// Copyright (C) 2017 Jan Delgado

package main

// TODO factor out templates, make configurable.

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	"text/template"

	"github.com/jandelgado/rabtap/pkg"
)

// BrokerInfoPrinter prints nicely treeish infos desribing a brokers
// topology
type BrokerInfoPrinter struct {
	//brokerInfo BrokerInfo
	config    PrintBrokerInfoConfig
	colorizer ColorPrinter
}

// NewBrokerInfoPrinter constructs a new object to print a broker info
func NewBrokerInfoPrinter(config PrintBrokerInfoConfig) *BrokerInfoPrinter {
	s := BrokerInfoPrinter{
		config:    config,
		colorizer: NewColorPrinter(config.NoColor),
	}
	return &s
}

// PrintBrokerInfoConfig controls bevaviour auf PrintBrokerInfo
type PrintBrokerInfoConfig struct {
	ShowDefaultExchange bool
	ShowConsumers       bool
	ShowStats           bool
	NoColor             bool
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

func (s BrokerInfoPrinter) renderQueueFlagsAsString(queue rabtap.RabbitQueue) string {
	var flags []string
	if queue.Durable {
		flags = append(flags, "D")
	}
	if queue.AutoDelete {
		flags = append(flags, "AD")
	}
	if queue.Exclusive {
		flags = append(flags, "EX")
	}
	return "[" + strings.Join(flags, "|") + "]"
}

func (s BrokerInfoPrinter) renderExchangeFlagsAsString(exchange rabtap.RabbitExchange) string {
	var flags []string
	if exchange.Durable {
		flags = append(flags, "D")
	}
	if exchange.AutoDelete {
		flags = append(flags, "AD")
	}
	if exchange.Internal {
		flags = append(flags, "I")
	}
	return "[" + strings.Join(flags, "|") + "]"
}

func (s BrokerInfoPrinter) renderConsumerElementAsString(consumer rabtap.RabbitConsumer) string {
	return fmt.Sprintf("%s (consumer, %s)",
		s.colorizer.Consumer(consumer.ConsumerTag),
		consumer.ChannelDetails.Name)
}

func (s BrokerInfoPrinter) renderQueueElementAsString(queue rabtap.RabbitQueue, binding rabtap.RabbitBinding) string {

	queueFlags := s.renderQueueFlagsAsString(queue)
	type QueueInfo struct {
		Config     PrintBrokerInfoConfig
		Binding    rabtap.RabbitBinding
		Queue      rabtap.RabbitQueue
		QueueFlags string
	}
	args := QueueInfo{s.config, binding, queue, queueFlags}

	const tpl = `
	    {{- QueueColor .Binding.Destination }} (queue,
		{{- with .Binding.RoutingKey }} key={{ KeyColor .}},{{end}}
		{{- with .Binding.Arguments}} args={{ KeyColor .}},{{end}}
		{{- if .Config.ShowStats }}
		{{- .Queue.Consumers  }} cons, (
		{{- .Queue.Messages }}, {{printf "%.1f" .Queue.MessagesDetails.Rate}}/s) msg, (
		{{- .Queue.MessagesReady }}, {{printf "%.1f" .Queue.MessagesReadyDetails.Rate}}/s) msg ready,
		{{- end }}
		{{- if .Queue.IdleSince}}{{- " idle since "}}{{ .Queue.IdleSince}}{{else}}{{ " running" }}{{end}}
		{{- " "}}{{ .QueueFlags}})`

	return s.resolveTemplate("queue-tpl", tpl, args)
}

func (s BrokerInfoPrinter) renderRootNodeAsString(rabbitURL url.URL, overview rabtap.RabbitOverview) string {

	type RootInfo struct {
		Config   PrintBrokerInfoConfig
		URL      url.URL
		Overview rabtap.RabbitOverview
	}
	args := RootInfo{s.config, rabbitURL, overview}

	const tpl = `{{ printf "%s://%s%s" .URL.Scheme .URL.Host .URL.Path | URLColor }} 
		   {{- if .Overview }} (broker ver={{ .Overview.RabbitmqVersion }},
		   {{- "" }} mgmt ver={{ .Overview.ManagementVersion }},
		   {{- "" }} cluster={{ .Overview.ClusterName }}{{end}})`

	return s.resolveTemplate("rootnode", tpl, args)
}

func (s BrokerInfoPrinter) renderExchangeElementAsString(exchange rabtap.RabbitExchange) string {
	printName := exchange.Name
	if printName == "" {
		printName = "(default)"
	}

	type ExchangeInfo struct {
		Config        PrintBrokerInfoConfig
		Exchange      rabtap.RabbitExchange
		ExchangeFlags string
		PrintName     string
	}
	exchangeFlags := s.renderExchangeFlagsAsString(exchange)
	args := ExchangeInfo{s.config, exchange, exchangeFlags, printName}

	const tpl = `{{ ExchangeColor .PrintName }} (exchange, type '
	                {{- .Exchange.Type  }}' {{ .ExchangeFlags  }})`

	return s.resolveTemplate("exchange-tpl", tpl, args)
}

func (s BrokerInfoPrinter) addConsumerNode(node *TreeNode, consumers []rabtap.RabbitConsumer,
	exchange *rabtap.RabbitExchange, queue *rabtap.RabbitQueue) {

	for _, consumer := range consumers {
		if consumer.Queue.Vhost == exchange.Vhost &&
			consumer.Queue.Name == queue.Name {
			node.AddChild(s.renderConsumerElementAsString(consumer))
		}
	}
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

// addExchange recursively (in case of exchange-exchange binding) an exchange to the
// given node.
// TODO simplify
func (s BrokerInfoPrinter) addExchange(node *TreeNode,
	exchange *rabtap.RabbitExchange, brokerInfo BrokerInfo) {

	exchangeNodeText := s.renderExchangeElementAsString(*exchange)
	exchangeNode := node.AddChild(exchangeNodeText)

	// process all bindings for current exchange
	for _, binding := range getBindingsForExchange(exchange, brokerInfo.Bindings) {
		switch binding.DestinationType {
		case "queue":
			// standard binding of queue to exchange
			queue := findQueueByName(brokerInfo.Queues,
				binding.Vhost,
				binding.Destination)
			if queue == nil {
				// we test for nil because (at least in theory) a queue
				// can disappear since we are making various non-transactional
				// API calls
				queue = &rabtap.RabbitQueue{Name: binding.Destination}
			}
			queueText := s.renderQueueElementAsString(*queue, binding)
			queueNode := exchangeNode.AddChild(queueText)

			// add consumers of queue
			if s.config.ShowConsumers {
				s.addConsumerNode(queueNode, brokerInfo.Consumers, exchange, queue)
			}
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

// Print render given brokerInfo into a tree-view:
//  RabbitMQ-Host
//  +--VHost
//     +--Exchange
//        +--Queue bound to exchange
//           +--Consumer
func (s BrokerInfoPrinter) Print(brokerInfo BrokerInfo,
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
