// Copyright (C) 2017-2019 Jan Delgado

package rabtap

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
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
	logger    *slog.Logger
	url       *url.URL
	tlsConfig *tls.Config
}

// NewAmqpConnector creates a new AmqpConnector object.
func NewAmqpConnector(url *url.URL, tlsConfig *tls.Config, logger *slog.Logger) *AmqpConnector {
	return &AmqpConnector{
		logger:    logger,
		url:       url,
		tlsConfig: tlsConfig,
	}
}

// Connect  (re-)establishes the connection to RabbitMQ broker.
func (s *AmqpConnector) Connect(ctx context.Context, worker AmqpWorkerFunc) error {
	sessions := redial(ctx, s.url, s.tlsConfig, s.logger, FailEarly)
	for session := range sessions {
		s.logger.Debug("waiting for new session", "url", s.url.Redacted())
		sub, more := <-session
		if !more {
			// closed. TODO propagate errors from redial()
			return errors.New("session factory closed")
		}
		s.logger.Debug("got new amqp session ...")
		action, err := worker(ctx, sub)
		if err != nil && action.shouldReconnect() {
			s.logger.Error("worker failed", "error", err)
		}
		if !action.shouldReconnect() {
			if errClose := sub.Connection.Close(); errClose != nil {
				return errors.Join(err, fmt.Errorf("connection close failed: %w", errClose))
			}
			return err
		}
	}
	return nil
}
