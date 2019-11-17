// Copyright (C) 2017 Jan Delgado

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
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

func startCmdInfo(args CommandLineArgs, title string) {
	queueFilter, err := NewPredicateExpression(args.QueueFilter)
	failOnError(err, fmt.Sprintf("invalid queue filter predicate '%s'", args.QueueFilter), os.Exit)
	apiURL, err := url.Parse(args.APIURI)
	failOnError(err, "invalid api url", os.Exit)
	cmdInfo(CmdInfoArg{
		rootNode: title,
		client:   rabtap.NewRabbitHTTPClient(apiURL, getTLSConfig(args.InsecureTLS)),
		treeConfig: BrokerInfoTreeBuilderConfig{
			Mode:                args.InfoMode,
			ShowConsumers:       args.ShowConsumers,
			ShowDefaultExchange: args.ShowDefaultExchange,
			QueueFilter:         queueFilter,
			OmitEmptyExchanges:  args.OmitEmptyExchanges},
		renderConfig: BrokerInfoRendererConfig{
			Format:    args.InfoFormat,
			ShowStats: args.ShowStats,
			NoColor:   args.NoColor},
		out: NewColorableWriter(os.Stdout)})
}

func startCmdPublish(ctx context.Context, args CommandLineArgs) {
	file := os.Stdin
	if args.PubFile != nil {
		var err error
		file, err = os.Open(*args.PubFile)
		failOnError(err, "error opening "+*args.PubFile, os.Exit)
		defer file.Close()
	}
	readerFunc, err := createMessageReaderFunc(args.Format, file)
	failOnError(err, "options", os.Exit)
	err = cmdPublish(ctx, CmdPublishArg{
		amqpURI:    args.AmqpURI,
		exchange:   args.PubExchange,
		routingKey: args.PubRoutingKey,
		tlsConfig:  getTLSConfig(args.InsecureTLS),
		readerFunc: readerFunc})
	failOnError(err, "error publishing message", os.Exit)
}

func startCmdSubscribe(ctx context.Context, args CommandLineArgs) {
	opts := MessageReceiveFuncOptions{
		noColor:    args.NoColor,
		format:     args.Format,
		optSaveDir: args.SaveDir,
	}
	messageReceiveFunc, err := createMessageReceiveFunc(
		NewColorableWriter(os.Stdout), opts)
	failOnError(err, "options", os.Exit)
	err = cmdSubscribe(ctx, CmdSubscribeArg{
		amqpURI:            args.AmqpURI,
		queue:              args.QueueName,
		AutoAck:            args.AutoAck,
		tlsConfig:          getTLSConfig(args.InsecureTLS),
		messageReceiveFunc: messageReceiveFunc})
	failOnError(err, "error subscribing messages", os.Exit)
}

func startCmdTap(ctx context.Context, args CommandLineArgs) {
	opts := MessageReceiveFuncOptions{
		noColor:    args.NoColor,
		format:     args.Format,
		optSaveDir: args.SaveDir,
	}
	messageReceiveFunc, err := createMessageReceiveFunc(
		NewColorableWriter(os.Stdout), opts)
	failOnError(err, "options", os.Exit)
	cmdTap(ctx, args.TapConfig, getTLSConfig(args.InsecureTLS),
		messageReceiveFunc)
}

func dispatchCmd(ctx context.Context, args CommandLineArgs, tlsConfig *tls.Config) {
	switch args.Cmd {
	case InfoCmd:
		startCmdInfo(args, args.APIURI)
	case SubCmd:
		startCmdSubscribe(ctx, args)
	case PubCmd:
		startCmdPublish(ctx, args)
	case TapCmd:
		startCmdTap(ctx, args)
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

func main() {
	args, err := ParseCommandLineArgs(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	initLogging(args.Verbose)
	tlsConfig := getTLSConfig(args.InsecureTLS)

	// translate ^C (Interrput) in ctx.Done()
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	dispatchCmd(ctx, args, tlsConfig)
}
