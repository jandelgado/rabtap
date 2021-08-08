// Copyright (C) 2017 Jan Delgado

package main

// exchange related cli command handlers

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
}

// cmdQueueCreate creates a new queue on the given broker
func cmdQueueCreate(cmd CmdQueueCreateArg) {
	failOnError(rabtap.SimpleAmqpConnector(cmd.amqpURL,
		cmd.tlsConfig,
		func(session rabtap.Session) error {
			log.Debugf("creating queue %s (ad=%t, durable=%t)",
				cmd.queue, cmd.autodelete, cmd.durable)
			return rabtap.CreateQueue(session, cmd.queue,
				cmd.durable, cmd.autodelete, false)
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
func cmdQueueBindToExchange(amqpURL *url.URL, queueName, key, exchangeName string,
	tlsConfig *tls.Config) {

	failOnError(rabtap.SimpleAmqpConnector(amqpURL,
		tlsConfig,
		func(session rabtap.Session) error {
			log.Debugf("binding queue %s to exchange %s w/ key %s",
				queueName, exchangeName, key)
			return rabtap.BindQueueToExchange(session, queueName, key, exchangeName)
		}), "bind queue failed", os.Exit)
}

// cmdQueueUnbindFromExchange unbinds a queue from an exchange
func cmdQueueUnbindFromExchange(amqpURL *url.URL, queueName, key, exchangeName string,
	tlsConfig *tls.Config) {

	failOnError(rabtap.SimpleAmqpConnector(amqpURL,
		tlsConfig,
		func(session rabtap.Session) error {
			log.Debugf("unbinding queue %s from exchange %s w/ key %s",
				queueName, exchangeName, key)
			return rabtap.UnbindQueueFromExchange(session, queueName, key, exchangeName)
		}), "unbind queue failed", os.Exit)
}
