package rabtap

import (
	"context"
	"sync"
)

// WrapChan takes a channel of any type and returns a new channel of type interface{}
func WrapChan[T any](c <-chan T) <-chan interface{} {
	wrapped := make(chan interface{})
	go func() {
		for m := range c {
			wrapped <- m
		}
		close(wrapped)
	}()
	return wrapped
}

// Fanin selects sumultanously from an array of channels and sends received
// messages to a new channel ("fan-in" of channels)
func Fanin(ctx context.Context, channels []<-chan interface{}) chan interface{} {
	var wg sync.WaitGroup
	out := make(chan interface{})

	receiver := func(ctx context.Context, c <-chan interface{}) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case val, ok := <-c:
				if !ok {
					return
				}
				out <- val
			}
		}
	}

	wg.Add(len(channels))
	for _, c := range channels {
		go receiver(ctx, c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
