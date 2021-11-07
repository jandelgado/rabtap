// Copyright (C) 2017 Jan Delgado

package rabtap

import (
	"context"
	"crypto/tls"
	"net/url"
	"time"

	uuid "github.com/google/uuid"
	"github.com/streadway/amqp"
)

const PrefetchCount = 1
const PrefetchSize = 0

// AmqpSubscriberConfig stores configuration of the subscriber
type AmqpSubscriberConfig struct {
	Exclusive bool
	AutoAck   bool
}

// AmqpSubscriber allows to tap to subscribe to queues
type AmqpSubscriber struct {
	config     AmqpSubscriberConfig
	connection *AmqpConnector
	logger     Logger
}

// NewAmqpSubscriber returns a new AmqpSubscriber object associated with the
// RabbitMQ broker denoted by the uri parameter.
func NewAmqpSubscriber(config AmqpSubscriberConfig, url *url.URL, tlsConfig *tls.Config, logger Logger) *AmqpSubscriber {
	return &AmqpSubscriber{
		config:     config,
		connection: NewAmqpConnector(url, tlsConfig, logger),
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
		return amqpMessageLoop(ctx, tapCh, fanin.Ch), nil
	}
}

func (s *AmqpSubscriber) consumeMessages(session Session,
	queueName string) (<-chan amqp.Delivery, error) {

	err := session.Qos(PrefetchCount, PrefetchSize, false)
	if err != nil {
		return nil, err
	}

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
