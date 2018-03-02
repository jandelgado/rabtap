// Copyright (C) 2017 Jan Delgado

package main

import "github.com/streadway/amqp"

// DefaultMessageFormatter is the standard message.
type DefaultMessageFormatter struct{}

// Format just returns the message body as string, no formatting applied.
func (s DefaultMessageFormatter) Format(message *amqp.Delivery) string {
	return string(message.Body)
}
