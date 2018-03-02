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

func TestFaninMulti(t *testing.T) {

	// create fanin of 3 int channels
	chan1 := make(chan interface{})
	chan2 := make(chan interface{})
	chan3 := make(chan interface{})
	fanin := NewFanin([]interface{}{chan1, chan2, chan3})

	assert.True(t, fanin.Alive())

	go func() {
		chan1 <- 99
		chan2 <- 100
		chan3 <- 101
		fanin.Stop()
	}()

	expectIntOnChan(t, 99, fanin.Ch)
	expectIntOnChan(t, 100, fanin.Ch)
	expectIntOnChan(t, 101, fanin.Ch)

	// fanin.Stop() closes fanin channel which in turn sends nil message
	expectNilOnChan(t, fanin.Ch)

	assert.False(t, fanin.Alive())
}
