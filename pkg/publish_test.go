// Copyright (C) 2017 Jan Delgado

//go:build integration
// +build integration

package rabtap

// pubishing integration test functionality. assumes running rabbitmq broker on
// address defined by AMQP_URL and RABBIT_API_URL environment variables.
// (to start a local rabbitmq instance:
//  $ sudo  docker run --rm -ti -p5672:5672 rabbitmq:3-management)

import (
	"context"
	"crypto/tls"
	"log/slog"
	"testing"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

const (
	numPublishingMessages = 10
)

func TestIntegrationAmqpPublishDirectExchange(t *testing.T) {

	// creates exchange "direct-exchange" and queues "queue-0" and "queue-1"
	conn, ch := testcommon.IntegrationTestConnection(t, "direct-exchange", "direct", 2, false)
	defer conn.Close()

	logger := slog.New(slog.DiscardHandler)
	mandatory := true
	confirms := true
	publisher := NewAmqpPublish(testcommon.IntegrationURIFromEnv(), &tls.Config{}, mandatory, confirms, logger)
	publishChannel := make(PublishChannel)
	errorChannel := make(PublishErrorChannel)
	ctx := context.Background()

	go publisher.EstablishConnection(ctx, publishChannel, errorChannel)

	// AmqpPublish now has started a go-routine which handles
	// connection to broker and expects messages on the publishChannel
	key := "queue-1"
	for i := 0; i < numPublishingMessages; i++ {
		routing := NewRouting("direct-exchange", key, amqp.Table{})
		publishChannel <- &PublishMessage{
			Routing:    routing,
			Publishing: &amqp.Publishing{Body: []byte("Hello")}}
	}

	doneChan := make(chan int)
	testcommon.VerifyTestMessageOnQueue(t, ch, "consumer", numPublishingMessages, key, doneChan)
	numReceivedOriginal := <-doneChan
	assert.Equal(t, numPublishingMessages, numReceivedOriginal)
}
