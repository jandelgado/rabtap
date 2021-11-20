// subscribe to message queues
// Copyright (C) 2017-2021 Jan Delgado

package rabtap

import (
	"context"
	"crypto/tls"
	"fmt"
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
	Args      amqp.Table
}

// AmqpSubscriber allows to tap to subscribe to queues
type AmqpSubscriber struct {
	config     AmqpSubscriberConfig
	connection *AmqpConnector
	logger     Logger
}

type SubscribeErrorReason int

const (
	SubscribeErrorChannelError SubscribeErrorReason = iota
)

// SubscribeError is sent back trough the error channel when there are problems
// during the subsription of messages
type SubscribeError struct {
	Reason SubscribeErrorReason
	// Cause holds the error when a ChannelError happened
	Cause error
}

type SubscribeErrorChannel chan *SubscribeError

func (s *SubscribeError) Error() string {
	switch s.Reason {
	case SubscribeErrorChannelError:
		return fmt.Sprintf("channel error: %s", s.Cause)
	default:
		return "unexpected error"
	}
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
//
// queueName is the queue to subscribe to. tapCh is where the consumed messages
// are sent to. errCh is the channel where errors are sent to.
//
func (s *AmqpSubscriber) EstablishSubscription(
	ctx context.Context,
	queueName string,
	tapCh TapChannel,
	errCh SubscribeErrorChannel) error {
	return s.connection.Connect(ctx, s.createWorkerFunc(queueName, tapCh, errCh))
}

func (s *AmqpSubscriber) createWorkerFunc(
	queueName string,
	outCh TapChannel,
	errOutCh SubscribeErrorChannel) AmqpWorkerFunc {

	return func(ctx context.Context, session Session) (ReconnectAction, error) {
		ch, err := s.consumeMessages(session, queueName)
		if err != nil {
			return doNotReconnect, err
		}

		// also subscribe to channel close notifications
		amqpErrorCh := session.Channel.NotifyClose(make(chan *amqp.Error, 1))
		fanin := NewFanin([]interface{}{ch, amqpErrorCh})

		return amqpMessageLoop(ctx, outCh, errOutCh, fanin.Ch)
	}
}

func (s *AmqpSubscriber) consumeMessages(session Session,
	queueName string) (<-chan amqp.Delivery, error) {

	err := session.Qos(PrefetchCount, PrefetchSize, false)
	if err != nil {
		return nil, err
	}

	return session.Consume(
		queueName,
		"__rabtap-consumer-"+uuid.Must(uuid.NewRandom()).String()[:8], // TODO param
		false, // no auto-ack
		s.config.Exclusive,
		false, // no-local - unsupported
		false, // wait
		s.config.Args,
	)
}
