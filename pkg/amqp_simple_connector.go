package rabtap

import (
	"crypto/tls"
	"net/url"

	amqp "github.com/rabbitmq/amqp091-go"
)

// openAMQPChannel tries to open a channel on the given broker
func openAMQPChannel(u *url.URL, tlsConfig *tls.Config) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := DialTLS(u, tlsConfig)
	if err != nil {
		return nil, nil, err
	}
	chn, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
	}
	return conn, chn, err
}

// SimpleAmqpConnector opens an AMQP connection and channel, and calls
// a function with the channel as argument. Use this function for simple,
// one-shot operations like creation of queues, exchanges etc.
func SimpleAmqpConnector(amqpURL *url.URL, tlsConfig *tls.Config,
	run func(session Session) error,
) error {
	conn, chn, err := openAMQPChannel(amqpURL, tlsConfig)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	return run(Session{conn, chn})
}
