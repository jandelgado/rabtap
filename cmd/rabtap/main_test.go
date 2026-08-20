// Copyright (C) 2017 Jan Delgado

package main

import (
	"errors"
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

func TestFormatCommandLineErrorReturnsNoMessageWhenNoArgsGiven(t *testing.T) {
	// docopt already prints the usage/help text itself in this case, so no
	// additional error message must be shown to the user.
	msg := formatCommandLineError([]string{}, errors.New(""))
	assert.Equal(t, "", msg)
}

func TestFormatCommandLineErrorReturnsMessageWhenInvalidArgsGiven(t *testing.T) {
	msg := formatCommandLineError([]string{"foo"}, errors.New(""))
	assert.Contains(t, msg, "invalid command or arguments")

	msg = formatCommandLineError([]string{"queue"}, errors.New("some error"))
	assert.Contains(t, msg, "some error")
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
