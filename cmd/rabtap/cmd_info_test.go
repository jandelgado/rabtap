// Copyright (C) 2017 Jan Delgado
// component tests of the text format info renderer called through cmdInfo
// top level entry point

package main

import (
	"crypto/tls"
	"net/url"
	"os"
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
	//     │   ├── direct-q1 (queue, key='direct-q1', running, [D|DLX])
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
	//         ├── topic-q1 (queue, key='topic-q1', idle since 2017-05-25 19:14:17, [D|AD|EX|DLX])
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
	//             └── direct-q1 (queue, running, [D|DLX])
}

const expectedResultDotByExchange = `digraph broker {
"root" [shape="record", label="{RabbitMQ 3.6.9 |http://rabbitmq/api |rabbit@08f57d1fe8ab }"];
"root" -> "vhost_/";
"vhost_/" [shape="box", label="Virtual host\n/"];

{ rank = same; "exchange_amq.direct"; "exchange_amq.fanout"; "exchange_amq.headers"; "exchange_amq.match"; "exchange_amq.rabbitmq.log"; "exchange_amq.rabbitmq.trace"; "exchange_amq.topic"; "exchange_test-direct"; "exchange_test-fanout"; "exchange_test-headers"; "exchange_test-topic"; };
"vhost_/" -> "exchange_amq.direct"[headport=n];
"vhost_/" -> "exchange_amq.fanout"[headport=n];
"vhost_/" -> "exchange_amq.headers"[headport=n];
"vhost_/" -> "exchange_amq.match"[headport=n];
"vhost_/" -> "exchange_amq.rabbitmq.log"[headport=n];
"vhost_/" -> "exchange_amq.rabbitmq.trace"[headport=n];
"vhost_/" -> "exchange_amq.topic"[headport=n];
"vhost_/" -> "exchange_test-direct"[headport=n];
"vhost_/" -> "exchange_test-fanout"[headport=n];
"vhost_/" -> "exchange_test-headers"[headport=n];
"vhost_/" -> "exchange_test-topic"[headport=n];
"exchange_amq.direct" [shape="none"; margin="0"; label=< 
		    <TABLE border='0' cellborder='1' cellspacing='0'>
              <TR><TD colspan='1' WIDTH='33%'> E </TD><TD colspan='2' balign='center'>direct</TD></TR>
              <TR><TD colspan='3' align='text'><B>amq.direct</B></TD></TR>
			  <TR><TD WIDTH='33%'> D </TD>
				  <TD WIDTH='33%'></TD>
				  <TD WIDTH='33%'></TD></TR>
		    </TABLE> >];
	"exchange_amq.fanout" [shape="none"; margin="0"; label=< 
		    <TABLE border='0' cellborder='1' cellspacing='0'>
              <TR><TD colspan='1' WIDTH='33%'> E </TD><TD colspan='2' balign='center'>fanout</TD></TR>
              <TR><TD colspan='3' align='text'><B>amq.fanout</B></TD></TR>
			  <TR><TD WIDTH='33%'> D </TD>
				  <TD WIDTH='33%'></TD>
				  <TD WIDTH='33%'></TD></TR>
		    </TABLE> >];
	"exchange_amq.headers" [shape="none"; margin="0"; label=< 
		    <TABLE border='0' cellborder='1' cellspacing='0'>
              <TR><TD colspan='1' WIDTH='33%'> E </TD><TD colspan='2' balign='center'>headers</TD></TR>
              <TR><TD colspan='3' align='text'><B>amq.headers</B></TD></TR>
			  <TR><TD WIDTH='33%'> D </TD>
				  <TD WIDTH='33%'></TD>
				  <TD WIDTH='33%'></TD></TR>
		    </TABLE> >];
	"exchange_amq.match" [shape="none"; margin="0"; label=< 
		    <TABLE border='0' cellborder='1' cellspacing='0'>
              <TR><TD colspan='1' WIDTH='33%'> E </TD><TD colspan='2' balign='center'>headers</TD></TR>
              <TR><TD colspan='3' align='text'><B>amq.match</B></TD></TR>
			  <TR><TD WIDTH='33%'> D </TD>
				  <TD WIDTH='33%'></TD>
				  <TD WIDTH='33%'></TD></TR>
		    </TABLE> >];
	"exchange_amq.rabbitmq.log" [shape="none"; margin="0"; label=< 
		    <TABLE border='0' cellborder='1' cellspacing='0'>
              <TR><TD colspan='1' WIDTH='33%'> E </TD><TD colspan='2' balign='center'>topic</TD></TR>
              <TR><TD colspan='3' align='text'><B>amq.rabbitmq.log</B></TD></TR>
			  <TR><TD WIDTH='33%'> D </TD>
				  <TD WIDTH='33%'></TD>
				  <TD WIDTH='33%'> I </TD></TR>
		    </TABLE> >];
	"exchange_amq.rabbitmq.trace" [shape="none"; margin="0"; label=< 
		    <TABLE border='0' cellborder='1' cellspacing='0'>
              <TR><TD colspan='1' WIDTH='33%'> E </TD><TD colspan='2' balign='center'>topic</TD></TR>
              <TR><TD colspan='3' align='text'><B>amq.rabbitmq.trace</B></TD></TR>
			  <TR><TD WIDTH='33%'> D </TD>
				  <TD WIDTH='33%'></TD>
				  <TD WIDTH='33%'> I </TD></TR>
		    </TABLE> >];
	"exchange_amq.topic" [shape="none"; margin="0"; label=< 
		    <TABLE border='0' cellborder='1' cellspacing='0'>
              <TR><TD colspan='1' WIDTH='33%'> E </TD><TD colspan='2' balign='center'>topic</TD></TR>
              <TR><TD colspan='3' align='text'><B>amq.topic</B></TD></TR>
			  <TR><TD WIDTH='33%'> D </TD>
				  <TD WIDTH='33%'></TD>
				  <TD WIDTH='33%'></TD></TR>
		    </TABLE> >];
	"exchange_test-direct" [shape="none"; margin="0"; label=< 
		    <TABLE border='0' cellborder='1' cellspacing='0'>
              <TR><TD colspan='1' WIDTH='33%'> E </TD><TD colspan='2' balign='center'>direct</TD></TR>
              <TR><TD colspan='3' align='text'><B>test-direct</B></TD></TR>
			  <TR><TD WIDTH='33%'> D </TD>
				  <TD WIDTH='33%'> AD </TD>
				  <TD WIDTH='33%'> I </TD></TR>
		    </TABLE> >];
	"exchange_test-direct" -> "boundqueue_direct-q1" [fontsize=10; label="direct-q1"];
"exchange_test-direct" -> "boundqueue_direct-q2" [fontsize=10; label="direct-q2"];

"boundqueue_direct-q1" [shape="none"; margin="0"; label=<
		     <TABLE border='0' cellborder='1' cellspacing='0'>
			   <TR><TD colspan='1' WIDTH='25%'> Q</TD><TD colspan='3'><B>direct-q1</B></TD></TR>
			   <TR><TD WIDTH='25%'> D </TD> 
			       <TD WIDTH='25%'> </TD>
			       <TD WIDTH='25%'> </TD>
			       <TD  WIDTH='25%' PORT='dlx'> DLX </TD></TR>
				</TABLE> >];
"boundqueue_direct-q1":dlx -> "exchange_mydlx" [style="dashed"];

"boundqueue_direct-q2" [shape="none"; margin="0"; label=<
		     <TABLE border='0' cellborder='1' cellspacing='0'>
			   <TR><TD colspan='1' WIDTH='25%'> Q</TD><TD colspan='3'><B>direct-q2</B></TD></TR>
			   <TR><TD WIDTH='25%'> D </TD> 
			       <TD WIDTH='25%'> </TD>
			       <TD WIDTH='25%'> </TD>
			       <TD  WIDTH='25%' PORT='dlx'></TD></TR>
				</TABLE> >];

"exchange_test-fanout" [shape="none"; margin="0"; label=< 
		    <TABLE border='0' cellborder='1' cellspacing='0'>
              <TR><TD colspan='1' WIDTH='33%'> E </TD><TD colspan='2' balign='center'>fanout</TD></TR>
              <TR><TD colspan='3' align='text'><B>test-fanout</B></TD></TR>
			  <TR><TD WIDTH='33%'> D </TD>
				  <TD WIDTH='33%'></TD>
				  <TD WIDTH='33%'></TD></TR>
		    </TABLE> >];
	"exchange_test-fanout" -> "boundqueue_fanout-q1" [fontsize=10; label=""];
"exchange_test-fanout" -> "boundqueue_fanout-q2" [fontsize=10; label=""];

"boundqueue_fanout-q1" [shape="none"; margin="0"; label=<
		     <TABLE border='0' cellborder='1' cellspacing='0'>
			   <TR><TD colspan='1' WIDTH='25%'> Q</TD><TD colspan='3'><B>fanout-q1</B></TD></TR>
			   <TR><TD WIDTH='25%'> D </TD> 
			       <TD WIDTH='25%'> </TD>
			       <TD WIDTH='25%'> </TD>
			       <TD  WIDTH='25%' PORT='dlx'></TD></TR>
				</TABLE> >];


"boundqueue_fanout-q2" [shape="none"; margin="0"; label=<
		     <TABLE border='0' cellborder='1' cellspacing='0'>
			   <TR><TD colspan='1' WIDTH='25%'> Q</TD><TD colspan='3'><B>fanout-q2</B></TD></TR>
			   <TR><TD WIDTH='25%'> D </TD> 
			       <TD WIDTH='25%'> </TD>
			       <TD WIDTH='25%'> </TD>
			       <TD  WIDTH='25%' PORT='dlx'></TD></TR>
				</TABLE> >];

"exchange_test-headers" [shape="none"; margin="0"; label=< 
		    <TABLE border='0' cellborder='1' cellspacing='0'>
              <TR><TD colspan='1' WIDTH='33%'> E </TD><TD colspan='2' balign='center'>headers</TD></TR>
              <TR><TD colspan='3' align='text'><B>test-headers</B></TD></TR>
			  <TR><TD WIDTH='33%'> D </TD>
				  <TD WIDTH='33%'> AD </TD>
				  <TD WIDTH='33%'></TD></TR>
		    </TABLE> >];
	"exchange_test-headers" -> "boundqueue_header-q1" [fontsize=10; label="headers-q1"];
"exchange_test-headers" -> "boundqueue_header-q2" [fontsize=10; label="headers-q2"];

"boundqueue_header-q1" [shape="none"; margin="0"; label=<
		     <TABLE border='0' cellborder='1' cellspacing='0'>
			   <TR><TD colspan='1' WIDTH='25%'> Q</TD><TD colspan='3'><B>header-q1</B></TD></TR>
			   <TR><TD WIDTH='25%'> D </TD> 
			       <TD WIDTH='25%'> </TD>
			       <TD WIDTH='25%'> </TD>
			       <TD  WIDTH='25%' PORT='dlx'></TD></TR>
				</TABLE> >];


"boundqueue_header-q2" [shape="none"; margin="0"; label=<
		     <TABLE border='0' cellborder='1' cellspacing='0'>
			   <TR><TD colspan='1' WIDTH='25%'> Q</TD><TD colspan='3'><B>header-q2</B></TD></TR>
			   <TR><TD WIDTH='25%'> D </TD> 
			       <TD WIDTH='25%'> </TD>
			       <TD WIDTH='25%'> </TD>
			       <TD  WIDTH='25%' PORT='dlx'></TD></TR>
				</TABLE> >];

"exchange_test-topic" [shape="none"; margin="0"; label=< 
		    <TABLE border='0' cellborder='1' cellspacing='0'>
              <TR><TD colspan='1' WIDTH='33%'> E </TD><TD colspan='2' balign='center'>topic</TD></TR>
              <TR><TD colspan='3' align='text'><B>test-topic</B></TD></TR>
			  <TR><TD WIDTH='33%'> D </TD>
				  <TD WIDTH='33%'></TD>
				  <TD WIDTH='33%'></TD></TR>
		    </TABLE> >];
	"exchange_test-topic" -> "boundqueue_topic-q1" [fontsize=10; label="topic-q1"];
"exchange_test-topic" -> "boundqueue_topic-q2" [fontsize=10; label="topic-q2"];

"boundqueue_topic-q1" [shape="none"; margin="0"; label=<
		     <TABLE border='0' cellborder='1' cellspacing='0'>
			   <TR><TD colspan='1' WIDTH='25%'> Q</TD><TD colspan='3'><B>topic-q1</B></TD></TR>
			   <TR><TD WIDTH='25%'> D </TD> 
			       <TD WIDTH='25%'> AD  </TD>
			       <TD WIDTH='25%'> EX  </TD>
			       <TD  WIDTH='25%' PORT='dlx'> DLX </TD></TR>
				</TABLE> >];
"boundqueue_topic-q1":dlx -> "exchange_mydlx" [style="dashed"];

"boundqueue_topic-q2" [shape="none"; margin="0"; label=<
		     <TABLE border='0' cellborder='1' cellspacing='0'>
			   <TR><TD colspan='1' WIDTH='25%'> Q</TD><TD colspan='3'><B>topic-q2</B></TD></TR>
			   <TR><TD WIDTH='25%'> D </TD> 
			       <TD WIDTH='25%'> </TD>
			       <TD WIDTH='25%'> </TD>
			       <TD  WIDTH='25%' PORT='dlx'></TD></TR>
				</TABLE> >];

}`

