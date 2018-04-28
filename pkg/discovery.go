// Copyright (C) 2017 Jan Delgado.

package rabtap

import (
	"errors"

	"github.com/streadway/amqp"
)

// DiscoverBindingsForExchange returns a string list of routing-keys that
// are used by the given exchange and broker. This list can be used to
// auto-tap to all queues on a given exchange
func DiscoverBindingsForExchange(rabbitAPIClient *RabbitHTTPClient, vhost, exchangeName string) ([]string, error) {

	var bindingKeys []string
	exchanges, err := rabbitAPIClient.Exchanges()

	if err != nil {
		return nil, err
	}

	// find type of given exchange
	var exchangeType *string
	for _, exchange := range exchanges {
		if exchange.Name == exchangeName && exchange.Vhost == vhost {
			exchangeType = &exchange.Type
			break
		}
	}

	if exchangeType == nil {
		return nil, errors.New("exchange " + exchangeName + " on vhost " +
			vhost + " not found")
	}

	switch *exchangeType {
	case amqp.ExchangeDirect:
		// filter out all bindings for given exchange
		bindings, err := rabbitAPIClient.Bindings()

		if err != nil {
			return nil, err
		}

		for _, binding := range bindings {
			if binding.Source == exchangeName {
				bindingKeys = append(bindingKeys, binding.RoutingKey)
			}
		}
		return bindingKeys, nil
	case amqp.ExchangeTopic:
		return []string{"#"}, nil
	case amqp.ExchangeFanout, amqp.ExchangeHeaders:
		return []string{""}, nil
	}

	return nil, errors.New("exchange '" + exchangeName + "': unknown type " +
		*exchangeType)
}
