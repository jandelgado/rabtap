// Copyright (C) 2017 Jan Delgado

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testMessage used troughout tests
var testMessage = &amqp.Delivery{
	Exchange:        "exchange",
	RoutingKey:      "routingkey",
	Priority:        99,
	Expiration:      "2017-05-22 17:00:00",
	ContentType:     "plain/text",
	ContentEncoding: "utf-8",
	MessageId:       "4711",
	Timestamp:       time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	Type:            "some type",
	CorrelationId:   "4712",
	Headers:         amqp.Table{"header": "value"},
	AppId:           "123",
	UserId:          "456",
	Body:            []byte("simple test message."),
}

func TestJSONMarshalIndentMarshalsToIndentedJSON(t *testing.T) {
	data, err := JSONMarshalIndent(map[string]string{"Test": "ABC"})
	assert.Nil(t, err)
	assert.Equal(t, `{
  "Test": "ABC"
}`, string(data))

}

func TestJSONMarshalMarshalsToSingleLineJSON(t *testing.T) {
	data, err := JSONMarshal(map[string]string{"Test": "ABC"})
	assert.Nil(t, err)
	assert.Equal(t, `{"Test":"ABC"}`, string(data))
}

// TestSaveMessageToFiles tests the SaveMessagesToFiles() function by
// writing to and reading from temporary files.
func TestSaveMessageToRawFile(t *testing.T) {
	testdir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(testdir)

	// SaveMessagesToFiles() will create files "test.dat" and "test.json" in
	// testdir.
	basename := filepath.Join(testdir, "test")
	createdTs := time.Date(2019, time.June, 13, 17, 45, 1, 0, time.UTC)
	err = SaveMessageToRawFiles(basename, rabtap.NewTapMessage(testMessage, createdTs), JSONMarshalIndent)
	assert.Nil(t, err)

	// check contents of message body .dat file
	datFilename := basename + ".dat"
	contentsBody, err := ioutil.ReadFile(datFilename)
	assert.Nil(t, err)
	assert.Equal(t, []byte("simple test message."), contentsBody)

	// check contents of metadata file
	metaFilename := basename + ".json"
	contentsMeta, err := ioutil.ReadFile(metaFilename)
	assert.Nil(t, err)
	// deserialize from .json file
	var jsonMetaActual RabtapPersistentMessage
	err = json.Unmarshal(contentsMeta, &jsonMetaActual)
	assert.Nil(t, err)

	// test some of the attributes
	assert.Equal(t, testMessage.AppId, jsonMetaActual.AppID)
	assert.Equal(t, len(testMessage.Headers), len(jsonMetaActual.Headers))
	assert.Equal(t, testMessage.Headers["header"], jsonMetaActual.Headers["header"])
	assert.Equal(t, testMessage.Timestamp, jsonMetaActual.Timestamp)
	assert.Equal(t, createdTs, jsonMetaActual.XRabtapReceivedTimestamp)
}

func TestSaveMessageToFilesToInvalidDir(t *testing.T) {
	// use nonexisting path
	filename := filepath.Join("/thispathshouldnotexist", "test")
	err := SaveMessageToRawFiles(filename, rabtap.NewTapMessage(testMessage, time.Now()), JSONMarshalIndent)
	assert.NotNil(t, err)
}

// TestSaveMessageToFile tests the SaveMessagesToFile() function by
// writing to and reading a temporary files.
func TestSaveMessageToJSONFile(t *testing.T) {
	testdir, err := ioutil.TempDir("", "")
	assert.Nil(t, err)
	defer os.RemoveAll(testdir)

	filename := filepath.Join(testdir, "test")
	createdTs := time.Date(2019, time.June, 13, 17, 45, 1, 0, time.UTC)
	err = SaveMessageToJSONFile(filename, rabtap.NewTapMessage(testMessage, createdTs), JSONMarshalIndent)
	assert.Nil(t, err)

	contents, err := ioutil.ReadFile(filename)
	assert.Nil(t, err)

	// deserialize from .json file
	var jsonActual RabtapPersistentMessage
	err = json.Unmarshal(contents, &jsonActual)
	assert.Nil(t, err)

	assert.Equal(t, testMessage.AppId, jsonActual.AppID)
	assert.Equal(t, len(testMessage.Headers), len(jsonActual.Headers))
	assert.Equal(t, testMessage.Headers["header"], jsonActual.Headers["header"])
	assert.Equal(t, testMessage.Timestamp, jsonActual.Timestamp)
	assert.Equal(t, createdTs, jsonActual.XRabtapReceivedTimestamp)
	assert.Equal(t, []byte("simple test message."), jsonActual.Body)
}

func TestSaveMessageToFileToInvalidDir(t *testing.T) {
	// use nonexisting path
	filename := filepath.Join("/thispathshouldnotexist", "test")
	err := SaveMessageToJSONFile(filename, rabtap.NewTapMessage(testMessage, time.Now()), JSONMarshalIndent)
	assert.NotNil(t, err)
}

func TestCreateTimestampFilename(t *testing.T) {
	tm := time.Date(2009, time.November, 10, 23, 1, 2, 3, time.UTC)
	filename := CreateTimestampFilename(tm)
	assert.Equal(t, "2009-11-10T23_01_02.000000003Z", filename)
}

func ExampleWriteMessage() {

	// serialize with message body, Body will be base64 encoded.
	createdTs := time.Date(2019, time.June, 13, 17, 45, 1, 0, time.UTC)
	err := WriteMessage(os.Stdout,
		rabtap.NewTapMessage(testMessage, createdTs),
		JSONMarshalIndent)
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// {
	//   "Headers": {
	//     "header": "value"
	//   },
	//   "ContentType": "plain/text",
	//   "ContentEncoding": "utf-8",
	//   "DeliveryMode": 0,
	//   "Priority": 99,
	//   "CorrelationID": "4712",
	//   "ReplyTo": "",
	//   "Expiration": "2017-05-22 17:00:00",
	//   "MessageID": "4711",
	//   "Timestamp": "2009-11-10T23:00:00Z",
	//   "Type": "some type",
	//   "UserID": "456",
	//   "AppID": "123",
	//   "DeliveryTag": 0,
	//   "Redelivered": false,
	//   "Exchange": "exchange",
	//   "RoutingKey": "routingkey",
	//   "XRabtapReceivedTimestamp": "2019-06-13T17:45:01Z",
	//   "Body": "c2ltcGxlIHRlc3QgbWVzc2FnZS4="
	// }
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

func TestMessageFromFireHoseIsTransformed(t *testing.T) {
	msg := `
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

	model, err := readMessageFromJSON(strings.NewReader(msg))
	require.NoError(t, err)

	assert.True(t, IsFromFireHoseExchange(model))
	transformed, err := FromFireHoseMessage(model)

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
	assert.Equal(t, model.Body, transformed.Body)
}
