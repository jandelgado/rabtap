// Copyright (C) 2017-2019 Jan Delgado
// Render broker info into a textual tree representation

package main

import (
	"fmt"
	"io"
	"net/url"
	"strings"
	"text/template"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

var (
	_ = func() struct{} {
		RegisterBrokerInfoRenderer("text", NewBrokerInfoRendererText)
		return struct{}{}
	}()
)

// brokerInfoRendererText renders a tree representation represented by a rootNode
// into a string representation
type brokerInfoRendererText struct {
	config        BrokerInfoRendererConfig
	templateFuncs template.FuncMap
}

// NewBrokerInfoRendererText returns a BrokerInfoRenderer that renders for
// text console.
func NewBrokerInfoRendererText(config BrokerInfoRendererConfig) BrokerInfoRenderer {
	colorizer := NewColorPrinter(config.NoColor)
	// TODO inject
	templateFuncs := MergeTemplateFuncs(colorizer.GetFuncMap(), RabtapTemplateFuncs)
	return &brokerInfoRendererText{
		config:        config,
		templateFuncs: templateFuncs,
	}
}

const (
	//{{- with .URL }} {{ .Redacted | URLColor }} {{ end }}
	tplRootNode = `
	    {{- printf "%s://%s%s" .URL.Scheme .URL.Host .URL.Path | URLColor }}
	 	{{- if .Overview }} (broker ver='{{ .Overview.RabbitmqVersion }}',
		{{- "" }} mgmt ver='{{ .Overview.ManagementVersion }}',
		{{- "" }} cluster='{{ .Overview.ClusterName }}{{end}}')`
	tplVhost = `
	    {{- printf "Vhost %s" .Vhost.Name | VHostColor }}`
	tplConnection = ` 
	    {{- ""}}{{- if .NotFound }}{{ "? (connection)" | ErrorColor }}{{else}}
		{{-   ""}}'{{ ConnectionColor .Connection.Name }}'
		{{-   ""}}{{- if .Connection.ClientProperties.ConnectionName }} ({{- .Connection.ClientProperties.ConnectionName }}) {{end}}
		{{-   ""}} (connection {{ .Connection.User}}@{{ .Connection.Host }}:{{ .Connection.Port }},
		{{-   ""}} state='{{ .Connection.State }}',
		{{-   ""}} client='{{ .Connection.ClientProperties.Product }}',
		{{-   ""}} ver='{{ .Connection.ClientProperties.Version }}',
		{{-   ""}} peer='{{ .Connection.PeerHost }}:{{ .Connection.PeerPort }}')
		{{- end}}`
	tplChannel = ` 
	    {{- ""}}{{- if .NotFound }}{{ "? (channel)" | ErrorColor }}{{else}}
	    {{-   ""}}'{{ ChannelColor .Channel.Name }}' (channel 
		{{-   ""}} prefetch={{ .Channel.PrefetchCount }},
		{{-   ""}} state={{ .Channel.State }},
		{{-   ""}} unacked={{ .Channel.MessagesUnacknowledged }},
		{{-   ""}} confirms={{ .Channel.Confirm | YesNo }}
		{{-   if .Channel.IdleSince}}{{- ", idle since "}}{{ .Channel.IdleSince}}{{else}}
		{{-     if and .Config.ShowStats .Channel.MessageStats }} (
		{{-       ", "}}
		{{-       with .Channel.MessageStats.PublishDetails }}{{ if gt .Rate 0. }}{{printf "pub=%.1f" .Rate}}{{end}}{{end}}
		{{-       with .Channel.MessageStats.ConfirmDetails }}{{ if gt .Rate 0. }}{{printf ", confirms=%.1f " .Rate}}{{end}}{{end}}
		{{-       with .Channel.MessageStats.ReturnUnroutableDetails }}{{ if gt .Rate 0. }}{{printf ", drop=%.1f " .Rate}}{{end}}{{end}}
		{{-       with .Channel.MessageStats.DeliverGetDetails }}{{ if gt .Rate 0. }}{{printf "get=%.1f" .Rate}}{{end}}{{end}}
		{{-       with .Channel.MessageStats.AckDetails }}{{ if gt .Rate 0. }}{{printf ", ack=%.1f" .Rate}}{{end}}{{end}}) msg/s
		{{-     end }}
		{{-   end}})
		{{- end}}`
	tplConsumer = `
		{{- ConsumerColor .Consumer.ConsumerTag }} (consumer
		{{- ""}} prefetch={{ .Consumer.PrefetchCount }},
		{{- ""}} ack_req={{ .Consumer.AckRequired | YesNo }},
		{{- ""}} active={{ .Consumer.Active | YesNo }},
		{{- ""}} status={{ .Consumer.ActivityStatus }})`
	tplExchange = `
	    {{- if eq .Exchange.Name "" }}{{ ExchangeColor "(default)" }}{{ else }}{{ ExchangeColor .Exchange.Name }}{{ end }}
	    {{- "" }} (exchange({{ .Exchange.Type }}),
		{{- if .Binding }}
		{{-   with .Binding.RoutingKey }} key='{{ KeyColor .}}',{{end}}
		{{-   with .Binding.Arguments}} args='{{ KeyColor .}}',{{end}}
		{{- end }}
		{{- if and .Config.ShowStats .Exchange.MessageStats }} in=(
		{{-   .Exchange.MessageStats.PublishIn }}, {{printf "%.1f" .Exchange.MessageStats.PublishInDetails.Rate}}/s) msg, out=(
		{{-   .Exchange.MessageStats.PublishOut }}, {{printf "%.1f" .Exchange.MessageStats.PublishOutDetails.Rate}}/s) msg,
		{{- end }} [{{ .ExchangeFlags  }}])`
	tplBoundQueue = `
	    {{- QueueColor .Queue.Name }} (queue({{ .Queue.Type}}),
		{{- if .Binding }}
		{{-   with .Binding.RoutingKey }} key='{{ KeyColor .}}',{{end}}
		{{-   with .Binding.Arguments}} args='{{ KeyColor .}}',{{end}}
		{{- end }}
		{{- " " }}
		{{- if .Config.ShowStats }}
		{{-   .Queue.Consumers  }} cons, (
		{{-   .Queue.Messages }}, {{printf "%.1f" .Queue.MessagesDetails.Rate}}/s) msg, (
		{{-   .Queue.MessagesReady }}, {{printf "%.1f" .Queue.MessagesReadyDetails.Rate}}/s) msg ready,
		{{-   " " }}{{ ToPercent .Queue.ConsumerUtilisation }}% utl,
		{{- end }}
		{{- if .Queue.IdleSince}}{{- " idle since "}}{{ .Queue.IdleSince}}{{else}}{{ " running" }}{{end}}
		{{- ""}}, [{{ .QueueFlags}}])`
)

func (s brokerInfoRendererText) renderQueueFlagsAsString(queue *rabtap.RabbitQueue) string {
	flags := []bool{queue.Durable, queue.AutoDelete, queue.Exclusive}
	names := []string{"D", "AD", "EX"}
	return strings.Join(filterStringList(flags, names), "|")
}

func (s brokerInfoRendererText) renderExchangeFlagsAsString(exchange *rabtap.RabbitExchange) string {
	flags := []bool{exchange.Durable, exchange.AutoDelete, exchange.Internal}
	names := []string{"D", "AD", "I"}
	return strings.Join(filterStringList(flags, names), "|")
}

func (s brokerInfoRendererText) renderVhostAsString(vhost *rabtap.RabbitVhost) string {
	var args = struct {
		Vhost *rabtap.RabbitVhost
	}{vhost}
	return resolveTemplate("vhost-tpl", tplVhost, args, s.templateFuncs)
}

func (s brokerInfoRendererText) renderConsumerElementAsString(consumer *rabtap.RabbitConsumer) string {
	var args = struct {
		Config   BrokerInfoRendererConfig
		Consumer *rabtap.RabbitConsumer
	}{s.config, consumer}
	return resolveTemplate("consumer-tpl", tplConsumer, args, s.templateFuncs)
}

func (s brokerInfoRendererText) renderConnectionElementAsString(conn *rabtap.RabbitConnection, notFound bool) string {
	var args = struct {
		Config     BrokerInfoRendererConfig
		Connection *rabtap.RabbitConnection
		NotFound   bool
	}{s.config, conn, notFound}
	return resolveTemplate("connnection-tpl", tplConnection, args, s.templateFuncs)
}

func (s brokerInfoRendererText) renderChannelElementAsString(channel *rabtap.RabbitChannel, notFound bool) string {
	var args = struct {
		Config   BrokerInfoRendererConfig
		Channel  *rabtap.RabbitChannel
		NotFound bool
	}{s.config, channel, notFound}
	return resolveTemplate("channel-tpl", tplChannel, args, s.templateFuncs)
}

func (s brokerInfoRendererText) renderBoundQueueElementAsString(queue *rabtap.RabbitQueue, binding *rabtap.RabbitBinding) string {
	queueFlags := s.renderQueueFlagsAsString(queue)
	var args = struct {
		Config     BrokerInfoRendererConfig
		Binding    *rabtap.RabbitBinding
		Queue      *rabtap.RabbitQueue
		QueueFlags string
	}{s.config, binding, queue, queueFlags}
	return resolveTemplate("bound-queue-tpl", tplBoundQueue, args, s.templateFuncs)
}

func (s brokerInfoRendererText) renderRootNodeAsString(rabbitURL *url.URL, overview *rabtap.RabbitOverview) string {
	var args = struct {
		Config   BrokerInfoRendererConfig
		URL      *url.URL
		Overview *rabtap.RabbitOverview
	}{s.config, rabbitURL, overview}
	return resolveTemplate("rootnode", tplRootNode, args, s.templateFuncs)
}

func (s brokerInfoRendererText) renderExchangeElementAsString(exchange *rabtap.RabbitExchange, binding *rabtap.RabbitBinding) string {
	exchangeFlags := s.renderExchangeFlagsAsString(exchange)
	var args = struct {
		Config        BrokerInfoRendererConfig
		Exchange      *rabtap.RabbitExchange
		ExchangeFlags string
		Binding       *rabtap.RabbitBinding
	}{s.config, exchange, exchangeFlags, binding}
	return resolveTemplate("exchange-tpl", tplExchange, args, s.templateFuncs)
}

func (s brokerInfoRendererText) renderNode(n interface{}) *TreeNode {
	var node *TreeNode

	switch e := n.(type) {
	case *rootNode:
		node = NewTreeNode(s.renderRootNodeAsString(e.URL, e.Overview))
	case *vhostNode:
		node = NewTreeNode(s.renderVhostAsString(e.Vhost))
	case *connectionNode:
		node = NewTreeNode(s.renderConnectionElementAsString(e.OptConnection, e.NotFound))
	case *channelNode:
		node = NewTreeNode(s.renderChannelElementAsString(e.OptChannel, e.NotFound))
	case *consumerNode:
		node = NewTreeNode(s.renderConsumerElementAsString(e.Consumer))
	case *queueNode:
		node = NewTreeNode(s.renderBoundQueueElementAsString(e.Queue, e.OptBinding))
	case *exchangeNode:
		node = NewTreeNode(s.renderExchangeElementAsString(e.Exchange, e.OptBinding))
	default:
		panic(fmt.Sprintf("unexpected node encountered %T", e))
	}

	for _, child := range n.(Node).Children() {
		node.Add(s.renderNode(child))
	}
	return node
}

func (s brokerInfoRendererText) Render(rootNode *rootNode, out io.Writer) error {
	root := s.renderNode(rootNode)
	PrintTree(root, out)
	return nil
}
