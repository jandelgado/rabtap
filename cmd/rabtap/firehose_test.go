// Copyright (C) 2017 Jan Delgado

package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoutingKeyFromHeaderReturnsRoutingKeyIfSet(t *testing.T) {
	headers := map[string]interface{}{
		"routing_keys": []interface{}{"key"}}

	assert.Equal(t, "key", routingKeyFromHeader(headers))
}

func TestRoutingKeyFromHeaderReturnsEmptyStringIfNotSet(t *testing.T) {
	headers := map[string]interface{}{}
	assert.Equal(t, "", routingKeyFromHeader(headers))
}

func TestIsFromFireHoseExchangeDetectsMessagesFromFireHose(t *testing.T) {
	assert.True(t, IsFromFireHoseExchange(RabtapPersistentMessage{Exchange: "amq.rabbitmq.trace"}))
	assert.False(t, IsFromFireHoseExchange(RabtapPersistentMessage{Exchange: "other"}))
}

func TestPropRetrievesElementByKey(t *testing.T) {
	m := map[string]interface{}{"key": int32(123)}

	assert.Equal(t, int32(123), prop(m, "key", int32(0)))
}

func TestPropReturnsDefaultIfKeyNotFound(t *testing.T) {
	m := map[string]interface{}{"key": int32(123)}

	assert.Equal(t, int32(99), prop(m, "other", int32(99)))
}

func TestFromFireHoseMessageTransformsMessage(t *testing.T) {
	// given
	headers := map[string]interface{}{
		"exchange_name": "newexchange",
		"routing_keys":  []interface{}{"newkey"},
		"properties": map[string]interface{}{
			"headers": map[string]interface{}{
				"a": 10,
				"b": "hello",
			},
			"content_type":     "newcontenttype",
			"content_encoding": "newcontentencoding",
			"delivery_mode":    json.Number("199"),
			"priority":         json.Number("198"),
			"correlation_id":   "newcorrelationid",
			"reply_to":         "newreplyto",
			"expiration":       "newexpiration",
			"message_id":       "newmessageid",
			"type":             "newtype",
			"user_id":          "newuserid",
			"app_id":           "newappid",
			"timestamp":        json.Number("123456"),
		}}

	m := RabtapPersistentMessage{
		Headers:                  headers,
		ContentType:              "contenttype",
		ContentEncoding:          "contentencoding",
		DeliveryMode:             99,
		Priority:                 98,
		CorrelationID:            "correlationid",
		ReplyTo:                  "replyto",
		Expiration:               "expiration",
		MessageID:                "12345",
		Timestamp:                time.Date(2020, time.June, 13, 17, 45, 1, 0, time.UTC),
		Type:                     "type",
		UserID:                   "userid",
		AppID:                    "appid",
		DeliveryTag:              97,
		Exchange:                 "exchange",
		RoutingKey:               "key",
		XRabtapReceivedTimestamp: time.Date(2021, time.June, 13, 17, 45, 1, 0, time.UTC),
		Body:                     []byte("body")}

	// when
	n, err := FromFireHoseMessage(m)

	// then
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"a": 10, "b": "hello"}, n.Headers)
	assert.Equal(t, "newcontenttype", n.ContentType)
	assert.Equal(t, "newcontentencoding", n.ContentEncoding)
	assert.Equal(t, uint8(199), n.DeliveryMode)
	assert.Equal(t, uint8(198), n.Priority)
	assert.Equal(t, "newcorrelationid", n.CorrelationID)
	assert.Equal(t, "newreplyto", n.ReplyTo)
	assert.Equal(t, "newexpiration", n.Expiration)
	assert.Equal(t, "newmessageid", n.MessageID)
	assert.Equal(t, "newtype", n.Type)
	assert.Equal(t, "newuserid", n.UserID)
	assert.Equal(t, "newappid", n.AppID)
	assert.Equal(t, "newexchange", n.Exchange)
	assert.Equal(t, "newkey", n.RoutingKey)
	assert.Equal(t, time.Unix(123456, 0), n.Timestamp)
}

