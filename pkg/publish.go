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

// Connected returns true if the tap is connected to an exchange, otherwise
// false
func (s *AmqpPublish) Connected() bool {
	return s.connection.Connected()
}

// createWorkerFunc receives messages on the provided channel and publishes
// the messages on an rabbitmq exchange
func (s *AmqpPublish) createWorkerFunc(publishChannel PublishChannel) AmqpWorkerFunc {

	return func(rabbitConn *amqp.Connection, controlChan chan ControlMessage) ReconnectAction {
		channel, err := rabbitConn.Channel()
		if err != nil {
			return doReconnect
		}
		defer channel.Close()
		errChan := make(chan *amqp.Error, 10)
		channel.NotifyClose(errChan)

		for {
			select {
			case err := <-errChan:
				s.logger.Fatalf("publishing error %#+v", err)
				return doReconnect

			case message, more := <-publishChannel:
				if !more {
					s.logger.Print("publishing channel closed.")
					return doNotReconnect
				}
				// TODO need to add notification hdlr to detect pub errors
				err := channel.Publish(message.Exchange,
					message.RoutingKey,
					false, // not mandatory
					false, // not immeadiate
					*message.Publishing)

				if err != nil {
					s.logger.Print("publishing error %#+v", err)
					// error publishing message
					// channel is invalid now - re-connect
					return doReconnect
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

// EstablishConnection sets up the connection to the broker
func (s *AmqpPublish) EstablishConnection(publishChannel PublishChannel) error {
	err := s.connection.Connect(s.createWorkerFunc(publishChannel))
	return err
}

// Close closes the connection to the broker
func (s *AmqpPublish) Close() error {
	return s.connection.Close()
}
