// Copyright (C) 2017 Jan Delgado
// TODO split in renderer, tree-builder-by-exchange, tree-builder-by-connection

package main

import (
	"io"
	"net/url"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// BrokerInfoRenderer renders a tree representation represented by a rootNode
// into a string representation
type BrokerInfoRenderer interface {
	Render(rootNode *rootNode, out io.Writer) error
}

type Node interface {
	Add(elem interface{})
	Children() []interface{}
}

type baseNode struct {
	children []interface{}
}

func (s *baseNode) Add(elem interface{}) {
	s.children = append(s.children, elem)
}

func (s *baseNode) Children() []interface{} {
	return s.children
}

func (s *baseNode) HasChildren() bool {
	return len(s.children) > 0
}

type rootNode struct {
	baseNode
	Overview rabtap.RabbitOverview
	URL      url.URL
}

type vhostNode struct {
	baseNode
	Vhost string
}

type exchangeNode struct {
	baseNode
	Exchange rabtap.RabbitExchange
}

type queueNode struct {
	baseNode
	Queue rabtap.RabbitQueue
}

type boundQueueNode struct {
	baseNode
	Queue   rabtap.RabbitQueue
	Binding rabtap.RabbitBinding
}

type connectionNode struct {
	baseNode
	Connection rabtap.RabbitConnection
}

type ChannelNode struct {
	baseNode
	Channel rabtap.RabbitConnection
}

type consumerNode struct {
	baseNode
	Consumer rabtap.RabbitConsumer
}

// BrokerInfoPrinterConfig controls bevaviour when rendering a broker info
type BrokerInfoPrinterConfig struct {
	ShowDefaultExchange bool
	ShowConsumers       bool
	ShowStats           bool
	ShowByConnection    bool
	QueueFilter         Predicate
	OmitEmptyExchanges  bool
	NoColor             bool
}

type BrokerInfoPrinter struct {
	config BrokerInfoPrinterConfig
}

func NewBrokerInfoPrinter(config BrokerInfoPrinterConfig) *BrokerInfoPrinter {
	return &BrokerInfoPrinter{config}
}

// uniqueVhosts returns the set of unique vhosts in the array of exchanges
func uniqueVhosts(exchanges []rabtap.RabbitExchange) (vhosts map[string]bool) {
	vhosts = make(map[string]bool)
	for _, exchange := range exchanges {
		vhosts[exchange.Vhost] = true
	}
	return
}

func findBindingsForExchange(exchange rabtap.RabbitExchange, bindings []rabtap.RabbitBinding) []rabtap.RabbitBinding {
	var result []rabtap.RabbitBinding
	for _, binding := range bindings {
		if binding.Source == exchange.Name &&
			binding.Vhost == exchange.Vhost {
			result = append(result, binding)
		}
	}
	return result
}

func (s BrokerInfoPrinter) shouldDisplayExchange(
	exchange rabtap.RabbitExchange, vhost string) bool {

	if exchange.Vhost != vhost {
		return false
	}
	if exchange.Name == "" && !s.config.ShowDefaultExchange {
		return false
	}

	return true
}

func (s BrokerInfoPrinter) shouldDisplayQueue(
	queue rabtap.RabbitQueue,
	exchange rabtap.RabbitExchange,
	binding rabtap.RabbitBinding) bool {

	// apply filter
	params := map[string]interface{}{"queue": queue, "binding": binding, "exchange": exchange}
	if res, err := s.config.QueueFilter.Eval(params); err != nil || !res {
		if err != nil {
			log.Warnf("error evaluating queue filter: %s", err)
		} else {
			return false
		}
	}
	return true
}

func (s BrokerInfoPrinter) createConnectionNodes(
	vhost string, connName string, brokerInfo rabtap.BrokerInfo) []*connectionNode {
	var conns []*connectionNode
	i := rabtap.FindConnectionByName(brokerInfo.Connections, vhost, connName)
	if i != -1 {
		return []*connectionNode{{baseNode{[]interface{}{}}, brokerInfo.Connections[i]}}
	}
	return conns
}

func (s BrokerInfoPrinter) createConsumerNodes(
	queue rabtap.RabbitQueue, brokerInfo rabtap.BrokerInfo) []*consumerNode {
	var nodes []*consumerNode
	vhost := queue.Vhost
	for _, consumer := range brokerInfo.Consumers {
		if consumer.Queue.Vhost == vhost &&
			consumer.Queue.Name == queue.Name {
			//consumerNode := NewTreeNode(s.renderConsumerElementAsString(consumer))
			consumerNode := consumerNode{baseNode{[]interface{}{}}, consumer}
			connectionNodes := s.createConnectionNodes(vhost, consumer.ChannelDetails.ConnectionName, brokerInfo)
			// consumerNode.AddList(s.createConnectionNodes(vhost, consumer.ChannelDetails.ConnectionName, brokerInfo))
			for _, connectionNode := range connectionNodes {
				consumerNode.Add(connectionNode)
			}
			nodes = append(nodes, &consumerNode)
		}
	}
	return nodes
}

func (s BrokerInfoPrinter) createQueueNodeFromBinding(
	binding rabtap.RabbitBinding,
	exchange rabtap.RabbitExchange,
	brokerInfo rabtap.BrokerInfo) []*boundQueueNode {

	// standard binding of queue to exchange
	i := rabtap.FindQueueByName(brokerInfo.Queues,
		binding.Vhost,
		binding.Destination)

	queue := rabtap.RabbitQueue{Name: binding.Destination} // default in case not found
	if i != -1 {
		// we test for -1 because (at least in theory) a queue can disappear
		// since we are making various non-transactional API calls
		queue = brokerInfo.Queues[i]
	}

	if !s.shouldDisplayQueue(queue, exchange, binding) {
		return []*boundQueueNode{}
	}

	queueNode := boundQueueNode{baseNode{[]interface{}{}}, queue, binding}

	if s.config.ShowConsumers {
		consumers := s.createConsumerNodes(queue, brokerInfo)
		//    queueNode.AddList(consumers)
		for _, consumer := range consumers {
			queueNode.Add(consumer)
		}
	}
	return []*boundQueueNode{&queueNode}
}

// addExchange recursively (in case of exchange-exchange binding) an exchange to the
// given node.
func (s BrokerInfoPrinter) createExchangeNode(
	exchange rabtap.RabbitExchange, brokerInfo rabtap.BrokerInfo) *exchangeNode {

	//exchangeNode := NewTreeNode(s.renderExchangeElementAsString(exchange))
	exchangeNode := exchangeNode{baseNode{[]interface{}{}}, exchange}

	// process all bindings for current exchange
	for _, binding := range findBindingsForExchange(exchange, brokerInfo.Bindings) {
		if binding.DestinationType == "exchange" {
			// exchange to exchange binding
			i := rabtap.FindExchangeByName(
				brokerInfo.Exchanges,
				binding.Vhost,
				binding.Destination)
			if i != -1 {
				exchangeNode.Add(
					s.createExchangeNode(
						brokerInfo.Exchanges[i],
						brokerInfo))
			} // TODO else log error
		} else {
			// queue to exchange binding
			queues := s.createQueueNodeFromBinding(binding, exchange, brokerInfo)
			for _, queue := range queues {
				exchangeNode.Add(queue)
			}
			//			exchangeNode.AddList(s.createQueueNodeFromBinding(binding, exchange, brokerInfo))
		}
	}
	return &exchangeNode
}

func (s BrokerInfoPrinter) createRootNode(rootNodeURL string,
	overview rabtap.RabbitOverview) (*rootNode, error) {
	// root of node is URL of rabtap.RabbitMQ broker.
	url, err := url.Parse(rootNodeURL)
	if err != nil {
		return nil, err
	}
	b := baseNode{[]interface{}{}}
	return &rootNode{b, overview, *url}, nil
}

// buildTree renders given brokerInfo into a tree:
//  RabbitMQ-Host
//  +--VHost
//     +--Exchange
//        +--Queue bound to exchange
//           +--Consumer  (optional)
//              +--Connection
//
func (s BrokerInfoPrinter) buildTreeByExchange(rootNodeURL string,
	brokerInfo rabtap.BrokerInfo) (*rootNode, error) {

	b := baseNode{[]interface{}{}}
	rootNode, err := s.createRootNode(rootNodeURL, brokerInfo.Overview)

	if err != nil {
		return nil, err
	}

	for vhost := range uniqueVhosts(brokerInfo.Exchanges) {
		// vhostNode := NewTreeNode(s.renderVhostAsString(vhost))
		// root.Add(vhostNode)
		vhostNode := vhostNode{b, vhost}
		for _, exchange := range brokerInfo.Exchanges {
			if !s.shouldDisplayExchange(exchange, vhost) {
				continue
			}
			exNode := s.createExchangeNode(exchange, brokerInfo)
			if s.config.OmitEmptyExchanges && !exNode.HasChildren() {
				continue
			}
			vhostNode.Add(exNode)
		}
		//rootNode2.Children = append(root.Children, vhostNode2)
		rootNode.Add(&vhostNode)
	}
	return rootNode, nil
}

func (s BrokerInfoPrinter) buildTreeByConnection(rootNodeURL string,
	brokerInfo rabtap.BrokerInfo) (*rootNode, error) {

	url, err := url.Parse(rootNodeURL)
	if err != nil {
		return nil, err
	}
	b := baseNode{[]interface{}{}}
	rootNode := rootNode{b, brokerInfo.Overview, *url}
	for vhost := range uniqueVhosts(brokerInfo.Exchanges) {
		vhostNode := vhostNode{b, vhost}
		for _, conn := range brokerInfo.Connections {
			connNode := connectionNode{b, conn}
			for _, consumer := range brokerInfo.Consumers {
				if consumer.ChannelDetails.ConnectionName == conn.Name {
					consNode := consumerNode{b, consumer}
					for _, queue := range brokerInfo.Queues {
						if consumer.Queue.Vhost == vhost && consumer.Queue.Name == queue.Name {
							queueNode := queueNode{b, queue}
							consNode.Add(&queueNode)
						}
					}
					connNode.Add(&consNode)
				}
			}
			if s.config.OmitEmptyExchanges && !connNode.HasChildren() {
				continue
			}
			vhostNode.Add(&connNode)
		}
		rootNode.Add(&vhostNode)
	}
	return &rootNode, nil
}

// Print renders given brokerInfo into a tree-view
func (s BrokerInfoPrinter) Print(brokerInfo rabtap.BrokerInfo,
	rootNodeURL string, out io.Writer) error {

	var root *rootNode
	var err error
	if s.config.ShowByConnection {
		root, err = s.buildTreeByConnection(rootNodeURL, brokerInfo)
	} else {
		root, err = s.buildTreeByExchange(rootNodeURL, brokerInfo)
	}

	if err != nil {
		return err
	}

	renderer := NewBrokerInfoRendererText(s.config)
	return renderer.Render(root, out)
}
