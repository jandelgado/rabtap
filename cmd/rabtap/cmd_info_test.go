// Copyright (C) 2017 Jan Delgado
// component tests of the text format info renderer called through cmdInfo
// top level entry point

package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"net/url"
	"strings"
	"testing"

	"github.com/fatih/color"
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

func TestCmdInfoByExchangeInTextFormatProducesExpectedTree(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := rabtap.NewRabbitHTTPClient(url, &tls.Config{})

	var actual bytes.Buffer
	rootURL, _ := url.Parse("http://rabbitmq/api")
	color.NoColor = true
	cmdInfo(context.TODO(),
		CmdInfoArg{
			rootNode: rootURL,
			client:   client,
			treeConfig: BrokerInfoTreeBuilderConfig{
				Mode:                "byExchange",
				ShowConsumers:       true,
				ShowDefaultExchange: false,
				Filter:              TruePredicate,
				OmitEmptyExchanges:  false},
			renderConfig: BrokerInfoRendererConfig{
				Format:    "text",
				ShowStats: false},
			out: &actual})

	expected := `http://rabbitmq/api (broker ver='3.6.9', mgmt ver='3.6.9', cluster='rabbit@08f57d1fe8ab')
└─ Vhost /
   ├─ amq.direct (exchange(direct), [D])
   ├─ amq.fanout (exchange(fanout), [D])
   ├─ amq.headers (exchange(headers), [D])
   ├─ amq.match (exchange(headers), [D])
   ├─ amq.rabbitmq.log (exchange(topic), [D|I])
   ├─ amq.rabbitmq.trace (exchange(topic), [D|I])
   ├─ amq.topic (exchange(topic), [D])
   │  └─ test-topic (exchange(topic), key='test', [D])
   ├─ test-direct (exchange(direct), [D|AD|I])
   │  ├─ direct-q1 (queue(classic), key='direct-q1',  running, [D])
   │  │  ├─ ? (connection )
   │  │  │  └─ ? (channel )
   │  │  │     └─ another_consumer w/ faulty channel (consumer prefetch=0, ack_req=no, active=no, status=)
   │  │  └─ '172.17.0.1:40874 -> 172.17.0.2:5672' (connection guest@172.17.0.2:5672, state='running', client='https://github.com/streadway/amqp', ver='β', peer='172.17.0.1:40874')
   │  │     └─ '172.17.0.1:40874 -> 172.17.0.2:5672 (1)' (channel prefetch=0, state=running, unacked=0, confirms=no)
   │  │        └─ another_consumer w/ faulty channel (consumer prefetch=0, ack_req=no, active=no, status=)
   │  └─ direct-q2 (queue(classic), key='direct-q2',  running, [D])
   ├─ test-fanout (exchange(fanout), [D])
   │  ├─ fanout-q1 (queue(classic),  idle since 2017-05-25 19:14:32, [D])
   │  └─ fanout-q2 (queue(classic),  idle since 2017-05-25 19:14:32, [D])
   ├─ test-headers (exchange(headers), [D|AD])
   │  ├─ header-q1 (queue(classic), key='headers-q1',  idle since 2017-05-25 19:14:53, [D])
   │  └─ header-q2 (queue(classic), key='headers-q2',  idle since 2017-05-25 19:14:47, [D])
   └─ test-topic (exchange(topic), [D])
      ├─ topic-q1 (queue(classic), key='topic-q1',  idle since 2017-05-25 19:14:17, [D|AD|EX])
      └─ topic-q2 (queue(classic), key='topic-q2',  idle since 2017-05-25 19:14:21, [D])
`

	assert.Equal(t, expected, actual.String())
}

