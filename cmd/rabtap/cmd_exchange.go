// Copyright (C) 2017 Jan Delgado

package main

// exchange related cli command handlers

import (
	"crypto/tls"
	"net/url"
	"os"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// CmdExchangeCreateArg contains argument for cmdExchangeCreate
type CmdExchangeCreateArg struct {
	amqpURL      *url.URL
	exchange     string
	exchangeType string
	durable      bool
	autodelete   bool
	args         rabtap.KeyValueMap
	tlsConfig    *tls.Config
}

type CmdExchangeBindArg struct {
	amqpURL        *url.URL
	sourceExchange string
	targetExchange string
	key            string
	args           rabtap.KeyValueMap
	headerMode     HeaderMode
	tlsConfig      *tls.Config
}

// exchangeCreate creates a new exchange on the given broker
func cmdExchangeCreate(cmd CmdExchangeCreateArg) {
	failOnError(rabtap.SimpleAmqpConnector(cmd.amqpURL,
		cmd.tlsConfig,
		func(session rabtap.Session) error {
			log.Debugf("creating exchange %s with type %s, args=%v",
				cmd.exchange, cmd.exchangeType, cmd.args)
			return rabtap.CreateExchange(session, cmd.exchange, cmd.exchangeType,
				cmd.durable, cmd.autodelete, rabtap.ToAMQPTable(cmd.args))
		}), "create exchange failed", os.Exit)
}

// exchangeCreate removes an exchange on the given broker
func cmdExchangeRemove(amqpURL *url.URL, exchangeName string, tlsConfig *tls.Config) {
	failOnError(rabtap.SimpleAmqpConnector(amqpURL,
		tlsConfig,
		func(session rabtap.Session) error {
			log.Debugf("removing exchange %s", exchangeName)
			return rabtap.RemoveExchange(session, exchangeName, false)
		}), "removing exchange failed", os.Exit)
}

// cmdExchangeBindToExchange binds an exchange  to another exchange
func cmdExchangeBindToExchange(cmd CmdExchangeBindArg) {
	failOnError(rabtap.SimpleAmqpConnector(cmd.amqpURL, cmd.tlsConfig,
		func(session rabtap.Session) error {
			if cmd.headerMode != HeaderNone {
				cmd.args["x-match"] = amqpHeaderRoutingMode(cmd.headerMode)
			}
			log.Debugf("binding exchange %s to exchange %s w/ key %s and headers %v",
				cmd.sourceExchange, cmd.targetExchange, cmd.key, cmd.args)

			return rabtap.BindExchangeToExchange(session, cmd.sourceExchange, cmd.key, cmd.targetExchange, rabtap.ToAMQPTable(cmd.args))
		}), "bind exchange failed", os.Exit)
}
