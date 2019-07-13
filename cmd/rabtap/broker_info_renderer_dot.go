// Copyright (C) 2017-2019 Jan Delgado
// Render broker info into a dot graph representation

package main

import (
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

var (
	_ = func() struct{} {
		RegisterBrokerInfoRenderer("dot", NewBrokerInfoRendererDot)
		return struct{}{}
	}()
)

// BrokerInfoRenderer renders a tree representation represented by a rootNode
// into a string representation
type BrokerInfoRendererDot struct {
	config BrokerInfoRendererConfig
	res    string
}

type dotNode struct {
	Name        string
	Text        string
	ParentAssoc string
}

func NewBrokerInfoRendererDot(config BrokerInfoRendererConfig) BrokerInfoRenderer {
	return &BrokerInfoRendererDot{config: config}
}

const (
	dotTplRootNode = `
		  {{- "" }}graph broker {
		  {{ q .Name }} [label="{{- printf "%s://%s%s" .URL.Scheme .URL.Host .URL.Path }}"];
		  {{ range $i, $e := .Children }}{{ q $.Name }} -- {{ q $e.Name }}{{ printf ";\n" }}{{end}}
		  {{ range $i, $e := .Children }}{{ $e.Text }}{{end}}
		}`

	dotTplVhost = `
		  {{- q .Name }} [label="Virtual host {{ .Vhost }}"];
	      {{ range $i, $e := .Children }}{{ q $.Name }} -- {{ q $e.Name }}{{ printf ";\n" }}{{end}}
		  {{ range $i, $e := .Children }}{{ $e.Text }}{{end}}`

	dotTplExchange = `
		  subgraph cluster_0 {
		    label="exchanges";
		    style=filled;
		    color=lightgrey;
		    node [style=filled,color=white];
		    {{ q .Name }} [label="{{ .Exchange.Name }}"];
		  }
		  {{ range $i, $e := .Children }}{{ q $.Name }} -- {{ q $e.Name }} [label={{ q $e.ParentAssoc }}]{{ printf ";\n" }}{{end}}
		  {{ range $i, $e := .Children }}{{ $e.Text }}{{end}}
		  `

	dotTplQueue = `
		  {{- q .Name }} [label="{{ .Queue.Name }}"];
		  {{ range $i, $e := .Children }}{{ q $.Name }} -- {{ q $e.Name }}{{ printf ";\n" }}{{end}}
		  {{ range $i, $e := .Children }}{{ $e.Text }}{{end}}`

	dotTplBoundQueue = `
		  {{- q .Name }} [label="{{ .Queue.Name }}"];
		  {{ range $i, $e := .Children }}{{ q $.Name }} -- {{ q $e.Name }}{{ printf ";\n" }}{{end}}
		  {{ range $i, $e := .Children }}{{ $e.Text }}{{end}}`

	dotTplConsumer   = `consume`
	dotTplConnection = `connection`
)

func (s BrokerInfoRendererDot) renderQueueFlagsAsString(queue rabtap.RabbitQueue) string {
	flags := []bool{queue.Durable, queue.AutoDelete, queue.Exclusive}
	names := []string{"D", "AD", "EX"}
	return "[" + strings.Join(filterStringList(flags, names), "|") + "]"
}

func (s BrokerInfoRendererDot) renderExchangeFlagsAsString(exchange rabtap.RabbitExchange) string {
	flags := []bool{exchange.Durable, exchange.AutoDelete, exchange.Internal}
	names := []string{"D", "AD", "I"}
	return "[" + strings.Join(filterStringList(flags, names), "|") + "]"
}

func (s BrokerInfoRendererDot) renderRootNodeAsString(name string, children []dotNode, rabbitURL url.URL, overview rabtap.RabbitOverview) string {
	var args = struct {
		Name     string
		Children []dotNode
		Config   BrokerInfoRendererConfig
		URL      url.URL
		Overview rabtap.RabbitOverview
	}{name, children, s.config, rabbitURL, overview}
	funcMap := map[string]interface{}{"q": strconv.Quote}
	return resolveTemplate("root-dotTpl", dotTplRootNode, args, funcMap)
}
func (s BrokerInfoRendererDot) renderVhostAsString(name string, children []dotNode, vhost string) string {
	var args = struct {
		Name     string
		Children []dotNode
		Vhost    string
	}{name, children, vhost}
	funcMap := map[string]interface{}{"q": strconv.Quote}
	return resolveTemplate("vhost-dotTpl", dotTplVhost, args, funcMap)
}

func (s BrokerInfoRendererDot) renderExchangeElementAsString(name string, children []dotNode, exchange rabtap.RabbitExchange) string {
	exchangeFlags := s.renderExchangeFlagsAsString(exchange)
	var args = struct {
		Name          string
		Children      []dotNode
		Config        BrokerInfoRendererConfig
		Exchange      rabtap.RabbitExchange
		ExchangeFlags string
	}{name, children, s.config, exchange, exchangeFlags}
	funcMap := map[string]interface{}{"q": strconv.Quote}
	return resolveTemplate("exchange-dotTpl", dotTplExchange, args, funcMap)
}

func (s BrokerInfoRendererDot) renderQueueElementAsString(name string, children []dotNode, queue rabtap.RabbitQueue) string {
	queueFlags := s.renderQueueFlagsAsString(queue)
	var args = struct {
		Name       string
		Children   []dotNode
		Config     BrokerInfoRendererConfig
		Queue      rabtap.RabbitQueue
		QueueFlags string
	}{name, children, s.config, queue, queueFlags}
	funcMap := map[string]interface{}{"q": strconv.Quote}
	return resolveTemplate("queue-dotTpl", dotTplQueue, args, funcMap)
}

func (s BrokerInfoRendererDot) renderBoundQueueElementAsString(name string, children []dotNode, queue rabtap.RabbitQueue, binding rabtap.RabbitBinding) string {
	queueFlags := s.renderQueueFlagsAsString(queue)
	var args = struct {
		Name       string
		Children   []dotNode
		Config     BrokerInfoRendererConfig
		Binding    rabtap.RabbitBinding
		Queue      rabtap.RabbitQueue
		QueueFlags string
	}{name, children, s.config, binding, queue, queueFlags}
	funcMap := map[string]interface{}{"q": strconv.Quote}
	return resolveTemplate("bound-queue-dotTpl", dotTplBoundQueue, args, funcMap)
}
func (s BrokerInfoRendererDot) renderConsumerElementAsString(name string, children []dotNode, consumer rabtap.RabbitConsumer) string {
	var args = struct {
		Name     string
		Children []dotNode
		Config   BrokerInfoRendererConfig
		Consumer rabtap.RabbitConsumer
	}{name, children, s.config, consumer}
	funcMap := map[string]interface{}{"q": strconv.Quote}
	return resolveTemplate("consumer-dotTpl", dotTplConsumer, args, funcMap)
}

func (s BrokerInfoRendererDot) renderConnectionElementAsString(name string, children []dotNode, conn rabtap.RabbitConnection) string {
	var args = struct {
		Name       string
		Children   []dotNode
		Config     BrokerInfoRendererConfig
		Connection rabtap.RabbitConnection
	}{name, children, s.config, conn}
	funcMap := map[string]interface{}{"q": strconv.Quote}
	return resolveTemplate("connnection-dotTpl", dotTplConnection, args, funcMap)
}

func (s *BrokerInfoRendererDot) renderNodeText(n interface{}) dotNode {
	var node dotNode

	children := []dotNode{}
	for _, child := range n.(Node).Children() {
		c := s.renderNodeText(child.(Node))
		children = append(children, c)
	}

	switch t := n.(type) {
	case *rootNode:
		name := "root"
		node = dotNode{name, s.renderRootNodeAsString(name, children, n.(*rootNode).URL, n.(*rootNode).Overview), ""}
	case *vhostNode:
		vhost := n.(*vhostNode)
		name := fmt.Sprintf("vhost_%s", vhost.Vhost)
		node = dotNode{name, s.renderVhostAsString(name, children, vhost.Vhost), ""}
	case *connectionNode:
		conn := n.(*connectionNode)
		name := fmt.Sprintf("connection_%s", conn.Connection.Name)
		node = dotNode{name, s.renderConnectionElementAsString(name, children, conn.Connection), ""}
	case *consumerNode:
		cons := n.(*consumerNode)
		name := fmt.Sprintf("consumer_%s", cons.Consumer.ConsumerTag)
		node = dotNode{name, s.renderConsumerElementAsString(name, children, cons.Consumer), ""}
	case *queueNode:
		queue := n.(*queueNode).Queue
		name := fmt.Sprintf("queue_%s", queue.Name)
		node = dotNode{name, s.renderQueueElementAsString(name, children, queue), ""}
	case *boundQueueNode:
		boundQueue := n.(*boundQueueNode)
		queue := boundQueue.Queue
		binding := boundQueue.Binding
		name := fmt.Sprintf("boundqueue_%s", queue.Name)
		node = dotNode{name, s.renderBoundQueueElementAsString(name, children, queue, binding), binding.RoutingKey}
	case *exchangeNode:
		exchange := n.(*exchangeNode).Exchange
		name := fmt.Sprintf("exchange_%s", exchange.Name)
		node = dotNode{name, s.renderExchangeElementAsString(name, children, exchange), ""}
	default:
		panic(fmt.Sprintf("unexpected node encountered %T", t))
	}
	return node
}

func (s *BrokerInfoRendererDot) Render(rootNode *rootNode, out io.Writer) error {
	res := s.renderNodeText(rootNode)
	//res := s.render(rootNode)
	fmt.Fprintf(out, res.Text)
	return nil
}
