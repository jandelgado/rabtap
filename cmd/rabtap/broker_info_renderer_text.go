// Copyright (C) 2017-2019 Jan Delgado
// Render broker info into a textual tree representation

package main

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	"text/template"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// BrokerInfoRenderer renders a tree representation represented by a RootNode
// into a string representation
type BrokerInfoRendererText struct {
	config    BrokerInfoPrinterConfig
	colorizer ColorPrinter
}

func NewBrokerInfoRendererText(config BrokerInfoPrinterConfig) BrokerInfoRenderer {
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
		{{- end }}, {{ .ExchangeFlags  }})`
	tplQueue = `
	    {{- QueueColor .Queue.Name }} (queue,
		{{- if .Config.ShowStats }}
		{{- .Queue.Consumers  }} cons, (
		{{- .Queue.Messages }}, {{printf "%.1f" .Queue.MessagesDetails.Rate}}/s) msg, (
		{{- .Queue.MessagesReady }}, {{printf "%.1f" .Queue.MessagesReadyDetails.Rate}}/s) msg ready,
		{{- end }}
		{{- if .Queue.IdleSince}}{{- " idle since "}}{{ .Queue.IdleSince}}{{else}}{{ " running" }}{{end}}
		{{- ""}}, {{ .QueueFlags}})`
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
		{{- ""}}, {{ .QueueFlags}})`
)

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
func (s BrokerInfoRendererText) resolveTemplate(name string, tpl string, args interface{}) string {
	tmpl := template.Must(template.New(name).Funcs(
		s.colorizer.GetFuncMap()).Parse(tpl))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, args)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func (s BrokerInfoRendererText) renderQueueFlagsAsString(queue rabtap.RabbitQueue) string {
	flags := []bool{queue.Durable, queue.AutoDelete, queue.Exclusive}
	names := []string{"D", "AD", "EX"}
	return "[" + strings.Join(filterStringList(flags, names), "|") + "]"
}

func (s BrokerInfoRendererText) renderExchangeFlagsAsString(exchange rabtap.RabbitExchange) string {
	flags := []bool{exchange.Durable, exchange.AutoDelete, exchange.Internal}
	names := []string{"D", "AD", "I"}
	return "[" + strings.Join(filterStringList(flags, names), "|") + "]"
}

func (s BrokerInfoRendererText) renderVhostAsString(vhost string) string {
	var args = struct {
		Vhost string
	}{vhost}
	return s.resolveTemplate("vhost-tpl", tplVhost, args)
}

func (s BrokerInfoRendererText) renderConsumerElementAsString(consumer rabtap.RabbitConsumer) string {
	var args = struct {
		Config   BrokerInfoPrinterConfig
		Consumer rabtap.RabbitConsumer
	}{s.config, consumer}
	return s.resolveTemplate("consumer-tpl", tplConsumer, args)
}

func (s BrokerInfoRendererText) renderConnectionElementAsString(conn rabtap.RabbitConnection) string {
	var args = struct {
		Config     BrokerInfoPrinterConfig
		Connection rabtap.RabbitConnection
	}{s.config, conn}
	return s.resolveTemplate("connnection-tpl", tplConnection, args)
}

func (s BrokerInfoRendererText) renderQueueElementAsString(queue rabtap.RabbitQueue) string {
	queueFlags := s.renderQueueFlagsAsString(queue)
	var args = struct {
		Config     BrokerInfoPrinterConfig
		Queue      rabtap.RabbitQueue
		QueueFlags string
	}{s.config, queue, queueFlags}
	return s.resolveTemplate("queue-tpl", tplQueue, args)
}

func (s BrokerInfoRendererText) renderBoundQueueElementAsString(queue rabtap.RabbitQueue, binding rabtap.RabbitBinding) string {
	queueFlags := s.renderQueueFlagsAsString(queue)
	var args = struct {
		Config     BrokerInfoPrinterConfig
		Binding    rabtap.RabbitBinding
		Queue      rabtap.RabbitQueue
		QueueFlags string
	}{s.config, binding, queue, queueFlags}
	return s.resolveTemplate("bound-queue-tpl", tplBoundQueue, args)
}

func (s BrokerInfoRendererText) renderRootNodeAsString(rabbitURL url.URL, overview rabtap.RabbitOverview) string {
	var args = struct {
		Config   BrokerInfoPrinterConfig
		URL      url.URL
		Overview rabtap.RabbitOverview
	}{s.config, rabbitURL, overview}
	return s.resolveTemplate("rootnode", tplRootNode, args)
}

func (s BrokerInfoRendererText) renderExchangeElementAsString(exchange rabtap.RabbitExchange) string {
	exchangeFlags := s.renderExchangeFlagsAsString(exchange)
	var args = struct {
		Config        BrokerInfoPrinterConfig
		Exchange      rabtap.RabbitExchange
		ExchangeFlags string
	}{s.config, exchange, exchangeFlags}
	return s.resolveTemplate("exchange-tpl", tplExchange, args)
}

func (s BrokerInfoRendererText) renderNodeText(n interface{}) *TreeNode {
	var node *TreeNode

	switch t := n.(type) {
	case *RootNode:
		node = NewTreeNode(s.renderRootNodeAsString(n.(*RootNode).URL, n.(*RootNode).Overview))
	case *VhostNode:
		node = NewTreeNode(s.renderVhostAsString(n.(*VhostNode).Vhost))
	case *ConnectionNode:
		node = NewTreeNode(s.renderConnectionElementAsString(n.(*ConnectionNode).Connection))
	case *ConsumerNode:
		node = NewTreeNode(s.renderConsumerElementAsString(n.(*ConsumerNode).Consumer))
	case *QueueNode:
		node = NewTreeNode(s.renderQueueElementAsString(n.(*QueueNode).Queue))
	case *BoundQueueNode:
		node = NewTreeNode(s.renderBoundQueueElementAsString(n.(*BoundQueueNode).Queue, n.(*BoundQueueNode).Binding))
	case *ExchangeNode:
		node = NewTreeNode(s.renderExchangeElementAsString(n.(*ExchangeNode).Exchange))
	default:
		panic(fmt.Sprintf("unexpected node encountered %T", t))
	}

	for _, child := range n.(Node).Children() {
		node.Add(s.renderNodeText(child))
	}
	return node
}

func (s BrokerInfoRendererText) Render(rootNode *RootNode, out io.Writer) error {
	root := s.renderNodeText(rootNode)
	PrintTree(root, out)
	return nil
}
