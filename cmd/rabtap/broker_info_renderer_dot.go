// Copyright (C) 2017-2019 Jan Delgado
// Render broker info into a dot graph representation
// https://www.graphviz.org/doc/info/lang.html

package main

import (
	"fmt"
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
	dotTplRootNode   string
	dotTplVhost      string
	dotTplExchange   string
	dotTplQueue      string
	dotTplBoundQueue string
	dotTplConsumer   string
	dotTplConnection string
}

// brokerInfoRendererDot renders into graphviz dot format
type brokerInfoRendererDot struct {
	config   BrokerInfoRendererConfig
	template dotRendererTpl
	rendered map[string]bool // keeps track of rendered elements
}

type dotNode struct {
	Name        string
	Text        string
	ParentAssoc string
}

// NewBrokerInfoRendererDot returns a BrokerInfoRenderer implementation that
// renders into graphviz dot format
func NewBrokerInfoRendererDot(config BrokerInfoRendererConfig) BrokerInfoRenderer {
	return &brokerInfoRendererDot{config: config, template: newDotRendererTpl()}
}

// newDotRendererTpl returns the dot template to use. For now, just one default
// template is used, later will support loading templates from the filesytem
func newDotRendererTpl() dotRendererTpl {
	return dotRendererTpl{dotTplRootNode: `digraph broker {
{{ q .Name }} [shape="record", label="{RabbitMQ {{ .Overview.RabbitmqVersion }} | 
               {{- printf "%s://%s%s" .URL.Scheme .URL.Host .URL.Path }} |
               {{- .Overview.ClusterName }} }"];

{{ range $i, $e := .Children }}{{ q $.Name }} -> {{ q $e.Name }}{{ printf ";\n" }}{{ end -}}
{{ range $i, $e := .Children }}{{ $e.Text -}}{{ end -}}
}`,

		dotTplVhost: `{{ q .Name }} [shape="box", label="Virtual host\n{{ .Vhost }}"];

{ rank = same; {{ range $i, $e := .Children }}{{ q $e.Name }}; {{ end -}} };
{{ range $i, $e := .Children }}{{ q $.Name }} -> {{ q $e.Name -}} [headport=n]{{ printf ";\n" }}{{ end -}}
{{ range $i, $e := .Children }}{{ $e.Text -}}{{ end -}}`,

		//         dotTplExchange: `
		// {{ q .Name }} [shape="record"; label="{ { E | {{ .Exchange.Type }} } | {
		//               {{- if eq .Exchange.Name "" }} (default) {{else}} {{ .Exchange.Name }} {{end}} } | {
		//               {{- if .Exchange.Durable }} D {{ end }} |
		//               {{- if .Exchange.AutoDelete }} AD {{ end }} |
		//               {{- if .Exchange.Internal }} I {{ end }} } } }"];
		dotTplExchange: `{{ q .Name }} [shape="none"; margin="0"; label=< 
		    <TABLE border='0' cellborder='1' cellspacing='0'>
              <TR><TD colspan='1' WIDTH='33%%'> E </TD><TD colspan='2' balign='center'>{{ .Exchange.Type }}</TD></TR>
              <TR><TD colspan='3' align='text'><B>{{- if eq .Exchange.Name "" }}(default){{else}}{{ .Exchange.Name }}{{end}}</B></TD></TR>
			  <TR><TD WIDTH='33%%'>{{- if .Exchange.Durable }} D {{ else }} &nbsp; {{ end }}</TD>
				  <TD WIDTH='33%%'>{{- if .Exchange.AutoDelete }} AD {{ end }}</TD>
				  <TD WIDTH='33%%'>{{- if .Exchange.Internal }} I {{ end }}</TD></TR>
		    </TABLE> >];

	{{ range $i, $e := .Children }}{{ q $.Name }} -> {{ q $e.Name }} [fontsize=10; label={{ q $e.ParentAssoc }}]{{ printf ";\n" }}{{ end -}}
{{ range $i, $e := .Children }}{{ $e.Text }}{{ end -}}`,

		dotTplQueue: `
{{ q .Name }} [shape="none"; margin="0"; label=<
		     <TABLE border='0' cellborder='1' cellspacing='0'>
			   <TR><TD colspan='4' WIDTH='25%%'> Q</TD><TD><B>{{ .Queue.Name }}</B></TD></TR>
			   <TR><TD WIDTH='25%%'>{{- if .Queue.Durable }} D {{else}} &nbsp; {{ end }}</TD> 
			       <TD WIDTH='25%%'>{{- if .Queue.AutoDelete }} AD {{ end }} </TD>
			       <TD WIDTH='25%%'>{{- if .Queue.Exclusive }} EX {{ end }} </TD>
			       <TD  WIDTH='25%%' PORT='dlx'>{{- if .Queue.HasDlx }}<dlx> DLX {{ end }}</TD></TR>
				</TABLE> >];

{{ if .Queue.HasDlx }}{{ q .Name }}:dlx -> {{ q ( print "exchange_"  .Queue.Dlx )}} [style="dashed"];{{ end -}}
{{ range $i, $e := .Children }}{{ q $.Name }} -> {{ q $e.Name }}{{ printf ";\n" }}{{ end -}}
{{ range $i, $e := .Children }}{{ $e.Text }}{{ end -}}`,

		//         dotTplQueue: `
		// {{ q .Name }} [shape="record"; label="{ { Q | {{ .Queue.Name }} } | {
		//               {{- if .Queue.Durable }} D {{ end }} |
		//               {{- if .Queue.AutoDelete }} AD {{ end }} |
		//               {{- if .Queue.Exclusive }} EX {{ end }} |
		//               {{- if .Queue.HasDlx }}<dlx> DLX {{ end }} } }"];

		// {{ if .Queue.HasDlx }}{{ q .Name }}:dlx -> {{ q ( print "exchange_"  .Queue.Dlx )}} [style="dashed"];{{ end -}}
		// {{ range $i, $e := .Children }}{{ q $.Name }} -> {{ q $e.Name }}{{ printf ";\n" }}{{ end -}}
		// {{ range $i, $e := .Children }}{{ $e.Text }}{{ end -}}`,

		dotTplBoundQueue: `
{{- if not .Skip }}
{{ q .Name }} [shape="none"; margin="0"; label=<
		     <TABLE border='0' cellborder='1' cellspacing='0'>
			   <TR><TD colspan='1' WIDTH='25%%'> Q</TD><TD colspan='3'><B>{{ .Queue.Name }}</B></TD></TR>
			   <TR><TD WIDTH='25%%'>{{- if .Queue.Durable }} D {{else}} &nbsp; {{ end }}</TD> 
			       <TD WIDTH='25%%'>{{- if .Queue.AutoDelete }} AD {{ end }} </TD>
			       <TD WIDTH='25%%'>{{- if .Queue.Exclusive }} EX {{ end }} </TD>
			       <TD  WIDTH='25%%' PORT='dlx'>{{- if .Queue.HasDlx }} DLX {{ end }}</TD></TR>
				</TABLE> >];

{{ if .Queue.HasDlx }}{{ q .Name }}:dlx -> {{ q ( print "exchange_"  .Queue.Dlx )}} [style="dashed"];{{ end -}}
{{- end}}

{{ range $i, $e := .Children }}{{ q $.Name }} -> {{ q $e.Name }}{{ end -}}
{{ range $i, $e := .Children }}{{ $e.Text -}}{{ end -}}`,

		// TODO add more details
		dotTplConnection: `
{{ q .Name }} [shape="record" label="{ Conn | {{ .Connection.Name }} }"];

{{ range $i, $e := .Children }}{{ q $.Name }} -> {{ q $e.Name }}{{ end -}}
{{ range $i, $e := .Children }}{{ $e.Text -}}{{ end -}}`,

		// TODO add more details
		dotTplConsumer: `
{{ q .Name }} [shape="record" label="{ Cons | {{ .Consumer.ConsumerTag}} }"];

{{ range $i, $e := .Children }}{{ q $.Name }} -> {{ q $e.Name }}{{ end -}}
{{ range $i, $e := .Children }}{{ $e.Text -}}{{ end -}}`,
	}
}

