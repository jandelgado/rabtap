// read persisted messages from files
// Copyright (C) 2019-2022 Jan Delgado
package main

import (
	"encoding/json"
	"fmt"
	"io"
)

func readMessageFromJSON(reader io.Reader) (RabtapPersistentMessage, error) {
	var message RabtapPersistentMessage
	decoder := json.NewDecoder(reader)
	decoder.UseNumber() // decode numbers as json.Number, not float64
	err := decoder.Decode(&message)
	return message, err
}

// readMessageFromJSONStream reads JSON messages from the given decoder as long
// as there are messages available.
func readMessageFromJSONStream(decoder *json.Decoder) (RabtapPersistentMessage, error) {
	var message RabtapPersistentMessage
	err := decoder.Decode(&message)
	return message, err
}

// NewReaderMessageSource returns a MessageSource that reads messages from
// the the given reader in the provided format
func NewReaderMessageSource(format string, reader io.ReadCloser) (MessageSource, error) {
	switch format {
	case "json-nopp":
		fallthrough
	case "json":
		decoder := json.NewDecoder(reader)
		return func() (RabtapPersistentMessage, error) {
			msg, err := readMessageFromJSONStream(decoder)
			return msg, err
		}, nil
	case "raw":
		read := false // only read one file, then return EOF
		return func() (RabtapPersistentMessage, error) {
			if read {
				return RabtapPersistentMessage{}, io.EOF
			}
			buf, err := io.ReadAll(reader) // note: does not return EOF
			read = true
			return RabtapPersistentMessage{Body: buf}, err
		}, nil
	}
	return nil, fmt.Errorf("invaild format %s", format)
}
