package rabtap

import (
	"crypto/tls"
	"net/url"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	defaultHeartbeat = 10 * time.Second
	defaultLocale    = "en_US"
)

// DialTLS is a Wrapper for amqp.DialTLS that supports EXTERNAL auth for mtls
// can be removed when https://github.com/streadway/amqp/pull/121 gets some day
// merged.
func DialTLS(uri string, tlsConfig *tls.Config) (*amqp.Connection, error) {

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	// if client certificates are specified and no explicit credentials in the
	// amqp connect url are given, then request EXTERNAL auth.
	var sasl []amqp.Authentication
	if tlsConfig.Certificates != nil && u.User == nil {
		sasl = []amqp.Authentication{&amqp.ExternalAuth{}}
	}

	return amqp.DialConfig(uri, amqp.Config{
		Heartbeat:       defaultHeartbeat,
		TLSClientConfig: tlsConfig,
		Locale:          defaultLocale,
		SASL:            sasl,
		Dial:            Dialer,
	})
}
