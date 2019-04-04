// Copyright (C) 2017 Jan Delgado

package main

import (
	"encoding/json"
	"io"
	"text/template"

	"github.com/streadway/amqp"
)

// messageTemplate is the default template to print a message
const messageTemplate = `------ {{ .Title }} ------
exchange.......: {{ ExchangeColor .Message.Exchange }}
{{with .Message.RoutingKey}}routingkey.....: {{ KeyColor .}}
{{end}}{{with .Message.Priority}}priority.......: {{.}}
{{end}}{{with .Message.Expiration}}expiration.....: {{.}}
{{end}}{{with .Message.ContentType}}content-type...: {{.}}
{{end}}{{with .Message.ContentEncoding}}content-enc....: {{.}}
{{end}}{{with .Message.MessageId}}app-message-id.: {{.}}
{{end}}{{if not .Message.Timestamp.IsZero}}app-timestamp..: {{ .Message.Timestamp }}
{{end}}{{with .Message.Type}}app-type.......: {{.}}
{{end}}{{with .Message.CorrelationId}}app-corr-id....: {{.}}
{{end}}{{with .Message.Headers}}app-headers....: {{.}}
{{end -}}
{{ MessageColor .Body }}

`

// PrintMessageInfo holds info for template
type PrintMessageInfo struct {
	// Title to print
	Title string
	// Message receveived
	Message amqp.Delivery
	// formatted body
	Body string
	// formatted headers
	Headers string
}

// MessageFormatter formats the body of ampq.Delivery objects according to its
// type
type MessageFormatter interface {
	Format(message *amqp.Delivery) string
}

// Registry of available message formatters. Key is contentType
var messageFormatters = map[string]MessageFormatter{}

// RegisterMessageFormatter registers a new message formatter by its
// content type.
func RegisterMessageFormatter(contentType string, formatter MessageFormatter) {
	messageFormatters[contentType] = formatter
}

// NewMessageFormatter return a message formatter suitable for the given
// message type, determined by the message headers content type.
func NewMessageFormatter(message *amqp.Delivery) MessageFormatter {
	if formatter, ok := messageFormatters[message.ContentType]; ok {
		return formatter
	}
	return DefaultMessageFormatter{}
}

// PrettyPrintMessage formats and prints a amqp.Delivery message
func PrettyPrintMessage(out io.Writer, message *amqp.Delivery,
	title string, noColor bool) error {

	colorizer := NewColorPrinter(noColor)

	// get mesagge formatter according to message type
	formatter := NewMessageFormatter(message)

	// nicely print headers as JSON for better readability
	headers, err := json.Marshal(message.Headers)
	if err != nil {
		return err
	}

	body := formatter.Format(message)
	printStruct := PrintMessageInfo{
		Title:   title,
		Message: *message,
		Body:    body,
		Headers: string(headers),
	}
	t := template.Must(template.New("message").
		Funcs(colorizer.GetFuncMap()).Parse(messageTemplate))
	return t.Execute(out, printStruct)
}
