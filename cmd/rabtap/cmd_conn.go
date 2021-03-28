package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

func cmdConnClose(ctx context.Context, apiURL, connName, reason string, tlsConfig *tls.Config) error {
	url, err := url.Parse(apiURL)
	if err != nil {
		return err
	}
	client := rabtap.NewRabbitHTTPClient(url, tlsConfig)
	err = client.CloseConnection(ctx, connName, reason)
	failOnError(err, fmt.Sprintf("close connection '%s'", connName), os.Exit)
	return err
}
