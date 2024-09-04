package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMessagePropertiesReturnsEmptyPropertiesWhenNoOptionsAreSet(t *testing.T) {

	args := map[string]string{}
	props, err := parseMessageProperties(args)

	require.NoError(t, err)
	assert.Nil(t, props.ContentType)
	assert.Nil(t, props.ContentEncoding)
	assert.Nil(t, props.DeliveryMode)
	assert.Nil(t, props.Priority)
	assert.Nil(t, props.CorrelationID)
	assert.Nil(t, props.ReplyTo)
	assert.Nil(t, props.Expiration)
	assert.Nil(t, props.MessageID)
	assert.Nil(t, props.Timestamp)
	assert.Nil(t, props.Type)
	assert.Nil(t, props.UserID)
	assert.Nil(t, props.AppID)
}

func TestParseMessagePropertiesReturnsFullyPopulatedProperties(t *testing.T) {

	// given
	args := map[string]string{
		"ContentTYPE":     "content-type",
		"contentencoding": "content-encoding",
		"deliverymode":    "persistent",
		"priority":        "2",
		"correlationId":   "correlation-id",
		"replyto":         "reply-to",
		"expiration":      "expiration",
		"messageid":       "message-id",
		"timestamp":       "2024-12-05T17:18:23.000Z",
		"type":            "type",
		"userid":          "user-id",
		"appid":           "app-id",
	}
	// when
	props, err := parseMessageProperties(args)

	// then
	require.NoError(t, err)
	assert.Equal(t, "content-type", *props.ContentType)
	assert.Equal(t, "content-encoding", *props.ContentEncoding)
	assert.Equal(t, uint8(0), *props.DeliveryMode)
	assert.Equal(t, uint8(2), *props.Priority)
	assert.Equal(t, "correlation-id", *props.CorrelationID)
	assert.Equal(t, "reply-to", *props.ReplyTo)
	assert.Equal(t, "expiration", *props.Expiration)
	assert.Equal(t, "message-id", *props.MessageID)
	assert.Equal(t, time.Date(2024, 12, 5, 17, 18, 23, 0, time.UTC), *props.Timestamp)
	assert.Equal(t, "type", *props.Type)
	assert.Equal(t, "user-id", *props.UserID)
	assert.Equal(t, "app-id", *props.AppID)
}

func TestParseMessagePropertiesParsesDeliveryMode(t *testing.T) {

	args := map[string]string{"deliverymode": "persistent"}
	props, err := parseMessageProperties(args)
	require.NoError(t, err)
	assert.Equal(t, uint8(2), *props.DeliveryMode)

	args = map[string]string{"deliverymode": "transient"}
	props, err = parseMessageProperties(args)
	require.NoError(t, err)
	assert.Equal(t, uint8(1), *props.DeliveryMode)
}