func TestMessageNotFromFireHoseIsReturnedAsIs(t *testing.T) {
	// given
	msgJSON := `
{
  "Exchange": "some.exchange"
}`

	msg, err := readMessageFromJSON(strings.NewReader(msgJSON))
	require.NoError(t, err)

	// when
	transformed, err := FireHoseTransformer(msg)

	// then
	assert.NoError(t, err)
	assert.Equal(t, "some.exchange", transformed.Exchange)
}

func TestMessageFromFireHoseIsTransformed(t *testing.T) {
	// given
	msgJSON := `
{
  "Headers": {
    "channel": 1,
    "connection": "172.18.0.1:58332 -\u003e 172.18.0.2:5672",
    "exchange_name": "amq.topic",
    "node": "rabbit@1a92b8526e33",
    "properties": {
      "app_id": "rabtap.testgen",
      "content_type": "text/plain",
      "correlation_id": "correlationId",
      "delivery_mode": 1,
      "expiration": "1234",
      "message_id": "messageId",
      "priority": 99,
      "reply_to": "replyTo",
      "timestamp": 1657468145,
      "type": "type",
      "user_id": "guest",
  	  "headers": {
        "header1": "test0"
      }
    },
    "routed_queues": [
      "test-q-amq.topic-0"
    ],
    "routing_keys": [
      "test-q-amq.topic-0"
    ],
    "user": "guest",
    "vhost": "/"
  },
  "ContentType": "",
  "ContentEncoding": "",
  "DeliveryMode": 0,
  "Priority": 0,
  "CorrelationID": "",
  "ReplyTo": "",
  "Expiration": "",
  "MessageID": "",
  "Timestamp": "0001-01-01T00:00:00Z",
  "Type": "",
  "UserID": "",
  "AppID": "",
  "DeliveryTag": 2,
  "Redelivered": false,
  "Exchange": "amq.rabbitmq.trace",
  "RoutingKey": "publish.amq.topic",
  "XRabtapReceivedTimestamp": "2022-07-10T17:49:05.425800307+02:00",
  "Body": "dGVzdCBtZXNzYWdlICM0NzQgd2FzIHB1c2hlZCB0byBleGNoYW5nZSAnYW1xLnRvcGljJyB3aXRoIHJvdXRpbmcga2V5ICd0ZXN0LXEtYW1xLnRvcGljLTAnIGFuZCBoZWFkZXJzIGFtcXAwOTEuVGFibGV7fQ=="
}`

	msg, err := readMessageFromJSON(strings.NewReader(msgJSON))
	require.NoError(t, err)

	// when
	transformed, err := FireHoseTransformer(msg)

	// then
	assert.NoError(t, err)
	assert.Equal(t, "rabtap.testgen", transformed.AppID)
	assert.Equal(t, "text/plain", transformed.ContentType)
	assert.Equal(t, "correlationId", transformed.CorrelationID)
	assert.Equal(t, byte(1), transformed.DeliveryMode)
	assert.Equal(t, "1234", transformed.Expiration)
	assert.Equal(t, "messageId", transformed.MessageID)
	assert.Equal(t, byte(99), transformed.Priority)
	assert.Equal(t, "replyTo", transformed.ReplyTo)
	assert.Equal(t, time.Unix(1657468145, 0), transformed.Timestamp)
	assert.Equal(t, "type", transformed.Type)
	assert.Equal(t, "guest", transformed.UserID)
	assert.Equal(t, "test-q-amq.topic-0", transformed.RoutingKey)
	assert.Equal(t, "amq.topic", transformed.Exchange)
	assert.Equal(t, map[string]interface{}{"header1": "test0"}, transformed.Headers)
	assert.Equal(t, msg.Body, transformed.Body)
}
