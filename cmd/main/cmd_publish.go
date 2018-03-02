// Copyright (C) 2017 Jan Delgado

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"os"

	"github.com/jandelgado/rabtap/pkg"
	"github.com/streadway/amqp"
)

// CmdPublishArg contains arguments for the publish command
type CmdPublishArg struct {
	amqpURI             string
	tlsConfig           *tls.Config
	exchange            string
	routingKey          string
	readNextMessageFunc MessageReaderFunc
}

// MessageReaderFunc provides messages that can be sent to an exchange. If no
// more messages are available io.EOF is returned as error.
type MessageReaderFunc func() (amqp.Publishing, error)

// SignalChannel transports os.Signal objects like e.g. os.Interrupt
// type SignalChannel chan os.Signal

// publishMessage publishes a single message on the given exchange with the
// provided routingkey
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

// readSingleMessageFromRawFile reads a single messages from the given io.Reader
// which is typically stdin or a file. On subsequent calls, it returns io.EOF.
func readSingleMessageFromRawFile(reader io.Reader) (amqp.Publishing, error) {
	buf := new(bytes.Buffer)
	if numRead, err := buf.ReadFrom(reader); err != nil {
		return amqp.Publishing{}, err
	} else if numRead == 0 {
		return amqp.Publishing{}, io.EOF
	}
	return amqp.Publishing{Body: buf.Bytes()}, nil
}

// readNextMessageFromJSONStream reads JSON messages from the given decoder as long
// as there are messages available.
func readNextMessageFromJSONStream(decoder *json.Decoder) (amqp.Publishing, error) {
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
		return func() (amqp.Publishing, error) {
			return readNextMessageFromJSONStream(decoder)
		}
	}
	return func() (amqp.Publishing, error) {
		return readSingleMessageFromRawFile(reader)
	}
}

// publishMessages reads messages with the provided readNextMessageFunc and
// publishes the messages to the given exchange.
func publishMessageStream(publishChannel rabtap.PublishChannel,
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

// cmdPublish reads messages with the provied readNextMessageFunc and
// publishes the messages to the given exchange.
func cmdPublish(cmd CmdPublishArg) {

	log.Debugf("publishing message(s) to exchange %s with routingkey %s",
		cmd.exchange, cmd.routingKey)
	publisher := rabtap.NewAmqpPublish(cmd.amqpURI, cmd.tlsConfig, log)
	defer publisher.Close()
	publishChannel := make(rabtap.PublishChannel)
	go publisher.EstablishConnection(publishChannel)
	publishMessageStream(publishChannel, cmd.exchange, cmd.routingKey,
		cmd.readNextMessageFunc)
}
