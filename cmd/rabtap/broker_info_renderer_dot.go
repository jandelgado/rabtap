// Copyright (C) 2017-2019 Jan Delgado
// Render broker info into a dot graph representation
// https://www.graphviz.org/doc/info/lang.html

package main

import (
	"fmt"
	"html"
	"io"
	"net/url"
	"strconv"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

var (
	_ = func() struct{} {
		RegisterBrokerInfoRenderer("dot", NewBrokerInfoRendererDot)
		return struct{}{}
	}()
)

// dotRendererTpl holds template fragments to use while rendering
type dotRendererTpl struct {
	dotTplRootNode     string
	dotTplVhost        string
	dotTplExchange     string
	dotTplBoundQueue   string
	dotTplQueueBinding string
	dotTplConnection   string
	dotTplChannel      string
	dotTplConsumer     string
}

// brokerInfoRendererDot renders into graphviz dot format
type brokerInfoRendererDot struct {
	config   BrokerInfoRendererConfig
	template dotRendererTpl
}

type dotNode struct {
	Name        string
	Text        string
	ParentAssoc string
}

var emptyDotNode = dotNode{}

// NewBrokerInfoRendererDot returns a BrokerInfoRenderer implementation that
// renders into graphviz dot format
func NewBrokerInfoRendererDot(config BrokerInfoRendererConfig) BrokerInfoRenderer {
	return &brokerInfoRendererDot{
		config: config, template: newDotRendererTpl()}
}

// newDotRendererTpl returns the dot template to use. For now, just one default
// template is used, later will support loading templates from the filesytem
func newDotRendererTpl() dotRendererTpl {
	return dotRendererTpl{dotTplRootNode: `graph broker {
{{ q .Name }} [shape="record", label="{RabbitMQ {{ esc .Overview.RabbitmqVersion }} | 
               {{- printf "%s://%s%s" .URL.Scheme .URL.Host .URL.Path }} |
               {{- .Overview.ClusterName }} }"];

{{ range $i, $e := .Children }}{{ q $.Name }} -- {{ q $e.Name }}{{ printf ";\n" }}{{ end -}}
{{ range $i, $e := .Children }}{{ $e.Text -}}{{ end -}}
}`,

		dotTplVhost: `{{ q .Name }} [shape="box", label="Virtual host {{ esc .Vhost.Name }}"];

{{ range $i, $e := .Children }}{{ q $.Name }} -- {{ q $e.Name -}} [headport=n]{{ printf ";\n" }}{{ end -}}
{{ range $i, $e := .Children }}{{ $e.Text -}}{{ end -}}`,

		dotTplExchange: `
{{ q .Name }} [shape="record"; label="{ {{ esc .Exchange.Name }} | {{- esc .Exchange.Type }} | {
			  {{- if .Exchange.Durable }} D {{ end }} | 
			  {{- if .Exchange.AutoDelete }} AD {{ end }} | 
			  {{- if .Exchange.Internal }} I {{ end }} } }"];

{{ range $i, $e := .Children }}{{ q $.Name }} -- {{ q $e.Name }} [fontsize=10; headport=n; label={{ $e.ParentAssoc | esc | q}}]{{ printf ";\n" }}{{ end -}}
{{ range $i, $e := .Children }}{{ $e.Text }}{{ end -}}`,

		dotTplBoundQueue: `
{{ q .Name }} [shape="record"; label="{ {{ esc .Queue.Name }} | {
			  {{- if .Queue.Durable }} D {{ end }} | 
			  {{- if .Queue.AutoDelete }} AD {{ end }} | 
			  {{- if .Queue.Exclusive }} EX {{ end }} } }"];

{{ range $i, $e := .Children }}{{ q $.Name }} -- {{ q $e.Name }}{{ end -}}
{{ range $i, $e := .Children }}{{ $e.Text -}}{{ end -}}`,

		dotTplQueueBinding: `
{{ range $i, $e := .Children }}{{ q $.Name }} -- {{ q $e.Name }}{{ end -}}
{{ range $i, $e := .Children }}{{ $e.Text -}}{{ end -}}`,

		dotTplConnection: `
{{ q .Name }} [shape="record" label="{{ esc .Connection.Name }}"];

{{ range $i, $e := .Children }}{{ q $.Name }} -- {{ q $e.Name }}{{ end -}}
{{ range $i, $e := .Children }}{{ $e.Text -}}{{ end -}}`,

		dotTplChannel: `
{{ q .Name }} [shape="record" label="{{ esc .Channel.Name }}"];

{{ range $i, $e := .Children }}{{ q $.Name }} -- {{ q $e.Name }}{{ end -}}
{{ range $i, $e := .Children }}{{ $e.Text -}}{{ end -}}`,

		dotTplConsumer: `
{{ q .Name }} [shape="record" label="{{ esc .Consumer.ConsumerTag}}"];

{{ range $i, $e := .Children }}{{ q $.Name }} -- {{ q $e.Name }}{{ end -}}
{{ range $i, $e := .Children }}{{ $e.Text -}}{{ end -}}`,
	}
}

func (s brokerInfoRendererDot) funcMap() map[string]interface{} {
	return map[string]interface{}{
		"q":   strconv.Quote,
		"esc": html.EscapeString}
}

func (s brokerInfoRendererDot) renderRootNodeAsString(name string,
	children []dotNode,
	rabbitURL *url.URL,
	overview *rabtap.RabbitOverview) string {
	var args = struct {
		Name     string
		Children []dotNode
		Config   BrokerInfoRendererConfig
		URL      *url.URL
		Overview *rabtap.RabbitOverview
	}{name, children, s.config, rabbitURL, overview}
	return resolveTemplate("root-dotTpl", s.template.dotTplRootNode, args, s.funcMap())
}

func (s brokerInfoRendererDot) renderVhostAsString(name string,
	children []dotNode,
	vhost *rabtap.RabbitVhost) string {
	var args = struct {
		Name     string
		Children []dotNode
		Vhost    *rabtap.RabbitVhost
	}{name, children, vhost}
	return resolveTemplate("vhost-dotTpl", s.template.dotTplVhost, args, s.funcMap())
}

func (s brokerInfoRendererDot) renderExchangeElementAsString(name string,
	children []dotNode,
	exchange *rabtap.RabbitExchange) string {
	var args = struct {
		Name     string
		Children []dotNode
		Config   BrokerInfoRendererConfig
		Exchange *rabtap.RabbitExchange
	}{name, children, s.config, exchange}
	return resolveTemplate("exchange-dotTpl", s.template.dotTplExchange, args, s.funcMap())
}

func (s brokerInfoRendererDot) renderBoundQueueElementAsString(name string,
	children []dotNode,
	queue *rabtap.RabbitQueue,
	binding *rabtap.RabbitBinding) string {
	var args = struct {
		Name     string
		Children []dotNode
		Config   BrokerInfoRendererConfig
		Binding  *rabtap.RabbitBinding
		Queue    *rabtap.RabbitQueue
	}{name, children, s.config, binding, queue}
	return resolveTemplate("bound-queue-dotTpl", s.template.dotTplBoundQueue, args, s.funcMap())
}

func (s brokerInfoRendererDot) renderConsumerElementAsString(name string,
	children []dotNode,
	consumer *rabtap.RabbitConsumer) string {
	var args = struct {
		Name     string
		Children []dotNode
		Config   BrokerInfoRendererConfig
		Consumer *rabtap.RabbitConsumer
	}{name, children, s.config, consumer}
	return resolveTemplate("consumer-dotTpl", s.template.dotTplConsumer, args, s.funcMap())
}

func (s brokerInfoRendererDot) renderChannelElementAsString(name string,
	children []dotNode,
	channel *rabtap.RabbitChannel) string {
	var args = struct {
		Name     string
		Children []dotNode
		Config   BrokerInfoRendererConfig
		Channel  *rabtap.RabbitChannel
	}{name, children, s.config, channel}
	return resolveTemplate("channel-dotTpl", s.template.dotTplChannel, args, s.funcMap())
}

func (s brokerInfoRendererDot) renderConnectionElementAsString(name string,
	children []dotNode,
	conn *rabtap.RabbitConnection) string {
	var args = struct {
		Name       string
		Children   []dotNode
		Config     BrokerInfoRendererConfig
		Connection *rabtap.RabbitConnection
	}{name, children, s.config, conn}
	return resolveTemplate("connnection-dotTpl", s.template.dotTplConnection, args, s.funcMap())
}

func (s *brokerInfoRendererDot) renderNode(n interface{}, queueRendered map[string]bool) dotNode {
	var node dotNode
	// render queues only once (otherwise in exchange-to-exchange binding
	// scenarios, queues would be rendered multiple times)
	children := []dotNode{}
	for _, child := range n.(Node).Children() {
		c := s.renderNode(child.(Node), queueRendered)
		if c != emptyDotNode {
			children = append(children, c)
		}
	}

	switch t := n.(type) {
	case *rootNode:
		name := "root"
		node = dotNode{name, s.renderRootNodeAsString(name, children, t.URL, t.Overview), ""}
	case *vhostNode:
		name := fmt.Sprintf("vhost_%s", t.Vhost.Name)
		node = dotNode{name, s.renderVhostAsString(name, children, t.Vhost), ""}
	case *exchangeNode:
		name := fmt.Sprintf("exchange_%s_%s", t.Exchange.Vhost, t.Exchange.Name)
		binding := t.OptBinding
		key := ""
		if binding != nil {
			key = binding.RoutingKey
		}
		node = dotNode{name, s.renderExchangeElementAsString(name, children, t.Exchange), key}
	case *queueNode:
		queue := t.Queue
		name := fmt.Sprintf("queue_%s_%s", queue.Vhost, queue.Name)
		queueRendered[name] = true
		binding := t.OptBinding
		key := ""
		if binding != nil {
			key = binding.RoutingKey
		}
		node = dotNode{name, s.renderBoundQueueElementAsString(name, children, queue, binding), key}
	case *connectionNode:
		name := fmt.Sprintf("connection_%s", t.Connection.Name)
		node = dotNode{name, s.renderConnectionElementAsString(name, children, t.Connection), ""}
	case *channelNode:
		name := fmt.Sprintf("channel_%s", t.Channel.Name)
		node = dotNode{name, s.renderChannelElementAsString(name, children, t.Channel), ""}
	case *consumerNode:
		name := fmt.Sprintf("consumer_%s", t.Consumer.ConsumerTag)
		node = dotNode{name, s.renderConsumerElementAsString(name, children, t.Consumer), ""}
	default:
		panic(fmt.Sprintf("unexpected node encountered %T", t))
	}
	return node
}

// Render renders the given tree in graphviz dot format. See
// https://www.graphviz.org/doc/info/lang.html
func (s *brokerInfoRendererDot) Render(rootNode *rootNode, out io.Writer) error {
	res := s.renderNode(rootNode, map[string]bool{})
	fmt.Fprintf(out, res.Text)
	return nil
}
