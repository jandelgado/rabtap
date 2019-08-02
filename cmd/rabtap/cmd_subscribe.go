// Copyright (C) 2017 Jan Delgado

package main

// subscribe cli command handler

import (
	"context"
	"crypto/tls"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// CmdSubscribeArg contains arguments for the subscribe command
type CmdSubscribeArg struct {
	amqpURI            string
	queue              string
	tlsConfig          *tls.Config
	messageReceiveFunc MessageReceiveFunc
}

// cmdSub subscribes to messages from the given queue
func cmdSubscribe(ctx context.Context, cmd CmdSubscribeArg) error {
	log.Debugf("cmdSub: subscribing to queue %s", cmd.queue)

	// this channel is used to decouple message receiving threads
	// with the main thread, which does the actual message processing
	messageChannel := make(rabtap.TapChannel)
	subscriber := rabtap.NewAmqpSubscriber(cmd.amqpURI, false, cmd.tlsConfig, log)
	defer subscriber.Close()
	go subscriber.EstablishSubscription(cmd.queue, messageChannel)
	return messageReceiveLoop(ctx, messageChannel, cmd.messageReceiveFunc)
}
