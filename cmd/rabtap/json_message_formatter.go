// Copyright (C) 2017 Jan Delgado

package main

import (
	"encoding/json"
	"strings"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// JSONMessageFormatter pretty prints JSON formatted messages.
type JSONMessageFormatter struct{}

var (
	_ = func() struct{} {
		RegisterMessageFormatter("application/json", JSONMessageFormatter{})
		return struct{}{}
	}()
)

// Format validates and formats a message in JSON format. The body can be a
// simple JSON object or an array of JSON objecs. If the message is not valid
// JSON, it will returned unformatted as-is.
func (s JSONMessageFormatter) Format(message rabtap.TapMessage) string {

	var formatted []byte
	originalMessage := strings.TrimSpace(string(message.AmqpMessage.Body))
	if originalMessage[0] == '[' {
		// try to unmarshal array to JSON objects
		var arrayJSONObj []map[string]interface{}
		err := json.Unmarshal([]byte(originalMessage), &arrayJSONObj)
		if err != nil {
			return string(message.AmqpMessage.Body)
		}
		// pretty print JSON
		formatted, err = json.MarshalIndent(arrayJSONObj, "", "  ")
		if err != nil {
			return string(message.AmqpMessage.Body)
		}
	} else {

		// try to unmarshal simple json object
		var simpleJSONObj map[string]interface{}
		err := json.Unmarshal([]byte(originalMessage), &simpleJSONObj)

		if err != nil {
			return string(message.AmqpMessage.Body)
		}

		formatted, err = json.MarshalIndent(simpleJSONObj, "", "  ")
		if err != nil {
			return string(message.AmqpMessage.Body)
		}
	}
	return string(formatted)
}
