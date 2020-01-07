// Copyright (C) 2017 Jan Delgado

package main

import (
	"context"
	"crypto/tls"
	"io"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/streadway/amqp"
	"golang.org/x/sync/errgroup"
)

// CmdPublishArg contains arguments for the publish command
type CmdPublishArg struct {
	amqpURI    string
	tlsConfig  *tls.Config
	exchange   *string
	routingKey *string
	readerFunc MessageReaderFunc
}

// publishMessage publishes a single message on the given exchange with the
// provided routingkey
func publishMessage(publishChannel rabtap.PublishChannel,
	exchange, routingKey string,
	amqpPublishing amqp.Publishing) {

	log.Debugf("publishing message to exchange '%s' with routing key '%s'",
		exchange, routingKey)

	publishChannel <- &rabtap.PublishMessage{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Publishing: &amqpPublishing}
}

// selectOptionalOrDefault returns either an optional string, if set, or
// a default value.
func selectOptionalOrDefault(optionalStr *string, defaultStr string) string {
	if optionalStr != nil {
		return *optionalStr
	}
	return defaultStr
}

// publishMessageStream publishes messages from the provided message stream
// provided by readNextMessageFunc. When done closes the publishChannel
func publishMessageStream(publishChannel rabtap.PublishChannel,
	optExchange, optRoutingKey *string, readNextMessageFunc MessageReaderFunc) error {
	for {
		msg, more, err := readNextMessageFunc()
		switch err {
		case io.EOF:
			close(publishChannel)
			return nil
		case nil:
			routingKey := selectOptionalOrDefault(optRoutingKey, msg.RoutingKey)
			exchange := selectOptionalOrDefault(optExchange, msg.Exchange)
			publishMessage(publishChannel, exchange, routingKey, msg.ToAmqpPublishing())
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

	errChan := make(chan error)
	publisher := rabtap.NewAmqpPublish(cmd.amqpURI, cmd.tlsConfig, log)
	publishChannel := make(rabtap.PublishChannel)

	go func() {
		// runs as long as readerFunc returns messages. Unfortunately, we
		// can not stop a blocking read on a file like we do with channels
		// and select. So we don't put the goroutine in the error group to
		// avoid blocking when e.g. the user presses CTRL+S and then CTRL+C.
		// TODO come up with better solution
		errChan <- publishMessageStream(publishChannel, cmd.exchange,
			cmd.routingKey, cmd.readerFunc)
	}()

	g.Go(func() error {
		return publisher.EstablishConnection(ctx, publishChannel)
	})

	if err := g.Wait(); err != nil {
		return err
	}

	select {
	case err := <-errChan:
		return err
	default:
	}

	return nil
}
