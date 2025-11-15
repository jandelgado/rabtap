// Copyright (C) 2017 Jan Delgado
//go:build integration

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/jandelgado/rabtap/pkg/testcommon"
)

func TestAmqpHeaderRoutingModeConverts(t *testing.T) {
	assert.Equal(t, "all", amqpHeaderRoutingMode(HeaderMatchAll))
	assert.Equal(t, "any", amqpHeaderRoutingMode(HeaderMatchAny))
	assert.Equal(t, "", amqpHeaderRoutingMode(HeaderNone))
}

func findQueueByName(apiURL *url.URL, vhost, name string) (*rabtap.RabbitQueue, error) {
	client := rabtap.NewRabbitHTTPClient(apiURL, &tls.Config{})
	queues, err := client.Queues(context.TODO())
	if err != nil {
		return nil, err
	}
	for _, q := range queues {
		if q.Name == name && q.Vhost == vhost {
			return &q, nil
		}
	}
	return nil, fmt.Errorf("queue not found")
}

func TestIntegrationCmdQueueCreatePurgeiBindUnbindQueue(t *testing.T) {
	// integration tests queue creation, bind to exchange, purge,
	// unbdind from exchange via calls through the main method
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	const testMessage = "SubHello"
	const testQueue = "purge-queue-test"
	const testKey = testQueue
	const testExchange = "amq.direct"

	amqpURL := testcommon.IntegrationURIFromEnv().String()
	apiURL, _ := url.Parse(testcommon.IntegrationAPIURIFromEnv())

	os.Args = []string{"rabtap", "queue", "create", testQueue, "--uri", amqpURL}
	main()

	os.Args = []string{
		"rabtap", "queue", "bind", testQueue, "to", testExchange,
		"--bindingkey", testQueue,
		"--uri", amqpURL,
	}
	main()

	// TODO publish some messages

	// purge queue and check size
	os.Args = []string{"rabtap", "queue", "purge", testQueue, "--uri", amqpURL}
	main()

	time.Sleep(2 * time.Second)

	// Check that the queue was created using the REST API
	queue, err := findQueueByName(apiURL, "/", testQueue)
	require.NoError(t, err)

	// check that queue is empty
	assert.Equal(t, 0, queue.Messages)

	// unbind queue
	os.Args = []string{
		"rabtap", "queue", "unbind", testQueue, "from", testExchange,
		"--bindingkey", testQueue,
		"--uri", amqpURL,
	}
	main()

	// remove queue
	os.Args = []string{"rabtap", "queue", "rm", testQueue, "--uri", amqpURL}
	main()

	// TODO check that queue is removed
}
