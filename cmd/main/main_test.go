// Copyright (C) 2017 Jan Delgado

package main

import (
	"errors"
	"testing"

	"github.com/jandelgado/rabtap/pkg/testcommon"
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

func ExamplestartCmdInfo() {
	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeEmpty)
	defer mock.Close()

	args := CommandLineArgs{APIURI: mock.URL, commonArgs: commonArgs{NoColor: true}}
	startCmdInfo(args, "http://rootnode")

	// Output:
	// http://rootnode (broker ver=3.6.9, mgmt ver=3.6.9, cluster=rabbit@08f57d1fe8ab)
}
