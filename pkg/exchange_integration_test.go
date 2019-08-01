// Copyright (C) 2017 Jan Delgado

// +build integration

package rabtap

// pubishing integration test functionality. assumes running rabbitmq broker on
// address defined by AMQP_URL and RABBIT_API_URL environment variables.
// (to start a local rabbitmq instance:
//  $ sudo  docker run --rm -ti -p5672:5672 rabbitmq:3-management)

import (
	"crypto/tls"
	"net/url"
	"testing"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
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

	const testName = "testexchange"

	url, _ := url.Parse(testcommon.IntegrationAPIURIFromEnv())
	client := NewRabbitHTTPClient(url, &tls.Config{})

	// make sure exchange does not exist before creation
	exchanges, err := client.Exchanges()
	assert.Nil(t, err)
	assert.Equal(t, -1, findExchange(testName, exchanges))

	// create exchange
	conn, ch := testcommon.IntegrationTestConnection(t, "", "", 0, false)
	defer conn.Close()
	err = CreateExchange(ch, testName, "topic", false, false)
	assert.Nil(t, err)

	// check if exchange was created
	exchanges, err = client.Exchanges()
	assert.Nil(t, err)
	assert.NotEqual(t, -1, findExchange(testName, exchanges))

	// finally remove exchange
	err = RemoveExchange(ch, testName, false)
	assert.Nil(t, err)

	// check if exchange was deleted
	exchanges, err = client.Exchanges()
	assert.Nil(t, err)
	assert.Equal(t, -1, findExchange(testName, exchanges))
}
