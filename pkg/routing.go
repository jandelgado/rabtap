package rabtap

import (
	"fmt"

	"github.com/streadway/amqp"
)

type Routing struct {
	key      string
	headers  amqp.Table
	exchange string
}

func (s Routing) String() string {
	if len(s.headers) > 0 {
		return fmt.Sprintf("exchange: '%s', headers: %v", s.exchange, s.headers)
	}
	if s.key != "" {
		return fmt.Sprintf("exchange: '%s', routingkey: '%s'", s.exchange, s.key)
	}
	return ""
}

func NewRouting(exchange, key string, headers amqp.Table) Routing {
	amqpHeaders := amqp.Table{}
	for k, v := range headers {
		amqpHeaders[k] = v
	}
	return Routing{exchange: exchange, key: key, headers: amqpHeaders}
}

func NewRoutingFromStrings(exchange, key string, headers map[string]string) Routing {
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

func mergeTables(first, second amqp.Table) amqp.Table {
	res := make(amqp.Table, len(first)+len(second))
	for k, v := range first {
		res[k] = v
	}
	for k, v := range second {
		res[k] = v
	}
	return res
}
