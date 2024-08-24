package main

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGunzipDecompressesValidBuffer(t *testing.T) {
	// echo "JAN" | gzip | hexdump -X
	buf, err := hex.DecodeString("1f8b0800000000000003f372f4e30200270b9a2a04000000")
	require.NoError(t, err)
	u, err := gunzip(bytes.NewReader(buf))
	require.NoError(t, err)
	assert.Equal(t, "JAN\n", string(u))
}
func TestGunzipFailsWithInvalidData(t *testing.T) {
	buf := []byte{1, 2, 3}
	_, err := gunzip(bytes.NewReader(buf))
	require.Error(t, err)
}
