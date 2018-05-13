// Copyright (C) 2017 Jan Delgado

package main

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	"text/template"

	"github.com/jandelgado/rabtap/pkg"
)

const (
	tplRootNode = `
	    {{- printf "%s://%s%s" .URL.Scheme .URL.Host .URL.Path | URLColor }} 
	 	{{- if .Overview }} (broker ver='{{ .Overview.RabbitmqVersion }}',
		{{- "" }} mgmt ver='{{ .Overview.ManagementVersion }}',
		{{- "" }} cluster='{{ .Overview.ClusterName }}{{end}}')`

	tplConsumer = `
		{{- ConsumerColor .Consumer.ConsumerTag }} (consumer 
		{{- ""}} user='{{ .Consumer.ChannelDetails.User }}', chan='
		{{- .Consumer.ChannelDetails.Name }}')`

	tplConnection = ` 
	    {{- ""}}'{{ ConnectionColor .Connection.Name }}' (connection 
		{{- ""}} client='{{ .Connection.ClientProperties.Product}}',
		{{- ""}} host='{{ .Connection.Host }}:{{ .Connection.Port }}',
		{{- ""}} peer='{{ .Connection.PeerHost }}:{{ .Connection.PeerPort }}')`

	tplExchange = `
	    {{- ExchangeColor .PrintName }} (exchange, type '{{ .Exchange.Type  }}'
		{{- if and .Config.ShowStats .Exchange.MessageStats }}, in=(
		{{- .Exchange.MessageStats.PublishIn }}, {{printf "%.1f" .Exchange.MessageStats.PublishInDetails.Rate}}/s) msg, out=(
		{{- .Exchange.MessageStats.PublishOut }}, {{printf "%.1f" .Exchange.MessageStats.PublishOutDetails.Rate}}/s) msg
		{{- end }}, {{ .ExchangeFlags  }})`

	tplQueue = `
	    {{- QueueColor .Binding.Destination }} (queue,
		{{- with .Binding.RoutingKey }} key='{{ KeyColor .}}',{{end}}
		{{- with .Binding.Arguments}} args='{{ KeyColor .}}',{{end}}
		{{- if .Config.ShowStats }}
		{{- .Queue.Consumers  }} cons, (
		{{- .Queue.Messages }}, {{printf "%.1f" .Queue.MessagesDetails.Rate}}/s) msg, (
		{{- .Queue.MessagesReady }}, {{printf "%.1f" .Queue.MessagesReadyDetails.Rate}}/s) msg ready,
		{{- end }}
		{{- if .Queue.IdleSince}}{{- " idle since "}}{{ .Queue.IdleSince}}{{else}}{{ " running" }}{{end}}
		{{- ""}}, {{ .QueueFlags}})`
)

// BrokerInfoPrinterConfig controls bevaviour auf PrintBrokerInfo
type BrokerInfoPrinterConfig struct {
	ShowDefaultExchange bool
	ShowConsumers       bool
	ShowStats           bool
	NoColor             bool
}

// BrokerInfoPrinter prints nicely treeish infos desribing a brokers
// topology
type BrokerInfoPrinter struct {
	config    BrokerInfoPrinterConfig
	colorizer ColorPrinter
}

// NewBrokerInfoPrinter constructs a new object to print a broker info
func NewBrokerInfoPrinter(config BrokerInfoPrinterConfig) *BrokerInfoPrinter {
	s := BrokerInfoPrinter{
		config:    config,
		colorizer: NewColorPrinter(config.NoColor),
	}
	return &s
}

// findQueueByName searches in the queues array for a queue with the given
// name and vhost. RabbitQueue element is returned on succes, otherwise nil.
func findQueueByName(queues []rabtap.RabbitQueue,
	vhost, queueName string) *rabtap.RabbitQueue {
	for _, queue := range queues {
		if queue.Name == queueName && queue.Vhost == vhost {
			return &queue
		}
	}
	return nil
}

func findExchangeByName(exchanges []rabtap.RabbitExchange,
	vhost, exchangeName string) *rabtap.RabbitExchange {
	for _, exchange := range exchanges {
		if exchange.Name == exchangeName && exchange.Vhost == vhost {
			return &exchange
		}
	}
	return nil
}

// currently not used.
// func findChannelByName(channels []rabtap.RabbitChannel,
//     vhost, channelName string) *rabtap.RabbitChannel {
//     for _, channel := range channels {
//         if channel.Name == channelName && channel.Vhost == vhost {
//             return &channel
//         }
//     }
//     return nil
// }

