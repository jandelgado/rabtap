// Copyright (C) 2017-2019 Jan Delgado

package rabtap

import (
	"context"
	"crypto/tls"
	"errors"

	"github.com/sirupsen/logrus"
)

// ReconnectAction signals if connection should be reconnected or not.
type ReconnectAction int

const (
	// doNotReconnect signals caller of worker func not to reconnect
	doNotReconnect ReconnectAction = iota
	// doReconnect signals caller of worker func to reconnect
	doReconnect
)

func (s ReconnectAction) shouldReconnect() bool {
	return s == doReconnect
}

// An AmqpWorkerFunc does the actual work after the connection is established.
// If the worker returns true, the caller should re-connect to the broker.  If
// the worker returne false, the caller should finish its processing.  The
// worker must return with NoReconnect if a ShutdownMessage is received via
// shutdownChan, otherwise with Reconnect.
type AmqpWorkerFunc func(ctx context.Context, session Session) (ReconnectAction, error)

// AmqpConnector manages the connection to the amqp broker and automatically
// reconnects after connections losses
type AmqpConnector struct {
	logger    logrus.StdLogger
	uri       string
	tlsConfig *tls.Config
}

// NewAmqpConnector creates a new AmqpConnector object.
func NewAmqpConnector(uri string, tlsConfig *tls.Config, logger logrus.StdLogger) *AmqpConnector {
	return &AmqpConnector{
		logger:    logger,
		uri:       uri,
		tlsConfig: tlsConfig}
}

// Connect  (re-)establishes the connection to RabbitMQ broker.
func (s *AmqpConnector) Connect(ctx context.Context, worker AmqpWorkerFunc) error {

	sessions := redial(ctx, s.uri, s.tlsConfig, s.logger, FailEarly)
	for session := range sessions {
		s.logger.Printf("waiting for new session ...")
		sub, more := <-session
		if !more {
			// closed. TODO propagate errors from redial()
			return errors.New("initial connection failed")
		}
		s.logger.Printf("got new amqp session ...")
		action, err := worker(ctx, sub)
		if !action.shouldReconnect() {
			return err
		}
	}
	return nil
}
