// Copyright (C) 2017-2020 Jan Delgado

// +build integration

package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultDurationReturnsCorrectValue(t *testing.T) {
	assert.Equal(t, time.Duration(50), multDuration(time.Duration(100), 0.5))
}

func TestDurationBetweenMessagesReturnsZeroIfAnyOfTheArgumentsIsNil(t *testing.T) {
	msg := RabtapPersistentMessage{XRabtapReceivedTimestamp: time.Now()}
	fixed := time.Duration(123)
	for _, delay := range []*time.Duration{nil, &fixed} {
		assert.Equal(t, time.Duration(0), durationBetweenMessages(&msg, nil, 1., delay))
		assert.Equal(t, time.Duration(0), durationBetweenMessages(nil, &msg, 1., delay))
		assert.Equal(t, time.Duration(0), durationBetweenMessages(nil, nil, 1., delay))
	}
}

func TestDurationBetweenMessagesReturnsFixedDurationIfSet(t *testing.T) {
	msg := RabtapPersistentMessage{XRabtapReceivedTimestamp: time.Now()}
	fixed := time.Duration(123)
	assert.Equal(t, time.Duration(123), durationBetweenMessages(&msg, &msg, 1., &fixed))
}

func TestDurationBetweenMessagesCorrectlyScalesDuration(t *testing.T) {
	first := time.Unix(0, 0)
	second := time.Unix(0, 1000)
	assert.Equal(t, time.Duration(500), durationBetweenMessages(
		&RabtapPersistentMessage{XRabtapReceivedTimestamp: first},
		&RabtapPersistentMessage{XRabtapReceivedTimestamp: second},
		.5, nil))
}

func TestDurationBetweenMessagesIsCalculatedCorrectly(t *testing.T) {
	first := time.Unix(0, 0)
	second := time.Unix(0, 1000)
	assert.Equal(t, time.Duration(1000), durationBetweenMessages(
		&RabtapPersistentMessage{XRabtapReceivedTimestamp: first},
		&RabtapPersistentMessage{XRabtapReceivedTimestamp: second},
		1., nil))
}

func TestSelectOptionOrDefaultReturnsOptionalIfSet(t *testing.T) {
	opt := "optional"
	assert.Equal(t, "optional", selectOptionalOrDefault(&opt, "default"))
}

func TestSelectOptionOrDefaultReturnsDefaultIfOptionalIsNil(t *testing.T) {
	assert.Equal(t, "default", selectOptionalOrDefault(nil, "default"))
}

func TestPublishMessageStreamPublishesNextMessage(t *testing.T) {
	mockReader := func() (RabtapPersistentMessage, bool, error) {
		return RabtapPersistentMessage{Body: []byte("hello")}, false, nil
	}
	delayer := func(first, second *RabtapPersistentMessage) {}

	pubCh := make(rabtap.PublishChannel, 1)
	exchange := "exchange"
	key := "key"
	err := publishMessageStream(pubCh, &exchange, &key, map[string]string{}, mockReader, delayer)

	assert.Nil(t, err)
	select {
	case message := <-pubCh:
		assert.Equal(t, "exchange", message.Routing.Exchange())
		assert.Equal(t, "key", message.Routing.Key())
		assert.Equal(t, "hello", string(message.Publishing.Body))
	case <-time.After(time.Second * 2):
		assert.Fail(t, "did not receive message within expected time")
	}
	// expect channel to be closed now
	select {
	case _, more := <-pubCh:
		assert.False(t, more)
	case <-time.After(time.Second * 2):
		assert.Fail(t, "did not receive message within expected time")
	}
}

func TestPublishMessageStreamPropagatesMessageReadError(t *testing.T) {
	mockReader := func() (RabtapPersistentMessage, bool, error) {
		return RabtapPersistentMessage{}, false, errors.New("error")
	}
	delayer := func(first, second *RabtapPersistentMessage) {}

	pubCh := make(rabtap.PublishChannel)
	exchange := ""
	key := "key"
	err := publishMessageStream(pubCh, &exchange, &key, map[string]string{}, mockReader, delayer)
	assert.Equal(t, errors.New("error"), err)
}