func (s brokerInfoRendererDot) renderRootNodeAsString(name string, children []dotNode, rabbitURL url.URL, overview rabtap.RabbitOverview) string {
	var args = struct {
		Name     string
		Children []dotNode
		Config   BrokerInfoRendererConfig
		URL      url.URL
		Overview rabtap.RabbitOverview
	}{name, children, s.config, rabbitURL, overview}
	funcMap := map[string]interface{}{"q": strconv.Quote}
	return resolveTemplate("root-dotTpl", s.template.dotTplRootNode, args, funcMap)
}

func (s brokerInfoRendererDot) renderVhostAsString(name string, children []dotNode, vhost string) string {
	var args = struct {
		Name     string
		Children []dotNode
		Vhost    string
	}{name, children, vhost}
	funcMap := map[string]interface{}{"q": strconv.Quote}
	return resolveTemplate("vhost-dotTpl", s.template.dotTplVhost, args, funcMap)
}

func (s brokerInfoRendererDot) renderExchangeElementAsString(name string, children []dotNode, exchange rabtap.RabbitExchange) string {
	var args = struct {
		Name     string
		Children []dotNode
		Config   BrokerInfoRendererConfig
		Exchange rabtap.RabbitExchange
	}{name, children, s.config, exchange}
	funcMap := map[string]interface{}{"q": strconv.Quote}
	return resolveTemplate("exchange-dotTpl", s.template.dotTplExchange, args, funcMap)
}

