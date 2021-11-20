// Copyright (C) 2017-2021 Jan Delgado
// +build integration

package rabtap

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/streadway/amqp"
)

func TestSubscribeReceivesMessages(t *testing.T) {

	// given

	// establish sending exchange.
	conn, ch := testcommon.IntegrationTestConnection(t, "subtest-direct-exchange", "direct", 0, false)
	session := Session{conn, ch}
	defer conn.Close()

	queueName := "queue"
	keyName := queueName // since using direct exchange

	// we need to create the queue non-exclusive, since exclusive queues are
	// bound to the connection which created them (other connections get
	// error RESOURCE_LOCKED (405)).
	CreateQueue(session, queueName, false /*durable*/, true /*ad*/, false /*excl*/, nil)
	BindQueueToExchange(session, queueName, keyName, "subtest-direct-exchange", amqp.Table{})

	finishChan := make(chan int)

	config := AmqpSubscriberConfig{Exclusive: false, AutoAck: true}
	log := testcommon.NewTestLogger()
	subscriber := NewAmqpSubscriber(config, testcommon.IntegrationURIFromEnv(), &tls.Config{}, log)
	resultChannel := make(TapChannel)
	resultErrChannel := make(SubscribeErrorChannel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go subscriber.EstablishSubscription(ctx, queueName, resultChannel, resultErrChannel)

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
	testcommon.PublishTestMessages(t, ch, MessagesPerTest, "subtest-direct-exchange", queueName, nil)

	// then
	requireIntFromChan(t, finishChan, MessagesPerTest)
}
