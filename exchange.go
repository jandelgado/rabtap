// Copyright (C) 2017 Jan Delgado
// RabbitMQ wire-tap.

package rabtap

import "github.com/streadway/amqp"

// CreateExchange creates a new echange on the given channel
func CreateExchange(channel *amqp.Channel, exchangeName, exchangeType string,
	durable, autoDelete bool) error {

	return channel.ExchangeDeclare(
		exchangeName,
		exchangeType,
		durable,
		autoDelete,
		false, // not internal
		false, // wait for response
		nil)
}

// RemoveExchange removes a echange on the given channel
func RemoveExchange(channel *amqp.Channel,
	exchangeName string, ifUnused bool) error {
	return channel.ExchangeDelete(exchangeName, ifUnused, false /* wait*/)
}
