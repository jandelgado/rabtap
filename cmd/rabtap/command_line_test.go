// Copyright (C) 2017 Jan Delgado

package main

// the command line is the UI of the tool. so we test it carefully.

import (
	"net/url"
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

// helper to match a string against an *url.URL
func assertEqualURL(t *testing.T, expected string, actual *url.URL) {
	expectedURI, err := url.Parse(expected)
	assert.NoError(t, err)
	assert.Equal(t, expectedURI, actual)
}

func TestParseAMQPURLParsesValidURI(t *testing.T) {
	// since multple --uri arguments are possible docopt returns an array
	args := map[string]interface{}{"--uri": []string{"uri"}}

	uri, err := parseAMQPURL(args)
	assert.Nil(t, err)

	assertEqualURL(t, "uri", uri)
}

func TestParseAMQPURLFailsIfNotSet(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	os.Unsetenv(key)
	args := map[string]interface{}{"--uri": []string{}}

	_, err := parseAMQPURL(args)

	assert.NotNil(t, err)
}

func TestParseAMQPURLTakesURIFromEnvironmentWhenNotSpecified(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	os.Setenv(key, "uri")
	defer os.Unsetenv(key)
	args := map[string]interface{}{"--uri": []string{}}

	uri, err := parseAMQPURL(args)

	assert.Nil(t, err)
	assertEqualURL(t, "uri", uri)
}

func TestParseTLSConfig(t *testing.T) {
	args := map[string]interface{}{
		"--tls-cert-file": "/tmp/tls-cert.pem",
		"--tls-key-file":  "/tmp/tls-key.pem",
		"--tls-ca-file":   "/tmp/tls-ca.pem",
		"--verbose":       false,
		"--insecure":      false,
		"--no-color":      false,
	}
	commonArgs := parseCommonArgs(args)
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
		"--tls-cert-file": nil,
		"--tls-key-file":  nil,
		"--tls-ca-file":   nil,
		"--verbose":       false,
		"--insecure":      false,
		"--no-color":      false,
	}
	commonArgs := parseCommonArgs(args)
	assert.Equal(t, "/tmp/tls-cert.pem", commonArgs.TLSCertFile)
	assert.Equal(t, "/tmp/tls-key.pem", commonArgs.TLSKeyFile)
	assert.Equal(t, "/tmp/tls-ca.pem", commonArgs.TLSCaFile)
}

func TestParseCommandLineArgsFailsWithInvalidSpec(t *testing.T) {
	_, err := parseCommandLineArgsWithSpec("invalid spec", []string{"invalid"})
	assert.NotNil(t, err)
}

func TestCliTapParsesSingleURL(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"tap", "--uri=uri", "exchange1:binding1"})

	assert.Nil(t, err)
	assert.Equal(t, TapCmd, args.Cmd)
	assert.False(t, args.Silent)
	assert.Equal(t, 1, len(args.TapConfig))
	assertEqualURL(t, "uri", args.TapConfig[0].AMQPURL)
	assert.Equal(t, 1, len(args.TapConfig[0].Exchanges))
	assert.Equal(t, "exchange1", args.TapConfig[0].Exchanges[0].Exchange)
	assert.Equal(t, "binding1", args.TapConfig[0].Exchanges[0].BindingKey)
	assert.False(t, args.NoColor)
}

func TestCliTapUsesURLFromEnvWhenNotSpecified(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	os.Setenv(key, "uri")
	defer os.Unsetenv(key)

	args, err := ParseCommandLineArgs(
		[]string{"tap", "exchange1:binding1"})

	assert.Nil(t, err)
	assert.Equal(t, TapCmd, args.Cmd)
	assert.Equal(t, 1, len(args.TapConfig))
	assertEqualURL(t, "uri", args.TapConfig[0].AMQPURL)
	assert.Equal(t, 1, len(args.TapConfig[0].Exchanges))
	assert.Equal(t, "exchange1", args.TapConfig[0].Exchanges[0].Exchange)
	assert.Equal(t, "binding1", args.TapConfig[0].Exchanges[0].BindingKey)
	assert.False(t, args.NoColor)
}

func TestCliTapFailsWhenNoURIisSpecified(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	defer os.Unsetenv(key)

	_, err := ParseCommandLineArgs(
		[]string{"tap", "exchange1:binding1"})

	assert.NotNil(t, err)
}

func TestCliTapParsesMultipleURLs(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"tap", "--uri=broker1", "exchange1:binding1,exchange2:binding2",
			"tap", "--uri=broker2", "exchange3:binding3,exchange4:binding4"})

	assert.Nil(t, err)
	assert.Equal(t, TapCmd, args.Cmd)
	assert.Equal(t, 2, len(args.TapConfig))
	assertEqualURL(t, "broker1", args.TapConfig[0].AMQPURL)
	assertEqualURL(t, "broker2", args.TapConfig[1].AMQPURL)
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
}

