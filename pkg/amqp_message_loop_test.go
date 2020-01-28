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

	go func() {
		assert.Panics(t, func() { amqpMessageLoop(ctx, out, in) }, "did not panic")
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

	go func() {
		done <- amqpMessageLoop(ctx, out, in)
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

	go func() {
		done <- amqpMessageLoop(ctx, out, in)
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

func TestAmqpMessageLoopCancelBlockingRead(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	out := make(TapChannel)
	in := make(chan interface{})
	done := make(chan ReconnectAction)

	go func() {
		done <- amqpMessageLoop(ctx, out, in)
	}()

	cancel()

	select {
	case result := <-done:
		assert.Equal(t, doNotReconnect, result)
	case <-time.After(2 * time.Second):
		assert.Fail(t, "amqpMessageLoop() did not terminate")
	}

}
