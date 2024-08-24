package main

import (
	"bytes"
	"compress/gzip"
	"io"

	amqp "github.com/rabbitmq/amqp091-go"
)

// body returns the message body, uncompressing if necessary
func body(m *amqp.Delivery) ([]byte, error) {
	if m.ContentEncoding == "gzip" {
		return gunzip(bytes.NewReader(m.Body))
	}
	return m.Body, nil
}

func gunzip(r io.Reader) ([]byte, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return io.ReadAll(zr)
}
