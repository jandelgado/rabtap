// Copyright (C) 2017 Jan Delgado
// +build integration

package rabtap

// integration test. assumes running rabbitmq broker on address
// defined by AMQP_URL environment variables.
// TODO add reconnection test (using toxiproxy)

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
)

func TestConnectFailsFastOnFirstNonSuccessfulConnect(t *testing.T) {

	ctx := context.Background()
	log := log.New(os.Stdout, "ampq_connector_inttest: ", log.Lshortfile)

	conn := NewAmqpConnector("amqp://localhost:1", &tls.Config{}, log)

	worker := func(ctx context.Context, session Session) (ReconnectAction, error) {
		assert.Fail(t, "worker unexpectedly called during test")
		return doNotReconnect, nil
	}

	err := conn.Connect(ctx, worker)
	assert.NotNil(t, err)
}

// TestIntegrationWorkerInteraction checks that our worker function is properly
// called an that the shutdown mechanism works.
func TestIntegrationWorkerInteraction(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	log := log.New(os.Stdout, "ampq_connector_inttest: ", log.Lshortfile)

	resultChan := make(chan int, 1)

	worker := func(ctx context.Context, session Session) (ReconnectAction, error) {
		for {
			select {
			case <-ctx.Done():
				resultChan <- 1337
				return doNotReconnect, nil
			}
		}
	}

	conn := NewAmqpConnector(testcommon.IntegrationURIFromEnv(), &tls.Config{}, log)
	go conn.Connect(ctx, worker)

	time.Sleep(time.Second * 2) // wait for connection to be established

	cancel()

	select {
	case <-time.After(5 * time.Second):
		assert.Fail(t, "worker did not shut down as expected")
	case val := <-resultChan:
		// make sure we received what we expect
		assert.Equal(t, 1337, val)
	}
}
