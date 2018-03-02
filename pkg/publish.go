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

// (re-)establish the connection to RabbitMQ in case the connection has died.
// this function is run in a go-routine. after the connection is established
// a channel is created and the list of provided exchanges is wire-tapped.
// To start the first connection process,  send an amqp.ErrClosed message
// through the errorChannel.
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
				//s.logger.Printf("publishing message %#v to %s/%s", message,
				//	message.Exchange, message.RoutingKey)
				err := channel.Publish(message.Exchange,
					message.RoutingKey,
					false, // not mandatory
					false, // not immeadiate
					*message.Publishing)

				if err != nil {
					s.logger.Print(err)
					// error publishing message
					// TODO send error back to client using an error channel?
					// TODO retry?
				}

			case controlMessage := <-controlChan:
				s.logger.Printf("received message on control channel: %#v", controlMessage)
				// true signals caller to re-connect, false to end processing
				if controlMessage.IsReconnect() {
					return doReconnect
				}
				return doNotReconnect
				//				return controlMessage == ReconnectMessage
			}
		}
	}
}

// EstablishConnection sets up the connection to the broker and sets up
// the tap, which is bound to the provided consumer function. Typically
// started as go-routine.
func (s *AmqpPublish) EstablishConnection(publishChannel PublishChannel) {
	s.connection.Connect(s.createWorkerFunc(publishChannel))
}

// Close closes the connection to the broker and ends tapping.
func (s *AmqpPublish) Close() error {
	return s.connection.Close()
}
