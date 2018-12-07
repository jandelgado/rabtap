// Copyright (C) 2017 Jan Delgado
// RabbitMQ wire-tap. Functions to hook to exchanges and to keep the
// connection to the broker alive in case of connection errors.

// Package rabtap provides funtionalities to tap to RabbitMQ exchanges using
// exchange-to-exchange bindings.
package rabtap

import (
	"crypto/tls"

	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// AmqpTap allows to tap to an RabbitMQ exchange.
type AmqpTap struct {
	*AmqpSubscriber
	exchanges []string // list of tap-exchanges created
	queues    []string // list of tap-queues created
}

// NewAmqpTap returns a new AmqpTap object associated with the RabbitMQ
// broker denoted by the uri parameter.
func NewAmqpTap(uri string, tlsConfig *tls.Config, logger logrus.StdLogger) *AmqpTap {
	return &AmqpTap{
		AmqpSubscriber: NewAmqpSubscriber(uri, tlsConfig, logger)}
}

func getTapExchangeNameForExchange(exchange, postfix string) string {
	return "__tap-exchange-for-" + exchange + "-" + postfix
}

func getTapQueueNameForExchange(exchange, postfix string) string {
	return "__tap-queue-for-" + exchange + "-" + postfix
}

// EstablishTap sets up the connection to the broker and sets up
// the tap, which is bound to the provided consumer function. Typically
// this function is run as a go-routine.
func (s *AmqpTap) EstablishTap(exchangeConfigList []ExchangeConfiguration,
	tapCh TapChannel) error {
	err := s.connection.Connect(s.createWorkerFunc(exchangeConfigList, tapCh))
	if err != nil {
		tapCh <- &TapMessage{nil, err}
	}
	return err
}

func (s *AmqpTap) createWorkerFunc(
	exchangeConfigList []ExchangeConfiguration,
	tapCh TapChannel) AmqpWorkerFunc {

	return func(rabbitConn *amqp.Connection, controlCh chan ControlMessage) ReconnectAction {
		amqpChs := s.setupTapsForExchanges(rabbitConn, exchangeConfigList, tapCh)
		fanin := NewFanin(amqpChs)
		defer func() { fanin.Stop() }()

		action := s.messageLoop(tapCh, fanin, controlCh)

		if !action.shouldReconnect() {
			err := s.cleanup(rabbitConn)
			if err != nil {
				s.logger.Printf("error while shutdown cleaning up tap: %s", err) // TODO WARN
			}
		}
		return action
	}
}

func (s *AmqpTap) setupTapsForExchanges(
	rabbitConn *amqp.Connection,
	exchangeConfigList []ExchangeConfiguration,
	tapCh TapChannel) []interface{} {

	// cleanup left-overs in case of re-connect
	s.cleanup(rabbitConn)

	var channels []interface{}

	// establish each tap with its own channel.
	for _, exchangeConfig := range exchangeConfigList {
		exchange, queue, err := s.setupTap(rabbitConn, exchangeConfig)
		if err != nil {
			// pass err to client, can decide to close tap.
			tapCh <- &TapMessage{nil, err}
			break
		}
		msgCh, err := s.consumeMessages(rabbitConn, queue)
		if err != nil {
			// pass err to client, can decide to close tap.
			tapCh <- &TapMessage{nil, err}
			break
		}
		channels = append(channels, msgCh)
		// store created exchanges and queues for later cleanup
		s.exchanges = append(s.exchanges, exchange)
		s.queues = append(s.queues, queue)
	}
	return channels
}

// setupTap sets up the a single tap to an exchange.  We create an
// exchange-to-exchange binding where the bound exchange (of type fanout) will
// receive all messages published to the original exchange. Returns
// (tapExchangeName, tapQueueName, error)
func (s *AmqpTap) setupTap(conn *amqp.Connection,
	exchangeConfig ExchangeConfiguration) (string, string, error) {

	id := uuid.NewV4().String()
	tapExchange := getTapExchangeNameForExchange(exchangeConfig.Exchange, id[:12])
	tapQueue := getTapQueueNameForExchange(exchangeConfig.Exchange, id[:12])

	var ch *amqp.Channel
	var err error

	if ch, err = conn.Channel(); err != nil {
		return "", "", err
	}
	defer ch.Close()

	err = s.createExchangeToExchangeBinding(conn,
		exchangeConfig.Exchange,
		exchangeConfig.BindingKey,
		tapExchange)
	if err != nil {
		return "", "", err
	}

	err = CreateQueue(ch, tapQueue,
		false, // non durable
		true,  // auto delete
		true)  // exclusive
	if err != nil {
		return "", "", err
	}

	if err = s.bindQueueToExchange(conn, tapExchange,
		exchangeConfig.BindingKey, tapQueue); err != nil {
		return "", "", err
	}
	return tapExchange, tapQueue, nil
}

// createExchangeToExchangeBinding creates a new exchange 'tapExchangeName'
// with a queue and bind the exchange to the existing exchange 'exchange'. By
// binding one exchange to another, we receive all messages published to to
// original exchange.
// The provided binding depends on the type of the observed exchange
// and must be set to
// - '#' on topic exchanges
// - a binding-key on direct exchanges (i.e. no wildcards)
// - '' on fanout or headers exchanges
// On errors delete prior created exchanges and/or queues to make sure
// that there are no leftovers lying around on the broker.
// TODO error handling must be improved - does not work if connection is lost
func (s *AmqpTap) createExchangeToExchangeBinding(conn *amqp.Connection,
	exchangeName, bindingKey, tapExchangeName string) error {

	var ch *amqp.Channel
	var err error

	if ch, err = conn.Channel(); err != nil {
		return err
	}
	defer ch.Close()

	if err := CreateExchange(ch, tapExchangeName, amqp.ExchangeFanout,
		false /* nondurable*/, true); err != nil {
		return err
	}

	// TODO when tapping to headers exchange the bindingKey must be "translated"
	//      into an amqp.Table{} struct. Right we are always seeing all messages
	//      sent to exchanges of type headers.
	if err = ch.ExchangeBind(
		tapExchangeName, // destination
		bindingKey,
		exchangeName, // source
		false,        // wait for response
		amqp.Table{}); err != nil {

		// bind failed, so we can also delete our tap-exchange
		ch, _ = conn.Channel()
		defer ch.Close()
		RemoveExchange(ch, tapExchangeName, false)
		return err
	}
	return nil
}

func (s *AmqpTap) cleanup(conn *amqp.Connection) error {

	// delete exchanges manually, since auto delete seems to
	// not work in all cases. TODO recheck & simplify
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	for _, exchange := range s.exchanges {
		if err := RemoveExchange(ch, exchange, false); err != nil {
			if ch, err = conn.Channel(); err != nil {
				return err
			}
		}
	}
	s.exchanges = []string{}

	// delete any created queues since auto delete seems to
	// not work in all cases
	for _, queue := range s.queues {
		if err := RemoveQueue(ch, queue, false, false); err != nil {
			// channel needs to re-opened after error
			if ch, err = conn.Channel(); err != nil {
				return err
			}
		}
	}
	s.queues = []string{}
	return nil
}

func (s *AmqpTap) bindQueueToExchange(conn *amqp.Connection,
	exchangeName, bindingKey, queueName string) error {

	var ch *amqp.Channel
	var err error

	if ch, err = conn.Channel(); err != nil {
		return err
	}

	err = BindQueueToExchange(ch, queueName, bindingKey, exchangeName)
	if err != nil {
		return err
	}
	return nil
}
