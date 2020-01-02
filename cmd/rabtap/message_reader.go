// read persisted messages from files
// Copyright (C) 2019 Jan Delgado

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/streadway/amqp"
)

// MessageReaderFunc provides messages that can be sent to an exchange.
// returns the message to be published, a flag if more messages are to be read,
// and an error.
type MessageReaderFunc func() (amqp.Publishing, bool, error)

// readMessageFromRawFile reads a single messages from the given io.Reader
// which is typically stdin or a file. If reading from stdin, CTRL+D (linux)
// or CTRL+Z (Win) on an empty line terminates the reader.
// -> readRawMessage
func readMessageFromRawFile(reader io.Reader) ([]byte, error) {
	return ioutil.ReadAll(reader)
	//return amqp.Publishing{Body: buf}, false, err
}

// -> readJSONMessage
func readMessageFromJSON(reader io.Reader) (RabtapPersistentMessage, error) {
	var message RabtapPersistentMessage

	contents, err := ioutil.ReadAll(reader)
	if err != nil {
		return message, err
	}
	err = json.Unmarshal(contents, &message)

	return message, err
}

// readMessageFromJSONStream reads JSON messages from the given decoder as long
// as there are messages available.
// -> readStreamedJSONMessage
func readMessageFromJSONStream(decoder *json.Decoder) (RabtapPersistentMessage, bool, error) {
	var message RabtapPersistentMessage
	err := decoder.Decode(&message)
	if err != nil {
		return message, false, err
	}
	//	return message.ToAmqpPublishing(), true, nil
	return message, true, nil
}

// createMessageFromDirReaderFunc returns a MessageReaderFunc that reads
// messages from the given list of filenames.
func createMessageFromDirReaderFunc(format string, files []filenameWithMetadata) (MessageReaderFunc, error) {

	i := 0

	switch format {
	case "json-nopp":
		fallthrough
	case "json":
		return func() (amqp.Publishing, bool, error) {
			fullMessage, err := readRabtapPersistentMessage(files[i].filename)
			i++
			return fullMessage.ToAmqpPublishing(), i < len(files), err
		}, nil
	case "raw":
		return func() (amqp.Publishing, bool, error) {
			body, err := ioutil.ReadFile(files[i].filename)
			message := files[i].metadata
			message.Body = body
			i++
			return message.ToAmqpPublishing(), i < len(files), err
		}, nil
	}
	return nil, fmt.Errorf("invaild format %s", format)
}

// createMessageReaderFunc returns a MessageReaderFunc that reads messages from
// the the given reader in the provided format
func createMessageReaderFunc(format string, reader io.ReadCloser) (MessageReaderFunc, error) {
	switch format {
	case "json-nopp":
		fallthrough
	case "json":
		decoder := json.NewDecoder(reader)
		return func() (amqp.Publishing, bool, error) {
			msg, more, err := readMessageFromJSONStream(decoder)
			return msg.ToAmqpPublishing(), more, err
		}, nil
	case "raw":
		return func() (amqp.Publishing, bool, error) {
			buf, err := readMessageFromRawFile(reader)
			return amqp.Publishing{Body: buf}, false, err
		}, nil
	}
	return nil, fmt.Errorf("invaild format %s", format)
}
