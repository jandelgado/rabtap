// publish messages
// Copyright (C) 2017-2019 Jan Delgado

package main

import (
	"context"
	"crypto/tls"
	"io"
	"net/url"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/streadway/amqp"
	"golang.org/x/sync/errgroup"
)

// CmdPublishArg contains arguments for the publish command
type CmdPublishArg struct {
	amqpURL    *url.URL
	tlsConfig  *tls.Config
	exchange   *string
	routingKey *string
	readerFunc MessageReaderFunc
	speed      float64
	fixedDelay *time.Duration
	confirms   bool
	mandatory  bool
}

type DelayFunc func(first, second *RabtapPersistentMessage)

func multDuration(duration time.Duration, factor float64) time.Duration {
	d := float64(duration.Nanoseconds()) * factor
	return time.Duration(int(d))
}

// durationBetweenMessages calculates the delay to make between the
// publishing of two previously recorded messages.
func durationBetweenMessages(first, second *RabtapPersistentMessage,
	speed float64, fixedDelay *time.Duration) time.Duration {
	if first == nil || second == nil {
		return time.Duration(0)
	}
	if fixedDelay != nil {
		return *fixedDelay
	}
	firstTs := first.XRabtapReceivedTimestamp
	secondTs := second.XRabtapReceivedTimestamp
	delta := secondTs.Sub(firstTs)
	return multDuration(delta, speed)
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
	optExchange, optRoutingKey *string,
	readNextMessageFunc MessageReaderFunc,
	delayFunc DelayFunc) error {
	var lastMsg *RabtapPersistentMessage
	for {
		msg, more, err := readNextMessageFunc()
		switch err {
		case io.EOF:
			close(publishChannel)
			return nil
		case nil:
			delayFunc(lastMsg, &msg)
			routingKey := selectOptionalOrDefault(optRoutingKey, msg.RoutingKey)
			exchange := selectOptionalOrDefault(optExchange, msg.Exchange)
			publishMessage(publishChannel, exchange, routingKey, msg.ToAmqpPublishing())
			lastMsg = &msg
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
	publisher := rabtap.NewAmqpPublish(cmd.amqpURL,
		cmd.tlsConfig, cmd.mandatory, cmd.confirms, log)
	publishChannel := make(rabtap.PublishChannel)

	delayFunc := func(first, second *RabtapPersistentMessage) {
		delay := durationBetweenMessages(first, second, cmd.speed, cmd.fixedDelay)
		log.Infof("sleeping for %s", delay)
		time.Sleep(delay) // TODO make interuptable
	}

	go func() {
		// runs as long as readerFunc returns messages. Unfortunately, we
		// can not stop a blocking read on a file like we do with channels
		// and select. So we don't put the goroutine in the error group to
		// avoid blocking when e.g. the user presses CTRL+S and then CTRL+C.
		// TODO find better solution
		errChan <- publishMessageStream(publishChannel, cmd.exchange,
			cmd.routingKey, cmd.readerFunc, delayFunc)
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
