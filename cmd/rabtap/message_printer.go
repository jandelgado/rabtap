// Copyright (C) 2017-2019 Jan Delgado

package main

import (
	"io"
	"text/template"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// messageTemplate is the default template to print a message
// TODO allow externalization of template
const messageTemplate = `------ message received on {{ .Message.ReceivedTimestamp.Format "2006-01-02T15:04:05Z07:00" }} ------
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
{{end}}{{with .Message.AmqpMessage.ReplyTo}}reply-to.......: {{.}}
{{end}}{{with .Message.AmqpMessage.AppId}}app-id.........: {{.}}
{{end}}{{with .Message.AmqpMessage.UserId}}user-id........: {{.}}
{{end}}{{with .Message.AmqpMessage.Headers}}app-headers....: {{.}}
{{end -}}
{{ MessageColor (call .Body) }}

`

// PrintMessageEnv holds info for template
type PrintMessageEnv struct {
	// Message receveived
	Message rabtap.TapMessage
	// formatted body
	Body func() string
}

// MessageBodyFormatter formats the body of a message
type MessageBodyFormatter interface {
	Format(body []byte) string
}

// Registry of available message formatters. Key is contentType
var messageFormatters = map[string]MessageBodyFormatter{}

// RegisterMessageFormatter registers a new message formatter by its
// content type.
func RegisterMessageFormatter(contentType string, formatter MessageBodyFormatter) {
	messageFormatters[contentType] = formatter
}

// NewMessageFormatter return a message formatter suitable the given
// contentType.
func NewMessageFormatter(contentType string) MessageBodyFormatter {
	if formatter, ok := messageFormatters[contentType]; ok {
		return formatter
	}
	return DefaultMessageFormatter{}
}

// PrettyPrintMessage formats and prints a tapped message
func PrettyPrintMessage(out io.Writer, message rabtap.TapMessage) error {

	formatter := NewMessageFormatter(message.AmqpMessage.ContentType)

	printEnv := PrintMessageEnv{
		Message: message,
		Body: func() string {
			if b, err := Body(message.AmqpMessage); err != nil {
				// decoding failed, printing body as-is
				return formatter.Format(message.AmqpMessage.Body)
			} else {
				return formatter.Format(b)
			}
		},
	}

	colorizer := NewColorPrinter()
	t := template.Must(template.New("message").
		Funcs(colorizer.GetFuncMap()).Parse(messageTemplate))
	return t.Execute(out, printEnv)
}
