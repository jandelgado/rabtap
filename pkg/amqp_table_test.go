// Copyright (C) 2017-2022 Jan Delgado

package rabtap

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestEnsureTableKeepsArray(t *testing.T) {
	array := []interface{}{"a"}
	assert.Equal(t, array, EnsureAMQPTable(array))
}

func TestEnsureTableKeepsBasicType(t *testing.T) {
	assert.Equal(t, "test", EnsureAMQPTable("test"))
}

func TestEnsureTableKeepsTable(t *testing.T) {
	table := amqp.Table{"test": "a"}
	assert.Equal(t, table, EnsureAMQPTable(table))
}

func TestEnsureTableTransformsMapToTable(t *testing.T) {
	m := map[string]interface{}{"k": "v", "x": map[string]interface{}{"a": "b"}}
	expected := amqp.Table{"k": "v", "x": amqp.Table{"a": "b"}}
	assert.Equal(t, expected, EnsureAMQPTable(m))
}
