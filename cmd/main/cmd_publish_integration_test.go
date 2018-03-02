// Copyright (C) 2017 Jan Delgado

// +build integration

package main

import (
	"crypto/tls"
	"strings"
	"testing"
	"time"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmdPublishRaw(t *testing.T) {

	conn, ch := testcommon.IntegrationTestConnection(t, "exchange", "topic", 1, false)
	defer conn.Close()

	queueName := testcommon.IntegrationQueueName(0)
	routingKey := queueName

	deliveries, err := ch.Consume(
		queueName,
		"test-consumer",
		false, // noAck
		true,  // exclusive
		false, // noLocal
		false, // noWait
		nil,   // arguments
	)
	require.Nil(t, err)

	reader := strings.NewReader("hello")
	cmdPublish(CmdPublishArg{
		amqpURI:             testcommon.IntegrationURIFromEnv(),
		exchange:            "exchange",
		routingKey:          routingKey,
		tlsConfig:           &tls.Config{},
		readNextMessageFunc: createMessageReaderFunc(false, reader)})

	select {
	case message := <-deliveries:
		assert.Equal(t, "exchange", message.Exchange)
		assert.Equal(t, routingKey, message.RoutingKey)
		assert.Equal(t, "hello", string(message.Body))
	case <-time.After(time.Second * 2):
		assert.Fail(t, "did not receive message within expected time")
	}
}

func TestCmdPublishJSON(t *testing.T) {

	// note: base64dec("aGVsbG8=") == "hello"
	message := `
	{
	  "Headers": null,
	  "ContentType": "text/plain",
	  "ContentEncoding": "",
	  "DeliveryMode": 0,
	  "Priority": 0,
	  "CorrelationID": "",
	  "ReplyTo": "",
	  "Expiration": "",
	  "MessageID": "",
	  "Timestamp": "2017-10-28T23:45:33+02:00",
	  "Type": "",
	  "UserID": "",
	  "AppID": "rabtap.testgen",
	  "DeliveryTag": 63,
	  "Redelivered": false,
	  "Exchange": "amq.topic",
	  "RoutingKey": "test-q-amq.topic-0",
	  "Body": "aGVsbG8="
    }`
	conn, ch := testcommon.IntegrationTestConnection(t, "exchange", "topic", 1, false)
	defer conn.Close()

	queueName := testcommon.IntegrationQueueName(0)
	routingKey := queueName

	deliveries, err := ch.Consume(
		queueName,
		"test-consumer",
		false, // noAck
		true,  // exclusive
		false, // noLocal
		false, // noWait
		nil,   // arguments
	)
	require.Nil(t, err)

	reader := strings.NewReader(message)
	cmdPublish(CmdPublishArg{
		amqpURI:             testcommon.IntegrationURIFromEnv(),
		exchange:            "exchange",
		routingKey:          routingKey,
		tlsConfig:           &tls.Config{},
		readNextMessageFunc: createMessageReaderFunc(true, reader)})

	select {
	case message := <-deliveries:
		assert.Equal(t, "exchange", message.Exchange)
		assert.Equal(t, routingKey, message.RoutingKey)
		assert.Equal(t, "hello", string(message.Body))
	case <-time.After(time.Second * 2):
		assert.Fail(t, "did not receive message within expected time")
	}
}
