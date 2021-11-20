// Copyright (C) 2017-2021 Jan Delgado

package rabtap

import (
	"context"
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

// amqpMessageLoop forwards incoming amqp messages from an "in" chan to an "out"
// chan, transforming them into TapMessage objects. Can be terminated
// using provided ctx or by closing the in chan.
func amqpMessageLoop(
	ctx context.Context,
	outCh TapChannel,
	errOutCh SubscribeErrorChannel,
	inCh <-chan interface{}) (ReconnectAction, error) {

	for {
		select {
		// case err, more := <-errInCh:
		//     if !more {
		//         return doReconnect, fmt.Errorf("channel closed")
		//     }
		//     errOutCh <- &SubscribeError{Reason: SubscribeErrorChannelError, Cause: err}

		case message, more := <-inCh:
			if !more {
				return doReconnect, fmt.Errorf("no more messages")
			}

			switch msg := message.(type) {

			case *amqp.Error:
				// TODO ctx?
				errOutCh <- &SubscribeError{Reason: SubscribeErrorChannelError, Cause: msg}

			case amqp.Delivery:
				received := time.Now()
				// Avoid blocking write to out when e.g. on the other end of the
				// channel the user pressed Ctrl+S to stop console output
				// TODO ctx.Done really needed?
				select {
				case <-ctx.Done():
					return doNotReconnect, nil
				case outCh <- NewTapMessage(&msg, received):
				}
			default:
				panic("unknown message type")
			}

		case <-ctx.Done():
			return doNotReconnect, nil
		}
	}
}
