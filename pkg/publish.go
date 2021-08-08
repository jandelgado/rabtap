// Copyright (C) 2017 Jan Delgado

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
}

// NewAmqpPublish returns a new AmqpPublish object associated with the RabbitMQ
// broker denoted by the uri parameter.
func NewAmqpPublish(url *url.URL, tlsConfig *tls.Config, logger Logger) *AmqpPublish {
	return &AmqpPublish{
		connection: NewAmqpConnector(url, tlsConfig, logger),
		logger:     logger}
}

// One would typically keep a channel of publishings, a sequence number, and a
// set of unacknowledged sequence numbers and loop until the publishing channel
// is closed.
// func confirmOne(confirms <-chan amqp.Confirmation) {
//     log.Printf("waiting for confirmation of one publishing")

//     if confirmed := <-confirms; confirmed.Ack {
//         log.Printf("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
//     } else {
//         log.Printf("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
//     }
// }

// createWorkerFunc receives messages on the provided channel and publishes
// the messages on an rabbitmq exchange
// TODO retry on failed publish
// TODO publish notification handler to detect problems

// Mandatory flag:
// When a published message cannot be routed to any queue (e.g. because there are
// no bindings defined for the target exchange), and the publisher set the
// mandatory message property to false (this is the default), the message is
// discarded or republished to an alternate exchange, if any.

// When a published message cannot be routed to any queue, and the publisher set
// the mandatory message property to true, the message will be returned to it. The
// publisher must have a returned message handler set up in order to handle the
//
// See https://www.rabbitmq.com/publishers.html#unroutable
func (s *AmqpPublish) createWorkerFunc(publishChannel PublishChannel) AmqpWorkerFunc {

	return func(ctx context.Context, session Session) (ReconnectAction, error) {

		reliable := true

		numPublished, numConfirmed, numReturned := 0, 0, 0

		// errors receives channel errors (e.g. publishing to non-existant exchange)
		errors := make(chan *amqp.Error, 1)
		// return receivces unroutable messages back from the server
		returns := make(chan amqp.Return, 1)
		// confirms receives confirmations from the server
		confirms := make(chan amqp.Confirmation, 1)

		skipDefer := false

		if reliable {
			if err := session.Confirm(false); err != nil {
				s.logger.Errorf("Channel could not be put into confirm mode: %s", err)
			} else {

				errors = session.Channel.NotifyClose(errors)
				returns = session.NotifyReturn(returns)
				session.Confirm(false)
				confirms = session.NotifyPublish(confirms)

				// wait a while for outstanding errors and returned messages
				// since these can arrive after we finished publishing.

				// TODO when an error was detected on the errors channel (event loop),
				//      then the defer needs not to be run (???)
				defer func() {

					if skipDefer {
						s.logger.Debugf("skipping defer")
						return // no need to wait for events
					}

					s.logger.Debugf("waiting for confirms & returns...")
					timeout := time.After(time.Second * 1)
					// wait for pending returned messages from the broker, whem
					// e.g. a message could not be routed. in this case the
					// message WILL be confirmed (ACK=true), but an async
					// return message will be send, for which we wait here.
					for {
						// we got a feedback for all messages  -> no need to
						// further wait for feedback from the broker
						// if numPublished == (numReturned + numConfirmed) {
						//     return
						// }

						select {
						case <-timeout:
							return

						case err, more := <-errors:
							if !more {
								continue //return // error chan closed -> channel closed -> no need to wait here
							}
							s.logger.Errorf("y publishing error: %v", err)

						case returned, more := <-returns:
							if !more {
								continue
							}
							s.logger.Errorf("y server returned message for exchange %s with routingkey %s: %s",
								returned.Exchange, returned.RoutingKey, returned.ReplyText)
							numReturned++

							// case confirmed, more := <-confirms:
							//     if !more {
							//         continue
							//     }
							//     s.logger.Debugf("y CONFIRM %+v", confirmed)
							//     if !confirmed.Ack {
							//         s.logger.Errorf("y delivery with delivery tag: %d not ACKed", confirmed.DeliveryTag)
							//     }
						}
					}
				}()

			}
		}

		for {
			select {
			case err := <-errors:
				// all errors render the channel invalid, so reconnect
				s.logger.Errorf("x publishing error (async): %v", err)
				skipDefer = true
				return doReconnect, fmt.Errorf("publishing error (async) %w", err)

			case returned, more := <-returns:
				if more {
					numReturned++
					s.logger.Errorf("x server returned message for exchange %s with routingkey %s: %s",
						returned.Exchange, returned.RoutingKey, returned.ReplyText)
				}

			// case confirmed := <-confirms:
			//     // TODO keep track of Acked/Nacked messages
			//     s.logger.Debugf("x CONFIRM %+v", confirmed)
			//     if !confirmed.Ack {
			//         s.logger.Errorf("x delivery with delivery tag: %d not ACKed", confirmed.DeliveryTag)
			//     }

			case message, more := <-publishChannel:
				if !more {
					s.logger.Infof("x publishing channel closed.")
					return doNotReconnect, nil
				}

				err := session.Publish(message.Exchange,
					message.RoutingKey,
					reliable, // not mandatory	// ENABLE RETURNS
					false,    //true, // not immeadiate
					*message.Publishing)

				if err != nil {
					s.logger.Errorf("x publishing error %v, reconnecting", err)
					// error publishing message - reconnect.
					return doReconnect, err
				}
				numPublished++

				// wait for the confirmation before publishing a new message
				select {
				case <-time.After(2 * time.Second): // TODO MEMORY LEAK when not firing FIXME
					s.logger.Errorf("x no confirmation for TODO")
				case confirmed := <-confirms:
					numConfirmed++
					// TODO keep track of Acked/Nacked messages
					if !confirmed.Ack {
						s.logger.Errorf("x delivery with delivery tag #%d was not ACKed by the server", confirmed.DeliveryTag)
					} else {
						s.logger.Debugf("x delivery with delivery tag #%d was ACKed by the server", confirmed.DeliveryTag)
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
