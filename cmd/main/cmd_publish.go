// Copyright (C) 2017 Jan Delgado

package main

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"

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

// MessageReaderFunc provides messages that can be sent to an exchange.
// returns the message to be published, a flag if more messages are to be read,
// and an error.
type MessageReaderFunc func() (amqp.Publishing, bool, error)

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
// which is typically stdin or a file. If reading from stdin, CTRL+D (linux)
// or CTRL+Z (Win) on an empty line terminates the reader.
func readSingleMessageFromRawFile(reader io.Reader) (amqp.Publishing, bool, error) {
	buf, err := ioutil.ReadAll(reader)
	return amqp.Publishing{Body: buf}, false, err
}

// readNextMessageFromJSONStream reads JSON messages from the given decoder as long
// as there are messages available.
func readNextMessageFromJSONStream(decoder *json.Decoder) (amqp.Publishing, bool, error) {
	message := RabtapPersistentMessage{}
	err := decoder.Decode(&message)
	if err != nil {
		return amqp.Publishing{}, false, err
	}
	return message.ToAmqpPublishing(), true, nil
}

// createMessageReaderFunc returns a function that reads messages from the
// the given reader in JSON or raw-format TODO drop boolean param
func createMessageReaderFunc(jsonFormat bool, reader io.Reader) MessageReaderFunc {
	if jsonFormat {
		decoder := json.NewDecoder(reader)
		return func() (amqp.Publishing, bool, error) {
			return readNextMessageFromJSONStream(decoder)
		}
	}
	return func() (amqp.Publishing, bool, error) {
		return readSingleMessageFromRawFile(reader)
	}
}

// publishMessages reads messages with the provided readNextMessageFunc and
// publishes the messages to the given exchange.
func publishMessageStream(publishChannel rabtap.PublishChannel,
	exchange, routingKey string,
	readNextMessageFunc MessageReaderFunc) error {

	for {
		msg, more, err := readNextMessageFunc()
		switch err {
		case io.EOF:
			close(publishChannel)
			return nil
		case nil:
			publishMessage(publishChannel, exchange, routingKey, msg)
		default:
			close(publishChannel)
			return err
		}

		if !more {
			close(publishChannel)
			return nil
		}
	}
}

// cmdPublish reads messages with the provied readNextMessageFunc and
// publishes the messages to the given exchange.
func cmdPublish(cmd CmdPublishArg) error {
	log.Debugf("publishing message(s) to exchange %s with routingkey %s",
		cmd.exchange, cmd.routingKey)
	publisher := rabtap.NewAmqpPublish(cmd.amqpURI, cmd.tlsConfig, log)
	defer publisher.Close()
	publishChannel := make(rabtap.PublishChannel)
	go publishMessageStream(publishChannel, cmd.exchange, cmd.routingKey, cmd.readNextMessageFunc)
	return publisher.EstablishConnection(publishChannel)
}
