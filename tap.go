// Copyright (C) 2017 Jan Delgado
// RabbitMQ wire-tap. Functions to hook to exchanges and to keep the
// connection to the broker alive in case of connection errors.

// Package rabtap provides funtionalities to tap to RabbitMQ exchanges using
// exchange-to-exchange bindings.
package rabtap

import (
	"crypto/tls"
	"reflect"

	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// AmqpTap allows to tap to an RabbitMQ exchange.
type AmqpTap struct {
	connection *AmqpConnector
	logger     logrus.StdLogger
	tlsConfig  *tls.Config
	exchanges  []string // list of tap-exchanges created
	queues     []string // list of tap-queue created
}

// NewAmqpTap returns a new AmqpTap object associated with the RabbitMQ
// broker denoted by the uri parameter.
func NewAmqpTap(uri string, tlsConfig *tls.Config, logger logrus.StdLogger) *AmqpTap {
	return &AmqpTap{
		connection: NewAmqpConnector(uri, logger),
		tlsConfig:  tlsConfig,
		logger:     logger}
}

// TapMessage objects are passed through a tapChannel from tap to client
// either AmqpMessage or Error is set
type TapMessage struct {
	AmqpMessage *amqp.Delivery
	Error       error
}

// TapChannel is a channel for *TapMessage objects
type TapChannel chan *TapMessage

func getTapExchangeNameForExchange(exchange, postfix string) string {
	return "__tap-exchange-for-" + exchange + "-" + postfix
}

func getTapQueueNameForExchange(exchange, postfix string) string {
	return "__tap-queue-for-" + exchange + "-" + postfix
}

// EstablishTap sets up the connection to the broker and sets up
// the tap, which is bound to the provided consumer function. Typicall
// this function is run as a go-routine.
func (s *AmqpTap) EstablishTap(exchangeConfigList []ExchangeConfiguration,
	tapChannel TapChannel) {
	s.connection.Connect(s.tlsConfig,
		s.createWorkerFunc(exchangeConfigList, tapChannel))
}

// Close closes the connection to the broker and ends tapping. Returns result
// of amqp.Connection.Close() operation.
func (s *AmqpTap) Close() error {
	return s.connection.Close()
}

// Connected returns true if the tap is connected to an exchange, otherwise
// false
func (s *AmqpTap) Connected() bool {
	return s.connection.Connected()
}

// (re-)establish the connection to RabbitMQ in case the connection has died.
// this function is run in a go-routine. after the connection is established
// a channel is created and the list of provided exchanges is wire-tapped.
// To start the first connection process,  send an amqp.ErrClosed message
// through the errorChannel. See EstablishTap() for example.
func (s *AmqpTap) createWorkerFunc(
	exchangeConfigList []ExchangeConfiguration,
	tapChannel TapChannel) AmqpWorkerFunc {

	return func(rabbitConn *amqp.Connection, controlChan chan ControlMessage) bool {

		// clear any previously created exchanges and queues (re-connect case)
		s.cleanup(rabbitConn)

		// establish each tap with its own channel. The errorChannel
		// will also be monitotred
		messageChannels := []reflect.SelectCase{}
		messageChannels = append(messageChannels,
			reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(controlChan)})
		for _, exchangeConfig := range exchangeConfigList {
			exchange, queue, msgs, err := s.setupTap(rabbitConn, exchangeConfig)
			if err != nil {
				// pass err to client, can decide to close tap.
				tapChannel <- &TapMessage{nil, err}
				break
			}
			// store created exchanges and queues for later cleanup
			messageChannels = append(messageChannels,
				reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(msgs)})
			s.exchanges = append(s.exchanges, exchange)
			s.queues = append(s.queues, queue)
		}

		// process array of message channels and the error channel, see
		// https://play.golang.org/p/8zwvSk4kjx
		remaining := len(messageChannels)
		for {
			chosen, message, ok := reflect.Select(messageChannels)
			if !ok {
				// The chosen channel has been closed, so zero
				// out the channel to disable the case (happens on normal
				// shutdown)
				remaining--
				messageChannels[chosen].Chan = reflect.ValueOf(nil)
				if remaining == 1 {
					// all message and the error (go-)channels were closed
					// (only the shutdown channel is remaining)
					// -> reconnect
					return true
				}
				continue
			}
			switch message.Interface().(type) {

			case amqp.Delivery: // message received on message channels
				amqpMessage, _ := message.Interface().(amqp.Delivery)
				tapChannel <- &TapMessage{&amqpMessage, nil}

			case ControlMessage: // mesage received on control channel
				controlMessage := message.Interface().(ControlMessage)
				switch controlMessage {
				case ShutdownMessage:
					err := s.cleanup(rabbitConn)
					if err != nil {
						s.logger.Printf("error while shutdown cleaning up tap: %s", err) // TODO WARN
					}
					return false // do not reconnect
				case ReconnectMessage:
					s.logger.Printf("ending worker due to reconnect")
					return true // force caller to reconnect
				}
			}
		}
	}
}

