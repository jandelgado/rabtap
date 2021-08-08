// Copyright (C) 2017 Jan Delgado

package main

import (
	"errors"
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
	tls := getTLSConfig(true, TLSCertFile, TLSKeyFile, TLSCaFile)
	assert.True(t, tls.InsecureSkipVerify)
	tls = getTLSConfig(false, TLSCertFile, TLSKeyFile, TLSCaFile)
	assert.False(t, tls.InsecureSkipVerify)
}

func TestFailOnError(t *testing.T) {

	exitFuncCalled := false
	exitFunc := func(int) {
		exitFuncCalled = true
	}

	// error case
	failOnError(errors.New("error"), "error test", exitFunc)
	assert.True(t, exitFuncCalled)

	// no error case
	exitFuncCalled = false
	failOnError(nil, "test", exitFunc)
	assert.False(t, exitFuncCalled)

}
