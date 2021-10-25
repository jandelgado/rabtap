package rabtap

import "github.com/streadway/amqp"

type KeyValueMap map[string]string

func ToAMQPTable(headers KeyValueMap) amqp.Table {
	table := amqp.Table{}
	for k, v := range headers {
		table[k] = v
	}
	return table
}
