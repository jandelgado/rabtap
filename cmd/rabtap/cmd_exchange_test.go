// Copyright (C) 2017 Jan Delgado

//go:build integration
// +build integration

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
)

func findExchangeByName(apiURL *url.URL, vhost, name string) (*rabtap.RabbitExchange, error) {
	client := rabtap.NewRabbitHTTPClient(apiURL, &tls.Config{})
	exchanges, err := client.Exchanges(context.TODO())
	if err != nil {
		return nil, err
	}
	for _, e := range exchanges {
		if e.Name == name && e.Vhost == vhost {
			return &e, nil
		}
	}
	return nil, fmt.Errorf("exchange not found")
}

func TestIntegrationCmdExchangeCreateRemoveExchange(t *testing.T) {
	// integration tests creation and removal of exchange
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	const testExchange = "cmd-exchange-test"

	amqpURL := testcommon.IntegrationURIFromEnv()
	apiURL, _ := url.Parse(testcommon.IntegrationAPIURIFromEnv())

	_, err := findExchangeByName(apiURL, "/", testExchange)
	assert.Error(t, fmt.Errorf("exchange not found"))

	os.Args = []string{"rabtap", "exchange", "create", testExchange, "--uri", amqpURL.String()}
	main()
	time.Sleep(2 * time.Second)
	_, err = findExchangeByName(apiURL, "/", testExchange)
	assert.NoError(t, err)

	// TODO validation

	os.Args = []string{"rabtap", "exchange", "rm", testExchange, "--uri", amqpURL.String()}
	main()
	time.Sleep(2 * time.Second)

	_, err = findExchangeByName(apiURL, "/", testExchange)
	assert.Error(t, fmt.Errorf("exchange not found"))
}
