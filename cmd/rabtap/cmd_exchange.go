// Copyright (C) 2017 Jan Delgado

package main

// exchange related cli command handlers

import (
	"crypto/tls"
	"os"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// CmdExchangeCreateArg contains argument for cmdExchangeCreate
type CmdExchangeCreateArg struct {
	amqpURI      string
	exchange     string
	exchangeType string
	durable      bool
	autodelete   bool
	tlsConfig    *tls.Config
}

// exchangeCreate creates a new exchange on the given broker
func cmdExchangeCreate(cmd CmdExchangeCreateArg) {
	failOnError(rabtap.SimpleAmqpConnector(cmd.amqpURI,
		cmd.tlsConfig,
		func(session rabtap.Session) error {
			log.Debugf("creating exchange %s with type %s",
				cmd.exchange, cmd.exchangeType)
			return rabtap.CreateExchange(session, cmd.exchange, cmd.exchangeType,
				cmd.durable, cmd.autodelete)
		}), "create exchange failed", os.Exit)
}

// exchangeCreate removes an exchange on the given broker
func cmdExchangeRemove(amqpURI, exchangeName string, tlsConfig *tls.Config) {
	failOnError(rabtap.SimpleAmqpConnector(amqpURI,
		tlsConfig,
		func(session rabtap.Session) error {
			log.Debugf("removing exchange %s", exchangeName)
			return rabtap.RemoveExchange(session, exchangeName, false)
		}), "removing exchange failed", os.Exit)
}
