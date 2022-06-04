package rabtap

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Routing describes where a message should be published
type Routing struct {
	key      string
	headers  amqp.Table
	exchange string
}

func (s Routing) String() string {
	r := fmt.Sprintf("exchange: '%s'", s.exchange)
	if len(s.headers) > 0 {
		r += fmt.Sprintf(", headers: %v", s.headers)
	}
	if s.key != "" {
		r += fmt.Sprintf(", routingkey: '%s'", s.key)
	}
	return r
}

func NewRouting(exchange, key string, headers amqp.Table) Routing {
	amqpHeaders := amqp.Table{}
	for k, v := range headers {
		amqpHeaders[k] = v
	}
	return Routing{exchange: exchange, key: key, headers: amqpHeaders}
}

func (s Routing) Exchange() string {
	return s.exchange
}

func (s Routing) Key() string {
	return s.key
}

func (s Routing) Headers() amqp.Table {
	return s.headers
}

// MergeTable merges the given amqp.Table's, the second one overrideing the
// values of the first one
func MergeTables(first, second amqp.Table) amqp.Table {
	res := make(amqp.Table, len(first)+len(second))
	for k, v := range first {
		res[k] = v
	}
	for k, v := range second {
		res[k] = v
	}
	return res
}
