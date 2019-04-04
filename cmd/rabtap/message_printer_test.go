// Copyright (C) 2017 Jan Delgado

package main

import (
	"os"
	"testing"
	"time"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestNewMessageFormatter(t *testing.T) {

	assert.Equal(t, JSONMessageFormatter{},
		NewMessageFormatter(&amqp.Delivery{ContentType: "application/json"}))
	assert.Equal(t, DefaultMessageFormatter{},
		NewMessageFormatter(&amqp.Delivery{ContentType: "unknown"}))
}

func ExamplePrettyPrintMessage() {

	message := amqp.Delivery{
		Exchange:        "exchange",
		RoutingKey:      "routingkey",
		Priority:        99,
		Expiration:      "2017-05-22 17:00:00",
		ContentType:     "plain/text",
		ContentEncoding: "utf-8",
		MessageId:       "4711",
		Timestamp:       time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
		Type:            "some type",
		CorrelationId:   "4712",
		Headers:         amqp.Table{"header": "value"},
		Body:            []byte("simple test message"),
	}

	_ = PrettyPrintMessage(os.Stdout, &message, "title", true)

	// Output:
	// ------ title ------
	// exchange.......: exchange
	// routingkey.....: routingkey
	// priority.......: 99
	// expiration.....: 2017-05-22 17:00:00
	// content-type...: plain/text
	// content-enc....: utf-8
	// app-message-id.: 4711
	// app-timestamp..: 2009-11-10 23:00:00 +0000 UTC
	// app-type.......: some type
	// app-corr-id....: 4712
	// app-headers....: map[header:value]
	// simple test message
	//
}

func ExamplePrettyPrintMessage_withFilteredAtributes() {

	message := amqp.Delivery{
		Exchange: "exchange",
		Body:     []byte("simple test message"),
	}

	_ = PrettyPrintMessage(os.Stdout, &message, "title", true)

	// Output:
	// ------ title ------
	// exchange.......: exchange
	// simple test message
	//
}
