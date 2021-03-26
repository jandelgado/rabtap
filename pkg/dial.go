package rabtap

import (
	"crypto/tls"
	"time"

	"github.com/streadway/amqp"
)

const (
	defaultHeartbeat = 10 * time.Second
	defaultLocale    = "en_US"
)

// DialTLS is a Wrapper for amqp.DialTLS that supports EXTERNAL auth for mtls
// can be removed when https://github.com/streadway/amqp/pull/121 gets some day
// merged.
func DialTLS(url string, amqps *tls.Config) (*amqp.Connection, error) {
	var sasl []amqp.Authentication

	if amqps.Certificates != nil {
		// client certificate are set to we must use EXTERNAL auth
		sasl = []amqp.Authentication{&ExternalAuth{}}
	}

	return amqp.DialConfig(url, amqp.Config{
		Heartbeat:       defaultHeartbeat,
		TLSClientConfig: amqps,
		Locale:          defaultLocale,
		SASL:            sasl,
	})
}
