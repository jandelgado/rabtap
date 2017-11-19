// Copyright (C) 2017 Jan Delgado

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/jandelgado/rabtap"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var log = logrus.New()

// MessageReaderFunc provides messages to be sent to an exchange. If no
// more messages are available io.EOF is returned as error.
type MessageReaderFunc func() (amqp.Publishing, error)

// MessageReceiveFunc processes receiced messages from a tap.
type MessageReceiveFunc func(*amqp.Delivery) error

// SignalChannel transports os.Signal objects like e.g. os.Interrupt
type SignalChannel chan os.Signal

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

func tapModeShutdownFunc(taps []*rabtap.AmqpTap) {
	log.Info("rabtap tap threads shutting down ...")
	for _, tap := range taps {
		tap.Close()
	}
}

// tapMode taps to the given exchanges and displays or saves the received
// messages.
func startTapMode(tapConfig []rabtap.TapConfiguration, insecureTLS bool,
	messageReceiveFunc MessageReceiveFunc, signalChannel chan os.Signal) {

	// this channel is used to decouple message receiving threads
	// with the main thread, which does the actual message processing
	tapMessageChannel := make(rabtap.TapChannel)
	taps := establishTaps(tapMessageChannel, tapConfig, insecureTLS)
	defer tapModeShutdownFunc(taps)

	signal.Notify(signalChannel, os.Interrupt)

ReceiverLoop:
	for {
		select {
		case message := <-tapMessageChannel:
			if message.Error != nil {
				// unrecoverable error received -> log and exit
				log.Error(message.Error)
				break ReceiverLoop
			}
			log.Debugf("received message %#+v", message.AmqpMessage)
			// let the receiveFunc do the actual message processing
			if err := messageReceiveFunc(message.AmqpMessage); err != nil {
				log.Error(err)
			}
		case <-signalChannel:
			break ReceiverLoop
		}
	}
}

// startInfoMode queries the rabbitMQ brokers REST api and dispays infos
// on exchanges, queues, bindings etc. in a human readably fashion.
func startInfoMode(rootNode string, client *rabtap.RabbitHTTPClient,
	printBrokerInfoConfig PrintBrokerInfoConfig) {
	brokerInfo, err := NewBrokerInfo(client)
	failOnError(err, "failed retrieving info from rabbitmq REST api", os.Exit)
	brokerInfoPrinter := NewBrokerInfoPrinter(printBrokerInfoConfig)
	brokerInfoPrinter.Print(brokerInfo, rootNode, NewColorableWriter(os.Stdout))
}

// piblishMessage publishes a message on the given exchange with the provided
// routingkey
func publishMessage(publishChannel rabtap.PublishChannel,
	exchange, routingKey string,
	amqpPublishing amqp.Publishing) {

	log.Debugf("publishing message %+v to exchange %s with routing key %s",
		amqpPublishing, exchange, routingKey)

	publishChannel <- &rabtap.PublishMessage{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Publishing: &amqpPublishing}
}

// readMessageFromRawFile reads a single messages from the given io.Reader
// which is typically stdin or a file. On subsequent calls, it returns io.EOF.
func readMessageFromRawFile(reader io.Reader) (amqp.Publishing, error) {
	buf := new(bytes.Buffer)
	if numRead, err := buf.ReadFrom(reader); err != nil {
		return amqp.Publishing{}, err
	} else if numRead == 0 {
		return amqp.Publishing{}, io.EOF
	}
	return amqp.Publishing{Body: buf.Bytes()}, nil
}

// readMessageFromJSON reads JSON messages from the given decoder as long
// as there are messages available.
func readMessageFromJSON(decoder *json.Decoder) (amqp.Publishing, error) {
	message := RabtapPersistentMessage{}
	err := decoder.Decode(&message)
	if err != nil {
		return amqp.Publishing{}, err
	}
	return message.ToAmqpPublishing(), nil
}

// createMessageReaderFunc returns a function that reads messages from the
// the given reader in JSON or raw-format
func createMessageReaderFunc(jsonFormat bool, reader io.Reader) MessageReaderFunc {
	if jsonFormat {
		decoder := json.NewDecoder(reader)
		return func() (amqp.Publishing, error) { return readMessageFromJSON(decoder) }
	}
	return func() (amqp.Publishing, error) { return readMessageFromRawFile(reader) }
}

// publishMessages reads messages with the provided readNextMessageFunc and
// publishes the messages to the given exchange.
func publishMessages(publishChannel rabtap.PublishChannel,
	exchange, routingKey string,
	readNextMessageFunc MessageReaderFunc) {
	for {
		amqpMessage, err := readNextMessageFunc()
		switch err {
		case io.EOF:
			return
		case nil:
			publishMessage(publishChannel, exchange, routingKey, amqpMessage)
		default:
			failOnError(err, "error reading message", os.Exit)
		}
	}
}

