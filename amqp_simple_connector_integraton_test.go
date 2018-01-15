// Copyright (C) 2017 Jan Delgado
// +build integration

package rabtap

import (
	"crypto/tls"
	"testing"

	"github.com/jandelgado/rabtap/testhelper"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestSimpleAmqpConnector(t *testing.T) {
	called := false
	err := SimpleAmqpConnector(testhelper.IntegrationURIFromEnv(),
		&tls.Config{},
		func(chn *amqp.Channel) error {
			if chn != nil {
				called = true
			}
			return nil
		})
	assert.Nil(t, err)
	assert.True(t, called)
}

func TestSimpleAmqpConnectorWithError(t *testing.T) {
	called := false
	err := SimpleAmqpConnector("invalid_uri",
		&tls.Config{},
		func(_ *amqp.Channel) error {
			// should not be called.
			called = true
			return nil
		})
	assert.NotNil(t, err)
	assert.False(t, called)
}
