package rabtap

import (
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestNewRoutingFromStringsConstructsRoutingWithAmqpTable(t *testing.T) {

	r := NewRoutingFromStrings("exchange", "key", map[string]string{"A": "B"})

	assert.Equal(t, "exchange", r.Exchange())
	assert.Equal(t, "key", r.Key())
	assert.Equal(t, amqp.Table{"A": "B"}, r.Headers())
}

func TestMergeTableMergesTwoTablesSecondOneOverridingFirstOne(t *testing.T) {
	first := amqp.Table{"A": "B", "X": "Y"}
	second := amqp.Table{"C": "D", "X": "W"}

	merged := mergeTables(first, second)

	assert.Equal(t, amqp.Table{"A": "B", "C": "D", "X": "W"}, merged)
}
