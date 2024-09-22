package main

import (
	"encoding/hex"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBodyDecompressesACompressedBody(t *testing.T) {
	// given
	buf, err := hex.DecodeString("f372f4e30200") // JAN\n
	require.NoError(t, err)
	d := amqp.Delivery{Body: buf, ContentEncoding: "deflate"}

	// when
	buf, err = Body(&d)

	// then
	require.NoError(t, err)
	assert.Equal(t, "JAN\n", string(buf))
}

func TestBodyFailsWithUnknownEncoding(t *testing.T) {
	// given
	d := amqp.Delivery{ContentEncoding: "invalid"}

	// when
	_, err := Body(&d)

	// then
	assert.ErrorContains(t, err, "decompress: unsupported encoding")
}

func TestBodyReturnsTheBodyAsIsWithoutAnEncoding(t *testing.T) {
	// given
	d := amqp.Delivery{Body: []byte("JAN")}

	// when
	buf, err := Body(&d)

	// then
	require.NoError(t, err)
	assert.Equal(t, "JAN", string(buf))
}
