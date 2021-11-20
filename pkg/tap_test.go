// Copyright (C) 2017-2021 Jan Delgado
// +build integration

package rabtap

// integration test functionality. assumes running rabbitmq broker on address
// defined by AMQP_URL and RABBIT_API_URL environment variables.
// (to start a local rabbitmq instance:
//  $ sudo  docker run --rm -ti -p5672:5672 rabbitmq:3-management)

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

const (
	MessagesPerTest = 5
	ResultTimeout   = time.Second * 5
	TapReadyDelay   = time.Millisecond * 500
)

func TestGetTapQueueNameForExchange(t *testing.T) {

	assert.Equal(t, "__tap-queue-for-exchange-1234",
		getTapQueueNameForExchange("exchange", "1234"))
}

func TestGetTapEchangeNameForExchange(t *testing.T) {

	assert.Equal(t, "__tap-exchange-for-exchange-1234",
		getTapExchangeNameForExchange("exchange", "1234"))
}

func verifyMessagesOnTap(t *testing.T, consumer string, numExpected int,
	tapExchangeName, tapQueueName string,
	success chan<- int) *AmqpTap {

	log := testcommon.NewTestLogger()
	tap := NewAmqpTap(testcommon.IntegrationURIFromEnv(), &tls.Config{}, log)
	resultChannel := make(TapChannel)
	resultErrChannel := make(SubscribeErrorChannel)

	// TODO cancel and return cancel func
	ctx, cancel := context.WithCancel(context.Background())
	go tap.EstablishTap(
		ctx,
		[]ExchangeConfiguration{
			{tapExchangeName, tapQueueName}},
		resultChannel,
		resultErrChannel)

	func() {
		numReceived := 0

		// sample messages for 3 seconds and return number of returned messages
		// through the success channel
		for {
			select {
			case <-time.After(time.Second * 3):
				success <- numReceived
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
	cancel()
	return tap
}

func requireIntFromChan(t *testing.T, c <-chan int, expected int) {
	select {
	case val := <-c:
		assert.Equal(t, expected, val)
		return
	case <-time.After(ResultTimeout):
		assert.Fail(t, "test result not received in expected time frame")
	}
}

func TestIntegrationHeadersExchange(t *testing.T) {

	// establish sending exchange
	conn, ch := testcommon.IntegrationTestConnection(t, "headers-exchange", "headers", 2, true)
	defer conn.Close()

	finishChan := make(chan int)

	// no binding key is needed for the headers exchange
	go verifyMessagesOnTap(t, "tap-consumer1", MessagesPerTest, "headers-exchange", "", finishChan)
	time.Sleep(TapReadyDelay)

	// inject messages into exchange. Each message should become visible
	// in the tap-exchange defined above. We use a headers exchange so we
	// must provide a amqp.Table struct with the messages headers, on which
	// routing is based. See integrationTestConnection() on how the routing
	// header is constructed.
	testcommon.PublishTestMessages(t, ch, MessagesPerTest, "headers-exchange", "", amqp.Table{"header1": "test0"})

	requireIntFromChan(t, finishChan, MessagesPerTest)

	// the original messages should also be delivered.
	testcommon.VerifyTestMessageOnQueue(t, ch, "consumer2", MessagesPerTest, "queue-0", finishChan)
	requireIntFromChan(t, finishChan, MessagesPerTest)
}

func TestIntegrationDirectExchange(t *testing.T) {

	// establish sending exchange
	conn, ch := testcommon.IntegrationTestConnection(t, "direct-exchange", "direct", 2, false)
	defer conn.Close()

	finishChan := make(chan int)

	// connect a test-tap and check if we received the test message
	MessagesPerTest := MessagesPerTest

	go verifyMessagesOnTap(t, "tap-consumer1", MessagesPerTest, "direct-exchange", "queue-0", finishChan)

	time.Sleep(TapReadyDelay)

	// inject messages into exchange. Each message should become visible
	// in the tap-exchange defined above.
	testcommon.PublishTestMessages(t, ch, MessagesPerTest, "direct-exchange", "queue-0", nil)

	requireIntFromChan(t, finishChan, MessagesPerTest)

	// the original messages should also be delivered.
	testcommon.VerifyTestMessageOnQueue(t, ch, "consumer2", MessagesPerTest, "queue-0", finishChan)
	requireIntFromChan(t, finishChan, MessagesPerTest)
}

// TestIntegrationTopicExchangeTapSingleQueue tests tapping to a topic
// exchange with a routing key so that only messages sent to one topic are
// tapped.
func TestIntegrationTopicExchangeTapSingleQueue(t *testing.T) {

	// establish sending exchange
	conn, ch := testcommon.IntegrationTestConnection(t, "topic-exchange", "topic", 2, false)
	defer conn.Close()

	finishChan := make(chan int)

	// connect a test-tap and check if we received the test message
	MessagesPerTest := MessagesPerTest

	// tap only messages routed to queue-0
	go verifyMessagesOnTap(t, "tap-consumer1", MessagesPerTest, "topic-exchange", "queue-0", finishChan)

	time.Sleep(TapReadyDelay)

	// inject messages into exchange. Each message should become visible
	// in the tap-exchange defined above.
	testcommon.PublishTestMessages(t, ch, MessagesPerTest, "topic-exchange", "queue-0", nil)
	testcommon.PublishTestMessages(t, ch, MessagesPerTest, "topic-exchange", "queue-1", nil)

	requireIntFromChan(t, finishChan, MessagesPerTest)

	// the original messages should also be delivered.
	testcommon.VerifyTestMessageOnQueue(t, ch, "consumer2", MessagesPerTest, "queue-0", finishChan)
	requireIntFromChan(t, finishChan, MessagesPerTest)

	testcommon.VerifyTestMessageOnQueue(t, ch, "consumer3", MessagesPerTest, "queue-1", finishChan)
	requireIntFromChan(t, finishChan, MessagesPerTest)
}

// TestIntegrationTopicExchangeTapWildcard tests tapping to an exechange
// of type topic. The tap-exchange s bound with the binding-key '#'.
func TestIntegrationTopicExchangeTapWildcard(t *testing.T) {

	// establish sending exchange
	conn, ch := testcommon.IntegrationTestConnection(t, "topic-exchange", "topic", 2, false)
	defer conn.Close()

	finishChan := make(chan int)

	// connect a test-tap and check if we received the test message
	MessagesPerTest := MessagesPerTest

	// tap all messages on the exchange
	go verifyMessagesOnTap(t, "tap-consumer1", MessagesPerTest*2, "topic-exchange", "#", finishChan)

	time.Sleep(TapReadyDelay)

	// inject messages into exchange. Each message should become visible
	// in the tap-exchange defined above.
	testcommon.PublishTestMessages(t, ch, MessagesPerTest, "topic-exchange", "queue-0", nil)
	testcommon.PublishTestMessages(t, ch, MessagesPerTest, "topic-exchange", "queue-1", nil)

	requireIntFromChan(t, finishChan, MessagesPerTest*2)

	// the original messages should also be delivered.
	testcommon.VerifyTestMessageOnQueue(t, ch, "consumer2", MessagesPerTest, "queue-0", finishChan)
	requireIntFromChan(t, finishChan, MessagesPerTest)

	testcommon.VerifyTestMessageOnQueue(t, ch, "consumer3", MessagesPerTest, "queue-1", finishChan)
	requireIntFromChan(t, finishChan, MessagesPerTest)
}

// TestIntegrationInvalidExchange tries to tap to a non existing exhange, we
// expect an error returned.
func TestIntegrationInvalidExchange(t *testing.T) {

	tapMessages := make(TapChannel)
	errChannel := make(SubscribeErrorChannel)
	log := testcommon.NewTestLogger()
	tap := NewAmqpTap(testcommon.IntegrationURIFromEnv(), &tls.Config{}, log)
	ctx := context.Background()
	err := tap.EstablishTap(
		ctx,
		[]ExchangeConfiguration{
			{"nonexisting-exchange", "test"}},
		tapMessages,
		errChannel)

	assert.NotNil(t, err)
}
