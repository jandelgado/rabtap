// test for the message subscriber
// TODO cleaner separation between component and unit test
// Copyright (C) 2017-2019 Jan Delgado

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// a mocked amqp.Acknowldger to test our AcknowledgeFunc
type MockAcknowledger struct {
	// store values in a map so being able to manipulate in a value receiver
	values map[string]bool
}

func NewMockAcknowledger() MockAcknowledger {
	return MockAcknowledger{values: map[string]bool{}}
}

func (s MockAcknowledger) isAcked() bool    { return s.values["acked"] }
func (s MockAcknowledger) isNacked() bool   { return s.values["nacked"] }
func (s MockAcknowledger) isRequeued() bool { return s.values["requeued"] }

func (s MockAcknowledger) Ack(tag uint64, multiple bool) error {
	s.values["acked"] = true
	return nil
}

func (s MockAcknowledger) Nack(tag uint64, multiple, requeue bool) error {
	s.values["nacked"] = true
	s.values["requeued"] = requeue
	return nil
}

func (s MockAcknowledger) Reject(tag uint64, requeue bool) error {
	s.values["nacked"] = true
	s.values["requeued"] = requeue
	return nil
}

func TestCreateMessagePredicateProvidesMessageContext(t *testing.T) {

	// given a predicate that accesses message attributes
	pred, err := NewExprPredicate("msg.MessageId=='match123'")
	require.NoError(t, err)
	filterPred := createMessagePred(pred)

	// when we evalute the predicate for the test Messages
	expectedMatch := rabtap.TapMessage{AmqpMessage: &amqp.Delivery{MessageId: "match123"}}
	res, err := filterPred(expectedMatch)
	require.NoError(t, err)
	// then we expect the expression evaluated in the given context
	assert.True(t, res)

	expectedNoMatch := rabtap.TapMessage{AmqpMessage: &amqp.Delivery{MessageId: "no match"}}
	res, err = filterPred(expectedNoMatch)
	require.NoError(t, err)
	assert.False(t, res)
}

func TestCreateAcknowledgeFuncReturnedFuncCorreclyAcknowledgesTheMessage(t *testing.T) {

	testcases := []struct {
		reject, requeue               bool // given
		isacked, isnacked, isrequeued bool // expected
	}{
		{false, false, true, false, false},
		{false, true, true, false, false},
		{true, false, false, true, false},
		{true, true, false, true, true},
	}

	for i, tc := range testcases {

		// given
		info := fmt.Sprintf("testcase %d, %+v", i, tc)
		mock := NewMockAcknowledger()
		ackFunc := createAcknowledgeFunc(tc.reject, tc.requeue)
		msg := rabtap.TapMessage{AmqpMessage: &amqp.Delivery{Acknowledger: mock}}

		// when
		err := ackFunc(msg)

		// then
		assert.NoError(t, err)
		assert.Equal(t, tc.isacked, mock.isAcked(), info)
		assert.Equal(t, tc.isnacked, mock.isNacked(), info)
		assert.Equal(t, tc.isrequeued, mock.isRequeued(), info)
	}
}

func TestCreateCountingMessageReceivePredReturnsTrueIfNumIsZero(t *testing.T) {
	pred := createCountingMessageReceivePred(0)
	res, err := pred(rabtap.TapMessage{})
	assert.NoError(t, err)
	assert.False(t, res)
}

func TestCreateCountingMessageReceivePredReturnsTrueOnNthCall(t *testing.T) {
	pred := createCountingMessageReceivePred(2)

	res, err := pred(rabtap.TapMessage{})
	assert.NoError(t, err)
	assert.False(t, res)
	res, err = pred(rabtap.TapMessage{})
	assert.NoError(t, err)
	assert.True(t, res)
}

func TestChainMessageReceiveFuncCallsBothFunctions(t *testing.T) {
	firstCalled := false
	secondCalled := false
	first := func(_ rabtap.TapMessage) error { firstCalled = true; return nil }
	second := func(_ rabtap.TapMessage) error { secondCalled = true; return nil }

	chained := chainedMessageReceiveFunc(first, second)
	err := chained(rabtap.TapMessage{})

	assert.Nil(t, err)
	assert.True(t, firstCalled)
	assert.True(t, secondCalled)
}

