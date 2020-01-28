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

// exchangeExists queries the API to check if a given exchange exists
func exchangeExists(t *testing.T, apiURL *url.URL, exchange string) bool {
	// TODO add a simple client to testcommon
	client := rabtap.NewRabbitHTTPClient(apiURL, &tls.Config{})
	exchanges, err := client.Exchanges()
	require.Nil(t, err)
	return rabtap.FindExchangeByName(exchanges, "/", exchange) != -1
}

func TestIntegrationCmdExchangeCreateRemoveExchange(t *testing.T) {
	// integration tests creation and removal of exchange
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	const testExchange = "cmd-exchange-test"

	amqpURI := testcommon.IntegrationURIFromEnv()
	apiURL, _ := url.Parse(testcommon.IntegrationAPIURIFromEnv())

	assert.False(t, exchangeExists(t, apiURL, testExchange))
	os.Args = []string{"rabtap", "exchange", "create", testExchange, "--uri", amqpURI}
	main()
	time.Sleep(2 * time.Second)
	assert.True(t, exchangeExists(t, apiURL, testExchange))

	// TODO validation

	os.Args = []string{"rabtap", "exchange", "rm", testExchange, "--uri", amqpURI}
	main()
	time.Sleep(2 * time.Second)
	assert.False(t, exchangeExists(t, apiURL, testExchange))
}
