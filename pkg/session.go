package rabtap

import (
	"context"
	"crypto/tls"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// taken from streadways amqplib examples

const (
	retryDelay = 3 * time.Second
	FailEarly  = true
)

// Session composes an amqp.Connection with an amqp.Channel
type Session struct {
	*amqp.Connection
	*amqp.Channel
}

// NewChannel opens a new Channel on the connection. Call when current
// got closed due to errors.
func (s *Session) NewChannel() error {
	if s.Channel != nil {
		_ = s.Channel.Close()
	}
	ch, err := s.Connection.Channel()
	s.Channel = ch
	return err
}

// redial continually connects to the URL and provides a AMQP connection and
// channel in a Session struct. Closes returned chan when initial connection
// attempt fails.
func redial(ctx context.Context, url string, tlsConfig *tls.Config,
	logger Logger, failEarly bool) chan chan Session {

	sessions := make(chan chan Session)

	go func() {
		sess := make(chan Session)
		defer close(sessions)

		for {
			select {
			case sessions <- sess:
			case <-ctx.Done():
				logger.Debugf("session: shutting down factory (cancel)")
				close(sess)
				return
			}

			// try to connect. fail early if initial connection cant be
			// established.
			var conn *amqp.Connection
			var ch *amqp.Channel
			var err error
			for {
				conn, err = DialTLS(url, tlsConfig)
				if err == nil {
					ch, err = conn.Channel()
					if err == nil {
						break
					}
				}
				logger.Errorf("session: cannot (re-)dial: %v: %q", err, url)
				if failEarly {
					close(sess)
					return
				}
				select {
				case <-ctx.Done():
					logger.Debugf("session: shutting down factory (cancel)")
					close(sess)
					return
				case <-time.After(retryDelay):
				}
			}

			failEarly = false

			select {
			case sess <- Session{conn, ch}:
			case <-ctx.Done():
				logger.Debugf("session: shutting down factory (cancel)")
				close(sess)
				return
			}
		}
	}()

	return sessions
}
