// Copyright (C) 2017 Jan Delgado

package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

// RabtapPersistentMessage is a messages as persistet from/to a JSON file
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

	Body *[]byte `json:",omitempty"`
}

// CreateTimestampFilename returns a filename based on a RFC3339Nano
// timstamp where all ":" are replaced with "_"
func CreateTimestampFilename(t time.Time) string {
	basename := t.Format(time.RFC3339Nano)
	return strings.Replace(basename, ":", "_", -1)
}

// NewRabtapPersistentMessage creates RabtapPersistentMessage a object
// from an amqp.Delivery
func NewRabtapPersistentMessage(m amqp.Delivery,
	includeBody bool) RabtapPersistentMessage {
	message := RabtapPersistentMessage{
		Headers:         m.Headers,
		ContentType:     m.ContentType,
		ContentEncoding: m.ContentEncoding,
		Priority:        m.Priority,
		CorrelationID:   m.CorrelationId,
		ReplyTo:         m.ReplyTo,
		Expiration:      m.Expiration,
		MessageID:       m.MessageId,
		Timestamp:       m.Timestamp,
		Type:            m.Type,
		UserID:          m.UserId,
		AppID:           m.AppId,
		DeliveryTag:     m.DeliveryTag,
		Exchange:        m.Exchange,
		RoutingKey:      m.RoutingKey,
		Body:            nil,
	}
	if includeBody {
		message.Body = &m.Body
	}
	return message
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
		Body:            *s.Body}
}

// WriteMessageBodyBlob writes the given message the provided stream.
func WriteMessageBodyBlob(out io.Writer, message *amqp.Delivery) error {
	_, err := out.Write(message.Body)
	return err
}

// WriteMessageJSON writes the given message as JSON, optionally with the
// body included to a stream.
func WriteMessageJSON(out io.Writer, includeBody bool, message *amqp.Delivery) error {
	// serialize message without body
	metadata, err := json.MarshalIndent(NewRabtapPersistentMessage(*message, includeBody), "", "  ")
	if err != nil {
		return err
	}
	_, err = out.Write(metadata)
	return err
}

func saveMessageBodyAsBlobFile(filename string, message *amqp.Delivery) error {
	// save the message as binary file 1:1
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	err = WriteMessageBodyBlob(writer, message)
	if err != nil {
		return err
	}
	return writer.Flush()
}

func saveMessageAsJSONFile(filename string, includeBody bool, message *amqp.Delivery) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	err = WriteMessageJSON(writer, includeBody, message)
	if err != nil {
		return err
	}
	return writer.Flush()
}

// SaveMessageToRawFile writes a message to 2 files, one with the metadata, and
// one with the payload
func SaveMessageToRawFile(basename string, message *amqp.Delivery) error {
	filenameRaw := basename + ".dat"
	filenameJSON := basename + ".json"
	log.Debugf("saving message  %s (RAW) with meta data in %s", filenameRaw, filenameJSON)
	err := saveMessageBodyAsBlobFile(filenameRaw, message)
	if err != nil {
		return err
	}
	return saveMessageAsJSONFile(filenameJSON, false, message)
}

// SaveMessageToJSONFile writes a message to a single JSON file, where
// the body will be BASE64 encoded
func SaveMessageToJSONFile(filename string, message *amqp.Delivery) error {
	log.Debugf("saving message to %s (JSON)", filename)
	return saveMessageAsJSONFile(filename, true, message)
}
