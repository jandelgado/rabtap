// Copyright (C) 2017 Jan Delgado
// RabbitMQ wire-tap.

package rabtap

// CreateExchange creates a new echange on the given channel
func CreateExchange(session Session, exchangeName, exchangeType string,
	durable, autoDelete bool) error {

	return session.ExchangeDeclare(
		exchangeName,
		exchangeType,
		durable,
		autoDelete,
		false, // not internal
		false, // wait for response
		nil)
}

// RemoveExchange removes a echange on the given channel
func RemoveExchange(session Session,
	exchangeName string, ifUnused bool) error {
	return session.ExchangeDelete(exchangeName, ifUnused, false /* wait*/)
}
