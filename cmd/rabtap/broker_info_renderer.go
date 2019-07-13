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
	ContentType string // "text", "dot"
	ShowStats   bool
	NoColor     bool
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
func RegisterBrokerInfoRenderer(contentType string, newFunc brokerInfoRendererNewFunc) {
	brokerInfoRenderers[contentType] = newFunc
}

func NewBrokerInfoRenderer(config BrokerInfoRendererConfig) BrokerInfoRenderer {
	newFunc := brokerInfoRenderers[config.ContentType]
	return newFunc(config)
}
