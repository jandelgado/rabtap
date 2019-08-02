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

func TestCreateMessageReceiveFuncRawToFile(t *testing.T) {
	testDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(testDir)

	var b bytes.Buffer
	rcvFunc := createMessageReceiveFunc(&b, false, &testDir, false)
	message := rabtap.NewTapMessage(&amqp.Delivery{Body: []byte("Testmessage")}, nil, time.Now())

	_ = rcvFunc(message)

	assert.True(t, strings.Contains(b.String(), "Testmessage"))

	// TODO make created filename predicatable and check written file
}

func TestCreateMessageReceiveFuncJSONToFile(t *testing.T) {
	testDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(testDir)

	var b bytes.Buffer
	rcvFunc := createMessageReceiveFunc(&b, true, &testDir, false)
	message := rabtap.NewTapMessage(&amqp.Delivery{Body: []byte("Testmessage")}, nil, time.Now())

	_ = rcvFunc(message)

	assert.True(t, strings.Contains(b.String(), "\"Body\": \"VGVzdG1lc3NhZ2U=\""))

	// TODO make created filename predicatable and check written file
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
	go messageReceiveLoop(ctx, messageChan, receiveFunc)

	messageChan <- rabtap.TapMessage{}
	<-done // TODO add timeout
	cancel()
	assert.Equal(t, 1, received)
}
