// Copyright (C) 2017-2021 Jan Delgado

package rabtap

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"time"

	"github.com/streadway/amqp"
)

// PublishMessage is a message to be published by AmqpPublish via
// a PublishChannel
type PublishMessage struct {
	Exchange   string
	RoutingKey string
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
func (s *AmqpPublish) createWorkerFunc(publishChannel PublishChannel) AmqpWorkerFunc {

	return func(ctx context.Context, session Session) (ReconnectAction, error) {

		// errors receives channel errors (e.g. publishing to non-existant exchange)
		errors := session.Channel.NotifyClose(make(chan *amqp.Error, 1))
		// return receivces unroutable messages back from the server
		returns := session.NotifyReturn(make(chan amqp.Return))
		// confirms receives confirmations from the server (if enabled below)
		confirms := session.NotifyPublish(make(chan amqp.Confirmation, 1))

		if s.confirms {
			if err := session.Confirm(false); err != nil {
				s.logger.Errorf("Channel could not be put into confirm mode: %s", err)
			}
		}
		// wait a while for outstanding errors and returned messages
		// since these can arrive after we finished publishing.

		// TODO when an error was detected on the errors channel (event loop),
		//      then the defer needs not to be run (???)
		defer func() {

			s.logger.Debugf("waiting for confirms & returns... ")
			timeout := time.After(time.Second * 1) // TODO config/const

			// wait for pending returned messages from the broker, whem
			// e.g. a message could not be routed. in this case the
			// message WILL be confirmed (ACK=true), but an async
			// return message will be send, for which we wait here.
			for {
				select {
				case <-timeout:
					return

				case returned, more := <-returns:
					if !more {
						continue
					}
					s.logger.Errorf("server returned message for exchange '%s' with routingkey '%s': %s",
						returned.Exchange, returned.RoutingKey, returned.ReplyText)

				// these events singal closing of the channel
				case err, more := <-errors:
					if !more {
						continue
					}
					s.logger.Errorf("channel error: %v", err)
				}
			}
		}()

		for {
			select {
			case err := <-errors:
				// all errors render the channel invalid, so reconnect
				s.logger.Errorf("channel error: %v", err)
				return doReconnect, fmt.Errorf("channel error: %w", err)

			case returned, more := <-returns:
				if more {
					s.logger.Errorf("server returned message for exchange '%s' with routingkey '%s': %s",
						returned.Exchange, returned.RoutingKey, returned.ReplyText)
				}

			case message, more := <-publishChannel:
				if !more {
					s.logger.Debugf("publishing channel closed.")
					return doNotReconnect, nil
				}

				if err := session.Publish(message.Exchange,
					message.RoutingKey,
					s.mandatory,
					// immeadiate flag was removed with RabbitMQ 3
					false,
					*message.Publishing); err != nil {

					s.logger.Errorf("publishing error %v, reconnecting", err)
				} else {

					// wait for the confirmation before publishing a new message
					if s.confirms {
						select {
						case <-time.After(2 * time.Second): // TODO MEMORY LEAK when not firing FIXME
							s.logger.Errorf("no confirmation for publish to '%s' with routing key '%s'", message.Exchange, message.RoutingKey)
						case confirmed := <-confirms:
							if !confirmed.Ack {
								s.logger.Errorf("delivery to exchange '%s' with routingkey '%s' and delivery tag #%d was not ACKed by the server",
									message.Exchange, message.RoutingKey, confirmed.DeliveryTag)
							} else {
								s.logger.Infof("delivery with delivery tag #%d was ACKed by the server",
									confirmed.DeliveryTag)
							}
						case <-ctx.Done():
							return doNotReconnect, nil
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
func (s *AmqpPublish) EstablishConnection(ctx context.Context, publishChannel PublishChannel) error {
	return s.connection.Connect(ctx, s.createWorkerFunc(publishChannel))
}
