// Copyright (C) 2017-2019 Jan Delgado
// RabbitMQ wire-tap. Functions to hook to exchanges and to keep the
// connection to the broker alive in case of connection errors.

// Package rabtap provides funtionalities to tap to RabbitMQ exchanges using
// exchange-to-exchange bindings.
package rabtap

import (
	"context"
	"crypto/tls"

	uuid "github.com/google/uuid"
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
	config := AmqpSubscriberConfig{Exclusive: true, AutoAck: true}
	return &AmqpTap{
		AmqpSubscriber: NewAmqpSubscriber(config, uri, tlsConfig, logger)}
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
func (s *AmqpTap) EstablishTap(ctx context.Context, exchangeConfigList []ExchangeConfiguration,
	tapCh TapChannel) error {
	err := s.connection.Connect(ctx, s.createWorkerFunc(exchangeConfigList, tapCh))
	// if err != nil {
	//     tapCh <- NewTapMessage(nil, err, time.Now())
	// }
	return err
}

func (s *AmqpTap) createWorkerFunc(
	exchangeConfigList []ExchangeConfiguration,
	tapCh TapChannel) AmqpWorkerFunc {

	return func(ctx context.Context, session Session) (ReconnectAction, error) {

		amqpChs, err := s.setupTapsForExchanges(session, exchangeConfigList, tapCh)
		if err != nil {
			return doNotReconnect, err
		}
		fanin := NewFanin(amqpChs)
		defer func() { _ = fanin.Stop() }()

		action := s.messageLoop(ctx, tapCh, fanin)

		return action, nil
	}
}

func (s *AmqpTap) setupTapsForExchanges(
	session Session,
	exchangeConfigList []ExchangeConfiguration,
	tapCh TapChannel) ([]interface{}, error) {

	// cleanup left-overs in case of re-connect
	//_ = s.cleanup(session)

	var channels []interface{}

	// establish each tap with its own channel. TODO
	for _, exchangeConfig := range exchangeConfigList {
		exchange, queue, err := s.setupTap(session, exchangeConfig)
		if err != nil {
			// pass err to client, can decide to close tap.
			// TODO log instead
			//tapCh <- NewTapMessage(nil, err, time.Now())
			return channels, err
		}
		msgCh, err := s.consumeMessages(session, queue)
		if err != nil {
			// pass err to client, can decide to close tap.
			// TODO log instead
			//tapCh <- NewTapMessage(nil, err, time.Now())
			return channels, err
		}
		channels = append(channels, msgCh)
		// store created exchanges and queues for later cleanup
		s.exchanges = append(s.exchanges, exchange)
		s.queues = append(s.queues, queue)
	}
	return channels, nil
}

// setupTap sets up the a single tap to an exchange.  We create an
// exchange-to-exchange binding where the bound exchange (of type fanout) will
// receive all messages published to the original exchange. Returns
// (tapExchangeName, tapQueueName, error)
func (s *AmqpTap) setupTap(session Session,
	exchangeConfig ExchangeConfiguration) (string, string, error) {

	id := uuid.Must(uuid.NewRandom()).String()
	tapExchange := getTapExchangeNameForExchange(exchangeConfig.Exchange, id[:12])
	tapQueue := getTapQueueNameForExchange(exchangeConfig.Exchange, id[:12])

	var err error

	err = s.createExchangeToExchangeBinding(session,
		exchangeConfig.Exchange,
		exchangeConfig.BindingKey,
		tapExchange)
	if err != nil {
		return "", "", err
	}

	err = CreateQueue(session, tapQueue,
		false, // non durable
		true,  // auto delete
		true)  // exclusive
	if err != nil {
		return "", "", err
	}

	if err = s.bindQueueToExchange(session, tapExchange,
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
func (s *AmqpTap) createExchangeToExchangeBinding(session Session,
	exchangeName, bindingKey, tapExchangeName string) error {

	var err error

	if err := CreateExchange(session, tapExchangeName, amqp.ExchangeFanout,
		false /* nondurable*/, true); err != nil {
		return err
	}

	// TODO when tapping to headers exchange the bindingKey must be
	// "translated" into an amqp.Table{} struct. Right know we are always
	// seeing all messages sent to exchanges of type headers.
	if err = session.ExchangeBind(
		tapExchangeName, // destination
		bindingKey,
		exchangeName, // source
		false,        // wait for response
		amqp.Table{}); err != nil {

		s.logger.Printf("tap: bind to exchange %s failed with %v", exchangeName, err)
		// bind failed, so we must also delete our tap-exchange since it
		// will not be auto-deleted when no binding exists.
		session.NewChannel()
		err2 := RemoveExchange(session, tapExchangeName, false)
		s.logger.Printf("delete exchange: %v", err2)
		return err
	}
	return nil
}

// func (s *AmqpTap) cleanup(session Session) error {

//     // delete exchanges manually, since auto delete seems to
//     // not work in all cases. TODO recheck & simplify
//     for _, exchange := range s.exchanges {
//         if err := RemoveExchange(session, exchange, false); err != nil {
//             //if ch, err = conn.Channel(); err != nil {
//             return err
//             //}
//         }
//     }
//     s.exchanges = []string{}

//     // delete any created queues since auto delete seems to
//     // not work in all cases
//     for _, queue := range s.queues {
//         if err := RemoveQueue(session, queue, false, false); err != nil {
//             // TODO channel needs to re-opened after error
//             //if ch, err = conn.Channel(); err != nil {
//             return err
//             //}
//         }
//     }
//     s.queues = []string{}
//     return nil
// }

func (s *AmqpTap) bindQueueToExchange(session Session,
	exchangeName, bindingKey, queueName string) error {

	var err error

	err = BindQueueToExchange(session, queueName, bindingKey, exchangeName)
	if err != nil {
		return err
	}
	return nil
}
