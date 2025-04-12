// Copyright (C) 2017 Jan Delgado

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"os"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"

	rabtap "github.com/jandelgado/rabtap/pkg"
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
	}

	if caFile != "" {
		caCert, err := os.ReadFile(caFile)
		failOnError(err, "invalid tls ca file", os.Exit)
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}
	return tlsConfig
}

func startCmdInfo(ctx context.Context, args CommandLineArgs, titleURL *url.URL, out *os.File) {
	filter, err := NewExprPredicate(args.Filter)
	failOnError(err, fmt.Sprintf("invalid queue filter predicate '%s'", args.Filter), os.Exit)

	cmdInfo(ctx,
		CmdInfoArg{
			rootNode: titleURL, // the title is constructed from this URL
			client:   rabtap.NewRabbitHTTPClient(args.APIURL, getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile)),
			treeConfig: BrokerInfoTreeBuilderConfig{
				Mode:                args.InfoMode,
				ShowConsumers:       args.ShowConsumers,
				ShowDefaultExchange: args.ShowDefaultExchange,
				Filter:              filter,
				OmitEmptyExchanges:  args.OmitEmptyExchanges,
			},
			renderConfig: BrokerInfoRendererConfig{
				Format:    args.Format,
				ShowStats: args.ShowStats,
			},
			out: NewColorableWriter(out),
		})
}

// createMessageReaderForPublish returns a message source that reads
// messages from the given source in the specified format. The source can
// be either empty (=stdin), a filename or a directory name
func newPublishMessageSource(source *string, format string) (MessageSource, error) {
	if source == nil {
		return NewReaderMessageSource(format, os.Stdin)
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
		return NewReaderMessageSource(format, file)
	} else {

		metadataFiles, err := LoadMetadataFilesFromDir(*source, os.ReadDir, NewRabtapFileInfoPredicate())
		if err != nil {
			return nil, err
		}

		sort.SliceStable(metadataFiles, func(i, j int) bool {
			return metadataFiles[i].metadata.XRabtapReceivedTimestamp.Before(
				metadataFiles[j].metadata.XRabtapReceivedTimestamp)
		})

		return NewReadFilesFromDirMessageSource(format, metadataFiles)
	}
}

func startCmdPublish(ctx context.Context, args CommandLineArgs) {
	if args.Format == "raw" && args.PubExchange == nil && args.PubRoutingKey == nil {
		fmt.Fprint(os.Stderr, "Warning: using raw message format but neither exchange or routing key are set.\n")
	}
	source, err := newPublishMessageSource(args.Source, args.Format)
	failOnError(err, "message-reader", os.Exit)
	source = NewTransformingMessageSource(source,
		FireHoseTransformer,
		NewPropertiesTransformer(args.Properties))

	err = cmdPublish(ctx, CmdPublishArg{
		amqpURL:    args.AMQPURL,
		exchange:   args.PubExchange,
		routingKey: args.PubRoutingKey,
		headers:    args.Args,
		fixedDelay: args.Delay,
		speed:      args.Speed,
		tlsConfig:  getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile),
		mandatory:  args.Mandatory,
		confirms:   args.Confirms,
		source:     source,
	})
	failOnError(err, "publish", os.Exit)
}

func startCmdSubscribe(ctx context.Context, args CommandLineArgs, out *os.File) {
	opts := MessageSinkOptions{
		out:              NewColorableWriter(out),
		format:           args.Format,
		silent:           args.Silent,
		optSaveDir:       args.SaveDir,
		filenameProvider: defaultFilenameProvider,
	}
	messageSink, err := NewMessageSink(opts)
	failOnError(err, "options", os.Exit)

	termPred, err := NewLoopCountPred(args.Limit)
	failOnError(err, "invalid message limit predicate", os.Exit)
	filterPred, err := NewExprPredicate(args.Filter)
	failOnError(err, fmt.Sprintf("invalid message filter predicate '%s'", args.Filter), os.Exit)

	err = cmdSubscribe(ctx, CmdSubscribeArg{
		amqpURL:     args.AMQPURL,
		queue:       args.QueueName,
		requeue:     args.Requeue,
		reject:      args.Reject,
		tlsConfig:   getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile),
		messageSink: messageSink,
		filterPred:  filterPred,
		termPred:    termPred,
		args:        args.Args,
		timeout:     args.IdleTimeout,
	})
	failOnError(err, "error subscribing messages", os.Exit)
}

