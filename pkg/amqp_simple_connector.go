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
	run func(session Session) error) error {
	conn, chn, err := openAMQPChannel(amqpURI, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()
	// Wait for channel close in case of errors,e.g. unbind of nonexistent
	// queue. But seems to not produce an error. So leave it out for now.
	// errCh := make(chan *amqp.Error)
	// chn.NotifyClose(errCh)
	err = run(Session{conn, chn})

	if err != nil {
		return err
	}
	// Wait for channel close in case of errors - see above
	// select {
	// case amqpErr := <-errCh:
	//     return errors.New(amqpErr.Reason)
	// case <-time.After(1 * time.Second):
	// }
	return nil
}
