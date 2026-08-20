// Copyright (C) 2017 Jan Delgado
//go:build integration

package rabtap

// queue integration test functionality. assumes running rabbitmq broker on
// address defined by AMQP_URL and RABBIT_API_URL environment variables.
// (to start a local rabbitmq instance:
//  $ sudo  docker run --rm -ti -p5672:5672 rabbitmq:3-management)

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"testing"
	"uuid"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jandelgado/rabtap/pkg/testcommon"
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

func TestIntegrationAmqpPurgeQueue(t *testing.T) {
	queueTestName := fmt.Sprintf("purgetest-%s", uuid.New().String())
	const exchangeTestName = "" // default exchange

	// TODO empty queue before test in case it exisits

	// create queue
	setup, err := testcommon.IntegrationTestConnection("", "", 0, false)
	require.NoError(t, err)
	session := Session{setup.Conn, setup.Chan}
	defer func() { _ = setup.Conn.Close() }()
	err = CreateQueue(session, queueTestName, true /*durable*/, true /*ad*/, false /*excl*/, nil)
	require.NoError(t, err)

	// publish & purge 10 messages
	const numMessages = 10
	testcommon.PublishTestMessages(t, setup.Chan, numMessages, exchangeTestName, queueTestName, nil)
	num, err := PurgeQueue(session, queueTestName)
	assert.NoError(t, err)
	assert.Equal(t, numMessages, num)
	// TODO additionally verifiy that queue is empty

	// TODO remove queue
}

func TestIntegrationAmqpQueueCreateBindUnbindAndRemove(t *testing.T) {
	// since in order to remove and unbind a  queue we must create it first, we
	// tests these functions together in one test case.
	queueTestName := fmt.Sprintf("bind-unbind-test-%s", uuid.New().String())
	const exchangeTestName = "amq.direct"
	const keyTestName = "key"

	url, err := url.Parse(testcommon.IntegrationAPIURIFromEnv())
	require.NoError(t, err)
	client := NewRabbitHTTPClient(url, &tls.Config{})

	// create queue
	setup, err := testcommon.IntegrationTestConnection("", "", 0, false)
	require.NoError(t, err)
	session := Session{setup.Conn, setup.Chan}
	defer func() { _ = setup.Conn.Close() }()
	err = CreateQueue(session, queueTestName, true /*durable*/, true /*ad*/, false /*excl*/, nil)
	require.NoError(t, err)

	// check if queue was created
	queues, err := client.Queues(context.TODO())
	require.NoError(t, err)
	assert.NotEqual(t, -1, findQueue(queueTestName, queues))

	// bind queue to exchange
	err = BindQueueToExchange(session, queueTestName, keyTestName, exchangeTestName, amqp.Table{})
	assert.NoError(t, err)
	bindings, err := client.Bindings(context.TODO())
	assert.NoError(t, err)
	assert.NotEqual(t, -1, findBinding(queueTestName, exchangeTestName, keyTestName, bindings))

	// unbind queue from exchange
	err = UnbindQueueFromExchange(session, queueTestName, keyTestName, exchangeTestName, amqp.Table{})
	assert.NoError(t, err)
	bindings, err = client.Bindings(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, -1, findBinding(queueTestName, exchangeTestName, keyTestName, bindings))

	// finally remove queue
	err = RemoveQueue(session, queueTestName, false, false)
	assert.NoError(t, err)
	queues, err = client.Queues(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, -1, findQueue(queueTestName, queues))
}
