// Copyright (C) 2017 Jan Delgado

package main

// exchange related cli command handlers

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/url"

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
func cmdExchangeCreate(cmd CmdExchangeCreateArg, logger *slog.Logger) error {
	return rabtap.SimpleAmqpConnector(cmd.amqpURL,
		cmd.tlsConfig,
		func(session rabtap.Session) error {
			logger.Debug(fmt.Sprintf("creating exchange %s with type %s, args=%v",
				cmd.exchange, cmd.exchangeType, cmd.args))
			return rabtap.CreateExchange(session, cmd.exchange, cmd.exchangeType,
				cmd.durable, cmd.autodelete, rabtap.ToAMQPTable(cmd.args))
		})
}

// exchangeCreate removes an exchange on the given broker
func cmdExchangeRemove(amqpURL *url.URL, exchangeName string, tlsConfig *tls.Config, logger *slog.Logger) error {
	return rabtap.SimpleAmqpConnector(amqpURL,
		tlsConfig,
		func(session rabtap.Session) error {
			logger.Debug(fmt.Sprintf("removing exchange %s", exchangeName))
			return rabtap.RemoveExchange(session, exchangeName, false)
		})
}

// cmdExchangeBindToExchange binds an exchange  to another exchange
func cmdExchangeBindToExchange(cmd CmdExchangeBindArg, logger *slog.Logger) error {
	return rabtap.SimpleAmqpConnector(cmd.amqpURL, cmd.tlsConfig,
		func(session rabtap.Session) error {
			if cmd.headerMode != HeaderNone {
				cmd.args["x-match"] = amqpHeaderRoutingMode(cmd.headerMode)
			}
			logger.Debug(fmt.Sprintf("binding exchange %s to exchange %s w/ key %s and headers %v",
				cmd.sourceExchange, cmd.targetExchange, cmd.key, cmd.args))

			return rabtap.BindExchangeToExchange(session, cmd.sourceExchange, cmd.key, cmd.targetExchange, rabtap.ToAMQPTable(cmd.args))
		})
}
