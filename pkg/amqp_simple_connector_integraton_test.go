// Copyright (C) 2017 Jan Delgado
// +build integration

package rabtap

import (
	"crypto/tls"
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

func TestSimpleAmqpConnectorWithError(t *testing.T) {
	called := false
	err := SimpleAmqpConnector("invalid_uri",
		&tls.Config{},
		func(_ Session) error {
			// should not be called.
			called = true
			return nil
		})
	assert.NotNil(t, err)
	assert.False(t, called)
}
