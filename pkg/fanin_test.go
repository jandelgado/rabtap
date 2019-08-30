package rabtap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func expectIntOnChan(t *testing.T, val int, ch <-chan interface{}) {
	select {
	case message := <-ch:
		assert.Equal(t, val, message.(int))
	case <-time.After(3 * time.Second):
		assert.Fail(t, "did not receive message in expected time")
	}
}

func expectNilOnChan(t *testing.T, ch <-chan interface{}) {
	select {
	case message := <-ch:
		assert.Nil(t, message)
	case <-time.After(3 * time.Second):
		assert.Fail(t, "did not receive message in expected time")
	}
}

func TestFaninReceivesFromMultipleChannels(t *testing.T) {

	chan1 := make(chan int)
	chan2 := make(chan int)
	chan3 := make(chan int)
	fanin := NewFanin([]interface{}{chan1, chan2, chan3})

	assert.True(t, fanin.Alive())

	go func() {
		chan1 <- 99
		chan2 <- 100
		chan3 <- 101
	}()

	expectIntOnChan(t, 99, fanin.Ch)
	expectIntOnChan(t, 100, fanin.Ch)
	expectIntOnChan(t, 101, fanin.Ch)

	// fanin.Stop() closes fanin channel which in turn sends nil message
	assert.Nil(t, fanin.Stop())
	expectNilOnChan(t, fanin.Ch)

	assert.False(t, fanin.Alive())
}

func TestFaninClosesChanWhenAllInputsAreClosed(t *testing.T) {

	chan1 := make(chan int)
	chan2 := make(chan int)
	fanin := NewFanin([]interface{}{chan1, chan2})

	go func() {
		chan1 <- 99
		chan2 <- 100
		close(chan1)
		close(chan2)
	}()

	expectIntOnChan(t, 99, fanin.Ch)
	expectIntOnChan(t, 100, fanin.Ch)

	// close of last channel closes fanin channel which in turn sends nil message
	expectNilOnChan(t, fanin.Ch)

	assert.False(t, fanin.Alive())
}
