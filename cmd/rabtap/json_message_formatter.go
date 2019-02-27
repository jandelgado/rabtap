// Copyright (C) 2017 Jan Delgado

package main

import (
	"encoding/json"
	"strings"

	"github.com/streadway/amqp"
)

// JSONMessageFormatter pretty prints JSON formatted messages.
type JSONMessageFormatter struct{}

func init() {
	RegisterMessageFormatter("application/json", JSONMessageFormatter{})
}

// Format validates and formats a message in JSON format. The body can be a
// simple JSON object or an array of JSON objecs. If the message is not valid
// JSON, it will returned unformatted as-is.
func (s JSONMessageFormatter) Format(d *amqp.Delivery) string {

	var message []byte
	originalMessage := strings.TrimSpace(string(d.Body))
	if originalMessage[0] == '[' {
		// try to unmarshal array to JSON objects
		var arrayJSONObj []map[string]interface{}
		err := json.Unmarshal([]byte(originalMessage), &arrayJSONObj)
		if err != nil {
			return string(d.Body)
		}
		// pretty print JSON
		message, err = json.MarshalIndent(arrayJSONObj, "", "  ")
		if err != nil {
			return string(d.Body)
		}
	} else {

		// try to unmarshal simple json object
		var simpleJSONObj map[string]interface{}
		err := json.Unmarshal([]byte(originalMessage), &simpleJSONObj)

		if err != nil {
			return string(d.Body)
		}

		message, err = json.MarshalIndent(simpleJSONObj, "", "  ")
		if err != nil {
			return string(d.Body)
		}
	}
	return string(message)
}
