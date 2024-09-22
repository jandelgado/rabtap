package main

import (
	"bytes"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Body returns the message Body, uncompressing if necessary
func Body(m *amqp.Delivery) ([]byte, error) {
	// currently we only expect a single encoding in the header
	if enc := m.ContentEncoding; enc != "" {
		dec, err := NewDecompressor(enc)
		if err != nil {
			return nil, fmt.Errorf("decompress: %w", err)
		}
		return dec(bytes.NewReader(m.Body))
	}
	return m.Body, nil
}
