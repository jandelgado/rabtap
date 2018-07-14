// Copyright (C) 2017 Jan Delgado

package main

import (
	"crypto/tls"
	"os"
	"testing"

	"github.com/jandelgado/rabtap/pkg"
	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
)

func TestResolveTemplate(t *testing.T) {
	type Info struct {
		Name string
	}
	args := Info{"Jan"}

	const tpl = "hello {{ .Name }}"

	brokerInfoPrinter := NewBrokerInfoPrinter(
		BrokerInfoPrinterConfig{
			NoColor: true},
	)

	result := brokerInfoPrinter.resolveTemplate("test", tpl, args)
	assert.Equal(t, "hello Jan", result)
}

func TestFindExchangeByName(t *testing.T) {
	exchanges := []rabtap.RabbitExchange{
		{Name: "exchange1", Vhost: "vhost"},
		{Name: "exchange2", Vhost: "vhost"},
	}
	exchange := findExchangeByName(exchanges, "vhost", "exchange2")
	assert.NotNil(t, exchange)
	assert.Equal(t, "exchange2", exchange.Name)
}

func TestFindExchangeByNameNotFound(t *testing.T) {
	exchanges := []rabtap.RabbitExchange{
		{Name: "exchange1", Vhost: "vhost"},
	}
	exchange := findExchangeByName(exchanges, "/", "not-available")
	assert.Nil(t, exchange)
}

func TestFindQueueByName(t *testing.T) {
	queues := []rabtap.RabbitQueue{
		{Name: "q1", Vhost: "vhost"},
		{Name: "q2", Vhost: "vhost"},
	}
	queue := findQueueByName(queues, "vhost", "q2")
	assert.Equal(t, "q2", queue.Name)
	assert.Equal(t, "vhost", queue.Vhost)
}

func TestFindQueueByNameNotFound(t *testing.T) {
	queues := []rabtap.RabbitQueue{
		{Name: "q1", Vhost: "vhost"},
		{Name: "q2", Vhost: "vhost"},
	}
	queue := findQueueByName(queues, "/", "not-available")
	assert.Nil(t, queue)
}

func TestFilterStringListOfEmptyLists(t *testing.T) {
	flags := []bool{}
	strs := []string{}
	assert.Equal(t, []string{}, filterStringList(flags, strs))
}

func TestFilterStringListOneElementKeptInList(t *testing.T) {
	flags := []bool{false, true, false}
	strs := []string{"A", "B", "C"}
	assert.Equal(t, []string{"B"}, filterStringList(flags, strs))
}

func TestFindConnectionByName(t *testing.T) {
	conns := []rabtap.RabbitConnection{
		{Name: "c1", Vhost: "vhost"},
		{Name: "c2", Vhost: "vhost"},
	}
	conn := findConnectionByName(conns, "vhost", "c2")
	assert.Equal(t, "c2", conn.Name)
	assert.Equal(t, "vhost", conn.Vhost)
}

func TestFindConnectionByNameNotFoundReturnsNil(t *testing.T) {
	assert.Nil(t, findConnectionByName([]rabtap.RabbitConnection{}, "vhost", "c2"))
}

func TestFindConsumerByQueue(t *testing.T) {
	con := rabtap.RabbitConsumer{}
	con.Queue.Name = "q1"
	con.Queue.Vhost = "vhost"
	cons := []rabtap.RabbitConsumer{con}
	foundCon := findConsumerByQueue(cons, "vhost", "q1")
	assert.Equal(t, "q1", foundCon.Queue.Name)
	assert.Equal(t, "vhost", foundCon.Queue.Vhost)
}

func TestFindConsumerByQueueNotFoundReturnsNil(t *testing.T) {
	assert.Nil(t, findConsumerByQueue([]rabtap.RabbitConsumer{}, "vhost", "q1"))
}

func TestBrokerInfoPrintFailsOnInvalidUri(t *testing.T) {
	brokerInfoPrinter := NewBrokerInfoPrinter(BrokerInfoPrinterConfig{})
	err := brokerInfoPrinter.Print(rabtap.BrokerInfo{}, "//:xxx::invalid uri", os.Stdout)
	assert.NotNil(t, err)

}

