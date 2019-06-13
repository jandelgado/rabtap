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

	// rabtap specific fields
	XRabtapReceivedTimestamp time.Time

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

// WriteMessageJSON writes the given message as JSON, optionally with the
// body included to a stream.
func WriteMessageJSON(out io.Writer, message rabtap.TapMessage) error {
	metadata, err := json.MarshalIndent(NewRabtapPersistentMessage(message), "", "  ")
	if err != nil {
		return err
	}
	_, err = out.Write(metadata)
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

func saveMessageAsJSONFile(filename string, message rabtap.TapMessage) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	err = WriteMessageJSON(writer, message)
	if err != nil {
		return err
	}
	return writer.Flush()
}

// SaveMessageToRawFile writes a message to 2 files, one with the metadata, and
// one with the payload
func SaveMessageToRawFile(basename string, message rabtap.TapMessage) error {
	filenameRaw := basename + ".dat"
	filenameJSON := basename + ".json"
	log.Debugf("saving message  %s (RAW) with meta data in %s", filenameRaw, filenameJSON)
	err := saveMessageBodyAsBlobFile(filenameRaw, message.AmqpMessage.Body)
	if err != nil {
		return err
	}
	//return saveMessageAsJSONFile(filenameJSON, false, message)
	// save metadata file without the body
	oldBody := message.AmqpMessage.Body
	message.AmqpMessage.Body = []byte{}
	err = saveMessageAsJSONFile(filenameJSON, message)
	message.AmqpMessage.Body = oldBody
	return err
}

// SaveMessageToJSONFile writes a message to a single JSON file, where
// the body will be BASE64 encoded
func SaveMessageToJSONFile(filename string, message rabtap.TapMessage) error {
	log.Debugf("saving message to %s (JSON)", filename)
	return saveMessageAsJSONFile(filename, message)
}
