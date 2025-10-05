package main

import (
	"compress/bzip2"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"strings"

	"github.com/klauspost/compress/zstd"
)

type DecompressionFunc func(r io.Reader) ([]byte, error)

// NewDecompressor returns a decompression function according to the given
// algorithmn
func NewDecompressor(alg string) (DecompressionFunc, error) {
	switch strings.ToLower(alg) {
	case "gzip":
		return decompressGunzip, nil
	case "zstd":
		return decompressZstd, nil
	case "bzip2":
		return decompressBunzip2, nil
	case "deflate":
		return decompressDeflate, nil
	case "identity":
		return func(r io.Reader) ([]byte, error) { return io.ReadAll(r) }, nil
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", alg)
	}
}

func decompressZstd(r io.Reader) ([]byte, error) {
	decoder, err := zstd.NewReader(r)
	if err != nil {
		return nil, err
	}
    cr := decoder.IOReadCloser()
    defer func() {_  = cr.Close()}()
	return io.ReadAll(cr)
}

func decompressDeflate(r io.Reader) ([]byte, error) {
	// compress/zlib is for zlib-formatted data (DEFLATE data with zlib header).
	// compress/flate is for raw DEFLATE data without any headers.
	cr := flate.NewReader(r)
    defer func() {_  = cr.Close()}()
	return io.ReadAll(cr)
}

func decompressBunzip2(r io.Reader) ([]byte, error) {
	cr := bzip2.NewReader(r)
	return io.ReadAll(cr)
}

func decompressGunzip(r io.Reader) ([]byte, error) {
	cr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
    defer func() {_  = cr.Close()}()
	return io.ReadAll(cr)
}