func TestCmdInfoByExchangeInDotFormat(t *testing.T) {

	api := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer api.Close()
	url, _ := url.Parse(api.URL)
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
	assert.Equal(t, expectedResultDotByExchange, result)
}

const expectedResultDotByConnection = `digraph broker {
"root" [shape="record", label="{RabbitMQ 3.6.9 |http://rabbitmq/api |rabbit@08f57d1fe8ab }"];
"root" -> "vhost_/";
"vhost_/" [shape="box", label="Virtual host\n/"];

{ rank = same; "connection_172.17.0.1:40874 -> 172.17.0.2:5672"; };
"vhost_/" -> "connection_172.17.0.1:40874 -> 172.17.0.2:5672"[headport=n];

"connection_172.17.0.1:40874 -> 172.17.0.2:5672" [shape="record" label="{ Conn | 172.17.0.1:40874 -> 172.17.0.2:5672 }"];
"connection_172.17.0.1:40874 -> 172.17.0.2:5672" -> "consumer_some_consumer"
"consumer_some_consumer" [shape="record" label="{ Cons | some_consumer }"];
"consumer_some_consumer" -> "queue_direct-q1"
"queue_direct-q1" [shape="none"; margin="0"; label=<
		     <TABLE border='0' cellborder='1' cellspacing='0'>
			   <TR><TD colspan='4' WIDTH='25%'> Q</TD><TD><B>direct-q1</B></TD></TR>
			   <TR><TD WIDTH='25%'> D </TD> 
			       <TD WIDTH='25%'> </TD>
			       <TD WIDTH='25%'> </TD>
			       <TD  WIDTH='25%' PORT='dlx'><dlx> DLX </TD></TR>
				</TABLE> >];
"queue_direct-q1":dlx -> "exchange_mydlx" [style="dashed"];}`

func TestCmdInfoByConnectionInDotFormat(t *testing.T) {

	api := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer api.Close()
	url, _ := url.Parse(api.URL)
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
	assert.Equal(t, expectedResultDotByConnection, result)
}
