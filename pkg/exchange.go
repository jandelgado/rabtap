// rabtap - exchange management
// Copyright (C) 2017-2021 Jan Delgado

package rabtap

import "github.com/streadway/amqp"

// CreateExchange creates a new echange on the given channel
func CreateExchange(session Session, exchangeName, exchangeType string,
	durable, autoDelete bool, args amqp.Table) error {

	return session.ExchangeDeclare(
		exchangeName,
		exchangeType,
		durable,
		autoDelete,
		false, // not internal
		false, // wait for response
		args)
}

// RemoveExchange removes a echange on the given channel
func RemoveExchange(session Session,
	exchangeName string, ifUnused bool) error {
	return session.ExchangeDelete(exchangeName, ifUnused, false /* wait*/)
}