func findConnectionByName(conns []rabtap.RabbitConnection,
	vhost, connName string) *rabtap.RabbitConnection {
	for _, conn := range conns {
		if conn.Name == connName && conn.Vhost == vhost {
			return &conn
		}
	}
	return nil
}

func findConsumerByQueue(consumers []rabtap.RabbitConsumer,
	vhost, queueName string) *rabtap.RabbitConsumer {
	for _, consumer := range consumers {
		if consumer.Queue.Vhost == vhost &&
			consumer.Queue.Name == queueName {
			return &consumer
		}
	}
	return nil
}

func getBindingsForExchange(exchange *rabtap.RabbitExchange, bindings []rabtap.RabbitBinding) []rabtap.RabbitBinding {
	var result []rabtap.RabbitBinding
	for _, binding := range bindings {
		if binding.Source == exchange.Name &&
			binding.Vhost == exchange.Vhost {
			result = append(result, binding)
		}
	}
	return result
}

func filterStringList(flags []bool, list []string) []string {
	result := []string{}
	for i, s := range list {
		if flags[i] {
			result = append(result, s)
		}
	}
	return result
}

// resolveTemplate resolves a template for use in the broker info printer,
// with support for colored output. name is just an informational name
// passed to the template ctor. tpl is the actual template and args
// the arguments used during rendering.
func (s BrokerInfoPrinter) resolveTemplate(name string, tpl string, args interface{}) string {
	tmpl := template.Must(template.New(name).Funcs(
		s.colorizer.GetFuncMap()).Parse(tpl))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, args)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func (s BrokerInfoPrinter) renderQueueFlagsAsString(queue *rabtap.RabbitQueue) string {
	flags := []bool{queue.Durable, queue.AutoDelete, queue.Exclusive}
	names := []string{"D", "AD", "EX"}
	return "[" + strings.Join(filterStringList(flags, names), "|") + "]"
}

func (s BrokerInfoPrinter) renderExchangeFlagsAsString(exchange *rabtap.RabbitExchange) string {
	flags := []bool{exchange.Durable, exchange.AutoDelete, exchange.Internal}
	names := []string{"D", "AD", "I"}
	return "[" + strings.Join(filterStringList(flags, names), "|") + "]"
}

func (s BrokerInfoPrinter) renderConsumerElementAsString(consumer *rabtap.RabbitConsumer) string {
	type ConsumerInfo struct {
		Config   BrokerInfoPrinterConfig
		Consumer *rabtap.RabbitConsumer
	}
	args := ConsumerInfo{s.config, consumer}
	return s.resolveTemplate("consumer-tpl", tplConsumer, args)
}

func (s BrokerInfoPrinter) renderConnectionElementAsString(conn *rabtap.RabbitConnection) string {
	type ConnectionInfo struct {
		Config     BrokerInfoPrinterConfig
		Connection *rabtap.RabbitConnection
	}
	args := ConnectionInfo{s.config, conn}
	return s.resolveTemplate("connnection-tpl", tplConnection, args)
}

func (s BrokerInfoPrinter) renderQueueElementAsString(queue *rabtap.RabbitQueue, binding *rabtap.RabbitBinding) string {

	queueFlags := s.renderQueueFlagsAsString(queue)
	type QueueInfo struct {
		Config     BrokerInfoPrinterConfig
		Binding    *rabtap.RabbitBinding
		Queue      *rabtap.RabbitQueue
		QueueFlags string
	}
	args := QueueInfo{s.config, binding, queue, queueFlags}
	return s.resolveTemplate("queue-tpl", tplQueue, args)
}

func (s BrokerInfoPrinter) renderRootNodeAsString(rabbitURL url.URL, overview rabtap.RabbitOverview) string {

	type RootInfo struct {
		Config   BrokerInfoPrinterConfig
		URL      url.URL
		Overview rabtap.RabbitOverview
	}
	args := RootInfo{s.config, rabbitURL, overview}

	return s.resolveTemplate("rootnode", tplRootNode, args)
}

func (s BrokerInfoPrinter) renderExchangeElementAsString(exchange *rabtap.RabbitExchange) string {
	printName := exchange.Name
	if printName == "" {
		printName = "(default)"
	}

	type ExchangeInfo struct {
		Config        BrokerInfoPrinterConfig
		Exchange      *rabtap.RabbitExchange
		ExchangeFlags string
		PrintName     string
	}
	exchangeFlags := s.renderExchangeFlagsAsString(exchange)
	args := ExchangeInfo{s.config, exchange, exchangeFlags, printName}
	return s.resolveTemplate("exchange-tpl", tplExchange, args)
}

