package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

func cmdConnClose(ctx context.Context, apiURL *url.URL, connName, reason string, tlsConfig *tls.Config) error {
	client := rabtap.NewRabbitHTTPClient(apiURL, tlsConfig)
	err := client.CloseConnection(ctx, connName, reason)
	failOnError(err, fmt.Sprintf("close connection '%s'", connName), os.Exit)
	return err
}
