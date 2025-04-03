package main

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/jandelgado/rabtap/pkg/testcommon"
)

func TestCaptureOutputRaceConditions(t *testing.T) {
	// Number of concurrent operations to run
	const concurrentOps = 500

	// Use a wait group to ensure all goroutines complete
	var wg sync.WaitGroup
	wg.Add(concurrentOps)

	for i := 0; i < concurrentOps; i++ {
		go func(id int) {
			defer wg.Done()

			output := testcommon.CaptureOutput(func() {
				// Produce different amounts of output
				for j := 0; j < id%10+1; j++ {
					fmt.Printf("Goroutine %d output line %d\n", id, j)
					log.Debug("log entry from goroutine")
					time.Sleep(time.Millisecond * time.Duration(rand.Intn(5)))
				}
			})

			if len(output) == 0 {
				t.Errorf("Goroutine %d received empty output", id)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
}