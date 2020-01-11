// Copyright (C) 2017 Jan Delgado

package rabtap

import (
	"context"
	"time"

	"github.com/streadway/amqp"
)

// amqpMessageLoop forwards incoming amqp messages from an "in" chan to an "out"
// chan, transforming them into TapMessage objects. Can be terminated
// using provided ctx or by closing the in chan.
func amqpMessageLoop(ctx context.Context,
	out TapChannel, in <-chan interface{}) ReconnectAction {

	for {
		select {
		case message, more := <-in:
			if !more {
				return doReconnect
			}
			// in is chan interface{} because we use FanIn.Ch
			amqpMessage, ok := message.(amqp.Delivery)

			if !ok {
				panic("amqp.Delivery expected")
			}

			received := time.Now()
			// Avoid blocking write to out when e.g. on the other end of the
			// channel the user pressed Ctrl+S to stop console output
			select {
			case <-ctx.Done():
				return doNotReconnect
			case out <- NewTapMessage(&amqpMessage, received):
			}

		case <-ctx.Done():
			return doNotReconnect
		}
	}
}
