// Copyright (C) 2017 Jan Delgado

//go:build integration
// +build integration

package main

import (
	"context"
	"crypto/tls"
	"os"
	"syscall"
	"testing"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/jandelgado/rabtap/pkg/testcommon"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmdTap(t *testing.T) {

	// given
	conn, ch := testcommon.IntegrationTestConnection(t, "int-test-exchange", "topic", 1, false)
	defer conn.Close()

	// receiveFunc must receive messages passed through tapMessageChannel
	done := make(chan bool)
	receiveFunc := func(message rabtap.TapMessage) error {
		log.Debug("received message on tap: #+v", message)
		if string(message.AmqpMessage.Body) == "Hello" {
			done <- true
		}
		return nil
	}

	exchangeConfig := []rabtap.ExchangeConfiguration{
		{Exchange: "int-test-exchange",
			BindingKey: "my-routing-key"}}
	tapConfig := []rabtap.TapConfiguration{
		{AMQPURL: testcommon.IntegrationURIFromEnv(),
			Exchanges: exchangeConfig}}

	ctx, cancel := context.WithCancel(context.Background())

	filterPred := func(rabtap.TapMessage) (bool, error) { return true, nil }
	termPred := func(rabtap.TapMessage) (bool, error) { return true, nil }

	// when
	go cmdTap(ctx, CmdTapArg{tapConfig: tapConfig,
		tlsConfig:          &tls.Config{},
		messageReceiveFunc: receiveFunc,
		filterPred:         filterPred,
		termPred:           termPred,
		timeout:            time.Second * 10})

	time.Sleep(time.Second * 1)
	err := ch.Publish(
		"int-test-exchange",
		"my-routing-key",
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Body:         []byte("Hello"),
			ContentType:  "text/plain",
			DeliveryMode: amqp.Transient,
		})

	// then: our tap received the message
	require.Nil(t, err)
	select {
	case <-done:
	case <-time.After(time.Second * 2):
		assert.Fail(t, "did not receive message within expected time")
	}
	cancel() // stop cmdTap()
}

func TestCmdTapIntegration(t *testing.T) {
	const testMessage = "TapHello"
	const testQueue = "tap-queue-test"
	testKey := testQueue
	testExchange := "amq.topic"

	go func() {
		time.Sleep(time.Second * 1)
		_, ch := testcommon.IntegrationTestConnection(t, "", "", 0, false)
		err := ch.Publish(
			testExchange,
			testKey,
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				Body:         []byte("Hello"),
				ContentType:  "text/plain",
				DeliveryMode: amqp.Transient,
				Headers:      amqp.Table{},
			})
		require.Nil(t, err)
		time.Sleep(time.Second * 1)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"rabtap", "tap",
		"--uri", testcommon.IntegrationURIFromEnv().String(),
		"amq.topic:" + testKey,
		"--format=raw",
		"--no-color"}
	output := testcommon.CaptureOutput(main)
	assert.Regexp(t, "(?s).*message received.*\nroutingkey.....: tap-queue-test\n.*Hello", output)
}
