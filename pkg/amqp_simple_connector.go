package rabtap

import (
	"crypto/tls"
	"net/url"

	"github.com/streadway/amqp"
)

// openAMQPChannel tries to open a channel on the given broker
func openAMQPChannel(uri *url.URL, tlsConfig *tls.Config) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := DialTLS(uri.String(), tlsConfig)
	if err != nil {
		return nil, nil, err
	}
	chn, err := conn.Channel()
	if err != nil {
		conn.Close()
	}
	return conn, chn, err
}

// SimpleAmqpConnector opens an AMQP connection and channel, and calls
// a function with the channel as argument. Use this function for simple,
// one-shot operations like creation of queues, exchanges etc.
func SimpleAmqpConnector(amqpURI *url.URL, tlsConfig *tls.Config,
	run func(session Session) error) error {
	conn, chn, err := openAMQPChannel(amqpURI, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()
	return run(Session{conn, chn})
}
