// Copyright (C) 2017 Jan Delgado

package main

import (
	"io"
	"text/template"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// messageTemplate is the default template to print a message
const messageTemplate = `------ {{ .Title }} ------
exchange.......: {{ ExchangeColor .Message.AmqpMessage.Exchange }}
{{with .Message.AmqpMessage.RoutingKey}}routingkey.....: {{ KeyColor .}}
{{end}}{{with .Message.AmqpMessage.Priority}}priority.......: {{.}}
{{end}}{{with .Message.AmqpMessage.Expiration}}expiration.....: {{.}}
{{end}}{{with .Message.AmqpMessage.ContentType}}content-type...: {{.}}
{{end}}{{with .Message.AmqpMessage.ContentEncoding}}content-enc....: {{.}}
{{end}}{{with .Message.AmqpMessage.MessageId}}app-message-id.: {{.}}
{{end}}{{if not .Message.AmqpMessage.Timestamp.IsZero}}app-timestamp..: {{ .Message.AmqpMessage.Timestamp }}
{{end}}{{with .Message.AmqpMessage.Type}}app-type.......: {{.}}
{{end}}{{with .Message.AmqpMessage.CorrelationId}}app-corr-id....: {{.}}
{{end}}{{with .Message.AmqpMessage.Headers}}app-headers....: {{.}}
{{end -}}
{{ MessageColor .Body }}

`

// PrintMessageInfo holds info for template
type PrintMessageInfo struct {
	// Title to print
	Title string
	// Message receveived
	Message rabtap.TapMessage
	// formatted body
	Body string
}

// MessageFormatter formats the body of tapped message
type MessageFormatter interface {
	Format(message rabtap.TapMessage) string
}

// Registry of available message formatters. Key is contentType
var messageFormatters = map[string]MessageFormatter{}

// RegisterMessageFormatter registers a new message formatter by its
// content type.
func RegisterMessageFormatter(contentType string, formatter MessageFormatter) {
	messageFormatters[contentType] = formatter
}

// NewMessageFormatter return a message formatter suitable the given
// contentType.
func NewMessageFormatter(contentType string) MessageFormatter {
	if formatter, ok := messageFormatters[contentType]; ok {
		return formatter
	}
	return DefaultMessageFormatter{}
}

// PrettyPrintMessage formats and prints a tapped message
func PrettyPrintMessage(out io.Writer, message rabtap.TapMessage,
	title string, noColor bool) error {

	colorizer := NewColorPrinter(noColor)

	formatter := NewMessageFormatter(message.AmqpMessage.ContentType)

	printStruct := PrintMessageInfo{
		Title:   title,
		Message: message,
		Body:    formatter.Format(message),
	}
	t := template.Must(template.New("message").
		Funcs(colorizer.GetFuncMap()).Parse(messageTemplate))
	return t.Execute(out, printStruct)
}
