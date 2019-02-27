package main

import (
	"crypto/tls"
	"fmt"
	"os"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

func cmdConnClose(apiURI, connName, reason string, tlsConfig *tls.Config) error {
	client := rabtap.NewRabbitHTTPClient(apiURI, tlsConfig)
	err := client.CloseConnection(connName, reason)
	failOnError(err, fmt.Sprintf("close connection '%s'", connName), os.Exit)
	return err
}
