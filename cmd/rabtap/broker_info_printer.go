// Copyright (C) 2017-2019 Jan Delgado
// construct & render broker infos as returned by the RabbitMQ API

package main

import (
	"io"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// PrintBrokerInfo renders given brokerInfo into something usable
func PrintBrokerInfo(treeBuilder BrokerInfoTreeBuilder,
	renderer BrokerInfoRenderer,
	rootNodeURL string,
	brokerInfo rabtap.BrokerInfo,
	out io.Writer) error {

	//builder := NewBrokerInfoTreeBuilder("byExchange")
	tree, err := treeBuilder.BuildTree(rootNodeURL, brokerInfo)

	// if s.config.ShowByConnection {
	//     root, err = s.buildTreeByConnection(rootNodeURL, brokerInfo)
	// } else {
	//     root, err = s.buildTreeByExchange(rootNodeURL, brokerInfo)
	// }

	if err != nil {
		return err
	}

	//	renderer := NewBrokerInfoRendererText(s.config)
	//renderer := brokerInfoRenderers[s.config.ContentType](s.config)
	//renderer := NewBrokerInfoRenderer(rendererConfig)
	return renderer.Render(tree, out)
}
