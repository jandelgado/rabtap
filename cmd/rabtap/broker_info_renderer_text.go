// Copyright (C) 2017-2019 Jan Delgado
// Render broker info into a textual tree representation

package main

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// BrokerInfoRenderer renders a tree representation represented by a rootNode
// into a string representation
type BrokerInfoRendererText struct {
	config    BrokerInfoRendererConfig
	colorizer ColorPrinter
}

func NewBrokerInfoRendererText(config BrokerInfoRendererConfig) BrokerInfoRenderer {
	return &BrokerInfoRendererText{
		config:    config,
		colorizer: NewColorPrinter(config.NoColor),
	}
}

const (
	tplRootNode = `
	    {{- printf "%s://%s%s" .URL.Scheme .URL.Host .URL.Path | URLColor }} 
	 	{{- if .Overview }} (broker ver='{{ .Overview.RabbitmqVersion }}',
		{{- "" }} mgmt ver='{{ .Overview.ManagementVersion }}',
		{{- "" }} cluster='{{ .Overview.ClusterName }}{{end}}')`
	tplVhost = `
	    {{- printf "Vhost %s" .Vhost | VHostColor }}`
	tplConsumer = `
		{{- ConsumerColor .Consumer.ConsumerTag }} (consumer 
		{{- ""}} user='{{ .Consumer.ChannelDetails.User }}', 
		{{- ""}} prefetch={{ .Consumer.PrefetchCount }}, chan='
		{{- .Consumer.ChannelDetails.Name }}')`
	tplConnection = ` 
	    {{- ""}}'{{ ConnectionColor .Connection.Name }}' (connection 
		{{- ""}} client='{{ .Connection.ClientProperties.Product}}',
		{{- ""}} host='{{ .Connection.Host }}:{{ .Connection.Port }}',
		{{- ""}} peer='{{ .Connection.PeerHost }}:{{ .Connection.PeerPort }}')`
	tplExchange = `
	    {{- if eq .Exchange.Name "" }}{{ ExchangeColor "(default)" }}{{ else }}{{ ExchangeColor .Exchange.Name }}{{ end }}
	    {{- "" }} (exchange, type '{{ .Exchange.Type  }}'
		{{- if and .Config.ShowStats .Exchange.MessageStats }}, in=(
		{{- .Exchange.MessageStats.PublishIn }}, {{printf "%.1f" .Exchange.MessageStats.PublishInDetails.Rate}}/s) msg, out=(
		{{- .Exchange.MessageStats.PublishOut }}, {{printf "%.1f" .Exchange.MessageStats.PublishOutDetails.Rate}}/s) msg
		{{- end }}, [{{ .ExchangeFlags  }}])`
	tplQueue = `
	    {{- QueueColor .Queue.Name }} (queue,
		{{- if .Config.ShowStats }}
		{{- .Queue.Consumers  }} cons, (
		{{- .Queue.Messages }}, {{printf "%.1f" .Queue.MessagesDetails.Rate}}/s) msg, (
		{{- .Queue.MessagesReady }}, {{printf "%.1f" .Queue.MessagesReadyDetails.Rate}}/s) msg ready,
		{{- end }}
		{{- if .Queue.IdleSince}}{{- " idle since "}}{{ .Queue.IdleSince}}{{else}}{{ " running" }}{{end}}
		{{- ""}}, [{{ .QueueFlags}}])`
	tplBoundQueue = `
	    {{- QueueColor .Binding.Destination }} (queue,
		{{- with .Binding.RoutingKey }} key='{{ KeyColor .}}',{{end}}
		{{- with .Binding.Arguments}} args='{{ KeyColor .}}',{{end}}
		{{- if .Config.ShowStats }}
		{{- .Queue.Consumers  }} cons, (
		{{- .Queue.Messages }}, {{printf "%.1f" .Queue.MessagesDetails.Rate}}/s) msg, (
		{{- .Queue.MessagesReady }}, {{printf "%.1f" .Queue.MessagesReadyDetails.Rate}}/s) msg ready,
		{{- end }}
		{{- if .Queue.IdleSince}}{{- " idle since "}}{{ .Queue.IdleSince}}{{else}}{{ " running" }}{{end}}
		{{- ""}}, [{{ .QueueFlags}}])`
)

func (s BrokerInfoRendererText) renderQueueFlagsAsString(queue rabtap.RabbitQueue) string {
	flags := []bool{queue.Durable, queue.AutoDelete, queue.Exclusive}
	names := []string{"D", "AD", "EX"}
	return strings.Join(filterStringList(flags, names), "|")
}

