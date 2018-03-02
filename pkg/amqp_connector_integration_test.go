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

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationWorkerInteraction checks that our worker function is properly
// called an that the shutdown mechanism works.
func TestIntegrationWorkerInteraction(t *testing.T) {

	log := log.New(os.Stdout, "ampq_connector_inttest: ", log.Lshortfile)

	resultChan := make(chan int, 1)

	worker := func(rabbitConn *amqp.Connection, controlChan chan ControlMessage) ReconnectAction {
		require.NotNil(t, rabbitConn)
		for {
			select {
			case controlMessage := <-controlChan:
				// when triggered by AmqpConnector.Close(), ShutdownMessage is expected
				assert.Equal(t, shutdownMessage, controlMessage)
				resultChan <- 1337
				if controlMessage.IsReconnect() {
					return doReconnect
				}
				return doNotReconnect
			}
		}
	}

	conn := NewAmqpConnector(testcommon.IntegrationURIFromEnv(), &tls.Config{}, log)
	go conn.Connect(worker)

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
