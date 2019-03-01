// Copyright (C) 2017 Jan Delgado

package main

import (
	"os"
	"regexp"
	"testing"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
)

func TestCmdInfoRootNodeOnly(t *testing.T) {

	// REST api mock returning "empty" broker
	apiMock := testcommon.NewRabbitAPIMock(testcommon.MockModeEmpty)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"rabtap", "info",
		"--api", apiMock.URL,
		"--no-color"}
	out := testcommon.CaptureOutput(main)
	assert.Regexp(t, regexp.MustCompile("http://(.*) \\(broker ver='3.6.9', mgmt ver='3.6.9', cluster='rabbit@08f57d1fe8ab'\\)"), out)
}
