// Copyright (C) 2017 Jan Delgado

package rabtap

import (
	"context"
	"crypto/tls"
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

		// errors receives channel errors (e.g. publishing to non-existant exchange)
		errors := make(chan *amqp.Error, 1)
		// return receivces unroutable messages back from the server
		returns := make(chan amqp.Return, 1)
		// confirms receives confirmations from the server
		confirms := make(chan amqp.Confirmation, 1)

		if reliable {
			if err := session.Confirm(false); err != nil {
				s.logger.Printf("WARNING Channel could not be put into confirm mode: %s", err)
			} else {

				errors = session.Channel.NotifyClose(errors)
				returns = session.NotifyReturn(returns)
				session.Confirm(false)
				confirms = session.NotifyPublish(confirms)

				// wait a while for outstanding confirmations
				defer func() {
					s.logger.Printf("waiting for confirms & returns...")
					timeout := time.After(time.Second * 2)
					for {
						select {
						case <-timeout:
							s.logger.Printf("Done waiting for confirmations & returns")
							return

						case err, more := <-errors:
							if !more {
								continue
							}
							s.logger.Printf("y publishing error:%+v", err)
						case returned, more := <-returns:
							if !more {
								continue
							}
							s.logger.Printf("y message returned by server: %s -> %s: %s", returned.Exchange, returned.RoutingKey, returned.ReplyText)

						case confirmed, more := <-confirms:
							if !more {
								continue
							}
							s.logger.Printf("y delivery with delivery tag: %d - %v", confirmed.DeliveryTag, confirmed.Ack)
						}
					}
				}()

			}
		}

		for {
			select {
			case err := <-errors:
				s.logger.Printf("x publishing error:%+v", err)

			case returned := <-returns:
				s.logger.Printf("x message returned by server: %s -> %s: %s", returned.Exchange, returned.RoutingKey, returned.ReplyText)

			case confirmed := <-confirms:
				s.logger.Printf("x delivery with delivery tag: %d - %v", confirmed.DeliveryTag, confirmed.Ack)

			case message, more := <-publishChannel:
				if !more {
					s.logger.Infof("publishing channel closed.")
					return doNotReconnect, nil
				}
				//TODO need to add notification hdlr to detect pub errors
				err := session.Publish(message.Exchange,
					message.RoutingKey,
					reliable, // not mandatory	// ENABLE RETURNS
					false,    //true, // not immeadiate
					*message.Publishing)

				if err != nil {
					s.logger.Errorf("publishing error %w", err)
					// error publishing message - reconnect.
					return doReconnect, err
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