func TestChainMessageReceiveFuncDoesNotCallSecondOnErrorOnFirst(t *testing.T) {
	firstCalled := false
	secondCalled := false
	expectedErr := errors.New("first failed")
	first := func(_ rabtap.TapMessage) error { firstCalled = true; return expectedErr }
	second := func(_ rabtap.TapMessage) error { secondCalled = true; return nil }

	chained := chainedMessageReceiveFunc(first, second)
	err := chained(rabtap.TapMessage{})

	assert.Equal(t, expectedErr, err)
	assert.True(t, firstCalled)
	assert.False(t, secondCalled)
}

func TestCreateMessageReceiveFuncReturnsErrorWithInvalidFormat(t *testing.T) {
	opts := MessageReceiveFuncOptions{
		format: "invalid",
	}
	_, err := createMessageReceiveFunc(opts)
	assert.NotNil(t, err)
}

func TestCreateMessageReceiveFuncRawToFile(t *testing.T) {
	testDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(testDir)

	var b bytes.Buffer
	opts := MessageReceiveFuncOptions{
		out:              &b,
		format:           "raw",
		optSaveDir:       &testDir,
		silent:           false,
		filenameProvider: func() string { return "tapfilename" },
	}
	rcvFunc, err := createMessageReceiveFunc(opts)
	assert.Nil(t, err)
	message := rabtap.NewTapMessage(&amqp.Delivery{Body: []byte("Testmessage")}, time.Now())

	err = rcvFunc(message)
	assert.Nil(t, err)

	assert.True(t, strings.Contains(b.String(), "Testmessage"))

	// check contents of written file(s)
	contents, err := ioutil.ReadFile(path.Join(testDir, "tapfilename.dat"))
	assert.Nil(t, err)
	assert.Equal(t, "Testmessage", string(contents))

	// TODO check contents of JSON metadata "tapfilename.json"
	contents, err = ioutil.ReadFile(path.Join(testDir, "tapfilename.json"))
	assert.Nil(t, err)
	var metadata map[string]interface{}
	err = json.Unmarshal(contents, &metadata)
	assert.Nil(t, err)
}

func TestCreateMessageReceiveFuncPrintsNothingWhenSilentOptionIsSet(t *testing.T) {
	var b bytes.Buffer
	opts := MessageReceiveFuncOptions{
		out:        &b,
		format:     "raw",
		optSaveDir: nil,
		silent:     true,
	}
	rcvFunc, err := createMessageReceiveFunc(opts)
	assert.Nil(t, err)
	message := rabtap.NewTapMessage(&amqp.Delivery{Body: []byte("Testmessage")}, time.Now())

	err = rcvFunc(message)
	assert.Nil(t, err)

	assert.Equal(t, b.String(), "")
}

func TestCreateMessageReceiveFuncJSON(t *testing.T) {
	var b bytes.Buffer
	opts := MessageReceiveFuncOptions{
		out:              &b,
		format:           "json",
		optSaveDir:       nil,
		filenameProvider: func() string { return "tapfilename" },
	}
	rcvFunc, err := createMessageReceiveFunc(opts)
	assert.Nil(t, err)
	message := rabtap.NewTapMessage(&amqp.Delivery{Body: []byte("Testmessage")}, time.Now())

	err = rcvFunc(message)
	assert.Nil(t, err)

	assert.True(t, strings.Count(b.String(), "\n") > 1)
	assert.True(t, strings.Contains(b.String(), "\"Body\": \"VGVzdG1lc3NhZ2U=\""))

}

func TestCreateMessageReceiveFuncJSONNoPPToFile(t *testing.T) {
	// message is written as json (no pretty print) to writer and
	// as json to file.

	testDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(testDir)

	var b bytes.Buffer
	opts := MessageReceiveFuncOptions{
		out:              &b,
		format:           "json-nopp",
		optSaveDir:       &testDir,
		filenameProvider: func() string { return "tapfilename" },
	}
	rcvFunc, err := createMessageReceiveFunc(opts)
	assert.Nil(t, err)
	message := rabtap.NewTapMessage(&amqp.Delivery{Body: []byte("Testmessage")}, time.Now())

	err = rcvFunc(message)
	assert.Nil(t, err)

	assert.Equal(t, 1, strings.Count(b.String(), "\n"))
	assert.True(t, strings.Contains(b.String(), ",\"Body\":\"VGVzdG1lc3NhZ2U=\""))

	contents, err := ioutil.ReadFile(path.Join(testDir, "tapfilename.json"))
	assert.Nil(t, err)
	assert.True(t, strings.Count(string(contents), "\n") > 1)
	assert.True(t, strings.Contains(string(contents), "\"Body\": \"VGVzdG1lc3NhZ2U=\""))

}

