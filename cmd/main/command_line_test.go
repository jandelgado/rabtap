// Copyright (C) 2017 Jan Delgado

package main

// the command line is the UI of the tool. so we test it carefully.

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAmqpURI(t *testing.T) {
	// since multple --uri arguments are possible docopt returns an array
	args := map[string]interface{}{"--uri": []string{"URI"}}
	uri, err := parseAmqpURI(args)
	assert.Nil(t, err)
	assert.Equal(t, "URI", uri)
}

func TestParseAmqpURINotSet(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	os.Unsetenv(key)
	args := map[string]interface{}{"--uri": []string{}}
	_, err := parseAmqpURI(args)
	assert.NotNil(t, err)
}

func TestParseAmqpURIUseEnvironment(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	os.Setenv(key, "URI")
	defer os.Unsetenv(key)
	args := map[string]interface{}{"--uri": []string{}}
	uri, err := parseAmqpURI(args)
	assert.Nil(t, err)
	assert.Equal(t, "URI", uri)
}

func TestCliTapSingleUri(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"tap", "--uri=broker1", "exchange1:binding1"})

	assert.Nil(t, err)
	assert.Equal(t, TapCmd, args.Cmd)
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
	assert.Equal(t, TapCmd, args.Cmd)
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
			"--saveto", "savedir"})

	assert.Nil(t, err)
	assert.Equal(t, TapCmd, args.Cmd)
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
	assert.False(t, args.NoColor)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
}

func TestCliInsecureLongOpt(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"tap", "--uri=broker", "exchange:binding",
			"--insecure"})

	assert.Nil(t, err)
	assert.Equal(t, 1, len(args.TapConfig))
	assert.Equal(t, TapCmd, args.Cmd)
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
	assert.Equal(t, TapCmd, args.Cmd)
	assert.Equal(t, "broker", args.TapConfig[0].AmqpURI)
	assert.Equal(t, 1, len(args.TapConfig[0].Exchanges))
	assert.Equal(t, "exchange", args.TapConfig[0].Exchanges[0].Exchange)
	assert.Equal(t, "binding", args.TapConfig[0].Exchanges[0].BindingKey)
	assert.True(t, args.JSONFormat)
	assert.Nil(t, args.SaveDir)
	assert.True(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
}

func TestCliInfoCmd(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"info", "--api=APIURI"})

	assert.Nil(t, err)
	assert.Equal(t, 0, len(args.TapConfig))
	assert.Equal(t, InfoCmd, args.Cmd)
	assert.Equal(t, "APIURI", args.APIURI)
	assert.False(t, args.Verbose)
	assert.False(t, args.ShowStats)
	assert.False(t, args.ShowConsumers)
	assert.False(t, args.InsecureTLS)
	assert.False(t, args.NoColor)
	assert.Nil(t, args.QueueFilter)
	assert.False(t, args.OmitEmptyExchanges)
}

func TestCliInfoCmdMissingApi(t *testing.T) {
	const key = "RABTAP_APIURI"
	os.Unsetenv(key)
	_, err := ParseCommandLineArgs(
		[]string{"info"})
	assert.NotNil(t, err)
}

func TestCliInfoCmdApiFromEnv(t *testing.T) {
	const key = "RABTAP_APIURI"
	os.Setenv(key, "APIURI")
	defer os.Unsetenv(key)

	args, err := ParseCommandLineArgs(
		[]string{"info"})

	assert.Nil(t, err)
	assert.Equal(t, 0, len(args.TapConfig))
	assert.Equal(t, InfoCmd, args.Cmd)
	assert.Equal(t, "APIURI", args.APIURI)
	assert.False(t, args.Verbose)
	assert.False(t, args.ShowConsumers)
	assert.False(t, args.InsecureTLS)
	assert.False(t, args.NoColor)
}

func TestCliInfoCmdAllOptionsAreSet(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"info", "--api=APIURI", "--stats", "--consumers",
			"--filter=EXPR", "--omit-empty",
			"--no-color", "-k", "--show-default"})

	assert.Nil(t, err)
	assert.Equal(t, 0, len(args.TapConfig))
	assert.Equal(t, InfoCmd, args.Cmd)
	assert.Equal(t, "APIURI", args.APIURI)
	assert.Nil(t, args.SaveDir)
	assert.False(t, args.Verbose)
	assert.Equal(t, "EXPR", *args.QueueFilter)
	assert.True(t, args.ShowStats)
	assert.True(t, args.ShowConsumers)
	assert.True(t, args.ShowDefaultExchange)
	assert.True(t, args.NoColor)
	assert.True(t, args.InsecureTLS)
	assert.True(t, args.OmitEmptyExchanges)
}

func TestCliPubCmdFromFile(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"pub", "--uri=broker", "exchange", "file"})

	assert.Nil(t, err)
	assert.Equal(t, PubCmd, args.Cmd)
	assert.Equal(t, "broker", args.AmqpURI)
	assert.Equal(t, "exchange", args.PubExchange)
	assert.Equal(t, "file", *args.PubFile)
	assert.Equal(t, "", args.PubRoutingKey)
	assert.False(t, args.JSONFormat)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
}