func TestCliAllOptsInTapCommandiAreRecognized(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"tap", "--uri=uri", "exchange:binding", "--silent", "--verbose",
			"--format=json-nopp", "--insecure", "--saveto", "savedir"})

	assert.Nil(t, err)
	assert.Equal(t, 1, len(args.TapConfig))
	assert.Equal(t, TapCmd, args.Cmd)
	assertEqualURL(t, "uri", args.TapConfig[0].AMQPURL)
	assert.Equal(t, 1, len(args.TapConfig[0].Exchanges))
	assert.Equal(t, "exchange", args.TapConfig[0].Exchanges[0].Exchange)
	assert.Equal(t, "binding", args.TapConfig[0].Exchanges[0].BindingKey)
	assert.Equal(t, "savedir", *args.SaveDir)
	assert.Equal(t, "json-nopp", args.Format)
	assert.True(t, args.Verbose)
	assert.True(t, args.Silent)
	assert.True(t, args.InsecureTLS)
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

func TestCliInfoCmdIsParsed(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"info", "--api=APIURI"})

	assert.Nil(t, err)
	assert.Equal(t, 0, len(args.TapConfig))
	assert.Equal(t, InfoCmd, args.Cmd)
	assertEqualURL(t, "APIURI", args.APIURL)
	assert.Equal(t, "true", args.QueueFilter)
	assert.False(t, args.Verbose)
	assert.False(t, args.ShowStats)
	assert.False(t, args.ShowConsumers)
	assert.Equal(t, "byExchange", args.InfoMode)
	assert.Equal(t, "text", args.Format)
	assert.False(t, args.InsecureTLS)
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

func TestCliInfoCmdApiTakesURIFromEnv(t *testing.T) {
	const key = "RABTAP_APIURI"
	os.Setenv(key, "APIURI")
	defer os.Unsetenv(key)

	args, err := ParseCommandLineArgs(
		[]string{"info"})

	assert.Nil(t, err)
	assert.Equal(t, 0, len(args.TapConfig))
	assert.Equal(t, InfoCmd, args.Cmd)
	assertEqualURL(t, "APIURI", args.APIURL)
	assert.False(t, args.Verbose)
	assert.False(t, args.ShowConsumers)
	assert.False(t, args.InsecureTLS)
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
	assertEqualURL(t, "APIURI", args.APIURL)
	assert.Nil(t, args.SaveDir)
	assert.False(t, args.Verbose)
	assert.Equal(t, "EXPR", args.QueueFilter)
	assert.True(t, args.ShowStats)
	assert.True(t, args.ShowConsumers)
	assert.True(t, args.ShowDefaultExchange)
	assert.True(t, args.NoColor)
	assert.True(t, args.InsecureTLS)
	assert.True(t, args.OmitEmptyExchanges)
}

func TestCliPubCmdFromFileMinimalOptsSet(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"pub", "--uri=uri"})

	assert.Nil(t, err)
	assert.Equal(t, PubCmd, args.Cmd)
	assertEqualURL(t, "uri", args.AMQPURL)
	assert.Nil(t, args.PubExchange)
	assert.Nil(t, args.Source)
	assert.Nil(t, args.PubRoutingKey)
	assert.Equal(t, "raw", args.Format)
	assert.Nil(t, args.Delay)
	assert.Equal(t, 1., args.Speed)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
}
func TestCliPubCmdFromFileAllOptsSet(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"pub", "--uri=uri", "--exchange=exchange", "file",
			"--routingkey=key", "--delay=5s", "--format=json"})

	assert.Nil(t, err)
	assert.Equal(t, PubCmd, args.Cmd)
	assertEqualURL(t, "uri", args.AMQPURL)
	assert.Equal(t, "exchange", *args.PubExchange)
	assert.Equal(t, "file", *args.Source)
	assert.Equal(t, "key", *args.PubRoutingKey)
	assert.Equal(t, "json", args.Format)
	assert.Equal(t, 5*time.Second, *args.Delay)
	assert.Equal(t, 1., args.Speed)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
}

func TestCliPubCmdURLFromEnv(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	os.Setenv(key, "uri")
	defer os.Unsetenv(key)
	args, err := ParseCommandLineArgs(
		[]string{"pub"})

	assert.Nil(t, err)
	assert.Equal(t, PubCmd, args.Cmd)
	assertEqualURL(t, "uri", args.AMQPURL)
}

func TestCliPubCmdFailsWithMissingURL(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	os.Unsetenv(key)
	_, err := ParseCommandLineArgs([]string{"pub"})
	assert.NotNil(t, err)
}

func TestCliPubCmdFailsWithInvalidDelay(t *testing.T) {
	_, err := ParseCommandLineArgs([]string{"pub", "--uri=uri", "--delay=invalid"})
	assert.NotNil(t, err)
}

func TestCliPubCmdFailsWithInvalidSpeedup(t *testing.T) {
	_, err := ParseCommandLineArgs([]string{"pub", "--uri=uri", "--speed=invalid"})
	assert.NotNil(t, err)
}

func TestCliPubCmdFailsWithInvalidFormatSpec(t *testing.T) {
	_, err := ParseCommandLineArgs([]string{"pub", "--uri=uri", "--format=invalid"})
	assert.NotNil(t, err)
}

