// Copyright (C) 2017 Jan Delgado

package main

// the command line is the UI of the tool. so we test it carefully.

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	os.Unsetenv("RABTAP_AMQPURI")
	os.Unsetenv("RABTAP_APIURI")
	code := m.Run()
	os.Exit(code)
}

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

func TestParseTLSConfig(t *testing.T) {
	args := map[string]interface{}{
		"--tls-cert-file": []string{"/tmp/tls-cert.pem"},
		"--tls-key-file": []string{"/tmp/tls-key.pem"},
		"--tls-ca-file": []string{"/tmp/tls-ca.pem"}
	}
	commonArgs, err := parseCommonArgs(args)
	assert.Nil(t, err)
	assert.Equal(t, "/tmp/tls-cert.pem", commonArgs.TLSCertFile)
	assert.Equal(t, "/tmp/tls-key.pem", commonArgs.TLSKeyFile)
	assert.Equal(t, "/tmp/tls-ca.pem", commonArgs.TLSCaFile)
}

func TestParseTLSConfigUseEnvironment(t *testing.T) {
	const key1 = "RABTAP_TLS_CERTFILE"
	const key2 = "RABTAP_TLS_KEYFILE"
	const key3 = "RABTAP_TLS_CAFILE"
	os.Setenv(key1, "/tmp/tls-cert.pem")
	os.Setenv(key2, "/tmp/tls-key.pem")
	os.Setenv(key3, "/tmp/tls-ca.pem")
	defer os.Unsetenv(key1)
	defer os.Unsetenv(key2)
	defer os.Unsetenv(key3)
	args := map[string]interface{}{
		"--tls-cert-file": []string{},
		"--tls-key-file": []string{},
		"--tls-ca-file": []string{}
	}
	commonArgs, err := parseCommonArgs(args)
	assert.Nil(t, err)
	assert.Equal(t, "/tmp/tls-cert.pem", commonArgs.TLSCertFile)
	assert.Equal(t, "/tmp/tls-key.pem", commonArgs.TLSKeyFile)
	assert.Equal(t, "/tmp/tls-ca.pem", commonArgs.TLSCaFile)
}

func TestParseCommandLineArgsFailsWithInvalidSpec(t *testing.T) {
	_, err := parseCommandLineArgsWithSpec("invalid spec", []string{"invalid"})
	assert.NotNil(t, err)
}

func TestCliTapSingleUri(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"tap", "--uri=broker1", "exchange1:binding1"})

	assert.Nil(t, err)
	assert.Equal(t, TapCmd, args.Cmd)
	assert.False(t, args.Silent)
	assert.Equal(t, 1, len(args.TapConfig))
	assert.Equal(t, "broker1", args.TapConfig[0].AmqpURI)
	assert.Equal(t, 1, len(args.TapConfig[0].Exchanges))
	assert.Equal(t, "exchange1", args.TapConfig[0].Exchanges[0].Exchange)
	assert.Equal(t, "binding1", args.TapConfig[0].Exchanges[0].BindingKey)
	assert.False(t, args.NoColor)
}

func TestParsePubSubFormatArgDefaultsToRaw(t *testing.T) {
	fmt, err := parsePubSubFormatArg(map[string]interface{}{})
	assert.Nil(t, err)
	assert.Equal(t, "raw", fmt)
}

func TestParsePubSubFormatArgDetectsValidOptions(t *testing.T) {
	fmt, err := parsePubSubFormatArg(map[string]interface{}{"--format": "json"})
	assert.Nil(t, err)
	assert.Equal(t, "json", fmt)
}

func TestParsePubSubFormatArgDetectsDeprecatedOptions(t *testing.T) {
	fmt, err := parsePubSubFormatArg(map[string]interface{}{"--json": true})
	assert.Nil(t, err)
	assert.Equal(t, "json", fmt)
}

func TestParsePubSubFormatArgRaisesErrorForInvalidOption(t *testing.T) {
	_, err := parsePubSubFormatArg(map[string]interface{}{"--format": "invalid"})
	assert.NotNil(t, err)
}

func TestCliTapSingleUriFromEnv(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	os.Setenv(key, "URI")
	defer os.Unsetenv(key)

	args, err := ParseCommandLineArgs(
		[]string{"tap", "exchange1:binding1", "--silent"})

	assert.Nil(t, err)
	assert.Equal(t, TapCmd, args.Cmd)
	assert.True(t, args.Silent)
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
	assert.False(t, args.Silent)
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
	assert.Nil(t, args.TLSCertFile)
	assert.Nil(t, args.TLSKeyFile)
	assert.Nil(t, args.TLSCaFile)
}

