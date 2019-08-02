// Copyright (C) 2017 Jan Delgado

// +build integration

package main

// cmd_{exchangeCreate, sub, queueCreate, queueBind, queueDelete}
// integration test

import (
	"context"
	"crypto/tls"
	"io"
	"testing"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestCmdSubFailsEarlyWhenBrokerIsNotAvailable(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan bool)
	go func() {
		// we expect cmdSubscribe to return
		cmdSubscribe(ctx, CmdSubscribeArg{
			amqpURI:            "invalid uri",
			queue:              "queue",
			tlsConfig:          &tls.Config{},
			messageReceiveFunc: func(rabtap.TapMessage) error { return nil },
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
	const testKey = testQueue
	const testExchange = "sub-exchange-test"
	tlsConfig := &tls.Config{}
	amqpURI := testcommon.IntegrationURIFromEnv()

	done := make(chan bool)
	receiveFunc := func(message rabtap.TapMessage) error {
		log.Debug("test: received message: #+v", message)
		if string(message.AmqpMessage.Body) == testMessage {
			done <- true
		}
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	// creat exchange
	cmdExchangeCreate(CmdExchangeCreateArg{amqpURI: amqpURI,
		exchange: testExchange, exchangeType: "fanout",
		durable: false, tlsConfig: tlsConfig})
	defer cmdExchangeRemove(amqpURI, testExchange, tlsConfig)

	// create and bind queue
	cmdQueueCreate(CmdQueueCreateArg{amqpURI: amqpURI,
		queue: testQueue, tlsConfig: tlsConfig})
	cmdQueueBindToExchange(amqpURI, testQueue, testKey, testExchange, tlsConfig)
	defer cmdQueueRemove(amqpURI, testQueue, tlsConfig)

	// subscribe to testQueue
	go cmdSubscribe(ctx, CmdSubscribeArg{
		amqpURI:            amqpURI,
		queue:              testQueue,
		tlsConfig:          tlsConfig,
		messageReceiveFunc: receiveFunc})

	time.Sleep(time.Second * 1)

	messageCount := 0
	cmdPublish(CmdPublishArg{
		amqpURI:    amqpURI,
		exchange:   testExchange,
		routingKey: testKey,
		tlsConfig:  tlsConfig,
		readNextMessageFunc: func() (amqp.Publishing, bool, error) {
			// provide exactly one message
			if messageCount > 0 {
				return amqp.Publishing{}, false, io.EOF
			}
			messageCount++
			return amqp.Publishing{
				Body:         []byte(testMessage),
				ContentType:  "text/plain",
				DeliveryMode: amqp.Transient,
			}, true, nil
		}})

	// test if our tap received the message
	select {
	case <-done:
	case <-time.After(time.Second * 2):
		assert.Fail(t, "did not receive message within expected time")
	}
	cancel() // stop cmdSubscribe()

	cmdQueueUnbindFromExchange(amqpURI, testQueue, testKey, testExchange, tlsConfig)
	// TODO check that queue is unbound
}
