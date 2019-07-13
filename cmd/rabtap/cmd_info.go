// Copyright (C) 2017 Jan Delgado

package main

import (
	"io"
	"os"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// CmdInfoArg contains arguments for the info command
type CmdInfoArg struct {
	rootNode     string
	client       *rabtap.RabbitHTTPClient
	treeConfig   BrokerInfoPrinterConfig
	renderConfig BrokerInfoRendererConfig
	out          io.Writer
}

// cmdInfo queries the rabbitMQ brokers REST api and dispays infos
// on exchanges, queues, bindings etc in a human readably fashion.
// TODO proper error handling
func cmdInfo(cmd CmdInfoArg) {
	brokerInfo, err := cmd.client.BrokerInfo()
	failOnError(err, "failed retrieving info from rabbitmq REST api", os.Exit)

	treeBuilder, err := NewBrokerInfoTreeBuilder(cmd.treeConfig)
	failOnError(err, "failed instanciating tree builder", os.Exit)
	renderer := NewBrokerInfoRenderer(cmd.renderConfig)

	tree, _ := treeBuilder.BuildTree(cmd.rootNode, brokerInfo)

	//	renderer := NewBrokerInfoRendererText(s.config)
	//renderer := brokerInfoRenderers[s.config.ContentType](s.config)
	//renderer := NewBrokerInfoRenderer(rendererConfig)
	renderer.Render(tree, os.Stdout)
	// brokerInfoPrinter := NewBrokerInfoPrinter(cmd.printConfig)
	// brokerInfoPrinter.Print(brokerInfo, cmd.rootNode, cmd.out)
}
