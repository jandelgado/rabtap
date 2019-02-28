// Copyright (C) 2017 Jan Delgado
// RabbitMQ wire-tap.

package rabtap

import "github.com/streadway/amqp"

// CreateQueue creates a new queue
// TODO(JD) get rid of bool types
func CreateQueue(channel *amqp.Channel, queueName string,
	durable, autoDelete, exclusive bool) error {

	_, err := channel.QueueDeclare(
		queueName,
		durable,
		autoDelete, // auto delete
		exclusive,  // exclusive
		false,      // wait for response
		nil)
	return err
}

// RemoveQueue removes a queue
func RemoveQueue(channel *amqp.Channel,
	queueName string, ifUnused, ifEmpty bool) error {
	_, err := channel.QueueDelete(queueName, ifUnused, ifEmpty, false /* wait*/)
	return err
}

// PurgeQueue clears a queue. Returns number of elements purged
func PurgeQueue(channel *amqp.Channel, queueName string) (int, error) {
	return channel.QueuePurge(queueName, false /* wait*/)
}

// BindQueueToExchange binds the given queue to the given exchange.
// TODO(JD) support for header based routing
func BindQueueToExchange(channel *amqp.Channel,
	queueName, key, exchangeName string) error {
	return channel.QueueBind(queueName, key, exchangeName, false /* wait */, nil)
}

// UnbindQueueFromExchange unbinds a queue from an exchange
func UnbindQueueFromExchange(channel *amqp.Channel,
	queueName, key, exchangeName string) error {
	return channel.QueueUnbind(queueName, key, exchangeName, nil)
}
