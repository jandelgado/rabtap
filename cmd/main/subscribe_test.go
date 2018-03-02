// Copyright (C) 2017 Jan Delgado

package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
)

func TestCreateMessageReceiveFuncRawToFile(t *testing.T) {
	testDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(testDir)

	out := os.Stdout
	rcvFunc := createMessageReceiveFunc(out, false, &testDir, false)
	_ = rcvFunc(&amqp.Delivery{})

	// TODO
}

func TestCreateMessageReceiveFuncJSONToFile(t *testing.T) {
	// testDir, err := ioutil.TempDir("", "")
	// require.Nil(t, err)
	// defer os.RemoveAll(testDir)

	// out := os.Stdout
	// rcvFunc := createMessageReceiveFunc(out, true, &testDir, false)
	// err = rcvFunc(&amqp.Delivery{})
}
