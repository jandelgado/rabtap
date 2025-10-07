// Copyright (C) 2017 Jan Delgado

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
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

// defaultFilenameProvider returns the default filename without extension to
// use when messages are saved to files during tap or subscribe.
func defaultFilenameProvider() string {
	return fmt.Sprintf("rabtap-%d", time.Now().UnixNano())
}

func getTLSConfig(insecureTLS bool, certFile string, keyFile string, caFile string) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecureTLS,
	}

	if certFile != "" && keyFile != "" {
		// Load client cert
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("load x509 key pair: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if caFile != "" {
		caCert, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("read ca file: %w", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}
	return tlsConfig, nil
}

func startCmdInfo(ctx context.Context, args CommandLineArgs, tlsConfig *tls.Config, titleURL *url.URL, out *os.File) error {
	filter, err := NewExprPredicate(args.Filter)
	if err != nil {
		return fmt.Errorf("invalid queue filter predicate '%s': %w", args.Filter, err)
	}
	return cmdInfo(ctx,
		CmdInfoArg{
			rootNode: titleURL, // the title is constructed from this URL
			client:   rabtap.NewRabbitHTTPClient(args.APIURL, tlsConfig),
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
		return nil, fmt.Errorf("stat message source file: %w", err)
	}

	if !fi.IsDir() {
		file, err := os.Open(*source)
		if err != nil {
			return nil, fmt.Errorf("open message source file: %w", err)
		}
		// TODO close file
		return NewReaderMessageSource(format, file)
	} else {

		metadataFiles, err := LoadMetadataFilesFromDir(*source, os.ReadDir, NewRabtapFileInfoPredicate())
		if err != nil {
			return nil, fmt.Errorf("load message metadata: %w", err)
		}

		sort.SliceStable(metadataFiles, func(i, j int) bool {
			return metadataFiles[i].metadata.XRabtapReceivedTimestamp.Before(
				metadataFiles[j].metadata.XRabtapReceivedTimestamp)
		})

		return NewReadFilesFromDirMessageSource(format, metadataFiles)
	}
}

func startCmdPublish(ctx context.Context, args CommandLineArgs, tlsConfig *tls.Config) error {
	if args.Format == "raw" && args.PubExchange == nil && args.PubRoutingKey == nil {
		slog.Warn("using raw message format but neither exchange or routing key are set.")
	}
	source, err := newPublishMessageSource(args.Source, args.Format)
	if err != nil {
		return fmt.Errorf("message source: %w", err)
	}
	source = NewTransformingMessageSource(source,
		FireHoseTransformer,
		NewPropertiesTransformer(args.Properties))

	return cmdPublish(ctx, CmdPublishArg{
		amqpURL:    args.AMQPURL,
		exchange:   args.PubExchange,
		routingKey: args.PubRoutingKey,
		headers:    args.Args,
		fixedDelay: args.Delay,
		speed:      args.Speed,
		tlsConfig:  tlsConfig,
		mandatory:  args.Mandatory,
		confirms:   args.Confirms,
		source:     source,
	})
}

func startCmdSubscribe(ctx context.Context, args CommandLineArgs, tlsConfig *tls.Config, out *os.File) error {
	opts := MessageSinkOptions{
		out:              NewColorableWriter(out),
		format:           args.Format,
		silent:           args.Silent,
		optSaveDir:       args.SaveDir,
		filenameProvider: defaultFilenameProvider,
	}
	messageSink, err := NewMessageSink(opts)
	if err != nil {
		return fmt.Errorf("create message sink: %w", err)
	}

	termPred, err := NewLoopCountPred(args.Limit)
	if err != nil {
		return fmt.Errorf("message limit predicate: %w", err)
	}
	filterPred, err := NewExprPredicate(args.Filter)
	if err != nil {
		return fmt.Errorf("message filter predicate: %w", err)
	}

	return cmdSubscribe(ctx, CmdSubscribeArg{
		amqpURL:     args.AMQPURL,
		queue:       args.QueueName,
		requeue:     args.Requeue,
		reject:      args.Reject,
		tlsConfig:   tlsConfig,
		messageSink: messageSink,
		filterPred:  filterPred,
		termPred:    termPred,
		args:        args.Args,
		timeout:     args.IdleTimeout,
	})
}

func startCmdTap(ctx context.Context, args CommandLineArgs, tlsConfig *tls.Config, out *os.File) error {
	opts := MessageSinkOptions{
		out:              NewColorableWriter(out),
		format:           args.Format,
		silent:           args.Silent,
		optSaveDir:       args.SaveDir,
		filenameProvider: defaultFilenameProvider,
	}
	messageSink, err := NewMessageSink(opts)
	if err != nil {
		return fmt.Errorf("create message sink: %w", err)
	}

	termPred, err := NewLoopCountPred(args.Limit)
	if err != nil {
		return fmt.Errorf("message limit predicate: %w", err)
	}

	filterPred, err := NewExprPredicate(args.Filter)
	if err != nil {
		return fmt.Errorf("message filter predicate: %w", err)
	}

	return cmdTap(ctx,
		CmdTapArg{
			tapConfig:   args.TapConfig,
			tlsConfig:   tlsConfig,
			messageSink: messageSink,
			filterPred:  filterPred,
			termPred:    termPred,
			timeout:     args.IdleTimeout,
		})
}

func dispatchCmd(ctx context.Context, args CommandLineArgs, tlsConfig *tls.Config, out *os.File) error {
	if args.NoColor {
		color.NoColor = true
	}
	if args.ForceColor {
		color.NoColor = false
	}
	switch args.Cmd {
	case HelpCmd:
		PrintHelp(args.HelpTopic)
		return nil
	case InfoCmd:
		return startCmdInfo(ctx, args, tlsConfig, args.APIURL, out)
	case SubCmd:
		return startCmdSubscribe(ctx, args, tlsConfig, out)
	case PubCmd:
		return startCmdPublish(ctx, args, tlsConfig)
	case TapCmd:
		return startCmdTap(ctx, args, tlsConfig, out)
	case ExchangeCreateCmd:
		return cmdExchangeCreate(CmdExchangeCreateArg{
			amqpURL:  args.AMQPURL,
			exchange: args.ExchangeName, exchangeType: args.ExchangeType,
			durable: args.Durable, autodelete: args.Autodelete,
			tlsConfig: tlsConfig, args: args.Args,
		})
	case ExchangeRemoveCmd:
		return cmdExchangeRemove(args.AMQPURL, args.ExchangeName, tlsConfig)
	case ExchangeBindToExchangeCmd:
		return cmdExchangeBindToExchange(CmdExchangeBindArg{
			amqpURL:        args.AMQPURL,
			sourceExchange: args.ExchangeName,
			targetExchange: args.DestExchangeName, key: args.BindingKey,
			headerMode: args.HeaderMode, args: args.Args,
			tlsConfig: tlsConfig,
		})
	case QueueCreateCmd:
		return cmdQueueCreate(CmdQueueCreateArg{
			amqpURL: args.AMQPURL,
			queue:   args.QueueName, durable: args.Durable,
			autodelete: args.Autodelete, tlsConfig: tlsConfig,
			args: args.Args,
		})
	case QueueRemoveCmd:
		return cmdQueueRemove(args.AMQPURL, args.QueueName, tlsConfig)
	case QueuePurgeCmd:
		return cmdQueuePurge(args.AMQPURL, args.QueueName, tlsConfig)
	case QueueBindCmd:
		return cmdQueueBindToExchange(CmdQueueBindArg{
			amqpURL:  args.AMQPURL,
			exchange: args.ExchangeName,
			queue:    args.QueueName, key: args.BindingKey,
			headerMode: args.HeaderMode, args: args.Args,
			tlsConfig: tlsConfig,
		})
	case QueueUnbindCmd:
		return cmdQueueUnbindFromExchange(CmdQueueBindArg{
			amqpURL:  args.AMQPURL,
			exchange: args.ExchangeName,
			queue:    args.QueueName, key: args.BindingKey,
			headerMode: args.HeaderMode, args: args.Args,
			tlsConfig: tlsConfig,
		})
	case ConnCloseCmd:
		return cmdConnClose(ctx, args.APIURL, args.ConnName,
			args.CloseReason, tlsConfig)
	default:
		return fmt.Errorf("unknown command %+v", args.Cmd)
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
	tlsConfig, err := getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go SigIntHandler(ctx, cancel)

	err = dispatchCmd(ctx, args, tlsConfig, out)
	if err != nil {
		log.Fatal(err)
	}
}
