package main

import (
	"context"
	"crypto/tls"
	"net/url"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

func cmdConnClose(ctx context.Context, apiURL *url.URL, connName, reason string, tlsConfig *tls.Config) error {
	client := rabtap.NewRabbitHTTPClient(apiURL, tlsConfig)
	return client.CloseConnection(ctx, connName, reason)
}
