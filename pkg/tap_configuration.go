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

// unescapeStr reutrns a string with all '\' characters removed from the
// given string
func unescapeStr(s string) string {
	res := ""
	inEscape := false
	for _, c := range s {
		if c == '\\' && !inEscape {
			inEscape = true
		} else {
			inEscape = false
			res += string(c)
		}
	}
	return res
}

// splitExchangeAndBinding splits a string of the form "exchange:binding"
// and return exchange and binding as string. A colon can be escaped with
// \: if it is part of the exchange or binding string, e.g.
// splitExchangeAndBinding("ex\\:change:binding") -> ("ex:change", "binding", nil)
func splitExchangeAndBinding(exchangeAndBinding string) (string, string, error) {
	inEscape := false
	pos := -1
	for i, c := range exchangeAndBinding {
		if c == ':' && !inEscape {
			pos = i
			break
		}
		inEscape = (c == '\\')
	}
	if pos == -1 {
		return "", "", errors.New("expected format `exchange`:`binding`, but got `" +
			exchangeAndBinding + "`")
	}
	return unescapeStr(exchangeAndBinding[:pos]), unescapeStr(exchangeAndBinding[pos+1:]), nil
}

// NewExchangeConfiguration returns a pointer to a newly created
// ExchangeConfiguration object
func NewExchangeConfiguration(exchangeAndBindingStr string) (*ExchangeConfiguration, error) {
	exchange, binding, err := splitExchangeAndBinding(exchangeAndBindingStr)
	if err != nil {
		return nil, err
	}
	return &ExchangeConfiguration{exchange, binding}, nil
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
