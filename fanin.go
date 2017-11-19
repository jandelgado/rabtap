package rabtap

import (
	"reflect"

	tomb "gopkg.in/tomb.v2"
)

// Fanin allows to do a select ("fan-in") on an set of channels
type Fanin struct {
	Ch       chan interface{}
	channels []reflect.SelectCase
	t        tomb.Tomb
}

// NewFanin creates a new Fanin object
func NewFanin(channels []interface{}) *Fanin {
	fanin := Fanin{Ch: make(chan interface{})}
	fanin.add(fanin.t.Dying())
	for _, c := range channels {
		fanin.add(c)
	}

	fanin.t.Go(fanin.loop)
	return &fanin
}

func (s *Fanin) add(c interface{}) {
	s.channels = append(s.channels,
		reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(c)})
}

// Stop stops the fanin go-routine
func (s *Fanin) Stop() error {
	s.t.Kill(nil)
	return s.t.Wait()
}

// Alive returns true if the fanin is running
func (s *Fanin) Alive() bool {
	return s.t.Alive()
}

// Select wait for activity on any of the channels
func (s *Fanin) loop() error {

	for {
		chosen, message, ok := reflect.Select(s.channels)

		// channels[0] is always the tomb Dying() chan. Request to end fanin.
		if chosen == 0 {
			close(s.Ch) // note: sends nil message on s.Ch
			return nil
		}

		if !ok {
			// The chosen channel has been closed, so zero
			// out the channel to disable the case (happens on normal shutdown)
			s.channels[chosen].Chan = reflect.ValueOf(nil)
			// TODO end fanin if no channels remain?
		} else {
			s.Ch <- message.Interface()
		}
	}
}
