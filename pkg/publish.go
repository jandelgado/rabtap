// Copyright (C) 2017-2021 Jan Delgado

package rabtap

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const timeoutWaitACK = time.Second * 2
const timeoutWaitServer = time.Millisecond * 500

// PublishMessage is a message to be published by AmqpPublish via
// a PublishChannel
type PublishMessage struct {
	Routing    Routing
	Publishing *amqp.Publishing
}

// PublishChannel is a channel for PublishMessage message objects
type PublishChannel chan *PublishMessage

// AmqpPublish allows to send to a RabbitMQ exchange.
type AmqpPublish struct {
	logger     Logger
	connection *AmqpConnector
	mandatory  bool
	confirms   bool
}

type PublishErrorReason int

const (
	PublishErrorAckTimeout PublishErrorReason = iota
	PublishErrorNack
	PublishErrorPublishFailed
	PublishErrorReturned
	PublishErrorChannelError
)

// PublishError is sent back trough the error channel when there are problems
// during the publishing of messages
type PublishError struct {
	Reason PublishErrorReason
	// Publishing stores the original message, if available (AckTimeout, Nack,
	// PublishFailed)
	Message *PublishMessage
	// ReturnedMessage stores the returned message in case of PublishErrorReturned
	ReturnedMessage *amqp.Return
	// Cause holds the error when a ChannelError happened
	Cause error
}

type PublishErrorChannel chan *PublishError

func (s *PublishError) Error() string {
	switch s.Reason {
	case PublishErrorAckTimeout:
		return fmt.Sprintf("publish to %s failed: timeout waiting for ACK",
			s.Message.Routing)
	case PublishErrorNack:
		return fmt.Sprintf("publish to %s failed: NACK",
			s.Message.Routing)
	case PublishErrorPublishFailed:
		return fmt.Sprintf("publish to %s failed: %s",
			s.Message.Routing, s.Cause)
	case PublishErrorReturned:
		// note: RabbitMQ seems not to set the headers on a returned message
		// when e.g. header based routing was used.
		routing := NewRouting(s.ReturnedMessage.Exchange,
			s.ReturnedMessage.RoutingKey,
			s.ReturnedMessage.Headers)
		return fmt.Sprintf("server returned message for %s: %s",
			routing, s.ReturnedMessage.ReplyText)
	case PublishErrorChannelError:
		return fmt.Sprintf("channel error: %s", s.Cause)
	}
	return "unexpected error"
}

// NewAmqpPublish returns a new AmqpPublish object associated with the RabbitMQ
// broker denoted by the uri parameter.
func NewAmqpPublish(url *url.URL, tlsConfig *tls.Config,
	mandatory, confirms bool, logger Logger) *AmqpPublish {
	return &AmqpPublish{
		connection: NewAmqpConnector(url, tlsConfig, logger),
		mandatory:  mandatory,
		confirms:   confirms,
		logger:     logger}
}

