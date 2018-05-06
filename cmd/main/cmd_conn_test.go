// Copyright (C) 2017 Jan Delgado
// +build integration

package main

import (
	"crypto/tls"
	"testing"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func findClosedConnName(connectionsBefore []rabtap.RabbitConnection,
	connectionsAfter []rabtap.RabbitConnection) string {
	// given to lists of connections, find the first connection by name which
	// is in the first, but not in the second list.
	for _, ca := range connectionsAfter {
		found := false
		for _, cb := range connectionsBefore {
			if ca.Name == cb.Name {
				found = true
				break
			}
		}
		if !found {
			return ca.Name
		}
	}
	return ""
}

func TestCmdCloseConnection(t *testing.T) {

	uri := testcommon.IntegrationAPIURIFromEnv()
	client := rabtap.NewRabbitHTTPClient(uri, &tls.Config{})

	// we can not get the name of a connection through the API of the AMQP client. So
	// we figure out the connections name by comparing the list of active
	// connection before and after we created our test connection. Therefore,
	// make sure this test runs isolated on the broker.
	connsBefore, err := client.Connections()
	require.Nil(t, err)

	// start the test connection to be terminated
	conn, _ := testcommon.IntegrationTestConnection(t, "", "", 0, false)

	// it takes a few seconds for the new connection to show up in the REST API
	time.Sleep(time.Second * 5)

	connsAfter, err := client.Connections()
	require.Nil(t, err)

	// we add a notification callback and expect the cb to be called
	// when we close the connection via the API
	errorChan := make(chan *amqp.Error)
	conn.NotifyClose(errorChan)

	connToClose := findClosedConnName(connsBefore, connsAfter)
	require.NotEqual(t, "", connToClose)

	// now close the newly created connection. TODO handle potential
	// call to failOnError in cmdConnClose
	err = cmdConnClose(uri, connToClose, "some reason", &tls.Config{})
	require.Nil(t, err)

	// ... and make sure it gets closed, notified by a message on the errorChan
	connClosed := false
	select {
	case <-errorChan:
		connClosed = true
	case <-time.After(time.Second * 2):
		assert.Fail(t, "did not receive notification within expected time")
	}
	assert.True(t, connClosed)
}
