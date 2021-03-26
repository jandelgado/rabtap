// test for the message subscriber
// TODO cleaner separation between component and unit test
// Copyright (C) 2017-2019 Jan Delgado

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCountingMessageReceivePredReturnsFalseAfterCalledNumTimes(t *testing.T) {
	pred := createCountingMessageReceivePred(2)

	assert.True(t, pred(rabtap.TapMessage{}))
	assert.True(t, pred(rabtap.TapMessage{}))
	assert.False(t, pred(rabtap.TapMessage{}))
}

func TestContinueMessageReceivePredReturnsTrue(t *testing.T) {
	assert.True(t, continueMessageReceivePred(rabtap.TapMessage{}))
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
		noColor:          false,
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
		noColor:    false,
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
	done := make(chan bool)
	received := 0

	receiveFunc := func(rabtap.TapMessage) error {
		received++
		done <- true
		return nil
	}
	continuePred := func(rabtap.TapMessage) bool { return true }
	go func() { _ = messageReceiveLoop(ctx, messageChan, receiveFunc, continuePred) }()

	messageChan <- rabtap.TapMessage{}
	<-done // TODO add timeout
	cancel()
	assert.Equal(t, 1, received)
}

func TestMessageReceiveLoopExitsOnChannelClose(t *testing.T) {
	ctx := context.Background()
	messageChan := make(rabtap.TapChannel)
	continuePred := func(rabtap.TapMessage) bool { return true }

	close(messageChan)
	err := messageReceiveLoop(ctx, messageChan, NullMessageReceiveFunc, continuePred)

	assert.Nil(t, err)
}

func TestMessageReceiveLoopExitsWhenLoopPredReturnsFalse(t *testing.T) {
	ctx := context.Background()
	messageChan := make(rabtap.TapChannel, 1)
	stopPred := func(rabtap.TapMessage) bool { return false }

	messageChan <- rabtap.TapMessage{}
	err := messageReceiveLoop(ctx, messageChan, NullMessageReceiveFunc, stopPred)

	assert.Nil(t, err)
}
