package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadMessageFromJSON(t *testing.T) {
	// note: base64dec("aGVsbG8=") == "hello"
	data := `
	{
	  "Headers": null,
	  "ContentType": "text/plain",
	  "ContentEncoding": "",
	  "DeliveryMode": 0,
	  "Priority": 0,
	  "CorrelationID": "",
	  "ReplyTo": "",
	  "Expiration": "",
	  "MessageID": "",
	  "Timestamp": "2017-10-28T23:45:33+02:00",
	  "Type": "",
	  "UserID": "",
	  "AppID": "rabtap.testgen",
	  "DeliveryTag": 63,
	  "Redelivered": false,
	  "Exchange": "amq.topic",
	  "RoutingKey": "test-q-amq.topic-0",
	  "Body": "aGVsbG8="
    }`
	reader := bytes.NewReader([]byte(data))

	msg, err := readMessageFromJSON(reader)
	assert.Nil(t, err)
	assert.Equal(t, []byte("hello"), msg.Body)
	assert.Equal(t, "amq.topic", msg.Exchange)
}

func TestReadMessageFromJSONStreamReturnsOneMessagePerCall(t *testing.T) {
	// note: base64dec("aGVsbG8=") == "hello"
	//        base64dec("c2Vjb25kCg==") == "second\n"
	data := `
	{
	  "Headers": null,
	  "ContentType": "text/plain",
	  "ContentEncoding": "",
	  "DeliveryMode": 0,
	  "Priority": 0,
	  "CorrelationID": "",
	  "ReplyTo": "",
	  "Expiration": "",
	  "MessageID": "",
	  "Timestamp": "2017-10-28T23:45:33+02:00",
	  "Type": "",
	  "UserID": "",
	  "AppID": "rabtap.testgen",
	  "DeliveryTag": 63,
	  "Redelivered": false,
	  "Exchange": "amq.topic",
	  "RoutingKey": "test-q-amq.topic-0",
	  "Body": "aGVsbG8="
    }
	{
		"Body": "c2Vjb25kCg=="
	}`
	reader := bytes.NewReader([]byte(data))
	decoder := json.NewDecoder(reader)

	msg, err := readMessageFromJSONStream(decoder)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello"), msg.Body)
	assert.Equal(t, "amq.topic", msg.Exchange)

	msg, err = readMessageFromJSONStream(decoder)
	assert.NoError(t, err)
	assert.Equal(t, []byte("second\n"), msg.Body)

	msg, err = readMessageFromJSONStream(decoder)
	assert.Equal(t, io.EOF, err)
}

func TestCreateMessageReaderFuncReturnsErrorForUnknownFormat(t *testing.T) {
	reader := ioutil.NopCloser(bytes.NewReader([]byte("")))
	_, err := CreateMessageReaderFunc("invalid", reader)
	assert.NotNil(t, err)
}

func TestCreateMessageReaderFuncReturnsJSONReaderForJSONFormats(t *testing.T) {

	for _, format := range []string{"json", "json-nopp"} {
		reader := ioutil.NopCloser(bytes.NewReader([]byte(`{"Body": "aGVsbG8="}`)))

		readFunc, err := CreateMessageReaderFunc(format, reader)
		assert.Nil(t, err)

		msg, err := readFunc()
		assert.NoError(t, err)
		assert.Equal(t, []byte("hello"), msg.Body)

		msg, err = readFunc()
		assert.Equal(t, io.EOF, err)
	}
}

func TestCreateMessageReaderFuncReturnsRawFileReaderForRawFormats(t *testing.T) {

	reader := ioutil.NopCloser(bytes.NewReader([]byte("hello")))

	readFunc, err := CreateMessageReaderFunc("raw", reader)
	assert.Nil(t, err)

	msg, err := readFunc()
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello"), msg.Body)

	msg, err = readFunc()
	assert.Equal(t, io.EOF, err)
}
