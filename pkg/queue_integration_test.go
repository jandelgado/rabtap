// Copyright (C) 2017 Jan Delgado

// +build integration

package rabtap

// queue integration test functionality. assumes running rabbitmq broker on
// address defined by AMQP_URL and RABBIT_API_URL environment variables.
// (to start a local rabbitmq instance:
//  $ sudo  docker run --rm -ti -p5672:5672 rabbitmq:3-management)

import (
	"crypto/tls"
	"testing"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
)

func findQueue(queue string, queues []RabbitQueue) int {
	for i, q := range queues {
		if q.Name == queue && q.Vhost == "/" {
			return i
		}
	}
	return -1
}

func findBinding(queue, exchange, key string, bindings []RabbitBinding) int {
	for i, b := range bindings {
		if b.Source == exchange &&
			b.Destination == queue &&
			b.RoutingKey == key {
			return i
		}
	}
	return -1
}

func TestIntegrationAmqpQueueCreateBindUnbindAndRemove(t *testing.T) {

	// since in order to remove and unbind a  queue we must create it first, we
	// tests these functions together in one test case.

	const queueTestName = "testqueue"
	const exchangeTestName = "amq.direct"
	const keyTestName = "key"

	client := NewRabbitHTTPClient(testcommon.IntegrationAPIURIFromEnv(),
		&tls.Config{})

	// make sure queue does not exist before creation
	queues, err := client.Queues()
	assert.Nil(t, err)
	assert.Equal(t, -1, findQueue(queueTestName, queues))

	// create queue
	conn, ch := testcommon.IntegrationTestConnection(t, "", "", 0, false)
	defer conn.Close()
	err = CreateQueue(ch, queueTestName, false, false, false)
	assert.Nil(t, err)

	// check if queue was created
	queues, err = client.Queues()
	assert.Nil(t, err)
	assert.NotEqual(t, -1, findQueue(queueTestName, queues))

	// bind queue to exchange
	err = BindQueueToExchange(ch, queueTestName, keyTestName, exchangeTestName)
	assert.Nil(t, err)
	bindings, err := client.Bindings()
	assert.Nil(t, err)
	assert.NotEqual(t, -1, findBinding(queueTestName, exchangeTestName, keyTestName, bindings))

	// unbind queue from exchange
	err = UnbindQueueFromExchange(ch, queueTestName, keyTestName, exchangeTestName)
	assert.Nil(t, err)
	bindings, err = client.Bindings()
	assert.Nil(t, err)
	assert.Equal(t, -1, findBinding(queueTestName, exchangeTestName, keyTestName, bindings))

	// finally remove queue
	err = RemoveQueue(ch, queueTestName, false, false)
	assert.Nil(t, err)
	queues, err = client.Queues()
	assert.Nil(t, err)
	assert.Equal(t, -1, findQueue(queueTestName, queues))
}
