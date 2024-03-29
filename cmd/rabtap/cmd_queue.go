// rabtap queue related commands
// Copyright (C) 2017-2021 Jan Delgado

package main

import (
	"crypto/tls"
	"net/url"
	"os"

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
		HeaderNone:     ""}
	return modes[mode]
}

// cmdQueueCreate creates a new queue on the given broker
func cmdQueueCreate(cmd CmdQueueCreateArg) {
	failOnError(rabtap.SimpleAmqpConnector(cmd.amqpURL,
		cmd.tlsConfig,
		func(session rabtap.Session) error {
			log.Debugf("creating queue %s (autodelete=%t, durable=%t, args=%v)",
				cmd.queue, cmd.autodelete, cmd.durable, cmd.args)
			return rabtap.CreateQueue(session, cmd.queue,
				cmd.durable, cmd.autodelete, false, rabtap.ToAMQPTable(cmd.args))
		}), "create queue failed", os.Exit)
}

// cmdQueueRemove removes an exchange on the given broker
// TODO(JD) add ifUnused, ifEmpty parameters
func cmdQueueRemove(amqpURL *url.URL, queueName string, tlsConfig *tls.Config) {
	failOnError(rabtap.SimpleAmqpConnector(amqpURL,
		tlsConfig,
		func(session rabtap.Session) error {
			log.Debugf("removing queue %s", queueName)
			return rabtap.RemoveQueue(session, queueName, false, false)
		}), "removing queue failed", os.Exit)
}

// cmdQueuePurge purges a queue, i.e. removes all queued elements
func cmdQueuePurge(amqpURL *url.URL, queueName string, tlsConfig *tls.Config) {
	failOnError(rabtap.SimpleAmqpConnector(amqpURL,
		tlsConfig,
		func(session rabtap.Session) error {
			log.Debugf("purging queue %s", queueName)
			num, err := rabtap.PurgeQueue(session, queueName)
			if err == nil {
				log.Infof("purged %d elements from queue %s", num, queueName)
			}
			return err
		}), "purge queue failed", os.Exit)
}

// cmdQueueBindToExchange binds a queue to an exchange
func cmdQueueBindToExchange(cmd CmdQueueBindArg) {
	failOnError(rabtap.SimpleAmqpConnector(cmd.amqpURL, cmd.tlsConfig,
		func(session rabtap.Session) error {
			if cmd.headerMode != HeaderNone {
				cmd.args["x-match"] = amqpHeaderRoutingMode(cmd.headerMode)
			}
			log.Debugf("binding queue %s to exchange %s w/ key %s and headers %v",
				cmd.queue, cmd.exchange, cmd.key, cmd.args)

			return rabtap.BindQueueToExchange(session, cmd.queue, cmd.key, cmd.exchange, rabtap.ToAMQPTable(cmd.args))
		}), "bind queue failed", os.Exit)
}

// cmdQueueUnbindFromExchange unbinds a queue from an exchange
func cmdQueueUnbindFromExchange(cmd CmdQueueBindArg) {
	failOnError(rabtap.SimpleAmqpConnector(cmd.amqpURL, cmd.tlsConfig,
		func(session rabtap.Session) error {
			if cmd.headerMode != HeaderNone {
				cmd.args["x-match"] = amqpHeaderRoutingMode(cmd.headerMode)
			}
			log.Debugf("unbinding queue %s from exchange %s w/ key %s and headers %v",
				cmd.queue, cmd.exchange, cmd.key, cmd.args)
			return rabtap.UnbindQueueFromExchange(session, cmd.queue, cmd.key, cmd.exchange, rabtap.ToAMQPTable(cmd.args))
		}), "unbind queue failed", os.Exit)
}
