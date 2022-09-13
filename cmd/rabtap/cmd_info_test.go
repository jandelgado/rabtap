// Copyright (C) 2017 Jan Delgado
// component tests of the text format info renderer called through cmdInfo
// top level entry point

package main

import (
	"context"
	"crypto/tls"
	"net/url"
	"os"
	"strings"
	"testing"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
)

func Example_startCmdInfo() {
	// TODO move to cmd_info_test
	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeEmpty)
	defer mock.Close()

	args, _ := ParseCommandLineArgs([]string{"info", "--api", mock.URL, "--no-color"})
	titleURL, _ := url.Parse("http://guest:guest@rootnode/vhost")
	startCmdInfo(context.TODO(), args, titleURL)

	// Output:
	// http://rootnode/vhost (broker ver='3.6.9', mgmt ver='3.6.9', cluster='rabbit@08f57d1fe8ab')
}

func Example_cmdInfoByExchangeInTextFormat() {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := rabtap.NewRabbitHTTPClient(url, &tls.Config{})

	rootURL, _ := url.Parse("http://rabbitmq/api")
	cmdInfo(context.TODO(),
		CmdInfoArg{
			rootNode: rootURL,
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
	// └─ Vhost /
	//    ├─ amq.direct (exchange(direct), [D])
	//    ├─ amq.fanout (exchange(fanout), [D])
	//    ├─ amq.headers (exchange(headers), [D])
	//    ├─ amq.match (exchange(headers), [D])
	//    ├─ amq.rabbitmq.log (exchange(topic), [D|I])
	//    ├─ amq.rabbitmq.trace (exchange(topic), [D|I])
	//    ├─ amq.topic (exchange(topic), [D])
	//    ├─ test-direct (exchange(direct), [D|AD|I])
	//    │  ├─ direct-q1 (queue(classic), key='direct-q1',  running, [D])
	//    │  │  ├─ '172.17.0.1:40874 -> 172.17.0.2:5672' (connection guest@172.17.0.2:5672, state='running', client='https://github.com/streadway/amqp', ver='β', peer='172.17.0.1:40874')
	//    │  │  │  └─ '172.17.0.1:40874 -> 172.17.0.2:5672 (1)' (channel prefetch=0, state=running, unacked=0, confirms=no)
	//    │  │  │     └─ some_consumer (consumer prefetch=0, ack_req=no, active=no, status=)
	//    │  │  └─ ? (connection)
	//    │  │     └─ ? (channel)
	//    │  │        └─ another_consumer w/ faulty channel (consumer prefetch=0, ack_req=no, active=no, status=)
	//    │  └─ direct-q2 (queue(classic), key='direct-q2',  running, [D])
	//    ├─ test-fanout (exchange(fanout), [D])
	//    │  ├─ fanout-q1 (queue(classic),  idle since 2017-05-25 19:14:32, [D])
	//    │  └─ fanout-q2 (queue(classic),  idle since 2017-05-25 19:14:32, [D])
	//    ├─ test-headers (exchange(headers), [D|AD])
	//    │  ├─ header-q1 (queue(classic), key='headers-q1',  idle since 2017-05-25 19:14:53, [D])
	//    │  └─ header-q2 (queue(classic), key='headers-q2',  idle since 2017-05-25 19:14:47, [D])
	//    └─ test-topic (exchange(topic), [D])
	//       ├─ topic-q1 (queue(classic), key='topic-q1',  idle since 2017-05-25 19:14:17, [D|AD|EX])
	//       ├─ topic-q2 (queue(classic), key='topic-q2',  idle since 2017-05-25 19:14:21, [D])
	//       └─ test-topic (exchange(topic), [D])
}

func Example_cmdInfoByConnectionInTextFormat() {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := rabtap.NewRabbitHTTPClient(url, &tls.Config{})
	rootURL, _ := url.Parse("http://rabbitmq/api")

	cmdInfo(context.TODO(),
		CmdInfoArg{
			rootNode: rootURL,
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
	// └─ Vhost /
	//    └─ '172.17.0.1:40874 -> 172.17.0.2:5672' (connection guest@172.17.0.2:5672, state='running', client='https://github.com/streadway/amqp', ver='β', peer='172.17.0.1:40874')
	//       └─ '172.17.0.1:40874 -> 172.17.0.2:5672 (1)' (channel prefetch=0, state=running, unacked=0, confirms=no)
	//          └─ some_consumer (consumer prefetch=0, ack_req=no, active=no, status=)
	//             └─ direct-q1 (queue(classic),  running, [D])
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

"exchange_test-direct" -- "queue_direct-q1" [fontsize=10; headport=n; label="direct-q1"];
"exchange_test-direct" -- "queue_direct-q2" [fontsize=10; headport=n; label="direct-q2"];

"queue_direct-q1" [shape="record"; label="{ direct-q1 | { D  | | } }"];


"queue_direct-q2" [shape="record"; label="{ direct-q2 | { D  | | } }"];


"exchange_test-fanout" [shape="record"; label="{ test-fanout |fanout | { D  | | } }"];

"exchange_test-fanout" -- "queue_fanout-q1" [fontsize=10; headport=n; label=""];
"exchange_test-fanout" -- "queue_fanout-q2" [fontsize=10; headport=n; label=""];

"queue_fanout-q1" [shape="record"; label="{ fanout-q1 | { D  | | } }"];


"queue_fanout-q2" [shape="record"; label="{ fanout-q2 | { D  | | } }"];


"exchange_test-headers" [shape="record"; label="{ test-headers |headers | { D  | AD  | } }"];

"exchange_test-headers" -- "queue_header-q1" [fontsize=10; headport=n; label="headers-q1"];
"exchange_test-headers" -- "queue_header-q2" [fontsize=10; headport=n; label="headers-q2"];

"queue_header-q1" [shape="record"; label="{ header-q1 | { D  | | } }"];


"queue_header-q2" [shape="record"; label="{ header-q2 | { D  | | } }"];


"exchange_test-topic" [shape="record"; label="{ test-topic |topic | { D  | | } }"];

"exchange_test-topic" -- "queue_topic-q1" [fontsize=10; headport=n; label="topic-q1"];
"exchange_test-topic" -- "queue_topic-q2" [fontsize=10; headport=n; label="topic-q2"];
"exchange_test-topic" -- "exchange_test-topic" [fontsize=10; headport=n; label=""];

"queue_topic-q1" [shape="record"; label="{ topic-q1 | { D  | AD  | EX  } }"];


"queue_topic-q2" [shape="record"; label="{ topic-q2 | { D  | | } }"];


"exchange_test-topic" [shape="record"; label="{ test-topic |topic | { D  | | } }"];

}`

func TestCmdInfoByExchangeInDotFormat(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := rabtap.NewRabbitHTTPClient(url, &tls.Config{})
	rootURL, _ := url.Parse("http://rabbitmq/api")

	testfunc := func() {
		cmdInfo(
			context.TODO(),
			CmdInfoArg{
				rootNode: rootURL,
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

"connection_172.17.0.1:40874 -> 172.17.0.2:5672" [shape="record" label="172.17.0.1:40874 -&gt; 172.17.0.2:5672"];

"connection_172.17.0.1:40874 -> 172.17.0.2:5672" -- "channel_172.17.0.1:40874 -> 172.17.0.2:5672 (1)"
"channel_172.17.0.1:40874 -> 172.17.0.2:5672 (1)" [shape="record" label="172.17.0.1:40874 -&gt; 172.17.0.2:5672 (1)"];

"channel_172.17.0.1:40874 -> 172.17.0.2:5672 (1)" -- "consumer_some_consumer"
"consumer_some_consumer" [shape="record" label="some_consumer"];

"consumer_some_consumer" -- "queue_direct-q1"
"queue_direct-q1" [shape="record"; label="{ direct-q1 | { D  | | } }"];

}`

func TestCmdInfoByConnectionInDotFormat(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := rabtap.NewRabbitHTTPClient(url, &tls.Config{})
	rootURL, _ := url.Parse("http://rabbitmq/api")

	testfunc := func() {
		cmdInfo(
			context.TODO(),
			CmdInfoArg{
				rootNode: rootURL,
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
