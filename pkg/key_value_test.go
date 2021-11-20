package rabtap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToAmqpTableInfersInteger(t *testing.T) {
	m := KeyValueMap{"x": "123"}

	assert.Equal(t, int(123), ToAMQPTable(m)["x"])
}

func TestToAmqpTableInfersRFC3339Timestamp(t *testing.T) {
	m := KeyValueMap{"x": "2021-11-18T23:05:02-02:00"}
	ts, _ := time.Parse(time.RFC3339, "2021-11-18T23:05:02-02:00")

	assert.Equal(t, ts, ToAMQPTable(m)["x"])
}

func TestToAmqpTableFallsBackToString(t *testing.T) {
	m := KeyValueMap{"x": "hello"}

	assert.Equal(t, "hello", ToAMQPTable(m)["x"])
}
