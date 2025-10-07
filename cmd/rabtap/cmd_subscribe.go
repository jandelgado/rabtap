// subscribe cli command handler
// Copyright (C) 2017-2021 Jan Delgado

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"golang.org/x/sync/errgroup"
)

// CmdSubscribeArg contains arguments for the subscribe command
type CmdSubscribeArg struct {
	amqpURL     *url.URL
	queue       string
	tlsConfig   *tls.Config
	messageSink MessageSink
	termPred    Predicate
	filterPred  Predicate
	reject      bool
	requeue     bool
	args        rabtap.KeyValueMap
	timeout     time.Duration
}

// cmdSub subscribes to messages from the given queue
func cmdSubscribe(ctx context.Context, cmd CmdSubscribeArg, logger *slog.Logger) error {
	logger.Debug(fmt.Sprintf("cmdSub: subscribing to queue %s", cmd.queue))

	ctx, cancel := context.WithCancel(ctx)
	g, ctx := errgroup.WithContext(ctx)

	config := rabtap.AmqpSubscriberConfig{
		Exclusive: false,
		Args:      rabtap.ToAMQPTable(cmd.args)}
	subscriber := rabtap.NewAmqpSubscriber(config, cmd.amqpURL, cmd.tlsConfig, logger)

	messageChannel := make(rabtap.TapChannel)
	errorChannel := make(rabtap.SubscribeErrorChannel)
	g.Go(func() error { return subscriber.EstablishSubscription(ctx, cmd.queue, messageChannel, errorChannel) })
	g.Go(func() error {

		acknowledger := CreateAcknowledgeFunc(cmd.reject, cmd.requeue)
		err := MessageReceiveLoop(ctx,
			messageChannel,
			errorChannel,
			cmd.messageSink,
			cmd.filterPred,
			cmd.termPred,
			acknowledger,
			cmd.timeout,
			logger)
		cancel()
		return err
	})

	if err := g.Wait(); err != nil && err != ErrIdleTimeout {
		return fmt.Errorf("subscribe failed: %w", err)
	}
	return nil
}
