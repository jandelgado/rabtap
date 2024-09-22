// Copyright (C) 2017 Jan Delgado

package main

import (
	"encoding/json"
	"strings"
)

// JSONMessageFormatter pretty prints JSON formatted messages.
type JSONMessageFormatter struct{}

var (
	_ = func() struct{} {
		RegisterMessageFormatter("application/json", JSONMessageFormatter{})
		return struct{}{}
	}()
)

// Format tries to format a message in JSON format. The body can be a simple
// JSON object or an array of JSON objects. If the message is not valid JSON,
// it will be returned unformatted as-is.
func (s JSONMessageFormatter) Format(body []byte) string {

	var formatted []byte
	originalMessage := strings.TrimSpace(string(body))
	if len(originalMessage) == 0 {
		return string(body)
	}
	if originalMessage[0] == '[' {
		// try to unmarshal array to JSON objects
		var arrayJSONObj []map[string]interface{}
		err := json.Unmarshal([]byte(originalMessage), &arrayJSONObj)
		if err != nil {
			return string(body)
		}
		// pretty print JSON
		formatted, err = json.MarshalIndent(arrayJSONObj, "", "  ")
		if err != nil {
			return string(body)
		}
	} else {

		// try to unmarshal simple json object
		var simpleJSONObj map[string]interface{}
		err := json.Unmarshal([]byte(originalMessage), &simpleJSONObj)

		if err != nil {
			return string(body)
		}

		formatted, err = json.MarshalIndent(simpleJSONObj, "", "  ")
		if err != nil {
			return string(body)
		}
	}
	return string(formatted)
}
