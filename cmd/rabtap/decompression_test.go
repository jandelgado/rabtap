package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecompressFailsWithInvalidData(t *testing.T) {
	testcases := []struct {
		name string
	}{
		{"zstd"},
		{"deflate"},
		{"gzip"},
		{"bzip2"},
	}
	for _, tc := range testcases {
		t.Run(fmt.Sprintf("algorihmn %s", tc.name), func(t *testing.T) {
		})
		dec, err := NewDecompressor(tc.name)
		require.NoError(t, err)

		buf := []byte{1, 2, 3}
		_, err = dec(bytes.NewReader(buf))
		require.Error(t, err)
	}
}

func TestDecompressesValidData(t *testing.T) {
	testcases := []struct {
		alg     string
		probe    string
		expected string
	}{
		// echo "JAN"|zstd|xxd -p -c0
		{"zstd", "28b52ffd04582100004a414e0a0a21908f", "JAN\n"},
		// echo "JAN"|python3 -c "import sys, zlib; compressor = zlib.compressobj(wbits=-zlib.MAX_WBITS); sys.stdout.buffer.write(compressor.compress(sys.stdin.buffer.read()) + compressor.flush())" | xxd -p -c0
		{"deflate", "f372f4e30200", "JAN\n"},
		// echo "JAN"|gzip|xxd -p -c0
		{"gzip", "1f8b0800000000000003f372f4e30200270b9a2a04000000", "JAN\n"},
		// echo "JAN"|bzip2|xxd -p -c0
		{"bzip2", "425a6839314159265359dab9c92b0000014400001020112000219a68334d173c5dc914e142436ae724ac", "JAN\n"},
		{"identity", "4a414e", "JAN"},
	}
	for _, tc := range testcases {
		t.Run(fmt.Sprintf("algorithmn %s", tc.alg), func(t *testing.T) {
			dec, err := NewDecompressor(tc.alg)
			require.NoError(t, err)

			buf, err := hex.DecodeString(tc.probe)
			require.NoError(t, err)

			u, err := dec(bytes.NewReader(buf))
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(u))
		})
	}
}
