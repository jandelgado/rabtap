// Copyright (C) 2017 Jan Delgado

package main

// subscribe cli command handler

import (
	"context"
	"crypto/tls"
	"net/url"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"golang.org/x/sync/errgroup"
)

// CmdSubscribeArg contains arguments for the subscribe command
type CmdSubscribeArg struct {
	amqpURL                *url.URL
	queue                  string
	tlsConfig              *tls.Config
	messageReceiveFunc     MessageReceiveFunc
	messageReceiveLoopPred MessageReceiveLoopPred
	reject                 bool
	requeue                bool
}

// cmdSub subscribes to messages from the given queue
func cmdSubscribe(ctx context.Context, cmd CmdSubscribeArg) error {
	log.Debugf("cmdSub: subscribing to queue %s", cmd.queue)

	ctx, cancel := context.WithCancel(ctx)
	g, ctx := errgroup.WithContext(ctx)

	messageChannel := make(rabtap.TapChannel)
	config := rabtap.AmqpSubscriberConfig{Exclusive: false, AutoAck: false}
	subscriber := rabtap.NewAmqpSubscriber(config, cmd.amqpURL, cmd.tlsConfig, log)

	g.Go(func() error { return subscriber.EstablishSubscription(ctx, cmd.queue, messageChannel) })
	g.Go(func() error {
		acknowledger := createAcknowledgeFunc(cmd.reject, cmd.requeue)
		err := messageReceiveLoop(ctx, messageChannel, cmd.messageReceiveFunc, cmd.messageReceiveLoopPred, acknowledger)
		cancel()
		return err
	})

	if err := g.Wait(); err != nil {
		log.Errorf("subscribe failed with %v", err)
		return err
	}
	return nil
}
