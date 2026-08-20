// Copyright (C) 2017-2021 Jan Delgado
//go:build integration

package rabtap

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"testing"
	"time"
	"uuid"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscribeReceivesMessages(t *testing.T) {
	// given

	// establish sending exchange.
	messagesPerTest := 5
	setup, err := testcommon.IntegrationTestConnection("subtest-direct-exchange", "direct", 0, false)
	require.NoError(t, err)
	session := Session{setup.Conn, setup.Chan}
	defer func() { _ = setup.Conn.Close() }()

	queueName := fmt.Sprintf("sub-test-%s", uuid.New().String())
	keyName := queueName // since using direct exchange

	// we need to create the queue non-exclusive, since exclusive queues are
	// bound to the connection which created them (other connections get
	// error RESOURCE_LOCKED (405)).
	err = CreateQueue(session, queueName, true /*durable*/, true /*ad*/, false /*excl*/, nil)
	require.NoError(t, err)
	err = BindQueueToExchange(session, queueName, keyName, "subtest-direct-exchange", amqp.Table{})
	assert.NoError(t, err)

	finishChan := make(chan int)

	config := AmqpSubscriberConfig{Exclusive: false}
	logger := slog.New(slog.DiscardHandler)
	subscriber := NewAmqpSubscriber(config, testcommon.IntegrationURIFromEnv(), &tls.Config{}, logger)
	resultChannel := make(TapChannel)
	resultErrChannel := make(SubscribeErrorChannel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = subscriber.EstablishSubscription(ctx, queueName, resultChannel, resultErrChannel) }()

	go func() {
		numReceived := 0

		// sample messages for 3 seconds and return number of returned messages
		// through the success channel
		for {
			select {
			case <-time.After(time.Second * 3):
				finishChan <- numReceived
				return
			case message := <-resultChannel:
				_ = message.AmqpMessage.Ack(false)
				if message.AmqpMessage != nil {
					if string(message.AmqpMessage.Body) == "Hello" {
						numReceived++
					}
				}
			}
		}
	}()

	time.Sleep(TapReadyDelay)

	// when: inject messages into exchange.
	testcommon.PublishTestMessages(t, setup.Chan, messagesPerTest, "subtest-direct-exchange", queueName, nil)

	// then
	requireIntFromChan(t, finishChan, messagesPerTest)
}
