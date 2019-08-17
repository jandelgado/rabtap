// Copyright (C) 2017 Jan Delgado
// component tests of the text format info renderer called through cmdInfo
// top level entry point

package main

import (
	"crypto/tls"
	"net/url"
	"os"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/jandelgado/rabtap/pkg/testcommon"
)

/****
func TestBrokerInfoPrintFailsOnInvalidUri(t *testing.T) {
	// TODO
	// brokerInfoPrinter := NewBrokerInfoTreeBuilder(BrokerInfoTreeBuilderConfig{})
	// err := brokerInfoPrinter.Print(rabtap.BrokerInfo{}, "//:xxx::invalid uri", os.Stdout)
	// assert.NotNil(t, err)

}
***/

func Example_cmdInfoByExchange() {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := rabtap.NewRabbitHTTPClient(url, &tls.Config{})

	cmdInfo(CmdInfoArg{
		rootNode: "http://rabbitmq/api",
		client:   client,
		treeConfig: BrokerInfoTreeBuilderConfig{
			Mode:                "byExchange",
			ShowConsumers:       true,
			ShowDefaultExchange: false,
			QueueFilter:         TruePredicate,
			OmitEmptyExchanges:  false},
		renderConfig: BrokerInfoRendererConfig{
			Format:    "text",
			ShowStats: false,
			NoColor:   true},
		out: os.Stdout})

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
	//     │   │   ├── some_consumer (consumer user='guest', prefetch=0, chan='172.17.0.1:40874 -> 172.17.0.2:5672 (1)')
	//     │   │   │   └── '172.17.0.1:40874 -> 172.17.0.2:5672' (connection client='https://github.com/streadway/amqp', host='172.17.0.2:5672', peer='172.17.0.1:40874')
	//     │   │   └── another_consumer w/ faulty channel (consumer user='', prefetch=0, chan='')
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

func Example_cmdInfoByConnection() {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := rabtap.NewRabbitHTTPClient(url, &tls.Config{})

	cmdInfo(CmdInfoArg{
		rootNode: "http://rabbitmq/api",
		client:   client,
		treeConfig: BrokerInfoTreeBuilderConfig{
			Mode:                "byConnection",
			ShowConsumers:       true,
			ShowDefaultExchange: false,
			QueueFilter:         TruePredicate,
			OmitEmptyExchanges:  false},
		renderConfig: BrokerInfoRendererConfig{
			Format:    "text",
			ShowStats: false,
			NoColor:   true},
		out: os.Stdout})

	// Output:
	// http://rabbitmq/api (broker ver='3.6.9', mgmt ver='3.6.9', cluster='rabbit@08f57d1fe8ab')
	// └── Vhost /
	//     └── '172.17.0.1:40874 -> 172.17.0.2:5672' (connection client='https://github.com/streadway/amqp', host='172.17.0.2:5672', peer='172.17.0.1:40874')
	//         └── some_consumer (consumer user='guest', prefetch=0, chan='172.17.0.1:40874 -> 172.17.0.2:5672 (1)')
	//             └── direct-q1 (queue, running, [D])
}