func TestCliPubCmdFromStdinWithRoutingKeyJsonFormatIsRecognized(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"pub", "--uri=uri", "--exchange=exchange1", "--routingkey=key", "--format=json"})

	assert.Nil(t, err)
	assert.Equal(t, PubCmd, args.Cmd)
	assertEqualURL(t, "uri", args.AMQPURL)
	assert.Equal(t, "exchange1", *args.PubExchange)
	assert.Equal(t, "key", *args.PubRoutingKey)
	assert.Nil(t, args.Source)
	assert.Equal(t, "json", args.Format)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
}

func TestCliPubCmdFromStdinWithJsonFormatDeprecatedIsRecognized(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"pub", "--uri=uri", "--json"})

	assert.Nil(t, err)
	assert.Equal(t, PubCmd, args.Cmd)
	assertEqualURL(t, "uri", args.AMQPURL)
	assert.Nil(t, args.PubExchange)
	assert.Nil(t, args.PubRoutingKey)
	assert.Nil(t, args.Source)
	assert.Equal(t, "json", args.Format)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
}

func TestCliSubCmdInvalidFormatReturnsError(t *testing.T) {
	_, err := ParseCommandLineArgs([]string{"sub", "queue", "--uri=uri", "--format=invalid"})
	assert.NotNil(t, err)
}

func TestCliSubCmdSaveToDir(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"sub", "queuename", "--uri=uri", "--saveto", "dir"})

	assert.Nil(t, err)
	assert.Equal(t, SubCmd, args.Cmd)
	assert.Equal(t, "queuename", args.QueueName)
	assertEqualURL(t, "uri", args.AMQPURL)
	assert.Equal(t, "dir", *args.SaveDir)
	assert.True(t, args.AutoAck)
}

func TestCliCreateQueue(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "create", "name", "--uri=uri"})

	assert.Nil(t, err)
	assert.Equal(t, QueueCreateCmd, args.Cmd)
	assert.Equal(t, "name", args.QueueName)
	assertEqualURL(t, "uri", args.AMQPURL)
	assert.False(t, args.Durable)
	assert.False(t, args.Autodelete)
}

func TestCliCreateDurableAutodeleteQueue(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "create", "name", "--uri=uri",
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
	assertEqualURL(t, "uri", args.AMQPURL)
}

func TestCliPurgeQueue(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "purge", "name", "--uri", "uri"})

	assert.Nil(t, err)
	assert.Equal(t, QueuePurgeCmd, args.Cmd)
	assert.Equal(t, "name", args.QueueName)
	assertEqualURL(t, "uri", args.AMQPURL)
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
	assertEqualURL(t, "uri", args.AMQPURL)
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
	assertEqualURL(t, "uri", args.AMQPURL)
}

func TestCliCreateExchange(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"exchange", "create", "name", "--type", "topic",
			"--uri", "uri"})

	assert.Nil(t, err)
	assert.Equal(t, ExchangeCreateCmd, args.Cmd)
	assert.Equal(t, "name", args.ExchangeName)
	assert.Equal(t, "topic", args.ExchangeType)
	assert.False(t, args.Durable)
	assert.False(t, args.Autodelete)
	assertEqualURL(t, "uri", args.AMQPURL)
}

func TestCliCreateDurableAutodeleteExchange(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"exchange", "create", "name", "--type", "topic",
			"--uri", "uri", "--durable", "--autodelete"})

	assert.Nil(t, err)
	assert.Equal(t, ExchangeCreateCmd, args.Cmd)
	assert.True(t, args.Durable)
	assert.True(t, args.Autodelete)
	assertEqualURL(t, "uri", args.AMQPURL)
}

func TestCliRemoveExchange(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"exchange", "rm", "name", "--uri", "uri"})

	assert.Nil(t, err)
	assert.Equal(t, ExchangeRemoveCmd, args.Cmd)
	assert.Equal(t, "name", args.ExchangeName)
	assertEqualURL(t, "uri", args.AMQPURL)
}

func TestCliCloseConnectionWithDefaultReason(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"conn", "close", "conn-name", "--api", "uri"})

	assert.Nil(t, err)
	assert.Equal(t, ConnCloseCmd, args.Cmd)
	assert.Equal(t, "conn-name", args.ConnName)
	assert.Equal(t, "closed by rabtap", args.CloseReason)
	assertEqualURL(t, "uri", args.APIURL)
}

func TestCliCloseConnection(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"conn", "close", "conn-name", "--api", "uri",
			"--reason", "reason"})

	assert.Nil(t, err)
	assert.Equal(t, ConnCloseCmd, args.Cmd)
	assertEqualURL(t, "uri", args.APIURL)
	assert.Equal(t, "conn-name", args.ConnName)
	assert.Equal(t, "reason", args.CloseReason)
}

func TestParseNoColorFromEnvironment(t *testing.T) {
	const key = "NO_COLOR"
	os.Setenv(key, "1")
	defer os.Unsetenv(key)

	args, err := ParseCommandLineArgs(
		[]string{"info", "--api=uri"})

	assert.Nil(t, err)
	assert.Equal(t, InfoCmd, args.Cmd)
	assertEqualURL(t, "uri", args.APIURL)
	assert.True(t, args.NoColor)
}
