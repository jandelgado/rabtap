// Copyright (C) 2017 Jan Delgado

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rabtap "github.com/jandelgado/rabtap/pkg"
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
	testdir, err := os.MkdirTemp("", "")
	require.Nil(t, err)
	t.Cleanup(func() {
        require.NoError(t ,os.RemoveAll(testdir))
    })

	// SaveMessagesToFiles() will create files "test.dat" and "test.json" in
	// testdir.
	basename := filepath.Join(testdir, "test")
	createdTs := time.Date(2019, time.June, 13, 17, 45, 1, 0, time.UTC)
	err = SaveMessageToRawFiles(basename, rabtap.NewTapMessage(testMessage, createdTs), JSONMarshalIndent)
	assert.Nil(t, err)

	// check contents of message body .dat file
	datFilename := basename + ".dat"
	contentsBody, err := os.ReadFile(datFilename)
	assert.Nil(t, err)
	assert.Equal(t, []byte("simple test message."), contentsBody)

	// check contents of metadata file
	metaFilename := basename + ".json"
	contentsMeta, err := os.ReadFile(metaFilename)
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
	testdir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	t.Cleanup(func() {
        require.NoError(t ,os.RemoveAll(testdir))
    })

	filename := filepath.Join(testdir, "test")
	createdTs := time.Date(2019, time.June, 13, 17, 45, 1, 0, time.UTC)
	err = SaveMessageToJSONFile(filename, rabtap.NewTapMessage(testMessage, createdTs), JSONMarshalIndent)
	assert.Nil(t, err)

	contents, err := os.ReadFile(filename)
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
