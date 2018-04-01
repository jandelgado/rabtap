// Copyright (C) 2017 Jan Delgado

package main

// subscribe cli command handler

import (
	"crypto/tls"
	"os"

	"github.com/jandelgado/rabtap/pkg"
)

// CmdSubscribeArg contains arguments for the subscribe command
type CmdSubscribeArg struct {
	amqpURI            string
	queue              string
	tlsConfig          *tls.Config
	messageReceiveFunc MessageReceiveFunc
	signalChannel      chan os.Signal
}

// cmdSub subscribes to messages from the given queue
func cmdSubscribe(cmd CmdSubscribeArg) {
	log.Debugf("cmdSub: subscribing to queue %s", cmd.queue)
	// this channel is used to decouple message receiving threads
	// with the main thread, which does the actual message processing
	messageChannel := make(rabtap.TapChannel)
	subscriber := rabtap.NewAmqpSubscriber(cmd.amqpURI, cmd.tlsConfig, log)
	defer subscriber.Close()
	go subscriber.EstablishSubscription(cmd.queue, messageChannel)

	messageReceiveLoop(messageChannel, cmd.messageReceiveFunc, cmd.signalChannel)
	log.Debug("cmdSub: cmd_subscribe ending")
}
