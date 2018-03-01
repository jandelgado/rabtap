// Copyright (C) 2017 Jan Delgado

package main

import (
	"bytes"
	"crypto/tls"
	"strings"
	"testing"

	"github.com/jandelgado/rabtap"
	"github.com/jandelgado/rabtap/testhelper"
	"github.com/stretchr/testify/assert"
)

func TestCmdInfoRootNodeOnly(t *testing.T) {
	// REST api mock returning "empty" broker
	apiMock := testhelper.NewRabbitAPIMock(testhelper.MockModeEmpty)
	client := rabtap.NewRabbitHTTPClient(apiMock.URL, &tls.Config{})

	printBrokerInfoConfig := PrintBrokerInfoConfig{
		ShowStats:           false,
		ShowConsumers:       false,
		ShowDefaultExchange: false,
		NoColor:             true}

	buf := bytes.NewBufferString("")
	cmdInfo(CmdInfoArg{
		rootNode:              "http://x:y@rootnode",
		client:                client,
		printBrokerInfoConfig: printBrokerInfoConfig,
		out: buf})
	assert.Equal(t, "http://rootnode (broker ver=3.6.9, mgmt ver=3.6.9, cluster=rabbit@08f57d1fe8ab)",
		strings.TrimSpace(buf.String()))
}
