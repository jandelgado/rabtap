// Copyright (C) 2017 Jan Delgado

// +build integration

package main

import (
	"crypto/tls"
	"net/url"
	"os"
	"testing"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmdPurgeQueue(t *testing.T) {

	// create a queue, publish some messages, purge queue and make
	// sure queue is empty
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	const testMessage = "SubHello"
	const testQueue = "purge-queue-test"
	const testKey = testQueue
	const testExchange = "amq.direct"

	amqpURI := testcommon.IntegrationURIFromEnv()
	apiURL, _ := url.Parse(testcommon.IntegrationAPIURIFromEnv())

	os.Args = []string{"rabtap", "queue",
		"create", testQueue,
		"--uri", amqpURI}
	main()

	os.Args = []string{"rabtap", "queue",
		"bind", testQueue,
		"to", testExchange,
		"--bindingkey", testQueue,
		"--uri", amqpURI}
	main()

	// TODO publish some messages

	// purge queue and check size
	os.Args = []string{"rabtap", "queue",
		"purge", testQueue,
		"--uri", amqpURI}
	main()

	time.Sleep(2 * time.Second)
	client := rabtap.NewRabbitHTTPClient(apiURL, &tls.Config{})
	queues, err := client.Queues()
	assert.Nil(t, err)
	i := rabtap.FindQueueByName(queues, "/", testQueue)
	require.True(t, i != -1)

	// check that queue is empty
	queue := queues[i]
	assert.Equal(t, 0, queue.Messages)

	// remove queue
	os.Args = []string{"rabtap", "queue",
		"rm", testQueue,
		"--uri", amqpURI}
	main()

}