func (s BrokerInfoRendererText) renderExchangeFlagsAsString(exchange rabtap.RabbitExchange) string {
	flags := []bool{exchange.Durable, exchange.AutoDelete, exchange.Internal}
	names := []string{"D", "AD", "I"}
	return strings.Join(filterStringList(flags, names), "|")
}

func (s BrokerInfoRendererText) renderVhostAsString(vhost string) string {
	var args = struct {
		Vhost string
	}{vhost}
	return resolveTemplate("vhost-tpl", tplVhost, args, s.colorizer.GetFuncMap())
}

func (s BrokerInfoRendererText) renderConsumerElementAsString(consumer rabtap.RabbitConsumer) string {
	var args = struct {
		Config   BrokerInfoRendererConfig
		Consumer rabtap.RabbitConsumer
	}{s.config, consumer}
	return resolveTemplate("consumer-tpl", tplConsumer, args, s.colorizer.GetFuncMap())
}

func (s BrokerInfoRendererText) renderConnectionElementAsString(conn rabtap.RabbitConnection) string {
	var args = struct {
		Config     BrokerInfoRendererConfig
		Connection rabtap.RabbitConnection
	}{s.config, conn}
	return resolveTemplate("connnection-tpl", tplConnection, args, s.colorizer.GetFuncMap())
}

func (s BrokerInfoRendererText) renderQueueElementAsString(queue rabtap.RabbitQueue) string {
	queueFlags := s.renderQueueFlagsAsString(queue)
	var args = struct {
		Config     BrokerInfoRendererConfig
		Queue      rabtap.RabbitQueue
		QueueFlags string
	}{s.config, queue, queueFlags}
	return resolveTemplate("queue-tpl", tplQueue, args, s.colorizer.GetFuncMap())
}

func (s BrokerInfoRendererText) renderBoundQueueElementAsString(queue rabtap.RabbitQueue, binding rabtap.RabbitBinding) string {
	queueFlags := s.renderQueueFlagsAsString(queue)
	var args = struct {
		Config     BrokerInfoRendererConfig
		Binding    rabtap.RabbitBinding
		Queue      rabtap.RabbitQueue
		QueueFlags string
	}{s.config, binding, queue, queueFlags}
	return resolveTemplate("bound-queue-tpl", tplBoundQueue, args, s.colorizer.GetFuncMap())
}

func (s BrokerInfoRendererText) renderRootNodeAsString(rabbitURL url.URL, overview rabtap.RabbitOverview) string {
	var args = struct {
		Config   BrokerInfoRendererConfig
		URL      url.URL
		Overview rabtap.RabbitOverview
	}{s.config, rabbitURL, overview}
	return resolveTemplate("rootnode", tplRootNode, args, s.colorizer.GetFuncMap())
}

func (s BrokerInfoRendererText) renderExchangeElementAsString(exchange rabtap.RabbitExchange) string {
	exchangeFlags := s.renderExchangeFlagsAsString(exchange)
	var args = struct {
		Config        BrokerInfoRendererConfig
		Exchange      rabtap.RabbitExchange
		ExchangeFlags string
	}{s.config, exchange, exchangeFlags}
	return resolveTemplate("exchange-tpl", tplExchange, args, s.colorizer.GetFuncMap())
}

func (s BrokerInfoRendererText) renderNodeText(n interface{}) *TreeNode {
	var node *TreeNode

	switch t := n.(type) {
	case *rootNode:
		node = NewTreeNode(s.renderRootNodeAsString(n.(*rootNode).URL, n.(*rootNode).Overview))
	case *vhostNode:
		node = NewTreeNode(s.renderVhostAsString(n.(*vhostNode).Vhost))
	case *connectionNode:
		node = NewTreeNode(s.renderConnectionElementAsString(n.(*connectionNode).Connection))
	case *consumerNode:
		node = NewTreeNode(s.renderConsumerElementAsString(n.(*consumerNode).Consumer))
	case *queueNode:
		node = NewTreeNode(s.renderQueueElementAsString(n.(*queueNode).Queue))
	case *boundQueueNode:
		node = NewTreeNode(s.renderBoundQueueElementAsString(n.(*boundQueueNode).Queue, n.(*boundQueueNode).Binding))
	case *exchangeNode:
		node = NewTreeNode(s.renderExchangeElementAsString(n.(*exchangeNode).Exchange))
	default:
		panic(fmt.Sprintf("unexpected node encountered %T", t))
	}

	for _, child := range n.(Node).Children() {
		node.Add(s.renderNodeText(child))
	}
	return node
}

func (s BrokerInfoRendererText) Render(rootNode *rootNode, out io.Writer) error {
	root := s.renderNodeText(rootNode)
	PrintTree(root, out)
	return nil
}
