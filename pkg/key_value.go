package rabtap

import (
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// KeyValueMap is a string -> string map used to store key value
// pairs defined on the command line.
type KeyValueMap map[string]string

// ToAMQPTable converts a KeyValueMap to an amqp.Table, trying to
// infer data types:
// - integers
// - timestamps in RFC3339 format
// - strings (default)
func ToAMQPTable(headers KeyValueMap) amqp.Table {
	table := amqp.Table{}

	for k, v := range headers {
		var val interface{}
		if i, err := strconv.Atoi(v); err == nil {
			val = i
		} else if d, err := time.Parse(time.RFC3339, v); err == nil {
			val = d
		} else {
			val = v
		}
		table[k] = val
	}
	return table
}
