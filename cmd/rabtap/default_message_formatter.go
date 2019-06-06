// Copyright (C) 2017 Jan Delgado

package main

import rabtap "github.com/jandelgado/rabtap/pkg"

// DefaultMessageFormatter is the standard message.
type DefaultMessageFormatter struct{}

// Format just returns the message body as string, no formatting applied.
func (s DefaultMessageFormatter) Format(message rabtap.TapMessage) string {
	return string(message.AmqpMessage.Body)
}