func startCmdTap(ctx context.Context, args CommandLineArgs, out *os.File) {
	opts := MessageSinkOptions{
		out:              NewColorableWriter(out),
		format:           args.Format,
		silent:           args.Silent,
		optSaveDir:       args.SaveDir,
		filenameProvider: defaultFilenameProvider,
	}
	messageSink, err := NewMessageSink(opts)
	failOnError(err, "options", os.Exit)
	termPred, err := NewLoopCountPred(args.Limit)
	failOnError(err, "invalid message limit predicate", os.Exit)
	filterPred, err := NewExprPredicate(args.Filter)
	failOnError(err, fmt.Sprintf("invalid message filter predicate '%s'", args.Filter), os.Exit)

	cmdTap(ctx,
		CmdTapArg{
			tapConfig:   args.TapConfig,
			tlsConfig:   getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile),
			messageSink: messageSink,
			filterPred:  filterPred,
			termPred:    termPred,
			timeout:     args.IdleTimeout,
		})
}

func dispatchCmd(ctx context.Context, args CommandLineArgs, tlsConfig *tls.Config, out *os.File) {
	if args.commonArgs.NoColor {
		color.NoColor = true
	}
	if args.commonArgs.ForceColor {
		color.NoColor = false
	}
	switch args.Cmd {
	case HelpCmd:
		PrintHelp(args.HelpTopic)
	case InfoCmd:
		startCmdInfo(ctx, args, args.APIURL, out)
	case SubCmd:
		startCmdSubscribe(ctx, args, out)
	case PubCmd:
		startCmdPublish(ctx, args)
	case TapCmd:
		startCmdTap(ctx, args, out)
	case ExchangeCreateCmd:
		cmdExchangeCreate(CmdExchangeCreateArg{
			amqpURL:  args.AMQPURL,
			exchange: args.ExchangeName, exchangeType: args.ExchangeType,
			durable: args.Durable, autodelete: args.Autodelete,
			tlsConfig: tlsConfig, args: args.Args,
		})
	case ExchangeRemoveCmd:
		cmdExchangeRemove(args.AMQPURL, args.ExchangeName, tlsConfig)
	case ExchangeBindToExchangeCmd:
		cmdExchangeBindToExchange(CmdExchangeBindArg{
			amqpURL:        args.AMQPURL,
			sourceExchange: args.ExchangeName,
			targetExchange: args.DestExchangeName, key: args.BindingKey,
			headerMode: args.HeaderMode, args: args.Args,
			tlsConfig: tlsConfig,
		})
	case QueueCreateCmd:
		cmdQueueCreate(CmdQueueCreateArg{
			amqpURL: args.AMQPURL,
			queue:   args.QueueName, durable: args.Durable,
			autodelete: args.Autodelete, tlsConfig: tlsConfig,
			args: args.Args,
		})
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
			tlsConfig: tlsConfig,
		})
	case QueueUnbindCmd:
		cmdQueueUnbindFromExchange(CmdQueueBindArg{
			amqpURL:  args.AMQPURL,
			exchange: args.ExchangeName,
			queue:    args.QueueName, key: args.BindingKey,
			headerMode: args.HeaderMode, args: args.Args,
			tlsConfig: tlsConfig,
		})
	case ConnCloseCmd:
		failOnError(cmdConnClose(ctx, args.APIURL, args.ConnName,
			args.CloseReason, tlsConfig),
			fmt.Sprintf("close connection '%s'", args.ConnName), os.Exit)
	}
}

func main() {
	rabtap_main(os.Stdout)
}

func rabtap_main(out *os.File) {
	args, err := ParseCommandLineArgs(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	initLogging(args.Verbose) // TODO pass out
	tlsConfig := getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile)

	ctx, cancel := context.WithCancel(context.Background())
	go SigIntHandler(ctx, cancel)

	dispatchCmd(ctx, args, tlsConfig, out)
}