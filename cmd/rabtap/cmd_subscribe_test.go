// Copyright (C) 2017 Jan Delgado

//go:build integration
// +build integration

package main

// cmd_{exchangeCreate, sub, queueCreate, queueBind, queueDelete}
// integration test

import (
	"context"
	"crypto/tls"
	"io"
	"net/url"
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

func TestCmdSubFailsEarlyWhenBrokerIsNotAvailable(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan bool)
	amqpURL, _ := url.Parse("amqp://invalid.url:5672/")
	go func() {
		// we expect cmdSubscribe to return
		cmdSubscribe(ctx, CmdSubscribeArg{
			amqpURL:                amqpURL,
			queue:                  "queue",
			tlsConfig:              &tls.Config{},
			messageReceiveFunc:     func(rabtap.TapMessage) error { return nil },
			messageReceiveLoopPred: func(rabtap.TapMessage) (bool, error) { return false, nil },
			timeout:                time.Second * 10,
		})
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(time.Second * 2):
		assert.Fail(t, "cmdSubscribe did not fail on initial connection error")
	}
	cancel()
}

func TestCmdSub(t *testing.T) {
	const testMessage = "SubHello"
	const testQueue = "sub-queue-test"
	testKey := testQueue

	testExchange := ""
	//	testExchange := "sub-exchange-test"
	tlsConfig := &tls.Config{}
	amqpURL := testcommon.IntegrationURIFromEnv()

	done := make(chan bool)
	receiveFunc := func(message rabtap.TapMessage) error {
		log.Debug("test: received message: #+v", message)
		if string(message.AmqpMessage.Body) == testMessage {
			done <- true
		}
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	// create and bind queue
	cmdQueueCreate(CmdQueueCreateArg{amqpURL: amqpURL,
		queue: testQueue, tlsConfig: tlsConfig})
	defer cmdQueueRemove(amqpURL, testQueue, tlsConfig)

	// subscribe to testQueue
	go cmdSubscribe(ctx, CmdSubscribeArg{
		amqpURL:                amqpURL,
		queue:                  testQueue,
		tlsConfig:              tlsConfig,
		messageReceiveFunc:     receiveFunc,
		filterPred:             func(rabtap.TapMessage) (bool, error) { return true, nil },
		messageReceiveLoopPred: func(rabtap.TapMessage) (bool, error) { return false, nil },
		timeout:                time.Second * 10,
	})

	time.Sleep(time.Second * 1)

	messageCount := 0

	// TODO test without cmdPublish
	cmdPublish(
		ctx,
		CmdPublishArg{
			amqpURL:    amqpURL,
			exchange:   &testExchange,
			routingKey: &testKey,
			headers:    rabtap.KeyValueMap{},
			tlsConfig:  tlsConfig,
			providerFunc: func() (RabtapPersistentMessage, error) {
				// provide exactly one message
				if messageCount > 0 {
					return RabtapPersistentMessage{}, io.EOF
				}
				messageCount++
				return RabtapPersistentMessage{
					Body:         []byte(testMessage),
					ContentType:  "text/plain",
					DeliveryMode: amqp.Transient,
				}, nil
			}})

	// test if we received the message
	select {
	case <-done:
	case <-time.After(time.Second * 2):
		assert.Fail(t, "did not receive message within expected time")
	}
	cancel() // stop cmdSubscribe()
}

func TestCmdSubIntegration(t *testing.T) {
	const testMessage = "SubHello"
	const testQueue = "sub-queue-test"
	testKey := testQueue
	testExchange := "" // default exchange

	tlsConfig := &tls.Config{}
	amqpURL := testcommon.IntegrationURIFromEnv()

	cmdQueueCreate(CmdQueueCreateArg{amqpURL: amqpURL,
		queue: testQueue, tlsConfig: tlsConfig})
	defer cmdQueueRemove(amqpURL, testQueue, tlsConfig)

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

	go func() {
		time.Sleep(time.Second * 2)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"rabtap", "sub",
		"--uri", amqpURL.String(),
		testQueue,
		"--format=raw",
		"--no-color"}
	output := testcommon.CaptureOutput(main)

	assert.Regexp(t, "(?s).*message received.*\nroutingkey.....: sub-queue-test\n.*Hello", output)
}