func TestCliAllOptsInTapCommand(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"tap", "--uri=broker", "exchange:binding", "--silent", "--verbose",
			"--format=json-nopp", "--insecure"})

	assert.Nil(t, err)
	assert.Equal(t, 1, len(args.TapConfig))
	assert.Equal(t, TapCmd, args.Cmd)
	assert.Equal(t, "broker", args.TapConfig[0].AmqpURI)
	assert.Equal(t, 1, len(args.TapConfig[0].Exchanges))
	assert.Equal(t, "exchange", args.TapConfig[0].Exchanges[0].Exchange)
	assert.Equal(t, "binding", args.TapConfig[0].Exchanges[0].BindingKey)
	assert.Nil(t, args.SaveDir)
	assert.Equal(t, "json-nopp", args.Format)
	assert.True(t, args.Verbose)
	assert.True(t, args.Silent)
	assert.True(t, args.InsecureTLS)
	assert.Nil(t, args.TLSCertFile)
	assert.Nil(t, args.TLSKeyFile)
	assert.Nil(t, args.TLSCaFile)
}

func TestCliInfoCmd(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"info", "--api=APIURI"})

	assert.Nil(t, err)
	assert.Equal(t, 0, len(args.TapConfig))
	assert.Equal(t, InfoCmd, args.Cmd)
	assert.Equal(t, "APIURI", args.APIURI)
	assert.Equal(t, "true", args.QueueFilter)
	assert.False(t, args.Verbose)
	assert.False(t, args.ShowStats)
	assert.False(t, args.ShowConsumers)
	assert.Equal(t, "byExchange", args.InfoMode)
	assert.Equal(t, "text", args.Format)
	assert.False(t, args.InsecureTLS)
	assert.Nil(t, args.TLSCertFile)
	assert.Nil(t, args.TLSKeyFile)
	assert.Nil(t, args.TLSCaFile)
	assert.False(t, args.NoColor)
	assert.False(t, args.OmitEmptyExchanges)
}

func TestCliInfoCmdShowByConnection(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"info", "--api=uri", "--mode=byConnection"})

	assert.Nil(t, err)
	assert.Equal(t, InfoCmd, args.Cmd)
	assert.Equal(t, "byConnection", args.InfoMode)
}

func TestCliInfoCmdOutputAsDotFile(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"info", "--api=uri", "--format=dot"})

	assert.Nil(t, err)
	assert.Equal(t, InfoCmd, args.Cmd)
	assert.Equal(t, "dot", args.Format)
}

func TestCliInfoCmdFailsWithInvalidMode(t *testing.T) {
	_, err := ParseCommandLineArgs(
		[]string{"info", "--api=uri", "--mode=INVALID"})

	assert.NotNil(t, err)
}

func TestCliInfoCmdFailsWithInvalidFormat(t *testing.T) {
	_, err := ParseCommandLineArgs(
		[]string{"info", "--api=uri", "--format=INVALID"})

	assert.NotNil(t, err)
}

func TestCliInfoCmdMissingApi(t *testing.T) {
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
	assert.Nil(t, args.TLSCertFile)
	assert.Nil(t, args.TLSKeyFile)
	assert.Nil(t, args.TLSCaFile)
	assert.False(t, args.NoColor)
	assert.Equal(t, "byExchange", args.InfoMode)
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
	assert.Equal(t, "EXPR", args.QueueFilter)
	assert.True(t, args.ShowStats)
	assert.True(t, args.ShowConsumers)
	assert.True(t, args.ShowDefaultExchange)
	assert.True(t, args.NoColor)
	assert.True(t, args.InsecureTLS)
	assert.Nil(t, args.TLSCertFile)
	assert.Nil(t, args.TLSKeyFile)
	assert.Nil(t, args.TLSCaFile)
	assert.True(t, args.OmitEmptyExchanges)
}

func TestCliPubCmdFromFileMinimalOptsSet(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"pub", "--uri=broker"})

	assert.Nil(t, err)
	assert.Equal(t, PubCmd, args.Cmd)
	assert.Equal(t, "broker", args.AmqpURI)
	assert.Nil(t, args.PubExchange)
	assert.Nil(t, args.Source)
	assert.Nil(t, args.PubRoutingKey)
	assert.Equal(t, "raw", args.Format)
	assert.Nil(t, args.Delay)
	assert.Equal(t, 1., args.Speed)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
	assert.Nil(t, args.TLSCertFile)
	assert.Nil(t, args.TLSKeyFile)
	assert.Nil(t, args.TLSCaFile)
}
func TestCliPubCmdFromFileAllOptsSet(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"pub", "--uri=broker", "--exchange=exchange", "file",
			"--routingkey=key", "--delay=5s", "--format=json"})

	assert.Nil(t, err)
	assert.Equal(t, PubCmd, args.Cmd)
	assert.Equal(t, "broker", args.AmqpURI)
	assert.Equal(t, "exchange", *args.PubExchange)
	assert.Equal(t, "file", *args.Source)
	assert.Equal(t, "key", *args.PubRoutingKey)
	assert.Equal(t, "json", args.Format)
	assert.Equal(t, 5*time.Second, *args.Delay)
	assert.Equal(t, 1., args.Speed)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
	assert.Nil(t, args.TLSCertFile)
	assert.Nil(t, args.TLSKeyFile)
	assert.Nil(t, args.TLSCaFile)
}

