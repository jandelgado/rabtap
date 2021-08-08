// Copyright (C) 2017 Jan Delgado
// +build integration

package rabtap

import (
	"crypto/tls"
	"net/url"
	"testing"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
)

func TestSimpleAmqpConnector(t *testing.T) {
	called := false
	err := SimpleAmqpConnector(testcommon.IntegrationURIFromEnv(),
		&tls.Config{},
		func(session Session) error {
			called = true
			return nil
		})
	assert.Nil(t, err)
	assert.True(t, called)
}

func TestSimpleAmqpConnectorFailsOnConnectionError(t *testing.T) {
	called := false

	url, _ := url.Parse("amqp://localhost:1")
	err := SimpleAmqpConnector(url,
		&tls.Config{},
		func(_ Session) error {
			// should not be called.
			called = true
			return nil
		})
	assert.NotNil(t, err)
	assert.False(t, called)
}
