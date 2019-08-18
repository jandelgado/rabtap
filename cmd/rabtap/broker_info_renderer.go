// Copyright (C) 2017-2019 Jan Delgado
// rendering of abstract broker info tree into concrete representations, e.g.
// as text, dot etc.

package main

import (
	"io"
)

// BrokerInfoRendererConfig holds configuration for a renderer. At the
// moment, all renderers share the same config.
type BrokerInfoRendererConfig struct {
	Format    string // "text", "dot"
	ShowStats bool
	NoColor   bool
}

// BrokerInfoRenderer renders a tree representation represented by a rootNode
// into a concrete representation. Implementations render the tree into text,
// dot etc. representations.
type BrokerInfoRenderer interface {
	// Render rendrs the given tree to the provided io.Writer
	Render(rootNode *rootNode, out io.Writer) error
}

type brokerInfoRendererNewFunc func(BrokerInfoRendererConfig) BrokerInfoRenderer

// Registry of available broker info renderers
var brokerInfoRenderers = map[string]brokerInfoRendererNewFunc{}

// RegisterBrokerInfoRenderer registers a new broker info renderer
func RegisterBrokerInfoRenderer(format string, newFunc brokerInfoRendererNewFunc) {
	brokerInfoRenderers[format] = newFunc
}

// NewBrokerInfoRenderer returns a new BrokerInfoRenderer for the given format
// config.Format that is initialized with the given config.
func NewBrokerInfoRenderer(config BrokerInfoRendererConfig) BrokerInfoRenderer {
	newFunc := brokerInfoRenderers[config.Format]
	return newFunc(config)
}
