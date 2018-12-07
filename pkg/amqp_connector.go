// Copyright (C) 2017 Jan Delgado

package rabtap

import (
	"context"
	"crypto/tls"
	"errors"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	reconnectDelayTime = 2000 * time.Millisecond
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

type connectionState int

const (
	stateConnecting connectionState = iota
	stateConnected
	stateClosed
)

// An AmqpWorkerFunc does the actual work after the connection is established.
// If the worker returns true, the caller should re-connect to the broker.  If
// the worker returne false, the caller should finish its processing.  The
// worker must return with NoReconnect if a ShutdownMessage is received via
// shutdownChan, otherwise with Reconnect.
type AmqpWorkerFunc func(conn *amqp.Connection, controlChan chan ControlMessage) ReconnectAction

// ControlMessage contols the amqp-worker.
type ControlMessage int

const (
	// ShutdownMessage signals shutdown. Worker should perform cleanup operations
	shutdownMessage ControlMessage = iota
	// ReconnectMessage signals that the connection to the broker is re-established
	reconnectMessage
)

var errConnectionClosed = errors.New("connection closed")

// IsReconnect returns true if the given control message is a reconnect message
func (s ControlMessage) IsReconnect() bool {
	return s == reconnectMessage
}

// AmqpConnector manages the connection to the amqp broker and automatically
// reconnects after connections losses
// TODO add amqp channel, would simplify code using AmqpConnector.
type AmqpConnector struct {
	logger         logrus.StdLogger
	uri            string
	tlsConfig      *tls.Config
	firstTry       bool
	connection     *amqp.Connection
	connected      *atomic.Value
	controlChan    chan ControlMessage // signal to worker to shutdown/reconnect
	workerFinished chan error          // worker signals result of shutdown
}

// NewAmqpConnector creates a new AmqpConnector object.
func NewAmqpConnector(uri string, tlsConfig *tls.Config, logger logrus.StdLogger) *AmqpConnector {
	connected := &atomic.Value{}
	connected.Store(stateClosed)
	return &AmqpConnector{
		logger:         logger,
		uri:            uri,
		tlsConfig:      tlsConfig,
		firstTry:       true,
		connected:      connected,
		controlChan:    make(chan ControlMessage, 5),
		workerFinished: make(chan error, 5)}
}

// Connected returns true if the connection is established, else false.
func (s *AmqpConnector) Connected() bool {
	return s.connected.Load().(connectionState) == stateConnected
}

// Try to connect to the RabbitMQ server as  long as it takes to establish a
// connection. Will be interrupted by any message on the control channel.
func (s *AmqpConnector) redial() (*amqp.Connection, error) {
	s.connection = nil
	s.connected.Store(stateConnecting)
	for {
		// loop can be interrupted by call to Close()
		select {
		case <-s.controlChan:
			s.connected.Store(stateClosed)
			return nil, errConnectionClosed
		default:
		}

		s.logger.Printf("(re-)connecting to %s", s.uri)
		conn, err := amqp.DialTLS(s.uri, s.tlsConfig)

		if err == nil {
			s.firstTry = false
			s.logger.Printf("connection established.")
			s.connection = conn
			s.connected.Store(stateConnected)
			return conn, nil
		}

		s.logger.Printf("error connecting to broker %+v", err)

		if err != nil && s.firstTry {
			s.logger.Printf("failed on first connection attempt - not retrying")
			s.connected.Store(stateClosed)
			return nil, err
		}

		time.Sleep(reconnectDelayTime)
	}
}

// Connect  (re-)establishes the connection to RabbitMQ broker.
func (s *AmqpConnector) Connect(worker AmqpWorkerFunc) error {
	for {
		// the error channel is used to detect when (re-)connect is needed will
		// be closed by amqp lib when connection is gracefully shut down.
		errorChan := make(chan *amqp.Error, 10)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel() // to prevent go-routine leaking

		// translate amqp notifications (*amqp.Error) to events for the worker
		go func() {
			select {
			case <-ctx.Done():
				return
			case <-errorChan:
				// let the worker know we are re-connecting
				s.controlChan <- reconnectMessage
				return
			}
		}()

		rabbitConn, err := s.redial()
		if err == errConnectionClosed {
			// graceful shutdown by call to Close()
			s.workerFinished <- err
			return nil
		}
		if err != nil {
			// connection could not be established
			s.workerFinished <- err
			return err
		}

		rabbitConn.NotifyClose(errorChan)
		if !worker(rabbitConn, s.controlChan).shouldReconnect() {
			break
		}
		s.shutdown()
	}
	err := s.shutdown()
	s.workerFinished <- err
	return nil
}

func (s *AmqpConnector) shutdown() error {
	err := s.connection.Close()
	s.connected.Store(stateClosed)
	return err
}

// Close closes the connection to the broker.
func (s *AmqpConnector) Close() error {
	if s.connected.Load().(connectionState) == stateClosed {
		return errors.New("already closed")
	}
	s.controlChan <- shutdownMessage
	err := <-s.workerFinished
	return err
}
