// Copyright (C) 2017 Jan Delgado
// RabbitMQ wire-tap.

package rabtap

// CreateQueue creates a new queue
// TODO(JD) get rid of bool types
func CreateQueue(session Session, queueName string,
	durable, autoDelete, exclusive bool) error {

	_, err := session.QueueDeclare(
		queueName,
		durable,
		autoDelete, // auto delete
		exclusive,  // exclusive
		false,      // wait for response
		nil)
	return err
}

// RemoveQueue removes a queue
func RemoveQueue(session Session,
	queueName string, ifUnused, ifEmpty bool) error {
	_, err := session.QueueDelete(queueName, ifUnused, ifEmpty, false /* wait*/)
	return err
}

// PurgeQueue clears a queue. Returns number of elements purged
func PurgeQueue(session Session, queueName string) (int, error) {
	return session.QueuePurge(queueName, false /* wait*/)
}

// BindQueueToExchange binds the given queue to the given exchange.
// TODO(JD) support for header based routing
func BindQueueToExchange(session Session,
	queueName, key, exchangeName string) error {
	return session.QueueBind(queueName, key, exchangeName, false /* wait */, nil)
}

// UnbindQueueFromExchange unbinds a queue from an exchange
func UnbindQueueFromExchange(session Session,
	queueName, key, exchangeName string) error {
	return session.QueueUnbind(queueName, key, exchangeName, nil)
}