// TODO remove exchange, use queue.Vhost
func (s BrokerInfoPrinter) addConsumerNodes(node *TreeNode,
	queue *rabtap.RabbitQueue, brokerInfo rabtap.BrokerInfo) {

	vhost := queue.Vhost
	for _, consumer := range brokerInfo.Consumers {
		if consumer.Queue.Vhost == vhost &&
			consumer.Queue.Name == queue.Name {
			consumerNode := node.AddChild(s.renderConsumerElementAsString(&consumer))
			s.addConnectionNode(consumerNode, vhost, consumer.ChannelDetails.ConnectionName, brokerInfo)
		}
	}
}

func (s BrokerInfoPrinter) addConnectionNode(node *TreeNode,
	vhost string, connName string, brokerInfo rabtap.BrokerInfo) {
	connInfo := findConnectionByName(brokerInfo.Connections, vhost, connName)
	if connInfo != nil {
		node.AddChild(s.renderConnectionElementAsString(connInfo))
	}
}

func (s BrokerInfoPrinter) addQueue(node *TreeNode, binding *rabtap.RabbitBinding, exchange *rabtap.RabbitExchange, brokerInfo rabtap.BrokerInfo) *TreeNode {
	// standard binding of queue to exchange
	queue := findQueueByName(brokerInfo.Queues,
		binding.Vhost,
		binding.Destination)
	if queue == nil {
		// we test for nil because (at least in theory) a queue can disappear
		// since we are making various non-transactional API calls
		queue = &rabtap.RabbitQueue{Name: binding.Destination}
	}
	queueText := s.renderQueueElementAsString(queue, binding)
	queueNode := node.AddChild(queueText)

	if s.config.ShowConsumers {
		s.addConsumerNodes(queueNode, queue, brokerInfo)
	}
	return queueNode
}

// addExchange recursively (in case of exchange-exchange binding) an exchange to the
// given node.
func (s BrokerInfoPrinter) addExchange(node *TreeNode,
	exchange *rabtap.RabbitExchange, brokerInfo rabtap.BrokerInfo) {

	exchangeNodeText := s.renderExchangeElementAsString(exchange)
	exchangeNode := node.AddChild(exchangeNodeText)

	// process all bindings for current exchange
	for _, binding := range getBindingsForExchange(exchange, brokerInfo.Bindings) {
		switch binding.DestinationType {
		case "queue":
			s.addQueue(exchangeNode, &binding, exchange, brokerInfo)
		case "exchange":
			s.addExchange(exchangeNode,
				findExchangeByName( // TODO can be nil
					brokerInfo.Exchanges,
					binding.Vhost,
					binding.Destination),
				brokerInfo)
		default:
			// unknown binding type
			exchangeNode.AddChild(fmt.Sprintf("%s (unknown binding type %s)",
				binding.Source, binding.DestinationType))
		}
	}
}

func (s BrokerInfoPrinter) getExchangesToDisplay(
	exchanges []rabtap.RabbitExchange,
	vhost string) []rabtap.RabbitExchange {

	var result []rabtap.RabbitExchange
	for _, exchange := range exchanges {
		if exchange.Vhost == vhost &&
			(exchange.Name != "" || s.config.ShowDefaultExchange) {
			result = append(result, exchange)
		}
	}
	return result
}

// Print renders given brokerInfo into a tree-view:
//  RabbitMQ-Host
//  +--VHost
//     +--Exchange
//        +--Queue bound to exchange
//           +--Consumer  (optional)
//              +--Connection
//
func (s BrokerInfoPrinter) Print(brokerInfo rabtap.BrokerInfo,
	rootNode string, out io.Writer) error {

	// root of node is URL of rabtap.RabbitMQ broker.
	// parse broker URL to filter out credentials for display.
	url, err := url.Parse(rootNode)
	if err != nil {
		return err
	}
	root := NewInfoTree(s.renderRootNodeAsString(*url, brokerInfo.Overview))

	// collect set of vhosts
	vhosts := make(map[string]bool)
	for _, exchange := range brokerInfo.Exchanges {
		vhosts[exchange.Vhost] = true
	}

	for vhost := range vhosts {
		node := root.AddChild(fmt.Sprintf("Vhost %s",
			s.colorizer.VHost(vhost)))

		for _, exchange := range s.getExchangesToDisplay(brokerInfo.Exchanges, vhost) {
			s.addExchange(node,
				&exchange,
				brokerInfo)
		}
	}

	PrintTree(root, out)
	return nil
}
