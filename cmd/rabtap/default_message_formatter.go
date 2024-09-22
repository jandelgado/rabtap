// Copyright (C) 2017 Jan Delgado

package main

// DefaultMessageFormatter is the standard message.
type DefaultMessageFormatter struct{}

// Format just returns the message body as string, no formatting applied.
func (s DefaultMessageFormatter) Format(body []byte) string {
	return string(body)
}
