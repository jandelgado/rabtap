// Copyright (C) 2017 Jan Delgado

package main

import (
	"context"
	"fmt"
	"io"
	"net/url"

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
func cmdInfo(ctx context.Context, cmd CmdInfoArg) error {
	brokerInfo, err := cmd.client.BrokerInfo(ctx)
	if err != nil {
		return fmt.Errorf("retrieving info from rabbitmq REST API: %w", err)
	}

	treeBuilder := NewBrokerInfoTreeBuilder(cmd.treeConfig)
	renderer := NewBrokerInfoRenderer(cmd.renderConfig)

	metadataService := rabtap.NewInMemoryMetadataService(brokerInfo)
	tree, err := treeBuilder.BuildTree(cmd.rootNode, metadataService)
	if err != nil {
		return fmt.Errorf("building info tree: %w", err)
	}
	if err := renderer.Render(tree, cmd.out); err != nil {
		return fmt.Errorf("render info tree: %w", err)
	}
	return nil
}
