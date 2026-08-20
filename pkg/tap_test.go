// Copyright (C) 2017-2021 Jan Delgado
//go:build integration

// TODO rewrite

package rabtap

// integration test functionality. assumes running rabbitmq broker on address
// defined by AMQP_URL and RABBIT_API_URL environment variables.
// (to start a local rabbitmq instance:
//  $ sudo  docker run --rm -ti -p5672:5672 rabbitmq:3-management)

import (
	"context"
	"crypto/tls"
	"log/slog"
	"testing"
	"time"
	"uuid"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ResultTimeout = time.Second * 5
	TapReadyDelay = time.Millisecond * 500
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
	success chan<- int,
) *AmqpTap {
	logger := slog.New(slog.DiscardHandler)
	tap := NewAmqpTap(testcommon.IntegrationURIFromEnv(), &tls.Config{}, logger)
	resultChannel := make(TapChannel)
	resultErrChannel := make(SubscribeErrorChannel)

	// TODO cancel and return cancel func
	ctx, cancel := context.WithCancel(context.Background())
	tapDone := make(chan struct{})
	go func() {
		defer close(tapDone)
		_ = tap.EstablishTap(
			ctx,
			[]ExchangeConfiguration{
				{tapExchangeName, tapQueueName},
			},
			resultChannel,
			resultErrChannel)
	}()

	func() {
		numReceived := 0

		// sample messages for 3 seconds and return number of returned messages
		// through the success channel
		for {
			select {
			case <-time.After(time.Second * 3):
				cancel()
				// wait for the tap's connection to be fully closed so its
				// auto-delete tap-exchange/-queue (and, if this was the last
				// binding, the tapped exchange itself) are torn down before
				// this function returns. Otherwise a subsequent test/run
				// re-declaring the same exchange can race with this teardown
				// and see a 404 NOT_FOUND on bind.
				<-tapDone
				success <- numReceived
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
	messagesPerTest := 5

	// establish sending exchange
	setup, err := testcommon.IntegrationTestConnection("headers-exchange", "headers", 2, true)
	require.NoError(t, err)
	defer func() { _ = setup.Conn.Close() }()

	finishChan := make(chan int)

	// no binding key is needed for the headers exchange
	go verifyMessagesOnTap(t, "tap-consumer1", messagesPerTest, "headers-exchange", "", finishChan)
	time.Sleep(TapReadyDelay)

	// inject messages into exchange. Each message should become visible
	// in the tap-exchange defined above. We use a headers exchange so we
	// must provide a amqp.Table struct with the messages headers, on which
	// routing is based. See integrationTestConnection() on how the routing
	// header is constructed.
	testcommon.PublishTestMessages(t, setup.Chan, messagesPerTest, "headers-exchange", "", amqp.Table{"header1": "test0"})

	requireIntFromChan(t, finishChan, messagesPerTest)

	// the original messages should also be delivered.
	testcommon.VerifyTestMessageOnQueue(t, setup.Chan, "consumer2", messagesPerTest, setup.QueueName(0), finishChan)
	requireIntFromChan(t, finishChan, messagesPerTest)
}

func TestIntegrationDirectExchange(t *testing.T) {
	// establish sending exchange
	setup, err := testcommon.IntegrationTestConnection("direct-exchange", "direct", 2, false)
	require.NoError(t, err)
	defer func() { _ = setup.Conn.Close() }()

	finishChan := make(chan int)

	// connect a test-tap and check if we received the test message
	messagesPerTest := 5

	go verifyMessagesOnTap(t, "tap-consumer1", messagesPerTest, "direct-exchange", setup.QueueName(0), finishChan)

	time.Sleep(TapReadyDelay)

	// inject messages into exchange. Each message should become visible
	// in the tap-exchange defined above.
	testcommon.PublishTestMessages(t, setup.Chan, messagesPerTest, "direct-exchange", setup.QueueName(0), nil)

	requireIntFromChan(t, finishChan, messagesPerTest)

	// the original messages should also be delivered.
	testcommon.VerifyTestMessageOnQueue(t, setup.Chan, "consumer2", messagesPerTest, setup.QueueName(0), finishChan)
	requireIntFromChan(t, finishChan, messagesPerTest)
}

// TestIntegrationTopicExchangeTapSingleQueue tests tapping to a topic
// exchange with a routing key so that only messages sent to one topic are
// tapped.
func TestIntegrationTopicExchangeTapSingleQueue(t *testing.T) {
	// establish sending exchange. use a unique exchange name per test run,
	// since it is auto-deleted when the last binding is removed and reusing
	// a fixed name races with that (still in-flight, async) teardown of a
	// previous run/test using the same name.
	exchangeName := "topic-exchange-" + uuid.New().String()
	setup, err := testcommon.IntegrationTestConnection(exchangeName, "topic", 2, false)
	require.NoError(t, err)
	defer func() { _ = setup.Conn.Close() }()

	finishChan := make(chan int)

	// connect a test-tap and check if we received the test message
	messagesPerTest := 5

	// tap only messages routed to queue-0
	go verifyMessagesOnTap(t, "tap-consumer1", messagesPerTest, exchangeName, setup.QueueName(0), finishChan)

	time.Sleep(TapReadyDelay)

	// inject messages into exchange. Each message should become visible
	// in the tap-exchange defined above.
	testcommon.PublishTestMessages(t, setup.Chan, messagesPerTest, exchangeName, setup.QueueName(0), nil)
	testcommon.PublishTestMessages(t, setup.Chan, messagesPerTest, exchangeName, setup.QueueName(1), nil)

	requireIntFromChan(t, finishChan, messagesPerTest)

	// the original messages should also be delivered.
	testcommon.VerifyTestMessageOnQueue(t, setup.Chan, "consumer2", messagesPerTest, setup.QueueName(0), finishChan)
	requireIntFromChan(t, finishChan, messagesPerTest)

	testcommon.VerifyTestMessageOnQueue(t, setup.Chan, "consumer3", messagesPerTest, setup.QueueName(1), finishChan)
	requireIntFromChan(t, finishChan, messagesPerTest)
}

// TestIntegrationTopicExchangeTapWildcard tests tapping to an exechange
// of type topic. The tap-exchange s bound with the binding-key '#'.
func TestIntegrationTopicExchangeTapWildcard(t *testing.T) {
	// establish sending exchange. use a unique exchange name per test run,
	// since it is auto-deleted when the last binding is removed and reusing
	// a fixed name races with that (still in-flight, async) teardown of a
	// previous run/test using the same name.
	exchangeName := "topic-exchange-" + uuid.New().String()
	setup, err := testcommon.IntegrationTestConnection(exchangeName, "topic", 2, false)
	require.NoError(t, err)
	defer func() { _ = setup.Conn.Close() }()

	finishChan := make(chan int)

	// connect a test-tap and check if we received the test message
	messagesPerTest := 5

	// tap all messages on the exchange
	go verifyMessagesOnTap(t, "tap-consumer1", messagesPerTest*2, exchangeName, "#", finishChan)

	time.Sleep(TapReadyDelay)

	// inject messages into exchange. Each message should become visible
	// in the tap-exchange defined above.
	testcommon.PublishTestMessages(t, setup.Chan, messagesPerTest, exchangeName, setup.QueueName(0), nil)
	testcommon.PublishTestMessages(t, setup.Chan, messagesPerTest, exchangeName, setup.QueueName(1), nil)

	requireIntFromChan(t, finishChan, messagesPerTest*2)

	// the original messages should also be delivered.
	testcommon.VerifyTestMessageOnQueue(t, setup.Chan, "consumer2", messagesPerTest, setup.QueueName(0), finishChan)
	requireIntFromChan(t, finishChan, messagesPerTest)

	testcommon.VerifyTestMessageOnQueue(t, setup.Chan, "consumer3", messagesPerTest, setup.QueueName(1), finishChan)
	requireIntFromChan(t, finishChan, messagesPerTest)
}

// TestIntegrationInvalidExchange tries to tap to a non existing exhange, we
// expect an error returned.
func TestIntegrationInvalidExchange(t *testing.T) {
	tapMessages := make(TapChannel)
	errChannel := make(SubscribeErrorChannel)
	logger := slog.New(slog.DiscardHandler)
	tap := NewAmqpTap(testcommon.IntegrationURIFromEnv(), &tls.Config{}, logger)
	ctx := context.Background()
	err := tap.EstablishTap(
		ctx,
		[]ExchangeConfiguration{
			{"nonexisting-exchange", "test"},
		},
		tapMessages,
		errChannel)

	assert.NotNil(t, err)
}
