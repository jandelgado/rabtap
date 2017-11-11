// Copyright (C) 2017 Jan Delgado
// +build integration

package rabtap

// integration test. assumes running rabbitmq broker on address
// defined by AMQP_URL environment variables.

import (
	"crypto/tls"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jandelgado/rabtap/testhelper"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

// TestIntegrationWorkerInteraction checks that our worker function is properly
// called an that the shutdown mechanism works.
func TestIntegrationWorkerInteraction(t *testing.T) {

	log := log.New(os.Stdout, "ampq_connector_inttest: ", log.Lshortfile)

	resultChan := make(chan int, 1)

	worker := func(rabbitConn *amqp.Connection, controlChan chan ControlMessage) bool {
		assert.NotNil(t, rabbitConn)
		for {
			select {
			case controlMessage := <-controlChan:
				// when triggered by AmqpConnector.Close(), ShutdownMessage is expected
				assert.Equal(t, ShutdownMessage, controlMessage)
				resultChan <- 1337
				// true signals caller to re-connect, false to end processing
				return controlMessage == ReconnectMessage
			}
		}
	}

	conn := NewAmqpConnector(testhelper.IntegrationURIFromEnv(), log)
	go conn.Connect(&tls.Config{}, worker)

	time.Sleep(time.Second * 2) // wait for connection to be established

	conn.Close()

	select {
	case <-time.After(5 * time.Second):
		assert.Fail(t, "worker did not shut down as expected")
	case val := <-resultChan:
		// make sure we received what we expect
		assert.Equal(t, 1337, val)
		assert.False(t, conn.Connected())
	}
}
