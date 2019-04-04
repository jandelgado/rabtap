// Copyright (C) 2017 Jan Delgado

package main

import (
	"crypto/tls"
	"os"
	"os/signal"

	//"net/http"
	//_ "net/http/pprof"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func initLogging(verbose bool) {
	log.Formatter = &logrus.TextFormatter{
		ForceColors:            true,
		DisableLevelTruncation: true,
		FullTimestamp:          false,
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

func createFilterPredicate(expr *string) (Predicate, error) {
	if expr != nil {
		return NewPredicateExpression(*expr)
	}
	return NewPredicateExpression("true")
}

func startCmdInfo(args CommandLineArgs, title string) {
	queueFilter, err := createFilterPredicate(args.QueueFilter)
	failOnError(err, "invalid queue filter predicate", os.Exit)
	cmdInfo(CmdInfoArg{
		rootNode: title,
		client:   rabtap.NewRabbitHTTPClient(args.APIURI, getTLSConfig(args.InsecureTLS)),
		printConfig: BrokerInfoPrinterConfig{
			ShowStats:           args.ShowStats,
			ShowConsumers:       args.ShowConsumers,
			ShowDefaultExchange: args.ShowDefaultExchange,
			ShowByConnection:    args.ShowByConnection,
			QueueFilter:         queueFilter,
			OmitEmptyExchanges:  args.OmitEmptyExchanges,
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
	err := cmdPublish(CmdPublishArg{
		amqpURI:             args.AmqpURI,
		exchange:            args.PubExchange,
		routingKey:          args.PubRoutingKey,
		tlsConfig:           getTLSConfig(args.InsecureTLS),
		readNextMessageFunc: readerFunc})
	failOnError(err, "error publishing message", os.Exit)
}

func startCmdSubscribe(args CommandLineArgs) {
	// signalChannel receives ctrl+C/interrput signal
	signalChannel := make(chan os.Signal, 5)
	signal.Notify(signalChannel, os.Interrupt)
	messageReceiveFunc := createMessageReceiveFunc(
		NewColorableWriter(os.Stdout), args.JSONFormat,
		args.SaveDir, args.NoColor)
	err := cmdSubscribe(CmdSubscribeArg{
		amqpURI:            args.AmqpURI,
		queue:              args.QueueName,
		tlsConfig:          getTLSConfig(args.InsecureTLS),
		messageReceiveFunc: messageReceiveFunc,
		signalChannel:      signalChannel})
	failOnError(err, "error subscribing messages", os.Exit)
}

func startCmdTap(args CommandLineArgs) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
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
		startCmdInfo(args, args.APIURI)
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
	case QueuePurgeCmd:
		cmdQueuePurge(args.AmqpURI, args.QueueName, tlsConfig)
	case QueueBindCmd:
		cmdQueueBindToExchange(args.AmqpURI, args.QueueName,
			args.QueueBindingKey, args.ExchangeName, tlsConfig)
	case QueueUnbindCmd:
		cmdQueueUnbindFromExchange(args.AmqpURI, args.QueueName,
			args.QueueBindingKey, args.ExchangeName, tlsConfig)
	case ConnCloseCmd:
		cmdConnClose(args.APIURI, args.ConnName,
			args.CloseReason, tlsConfig)
	}
}
