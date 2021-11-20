// (c) copyright jan delgado 2017-2021
package rabtap

import (
	"context"
	"testing"
	"time"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestAmqpMessageLoopPanicsWithInvalidMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	out := make(TapChannel)
	in := make(chan interface{})
	done := make(chan bool)
	errOut := make(SubscribeErrorChannel)

	go func() {
		assert.Panics(t, func() { _, _ = amqpMessageLoop(ctx, out, errOut, in) }, "did not panic")
		done <- true
	}()

	// we expect amqpMessageLoop to panic when a non amqp.Delivery is passed in
	in <- 1337

	cancel()

	select {
	case result := <-done:
		assert.Equal(t, true, result)
	case <-time.After(2 * time.Second):
		assert.Fail(t, "amqpMessageLoop() did not terminate")
	}
}

func TestAmqpMessageLoopTerminatesWhenInputChannelIsClosed(t *testing.T) {
	ctx := context.Background()
	out := make(TapChannel)
	in := make(chan interface{})
	done := make(chan ReconnectAction)
	errOut := make(SubscribeErrorChannel)

	go func() {
		result, _ := amqpMessageLoop(ctx, out, errOut, in)
		done <- result
	}()

	close(in)

	select {
	case result := <-done:
		assert.Equal(t, doReconnect, result)
	case <-time.After(2 * time.Second):
		assert.Fail(t, "amqpMessageLoop() did not terminate")
	}
}

func TestAmqpMessageLoopCancelBlockingWrite(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	out := make(TapChannel)
	in := make(chan interface{}, 5)
	done := make(chan ReconnectAction)
	errOut := make(SubscribeErrorChannel)

	go func() {
		result, _ := amqpMessageLoop(ctx, out, errOut, in)
		done <- result
	}()

	in <- amqp.Delivery{}
	// this second write blocks the write in the messageloop
	in <- amqp.Delivery{}

	time.Sleep(1 * time.Second)
	cancel()

	select {
	case result := <-done:
		assert.Equal(t, doNotReconnect, result)
	case <-time.After(2 * time.Second):
		assert.Fail(t, "amqpMessageLoop() did not terminate")
	}

}

func TestAmqpMessageLoopForwardsAMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	out := make(TapChannel)
	in := make(chan interface{})
	done := make(chan ReconnectAction)
	errOut := make(SubscribeErrorChannel)

	go func() {
		result, _ := amqpMessageLoop(ctx, out, errOut, in)
		done <- result
	}()

	expected := amqp.Delivery{}
	in <- expected

	select {
	case msg := <-out:
		assert.Equal(t, expected, *msg.AmqpMessage)
	case <-time.After(2 * time.Second):
		assert.Fail(t, "amqpMessageLoop() did not terminate")
	}
	cancel()
}

func TestAmqpMessageLoopForwardsAnError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	out := make(TapChannel)
	in := make(chan interface{})
	done := make(chan ReconnectAction)
	errOut := make(SubscribeErrorChannel)

	go func() {
		result, _ := amqpMessageLoop(ctx, out, errOut, in)
		done <- result
	}()

	expected := &amqp.Error{}
	in <- expected

	select {
	case err := <-errOut:
		assert.Equal(t, SubscribeError{Reason: SubscribeErrorChannelError, Cause: expected}, *err)
	case <-time.After(2 * time.Second):
		assert.Fail(t, "amqpMessageLoop() did not terminate")
	}
	cancel()
}

func TestAmqpMessageLoopCancelBlockingRead(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	out := make(TapChannel)
	in := make(chan interface{})
	done := make(chan ReconnectAction)
	errOut := make(SubscribeErrorChannel)

	go func() {
		result, _ := amqpMessageLoop(ctx, out, errOut, in)
		done <- result
	}()

	cancel()

	select {
	case result := <-done:
		assert.Equal(t, doNotReconnect, result)
	case <-time.After(2 * time.Second):
		assert.Fail(t, "amqpMessageLoop() did not terminate")
	}
}
