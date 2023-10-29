//go:build !wasip1

package main

import (
	"context"
	"os"
	"os/signal"
)

func SigIntHandler(ctx context.Context, cancel func()) {
	// translate ^C (Interrput) in ctx.Done()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt) // NOT WITH WASM!
	select {
	case <-c:
		cancel()
	case <-ctx.Done():
	}

	signal.Stop(c)
}