func TestMessageReceiveLoopForwardsMessagesOnChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	messageChan := make(rabtap.TapChannel)
	errorChan := make(rabtap.SubscribeErrorChannel)
	done := make(chan bool)
	received := 0

	receiveFunc := func(rabtap.TapMessage) error {
		received++
		done <- true
		return nil
	}
	termPred := func(rabtap.TapMessage) (bool, error) { return false, nil }
	passPred := func(rabtap.TapMessage) (bool, error) { return true, nil }
	acknowledger := func(rabtap.TapMessage) error { return nil }
	go func() {
		_ = messageReceiveLoop(ctx, messageChan, errorChan, receiveFunc, passPred, termPred, acknowledger, time.Second*10)
	}()

	messageChan <- rabtap.TapMessage{}
	<-done // TODO add timeout
	cancel()
	assert.Equal(t, 1, received)
}

func TestMessageReceiveLoopExitsOnChannelClose(t *testing.T) {
	ctx := context.Background()
	messageChan := make(rabtap.TapChannel)
	errorChan := make(rabtap.SubscribeErrorChannel)
	termPred := func(rabtap.TapMessage) (bool, error) { return false, nil }
	passPred := func(rabtap.TapMessage) (bool, error) { return true, nil }

	close(messageChan)
	acknowledger := func(rabtap.TapMessage) error { return nil }
	err := messageReceiveLoop(ctx, messageChan, errorChan, NullMessageReceiveFunc, passPred, termPred, acknowledger, time.Second*10)

	assert.Nil(t, err)
}

func TestMessageReceiveLoopExitsWhenTermPredReturnsTrue(t *testing.T) {
	ctx := context.Background()
	messageChan := make(rabtap.TapChannel, 1)
	errorChan := make(rabtap.SubscribeErrorChannel)
	termPred := func(rabtap.TapMessage) (bool, error) { return true, nil }
	passPred := func(rabtap.TapMessage) (bool, error) { return true, nil }

	messageChan <- rabtap.TapMessage{}
	acknowledger := func(rabtap.TapMessage) error { return nil }
	err := messageReceiveLoop(ctx, messageChan, errorChan, NullMessageReceiveFunc, passPred, termPred, acknowledger, time.Second*10)

	assert.Nil(t, err)
}

func TestMessageReceiveLoopIgnoresFilteredMessages(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	messageChan := make(rabtap.TapChannel, 3)
	errorChan := make(rabtap.SubscribeErrorChannel)
	received := 0

	receiveFunc := func(rabtap.TapMessage) error {
		received++
		return nil
	}
	termPred := func(rabtap.TapMessage) (bool, error) { return false, nil }
	// create a message predicate that lets only pass message with MessageID set to "test"
	filterPred := func(m rabtap.TapMessage) (bool, error) { return m.AmqpMessage.MessageId == "test", nil }
	acknowledger := func(rabtap.TapMessage) error { return nil }

	// when we send 3 messages
	messageChan <- rabtap.TapMessage{AmqpMessage: &amqp.Delivery{MessageId: ""}}
	messageChan <- rabtap.TapMessage{AmqpMessage: &amqp.Delivery{MessageId: "test"}}
	messageChan <- rabtap.TapMessage{AmqpMessage: &amqp.Delivery{MessageId: ""}}

	_ = messageReceiveLoop(ctx, messageChan, errorChan, receiveFunc,
		filterPred, termPred, acknowledger, time.Second*1)

	// we expect 2 of them to be filtered out
	cancel()
	assert.Equal(t, 1, received)
}

func TestMessageReceiveLoopExitsWithErrorWhenIdle(t *testing.T) {
	// given
	ctx := context.Background()
	messageChan := make(rabtap.TapChannel)
	errorChan := make(rabtap.SubscribeErrorChannel)
	termPred := func(rabtap.TapMessage) (bool, error) { return false, nil }
	passPred := func(rabtap.TapMessage) (bool, error) { return true, nil }
	acknowledger := func(rabtap.TapMessage) error { return nil }

	// when
	err := messageReceiveLoop(ctx, messageChan, errorChan, NullMessageReceiveFunc, passPred, termPred, acknowledger, time.Second*1)

	// Then
	assert.Equal(t, ErrIdleTimeout, err)
}
