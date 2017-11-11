// Copyright (C) 2017 Jan Delgado

package main

// the command line is the UI of the tool. so we test it carefully.

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCliTapNoBinding(t *testing.T) {
	// EXCHANGE has not required format of "exchange:key[,exchange:key]*"
	_, err := ParseCommandLineArgs([]string{"tap", "exchange"})
	assert.NotNil(t, err)
}

func TestCliTapSingleUri(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"tap", "--uri=broker1", "exchange1:binding1"})

	assert.Nil(t, err)
	assert.Equal(t, TapMode, args.Mode)
	assert.Equal(t, 1, len(args.TapConfig))
	assert.Equal(t, "broker1", args.TapConfig[0].AmqpURI)
	assert.Equal(t, 1, len(args.TapConfig[0].Exchanges))
	assert.Equal(t, "exchange1", args.TapConfig[0].Exchanges[0].Exchange)
	assert.Equal(t, "binding1", args.TapConfig[0].Exchanges[0].BindingKey)
	assert.False(t, args.NoColor)
}

func TestCliTapSingleUriFromEnv(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	os.Setenv(key, "URI")
	defer os.Unsetenv(key)

	args, err := ParseCommandLineArgs(
		[]string{"tap", "exchange1:binding1"})

	assert.Nil(t, err)
	assert.Equal(t, TapMode, args.Mode)
	assert.Equal(t, 1, len(args.TapConfig))
	assert.Equal(t, "URI", args.TapConfig[0].AmqpURI)
	assert.Equal(t, 1, len(args.TapConfig[0].Exchanges))
	assert.Equal(t, "exchange1", args.TapConfig[0].Exchanges[0].Exchange)
	assert.Equal(t, "binding1", args.TapConfig[0].Exchanges[0].BindingKey)
	assert.False(t, args.NoColor)
}

func TestCliTapSingleMissingUri(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	defer os.Unsetenv(key)
	_, err := ParseCommandLineArgs(
		[]string{"tap", "exchange1:binding1"})
	assert.NotNil(t, err)
}
func TestCliTapWithMultipleUris(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"tap", "--uri=broker1", "exchange1:binding1,exchange2:binding2",
			"tap", "--uri=broker2", "exchange3:binding3,exchange4:binding4",
			"--saveto", "savedir", "--no-color"})

	assert.Nil(t, err)
	assert.Equal(t, TapMode, args.Mode)
	assert.Equal(t, 2, len(args.TapConfig))
	assert.Equal(t, "broker1", args.TapConfig[0].AmqpURI)
	assert.Equal(t, "broker2", args.TapConfig[1].AmqpURI)
	assert.Equal(t, 2, len(args.TapConfig[0].Exchanges))
	assert.Equal(t, "exchange1", args.TapConfig[0].Exchanges[0].Exchange)
	assert.Equal(t, "binding1", args.TapConfig[0].Exchanges[0].BindingKey)
	assert.Equal(t, "exchange2", args.TapConfig[0].Exchanges[1].Exchange)
	assert.Equal(t, "binding2", args.TapConfig[0].Exchanges[1].BindingKey)
	assert.Equal(t, 2, len(args.TapConfig[1].Exchanges))
	assert.Equal(t, "exchange3", args.TapConfig[1].Exchanges[0].Exchange)
	assert.Equal(t, "binding3", args.TapConfig[1].Exchanges[0].BindingKey)
	assert.Equal(t, "exchange4", args.TapConfig[1].Exchanges[1].Exchange)
	assert.Equal(t, "binding4", args.TapConfig[1].Exchanges[1].BindingKey)
	assert.Equal(t, "savedir", *args.SaveDir)
	assert.True(t, args.NoColor)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
}

func TestCliInsecureLongOpt(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"tap", "--uri=broker", "exchange:binding",
			"--insecure"})

	assert.Nil(t, err)
	assert.Equal(t, 1, len(args.TapConfig))
	assert.Equal(t, TapMode, args.Mode)
	assert.Equal(t, "broker", args.TapConfig[0].AmqpURI)
	assert.Equal(t, 1, len(args.TapConfig[0].Exchanges))
	assert.Equal(t, "exchange", args.TapConfig[0].Exchanges[0].Exchange)
	assert.Equal(t, "binding", args.TapConfig[0].Exchanges[0].BindingKey)
	assert.Nil(t, args.SaveDir)
	assert.False(t, args.JSONFormat)
	assert.True(t, args.InsecureTLS)
}

func TestCliInsecureOpt(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"tap", "--uri=broker", "exchange:binding", "--insecure"})

	assert.Nil(t, err)
	assert.Equal(t, 1, len(args.TapConfig))
	assert.Equal(t, TapMode, args.Mode)
	assert.Equal(t, "broker", args.TapConfig[0].AmqpURI)
	assert.Equal(t, 1, len(args.TapConfig[0].Exchanges))
	assert.Equal(t, "exchange", args.TapConfig[0].Exchanges[0].Exchange)
	assert.Equal(t, "binding", args.TapConfig[0].Exchanges[0].BindingKey)
	assert.Nil(t, args.SaveDir)
	assert.False(t, args.JSONFormat)
	assert.True(t, args.InsecureTLS)
}

