// Copyright (C) 2017 Jan Delgado

package main

import (
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestJSONFormatterInvalidArray(t *testing.T) {

	message := amqp.Delivery{
		Body: []byte("[ {\"a\":1} "),
	}
	formattedMessage := JSONMessageFormatter{}.Format(&message)
	// message is expected to be returned untouched
	assert.Equal(t, "[ {\"a\":1} ", formattedMessage)
}

func TestJSONFormatterValidArray(t *testing.T) {

	message := amqp.Delivery{
		Body: []byte(" [   {\"a\":1}    ] "),
	}
	formattedMessage := JSONMessageFormatter{}.Format(&message)
	assert.Equal(t, "[\n  {\n    \"a\": 1\n  }\n]", formattedMessage)
}

func TestJSONFormatterInvalidObject(t *testing.T) {

	message := amqp.Delivery{
		Body: []byte("[ {\"a\":1 "),
	}
	formattedMessage := JSONMessageFormatter{}.Format(&message)
	// message is expected to be returned untouched
	assert.Equal(t, "[ {\"a\":1 ", formattedMessage)
}

func TestJSONFormatterValidObject(t *testing.T) {

	message := amqp.Delivery{
		Body: []byte("  {\"a\":1}   "),
	}
	formattedMessage := JSONMessageFormatter{}.Format(&message)
	assert.Equal(t, "{\n  \"a\": 1\n}", formattedMessage)
}