func TestCliPubCmdUriFromEnv(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	os.Setenv(key, "URI")
	defer os.Unsetenv(key)
	args, err := ParseCommandLineArgs(
		[]string{"pub", "exchange"})

	assert.Nil(t, err)
	assert.Equal(t, PubCmd, args.Cmd)
	assert.Equal(t, "URI", args.AmqpURI)
	assert.Equal(t, "exchange", args.PubExchange)
}

func TestCliPubCmdMissingUri(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	os.Unsetenv(key)
	_, err := ParseCommandLineArgs(
		[]string{"pub", "exchange"})
	assert.NotNil(t, err)
}

func TestCliPubCmdFromStdinWithRoutingKey(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"pub", "--uri=broker1", "exchange1", "--routingkey=key", "--json"})

	assert.Nil(t, err)
	assert.Equal(t, PubCmd, args.Cmd)
	assert.Equal(t, "broker1", args.AmqpURI)
	assert.Equal(t, "exchange1", args.PubExchange)
	assert.Nil(t, args.PubFile)
	assert.True(t, args.JSONFormat)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
}

func TestCliSubCmd(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"sub", "queuename", "--uri", "uri"})

	assert.Nil(t, err)
	assert.Equal(t, SubCmd, args.Cmd)
	assert.Equal(t, "queuename", args.QueueName)
	assert.Equal(t, "uri", args.AmqpURI)
}

func TestCliCreateQueue(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "create", "name", "--uri", "uri"})

	assert.Nil(t, err)
	assert.Equal(t, QueueCreateCmd, args.Cmd)
	assert.Equal(t, "name", args.QueueName)
	assert.Equal(t, "uri", args.AmqpURI)
	assert.False(t, args.Durable)
	assert.False(t, args.Autodelete)
}

func TestCliCreateDurableAutodeleteQueue(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "create", "name", "--uri", "uri",
			"--durable", "--autodelete"})

	assert.Nil(t, err)
	assert.Equal(t, QueueCreateCmd, args.Cmd)
	assert.True(t, args.Durable)
	assert.True(t, args.Autodelete)
}

func TestCliRemoveQueue(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "rm", "name", "--uri", "uri"})

	assert.Nil(t, err)
	assert.Equal(t, QueueRemoveCmd, args.Cmd)
	assert.Equal(t, "name", args.QueueName)
	assert.Equal(t, "uri", args.AmqpURI)
}

func TestCliBindQueue(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "bind", "queuename", "to", "exchangename",
			"--bindingkey", "key", "--uri", "uri"})

	assert.Nil(t, err)
	assert.Equal(t, QueueBindCmd, args.Cmd)
	assert.Equal(t, "queuename", args.QueueName)
	assert.Equal(t, "exchangename", args.ExchangeName)
	assert.Equal(t, "key", args.QueueBindingKey)
	assert.Equal(t, "uri", args.AmqpURI)
}

func TestCliCreateExchange(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"exchange", "create", "name", "--type", "topic",
			"--uri", "uri"})

	assert.Nil(t, err)
	assert.Equal(t, ExchangeCreateCmd, args.Cmd)
	assert.Equal(t, "name", args.ExchangeName)
	assert.Equal(t, "topic", args.ExchangeType)
	assert.Equal(t, "uri", args.AmqpURI)
	assert.False(t, args.Durable)
	assert.False(t, args.Autodelete)
}

func TestCliCreateDurableAutodeleteExchange(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"exchange", "create", "name", "--type", "topic",
			"--uri", "uri", "--durable", "--autodelete"})

	assert.Nil(t, err)
	assert.Equal(t, ExchangeCreateCmd, args.Cmd)
	assert.True(t, args.Durable)
	assert.True(t, args.Autodelete)
}

func TestCliRemoveExchange(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"exchange", "rm", "name", "--uri", "uri"})

	assert.Nil(t, err)
	assert.Equal(t, ExchangeRemoveCmd, args.Cmd)
	assert.Equal(t, "name", args.ExchangeName)
	assert.Equal(t, "uri", args.AmqpURI)
}

func TestCliCloseConnectionWithDefaultReason(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"conn", "close", "conn-name", "--api", "uri"})

	assert.Nil(t, err)
	assert.Equal(t, ConnCloseCmd, args.Cmd)
	assert.Equal(t, "uri", args.APIURI)
	assert.Equal(t, "conn-name", args.ConnName)
	assert.Equal(t, "closed by rabtap", args.CloseReason)
}

func TestCliCloseConnection(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"conn", "close", "conn-name", "--api", "uri",
			"--reason", "reason"})

	assert.Nil(t, err)
	assert.Equal(t, ConnCloseCmd, args.Cmd)
	assert.Equal(t, "uri", args.APIURI)
	assert.Equal(t, "conn-name", args.ConnName)
	assert.Equal(t, "reason", args.CloseReason)
}

func TestParseNoColorFromEnvironment(t *testing.T) {
	const key = "NO_COLOR"
	os.Setenv(key, "1")
	defer os.Unsetenv(key)

	args, err := ParseCommandLineArgs(
		[]string{"info", "--api=APIURI"})

	assert.Nil(t, err)
	assert.Equal(t, InfoCmd, args.Cmd)
	assert.Equal(t, "APIURI", args.APIURI)
	assert.True(t, args.NoColor)
}
