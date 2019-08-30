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
type TapMessage struct {
	AmqpMessage       *amqp.Delivery
	ReceivedTimestamp time.Time
}

// NewTapMessage constructs a new TapMessage
func NewTapMessage(message *amqp.Delivery, ts time.Time) TapMessage {
	return TapMessage{AmqpMessage: message, ReceivedTimestamp: ts}
}

// TapChannel is a channel for *TapMessage objects
type TapChannel chan TapMessage

// EstablishSubscription sets up the connection to the broker and sets up
// the tap, which is bound to the provided consumer function. Typically
// this function is run as a go-routine.
func (s *AmqpSubscriber) EstablishSubscription(ctx context.Context, queueName string, tapCh TapChannel) error {
	return s.connection.Connect(ctx, s.createWorkerFunc(queueName, tapCh))
}

func (s *AmqpSubscriber) createWorkerFunc(
	queueName string, tapCh TapChannel) AmqpWorkerFunc {

	return func(ctx context.Context, session Session) (ReconnectAction, error) {
		ch, err := s.consumeMessages(session, queueName)
		if err != nil {
			return doNotReconnect, err
		}
		// messageLoop expects Fanin object, which expects array of channels.
		var channels []interface{}
		fanin := NewFanin(append(channels, ch))
		return s.messageLoop(ctx, tapCh, fanin), nil
	}
}

// messageLoop forwards incoming amqp messages from the fanin to the provided
// tapCh.
// TODO need not be "method"
// TODO pass chan instead of Fanin and using fanin.Ch
func (s *AmqpSubscriber) messageLoop(ctx context.Context, tapCh TapChannel,
	fanin *Fanin) ReconnectAction {

	for {

		select {
		case message, more := <-fanin.Ch:
			if !more {
				return doReconnect
			}

			amqpMessage, _ := message.(amqp.Delivery)
			// Avoid blocking write to tapCh when e.g. on the other end of the
			// channel the user pressed Ctrl+S to stop console output
			select {
			case tapCh <- NewTapMessage(&amqpMessage, time.Now()):
			case <-ctx.Done():
				return doNotReconnect
			}

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
