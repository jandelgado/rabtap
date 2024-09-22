// Copyright (C) 2017-2019 Jan Delgado

package main

import (
	"os"
	"testing"
	"time"

	"github.com/fatih/color"
	rabtap "github.com/jandelgado/rabtap/pkg"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestNewMessageFormatter(t *testing.T) {

	assert.Equal(t, JSONMessageFormatter{},
		NewMessageFormatter("application/json"))
	assert.Equal(t, DefaultMessageFormatter{},
		NewMessageFormatter("unknown"))
}

func ExamplePrettyPrintMessage() {

	message := amqp.Delivery{
		Exchange:        "exchange",
		RoutingKey:      "routingkey",
		Priority:        99,
		Expiration:      "2017-05-22 17:00:00",
		ContentType:     "plain/text",
		ContentEncoding: "identity",
		MessageId:       "4711",
		Timestamp:       time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
		Type:            "some type",
		CorrelationId:   "4712",
		Headers:         amqp.Table{"header": "value"},
		UserId:          "jan",
		AppId:           "rabtap",
		ReplyTo:         "message123",
		Body:            []byte("simple test message"),
	}

	ts := time.Date(2019, time.June, 6, 23, 0, 0, 0, time.UTC)
	color.NoColor = true // disable colors for test
	_ = PrettyPrintMessage(os.Stdout, rabtap.NewTapMessage(&message, ts))

	// Output:
	// ------ message received on 2019-06-06T23:00:00Z ------
	// exchange.......: exchange
	// routingkey.....: routingkey
	// priority.......: 99
	// expiration.....: 2017-05-22 17:00:00
	// content-type...: plain/text
	// content-enc....: identity
	// app-message-id.: 4711
	// app-timestamp..: 2009-11-10 23:00:00 +0000 UTC
	// app-type.......: some type
	// app-corr-id....: 4712
	// reply-to.......: message123
	// app-id.........: rabtap
	// user-id........: jan
	// app-headers....: map[header:value]
	// simple test message
	//
}

func ExamplePrettyPrintMessage_withFilteredAtributes() {

	message := amqp.Delivery{
		Exchange: "exchange",
		Body:     []byte("simple test message"),
	}

	color.NoColor = true
	ts := time.Date(2019, time.June, 6, 23, 0, 0, 0, time.UTC)
	_ = PrettyPrintMessage(os.Stdout, rabtap.NewTapMessage(&message, ts))

	// Output:
	// ------ message received on 2019-06-06T23:00:00Z ------
	// exchange.......: exchange
	// simple test message
	//
}
