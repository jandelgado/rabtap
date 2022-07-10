// read persisted messages from files
// Copyright (C) 2019 Jan Delgado

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

// MessageReaderFunc provides messages that can be sent to an exchange.
// returns the message to be published, a flag if more messages are to be read,
// and an error.
type MessageReaderFunc func() (RabtapPersistentMessage, bool, error)

// readMessageFromRawFile reads a single messages from the given io.Reader
// which is typically stdin or a file. If reading from stdin, CTRL+D (linux)
// or CTRL+Z (Win) on an empty line terminates the reader.
func readMessageFromRawFile(reader io.Reader) ([]byte, error) {
	return ioutil.ReadAll(reader)
}

func readMessageFromJSON(reader io.Reader) (RabtapPersistentMessage, error) {
	var message RabtapPersistentMessage

	decoder := json.NewDecoder(reader)
	decoder.UseNumber() // decode numbers as json.Number, not float64
	err := decoder.Decode(&message)
	return message, err
}

// readMessageFromJSONStream reads JSON messages from the given decoder as long
// as there are messages available.
func readMessageFromJSONStream(decoder *json.Decoder) (RabtapPersistentMessage, bool, error) {
	var message RabtapPersistentMessage
	err := decoder.Decode(&message)
	if err != nil {
		return message, false, err
	}
	return message, true, nil
}

// CreateMessageReaderFunc returns a MessageReaderFunc that reads messages from
// the the given reader in the provided format
func CreateMessageReaderFunc(format string, reader io.ReadCloser) (MessageReaderFunc, error) {
	switch format {
	case "json-nopp":
		fallthrough
	case "json":
		decoder := json.NewDecoder(reader)
		return func() (RabtapPersistentMessage, bool, error) {
			msg, more, err := readMessageFromJSONStream(decoder)
			return msg, more, err
		}, nil
	case "raw":
		return func() (RabtapPersistentMessage, bool, error) {
			buf, err := readMessageFromRawFile(reader)
			return RabtapPersistentMessage{Body: buf}, false, err
		}, nil
	}
	return nil, fmt.Errorf("invaild format %s", format)
}
