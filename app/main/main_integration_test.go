// Copyright (C) 2017 Jan Delgado

// +build integration

package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jandelgado/rabtap"
	"github.com/jandelgado/rabtap/testhelper"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExampleInfoMode() {
	// REST api mock returning only empty messages
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "[ ]")
	}
	apiMock := httptest.NewServer(http.HandlerFunc(handler))
	client := rabtap.NewRabbitHTTPClient(apiMock.URL, &tls.Config{})

	printBrokerInfoConfig := PrintBrokerInfoConfig{
		ShowStats:           false,
		ShowConsumers:       false,
		ShowDefaultExchange: false,
		NoColor:             true}

	startInfoMode("http://x:y@rootnode", client, printBrokerInfoConfig)

	// Output:
	// http://rootnode
}

func TestSendModeRaw(t *testing.T) {

	conn, ch := testhelper.IntegrationTestConnection(t, "exchange", "topic", 1, false)
	defer conn.Close()

	queueName := testhelper.IntegrationQueueName(0)
	bindingKey := queueName

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
	startSendMode(testhelper.IntegrationURIFromEnv(),
		"exchange", bindingKey, false, createMessageReaderFunc(false, reader))

	select {
	case message := <-deliveries:
		assert.Equal(t, "exchange", message.Exchange)
		assert.Equal(t, bindingKey, message.RoutingKey)
		assert.Equal(t, "hello", string(message.Body))
	case <-time.After(time.Second * 2):
		assert.Fail(t, "did not receive message within expected time")
	}
}

func TestSendModeJSON(t *testing.T) {

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
	conn, ch := testhelper.IntegrationTestConnection(t, "exchange", "topic", 1, false)
	defer conn.Close()

	queueName := testhelper.IntegrationQueueName(0)
	bindingKey := queueName

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
	startSendMode(testhelper.IntegrationURIFromEnv(),
		"exchange", bindingKey, false, createMessageReaderFunc(true, reader))

	select {
	case message := <-deliveries:
		assert.Equal(t, "exchange", message.Exchange)
		assert.Equal(t, bindingKey, message.RoutingKey)
		assert.Equal(t, "hello", string(message.Body))
	case <-time.After(time.Second * 2):
		assert.Fail(t, "did not receive message within expected time")
	}
}

func TestTapMode(t *testing.T) {

	conn, ch := testhelper.IntegrationTestConnection(t, "int-test-exchange", "topic", 1, false)
	require.NotNil(t, conn)
	require.NotNil(t, ch)
	defer conn.Close()

	// receiveFunc must receive messages passed through tapMessageChannel
	done := make(chan bool)
	receiveFunc := func(message *amqp.Delivery) error {
		log.Debug("received message on tap: #+v", message)
		if string(message.Body) == "Hello" {
			done <- true
		}
		return nil
	}

	exchangeConfig := []rabtap.ExchangeConfiguration{
		rabtap.ExchangeConfiguration{Exchange: "int-test-exchange",
			BindingKey: "my-routing-key"}}
	tapConfig := []rabtap.TapConfiguration{
		rabtap.TapConfiguration{AmqpURI: testhelper.IntegrationURIFromEnv(),
			Exchanges: exchangeConfig}}
	// signalChannel receives ctrl+C/interrput signal
	signalChannel := make(chan os.Signal, 1)
	go startTapMode(tapConfig, false, receiveFunc, signalChannel)

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
	signalChannel <- os.Interrupt
}
