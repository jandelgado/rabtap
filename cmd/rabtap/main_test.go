// Copyright (C) 2017 Jan Delgado

package main

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestInitLogging(t *testing.T) {
	initLogging(false)
	assert.Equal(t, logrus.WarnLevel, log.Level)
	initLogging(true)
	assert.Equal(t, logrus.DebugLevel, log.Level)
	initLogging(false)
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
