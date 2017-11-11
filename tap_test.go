// Copyright (C) 2017 Jan Delgado

package rabtap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTapQueueNameForExchange(t *testing.T) {

	assert.Equal(t, "__tap-queue-for-exchange-1234",
		getTapQueueNameForExchange("exchange", "1234"))
}

func TestGetTapEchangeNameForExchange(t *testing.T) {

	assert.Equal(t, "__tap-exchange-for-exchange-1234",
		getTapExchangeNameForExchange("exchange", "1234"))
}
