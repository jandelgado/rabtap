// Copyright (C) 2017 Jan Delgado

package rabtap

import (
	"context"
	"crypto/tls"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// PublishMessage is a message to be published by AmqpPublish via
// a PublishChannel
type PublishMessage struct {
	Exchange   string
	RoutingKey string
	Publishing *amqp.Publishing
	Error      *error
}

// PublishChannel is a channel for PublishMessage message objects
type PublishChannel chan *PublishMessage

// AmqpPublish allows to send to a RabbitMQ exchange.
type AmqpPublish struct {
	logger     logrus.StdLogger
	connection *AmqpConnector
}

// NewAmqpPublish returns a new AmqpPublish object associated with the RabbitMQ
// broker denoted by the uri parameter.
func NewAmqpPublish(uri string, tlsConfig *tls.Config, logger logrus.StdLogger) *AmqpPublish {
	return &AmqpPublish{
		connection: NewAmqpConnector(uri, tlsConfig, logger),
		logger:     logger}
}

// createWorkerFunc receives messages on the provided channel and publishes
// the messages on an rabbitmq exchange
// TODO retry on failed publish
// TODO publish notification handler to detect problems
func (s *AmqpPublish) createWorkerFunc(publishChannel PublishChannel) AmqpWorkerFunc {

	return func(ctx context.Context, session Session) (ReconnectAction, error) {

		for {
			select {
			case message, more := <-publishChannel:
				if !more {
					s.logger.Print("publishing channel closed.")
					return doNotReconnect, nil
				}
				// TODO need to add notification hdlr to detect pub errors
				err := session.Publish(message.Exchange,
					message.RoutingKey,
					false, // not mandatory
					false, // not immeadiate
					*message.Publishing)

				if err != nil {
					s.logger.Print("publishing error %#+v", err)
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
