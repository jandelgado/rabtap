package rabtap

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func expectOnChan[T any](t *testing.T, val T, ch <-chan interface{}) {
	t.Helper()
	select {
	case message := <-ch:
		assert.Equal(t, val, message.(T))
	case <-time.After(1 * time.Second):
		assert.Fail(t, "did not receive message in expected time")
	}
}

func TestFaninReceivesFromMultipleChannels(t *testing.T) {

	chan1 := make(chan int, 1)
	defer close(chan1)
	chan2 := make(chan string, 1)
	defer close(chan2)
	fanin := Fanin(context.TODO(), []<-chan interface{}{Wrap(chan1), Wrap(chan2)})

	chan1 <- 99
	expectOnChan(t, 99, fanin)
	chan2 <- "hello"
	expectOnChan(t, "hello", fanin)
}

func TestFaninClosesChanWhenAllInputsAreClosed(t *testing.T) {

	chan1 := make(chan int)
	chan2 := make(chan int)
	fanin := Fanin(context.TODO(), []<-chan interface{}{Wrap(chan1), Wrap(chan2)})

	close(chan1)
	close(chan2)

	_, ok := <-fanin

	assert.False(t, ok)
}
