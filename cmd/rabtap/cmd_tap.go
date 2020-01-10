// Copyright (C) 2017 Jan Delgado

package main

import (
	"context"
	"crypto/tls"

	"golang.org/x/sync/errgroup"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// cmdTap taps to the given exchanges and displays or saves the received
// messages.
// TODO feature: discover bindings when no binding keys are given (-> discovery.go)
func cmdTap(ctx context.Context, tapConfig []rabtap.TapConfiguration, tlsConfig *tls.Config,
	messageReceiveFunc MessageReceiveFunc) {

	g, ctx := errgroup.WithContext(ctx)

	tapMessageChannel := make(rabtap.TapChannel)

	for _, config := range tapConfig {
		tap := rabtap.NewAmqpTap(config.AmqpURI, tlsConfig, log)
		g.Go(func() error {
			return tap.EstablishTap(ctx, config.Exchanges, tapMessageChannel)
		})
	}
	g.Go(func() error {
		return messageReceiveLoop(ctx, tapMessageChannel, messageReceiveFunc)
	})
	if err := g.Wait(); err != nil {
		log.Errorf("tap failed with %v", err)
	}
}
