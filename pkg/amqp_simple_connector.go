package rabtap

import (
	"crypto/tls"

	"github.com/streadway/amqp"
)

// openAMQPChannel tries to open a channel on the given broker
func openAMQPChannel(uri string, tlsConfig *tls.Config) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.DialTLS(uri, tlsConfig)
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
func SimpleAmqpConnector(amqpURI string, tlsConfig *tls.Config,
	run func(*amqp.Channel) error) error {
	conn, chn, err := openAMQPChannel(amqpURI, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()
	return run(chn)
}
