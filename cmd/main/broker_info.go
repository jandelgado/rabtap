// Copyright (C) 2017 Jan Delgado

package main

import "github.com/jandelgado/rabtap/pkg"

// BrokerInfo collects information of an RabbitMQ broker
type BrokerInfo struct {
	Overview    rabtap.RabbitOverview
	Exchanges   []rabtap.RabbitExchange
	Queues      []rabtap.RabbitQueue
	Bindings    []rabtap.RabbitBinding
	Connections []rabtap.RabbitConnection
	Consumers   []rabtap.RabbitConsumer
	//	Channels    []rabtap.RabbitChannel  // not yet used.
}

// NewBrokerInfo obtains infos on broker using the provided client object
func NewBrokerInfo(client *rabtap.RabbitHTTPClient) (BrokerInfo, error) {

	var err error
	var bi BrokerInfo

	// collect infos from rabtap.RabbitMQ API
	bi.Overview, err = client.Overview()
	if err != nil {
		return bi, err
	}

	bi.Exchanges, err = client.Exchanges()
	if err != nil {
		return bi, err
	}

	bi.Bindings, err = client.Bindings()
	if err != nil {
		return bi, err
	}

	bi.Queues, err = client.Queues()
	if err != nil {
		return bi, err
	}

	bi.Connections, err = client.Connections()
	if err != nil {
		return bi, err
	}

	bi.Consumers, err = client.Consumers()
	if err != nil {
		return bi, err
	}

	// bi.Channels, err = client.Channels()
	// if err != nil {
	//     return bi, err
	// }

	return bi, nil
}
