// Copyright (C) 2017 Jan Delgado
// TODO split in renderer, tree-builder-by-exchange, tree-builder-by-connection

package main

import (
	"io"
	"net/url"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// BrokerInfoRenderer renders a tree representation represented by a RootNode
// into a string representation
type BrokerInfoRenderer interface {
	Render(rootNode *RootNode, out io.Writer) error
}

type Node interface {
	Add(elem interface{})
	Children() []interface{}
}

type BaseNode struct {
	Children_ []interface{}
}

func (s *BaseNode) Add(elem interface{}) {
	s.Children_ = append(s.Children_, elem)
}

func (s *BaseNode) Children() []interface{} {
	return s.Children_
}

func (s *BaseNode) HasChildren() bool {
	return len(s.Children_) > 0
}

type RootNode struct {
	BaseNode
	Overview rabtap.RabbitOverview
	URL      url.URL
}

type VhostNode struct {
	BaseNode
	Vhost string
}

type ExchangeNode struct {
	BaseNode
	Exchange rabtap.RabbitExchange
}

type QueueNode struct {
	BaseNode
	Queue rabtap.RabbitQueue
}

type BoundQueueNode struct {
	BaseNode
	Queue   rabtap.RabbitQueue
	Binding rabtap.RabbitBinding
}

type ConnectionNode struct {
	BaseNode
	Connection rabtap.RabbitConnection
}

type ChannelNode struct {
	BaseNode
	Channel rabtap.RabbitConnection
}

type ConsumerNode struct {
	BaseNode
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
	vhost string, connName string, brokerInfo rabtap.BrokerInfo) []*ConnectionNode {
	var conns []*ConnectionNode
	i := rabtap.FindConnectionByName(brokerInfo.Connections, vhost, connName)
	if i != -1 {
		//		conns = append(conns, NewTreeNode(s.renderConnectionElementAsString(brokerInfo.Connections[i])))
		return []*ConnectionNode{{BaseNode{[]interface{}{}}, brokerInfo.Connections[i]}}
	}
	return conns
}

func (s BrokerInfoPrinter) createConsumerNodes(
	queue rabtap.RabbitQueue, brokerInfo rabtap.BrokerInfo) []*ConsumerNode {
	var nodes []*ConsumerNode
	vhost := queue.Vhost
	for _, consumer := range brokerInfo.Consumers {
		if consumer.Queue.Vhost == vhost &&
			consumer.Queue.Name == queue.Name {
			//consumerNode := NewTreeNode(s.renderConsumerElementAsString(consumer))
			consumerNode := ConsumerNode{BaseNode{[]interface{}{}}, consumer}
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
	brokerInfo rabtap.BrokerInfo) []*BoundQueueNode {

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
		return []*BoundQueueNode{}
	}

	queueNode := BoundQueueNode{BaseNode{[]interface{}{}}, queue, binding}

	if s.config.ShowConsumers {
		consumers := s.createConsumerNodes(queue, brokerInfo)
		//    queueNode.AddList(consumers)
		for _, consumer := range consumers {
			queueNode.Add(consumer)
		}
	}
	return []*BoundQueueNode{&queueNode}
}

// addExchange recursively (in case of exchange-exchange binding) an exchange to the
// given node.
func (s BrokerInfoPrinter) createExchangeNode(
	exchange rabtap.RabbitExchange, brokerInfo rabtap.BrokerInfo) *ExchangeNode {

	//exchangeNode := NewTreeNode(s.renderExchangeElementAsString(exchange))
	exchangeNode := ExchangeNode{BaseNode{[]interface{}{}}, exchange}

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

// func (s BrokerInfoPrinter) makeRootNode(rootNodeURL string,
//     overview rabtap.RabbitOverview) (*TreeNode, error) {
//     // root of node is URL of rabtap.RabbitMQ broker.
//     url, err := url.Parse(rootNodeURL)
//     if err != nil {
//         return nil, err
//     }
//     return NewTreeNode(s.renderRootNodeAsString(*url, overview)), nil
// }

// buildTree renders given brokerInfo into a tree:
//  RabbitMQ-Host
//  +--VHost
//     +--Exchange
//        +--Queue bound to exchange
//           +--Consumer  (optional)
//              +--Connection
//
func (s BrokerInfoPrinter) buildTreeByExchange(rootNodeURL string,
	brokerInfo rabtap.BrokerInfo) (*RootNode, error) {

	url, err := url.Parse(rootNodeURL)
	if err != nil {
		return nil, err
	}

	b := BaseNode{[]interface{}{}}
	rootNode2 := RootNode{b, brokerInfo.Overview, *url}

	for vhost := range uniqueVhosts(brokerInfo.Exchanges) {
		// vhostNode := NewTreeNode(s.renderVhostAsString(vhost))
		// root.Add(vhostNode)
		vhostNode2 := VhostNode{b, vhost}
		for _, exchange := range brokerInfo.Exchanges {
			if !s.shouldDisplayExchange(exchange, vhost) {
				continue
			}
			exNode := s.createExchangeNode(exchange, brokerInfo)
			if s.config.OmitEmptyExchanges && !exNode.HasChildren() {
				continue
			}
			vhostNode2.Add(exNode)
		}
		//rootNode2.Children = append(root.Children, vhostNode2)
		rootNode2.Add(&vhostNode2)
	}
	return &rootNode2, nil
}

func (s BrokerInfoPrinter) buildTreeByConnection2(rootNodeURL string,
	brokerInfo rabtap.BrokerInfo) (*RootNode, error) {

	url, err := url.Parse(rootNodeURL)
	if err != nil {
		return nil, err
	}
	b := BaseNode{[]interface{}{}}
	rootNode2 := RootNode{b, brokerInfo.Overview, *url}
	for vhost := range uniqueVhosts(brokerInfo.Exchanges) {
		vhostNode2 := VhostNode{b, vhost}
		for _, conn := range brokerInfo.Connections {
			connNode2 := ConnectionNode{b, conn}
			for _, consumer := range brokerInfo.Consumers {
				if consumer.ChannelDetails.ConnectionName == conn.Name {
					consNode2 := ConsumerNode{b, consumer}
					for _, queue := range brokerInfo.Queues {
						if consumer.Queue.Vhost == vhost && consumer.Queue.Name == queue.Name {
							queueNode2 := QueueNode{b, queue}
							consNode2.Add(&queueNode2)
						}
					}
					connNode2.Add(&consNode2)
				}
			}
			if s.config.OmitEmptyExchanges && !connNode2.HasChildren() {
				continue
			}
			vhostNode2.Add(&connNode2)
		}
		rootNode2.Add(&vhostNode2)
	}
	return &rootNode2, nil
}

// Print renders given brokerInfo into a tree-view
func (s BrokerInfoPrinter) Print(brokerInfo rabtap.BrokerInfo,
	rootNodeURL string, out io.Writer) error {

	var root *RootNode
	var err error
	if s.config.ShowByConnection {
		root, err = s.buildTreeByConnection2(rootNodeURL, brokerInfo)
	} else {
		root, err = s.buildTreeByExchange(rootNodeURL, brokerInfo)
	}

	if err != nil {
		return err
	}

	renderer := NewBrokerInfoRendererText(s.config)
	return renderer.Render(root, out)
}