func TestCmdInfoByConnectionInTextFormatProducesExpectedTree(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := rabtap.NewRabbitHTTPClient(url, &tls.Config{})
	rootURL, _ := url.Parse("http://rabbitmq/api")

	var actual bytes.Buffer
	color.NoColor = true
	cmdInfo(context.TODO(),
		CmdInfoArg{
			rootNode: rootURL,
			client:   client,
			treeConfig: BrokerInfoTreeBuilderConfig{
				Mode:                "byConnection",
				ShowConsumers:       true,
				ShowDefaultExchange: false,
				Filter:              TruePredicate,
				OmitEmptyExchanges:  false},
			renderConfig: BrokerInfoRendererConfig{
				Format:    "text",
				ShowStats: false},
			out: &actual})

	expected := `http://rabbitmq/api (broker ver='3.6.9', mgmt ver='3.6.9', cluster='rabbit@08f57d1fe8ab')
└─ Vhost /
   └─ '172.17.0.1:40874 -> 172.17.0.2:5672' (connection guest@172.17.0.2:5672, state='running', client='https://github.com/streadway/amqp', ver='β', peer='172.17.0.1:40874')
      └─ '172.17.0.1:40874 -> 172.17.0.2:5672 (1)' (channel prefetch=0, state=running, unacked=0, confirms=no)
         └─ some_consumer (consumer prefetch=0, ack_req=no, active=no, status=)
            └─ direct-q1 (queue(classic),  running, [D])
`
	assert.Equal(t, expected, actual.String())

}

const expectedResultDotByExchange = `graph broker {
"root" [shape="record", label="{RabbitMQ 3.6.9 |http://rabbitmq/api |rabbit@08f57d1fe8ab }"];

"root" -- "vhost_/";
"vhost_/" [shape="box", label="Virtual host /"];

"vhost_/" -- "exchange_/_amq.direct"[headport=n];
"vhost_/" -- "exchange_/_amq.fanout"[headport=n];
"vhost_/" -- "exchange_/_amq.headers"[headport=n];
"vhost_/" -- "exchange_/_amq.match"[headport=n];
"vhost_/" -- "exchange_/_amq.rabbitmq.log"[headport=n];
"vhost_/" -- "exchange_/_amq.rabbitmq.trace"[headport=n];
"vhost_/" -- "exchange_/_amq.topic"[headport=n];
"vhost_/" -- "exchange_/_test-direct"[headport=n];
"vhost_/" -- "exchange_/_test-fanout"[headport=n];
"vhost_/" -- "exchange_/_test-headers"[headport=n];
"vhost_/" -- "exchange_/_test-topic"[headport=n];

"exchange_/_amq.direct" [shape="record"; label="{ amq.direct |direct | { D  | | } }"];


"exchange_/_amq.fanout" [shape="record"; label="{ amq.fanout |fanout | { D  | | } }"];


"exchange_/_amq.headers" [shape="record"; label="{ amq.headers |headers | { D  | | } }"];


"exchange_/_amq.match" [shape="record"; label="{ amq.match |headers | { D  | | } }"];


"exchange_/_amq.rabbitmq.log" [shape="record"; label="{ amq.rabbitmq.log |topic | { D  | | I  } }"];


"exchange_/_amq.rabbitmq.trace" [shape="record"; label="{ amq.rabbitmq.trace |topic | { D  | | I  } }"];


"exchange_/_amq.topic" [shape="record"; label="{ amq.topic |topic | { D  | | } }"];

"exchange_/_amq.topic" -- "exchange_/_test-topic" [fontsize=10; headport=n; label="test"];

"exchange_/_test-topic" [shape="record"; label="{ test-topic |topic | { D  | | } }"];


"exchange_/_test-direct" [shape="record"; label="{ test-direct |direct | { D  | AD  | I  } }"];

"exchange_/_test-direct" -- "queue_/_direct-q1" [fontsize=10; headport=n; label="direct-q1"];
"exchange_/_test-direct" -- "queue_/_direct-q2" [fontsize=10; headport=n; label="direct-q2"];

"queue_/_direct-q1" [shape="record"; label="{ direct-q1 | { D  | | } }"];


"queue_/_direct-q2" [shape="record"; label="{ direct-q2 | { D  | | } }"];


"exchange_/_test-fanout" [shape="record"; label="{ test-fanout |fanout | { D  | | } }"];

"exchange_/_test-fanout" -- "queue_/_fanout-q1" [fontsize=10; headport=n; label=""];
"exchange_/_test-fanout" -- "queue_/_fanout-q2" [fontsize=10; headport=n; label=""];

"queue_/_fanout-q1" [shape="record"; label="{ fanout-q1 | { D  | | } }"];


"queue_/_fanout-q2" [shape="record"; label="{ fanout-q2 | { D  | | } }"];


"exchange_/_test-headers" [shape="record"; label="{ test-headers |headers | { D  | AD  | } }"];

"exchange_/_test-headers" -- "queue_/_header-q1" [fontsize=10; headport=n; label="headers-q1"];
"exchange_/_test-headers" -- "queue_/_header-q2" [fontsize=10; headport=n; label="headers-q2"];

"queue_/_header-q1" [shape="record"; label="{ header-q1 | { D  | | } }"];


"queue_/_header-q2" [shape="record"; label="{ header-q2 | { D  | | } }"];


"exchange_/_test-topic" [shape="record"; label="{ test-topic |topic | { D  | | } }"];

"exchange_/_test-topic" -- "queue_/_topic-q1" [fontsize=10; headport=n; label="topic-q1"];
"exchange_/_test-topic" -- "queue_/_topic-q2" [fontsize=10; headport=n; label="topic-q2"];

"queue_/_topic-q1" [shape="record"; label="{ topic-q1 | { D  | AD  | EX  } }"];


"queue_/_topic-q2" [shape="record"; label="{ topic-q2 | { D  | | } }"];

}`

