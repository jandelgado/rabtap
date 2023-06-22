// Copyright (C) 2017 Jan Delgado

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	// log2 "log"

	"net/url"
	"os"
	"os/signal"
	"sort"
	"time"

	"github.com/fatih/color"
	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/sirupsen/logrus"
)

//var log = slog.Default()

// type Logger struct{}

// func (s Logger) Debug(format string, a ...interface{})  { log2.Printf("FATAL "+format, a...) }
// func (s Logger) Error(err error)                        { log2.Printf("ERROR %v", err) }
// func (s Logger) Fatal(err error)                        { log2.Fatalf("FATAL %v", err) }
// func (s Logger) Info(format string, a ...interface{})   { log2.Printf("INFO "+format, a...) }
// func (s Logger) Warnf(format string, a ...interface{})  { log2.Printf("WARN "+format, a...) }
// func (s Logger) Debugf(format string, a ...interface{}) { log2.Printf("DEBUG "+format, a...) }
// func (s Logger) Infof(format string, a ...interface{})  { log2.Printf("INFO "+format, a...) }
// func (s Logger) Errorf(format string, a ...interface{}) { log2.Printf("ERROR "+format, a...) }

// var log = Logger{}

var log = logrus.New()

// func initLogging(verbose bool) {
//     log.Formatter = &logrus.TextFormatter{
//         ForceColors:            true,
//         DisableLevelTruncation: true,
//         FullTimestamp:          false,
//     }
//     log.Out = NewColorableWriter(os.Stderr)
//     if verbose {
//         log.SetLevel(logrus.DebugLevel)
//     } else {
//         log.SetLevel(logrus.WarnLevel)
//     }
// }

func failOnError(err error, msg string, exitFunc func(int)) {
	if err != nil {
		log.Errorf("%s: %s", msg, err)
		exitFunc(1)
	}
}

// defaultFilenameProvider returns the default filename without extension to
// use when messages are saved to files during tap or subscribe.
func defaultFilenameProvider() string {
	return fmt.Sprintf("rabtap-%d", time.Now().UnixNano())
}

func getTLSConfig(insecureTLS bool, certFile string, keyFile string, caFile string) *tls.Config {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecureTLS,
	}

	if certFile != "" && keyFile != "" {
		// Load client cert
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		failOnError(err, "invalid client tls cert/key file", os.Exit)
		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.BuildNameToCertificate()
	}

	if caFile != "" {
		caCert, err := ioutil.ReadFile(caFile)
		failOnError(err, "invalid tls ca file", os.Exit)
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
		tlsConfig.BuildNameToCertificate()
	}
	return tlsConfig
}

func startCmdInfo(ctx context.Context, args CommandLineArgs, titleURL *url.URL) {
	queueFilter, err := NewPredicateExpression(args.QueueFilter)
	failOnError(err, fmt.Sprintf("invalid queue filter predicate '%s'", args.QueueFilter), os.Exit)

	cmdInfo(ctx,
		CmdInfoArg{
			rootNode: titleURL, // the title is constructed from this URL
			client:   rabtap.NewRabbitHTTPClient(args.APIURL, getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile)),
			treeConfig: BrokerInfoTreeBuilderConfig{
				Mode:                args.InfoMode,
				ShowConsumers:       args.ShowConsumers,
				ShowDefaultExchange: args.ShowDefaultExchange,
				Filter:              queueFilter,
				OmitEmptyExchanges:  args.OmitEmptyExchanges},
			renderConfig: BrokerInfoRendererConfig{
				Format:    args.Format,
				ShowStats: args.ShowStats,
				NoColor:   args.NoColor},
			out: NewColorableWriter(os.Stdout)})
}

// createMessageReaderForPublish returns a MessageReaderFunc that reads
// messages from the given source in the specified format. The source can
// be either empty (=stdin), a filename or a directory name
func createMessageReaderForPublishFunc(source *string, format string) (MessageProviderFunc, error) {
	if source == nil {
		return CreateMessageReaderFunc(format, os.Stdin)
	}

	fi, err := os.Stat(*source)
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		file, err := os.Open(*source)
		if err != nil {
			return nil, err
		}
		// TODO close file
		return CreateMessageReaderFunc(format, file)
	} else {

		metadataFiles, err := LoadMetadataFilesFromDir(*source, ioutil.ReadDir, NewRabtapFileInfoPredicate())
		if err != nil {
			return nil, err
		}

		sort.SliceStable(metadataFiles, func(i, j int) bool {
			return metadataFiles[i].metadata.XRabtapReceivedTimestamp.Before(
				metadataFiles[j].metadata.XRabtapReceivedTimestamp)
		})

		return CreateMessageFromDirReaderFunc(format, metadataFiles)
	}
}

func startCmdPublish(ctx context.Context, args CommandLineArgs) {
	if args.Format == "raw" && args.PubExchange == nil && args.PubRoutingKey == nil {
		fmt.Fprint(os.Stderr, "Warning: using raw message format but neither exchange or routing key are set.\n")
	}
	provider, err := createMessageReaderForPublishFunc(args.Source, args.Format)
	provider = NewTransformingMessageProvider(FireHoseTransformer, provider)

	failOnError(err, "message-reader", os.Exit)
	err = cmdPublish(ctx, CmdPublishArg{
		amqpURL:      args.AMQPURL,
		exchange:     args.PubExchange,
		routingKey:   args.PubRoutingKey,
		headers:      args.Args,
		fixedDelay:   args.Delay,
		speed:        args.Speed,
		tlsConfig:    getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile),
		mandatory:    args.Mandatory,
		confirms:     args.Confirms,
		providerFunc: provider})
	failOnError(err, "publish", os.Exit)
}