func TestCliPubCmdUriFromEnv(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	os.Setenv(key, "URI")
	defer os.Unsetenv(key)
	args, err := ParseCommandLineArgs(
		[]string{"pub"})

	assert.Nil(t, err)
	assert.Equal(t, PubCmd, args.Cmd)
	assert.Equal(t, "URI", args.AmqpURI)
}

func TestCliPubCmdMissingUriReturnsError(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	os.Unsetenv(key)
	_, err := ParseCommandLineArgs([]string{"pub"})
	assert.NotNil(t, err)
}

func TestCliPubCmdInvalidDelayReturnsError(t *testing.T) {
	_, err := ParseCommandLineArgs([]string{"pub", "--uri=uri", "--delay=invalid"})
	assert.NotNil(t, err)
}

func TestCliPubCmdInvalidSpeedupReturnsError(t *testing.T) {
	_, err := ParseCommandLineArgs([]string{"pub", "--uri=uri", "--speed=invalid"})
	assert.NotNil(t, err)
}

func TestCliPubCmdInvalidFormatReturnsError(t *testing.T) {
	_, err := ParseCommandLineArgs([]string{"pub", "--uri=uri", "--format=invalid"})
	assert.NotNil(t, err)
}

func TestCliPubCmdFromStdinWithRoutingKeyJsonFormat(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"pub", "--uri=broker1", "--exchange=exchange1", "--routingkey=key", "--format=json"})

	assert.Nil(t, err)
	assert.Equal(t, PubCmd, args.Cmd)
	assert.Equal(t, "broker1", args.AmqpURI)
	assert.Equal(t, "exchange1", *args.PubExchange)
	assert.Equal(t, "key", *args.PubRoutingKey)
	assert.Nil(t, args.Source)
	assert.Equal(t, "json", args.Format)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
	assert.Nil(t, args.TLSCertFile)
	assert.Nil(t, args.TLSKeyFile)
	assert.Nil(t, args.TLSCaFile)
}

func TestCliPubCmdFromStdinWithJsonFormatDeprecated(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"pub", "--uri=broker1", "--json"})

	assert.Nil(t, err)
	assert.Equal(t, PubCmd, args.Cmd)
	assert.Equal(t, "broker1", args.AmqpURI)
	assert.Nil(t, args.PubExchange)
	assert.Nil(t, args.PubRoutingKey)
	assert.Nil(t, args.Source)
	assert.Equal(t, "json", args.Format)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
	assert.Nil(t, args.TLSCertFile)
	assert.Nil(t, args.TLSKeyFile)
	assert.Nil(t, args.TLSCaFile)
}

func TestCliSubCmdInvalidFormatReturnsError(t *testing.T) {
	_, err := ParseCommandLineArgs([]string{"sub", "queue", "--uri=uri", "--format=invalid"})
	assert.NotNil(t, err)
}

func TestCliSubCmdSaveToDir(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"sub", "queuename", "--uri", "uri", "--saveto", "dir"})

	assert.Nil(t, err)
	assert.Equal(t, SubCmd, args.Cmd)
	assert.Equal(t, "queuename", args.QueueName)
	assert.Equal(t, "uri", args.AmqpURI)
	assert.Equal(t, "dir", *args.SaveDir)
	assert.True(t, args.AutoAck)
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

func TestCliPurgeQueue(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "purge", "name", "--uri", "uri"})

	assert.Nil(t, err)
	assert.Equal(t, QueuePurgeCmd, args.Cmd)
	assert.Equal(t, "name", args.QueueName)
	assert.Equal(t, "uri", args.AmqpURI)
}

func TestCliUnbindQueue(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "unbind", "queuename", "from", "exchangename",
			"--bindingkey", "key", "--uri", "uri"})

	assert.Nil(t, err)
	assert.Equal(t, QueueUnbindCmd, args.Cmd)
	assert.Equal(t, "queuename", args.QueueName)
	assert.Equal(t, "exchangename", args.ExchangeName)
	assert.Equal(t, "key", args.QueueBindingKey)
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
