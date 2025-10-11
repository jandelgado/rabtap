// Copyright (C) 2017 Jan Delgado

package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitLogging(t *testing.T) {
	logger := initLogging(os.Stderr, false, false)
	assert.NotNil(t, logger)
	logger = initLogging(os.Stderr, true, false)
	assert.NotNil(t, logger)
}

func TestDefaultFilenameProviderReturnsFilenameInExpectedFormat(t *testing.T) {
	fn := defaultFilenameProvider()
	assert.Regexp(t, "^rabtap-[0-9]+$", fn)
}

func TestGetTLSConfig(t *testing.T) {
	var TLSCertFile string
	var TLSKeyFile string
	var TLSCaFile string
	tls, err := getTLSConfig(true, TLSCertFile, TLSKeyFile, TLSCaFile)
	assert.NoError(t, err)
	assert.True(t, tls.InsecureSkipVerify)

	tls, err = getTLSConfig(false, TLSCertFile, TLSKeyFile, TLSCaFile)
	assert.NoError(t, err)
	assert.False(t, tls.InsecureSkipVerify)
}
