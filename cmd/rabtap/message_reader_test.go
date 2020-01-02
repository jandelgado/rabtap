package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadMessageFromRawFile(t *testing.T) {
	reader := bytes.NewReader([]byte("hello"))

	buf, err := readMessageFromRawFile(reader)
	assert.Nil(t, err)
	assert.Equal(t, []byte("hello"), buf)
}

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
	// TODO test additional attributes
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
	//reader := ioutil.NopCloser(bytes.NewReader([]byte("hello world"))) // r type is io.ReadCloser
	reader := bytes.NewReader([]byte(data))
	decoder := json.NewDecoder(reader)

	msg, more, err := readMessageFromJSONStream(decoder)
	assert.Nil(t, err)
	assert.True(t, more)
	assert.Equal(t, []byte("hello"), msg.Body)
	assert.Equal(t, "amq.topic", msg.Exchange)
	// TODO test additional attributes

	msg, more, err = readMessageFromJSONStream(decoder)
	assert.Nil(t, err)
	assert.True(t, more)
	assert.Equal(t, []byte("second\n"), msg.Body)

	msg, more, err = readMessageFromJSONStream(decoder)
	assert.Equal(t, io.EOF, err)
	assert.False(t, more)
}

func TestCreateMessageReaderFuncReturnsErrorForUnknownFormat(t *testing.T) {
	reader := ioutil.NopCloser(bytes.NewReader([]byte("")))
	_, err := createMessageReaderFunc("invalid", reader)
	assert.NotNil(t, err)
}

func TestCreateMessageReaderFuncReturnsJSONReaderForJSONFormats(t *testing.T) {

	for _, format := range []string{"json", "json-nopp"} {
		reader := ioutil.NopCloser(bytes.NewReader([]byte(`{"Body": "aGVsbG8="}`)))

		readFunc, err := createMessageReaderFunc(format, reader)
		assert.Nil(t, err)

		msg, more, err := readFunc()
		assert.Nil(t, err)
		assert.True(t, more)
		assert.Equal(t, []byte("hello"), msg.Body)

		msg, more, err = readFunc()
		assert.Equal(t, err, io.EOF)
		assert.False(t, more)
	}
}

func TestCreateMessageReaderFuncReturnsRawFileReaderForRawFormats(t *testing.T) {

	reader := ioutil.NopCloser(bytes.NewReader([]byte("hello")))

	readFunc, err := createMessageReaderFunc("raw", reader)
	assert.Nil(t, err)

	msg, more, err := readFunc()
	assert.Nil(t, err)
	assert.False(t, more)
	assert.Equal(t, []byte("hello"), msg.Body)
}
