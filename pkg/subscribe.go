// Copyright (C) 2017 Jan Delgado

package rabtap

import (
	"context"
	"crypto/tls"
	"time"

	uuid "github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// AmqpSubscriberConfig stores configuration of the subscriber
type AmqpSubscriberConfig struct {
	Exclusive bool
	AutoAck   bool
}

// AmqpSubscriber allows to tap to subscribe to queues
type AmqpSubscriber struct {
	config     AmqpSubscriberConfig
	connection *AmqpConnector
	logger     logrus.StdLogger
}

// NewAmqpSubscriber returns a new AmqpSubscriber object associated with the
// RabbitMQ broker denoted by the uri parameter.
func NewAmqpSubscriber(config AmqpSubscriberConfig, uri string, tlsConfig *tls.Config, logger logrus.StdLogger) *AmqpSubscriber {
	return &AmqpSubscriber{
		config:     config,
		connection: NewAmqpConnector(uri, tlsConfig, logger),
		logger:     logger}
}

// TapMessage objects are passed through a tapChannel from tap to client
// either AmqpMessage or Error is set
type TapMessage struct {
	AmqpMessage       *amqp.Delivery
	Error             error
	ReceivedTimestamp time.Time
}

// NewTapMessage constructs a new TapMessage
func NewTapMessage(message *amqp.Delivery, err error, ts time.Time) TapMessage {
	return TapMessage{AmqpMessage: message, Error: err, ReceivedTimestamp: ts}
}

// TapChannel is a channel for *TapMessage objects
type TapChannel chan TapMessage

// Close closes the connection to the broker and ends tapping. Returns result
// of amqp.Connection.Close() operation.
// func (s *AmqpSubscriber) Close() error {
//     return s.connection.Close()
// }

// Connected returns true if the tap is connected to an exchange, otherwise
// false
// func (s *AmqpSubscriber) Connected() bool {
//     return s.connection.Connected()
// }

// EstablishSubscription sets up the connection to the broker and sets up
// the tap, which is bound to the provided consumer function. Typically
// this function is run as a go-routine.
func (s *AmqpSubscriber) EstablishSubscription(ctx context.Context, queueName string, tapCh TapChannel) error {
	err := s.connection.Connect(ctx, s.createWorkerFunc(queueName, tapCh))
	return err
}

func (s *AmqpSubscriber) createWorkerFunc(
	queueName string, tapCh TapChannel) AmqpWorkerFunc {

	return func(ctx context.Context, session Session) (ReconnectAction, error) {
		ch, err := s.consumeMessages(session, queueName)
		if err != nil {
			return doNotReconnect, err
		}
		// messageloop expects Fanin object, which expects array of channels.
		var channels []interface{}
		fanin := NewFanin(append(channels, ch))
		return s.messageLoop(ctx, tapCh, fanin), nil
	}
}

// messageLoop forwards incoming amqp messages from the fanin to the provided
// tapCh.
func (s *AmqpSubscriber) messageLoop(ctx context.Context, tapCh TapChannel,
	fanin *Fanin) ReconnectAction {

	for {
		select {
		case message, more := <-fanin.Ch:
			s.logger.Printf("AmqpSubscriber: more=%v", more)
			if !more {
				return doReconnect
			}
			amqpMessage, _ := message.(amqp.Delivery)
			tapCh <- NewTapMessage(&amqpMessage, nil, time.Now())

		case <-ctx.Done():
			return doNotReconnect
		}
	}
}

func (s *AmqpSubscriber) consumeMessages(session Session,
	queueName string) (<-chan amqp.Delivery, error) {

	msgs, err := session.Consume(
		queueName,
		"__rabtap-consumer-"+uuid.Must(uuid.NewRandom()).String()[:8], // TODO param
		s.config.AutoAck,
		s.config.Exclusive,
		false, // no-local - unsupported
		false, // wait
		nil,   // args
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}
