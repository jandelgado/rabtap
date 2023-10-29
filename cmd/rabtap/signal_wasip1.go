//go:build wasip1

package main

import (
	"context"
)

func SigIntHandler(ctx context.Context, cancel func()) {
	// does currently not work with WASM and makes the program hang.
	// signal.Notify(c, os.Interrupt)
}
