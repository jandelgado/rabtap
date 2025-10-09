// rabtap queue related commands
// Copyright (C) 2017-2021 Jan Delgado

package main

import (
	"crypto/tls"
	"log/slog"
	"net/url"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// CmdQueueCreateArg contains the arguments for cmdQueueCreate
type CmdQueueCreateArg struct {
	amqpURL    *url.URL
	queue      string
	durable    bool
	autodelete bool
	tlsConfig  *tls.Config
	args       rabtap.KeyValueMap
}

type CmdQueueBindArg struct {
	amqpURL    *url.URL
	queue      string
	exchange   string
	key        string
	args       rabtap.KeyValueMap
	headerMode HeaderMode
	tlsConfig  *tls.Config
}

func amqpHeaderRoutingMode(mode HeaderMode) string {
	modes := map[HeaderMode]string{
		HeaderMatchAny: "any",
		HeaderMatchAll: "all",
		HeaderNone:     "",
	}
	return modes[mode]
}

// cmdQueueCreate creates a new queue on the given broker
func cmdQueueCreate(cmd CmdQueueCreateArg, logger *slog.Logger) error {
	return rabtap.SimpleAmqpConnector(cmd.amqpURL,
		cmd.tlsConfig,
		func(session rabtap.Session) error {
			logger.Debug("creating queue", "queue", cmd.queue, "autodelete", cmd.autodelete,
				"durable", cmd.durable, "args", cmd.args)
			return rabtap.CreateQueue(session, cmd.queue,
				cmd.durable, cmd.autodelete, false, rabtap.ToAMQPTable(cmd.args))
		})
}

// cmdQueueRemove removes an exchange on the given broker
// TODO(JD) add ifUnused, ifEmpty parameters
func cmdQueueRemove(amqpURL *url.URL, queueName string, tlsConfig *tls.Config, logger *slog.Logger) error {
	return rabtap.SimpleAmqpConnector(amqpURL,
		tlsConfig,
		func(session rabtap.Session) error {
			logger.Debug("removing queue", "queue", queueName)
			return rabtap.RemoveQueue(session, queueName, false, false)
		})
}

// cmdQueuePurge purges a queue, i.e. removes all queued elements
func cmdQueuePurge(amqpURL *url.URL, queueName string, tlsConfig *tls.Config, logger *slog.Logger) error {
	return rabtap.SimpleAmqpConnector(amqpURL,
		tlsConfig,
		func(session rabtap.Session) error {
			logger.Debug("purging queue", "queue", queueName)
			num, err := rabtap.PurgeQueue(session, queueName)
			if err == nil {
				logger.Info("purged queue", "num_elements", num, "queue", queueName)
			}
			return err
		})
}

// cmdQueueBindToExchange binds a queue to an exchange
func cmdQueueBindToExchange(cmd CmdQueueBindArg, logger *slog.Logger) error {
	return rabtap.SimpleAmqpConnector(cmd.amqpURL, cmd.tlsConfig,
		func(session rabtap.Session) error {
			if cmd.headerMode != HeaderNone {
				cmd.args["x-match"] = amqpHeaderRoutingMode(cmd.headerMode)
			}
			logger.Debug("binding queue to exchange",
				"queue", cmd.queue, "exchange", cmd.exchange, "key", cmd.key, "args", cmd.args)

			return rabtap.BindQueueToExchange(session, cmd.queue, cmd.key, cmd.exchange, rabtap.ToAMQPTable(cmd.args))
		})
}

// cmdQueueUnbindFromExchange unbinds a queue from an exchange
func cmdQueueUnbindFromExchange(cmd CmdQueueBindArg, logger *slog.Logger) error {
	return rabtap.SimpleAmqpConnector(cmd.amqpURL, cmd.tlsConfig,
		func(session rabtap.Session) error {
			if cmd.headerMode != HeaderNone {
				cmd.args["x-match"] = amqpHeaderRoutingMode(cmd.headerMode)
			}
			logger.Debug("unbinding queue from exchange",
				"queue", cmd.queue, "exchange", cmd.exchange, "key", cmd.key, "args", cmd.args)
			return rabtap.UnbindQueueFromExchange(session, cmd.queue, cmd.key, cmd.exchange, rabtap.ToAMQPTable(cmd.args))
		})
}
