// Copyright (C) 2017 Jan Delgado
// component tests of the text format info renderer called through cmdInfo
// top level entry point

package main

import (
	"crypto/tls"
	"net/url"
	"os"
	"strings"
	"testing"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
)

/****
func TestBrokerInfoPrintFailsOnInvalidUri(t *testing.T) {
	// TODO
	// brokerInfoPrinter := NewBrokerInfoTreeBuilder(BrokerInfoTreeBuilderConfig{})
	// err := brokerInfoPrinter.Print(rabtap.BrokerInfo{}, "//:xxx::invalid uri", os.Stdout)
	// assert.NotNil(t, err)

}
***/

func Example_cmdInfoByExchangeInTextFormat() {

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

func Example_cmdInfoByConnectionInTextFormat() {

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

const expectedResultDotByExchange = `graph broker {
"root" [shape="record", label="{RabbitMQ 3.6.9 |http://rabbitmq/api |rabbit@08f57d1fe8ab }"];

"root" -- "vhost_/";
"vhost_/" [shape="box", label="Virtual host /"];

"vhost_/" -- "exchange_amq.direct"[headport=n];
"vhost_/" -- "exchange_amq.fanout"[headport=n];
"vhost_/" -- "exchange_amq.headers"[headport=n];
"vhost_/" -- "exchange_amq.match"[headport=n];
"vhost_/" -- "exchange_amq.rabbitmq.log"[headport=n];
"vhost_/" -- "exchange_amq.rabbitmq.trace"[headport=n];
"vhost_/" -- "exchange_amq.topic"[headport=n];
"vhost_/" -- "exchange_test-direct"[headport=n];
"vhost_/" -- "exchange_test-fanout"[headport=n];
"vhost_/" -- "exchange_test-headers"[headport=n];
"vhost_/" -- "exchange_test-topic"[headport=n];

"exchange_amq.direct" [shape="record"; label="{ amq.direct |direct | { D  | | } }"];


"exchange_amq.fanout" [shape="record"; label="{ amq.fanout |fanout | { D  | | } }"];


"exchange_amq.headers" [shape="record"; label="{ amq.headers |headers | { D  | | } }"];


"exchange_amq.match" [shape="record"; label="{ amq.match |headers | { D  | | } }"];


"exchange_amq.rabbitmq.log" [shape="record"; label="{ amq.rabbitmq.log |topic | { D  | | I  } }"];


"exchange_amq.rabbitmq.trace" [shape="record"; label="{ amq.rabbitmq.trace |topic | { D  | | I  } }"];


"exchange_amq.topic" [shape="record"; label="{ amq.topic |topic | { D  | | } }"];


"exchange_test-direct" [shape="record"; label="{ test-direct |direct | { D  | AD  | I  } }"];

"exchange_test-direct" -- "boundqueue_direct-q1" [fontsize=10; headport=n; label="direct-q1"];
"exchange_test-direct" -- "boundqueue_direct-q2" [fontsize=10; headport=n; label="direct-q2"];

"boundqueue_direct-q1" [shape="record"; label="{ direct-q1 | { D  | | } }"];


"boundqueue_direct-q2" [shape="record"; label="{ direct-q2 | { D  | | } }"];


"exchange_test-fanout" [shape="record"; label="{ test-fanout |fanout | { D  | | } }"];

"exchange_test-fanout" -- "boundqueue_fanout-q1" [fontsize=10; headport=n; label=""];
"exchange_test-fanout" -- "boundqueue_fanout-q2" [fontsize=10; headport=n; label=""];

"boundqueue_fanout-q1" [shape="record"; label="{ fanout-q1 | { D  | | } }"];


"boundqueue_fanout-q2" [shape="record"; label="{ fanout-q2 | { D  | | } }"];


"exchange_test-headers" [shape="record"; label="{ test-headers |headers | { D  | AD  | } }"];

"exchange_test-headers" -- "boundqueue_header-q1" [fontsize=10; headport=n; label="headers-q1"];
"exchange_test-headers" -- "boundqueue_header-q2" [fontsize=10; headport=n; label="headers-q2"];

"boundqueue_header-q1" [shape="record"; label="{ header-q1 | { D  | | } }"];


"boundqueue_header-q2" [shape="record"; label="{ header-q2 | { D  | | } }"];


"exchange_test-topic" [shape="record"; label="{ test-topic |topic | { D  | | } }"];

"exchange_test-topic" -- "boundqueue_topic-q1" [fontsize=10; headport=n; label="topic-q1"];
"exchange_test-topic" -- "boundqueue_topic-q2" [fontsize=10; headport=n; label="topic-q2"];

"boundqueue_topic-q1" [shape="record"; label="{ topic-q1 | { D  | AD  | EX  } }"];


"boundqueue_topic-q2" [shape="record"; label="{ topic-q2 | { D  | | } }"];

}`

func TestCmdInfoByExchangeInDotFormat(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := rabtap.NewRabbitHTTPClient(url, &tls.Config{})

	testfunc := func() {
		cmdInfo(CmdInfoArg{
			rootNode: "http://rabbitmq/api",
			client:   client,
			treeConfig: BrokerInfoTreeBuilderConfig{
				Mode:                "byExchange",
				ShowConsumers:       false,
				ShowDefaultExchange: false,
				QueueFilter:         TruePredicate,
				OmitEmptyExchanges:  false},
			renderConfig: BrokerInfoRendererConfig{Format: "dot"},
			out:          os.Stdout})
	}
	result := testcommon.CaptureOutput(testfunc)
	assert.Equal(t, strings.Trim(expectedResultDotByExchange, " \n"),
		strings.Trim(result, " \n"))
}

const expectedResultDotByConnection = `graph broker {
"root" [shape="record", label="{RabbitMQ 3.6.9 |http://rabbitmq/api |rabbit@08f57d1fe8ab }"];

"root" -- "vhost_/";
"vhost_/" [shape="box", label="Virtual host /"];

"vhost_/" -- "connection_172.17.0.1:40874 -> 172.17.0.2:5672"[headport=n];

"connection_172.17.0.1:40874 -> 172.17.0.2:5672" [shape="recored" label="172.17.0.1:40874 -> 172.17.0.2:5672"];

"connection_172.17.0.1:40874 -> 172.17.0.2:5672" -- "consumer_some_consumer"
"consumer_some_consumer" [shape="recored" label="some_consumer"];

"consumer_some_consumer" -- "queue_direct-q1"
"queue_direct-q1" [shape="record"; label="{ direct-q1 | { D  | | } }"];

}`

func TestCmdInfoByConnectionInDotFormat(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := rabtap.NewRabbitHTTPClient(url, &tls.Config{})

	testfunc := func() {
		cmdInfo(CmdInfoArg{
			rootNode: "http://rabbitmq/api",
			client:   client,
			treeConfig: BrokerInfoTreeBuilderConfig{
				Mode:                "byConnection",
				ShowConsumers:       false,
				ShowDefaultExchange: false,
				QueueFilter:         TruePredicate,
				OmitEmptyExchanges:  false},
			renderConfig: BrokerInfoRendererConfig{Format: "dot"},
			out:          os.Stdout})
	}
	result := testcommon.CaptureOutput(testfunc)
	assert.Equal(t, strings.Trim(expectedResultDotByConnection, " \n"),
		strings.Trim(result, " \n"))
}
