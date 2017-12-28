// Copyright (C) 2017 Jan Delgado

package rabtap

import (
	"crypto/tls"

	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// AmqpSubscriber allows to tap to subscribe to queues
type AmqpSubscriber struct {
	connection *AmqpConnector
	logger     logrus.StdLogger
}

// NewAmqpSubscriber returns a new AmqpSubscriber object associated with the
// RabbitMQ broker denoted by the uri parameter.
func NewAmqpSubscriber(uri string, tlsConfig *tls.Config, logger logrus.StdLogger) *AmqpSubscriber {
	return &AmqpSubscriber{
		connection: NewAmqpConnector(uri, tlsConfig, logger),
		logger:     logger}
}

// TapMessage objects are passed through a tapChannel from tap to client
// either AmqpMessage or Error is set
type TapMessage struct {
	AmqpMessage *amqp.Delivery
	Error       error
}

// TapChannel is a channel for *TapMessage objects
type TapChannel chan *TapMessage

// Close closes the connection to the broker and ends tapping. Returns result
// of amqp.Connection.Close() operation.
func (s *AmqpSubscriber) Close() error {
	return s.connection.Close()
}

// Connected returns true if the tap is connected to an exchange, otherwise
// false
func (s *AmqpSubscriber) Connected() bool {
	return s.connection.Connected()
}

// EstablishSubscription sets up the connection to the broker and sets up
// the tap, which is bound to the provided consumer function. Typically
// this function is run as a go-routine.
func (s *AmqpSubscriber) EstablishSubscription(queueName string, tapCh TapChannel) {
	s.connection.Connect(s.createWorkerFunc(queueName, tapCh))
}

func (s *AmqpSubscriber) createWorkerFunc(
	queueName string, tapCh TapChannel) AmqpWorkerFunc {

	return func(rabbitConn *amqp.Connection, controlCh chan ControlMessage) ReconnectAction {
		ch, err := s.consumeMessages(rabbitConn, queueName)
		if err != nil {
			tapCh <- &TapMessage{nil, err}
			return doReconnect
		}
		// messageloop expects Fanin object, which expects array of channels.
		var channels []interface{}
		fanin := NewFanin(append(channels, ch))
		return s.messageLoop(tapCh, fanin, controlCh)
	}
}

// messageLoop forwards incoming amqp messages from the fanin to the provided
// tapCh.
func (s *AmqpSubscriber) messageLoop(tapCh TapChannel,
	fanin *Fanin, controlCh <-chan ControlMessage) ReconnectAction {

	for {
		select {
		case message := <-fanin.Ch:
			//s.logger.Printf("AmqpSubscriber: received message %#v", message)
			amqpMessage, _ := message.(amqp.Delivery)
			tapCh <- &TapMessage{&amqpMessage, nil}

		case controlMessage := <-controlCh:
			switch controlMessage {
			case shutdownMessage:
				s.logger.Printf("AmqpSubscriber: shutdown")
				return doNotReconnect
			case reconnectMessage:
				s.logger.Printf("AmqpSubscriber: ending worker due to reconnect")
				return doReconnect
			}
		}
	}
}

func (s *AmqpSubscriber) consumeMessages(conn *amqp.Connection,
	queueName string) (<-chan amqp.Delivery, error) {

	var ch *amqp.Channel
	var err error

	if ch, err = conn.Channel(); err != nil {
		return nil, err
	}

	msgs, err := ch.Consume(
		queueName,
		"__rabtap-consumer-"+uuid.NewV4().String()[:8], // TODO param
		true,  // auto-ack
		true,  // exclusive
		false, // no-local - unsupported
		false, // wait
		nil,   // args
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}
