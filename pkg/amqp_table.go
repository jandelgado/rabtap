// Copyright (C) 2017-2022 Jan Delgado

package rabtap

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// EnsureAMQPTable returns an object where all map[string]interface{}
// are replaced by amqp.Table{} so it is compatible with the amqp
// libs type system when it comes to passing headers, which expects (nested)
// amqp.Table structures.
//
// See https://github.com/streadway/amqp/blob/e6b33f460591b0acb2f13b04ef9cf493720ffe17/types.go#L227
func EnsureAMQPTable(m interface{}) interface{} {
	switch x := m.(type) {

	case []interface{}:
		a := make([]interface{}, len(x))
		for i := range x {
			a[i] = EnsureAMQPTable(x[i])
		}
		return a

	case amqp.Table:
		m := amqp.Table{}
		for k, v := range x {
			m[k] = EnsureAMQPTable(v)
		}
		return m

	case map[string]interface{}:
		m := amqp.Table{}
		for k, v := range x {
			m[k] = EnsureAMQPTable(v)
		}
		return m

	default:
		return x
	}
}
