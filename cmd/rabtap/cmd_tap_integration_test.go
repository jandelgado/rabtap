// Copyright (C) 2017 Jan Delgado

// +build integration

package main

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/jandelgado/rabtap/pkg"
	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmdTap(t *testing.T) {

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
		{AmqpURI: testcommon.IntegrationURIFromEnv(),
			Exchanges: exchangeConfig}}

	ctx, cancel := context.WithCancel(context.Background())
	go cmdTap(ctx, tapConfig, &tls.Config{}, receiveFunc)

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
	require.Nil(t, err)

	// test if our tap received the message
	select {
	case <-done:
	case <-time.After(time.Second * 2):
		assert.Fail(t, "did not receive message within expected time")
	}
	cancel() // stop cmdTap()
}
