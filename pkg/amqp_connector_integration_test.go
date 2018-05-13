// Copyright (C) 2017 Jan Delgado
// +build integration

package rabtap

// integration test. assumes running rabbitmq broker on address
// defined by AMQP_URL environment variables.
// TODO add reconnection test (using toxiproxy)

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

func TestConnectFailsFastOnFirstNonSuccessfulConnect(t *testing.T) {

	log := log.New(os.Stdout, "ampq_connector_inttest: ", log.Lshortfile)

	conn := NewAmqpConnector("amqp://localhost:1", &tls.Config{}, log)

	worker := func(rabbitConn *amqp.Connection, controlChan chan ControlMessage) ReconnectAction {
		assert.Fail(t, "worker unexpectedly called during test")
		return doNotReconnect
	}

	err := conn.Connect(worker)
	assert.NotNil(t, err)
}

func XTestAbortConnectionPhase(t *testing.T) {

	// test must be rewritten, since connector now "fails fast" on error
	// on first connection attempt

	log := log.New(os.Stdout, "ampq_connector_inttest: ", log.Lshortfile)

	conn := NewAmqpConnector("amqp://localhost:1", &tls.Config{}, log)

	worker := func(rabbitConn *amqp.Connection, controlChan chan ControlMessage) ReconnectAction {
		assert.Fail(t, "worker unepctedly called during test")
		return doNotReconnect
	}

	done := make(chan error)
	go func() {
		done <- conn.Connect(worker)
	}()

	time.Sleep(time.Second * 1) // wait for go-routine to start
	conn.Close()

	select {
	case <-time.After(5 * time.Second):
		assert.Fail(t, "Connect() did not shut down as expected")
	case err := <-done:
		assert.Nil(t, err)
	}
}

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
