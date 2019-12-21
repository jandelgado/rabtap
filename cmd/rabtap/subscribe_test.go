// Copyright (C) 2017 Jan Delgado

package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateMessageReceiveFuncReturnsErrorWithInvalidFormat(t *testing.T) {
	testDir := "test"

	var b bytes.Buffer
	opts := MessageReceiveFuncOptions{
		out:        &b,
		format:     "invalid",
		optSaveDir: &testDir,
		noColor:    false,
		silent:     false,
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
		out:        &b,
		format:     "raw",
		optSaveDir: &testDir,
		noColor:    false,
		silent:     false,
	}
	rcvFunc, err := createMessageReceiveFunc(opts)
	assert.Nil(t, err)
	message := rabtap.NewTapMessage(&amqp.Delivery{Body: []byte("Testmessage")}, time.Now())

	err = rcvFunc(message)
	assert.Nil(t, err)

	assert.True(t, strings.Contains(b.String(), "Testmessage"))

	// TODO make contents of created filename predicatable (Timestamp, Name)
	//      and check written file
}

func TestCreateMessageReceiveFuncWritesNothingWhenSilentOptionIsSet(t *testing.T) {
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

func TestCreateMessageReceiveFuncJSONToFile(t *testing.T) {
	testDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(testDir)

	var b bytes.Buffer
	opts := MessageReceiveFuncOptions{
		out:        &b,
		format:     "json",
		optSaveDir: &testDir,
		noColor:    false,
		silent:     false,
	}
	rcvFunc, err := createMessageReceiveFunc(opts)
	assert.Nil(t, err)
	message := rabtap.NewTapMessage(&amqp.Delivery{Body: []byte("Testmessage")}, time.Now())

	err = rcvFunc(message)
	assert.Nil(t, err)

	assert.True(t, strings.Contains(b.String(), "\"Body\": \"VGVzdG1lc3NhZ2U=\""))

	// TODO make contents of created filename predicatable (Timestamp, Name)
	//      and check written file
}

func TestMessageReceiveLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	messageChan := make(rabtap.TapChannel)
	done := make(chan bool)
	received := 0

	receiveFunc := func(rabtap.TapMessage) error {
		received++
		done <- true
		return nil
	}
	go func() { _ = messageReceiveLoop(ctx, messageChan, receiveFunc) }()

	messageChan <- rabtap.TapMessage{}
	<-done // TODO add timeout
	cancel()
	assert.Equal(t, 1, received)
}
