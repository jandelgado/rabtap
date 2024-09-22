// Copyright (C) 2017 Jan Delgado

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONFormatterInvalidArray(t *testing.T) {

	body := []byte("[ {\"a\":1} ")
	formattedMessage := JSONMessageFormatter{}.Format(body)
	// message is expected to be returned untouched
	assert.Equal(t, "[ {\"a\":1} ", formattedMessage)
}

func TestJSONFormatterValidArray(t *testing.T) {

	body := []byte(" [   {\"a\":1}    ] ")
	formattedMessage := JSONMessageFormatter{}.Format(body)
	assert.Equal(t, "[\n  {\n    \"a\": 1\n  }\n]", formattedMessage)
}

func TestJSONFormatterInvalidObject(t *testing.T) {

	body := []byte("[ {\"a\":1 ")
	formattedMessage := JSONMessageFormatter{}.Format(body)
	// message is expected to be returned untouched
	assert.Equal(t, "[ {\"a\":1 ", formattedMessage)
}

func TestJSONFormatterValidObject(t *testing.T) {

	body := []byte("  {\"a\":1}   ")
	formattedMessage := JSONMessageFormatter{}.Format(body)
	assert.Equal(t, "{\n  \"a\": 1\n}", formattedMessage)
}

func TestJSONFormatterEmptyValue(t *testing.T) {
	// An empty buffer effectively should be returned unmodified
	body := []byte("")
	formattedMessage := JSONMessageFormatter{}.Format(body)
	// message is expected to be returned untouched
	assert.Equal(t, "", formattedMessage)
}
