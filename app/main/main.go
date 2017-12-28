// Copyright (C) 2017 Jan Delgado

package main

import (
	"crypto/tls"
	"os"

	"github.com/jandelgado/rabtap"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func initLogging(verbose bool) {
	log.Formatter = &logrus.TextFormatter{
		ForceColors: true,
	}
	log.Out = NewColorableWriter(os.Stderr)
	if verbose {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.WarnLevel)
	}
}

func failOnError(err error, msg string, exitFunc func(int)) {
	if err != nil {
		log.Errorf("%s: %s", msg, err)
		exitFunc(1)
	}
}

func getTLSConfig(insecureTLS bool) *tls.Config {
	return &tls.Config{InsecureSkipVerify: insecureTLS}
}

func startCmdInfo(args CommandLineArgs) {
	cmdInfo(CmdInfoArg{
		rootNode: args.APIURI,
		client:   rabtap.NewRabbitHTTPClient(args.APIURI, getTLSConfig(args.InsecureTLS)),
		printBrokerInfoConfig: PrintBrokerInfoConfig{
			ShowStats:           args.ShowStats,
			ShowConsumers:       args.ShowConsumers,
			ShowDefaultExchange: args.ShowDefaultExchange,
			NoColor:             args.NoColor},
		out: NewColorableWriter(os.Stdout)})
}

func startCmdPublish(args CommandLineArgs) {
	reader := os.Stdin
	if args.PubFile != nil {
		var err error
		reader, err = os.Open(*args.PubFile)
		failOnError(err, "error opening "+*args.PubFile, os.Exit)
	}
	readerFunc := createMessageReaderFunc(args.JSONFormat, reader)
	cmdPublish(CmdPublishArg{
		amqpURI:             args.AmqpURI,
		exchange:            args.PubExchange,
		routingKey:          args.PubRoutingKey,
		tlsConfig:           getTLSConfig(args.InsecureTLS),
		readNextMessageFunc: readerFunc})
}

func startCmdSubscribe(args CommandLineArgs) {
	// signalChannel receives ctrl+C/interrput signal
	signalChannel := make(chan os.Signal, 1)
	// messageReceiveFunc receives the tapped messages, prints
	// and optionally saves them.
	messageReceiveFunc := createMessageReceiveFunc(
		NewColorableWriter(os.Stdout), args.JSONFormat,
		args.SaveDir, args.NoColor)
	cmdSubscribe(CmdSubscribeArg{
		amqpURI:            args.AmqpURI,
		queue:              args.QueueName,
		tlsConfig:          getTLSConfig(args.InsecureTLS),
		messageReceiveFunc: messageReceiveFunc,
		signalChannel:      signalChannel})
}

func startCmdTap(args CommandLineArgs) {
	signalChannel := make(chan os.Signal, 1)
	messageReceiveFunc := createMessageReceiveFunc(
		NewColorableWriter(os.Stdout), args.JSONFormat,
		args.SaveDir, args.NoColor)
	cmdTap(args.TapConfig, getTLSConfig(args.InsecureTLS),
		messageReceiveFunc, signalChannel)
}

func main() {
	args, err := ParseCommandLineArgs(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	initLogging(args.Verbose)
	log.Debugf("parsed cli-args: %+v", args)

	tlsConfig := getTLSConfig(args.InsecureTLS)

	switch args.Cmd {
	case InfoCmd:
		startCmdInfo(args)
	case SubCmd:
		startCmdSubscribe(args)
	case PubCmd:
		startCmdPublish(args)
	case TapCmd:
		startCmdTap(args)
	case ExchangeCreateCmd:
		cmdExchangeCreate(CmdExchangeCreateArg{amqpURI: args.AmqpURI,
			exchange: args.ExchangeName, exchangeType: args.ExchangeType,
			durable: args.Durable, autodelete: args.Autodelete,
			tlsConfig: tlsConfig})
	case ExchangeRemoveCmd:
		cmdExchangeRemove(args.AmqpURI, args.ExchangeName, tlsConfig)
	case QueueCreateCmd:
		cmdQueueCreate(CmdQueueCreateArg{amqpURI: args.AmqpURI,
			queue: args.QueueName, durable: args.Durable,
			autodelete: args.Autodelete, tlsConfig: tlsConfig})
	case QueueRemoveCmd:
		cmdQueueRemove(args.AmqpURI, args.QueueName, tlsConfig)
	case QueueBindCmd:
		cmdQueueBindToExchange(args.AmqpURI, args.QueueName,
			args.QueueBindingKey, args.ExchangeName, tlsConfig)
	}
}
