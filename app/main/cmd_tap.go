// Copyright (C) 2017 Jan Delgado

package main

// subscribe cli command handler

import (
	"crypto/tls"
	"os"
	"os/signal"

	"github.com/jandelgado/rabtap"
)

func tapCmdShutdownFunc(taps []*rabtap.AmqpTap) {
	log.Info("rabtap tap threads shutting down ...")
	for _, tap := range taps {
		tap.Close()
	}
}

// cmdTap taps to the given exchanges and displays or saves the received
// messages.
func cmdTap(tapConfig []rabtap.TapConfiguration, tlsConfig *tls.Config,
	messageReceiveFunc MessageReceiveFunc, signalChannel chan os.Signal) {

	// this channel is used to decouple message receiving threads
	// with the main thread, which does the actual message processing
	tapMessageChannel := make(rabtap.TapChannel)
	taps := establishTaps(tapMessageChannel, tapConfig, tlsConfig)
	defer tapCmdShutdownFunc(taps)

	signal.Notify(signalChannel, os.Interrupt)

	messageReceiveLoop(tapMessageChannel, messageReceiveFunc, signalChannel)
}

// establishTaps establishes all message taps as specified by tapConfiguration
// array. All received messages will be send to the provided tapMessageChannel
// channel. Returns array of tabtap.AmqpTap objects and immeadiately starts
// the processing.
// TODO feature: discover bindings when no binding keys are given (-> discovery.go)
func establishTaps(tapMessageChannel rabtap.TapChannel,
	tapConfigs []rabtap.TapConfiguration, tlsConfig *tls.Config) []*rabtap.AmqpTap {
	taps := []*rabtap.AmqpTap{}
	for _, config := range tapConfigs {
		tap := rabtap.NewAmqpTap(config.AmqpURI, tlsConfig, log)
		go tap.EstablishTap(config.Exchanges, tapMessageChannel)
		taps = append(taps, tap)
	}
	return taps
}