// setupTap sets the actual tap after the connection was established.
// We create an exchange-to-exchange binding where the bound
// exchange (of type fanout) will receive all messages published to the
// original exchange. Returns (tapExchangeName, tapQueueName, <- chan
// amqp.Delivery, error)
func (s *AmqpTap) setupTap(conn *amqp.Connection,
	exchangeConfig ExchangeConfiguration) (string, string, <-chan amqp.Delivery, error) {

	id := uuid.NewV4().String()
	tapExchange := getTapExchangeNameForExchange(exchangeConfig.Exchange, id[:12])
	tapQueue := getTapQueueNameForExchange(exchangeConfig.Exchange, id[:12])

	err := s.tapToExchange(conn,
		exchangeConfig.Exchange,
		exchangeConfig.BindingKey,
		tapExchange,
		tapQueue)

	if err != nil {
		return "", "", nil, err
	}

	// create RabbitMQ channel per tap
	channel, err := conn.Channel()
	if err != nil {
		return "", "", nil, err
	}

	msgs, err := channel.Consume(
		tapQueue,
		"__tap-consumer-"+uuid.NewV4().String()[:8],
		true,  // auto-ack
		true,  // exclusive
		false, // no-local
		false, // wait
		nil,   // args
	)
	if err != nil {
		return "", "", nil, err
	}
	return tapExchange, tapQueue, msgs, nil
}

func (s *AmqpTap) cleanup(conn *amqp.Connection) error {

	// delete s.queues, s.exchanges manually, since auto delete seems to
	// not work in all cases
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	for _, queue := range s.queues {
		_, err := channel.QueueDelete(queue, false, false, false)
		if err != nil {
			// channel needs to re-opened after error
			channel, _ = conn.Channel()
		}
	}
	s.queues = []string{}
	for _, exchange := range s.exchanges {
		err := channel.ExchangeDelete(exchange, false, false)
		if err != nil {
			channel, _ = conn.Channel()
		}
	}
	s.exchanges = []string{}
	return nil
}

// tapToExchange creates a new exchange 'tapExchangeName' with a queue
// and bind the exchange to the existing exchange 'exchange'. By binding
// one exchange to another, we receive all messages published to to
// original exchange.
// The provided binding depends on the type of the observed exchange
// and must be set to
// - '#' on topic exchanges
// - a binding-key on direct exchanges (i.e. no wildcards)
// - '' on fanout or headers exchanges
// On errors delete prior created exchanges and/or queues to make sure
// that there are no leftovers lying around on the broker.
func (s *AmqpTap) tapToExchange(conn *amqp.Connection,
	exchangeName, bindingKey, tapExchangeName, tapQueueName string) error {

	channel, err := conn.Channel()
	if err != nil {
		return err
	}

	err = channel.ExchangeDeclare(
		tapExchangeName,
		amqp.ExchangeFanout,
		false, // non durable
		true,  // auto delete
		true,  // internal
		false, // wait for response
		nil)
	if err != nil {
		return err
	}
	// TODO when tapping to headers exchange the bindingKey must be "translated"
	//      into an amqp.Table{} struct. Right we are always seeing all messages
	//      sent to exchanges of type headers.
	err = channel.ExchangeBind(
		tapExchangeName, // destination
		bindingKey,
		exchangeName, // source
		false,        // wait for response
		amqp.Table{})
	if err != nil {
		// bind failed, so we can also delete our tap-exchange
		channel, _ = conn.Channel()
		defer channel.Close()
		channel.ExchangeDelete(tapExchangeName, false, false)
		return err
	}
	_, err = channel.QueueDeclare(
		tapQueueName,
		false, // non durable
		true,  // auto delete
		true,  // exclusive
		false, // wait for response
		nil)
	if err != nil {
		channel, _ = conn.Channel()
		defer channel.Close()
		channel.ExchangeDelete(tapExchangeName, false, false)
		return err
	}
	err = channel.QueueBind(
		tapQueueName,    // queue name
		"",              // binding key not needed for fanout
		tapExchangeName, // exchange
		false,           // wait for response
		nil)
	if err != nil {
		channel, _ = conn.Channel()
		defer channel.Close()
		channel.QueueDelete(tapQueueName, false, false, false)
		channel.ExchangeDelete(tapExchangeName, false, false)
		return err
	}
	channel.Close()
	return nil
}