func TestCmdInfoByExchangeInDotFormat(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := rabtap.NewRabbitHTTPClient(url, &tls.Config{})
	rootURL, _ := url.Parse("http://rabbitmq/api")

	var actual bytes.Buffer
	cmdInfo(
		context.TODO(),
		CmdInfoArg{
			rootNode: rootURL,
			client:   client,
			treeConfig: BrokerInfoTreeBuilderConfig{
				Mode:                "byExchange",
				ShowConsumers:       false,
				ShowDefaultExchange: false,
				Filter:              TruePredicate,
				OmitEmptyExchanges:  false},
			renderConfig: BrokerInfoRendererConfig{Format: "dot"},
			out:          &actual})

	assert.Equal(t, strings.Trim(expectedResultDotByExchange, " \n"),
		strings.Trim(actual.String(), " \n"))
}

func TestCmdInfoByConnectionInDotFormat(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := rabtap.NewRabbitHTTPClient(url, &tls.Config{})
	rootURL, _ := url.Parse("http://rabbitmq/api")

	var actual bytes.Buffer
	cmdInfo(
		context.TODO(),
		CmdInfoArg{
			rootNode: rootURL,
			client:   client,
			treeConfig: BrokerInfoTreeBuilderConfig{
				Mode:                "byConnection",
				ShowConsumers:       false,
				ShowDefaultExchange: false,
				Filter:              TruePredicate,
				OmitEmptyExchanges:  false},
			renderConfig: BrokerInfoRendererConfig{Format: "dot"},
			out:          &actual})

	const expected = `graph broker {
"root" [shape="record", label="{RabbitMQ 3.6.9 |http://rabbitmq/api |rabbit@08f57d1fe8ab }"];

"root" -- "vhost_/";
"vhost_/" [shape="box", label="Virtual host /"];

"vhost_/" -- "connection_172.17.0.1:40874 -> 172.17.0.2:5672"[headport=n];

"connection_172.17.0.1:40874 -> 172.17.0.2:5672" [shape="record" label="172.17.0.1:40874 -&gt; 172.17.0.2:5672"];

"connection_172.17.0.1:40874 -> 172.17.0.2:5672" -- "channel_172.17.0.1:40874 -> 172.17.0.2:5672 (1)"
"channel_172.17.0.1:40874 -> 172.17.0.2:5672 (1)" [shape="record" label="172.17.0.1:40874 -&gt; 172.17.0.2:5672 (1)"];

"channel_172.17.0.1:40874 -> 172.17.0.2:5672 (1)" -- "consumer_some_consumer"
"consumer_some_consumer" [shape="record" label="some_consumer"];

"consumer_some_consumer" -- "queue_/_direct-q1"
"queue_/_direct-q1" [shape="record"; label="{ direct-q1 | { D  | | } }"];

}`
	assert.Equal(t, strings.Trim(expected, " \n"),
		strings.Trim(actual.String(), " \n"))
}
