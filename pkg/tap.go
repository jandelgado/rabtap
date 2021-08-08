// Copyright (C) 2017-2019 Jan Delgado
// RabbitMQ wire-tap. Functions to hook to exchanges and to keep the
// connection to the broker alive in case of connection errors.

package rabtap

import (
	"context"
	"crypto/tls"

	uuid "github.com/google/uuid"
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
func NewAmqpTap(uri string, tlsConfig *tls.Config, logger Logger) *AmqpTap {
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
	return s.connection.Connect(ctx, s.createWorkerFunc(exchangeConfigList, tapCh))
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
		action := amqpMessageLoop(ctx, tapCh, fanin.Ch)
		return action, nil
	}
}

func (s *AmqpTap) setupTapsForExchanges(
	session Session,
	exchangeConfigList []ExchangeConfiguration,
	tapCh TapChannel) ([]interface{}, error) {

	var channels []interface{}

	for _, exchangeConfig := range exchangeConfigList {
		exchange, queue, err := s.setupTap(session, exchangeConfig)
		if err != nil {
			return channels, err
		}
		msgCh, err := s.consumeMessages(session, queue)
		if err != nil {
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

	err := s.createExchangeToExchangeBinding(session,
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

		s.logger.Errorf("tap: bind to exchange %s failed with %v", exchangeName, err)

		// bind failed, so we must also delete our tap-exchange since it
		// will not be auto-deleted when no binding exists.
		// TODO handle errors
		_ = session.NewChannel()
		err2 := RemoveExchange(session, tapExchangeName, false)
		s.logger.Errorf("tap: delete of exchange %s failed with: %v", tapExchangeName, err2)
		return err
	}
	return nil
}

func (s *AmqpTap) bindQueueToExchange(session Session,
	exchangeName, bindingKey, queueName string) error {
	return BindQueueToExchange(session, queueName, bindingKey, exchangeName)
}
