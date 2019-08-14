// Copyright (C) 2017 Jan Delgado

package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/streadway/amqp"
	"golang.org/x/sync/errgroup"
)

// MessageReaderFunc provides messages that can be sent to an exchange.
// returns the message to be published, a flag if more messages are to be read,
// and an error.
type MessageReaderFunc func() (amqp.Publishing, bool, error)

// CmdPublishArg contains arguments for the publish command
type CmdPublishArg struct {
	amqpURI    string
	tlsConfig  *tls.Config
	exchange   string
	routingKey string
	readerFunc MessageReaderFunc
}

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
func createMessageReaderFunc(jsonFormat bool, reader io.ReadCloser) MessageReaderFunc {
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
// publishes the messages to the given exchange. When done closes
// the publishChannel
func publishMessageStream(publishChannel rabtap.PublishChannel,
	exchange, routingKey string, readNextMessageFunc MessageReaderFunc) error {
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
// Termination is a little bit tricky here, since we can not use "select"
// on a File object to stop a blocking read. There are 3 ways publishing
// can be stopped:
// * by an EOF or error on the input file
// * by ctx.Context() signaling cancellation (e.g. ctrl+c)
// * by an initial connection failure to the broker
func cmdPublish(ctx context.Context, cmd CmdPublishArg) error {

	g, ctx := errgroup.WithContext(ctx)

	publisher := rabtap.NewAmqpPublish(cmd.amqpURI, cmd.tlsConfig, log)
	publishChannel := make(rabtap.PublishChannel)

	go func() {
		// runs as long as readerFunc returns messages. Unfortunately, we
		// can not stop a blocking read on a file like we do with channels
		// and select.
		publishMessageStream(publishChannel, cmd.exchange,
			cmd.routingKey, cmd.readerFunc)
	}()

	g.Go(func() error {
		err := publisher.EstablishConnection(ctx, publishChannel)
		return err
	})

	if err := g.Wait(); err != nil {
		log.Errorf("publish failed with %v", err)
		log.Debug("cmd_publish_end with error")
		return err
	}
	log.Debug("cmd_publish_end ok")
	return nil
}
