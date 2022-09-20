// Copyright (C) 2017-2019 Jan Delgado

// rendering of info return by broker API into abstract tree represenations,
// which can later be rendered into something useful (e.g. text, dot etc.)
// Definition of interface and default implementation.
// TODO add unit test. currently only component tested in cmd_info_test.go

package main

import (
	"fmt"
	"net/url"
	"sort"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// BrokerInfoTreeBuilderConfig controls bevaviour when rendering a broker info
type BrokerInfoTreeBuilderConfig struct {
	Mode                string
	ShowDefaultExchange bool
	ShowConsumers       bool
	ShowStats           bool
	QueueFilter         Predicate
	OmitEmptyExchanges  bool
}

// BrokerInfoTreeBuilder transforms a rabtap.BrokerInfo into a tree
// representation that can be easily rendered (e.g. into text, dot fomats)
type BrokerInfoTreeBuilder interface {
	BuildTree(rootNodeURL *url.URL, metadataService rabtap.MetadataService) (*rootNode, error)
}

type brokerInfoTreeBuilderByConnection struct{ config BrokerInfoTreeBuilderConfig }
type brokerInfoTreeBuilderByExchange struct{ config BrokerInfoTreeBuilderConfig }

// NewBrokerInfoTreeBuilder returns a BrokerInfoTreeBuilder implementation
// that builds a tree for the config.Mode
func NewBrokerInfoTreeBuilder(config BrokerInfoTreeBuilderConfig) BrokerInfoTreeBuilder {
	switch config.Mode {
	case "byConnection":
		return &brokerInfoTreeBuilderByConnection{config}
	case "byExchange":
		return &brokerInfoTreeBuilderByExchange{config}
	default:
		panic(fmt.Sprintf("invalid mode %s", config.Mode))
	}
}

// Node represents functionality to add/access child nodes on a tree node
// TODO rename to something less generic
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

func (s *baseNode) hasChildren() bool {
	return len(s.children) > 0
}

type rootNode struct {
	baseNode
	Overview *rabtap.RabbitOverview
	URL      *url.URL
}

type vhostNode struct {
	baseNode
	Vhost *rabtap.RabbitVhost
}

func newVhostNode(vhost *rabtap.RabbitVhost) *vhostNode {
	return &vhostNode{baseNode{[]interface{}{}}, vhost}
}

type exchangeNode struct {
	baseNode
	Exchange   *rabtap.RabbitExchange
	OptBinding *rabtap.RabbitBinding // optional binding in case of e-to-e binding
}

func newExchangeNode(exchange *rabtap.RabbitExchange, binding *rabtap.RabbitBinding) *exchangeNode {
	return &exchangeNode{baseNode{[]interface{}{}}, exchange, binding}
}

type queueNode struct {
	baseNode
	Queue      *rabtap.RabbitQueue
	OptBinding *rabtap.RabbitBinding // optional binding if queue is bound to exchange
}

func newQueueNode(queue *rabtap.RabbitQueue) *queueNode {
	return &queueNode{baseNode{[]interface{}{}}, queue, nil}
}

func newQueueNodeWithBinding(queue *rabtap.RabbitQueue, binding *rabtap.RabbitBinding) *queueNode {
	return &queueNode{baseNode{[]interface{}{}}, queue, binding}
}

type nodeStatus int

const (
	// object was found and contains actual data
	Valid nodeStatus = iota
	// object was not found and usually contains only the name attribute set
	NotFound
)

type connectionNode struct {
	baseNode
	Connection *rabtap.RabbitConnection
	Status     nodeStatus
}

func newConnectionNode(connection *rabtap.RabbitConnection, status nodeStatus) *connectionNode {
	return &connectionNode{baseNode{[]interface{}{}}, connection, status}
}

type channelNode struct {
	baseNode
	Channel *rabtap.RabbitChannel
	Status  nodeStatus
}

func newChannelNode(channel *rabtap.RabbitChannel, status nodeStatus) *channelNode {
	return &channelNode{baseNode{[]interface{}{}}, channel, status}
}

type consumerNode struct {
	baseNode
	Consumer *rabtap.RabbitConsumer
}

func newConsumerNode(consumer *rabtap.RabbitConsumer) *consumerNode {
	return &consumerNode{baseNode{[]interface{}{}}, consumer}
}

type defaultBrokerInfoTreeBuilder struct {
	config BrokerInfoTreeBuilderConfig
}

func newDefaultBrokerInfoTreeBuilder(config BrokerInfoTreeBuilderConfig) *defaultBrokerInfoTreeBuilder {
	return &defaultBrokerInfoTreeBuilder{config}
}

func (s defaultBrokerInfoTreeBuilder) shouldDisplayExchange(
	exchange *rabtap.RabbitExchange, vhost *rabtap.RabbitVhost) bool {

	if exchange.Vhost != vhost.Name {
		return false
	}
	if exchange.Name == "" && !s.config.ShowDefaultExchange {
		return false
	}

	return true
}

// orderedKeySet returns the key set of the given map as a sorted array of strings
func orderedKeySet[T any](m map[string]T) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (s defaultBrokerInfoTreeBuilder) shouldDisplayQueue(
	queue *rabtap.RabbitQueue,
	exchange *rabtap.RabbitExchange,
	binding *rabtap.RabbitBinding) bool {

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

// createConnectionNodes creates a tree for the given queue with all
// connections -> channels -> consumers consuming from this queue.
func (s defaultBrokerInfoTreeBuilder) createConnectionNodes(
	queue *rabtap.RabbitQueue,
	metadataService rabtap.MetadataService) []*connectionNode {

	connectionNodes := map[string]*connectionNode{}
	channelNodes := map[string]*channelNode{}

	vhostName := queue.Vhost
	for _, consumer := range metadataService.Consumers() { // TODO AllConsumersByQueue
		if !(consumer.Queue.Vhost == vhostName && consumer.Queue.Name == queue.Name) {
			continue
		}
		consumerNode := newConsumerNode(&consumer)

		var ok bool

		connectionName := consumer.ChannelDetails.ConnectionName
		var connNode *connectionNode
		if connNode, ok = connectionNodes[connectionName]; !ok {
			connection := metadataService.FindConnectionByName(vhostName, connectionName)
			if connection != nil {
				connNode = newConnectionNode(connection, Valid)
			} else {
				// for some reason, the connection could not be found by it's name.
				// So we create an empty connection object and mark it as "Not found"
				// and let the renderer decide what to do.
				dummyConnection := rabtap.RabbitConnection{Name: connectionName}
				connNode = newConnectionNode(&dummyConnection, NotFound)
			}
			connectionNodes[connectionName] = connNode
		}

		channelName := consumer.ChannelDetails.Name
		var chanNode *channelNode
		if chanNode, ok = channelNodes[channelName]; !ok {
			channel := metadataService.FindChannelByName(vhostName, channelName)
			if channel != nil {
				chanNode = newChannelNode(channel, Valid)
			} else {
				dummyChan := rabtap.RabbitChannel{Name: channelName}
				chanNode = newChannelNode(&dummyChan, NotFound)
			}
			channelNodes[channelName] = chanNode
		}

		connNode.Add(chanNode)
		chanNode.Add(consumerNode)
	}

	var nodes []*connectionNode
	for _, key := range orderedKeySet(connectionNodes) {
		nodes = append(nodes, connectionNodes[key])
	}
	return nodes
}

func (s defaultBrokerInfoTreeBuilder) createQueueNodeFromBinding(
	binding *rabtap.RabbitBinding,
	exchange *rabtap.RabbitExchange,
	metadataService rabtap.MetadataService) []*queueNode {

	// standard binding of queue to exchange
	queue := metadataService.FindQueueByName(binding.Vhost, binding.Destination)

	if queue == nil {
		// can happen in theory since REST calls are not transactional
		return []*queueNode{}
	}

	if !s.shouldDisplayQueue(queue, exchange, binding) {
		return []*queueNode{}
	}

	node := newQueueNodeWithBinding(queue, binding)

	if s.config.ShowConsumers {
		consumers := s.createConnectionNodes(queue, metadataService)
		for _, consumer := range consumers {
			node.Add(consumer)
		}
	}
	return []*queueNode{node}
}

// createExchangeNode recursively (in case of exchange-exchange binding) an
// exchange to the given node.
func (s defaultBrokerInfoTreeBuilder) createExchangeNode(
	exchange *rabtap.RabbitExchange,
	metadataService rabtap.MetadataService,
	binding *rabtap.RabbitBinding) *exchangeNode {

	// to detect cyclic exchange-to-exchange bindings. Yes, this is possible.
	visited := map[string]bool{}

	var create func(*rabtap.RabbitExchange, rabtap.MetadataService, *rabtap.RabbitBinding) *exchangeNode
	create = func(exchange *rabtap.RabbitExchange, metadataService rabtap.MetadataService, binding *rabtap.RabbitBinding) *exchangeNode {

		exchangeNode := newExchangeNode(exchange, binding)

		// process all bindings for current exchange. Can be exchange-exchange-
		// as well as queue-to-exchange bindings.
		//for _, binding := range rabtap.FindBindingsForExchange(exchange, brokerInfo.Bindings) {
		for _, binding := range metadataService.AllBindingsForExchange(exchange.Vhost, exchange.Name) {
			if binding.IsExchangeToExchange() {
				boundExchange := metadataService.FindExchangeByName(binding.Vhost, binding.Destination)
				if boundExchange == nil {
					// ignore if not found
					continue
				}
				if _, found := visited[boundExchange.Name]; found {
					// cyclic exchange-to-exchange binding detected
					continue
				}
				visited[boundExchange.Name] = true
				exchangeNode.Add(create(boundExchange, metadataService, binding))
			} else {
				// do not add (redundant) queues if in recursive exchange-to-exchange
				// binding: show queues only below top-level exchange
				if len(visited) > 0 {
					continue
				}
				// queue to exchange binding
				queues := s.createQueueNodeFromBinding(binding, exchange, metadataService)
				for _, queue := range queues {
					exchangeNode.Add(queue)
				}
			}
		}
		return exchangeNode
	}
	return create(exchange, metadataService, binding)
}

func (s defaultBrokerInfoTreeBuilder) createRootNode(rootNodeURL *url.URL,
	overview *rabtap.RabbitOverview) *rootNode {
	b := baseNode{[]interface{}{}}
	return &rootNode{b, overview, rootNodeURL}
}

// buildTree renders given brokerInfo into a tree:
//  RabbitMQ-Host
//  +--VHost
//     +--Exchange
//        +--Queue bound to exchange
//           +--Connection (optional)
//              +--Channel
//                 +--Consumer
//
func (s defaultBrokerInfoTreeBuilder) buildTreeByExchange(
	rootNodeURL *url.URL,
	metadataService rabtap.MetadataService) (*rootNode, error) {

	overview := metadataService.Overview()
	rootNode := s.createRootNode(rootNodeURL, &overview)

	for _, vhost := range metadataService.Vhosts() {
		vhost := vhost
		vhostNode := newVhostNode(&vhost)
		for _, exchange := range metadataService.Exchanges() {
			exchange := exchange
			if !s.shouldDisplayExchange(&exchange, &vhost) {
				continue
			}
			exNode := s.createExchangeNode(&exchange, metadataService, nil)
			if s.config.OmitEmptyExchanges && !exNode.hasChildren() {
				continue
			}
			vhostNode.Add(exNode)
		}
		rootNode.Add(vhostNode)
	}
	return rootNode, nil
}

// buildTree renders given brokerInfo into a tree:
//  RabbitMQ-Host
//  +--VHost
//     +--Connection
//        +--Channel
//          +--Consumer (opt)
//             +--Queue
func (s defaultBrokerInfoTreeBuilder) buildTreeByConnection(
	rootNodeURL *url.URL,
	metadataService rabtap.MetadataService) (*rootNode, error) {

	overview := metadataService.Overview()
	rootNode := s.createRootNode(rootNodeURL, &overview)

	vhosts := map[string]*vhostNode{}
	for _, conn := range metadataService.Connections() {
		conn := conn
		vhostName := conn.Vhost
		var ok bool
		if _, ok = vhosts[vhostName]; !ok {
			vhost := metadataService.FindVhostByName(vhostName)
			vhosts[vhostName] = newVhostNode(vhost)
		}

		connNode := newConnectionNode(&conn, Valid)

		channels := metadataService.AllChannelsForConnection(vhostName, conn.Name)
		for _, channel := range channels {
			channel := channel

			params := map[string]interface{}{"connection": conn, "channel": channel}
			if res, err := s.config.QueueFilter.Eval(params); err != nil || !res {
				if err != nil {
					log.Warnf("error evaluating queue filter: %s", err)
				}
				continue
			}

			chanNode := newChannelNode(channel, Valid)

			consumers := metadataService.AllConsumersForChannel(vhostName, channel.Name)
			for _, consumer := range consumers {
				consumer := consumer
				consNode := newConsumerNode(consumer)
				if queue := metadataService.FindQueueByName(vhostName, consumer.Queue.Name); queue != nil {
					queueNode := newQueueNode(queue)
					consNode.Add(queueNode)
				}
				chanNode.Add(consNode)
			}
			// if s.config.OmitEmptyExchanges && !connNode.HasChildren() {
			//     continue
			// }
			connNode.Add(chanNode)
		}
		if connNode.hasChildren() {
			vhosts[vhostName].Add(connNode)
		}
	}

	for _, key := range orderedKeySet(vhosts) {
		rootNode.Add(vhosts[key])
	}
	return rootNode, nil
}

func (s brokerInfoTreeBuilderByConnection) BuildTree(
	rootNodeURL *url.URL,
	metadataService rabtap.MetadataService) (*rootNode, error) {

	builder := newDefaultBrokerInfoTreeBuilder(s.config)
	return builder.buildTreeByConnection(rootNodeURL, metadataService)
}

func (s brokerInfoTreeBuilderByExchange) BuildTree(
	rootNodeURL *url.URL,
	metadataService rabtap.MetadataService) (*rootNode, error) {

	builder := newDefaultBrokerInfoTreeBuilder(s.config)
	return builder.buildTreeByExchange(rootNodeURL, metadataService)
}
