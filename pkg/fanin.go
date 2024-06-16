package rabtap

import (
	"context"
	"sync"
)

func Wrap[T any](c <-chan T) <-chan interface{} {
	wrapped := make(chan interface{}, 0)
	go func() {
		for m := range c {
			wrapped <- m
		}
		close(wrapped)
	}()
	return wrapped
}

// Fanin allows to do a select ("fan-in") on an set of channels
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
