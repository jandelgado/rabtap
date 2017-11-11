// Copyright (C) 2017 Jan Delgado

package main

import "github.com/jandelgado/rabtap"

// BrokerInfo collects information of an RabbitMQ broker
type BrokerInfo struct {
	Exchanges []rabtap.RabbitExchange
	Queues    []rabtap.RabbitQueue
	Bindings  []rabtap.RabbitBinding
	Consumers []rabtap.RabbitConsumer
}

// NewBrokerInfo obtains infos on broker using the provided client object
func NewBrokerInfo(client *rabtap.RabbitHTTPClient) (BrokerInfo, error) {

	var err error
	var s BrokerInfo

	// collect infos from rabtap.RabbitMQ API
	s.Exchanges, err = client.GetExchanges()
	if err != nil {
		return s, err
	}

	s.Bindings, err = client.GetBindings()
	if err != nil {
		return s, err
	}

	s.Queues, err = client.GetQueues()
	if err != nil {
		return s, err
	}

	s.Consumers, err = client.GetConsumers()
	if err != nil {
		return s, err
	}

	return s, nil
}
