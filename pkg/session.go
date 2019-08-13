package rabtap

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// taken from streadways amqplib examples

// Session composes an amqp.Connection with an amqp.Channel
type Session struct {
	*amqp.Connection
	*amqp.Channel
}

// Close tears the connection down, taking the channel with it.
// func (s Session) Close() error {
//     if s.Connection == nil {
//         return nil
//     }
//     return s.Connection.Close()
// }

// NewChannel opens a new Channel on the connection. Call when current
// got closed due to errors.
func (s *Session) NewChannel() error {
	if s.Channel != nil {
		s.Channel.Close()
	}
	ch, err := s.Connection.Channel()
	s.Channel = ch
	return err
}

// redial continually connects to the URL, exiting the program when no longer possible
func redial(ctx context.Context, url string, tlsConfig *tls.Config) chan chan Session {
	sessions := make(chan chan Session)

	go func() {
		initial := true
		sess := make(chan Session)
		defer close(sessions)

		for {
			select {
			case sessions <- sess:
			case <-ctx.Done():
				log.Println("shutting down session factory")
				close(sess)
				return
			}

			// try to connect. fail early if initial connection cant be
			// established.
			var conn *amqp.Connection
			var err error
			for {
				conn, err = amqp.DialTLS(url, tlsConfig)
				if err == nil {
					break
				}
				log.Printf("cannot (re)dial: %v: %q", err, url)
				if initial {
					log.Printf("initial connection failed")
					close(sess)
					return
				}
				select {
				case <-ctx.Done():
					log.Println("shutting down session factory")
					close(sess)
					return
				case <-time.After(2 * time.Second):
				}
			}

			initial = false
			log.Printf("connected to %s", url)

			ch, err := conn.Channel()
			if err != nil {
				log.Printf("cannot create channel: %v", err)
			}

			select {
			case sess <- Session{conn, ch}:
			case <-ctx.Done():
				log.Println("shutting down new session")
				close(sess)
				return
			}
		}
	}()

	return sessions
}
