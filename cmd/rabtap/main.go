// Copyright (C) 2017 Jan Delgado

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"time"

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
		if err != nil {
			failOnError(err, "invalid client tls cert/key file", os.Exit)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.BuildNameToCertificate()
	}

	if caFile != "" {
		// Load CA cert
		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			failOnError(err, "invalid client tls ca file", os.Exit)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
		tlsConfig.BuildNameToCertificate()
	}
	return tlsConfig
}

func startCmdInfo(args CommandLineArgs, title string) {
	queueFilter, err := NewPredicateExpression(args.QueueFilter)
	failOnError(err, fmt.Sprintf("invalid queue filter predicate '%s'", args.QueueFilter), os.Exit)
	apiURL, err := url.Parse(args.APIURI)
	failOnError(err, "invalid api url", os.Exit)
	cmdInfo(CmdInfoArg{
		rootNode: title,
		client:   rabtap.NewRabbitHTTPClient(apiURL, getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile)),
		treeConfig: BrokerInfoTreeBuilderConfig{
			Mode:                args.InfoMode,
			ShowConsumers:       args.ShowConsumers,
			ShowDefaultExchange: args.ShowDefaultExchange,
			QueueFilter:         queueFilter,
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
func createMessageReaderForPublishFunc(source *string, format string) (MessageReaderFunc, error) {
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
	}

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

func startCmdPublish(ctx context.Context, args CommandLineArgs) {
	if args.Format == "raw" && args.PubExchange == nil && args.PubRoutingKey == nil {
		fmt.Fprint(os.Stderr, "Warning: using raw message format but neither exchange or routing key are set.\n")
	}
	readerFunc, err := createMessageReaderForPublishFunc(args.Source, args.Format)
	failOnError(err, "message-reader", os.Exit)
	err = cmdPublish(ctx, CmdPublishArg{
		amqpURI:    args.AmqpURI,
		exchange:   args.PubExchange,
		routingKey: args.PubRoutingKey,
		fixedDelay: args.Delay,
		speed:      args.Speed,
		tlsConfig:  getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile),
		readerFunc: readerFunc})
	failOnError(err, "error publishing message", os.Exit)
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
	err = cmdSubscribe(ctx, CmdSubscribeArg{
		amqpURI:            args.AmqpURI,
		queue:              args.QueueName,
		AutoAck:            args.AutoAck,
		tlsConfig:          getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile),
		messageReceiveFunc: messageReceiveFunc})
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
	cmdTap(ctx, args.TapConfig, getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile),
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
	tlsConfig := getTLSConfig(args.InsecureTLS, args.TLSCertFile, args.TLSKeyFile, args.TLSCaFile)

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
