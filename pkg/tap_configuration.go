// Copyright (C) 2017 Jan Delgado

package rabtap

import (
	"errors"
	"strings"
)

// ExchangeConfiguration holds exchange and bindingkey for a single tap
// TODO add vhost?
type ExchangeConfiguration struct {
	// Exchange is the name of the exchange to bind to
	Exchange string
	// BindingKey is the binding key to use. The key depends on the the type
	// of exchange being tapped (e.g. direct, topic).
	BindingKey string
}

// NewExchangeConfiguration returns a pointer to a newly created
// ExchangeConfiguration object
func NewExchangeConfiguration(exchangeAndBindingStr string) (*ExchangeConfiguration, error) {
	exchangeAndBinding := strings.Split(exchangeAndBindingStr, ":")
	if len(exchangeAndBinding) != 2 {
		return nil, errors.New("expected format `exchange`:`binding`, but got `" +
			exchangeAndBindingStr + "`")
	}
	return &ExchangeConfiguration{exchangeAndBinding[0], exchangeAndBinding[1]}, nil
}

// TapConfiguration holds the set of ExchangeCOnfigurations to tap to for a
// single RabbitMQ host
type TapConfiguration struct {
	AmqpURI   string
	Exchanges []ExchangeConfiguration
}

// NewTapConfiguration returns a TapConfiguration object for a an rabbitMQ
// broker specified by an URI and a list of exchanges and bindings in the
// form of "exchange:binding,exchange:binding). Returns configuration object
// or an error if parsing failed.
func NewTapConfiguration(amqpURI string, exchangesAndBindings string) (*TapConfiguration, error) {
	result := TapConfiguration{}
	result.AmqpURI = amqpURI
	for _, item := range strings.Split(exchangesAndBindings, ",") {
		exchangeConfig, err := NewExchangeConfiguration(item)
		if err != nil {
			return nil, err
		}
		result.Exchanges = append(result.Exchanges, *exchangeConfig)
	}
	return &result, nil
}
