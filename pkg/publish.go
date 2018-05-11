// Copyright (C) 2017 Jan Delgado

package rabtap

import (
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
}

// PublishChannel is a channel for PublishMessage message objects
type PublishChannel chan *PublishMessage

// AmqpPublish allows to send to a RabbitMQ exchange.
type AmqpPublish struct {
	logger     logrus.StdLogger
	connection *AmqpConnector
}

// NewAmqpPublish returns a new AmqpTap object associated with the RabbitMQ
// broker denoted by the uri parameter.
func NewAmqpPublish(uri string, tlsConfig *tls.Config, logger logrus.StdLogger) *AmqpPublish {
	return &AmqpPublish{
		connection: NewAmqpConnector(uri, tlsConfig, logger),
		logger:     logger}
}

// Connected returns true if the tap is connected to an exchange, otherwise
// false
func (s *AmqpPublish) Connected() bool {
	return s.connection.Connected()
}

// createWorkerFunc receives messages on the provides channel and publishes
// the messages on an rabbitmq exchange
func (s *AmqpPublish) createWorkerFunc(publishChannel PublishChannel) AmqpWorkerFunc {

	return func(rabbitConn *amqp.Connection, controlChan chan ControlMessage) ReconnectAction {

		channel, err := rabbitConn.Channel()
		if err != nil {
			return doReconnect
		}
		defer channel.Close()

		for {
			select {
			case message, more := <-publishChannel:
				if !more {
					s.logger.Print("publishing channel closed.")
					return doNotReconnect
				}
				err := channel.Publish(message.Exchange,
					message.RoutingKey,
					false, // not mandatory
					false, // not immeadiate
					*message.Publishing)

				if err != nil {
					s.logger.Print(err)
					// error publishing message
					// TODO should we do something here, e.g. retry?
				}

			case controlMessage := <-controlChan:
				if controlMessage.IsReconnect() {
					return doReconnect
				}
				return doNotReconnect
			}
		}
	}
}

// EstablishConnection sets up the connection to the broker and sets up
// the tap, which is bound to the provided consumer function. Typically
// started as go-routine.
func (s *AmqpPublish) EstablishConnection(publishChannel PublishChannel) error {
	return s.connection.Connect(s.createWorkerFunc(publishChannel))
}

// Close closes the connection to the broker and ends tapping.
func (s *AmqpPublish) Close() error {
	return s.connection.Close()
}