func TestCliVerboseOpt(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"tap", "--uri=broker", "exchange:binding", "-v", "--json"})

	assert.Nil(t, err)
	assert.Equal(t, 1, len(args.TapConfig))
	assert.Equal(t, TapMode, args.Mode)
	assert.Equal(t, "broker", args.TapConfig[0].AmqpURI)
	assert.Equal(t, 1, len(args.TapConfig[0].Exchanges))
	assert.Equal(t, "exchange", args.TapConfig[0].Exchanges[0].Exchange)
	assert.Equal(t, "binding", args.TapConfig[0].Exchanges[0].BindingKey)
	assert.True(t, args.JSONFormat)
	assert.Nil(t, args.SaveDir)
	assert.True(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
}

func TestCliInfoMode(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"info", "--api=APIURI"})

	assert.Nil(t, err)
	assert.Equal(t, 0, len(args.TapConfig))
	assert.Equal(t, InfoMode, args.Mode)
	assert.Equal(t, "APIURI", args.APIURI)
	assert.False(t, args.Verbose)
	assert.False(t, args.ShowStats)
	assert.False(t, args.ShowConsumers)
	assert.False(t, args.InsecureTLS)
	assert.False(t, args.NoColor)
}

func TestCliInfoModeMissingApi(t *testing.T) {
	const key = "RABTAP_APIURI"
	os.Unsetenv(key)
	_, err := ParseCommandLineArgs(
		[]string{"info"})
	assert.NotNil(t, err)
}
func TestCliInfoModeApiFromEnv(t *testing.T) {
	const key = "RABTAP_APIURI"
	os.Setenv(key, "APIURI")
	defer os.Unsetenv(key)

	args, err := ParseCommandLineArgs(
		[]string{"info"})

	assert.Nil(t, err)
	assert.Equal(t, 0, len(args.TapConfig))
	assert.Equal(t, InfoMode, args.Mode)
	assert.Equal(t, "APIURI", args.APIURI)
	assert.False(t, args.Verbose)
	assert.False(t, args.ShowConsumers)
	assert.False(t, args.InsecureTLS)
	assert.False(t, args.NoColor)
}

func TestCliInfoModeShowConsumersWithoutColor(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"info", "--api=APIURI", "--stats", "--consumers", "--no-color"})

	assert.Nil(t, err)
	assert.Equal(t, 0, len(args.TapConfig))
	assert.Equal(t, InfoMode, args.Mode)
	assert.Equal(t, "APIURI", args.APIURI)
	assert.Nil(t, args.SaveDir)
	assert.False(t, args.Verbose)
	assert.True(t, args.ShowStats)
	assert.True(t, args.ShowConsumers)
	assert.False(t, args.ShowDefaultExchange)
	assert.True(t, args.NoColor)
	assert.False(t, args.InsecureTLS)
}

func TestCliInfoModeShowDefault(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"info", "--api=APIURI", "--show-default"})

	assert.Nil(t, err)
	assert.Equal(t, 0, len(args.TapConfig))
	assert.Equal(t, InfoMode, args.Mode)
	assert.Equal(t, "APIURI", args.APIURI)
	assert.False(t, args.Verbose)
	assert.False(t, args.ShowConsumers)
	assert.True(t, args.ShowDefaultExchange)
	assert.False(t, args.InsecureTLS)
}

func TestCliSendModeFromFile(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"send", "--uri=broker", "exchange", "file"})

	assert.Nil(t, err)
	assert.Equal(t, SendMode, args.Mode)
	assert.Equal(t, "broker", args.SendAmqpURI)
	assert.Equal(t, "exchange", args.SendExchange)
	assert.Equal(t, "file", *args.SendFile)
	assert.Equal(t, "", args.SendRoutingKey)
	assert.False(t, args.JSONFormat)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
}

func TestCliSendModeUriFromEnv(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	os.Setenv(key, "URI")
	defer os.Unsetenv(key)
	args, err := ParseCommandLineArgs(
		[]string{"send", "exchange"})

	assert.Nil(t, err)
	assert.Equal(t, SendMode, args.Mode)
	assert.Equal(t, "URI", args.SendAmqpURI)
	assert.Equal(t, "exchange", args.SendExchange)
}

func TestCliSendModeMissingUri(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	os.Unsetenv(key)
	_, err := ParseCommandLineArgs(
		[]string{"send", "exchange"})
	assert.NotNil(t, err)
}
func TestCliSendModeFromStdinWithRoutingKey(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"send", "--uri=broker1", "exchange1", "--routingkey=key", "--json"})

	assert.Nil(t, err)
	assert.Equal(t, SendMode, args.Mode)
	assert.Equal(t, "broker1", args.SendAmqpURI)
	assert.Equal(t, "exchange1", args.SendExchange)
	assert.Nil(t, args.SendFile)
	assert.True(t, args.JSONFormat)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
}
