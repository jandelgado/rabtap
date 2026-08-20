// Copyright (C) 2017 Jan Delgado

//go:build integration

package rabtap

// pubishing integration test functionality. assumes running rabbitmq broker on
// address defined by AMQP_URL and RABBIT_API_URL environment variables.
// (to start a local rabbitmq instance:
//  $ sudo  docker run --rm -ti -p5672:5672 rabbitmq:3-management)

import (
	"context"
	"crypto/tls"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jandelgado/rabtap/pkg/testcommon"
)

func findExchange(exchange string, exchanges []RabbitExchange) int {
	for i, exc := range exchanges {
		if exc.Name == exchange && exc.Vhost == "/" {
			return i
		}
	}
	return -1
}

func TestIntegrationAmqpExchangeCreateRemove(t *testing.T) {
	// since in order to remove an exchange we must create it first, we
	// tests both functions together in one test case.

	const testName = "rabtaptestexchange"

	url, _ := url.Parse(testcommon.IntegrationAPIURIFromEnv())
	client := NewRabbitHTTPClient(url, &tls.Config{})

	// make sure exchange does not exist before creation
	exchanges, err := client.Exchanges(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, -1, findExchange(testName, exchanges))

	// create exchange
	setup, err := testcommon.IntegrationTestConnection("", "", 0, false)
	assert.NoError(t, err)
	session := Session{setup.Conn, setup.Chan}
	defer func() { _ = setup.Conn.Close() }()
	err = CreateExchange(session, testName, "topic", false, false, nil)
	assert.NoError(t, err)

	// check if exchange was created
	exchanges, err = client.Exchanges(context.TODO())
	assert.NoError(t, err)
	assert.NotEqual(t, -1, findExchange(testName, exchanges))

	// finally remove exchange
	err = RemoveExchange(session, testName, false)
	assert.NoError(t, err)

	// check if exchange was deleted
	exchanges, err = client.Exchanges(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, -1, findExchange(testName, exchanges))
}
