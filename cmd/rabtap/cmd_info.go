// Copyright (C) 2017 Jan Delgado

package main

import (
	"context"
	"io"
	"net/url"
	"os"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// CmdInfoArg contains arguments for the info command
type CmdInfoArg struct {
	rootNode     *url.URL
	client       *rabtap.RabbitHTTPClient
	treeConfig   BrokerInfoTreeBuilderConfig
	renderConfig BrokerInfoRendererConfig
	out          io.Writer
}

// cmdInfo queries the rabbitMQ brokers REST api and dispays infos
// on exchanges, queues, bindings etc in a human readably fashion.
// TODO proper error handling
func cmdInfo(ctx context.Context, cmd CmdInfoArg) {
	brokerInfo, err := cmd.client.BrokerInfo(ctx)
	failOnError(err, "failed retrieving info from rabbitmq REST api", os.Exit)

	treeBuilder := NewBrokerInfoTreeBuilder(cmd.treeConfig)
	failOnError(err, "failed instanciating tree builder", os.Exit)
	renderer := NewBrokerInfoRenderer(cmd.renderConfig)

	tree, _ := treeBuilder.BuildTree(cmd.rootNode, brokerInfo)
	failOnError(renderer.Render(tree, os.Stdout), "rendering failed", os.Exit)
}
