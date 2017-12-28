// Copyright (C) 2017 Jan Delgado

// +build integration

package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jandelgado/rabtap"
	"github.com/stretchr/testify/assert"
)

func TestCmdInfo(t *testing.T) {
	// REST api mock returning only empty messages
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "[ ]")
	}
	apiMock := httptest.NewServer(http.HandlerFunc(handler))
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
	assert.Equal(t, "http://rootnode", strings.TrimSpace(buf.String()))
}
