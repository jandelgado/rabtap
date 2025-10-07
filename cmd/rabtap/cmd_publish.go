// publish messages
// Copyright (C) 2017-2019 Jan Delgado

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/errgroup"
)

// CmdPublishArg contains arguments for the publish command
type CmdPublishArg struct {
	amqpURL    *url.URL
	tlsConfig  *tls.Config
	exchange   *string
	routingKey *string
	headers    rabtap.KeyValueMap
	source     MessageSource
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
	routing rabtap.Routing,
	amqpPublishing amqp.Publishing) {

	publishChannel <- &rabtap.PublishMessage{
		Routing:    routing,
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

// routingFromMessage creates a Routing from a message and optional defaults
// that can override fields in the message
func routingFromMessage(optExchange, optRoutingKey *string, headers rabtap.KeyValueMap, msg RabtapPersistentMessage) rabtap.Routing {
	routingKey := selectOptionalOrDefault(optRoutingKey, msg.RoutingKey)
	exchange := selectOptionalOrDefault(optExchange, msg.Exchange)
	mergedHeaders := rabtap.MergeTables(msg.Headers, rabtap.ToAMQPTable(headers))
	return rabtap.NewRouting(exchange, routingKey, mergedHeaders)
}

// publishMessageStream publishes messages from the provided message stream
// provided by readNextMessageFunc. When done closes the publishChannel
func publishMessageStream(publishCh rabtap.PublishChannel,
	optExchange *string,
	optRoutingKey *string,
	headers rabtap.KeyValueMap,
	source MessageSource,
	delayFunc DelayFunc) error {

	defer func() {
		close(publishCh)
	}()

	var lastMsg *RabtapPersistentMessage
	for {
		msg, err := source()
		switch err {
		case io.EOF: //  if errors.Is(err, io.EOF)
			return nil
		case nil:
			delayFunc(lastMsg, &msg)

			// the per-message routing key (in case it was read from a json
			// file) can be overriden by the command line, if set.
			routing := routingFromMessage(optExchange, optRoutingKey, headers, msg)

			// during publishing, header information in msg.Header will be overriden
			// by header information in the routing object (if present). The
			// latter are set on the command line using --header K=V options.
			publishMessage(publishCh, routing, msg.ToAmqpPublishing())
			lastMsg = &msg
		default:
			return err
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
func cmdPublish(ctx context.Context, cmd CmdPublishArg, logger *slog.Logger) error {

	g, ctx := errgroup.WithContext(ctx)

	resultCh := make(chan error, 1)
	publisher := rabtap.NewAmqpPublish(cmd.amqpURL,
		cmd.tlsConfig, cmd.mandatory, cmd.confirms, logger)
	publishCh := make(rabtap.PublishChannel)
	errorCh := make(rabtap.PublishErrorChannel)

	delayFunc := func(first, second *RabtapPersistentMessage) {
		if first == nil || second == nil {
			return
		}
		delay := durationBetweenMessages(first, second, cmd.speed, cmd.fixedDelay)
		logger.Info(fmt.Sprintf("publish delay: sleeping for %s", delay))
		select {
		case <-time.After(delay):
		case <-ctx.Done():
		}
	}

	go func() {
		// runs as long as source returns messages. Unfortunately, we
		// can not stop a blocking read on a file like we do with channels
		// and select. So we don't put the goroutine in the error group to
		// avoid blocking when e.g. the user presses CTRL+S and then CTRL+C.
		// TODO find better solution
		resultCh <- publishMessageStream(publishCh, cmd.exchange,
			cmd.routingKey, cmd.headers, cmd.source, delayFunc)
	}()

	g.Go(func() error {
		numPublishErrors := 0
		// log all publishing errors
		for err := range errorCh {
			numPublishErrors++
			logger.Error("publishing error", "error", err)
		}
		if numPublishErrors > 0 {
			return fmt.Errorf("published with errors")
		}
		return nil
	})

	g.Go(func() error {
		err := publisher.EstablishConnection(ctx, publishCh, errorCh)
		logger.Info("Publisher ending")
		close(errorCh)
		return err
	})

	if err := g.Wait(); err != nil {
		return err
	}

	select {
	case err := <-resultCh:
		return err
	default:
	}

	return nil
}
