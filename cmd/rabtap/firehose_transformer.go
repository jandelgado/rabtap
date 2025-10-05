// detect and transform messages recorded from the RabbitMQ FireHose tracer.
// (see https://www.rabbitmq.com/firehose.html)
// Copyright (C) 2022 Jan Delgado

package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// FireHoseTransformer checks if a messages was recorded from the firehose
// exchange and then transforms the message so it will be replayed as the
// originally tapped message
func FireHoseTransformer(m RabtapPersistentMessage) (RabtapPersistentMessage, error) {
	if IsFromFireHoseExchange(m) {
		return FromFireHoseMessage(m)
	}
	return m, nil
}

// IsFromIsFromFireHoseExchange returns true if the given message was
// sent to the amq.rabbitmq.trace exchange, the FireHose exchange.
func IsFromFireHoseExchange(m RabtapPersistentMessage) bool {
	return m.Exchange == "amq.rabbitmq.trace"
}

// prop accesses a property in the given map m or return the provided default,
// if not found
func prop[T any](m map[string]interface{}, key string, def T) T {
	if val, found := m[key]; found {
		return val.(T)
	}
	return def
}

// propInt accesses a int64 property in the given map m or return the provided default,
// if not found
func propInt(m map[string]interface{}, key string, def int64) (int64, error) {
	num := prop(m, key, json.Number(fmt.Sprintf("%d", def)))
	return num.Int64()
}

func routingKeyFromHeader(header map[string]interface{}) string {
	routingkeys := prop(header, "routing_keys", []interface{}{})
	if len(routingkeys) > 0 {
		return routingkeys[0].(string) // TODO consider CC, BCC ?
	}
	return ""
}

// FromFireHoseMessage takes a message that was originally read from a FireHose
// exchange and transforms it into a message that can be replayed like the
// original message.  In messages tapped from the FireHose exchange, the actual
// message meta data is stored in the Header and Header.properties attributes
// (e.g.  exchange_name, content_type etc)
// See https://www.rabbitmq.com/firehose.html
func FromFireHoseMessage(m RabtapPersistentMessage) (RabtapPersistentMessage, error) {

	if m.Headers == nil {
		return RabtapPersistentMessage{}, fmt.Errorf("headers not set")
	}

	if props, found := m.Headers["properties"]; !found {
		return RabtapPersistentMessage{}, fmt.Errorf("headers.properties attribute missing")
	} else {
		props := props.(map[string]interface{})

		var err error
		var priority int64
		if priority, err = propInt(props, "priority", 0); err != nil {
			return RabtapPersistentMessage{}, fmt.Errorf("priority: %w", err)
		}
		var delivery_mode int64
		if delivery_mode, err = propInt(props, "delivery_mode", 0); err != nil {
			return RabtapPersistentMessage{}, fmt.Errorf("delivery_mode: %w", err)
		}
		var timestamp_s int64
		if timestamp_s, err = propInt(props, "timestamp", 0); err != nil {
			return RabtapPersistentMessage{}, fmt.Errorf("timestamp: %w", err)
		}
		return RabtapPersistentMessage{
			Headers:                  prop(props, "headers", map[string]interface{}{}),
			ContentType:              prop(props, "content_type", ""),
			ContentEncoding:          prop(props, "content_encoding", ""),
			DeliveryMode:             uint8(delivery_mode),
			Priority:                 uint8(priority),
			CorrelationID:            prop(props, "correlation_id", ""),
			ReplyTo:                  prop(props, "reply_to", ""),
			Expiration:               prop(props, "expiration", ""),
			MessageID:                prop(props, "message_id", ""),
			Timestamp:                time.Unix(timestamp_s, 0),
			Type:                     prop(props, "type", ""),
			UserID:                   prop(props, "user_id", ""),
			AppID:                    prop(props, "app_id", ""),
			Exchange:                 prop(m.Headers, "exchange_name", ""),
			RoutingKey:               routingKeyFromHeader(m.Headers),
			DeliveryTag:              m.DeliveryTag,
			XRabtapReceivedTimestamp: m.XRabtapReceivedTimestamp,
			Body:                     m.Body,
		}, nil
	}
}
