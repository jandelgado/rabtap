// Copyright (C) 2017 Jan Delgado

// +build integration

package rabtap

// pubishing integration test functionality. assumes running rabbitmq broker on
// address defined by AMQP_URL and RABBIT_API_URL environment variables.
// (to start a local rabbitmq instance:
//  $ sudo  docker run --rm -ti -p5672:5672 rabbitmq:3-management)

import (
	"crypto/tls"
	"log"
	"os"
	"testing"

	"github.com/jandelgado/rabtap/testhelper"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

const (
	numPublishingMessages = 10
)

func TestIntegrationAmqpPublishDirectExchange(t *testing.T) {

	// creates exchange "direct-exchange" and queues "queue-0" and "queue-1"
	conn, ch := testhelper.IntegrationTestConnection(t, "direct-exchange", "direct", 2, false)
	defer conn.Close()

	publisher := NewAmqpPublish(testhelper.IntegrationURIFromEnv(), &tls.Config{}, log.New(os.Stderr, "", log.LstdFlags))
	publishChannel := make(PublishChannel)
	go publisher.EstablishConnection(publishChannel)

	// AmqpPublish now has started a go-routine which handles
	// connection to broker and expects messages on the publishChannel
	for i := 0; i < numPublishingMessages; i++ {
		publishChannel <- &PublishMessage{Exchange: "direct-exchange",
			RoutingKey: "queue-1",
			Publishing: &amqp.Publishing{Body: []byte("Hello")}}
	}

	log.Println("receiving message...")
	doneChan := make(chan int)
	testhelper.VerifyTestMessageOnQueue(t, ch, "consumer", numPublishingMessages, "queue-1", doneChan)
	numReceivedOriginal := <-doneChan
	assert.Equal(t, numPublishingMessages, numReceivedOriginal)
	log.Println("good bye.")

	err := publisher.Close()
	assert.Nil(t, err)
	assert.False(t, publisher.Connected())
}
