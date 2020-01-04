// Copyright (C) 2017 Jan Delgado

// +build integration

package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmdPublishRaw(t *testing.T) {

	tmpfile, err := ioutil.TempFile("", "rabtap")
	require.Nil(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte("hello"))
	require.Nil(t, err)

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

	// execution: run publish command through call of main(), the actual
	// message is in tmpfile.Name()
	os.Args = []string{"rabtap", "pub",
		"--uri", testcommon.IntegrationURIFromEnv(),
		"--exchange=exchange",
		tmpfile.Name(),
		"--routingkey", routingKey}

	main()

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

	// in the test we send a stream of 2 messages.
	// note: base64dec("aGVsbG8=") == "hello"
	//        base64dec("c2Vjb25kCg==") == "second\n"
	testmessage := `
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
    }
	{
		"Body": "c2Vjb25kCg=="
	}`

	tmpfile, err := ioutil.TempFile("", "rabtap")
	require.Nil(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(testmessage))
	require.Nil(t, err)

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

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// execution: run publish command through call of main(), the actual
	// message is in tmpfile.Name()
	os.Args = []string{"rabtap", "pub",
		"--uri", testcommon.IntegrationURIFromEnv(),
		"--exchange=exchange",
		tmpfile.Name(),
		"--routingkey", routingKey,
		"--format=json"}
	main()

	// verification: we expect 2 messages to be sent by above call
	var message [2]amqp.Delivery
	for i := 0; i < 2; i++ {
		select {
		case message[i] = <-deliveries:
		case <-time.After(time.Second * 2):
			assert.Fail(t, "did not receive message within expected time")
		}
	}

	assert.Equal(t, "exchange", message[0].Exchange)
	assert.Equal(t, routingKey, message[0].RoutingKey)
	assert.Equal(t, "hello", string(message[0].Body))

	assert.Equal(t, "exchange", message[1].Exchange)
	assert.Equal(t, routingKey, message[1].RoutingKey)
	assert.Equal(t, "second\n", string(message[1].Body))
}
