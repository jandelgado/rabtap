package rabtap

import (
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestNewRoutingConstructsRoutingWithAmqpTable(t *testing.T) {

	r := NewRouting("exchange", "key", amqp.Table{"A": "B"})

	assert.Equal(t, "exchange", r.Exchange())
	assert.Equal(t, "key", r.Key())
	assert.Equal(t, amqp.Table{"A": "B"}, r.Headers())
}

func TestMergeTableMergesNilTablesIntoAnEmptyTable(t *testing.T) {
	assert.Equal(t, amqp.Table{}, MergeTables(nil, nil))
}

func TestMergeTableMergesTwoTablesSecondOneOverridingFirstOne(t *testing.T) {
	first := amqp.Table{"A": "B", "X": "Y"}
	second := amqp.Table{"C": "D", "X": "W"}

	merged := MergeTables(first, second)

	assert.Equal(t, amqp.Table{"A": "B", "C": "D", "X": "W"}, merged)
}
