// write messages to writers and files
// Copyright (C) 2017-2019 Jan Delgado

package main

import (
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	amqp "github.com/rabbitmq/amqp091-go"
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

type PropertiesOverride struct {
	ContentType     *string
	ContentEncoding *string
	DeliveryMode    *uint8
	Priority        *uint8
	CorrelationID   *string
	ReplyTo         *string
	Expiration      *string
	MessageID       *string
	Timestamp       *time.Time
	Type            *string
	UserID          *string
	AppID           *string
}

// NewRabtapPersistentMessage creates RabtapPersistentMessage object
// from a rabtap.TapMessage
func NewRabtapPersistentMessage(message rabtap.TapMessage) RabtapPersistentMessage {
	m := message.AmqpMessage
	return RabtapPersistentMessage{
		Headers:                  m.Headers,
		ContentType:              m.ContentType,
		ContentEncoding:          m.ContentEncoding,
		DeliveryMode:             m.DeliveryMode,
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
func (s *RabtapPersistentMessage) ToAmqpPublishing() amqp.Publishing {
	return amqp.Publishing{
		Headers:         s.Headers,
		ContentType:     s.ContentType,
		ContentEncoding: s.ContentEncoding,
		DeliveryMode:    s.DeliveryMode,
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

func (s *RabtapPersistentMessage) WithProperties(props PropertiesOverride) *RabtapPersistentMessage {
	if props.ContentType != nil {
		s.ContentType = *props.ContentType
	}
	if props.ContentEncoding != nil {
		s.ContentEncoding = *props.ContentEncoding
	}
	if props.DeliveryMode != nil {
		s.DeliveryMode = *props.DeliveryMode
	}
	if props.Priority != nil {
		s.Priority = *props.Priority
	}
	if props.CorrelationID != nil {
		s.CorrelationID = *props.CorrelationID
	}
	if props.ReplyTo != nil {
		s.ReplyTo = *props.ReplyTo
	}
	if props.Expiration != nil {
		s.Expiration = *props.Expiration
	}
	if props.MessageID != nil {
		s.MessageID = *props.MessageID
	}
	if props.Timestamp != nil {
		s.Timestamp = *props.Timestamp
	}
	if props.Type != nil {
		s.Type = *props.Type
	}
	if props.UserID != nil {
		s.UserID = *props.UserID
	}
	if props.AppID != nil {
		s.AppID = *props.AppID
	}
	return s
}