func startCmdSubscribe(ctx context.Context, args CommandLineArgs) {
	opts := MessageReceiveFuncOptions{
		out:              NewColorableWriter(os.Stdout),
		noColor:          args.NoColor,
		format:           args.Format,
		silent:           args.Silent,
		optSaveDir:       args.SaveDir,
		filenameProvider: defaultFilenameProvider,
	}
	messageReceiveFunc, err := createMessageReceiveFunc(opts)
	failOnError(err, "options", os.Exit)

	pred := createCountingMessageReceivePred(args.Limit)
	err = cmdSubscribe(ctx, CmdSubscribeArg{
		amqpURL:                args.AMQPURL,
		queue:                  args.QueueName,
		requeue:                args.Requeue,
		reject:                 args.Reject,
		tlsConfig:              getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile),
		messageReceiveFunc:     messageReceiveFunc,
		messageReceiveLoopPred: pred,
		args:                   args.Args,
		timeout:                args.IdleTimeout,
	})
	failOnError(err, "error subscribing messages", os.Exit)
}

func startCmdTap(ctx context.Context, args CommandLineArgs) {
	opts := MessageReceiveFuncOptions{
		out:              NewColorableWriter(os.Stdout),
		noColor:          args.NoColor,
		format:           args.Format,
		silent:           args.Silent,
		optSaveDir:       args.SaveDir,
		filenameProvider: defaultFilenameProvider,
	}
	messageReceiveFunc, err := createMessageReceiveFunc(opts)
	failOnError(err, "options", os.Exit)
	pred := createCountingMessageReceivePred(args.Limit)
	cmdTap(ctx,
		CmdTapArg{
			tapConfig:          args.TapConfig,
			tlsConfig:          getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile),
			messageReceiveFunc: messageReceiveFunc,
			pred:               pred,
			timeout:            args.IdleTimeout,
		})
}

func dispatchCmd(ctx context.Context, args CommandLineArgs, tlsConfig *tls.Config) {
	switch args.Cmd {
	case InfoCmd:
		startCmdInfo(ctx, args, args.APIURL)
	case SubCmd:
		startCmdSubscribe(ctx, args)
	case PubCmd:
		startCmdPublish(ctx, args)
	case TapCmd:
		startCmdTap(ctx, args)
	case ExchangeCreateCmd:
		cmdExchangeCreate(CmdExchangeCreateArg{amqpURL: args.AMQPURL,
			exchange: args.ExchangeName, exchangeType: args.ExchangeType,
			durable: args.Durable, autodelete: args.Autodelete,
			tlsConfig: tlsConfig, args: args.Args})
	case ExchangeRemoveCmd:
		cmdExchangeRemove(args.AMQPURL, args.ExchangeName, tlsConfig)
	case ExchangeBindToExchangeCmd:
		cmdExchangeBindToExchange(CmdExchangeBindArg{
			amqpURL:        args.AMQPURL,
			sourceExchange: args.ExchangeName,
			targetExchange: args.DestExchangeName, key: args.BindingKey,
			headerMode: args.HeaderMode, args: args.Args,
			tlsConfig: tlsConfig})
	case QueueCreateCmd:
		cmdQueueCreate(CmdQueueCreateArg{amqpURL: args.AMQPURL,
			queue: args.QueueName, durable: args.Durable,
			autodelete: args.Autodelete, tlsConfig: tlsConfig,
			args: args.Args})
	case QueueRemoveCmd:
		cmdQueueRemove(args.AMQPURL, args.QueueName, tlsConfig)
	case QueuePurgeCmd:
		cmdQueuePurge(args.AMQPURL, args.QueueName, tlsConfig)
	case QueueBindCmd:
		cmdQueueBindToExchange(CmdQueueBindArg{
			amqpURL:  args.AMQPURL,
			exchange: args.ExchangeName,
			queue:    args.QueueName, key: args.BindingKey,
			headerMode: args.HeaderMode, args: args.Args,
			tlsConfig: tlsConfig})
	case QueueUnbindCmd:
		cmdQueueUnbindFromExchange(CmdQueueBindArg{
			amqpURL:  args.AMQPURL,
			exchange: args.ExchangeName,
			queue:    args.QueueName, key: args.BindingKey,
			headerMode: args.HeaderMode, args: args.Args,
			tlsConfig: tlsConfig})
	case ConnCloseCmd:
		failOnError(cmdConnClose(ctx, args.APIURL, args.ConnName,
			args.CloseReason, tlsConfig),
			fmt.Sprintf("close connection '%s'", args.ConnName), os.Exit)

	}
}

func main() {
	color.NoColor = false

	ctx, _ := context.WithCancel(context.Background())

	args, err := ParseCommandLineArgs(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	//	initLogging(args.Verbose)
	tlsConfig := getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile)

	// translate ^C (Interrput) in ctx.Done()
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	//	signal.Notify(c, os.Interrupt)		// NOT WITH WASM!
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
