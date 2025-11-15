// Copyright (C) 2022 Jan Delgado

//go:build integration

package rabtap

import (
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"os"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

// relative location of certificates used during the integration tests
const certDir = "../inttest/pki/certs/"

func TestDialTLSConnectsToNonTLSEndpoint(t *testing.T) {
	u, _ := url.Parse("amqp://guest:password@localhost:5672")
	conn, err := DialTLS(u, &tls.Config{})
	assert.NoError(t, err)
	conn.Close()
}

func TestDialTLSConnectsToTLSEndpoint(t *testing.T) {
	testcases := []struct {
		certFile, keyFile string
		url               string
		err               error
	}{
		// credentials in URL will force PLAIN auth
		{"unknown.crt", "unknown.key", "amqps://guest:password@localhost:5671", nil},
		{
			"unknown.crt", "unknown.key", "amqps://invalid:pass@localhost:5671",
			&amqp.Error{Code: 403, Reason: "username or password not allowed"},
		},

		// client cert with unknown user in RabbitMQ will not proceed
		{
			"unknown.crt", "unknown.key", "amqps://localhost:5671",
			&amqp.Error{Code: 403, Reason: "username or password not allowed"},
		},

		// client cert with known user in RabbitMQ will proceed
		{"testuser.crt", "testuser.key", "amqps://localhost:5671", nil},

		// client cert with known user in RabbitMQ but unknown credentials will not proceed
		{
			"testuser.crt", "testuser.key", "amqps://invalid:pass@localhost:5671",
			&amqp.Error{Code: 403, Reason: "username or password not allowed"},
		},
	}

	for _, tc := range testcases {
		// given
		tlsConfig := &tls.Config{}
		caCert, err := os.ReadFile(certDir + "ca.crt")
		assert.NoError(t, err)
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool

		cert, err := tls.LoadX509KeyPair(certDir+tc.certFile, certDir+tc.keyFile)
		assert.NoError(t, err)
		tlsConfig.Certificates = []tls.Certificate{cert}

		u, _ := url.Parse(tc.url)

		// when
		conn, err := DialTLS(u, tlsConfig)

		// then
		assert.Equal(t, tc.err, err)

		conn.Close()
	}
}
