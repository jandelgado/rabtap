// Copyright (C) 2017 Jan Delgado

package main

// exchange related cli command handlers

import (
	"crypto/tls"
	"os"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/streadway/amqp"
)

// CmdQueueCreateArg contains the arguments for cmdQueueCreate
type CmdQueueCreateArg struct {
	amqpURI    string
	queue      string
	durable    bool
	autodelete bool
	tlsConfig  *tls.Config
}

// cmdQueueCreate creates a new queue on the given broker
func cmdQueueCreate(cmd CmdQueueCreateArg) {
	failOnError(rabtap.SimpleAmqpConnector(cmd.amqpURI,
		cmd.tlsConfig,
		func(chn *amqp.Channel) error {
			log.Debugf("creating queue %s (ad=%t, durable=%t)",
				cmd.queue, cmd.autodelete, cmd.durable)
			return rabtap.CreateQueue(chn, cmd.queue,
				cmd.durable, cmd.autodelete, false)
		}), "create queue failed", os.Exit)
}

// cmdQueueRemove removes an exchange on the given broker
// TODO(JD) add ifUnused, ifEmpty parameters
func cmdQueueRemove(amqpURI, queueName string, tlsConfig *tls.Config) {
	failOnError(rabtap.SimpleAmqpConnector(amqpURI,
		tlsConfig,
		func(chn *amqp.Channel) error {
			log.Debugf("removing queue %s", queueName)
			return rabtap.RemoveQueue(chn, queueName, false, false)
		}), "removing queue failed", os.Exit)
}

// cmdQueuePurge purges a queue, i.e. removes all queued elements
func cmdQueuePurge(amqpURI, queueName string, tlsConfig *tls.Config) {
	failOnError(rabtap.SimpleAmqpConnector(amqpURI,
		tlsConfig,
		func(chn *amqp.Channel) error {
			log.Debugf("purging queue %s", queueName)
			num, err := rabtap.PurgeQueue(chn, queueName)
			if err == nil {
				log.Infof("purged %d elements from queue %s", num, queueName)
			}
			return err
		}), "purge queue failed", os.Exit)
}

// cmdQueueBindToExchange binds a queue to an exchange
func cmdQueueBindToExchange(amqpURI, queueName, key, exchangeName string,
	tlsConfig *tls.Config) {

	failOnError(rabtap.SimpleAmqpConnector(amqpURI,
		tlsConfig,
		func(chn *amqp.Channel) error {
			log.Debugf("binding queue %s to exchange %s w/ key %s",
				queueName, exchangeName, key)
			return rabtap.BindQueueToExchange(chn, queueName, key, exchangeName)
		}), "bind queue failed", os.Exit)
}

// cmdQueueUnbindFromExchange unbinds a queue from an exchange
func cmdQueueUnbindFromExchange(amqpURI, queueName, key, exchangeName string,
	tlsConfig *tls.Config) {

	failOnError(rabtap.SimpleAmqpConnector(amqpURI,
		tlsConfig,
		func(chn *amqp.Channel) error {
			log.Debugf("unbinding queue %s from exchange %s w/ key %s",
				queueName, exchangeName, key)
			return rabtap.UnbindQueueFromExchange(chn, queueName, key, exchangeName)
		}), "unbind queue failed", os.Exit)
}
