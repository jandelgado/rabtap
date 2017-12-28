// Copyright (C) 2017 Jan Delgado
// +build integration

package rabtap

import (
	"crypto/tls"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jandelgado/rabtap/testhelper"
)

func TestSubscribe(t *testing.T) {

	// establish sending exchange.
	conn, ch := testhelper.IntegrationTestConnection(t, "subtest-direct-exchange", "direct", 0, false)
	defer conn.Close()

	queueName := "queue"
	keyName := queueName // since using direct exchange

	// we need to create the queue non-exclusive, since exclusive queues are
	// bound to the connection which created them (other connections get
	// error RESOURCE_LOCKED (405)).
	CreateQueue(ch, queueName, false /*durable*/, true /*ad*/, false /*excl*/)
	BindQueueToExchange(ch, queueName, keyName, "subtest-direct-exchange")

	finishChan := make(chan int)

	subscriber := NewAmqpSubscriber(testhelper.IntegrationURIFromEnv(), &tls.Config{}, log.New(os.Stderr, "", log.LstdFlags))
	defer subscriber.Close()
	resultChannel := make(TapChannel)
	go subscriber.EstablishSubscription(queueName, resultChannel)

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
				if message.Error != nil {
					log.Print(message.Error)
				}
				if message.AmqpMessage != nil {
					if string(message.AmqpMessage.Body) == "Hello" {
						numReceived++
					}
				}
			}
		}
	}()

	time.Sleep(TapReadyDelay)

	// inject messages into exchange.
	testhelper.PublishTestMessages(t, ch, MessagesPerTest, "subtest-direct-exchange", queueName, nil)
	requireIntFromChan(t, finishChan, MessagesPerTest)
}
