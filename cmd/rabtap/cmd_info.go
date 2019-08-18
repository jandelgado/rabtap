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
	treeConfig   BrokerInfoTreeBuilderConfig
	renderConfig BrokerInfoRendererConfig
	out          io.Writer
}

// cmdInfo queries the rabbitMQ brokers REST api and dispays infos
// on exchanges, queues, bindings etc in a human readably fashion.
// TODO proper error handling
func cmdInfo(cmd CmdInfoArg) {
	brokerInfo, err := cmd.client.BrokerInfo()
	failOnError(err, "failed retrieving info from rabbitmq REST api", os.Exit)

	treeBuilder := NewBrokerInfoTreeBuilder(cmd.treeConfig)
	failOnError(err, "failed instanciating tree builder", os.Exit)
	renderer := NewBrokerInfoRenderer(cmd.renderConfig)

	tree, _ := treeBuilder.BuildTree(cmd.rootNode, brokerInfo)
	renderer.Render(tree, os.Stdout)
}