func TestCmdPublishARawFileWithExchangeAndRoutingKey(t *testing.T) {
	// integrative test publishing a raw file

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
		"--uri", testcommon.IntegrationURIFromEnv().String(),
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

func TestCmdPublishAJSONFileWithIncludedRoutingKeyAndExchange(t *testing.T) {

	conn, ch := testcommon.IntegrationTestConnection(t, "myexchange", "topic", 1, false)
	defer conn.Close()

	queueName := testcommon.IntegrationQueueName(0)
	routingKey := queueName

	// in the integrative test we send a stream of 2 messages.
	// note: base64dec("aGVsbG8=") == "hello"
	//        base64dec("c2Vjb25kCg==") == "second\n"
	testmessages := `
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
	  "Exchange": "myexchange",
	  "RoutingKey": "` + routingKey + `",
	  "Body": "aGVsbG8=",
	  "XRabtapReceivedTimestamp": "2017-10-28T23:45:33+02:00"
    }
	{
	  "Exchange": "myexchange",
	  "RoutingKey": "` + routingKey + `",
      "Body": "c2Vjb25kCg==",
	  "XRabtapReceivedTimestamp": "2017-10-28T23:45:34+02:00"
	}
	{
      "Body": "c2Vjb25kCg==",
	  "XRabtapReceivedTimestamp": "2017-10-28T23:45:35+02:00"
	}`

	tmpfile, err := ioutil.TempFile("", "rabtap")
	require.Nil(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(testmessages))
	require.Nil(t, err)

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
	// message is in tmpfile.Name(). We expect exchange and routingkey to
	// be taken from the JSON metadata as they are not specified.
	os.Args = []string{"rabtap", "pub",
		"--uri", testcommon.IntegrationURIFromEnv().String(),
		tmpfile.Name(),
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

	assert.Equal(t, "myexchange", message[0].Exchange)
	assert.Equal(t, routingKey, message[0].RoutingKey)
	assert.Equal(t, "hello", string(message[0].Body))

	assert.Equal(t, "myexchange", message[1].Exchange)
	assert.Equal(t, routingKey, message[1].RoutingKey)
	assert.Equal(t, "second\n", string(message[1].Body))
}

func TestCmdPublishFilesFromDirectory(t *testing.T) {
	// publish message stored in a directory as separate files (json-metadata
	// and raw message file)

	conn, ch := testcommon.IntegrationTestConnection(t, "myexchange", "topic", 1, false)
	defer conn.Close()

	queueName := testcommon.IntegrationQueueName(0)
	routingKey := queueName

	msg := `{ "Exchange": "myexchange", "RoutingKey": "` + routingKey + `", "Body": "ixxx" }`

	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	err = ioutil.WriteFile(filepath.Join(dir, "rabtap-1.json"), []byte(msg), 0666)
	require.Nil(t, err)
	err = ioutil.WriteFile(filepath.Join(dir, "rabtap-1.dat"), []byte("Hello123"), 0666)
	require.Nil(t, err)

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
	// message is read from the provided directory. We expect exchange and
	// routingkey to be taken from the JSON metadata as they are not specified.
	os.Args = []string{"rabtap", "pub",
		"--uri", testcommon.IntegrationURIFromEnv().String(),
		dir,
		"--format=raw"}
	main()

	// verification: we expect 1 messages to be sent by above call
	var message [1]amqp.Delivery
	for i := 0; i < 1; i++ {
		select {
		case message[i] = <-deliveries:
		case <-time.After(time.Second * 2):
			assert.Fail(t, "did not receive message within expected time")
		}
	}

	assert.Equal(t, "myexchange", message[0].Exchange)
	assert.Equal(t, routingKey, message[0].RoutingKey)
	assert.Equal(t, "Hello123", string(message[0].Body))
}
