// Copyright (C) 2017 Jan Delgado

package main

// subscribe cli command handler

import (
	"context"
	"crypto/tls"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"golang.org/x/sync/errgroup"
)

// CmdSubscribeArg contains arguments for the subscribe command
type CmdSubscribeArg struct {
	amqpURI            string
	queue              string
	tlsConfig          *tls.Config
	messageReceiveFunc MessageReceiveFunc
	AutoAck            bool
}

// cmdSub subscribes to messages from the given queue
func cmdSubscribe(ctx context.Context, cmd CmdSubscribeArg) error {
	log.Debugf("cmdSub: subscribing to queue %s", cmd.queue)

	g, ctx := errgroup.WithContext(ctx)

	// this channel is used to decouple message receiving threads
	// with the main thread, which does the actual message processing
	messageChannel := make(rabtap.TapChannel)
	config := rabtap.AmqpSubscriberConfig{Exclusive: false, AutoAck: cmd.AutoAck}
	subscriber := rabtap.NewAmqpSubscriber(config, cmd.amqpURI, cmd.tlsConfig, log)
	//defer subscriber.Close()

	g.Go(func() error { return subscriber.EstablishSubscription(ctx, cmd.queue, messageChannel) })
	g.Go(func() error { return messageReceiveLoop(ctx, messageChannel, cmd.messageReceiveFunc) })

	if err := g.Wait(); err != nil {
		log.Errorf("subscribe failed with %v", err)
		return err
	}
	return nil
}
