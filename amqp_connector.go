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

// ControlMessage contols the amqp-worker.
type ControlMessage int

const (
	// ShutdownMessage signals shutdown. Worker should perform cleanup operations
	ShutdownMessage ControlMessage = iota
	// ReconnectMessage signals that the connection to the broker is re-established
	ReconnectMessage
)

// AmqpConnector manages the connection to the amqp broker
type AmqpConnector struct {
	logger         logrus.StdLogger
	uri            string
	connection     *amqp.Connection
	connected      *atomic.Value
	controlChan    chan ControlMessage // signal to worker to shutdown/reconnect
	workerFinished chan error          // worker signals result of shutdown
}

// NewAmqpConnector creates a new AmqpConnector object.
func NewAmqpConnector(uri string, logger logrus.StdLogger) *AmqpConnector {
	connected := &atomic.Value{}
	connected.Store(false)
	return &AmqpConnector{
		uri:            uri,
		logger:         logger,
		connected:      connected,
		controlChan:    make(chan ControlMessage),
		workerFinished: make(chan error)}
}

// Connected returns true if the connection is established, else false.
func (s *AmqpConnector) Connected() bool {
	return s.connected.Load().(bool)
}

// Try to connect to the RabbitMQ server as  long as it takes to establish a
// connection
func (s *AmqpConnector) connect(tlsConfig *tls.Config) *amqp.Connection {
	s.connection = nil
	s.connected.Store(false)
	for {
		s.logger.Printf("(re-)connecting to %s\n", s.uri)
		conn, err := amqp.DialTLS(s.uri, tlsConfig)
		if err == nil {
			s.logger.Printf("connection established.")
			s.connection = conn
			s.connected.Store(true)
			return conn
		}
		s.logger.Printf("error connecting to broker %+v", err)
		time.Sleep(reconnectDelayTime)
	}
}

// An AmqpWorkerFunc does the actual work after the connection is established.
// If the worker returns true, the caller should re-connect to the broker.  If
// the worker returne false, the caller should finish its processing.  The
// worker must return with false if a ShutdownMessage is received via
// shutdownChan.
type AmqpWorkerFunc func(conn *amqp.Connection, controlChan chan ControlMessage) bool

// Connect  (re-)establishes the connection to RabbitMQ broker.
func (s *AmqpConnector) Connect(tlsConfig *tls.Config, worker AmqpWorkerFunc) {

	for {

		// the error channel is used to detect when (re-)connect is needed
		// will be closed by amqp lib when event is sent.
		errorChan := make(chan *amqp.Error)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// translate amqp notifications (*amqp.Error) to events for the worker
		go func() {
			select {
			case <-ctx.Done():
				// prevents go-routine leaking
				return
			case <-errorChan:
				// let the worker know we are re-connecting
				s.controlChan <- ReconnectMessage
				// amqp lib closes channel afterwards.
				return
			}
		}()
		rabbitConn := s.connect(tlsConfig)
		rabbitConn.NotifyClose(errorChan)

		if !worker(rabbitConn, s.controlChan) {
			break
		}
		s.shutdown()
	}
	err := s.shutdown()
	s.workerFinished <- err
}

func (s *AmqpConnector) shutdown() error {
	err := s.connection.Close() // this should be a critical section
	s.connected.Store(false)
	return err
}

// Close closes the connection to the broker.
func (s *AmqpConnector) Close() error {
	if !s.Connected() {
		return errors.New("not connected")
	}
	s.controlChan <- ShutdownMessage
	return <-s.workerFinished
}