// createWorkerFunc creates a function that receives messages on the provided
// channel and publishes the messages on an rabbitmq exchange
//
// Mandatory flag:
// When a published message cannot be routed to any queue (e.g. because there are
// no bindings defined for the target exchange), and the publisher set the
// mandatory message property to false (this is the default), the message is
// discarded or republished to an alternate exchange, if any.
//
// When a published message cannot be routed to any queue, and the publisher set
// the mandatory message property to true, the message will be returned to it. The
// publisher must have a returned message handler set up in order to handle the
//
// See https://www.rabbitmq.com/publishers.html#unroutable
//
// The immedeate flag is not supported since RabbitMQ 3.0, see
// https://blog.rabbitmq.com/posts/2012/11/breaking-things-with-rabbitmq-3-0
//
// TODO detect throttling
// TODO simplify
func (s *AmqpPublish) createWorkerFunc(
	publishCh PublishChannel,
	errorCh PublishErrorChannel) AmqpWorkerFunc {

	return func(ctx context.Context, session Session) (ReconnectAction, error) {

		// errors receives channel errors (e.g. publishing to non-existant exchange)
		errors := session.Channel.NotifyClose(make(chan *amqp.Error, 1))
		// return receivces unroutable messages back from the server
		returns := session.NotifyReturn(make(chan amqp.Return, 1))
		// confirms receives confirmations from the server (if enabled below)
		confirms := session.NotifyPublish(make(chan amqp.Confirmation, 1))

		if s.confirms {
			if err := session.Confirm(false); err != nil {
				s.logger.Errorf("Channel could not be put into confirm mode: %s", err)
			}
		}

		// wait a while for outstanding errors and returned messages
		// since these can arrive after we finished publishing.
		defer func() {

			s.logger.Debugf("waiting for pending server messages ... ")
			timeout := time.After(timeoutWaitServer)

			// wait for pending returned messages from the broker, when e.g. a
			// message could not be routed. in this case the message WILL be
			// confirmed (ACK=true), but an async return message will be send,
			// for which we wait here.
			for {
				select {
				case <-timeout:
					return

				case returned, more := <-returns:
					if more {
						errorCh <- &PublishError{Reason: PublishErrorReturned, ReturnedMessage: &returned}
					}

				case err, more := <-errors:
					if more {
						errorCh <- &PublishError{Reason: PublishErrorChannelError, Cause: err}
					}
				}
			}
		}()

		ackTimeout := time.NewTimer(timeoutWaitACK)
		defer ackTimeout.Stop()

		for {
			select {
			case err := <-errors:
				// all errors render the channel invalid, so reconnect
				errorCh <- &PublishError{Reason: PublishErrorChannelError, Cause: err}
				return doReconnect, fmt.Errorf("channel error: %w", err)

			case returned, more := <-returns:
				if more {
					errorCh <- &PublishError{Reason: PublishErrorReturned, ReturnedMessage: &returned}
				}

			case message, more := <-publishCh:
				if !more {
					s.logger.Debugf("publishing channel closed.")
					return doNotReconnect, nil
				}

				size := len((*message.Publishing).Body)
				s.logger.Debugf("publish message to %s (%d bytes)", message.Routing, size)
				headers := EnsureAMQPTable(message.Routing.Headers()).(amqp.Table)
				message.Publishing.Headers = headers
				err := session.PublishWithContext(
					ctx,
					message.Routing.Exchange(),
					message.Routing.Key(),
					s.mandatory,
					false, // immeadiate flag was removed with RabbitMQ 3
					*message.Publishing)

				if err != nil {
					errorCh <- &PublishError{Reason: PublishErrorPublishFailed, Message: message, Cause: err}
				} else {

					// wait for the confirmation before publishing a new message
					// https://www.rabbitmq.com/confirms.html
					// TODO batched confirms
					//
					// "For unroutable messages, the broker will issue a confirm
					// once the exchange verifies a message won't route to any
					// queue (returns an empty list of queues). If the message
					// is also published as mandatory, the basic.return is sent
					// to the client before basic.ack. The same is true for
					// negative acknowledgements (basic.nack)."
					if s.confirms {
						ackTimeout.Reset(timeoutWaitACK)
					Outer:
						for {
							select {
							case <-ackTimeout.C:
								errorCh <- &PublishError{Reason: PublishErrorAckTimeout, Message: message}
								break Outer
							case returned, more := <-returns:
								if more {
									errorCh <- &PublishError{Reason: PublishErrorReturned, ReturnedMessage: &returned}
								}
							case confirmed := <-confirms:
								if !confirmed.Ack {
									errorCh <- &PublishError{Reason: PublishErrorNack, Message: message}
								} else {
									s.logger.Infof("delivery with delivery tag #%d was ACKed by the server",
										confirmed.DeliveryTag)
								}
								break Outer
							case <-ctx.Done():
								return doNotReconnect, nil
							}
						}
					}
				}

			case <-ctx.Done():
				return doNotReconnect, nil

			}
		}
	}
}

// EstablishConnection sets up the connection to the broker
func (s *AmqpPublish) EstablishConnection(
	ctx context.Context,
	publishChannel PublishChannel,
	errorChannel PublishErrorChannel) error {
	return s.connection.Connect(ctx, s.createWorkerFunc(publishChannel, errorChannel))
}
