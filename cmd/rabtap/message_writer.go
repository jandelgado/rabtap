// write messages to writers and files
// Copyright (C) 2017-2019 Jan Delgado

package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/streadway/amqp"
)

// RabtapPersistentMessage is a messages as persisted from/to a JSON file
// object can be initialiazed from amqp.Delivery and to amqp.Publishing
type RabtapPersistentMessage struct {
	Headers map[string]interface{} `json:",omitempty"`

	// Properties
	ContentType     string
	ContentEncoding string
	DeliveryMode    uint8
	Priority        uint8
	CorrelationID   string
	ReplyTo         string
	Expiration      string
	MessageID       string
	Timestamp       time.Time
	Type            string
	UserID          string
	AppID           string

	DeliveryTag uint64
	Redelivered bool
	Exchange    string
	RoutingKey  string

	// rabtap specific fields
	XRabtapReceivedTimestamp time.Time

	// will be serialized as base64
	Body []byte
}

// CreateTimestampFilename returns a filename based on a RFC3339Nano
// timstamp where all ":" are replaced with "_"
func CreateTimestampFilename(t time.Time) string {
	basename := t.Format(time.RFC3339Nano)
	return strings.Replace(basename, ":", "_", -1)
}

// NewRabtapPersistentMessage creates RabtapPersistentMessage object
// from a rabtap.TapMessage
func NewRabtapPersistentMessage(message rabtap.TapMessage) RabtapPersistentMessage {
	m := message.AmqpMessage
	return RabtapPersistentMessage{
		Headers:                  m.Headers,
		ContentType:              m.ContentType,
		ContentEncoding:          m.ContentEncoding,
		Priority:                 m.Priority,
		CorrelationID:            m.CorrelationId,
		ReplyTo:                  m.ReplyTo,
		Expiration:               m.Expiration,
		MessageID:                m.MessageId,
		Timestamp:                m.Timestamp,
		Type:                     m.Type,
		UserID:                   m.UserId,
		AppID:                    m.AppId,
		DeliveryTag:              m.DeliveryTag,
		Exchange:                 m.Exchange,
		RoutingKey:               m.RoutingKey,
		XRabtapReceivedTimestamp: message.ReceivedTimestamp,
		Body:                     m.Body,
	}
}

// ToAmqpPublishing converts message to an amqp.Publishing object
func (s RabtapPersistentMessage) ToAmqpPublishing() amqp.Publishing {
	return amqp.Publishing{
		Headers:         s.Headers,
		ContentType:     s.ContentType,
		ContentEncoding: s.ContentEncoding,
		Priority:        s.Priority,
		CorrelationId:   s.CorrelationID,
		ReplyTo:         s.ReplyTo,
		Expiration:      s.Expiration,
		MessageId:       s.MessageID,
		Timestamp:       s.Timestamp,
		Type:            s.Type,
		UserId:          s.UserID,
		AppId:           s.AppID,
		Body:            s.Body}
}

// marshalFunc marshals messages prior to writing them e.g. in JSON format
type marshalFunc func(m interface{}) ([]byte, error)

// JSONMarshmarshallIndent the given message as a formatted JSON
func JSONMarshalIndent(m interface{}) ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}

// JSONMarshal marshalls the given message as a single line JSON
func JSONMarshal(m interface{}) ([]byte, error) {
	return json.Marshal(m)
}

// WriteMessage writes the given message using the proviced marshaller and writer
func WriteMessage(out io.Writer, message rabtap.TapMessage, marshaller marshalFunc) error {
	data, err := marshaller(NewRabtapPersistentMessage(message))
	if err != nil {
		return err
	}
	nl := []byte("\n")
	_, err = out.Write(append(data, nl...))
	return err
}

func saveMessageBodyAsBlobFile(filename string, body []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	_, err = writer.Write(body)
	if err != nil {
		return err
	}
	return writer.Flush()
}

func saveMessageAsJSONFile(filename string, message rabtap.TapMessage, marshaller marshalFunc) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	err = WriteMessage(writer, message, marshaller)
	if err != nil {
		return err
	}
	return writer.Flush()
}

// SaveMessageToRawFile writes a message to 2 files, one with the metadata, and
// one with the payload. The metadata will be serialized using the proviced marshaller
func SaveMessageToRawFiles(basename string, message rabtap.TapMessage, marshaller marshalFunc) error {
	filenameRaw := basename + ".dat"
	filenameMeta := basename + ".json"
	log.Debugf("saving message  %s (RAW) with meta data in %s", filenameRaw, filenameMeta)
	err := saveMessageBodyAsBlobFile(filenameRaw, message.AmqpMessage.Body)
	if err != nil {
		return err
	}
	// save metadata file without the body
	oldBody := message.AmqpMessage.Body
	message.AmqpMessage.Body = []byte{}
	err = saveMessageAsJSONFile(filenameMeta, message, marshaller)
	message.AmqpMessage.Body = oldBody
	return err
}

// SaveMessageToJSONFile writes a message to a single JSON file, where
// the body will be BASE64 encoded
func SaveMessageToJSONFile(filename string, message rabtap.TapMessage, marshaller marshalFunc) error {
	log.Debugf("saving message to %s (JSON)", filename)
	return saveMessageAsJSONFile(filename, message, marshaller)
}
