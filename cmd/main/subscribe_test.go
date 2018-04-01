// Copyright (C) 2017 Jan Delgado

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

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
	_ = rcvFunc(&amqp.Delivery{Body: []byte("Testmessage")})

	assert.True(t, strings.Contains(b.String(), "Testmessage"))

	// TODO make created filename predicatable and check written file
}

func TestCreateMessageReceiveFuncJSONToFile(t *testing.T) {
	testDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(testDir)

	var b bytes.Buffer
	rcvFunc := createMessageReceiveFunc(&b, true, &testDir, false)
	_ = rcvFunc(&amqp.Delivery{Body: []byte("Testmessage")})

	assert.True(t, strings.Contains(b.String(), "\"Body\": \"VGVzdG1lc3NhZ2U=\""))

	// TODO make created filename predicatable and check written file
}

func TestMessageReceiveLoop(t *testing.T) {
	messageChan := make(rabtap.TapChannel)
	signalChannel := make(chan os.Signal)
	done := make(chan bool)
	received := 0

	receiveFunc := func(*amqp.Delivery) error {
		received = received + 1
		done <- true
		return nil
	}
	go messageReceiveLoop(messageChan, receiveFunc, signalChannel)

	messageChan <- &rabtap.TapMessage{}
	<-done                        // TODO add timeout
	signalChannel <- os.Interrupt // terminates go routine
	assert.Equal(t, 1, received)
}