// startPublishMode reads messages with the provied readNextMessageFunc and
// publishes the messages to the given exchange.
func startPublishMode(amqpURI, exchange, routingKey string,
	insecureTLS bool, readNextMessageFunc MessageReaderFunc) {

	log.Debugf("publishing message(s) to exchange %s with routingkey %s",
		exchange, routingKey)
	publisher := rabtap.NewAmqpPublish(amqpURI,
		&tls.Config{InsecureSkipVerify: insecureTLS},
		log)
	defer publisher.Close()
	publishChannel := make(rabtap.PublishChannel)
	go publisher.EstablishConnection(publishChannel)
	publishMessages(publishChannel, exchange, routingKey, readNextMessageFunc)
}

// establishTaps establish all message taps as specified by tapConfiguration
// array. All received messages will be send to the provided tapMessageChannel
// channel. Returns array of tabtap.AmqpTap objects and immeadiately starts
// the processing.
// TODO feature: discover bindings when no binding keys are given (-> discovery.go)
func establishTaps(tapMessageChannel rabtap.TapChannel,
	tapConfigs []rabtap.TapConfiguration, insecureTLS bool) []*rabtap.AmqpTap {
	taps := []*rabtap.AmqpTap{}
	for _, config := range tapConfigs {
		tap := rabtap.NewAmqpTap(config.AmqpURI,
			&tls.Config{InsecureSkipVerify: insecureTLS}, log)
		go tap.EstablishTap(config.Exchanges, tapMessageChannel)
		taps = append(taps, tap)
	}
	return taps
}

// createMessageReceiveFuncJSON returns a function that processes received
// messages as JSON messages
// TODO make testable and write test
func createMessageReceiveFuncJSON(out io.Writer, optSaveDir *string, _ bool) MessageReceiveFunc {
	return func(message *amqp.Delivery) error {
		err := WriteMessageJSON(out, true, message)
		if err != nil || optSaveDir == nil {
			return err
		}
		filename := path.Join(*optSaveDir,
			fmt.Sprintf("rabtap-%d.json", time.Now().UnixNano()))
		return SaveMessageToJSONFile(filename, message)
	}
}

// createMessageReceiveFuncRaw returns a function that processes received
// messages as "raw" messages
// TODO make testable and write test
func createMessageReceiveFuncRaw(out io.Writer, optSaveDir *string, noColor bool) MessageReceiveFunc {
	return func(message *amqp.Delivery) error {
		err := PrettyPrintMessage(out, message,
			fmt.Sprintf("message received on %s",
				time.Now().Format(time.RFC3339)),
			noColor,
		)
		if err != nil || optSaveDir == nil {
			return err
		}
		basename := path.Join(*optSaveDir,
			fmt.Sprintf("rabtap-%d", time.Now().UnixNano()))
		return SaveMessageToRawFile(basename, message)
	}
}

func createMessageReceiveFunc(out io.Writer, jsonFormat bool, optSaveDir *string, noColor bool) MessageReceiveFunc {
	if jsonFormat {
		return createMessageReceiveFuncJSON(out, optSaveDir, noColor)
	}
	return createMessageReceiveFuncRaw(out, optSaveDir, noColor)
}

func main() {

	args, err := ParseCommandLineArgs(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	initLogging(args.Verbose)
	log.Debugf("parsed cli-args: %+v", args)

	switch args.Mode {
	case PubMode:
		reader := os.Stdin
		if args.PubFile != nil {
			var err error
			reader, err = os.Open(*args.PubFile)
			failOnError(err, "error opening "+*args.PubFile, os.Exit)
		}
		readerFunc := createMessageReaderFunc(args.JSONFormat, reader)
		startPublishMode(args.PubAmqpURI,
			args.PubExchange,
			args.PubRoutingKey,
			args.InsecureTLS,
			readerFunc)

	case TapMode:
		// signalChannel receives ctrl+C/interrput signal
		signalChannel := make(chan os.Signal, 1)
		// messageReceiveFunc receives the tapped messages, prints
		// and optionally saves them.
		messageReceiveFunc := createMessageReceiveFunc(NewColorableWriter(os.Stdout),
			args.JSONFormat, args.SaveDir, args.NoColor)
		startTapMode(args.TapConfig, args.InsecureTLS, messageReceiveFunc, signalChannel)

	case InfoMode:
		printBrokerInfoConfig := PrintBrokerInfoConfig{
			ShowStats:           args.ShowStats,
			ShowConsumers:       args.ShowConsumers,
			ShowDefaultExchange: args.ShowDefaultExchange,
			NoColor:             args.NoColor}
		startInfoMode(args.APIURI,
			rabtap.NewRabbitHTTPClient(args.APIURI,
				&tls.Config{InsecureSkipVerify: args.InsecureTLS}),
			printBrokerInfoConfig)
	}
}