func (s brokerInfoRendererDot) renderQueueElementAsString(name string, children []dotNode, queue rabtap.RabbitQueue) string {
	var args = struct {
		Name     string
		Children []dotNode
		Config   BrokerInfoRendererConfig
		Queue    rabtap.RabbitQueue
	}{name, children, s.config, queue}
	funcMap := map[string]interface{}{"q": strconv.Quote}
	return resolveTemplate("queue-dotTpl", s.template.dotTplQueue, args, funcMap)
}

func (s brokerInfoRendererDot) renderBoundQueueElementAsString(name string, children []dotNode, queue rabtap.RabbitQueue, binding rabtap.RabbitBinding, skip bool) string {
	var args = struct {
		Name     string
		Children []dotNode
		Config   BrokerInfoRendererConfig
		Binding  rabtap.RabbitBinding
		Queue    rabtap.RabbitQueue
		Skip     bool
	}{name, children, s.config, binding, queue, skip}
	funcMap := map[string]interface{}{"q": strconv.Quote}
	return resolveTemplate("bound-queue-dotTpl", s.template.dotTplBoundQueue, args, funcMap)
}
func (s brokerInfoRendererDot) renderConsumerElementAsString(name string, children []dotNode, consumer rabtap.RabbitConsumer) string {
	var args = struct {
		Name     string
		Children []dotNode
		Config   BrokerInfoRendererConfig
		Consumer rabtap.RabbitConsumer
	}{name, children, s.config, consumer}
	funcMap := map[string]interface{}{"q": strconv.Quote}
	return resolveTemplate("consumer-dotTpl", s.template.dotTplConsumer, args, funcMap)
}

func (s brokerInfoRendererDot) renderConnectionElementAsString(name string, children []dotNode, conn rabtap.RabbitConnection) string {
	var args = struct {
		Name       string
		Children   []dotNode
		Config     BrokerInfoRendererConfig
		Connection rabtap.RabbitConnection
	}{name, children, s.config, conn}
	funcMap := map[string]interface{}{"q": strconv.Quote}
	return resolveTemplate("connnection-dotTpl", s.template.dotTplConnection, args, funcMap)
}

func (s *brokerInfoRendererDot) renderNode(n interface{}) dotNode {
	var node dotNode

	children := []dotNode{}
	for _, child := range n.(Node).Children() {
		c := s.renderNode(child.(Node))
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
	case *exchangeNode:
		exchange := n.(*exchangeNode).Exchange
		name := fmt.Sprintf("exchange_%s", exchange.Name)
		node = dotNode{name, s.renderExchangeElementAsString(name, children, exchange), ""}
	case *queueNode:
		queue := n.(*queueNode).Queue
		name := fmt.Sprintf("queue_%s", queue.Name)
		node = dotNode{name, s.renderQueueElementAsString(name, children, queue), ""}
	case *boundQueueNode:
		boundQueue := n.(*boundQueueNode)
		queue := boundQueue.Queue
		binding := boundQueue.Binding
		name := fmt.Sprintf("boundqueue_%s", queue.Name)
		// don't render bound queue nodes more than once
		_, skip := s.rendered[name]
		s.rendered[name] = true
		node = dotNode{name, s.renderBoundQueueElementAsString(name, children, queue, binding, skip), binding.RoutingKey}
	case *connectionNode:
		conn := n.(*connectionNode)
		name := fmt.Sprintf("connection_%s", conn.Connection.Name)
		node = dotNode{name, s.renderConnectionElementAsString(name, children, conn.Connection), ""}
	case *consumerNode:
		cons := n.(*consumerNode)
		name := fmt.Sprintf("consumer_%s", cons.Consumer.ConsumerTag)
		node = dotNode{name, s.renderConsumerElementAsString(name, children, cons.Consumer), ""}
	default:
		panic(fmt.Sprintf("unexpected node encountered %T", t))
	}
	return node
}

// Render renders the given tree in graphviz dot format. See
// https://www.graphviz.org/doc/info/lang.html
func (s *brokerInfoRendererDot) Render(rootNode *rootNode, out io.Writer) error {
	s.rendered = map[string]bool{}
	res := s.renderNode(rootNode)
	fmt.Fprintf(out, res.Text)
	return nil
}