func ExampleBrokerInfoPrinter_Print() {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := rabtap.NewRabbitHTTPClient(mock.URL, &tls.Config{})

	brokerInfoPrinter := NewBrokerInfoPrinter(
		BrokerInfoPrinterConfig{
			ShowStats:           false,
			ShowConsumers:       true,
			ShowDefaultExchange: false,
			QueueFilter:         TruePredicate,
			OmitEmptyExchanges:  false,
			NoColor:             true},
	)
	brokerInfo, err := client.BrokerInfo()
	if err != nil {
		log.Fatal(err)
	}

	if err := brokerInfoPrinter.Print(brokerInfo, "http://rabbitmq/api", os.Stdout); err != nil {
		log.Fatal(err)
	}

	// Output:
	// http://rabbitmq/api (broker ver='3.6.9', mgmt ver='3.6.9', cluster='rabbit@08f57d1fe8ab')
	// └── Vhost /
	//     ├── amq.direct (exchange, type 'direct', [D])
	//     ├── amq.fanout (exchange, type 'fanout', [D])
	//     ├── amq.headers (exchange, type 'headers', [D])
	//     ├── amq.match (exchange, type 'headers', [D])
	//     ├── amq.rabbitmq.log (exchange, type 'topic', [D|I])
	//     ├── amq.rabbitmq.trace (exchange, type 'topic', [D|I])
	//     ├── amq.topic (exchange, type 'topic', [D])
	//     ├── test-direct (exchange, type 'direct', [D|AD|I])
	//     │   ├── direct-q1 (queue, key='direct-q1', running, [D])
	//     │   │   ├── some_consumer (consumer user='guest', chan='172.17.0.1:40874 -> 172.17.0.2:5672 (1)')
	//     │   │   │   └── '172.17.0.1:40874 -> 172.17.0.2:5672' (connection client='https://github.com/streadway/amqp', host='172.17.0.2:5672', peer='172.17.0.1:40874')
	//     │   │   └── another_consumer w/ faulty channel (consumer user='', chan='')
	//     │   └── direct-q2 (queue, key='direct-q2', running, [D])
	//     ├── test-fanout (exchange, type 'fanout', [D])
	//     │   ├── fanout-q1 (queue, idle since 2017-05-25 19:14:32, [D])
	//     │   └── fanout-q2 (queue, idle since 2017-05-25 19:14:32, [D])
	//     ├── test-headers (exchange, type 'headers', [D|AD])
	//     │   ├── header-q1 (queue, key='headers-q1', idle since 2017-05-25 19:14:53, [D])
	//     │   └── header-q2 (queue, key='headers-q2', idle since 2017-05-25 19:14:47, [D])
	//     └── test-topic (exchange, type 'topic', [D])
	//         ├── topic-q1 (queue, key='topic-q1', idle since 2017-05-25 19:14:17, [D|AD|EX])
	//         └── topic-q2 (queue, key='topic-q2', idle since 2017-05-25 19:14:21, [D])

}

func ExampleBrokerInfoPrinter_printWithQueueFilter() {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := rabtap.NewRabbitHTTPClient(mock.URL, &tls.Config{})

	queueFilter, err := NewPredicateExpression("queue.Name == 'fanout-q2'")
	if err != nil {
		log.Fatal(err)
	}

	brokerInfoPrinter := NewBrokerInfoPrinter(
		BrokerInfoPrinterConfig{
			ShowStats:           false,
			ShowConsumers:       true,
			ShowDefaultExchange: false,
			QueueFilter:         queueFilter,
			OmitEmptyExchanges:  true,
			NoColor:             true},
	)
	brokerInfo, err := client.BrokerInfo()
	if err != nil {
		log.Fatal(err)
	}

	if err := brokerInfoPrinter.Print(brokerInfo, "http://rabbitmq/api", os.Stdout); err != nil {
		log.Fatal(err)
	}

	// Output:
	// http://rabbitmq/api (broker ver='3.6.9', mgmt ver='3.6.9', cluster='rabbit@08f57d1fe8ab')
	// └── Vhost /
	//     └── test-fanout (exchange, type 'fanout', [D])
	//         └── fanout-q2 (queue, idle since 2017-05-25 19:14:32, [D])

}
