// Copyright (C) 2017 Jan Delgado

// +build integration

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

func TestGetTLSConfig(t *testing.T) {

	tls := getTLSConfig(true)
	assert.True(t, tls.InsecureSkipVerify)
	tls = getTLSConfig(false)
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
