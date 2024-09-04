// Copyright (C) 2017-2021 Jan Delgado

package main

// the command line is the UI of the tool. so we test it carefully.

import (
	"math"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestParseKeyValueExpressions(t *testing.T) {
	testcases := []struct {
		desc                     string
		probe                    string
		expectedKey, expectedVal string
		err                      bool
	}{
		{"empty is invalid", "", "", "", true},
		{"without assignment is invalid", "a b", "", "", true},
		{"without lhs is invalid", "=rhs", "", "", true},
		{"without rhs is invalid", "lhs=", "", "", true},
		{"standard case", "key=value", "key", "value", false},
		{"standard case with whitespace", "  key = value ", "key", "value", false},
		{"standard case with whitespace and special chars", "  key_1.2 = value%3 ", "key_1.2", "value%3", false},
	}

	for _, tc := range testcases {
		k, v, err := parseKeyValue(tc.probe)
		assert.Equal(t, tc.expectedKey, k, tc.desc)
		assert.Equal(t, tc.expectedVal, v, tc.desc)
		assert.Equal(t, tc.err, err != nil, tc.desc)
	}
}

func TestParseKeyValueListParsesValidList(t *testing.T) {
	kv, err := parseKeyValueList([]string{"a=b"})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "b"}, kv)
}

func TestParseKeyValueListParsesFailsOnInvalidInput(t *testing.T) {
	_, err := parseKeyValueList([]string{"a="})
	assert.Error(t, err)
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
	t.Setenv(key, "uri")
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
		"--color":         false,
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
	t.Setenv(key1, "/tmp/tls-cert.pem")
	t.Setenv(key2, "/tmp/tls-key.pem")
	t.Setenv(key3, "/tmp/tls-ca.pem")
	args := map[string]interface{}{
		"--tls-cert-file": nil,
		"--tls-key-file":  nil,
		"--tls-ca-file":   nil,
		"--verbose":       false,
		"--insecure":      false,
		"--no-color":      false,
		"--color":         false,
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

func TestCliTapCmdSingleUriFromEnv(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	t.Setenv(key, "uri")

	args, err := ParseCommandLineArgs(
		[]string{"tap", "exchange1:binding1", "--silent", "--limit=99"})

	assert.Nil(t, err)
	assert.Equal(t, TapCmd, args.Cmd)
	assert.Equal(t, 1, len(args.TapConfig))
	assertEqualURL(t, "uri", args.TapConfig[0].AMQPURL)
	assert.Equal(t, 1, len(args.TapConfig[0].Exchanges))
	assert.Equal(t, "exchange1", args.TapConfig[0].Exchanges[0].Exchange)
	assert.Equal(t, "binding1", args.TapConfig[0].Exchanges[0].BindingKey)
	assert.Equal(t, time.Duration(math.MaxInt64), args.IdleTimeout)
	assert.False(t, args.NoColor)
	assert.Equal(t, int64(99), args.Limit)
}

func TestCliTapFailsWhenNoURIisSpecified(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	defer os.Unsetenv(key)

	_, err := ParseCommandLineArgs(
		[]string{"tap", "exchange1:binding1"})

	assert.NotNil(t, err)
}

func TestCliTapCmdInvalidNumReturnsError(t *testing.T) {
	_, err := ParseCommandLineArgs([]string{"tap", "exchange:binding", "--uri=uri", "--limit=invalid"})
	assert.NotNil(t, err)
}

func TestCliTapCmdWithMultipleUris(t *testing.T) {
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
	assert.Nil(t, args.SaveDir)
	assert.Equal(t, InfiniteMessages, args.Limit)
	assert.False(t, args.NoColor)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
}

func TestCliAllOptsInTapCommandiAreRecognized(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"tap", "--uri=uri", "exchange:binding", "--silent", "--verbose",
			"--format=json-nopp", "--insecure", "--saveto", "savedir", "--limit", "123",
			"--idle-timeout=10s", "--filter=filter"})

	assert.Nil(t, err)
	assert.Equal(t, 1, len(args.TapConfig))
	assert.Equal(t, TapCmd, args.Cmd)
	assertEqualURL(t, "uri", args.TapConfig[0].AMQPURL)
	assert.Equal(t, 1, len(args.TapConfig[0].Exchanges))
	assert.Equal(t, "exchange", args.TapConfig[0].Exchanges[0].Exchange)
	assert.Equal(t, "binding", args.TapConfig[0].Exchanges[0].BindingKey)
	assert.Equal(t, int64(123), args.Limit)
	assert.Equal(t, time.Duration(time.Second*10), args.IdleTimeout)
	assert.Equal(t, "savedir", *args.SaveDir)
	assert.Equal(t, "json-nopp", args.Format)
	assert.Equal(t, "filter", args.Filter)
	assert.True(t, args.Verbose)
	assert.True(t, args.Silent)
	assert.True(t, args.InsecureTLS)
}

func TestCliInfoCmdIsParsed(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"info", "--api=APIURI"})

	assert.Nil(t, err)
	assert.Equal(t, 0, len(args.TapConfig))
	assert.Equal(t, InfoCmd, args.Cmd)
	assertEqualURL(t, "APIURI", args.APIURL)
	assert.Equal(t, "true", args.Filter)
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
	t.Setenv(key, "APIURI")

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
	assert.False(t, args.ForceColor)
	assert.Equal(t, "byExchange", args.InfoMode)
}

func TestCliInfoCmdAllOptionsAreSet(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"info", "--api=APIURI", "--stats", "--consumers",
			"--filter=EXPR", "--omit-empty",
			"--no-color", "--color", "-k", "--show-default"})

	assert.Nil(t, err)
	assert.Equal(t, 0, len(args.TapConfig))
	assert.Equal(t, InfoCmd, args.Cmd)
	assertEqualURL(t, "APIURI", args.APIURL)
	assert.Nil(t, args.SaveDir)
	assert.False(t, args.Verbose)
	assert.Equal(t, "EXPR", args.Filter)
	assert.True(t, args.ShowStats)
	assert.True(t, args.ShowConsumers)
	assert.True(t, args.ShowDefaultExchange)
	assert.True(t, args.NoColor)
	assert.True(t, args.ForceColor)
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
	assert.False(t, args.Confirms)
	assert.False(t, args.Mandatory)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
	assert.Nil(t, args.Properties.ContentType)
}

func TestCliPubCmdFromFileAllOptsSet(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"pub", "--uri=uri", "--exchange=exchange", "file",
			"--routingkey=key", "--delay=5s", "--format=json",
			"--confirms", "--mandatory", "--property=ContentEncoding=gzip"})

	require.Nil(t, err)
	assert.Equal(t, PubCmd, args.Cmd)
	assertEqualURL(t, "uri", args.AMQPURL)
	assert.Equal(t, "exchange", *args.PubExchange)
	assert.Equal(t, "file", *args.Source)
	assert.Equal(t, "key", *args.PubRoutingKey)
	assert.Equal(t, "json", args.Format)
	assert.Equal(t, 5*time.Second, *args.Delay)
	assert.Equal(t, 1., args.Speed)
	assert.True(t, args.Confirms)
	assert.True(t, args.Mandatory)
	assert.False(t, args.Verbose)
	assert.False(t, args.InsecureTLS)
	assert.Equal(t, "gzip", *args.Properties.ContentEncoding)
}

func TestCliPubCmdURLFromEnv(t *testing.T) {
	const key = "RABTAP_AMQPURI"
	t.Setenv(key, "uri")
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

func TestCliSubCmdOffsetSetsStreamOffsetArg(t *testing.T) {
	args, err := ParseCommandLineArgs([]string{"sub", "queue", "--uri=uri", "--offset=123"})
	assert.NoError(t, err)
	assert.Equal(t, args.Args["x-stream-offset"], "123")
}

func TestCliSubSetsInfiniteTimeoutWhenNotSpecified(t *testing.T) {
	args, err := ParseCommandLineArgs([]string{"sub", "queue", "--uri=uri"})
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(math.MaxInt64), args.IdleTimeout)
}

func TestCliSubCmdInvalidFormatReturnsError(t *testing.T) {
	_, err := ParseCommandLineArgs([]string{"sub", "queue", "--uri=uri", "--format=invalid"})
	assert.Error(t, err)
}

func TestCliSubCmdInvalidNumReturnsError(t *testing.T) {
	_, err := ParseCommandLineArgs([]string{"sub", "queue", "--uri=uri", "--limit=invalid"})
	assert.Error(t, err)
}

func TestCliSubCmdAllOptsSet(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"sub", "queuename", "--uri", "uri", "--saveto", "dir",
			"--limit=99", "--offset=123", "--idle-timeout=10s", "--filter=filter"})

	assert.Nil(t, err)
	assert.Equal(t, SubCmd, args.Cmd)
	assert.Equal(t, "queuename", args.QueueName)
	assertEqualURL(t, "uri", args.AMQPURL)
	assert.Equal(t, "dir", *args.SaveDir)
	assert.Equal(t, int64(99), args.Limit)
	assert.Equal(t, time.Duration(time.Second*10), args.IdleTimeout)
	assert.Equal(t, "filter", args.Filter)
	assert.False(t, args.Reject)
	assert.False(t, args.Requeue)
}

func TestCliCreateQueue(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "create", "name", "--uri=uri", "--args=x=y"})

	assert.NoError(t, err)
	assert.Equal(t, QueueCreateCmd, args.Cmd)
	assert.Equal(t, "name", args.QueueName)
	assertEqualURL(t, "uri", args.AMQPURL)
	assert.Equal(t, map[string]string{"x": "y", "x-queue-type": "classic"}, args.Args)
	assert.False(t, args.Durable)
	assert.False(t, args.Autodelete)
}

func TestCliCreateQueueAllOptsSet(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "create", "name", "--uri=uri",
			"--durable", "--autodelete", "--lazy", "--queue-type=quorum"})

	assert.NoError(t, err)
	assert.Equal(t, QueueCreateCmd, args.Cmd)
	assert.Equal(t, "name", args.QueueName)
	assertEqualURL(t, "uri", args.AMQPURL)
	assert.Equal(t, map[string]string{"x-queue-type": "quorum", "x-queue-mode": "lazy"}, args.Args)
	assert.True(t, args.Durable)
	assert.True(t, args.Autodelete)
}

func TestCliRemoveQueue(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "rm", "name", "--uri", "uri"})

	assert.NoError(t, err)
	assert.Equal(t, QueueRemoveCmd, args.Cmd)
	assert.Equal(t, "name", args.QueueName)
	assertEqualURL(t, "uri", args.AMQPURL)
}

func TestCliPurgeQueue(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "purge", "name", "--uri", "uri"})

	assert.NoError(t, err)
	assert.Equal(t, QueuePurgeCmd, args.Cmd)
	assert.Equal(t, "name", args.QueueName)
	assertEqualURL(t, "uri", args.AMQPURL)
}

func TestCliUnbindQueue(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "unbind", "queuename", "from", "exchangename",
			"--bindingkey", "key", "--uri", "uri"})

	assert.NoError(t, err)
	assert.Equal(t, QueueUnbindCmd, args.Cmd)
	assert.Equal(t, "queuename", args.QueueName)
	assert.Equal(t, "exchangename", args.ExchangeName)
	assert.Equal(t, "key", args.BindingKey)
	assertEqualURL(t, "uri", args.AMQPURL)
}

func TestCliBindQueueWithBindingKey(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "bind", "queuename", "to", "exchangename",
			"--bindingkey", "key", "--uri", "uri"})

	assert.NoError(t, err)
	assert.Equal(t, QueueBindCmd, args.Cmd)
	assert.Equal(t, "queuename", args.QueueName)
	assert.Equal(t, "exchangename", args.ExchangeName)
	assert.Equal(t, "key", args.BindingKey)
	assert.Equal(t, map[string]string{}, args.Args)
	assert.Equal(t, HeaderNone, args.HeaderMode)
	assertEqualURL(t, "uri", args.AMQPURL)
}

func TestCliBindQueueWithHeaders(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"queue", "bind", "queuename", "to", "exchangename",
			"--header", "a=b", "--header", "c=d", "--any", "--uri", "uri"})

	assert.Nil(t, err)
	assert.Equal(t, QueueBindCmd, args.Cmd)
	assert.Equal(t, "queuename", args.QueueName)
	assert.Equal(t, "exchangename", args.ExchangeName)
	assert.Equal(t, "", args.BindingKey)
	assert.Equal(t, map[string]string{"a": "b", "c": "d"}, args.Args)
	assert.Equal(t, HeaderMatchAny, args.HeaderMode)
	assertEqualURL(t, "uri", args.AMQPURL)
}

func TestCliCreateExchange(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"exchange", "create", "name", "--type", "topic", "--args=x=y",
			"--uri", "uri"})

	assert.Nil(t, err)
	assert.Equal(t, ExchangeCreateCmd, args.Cmd)
	assert.Equal(t, "name", args.ExchangeName)
	assert.Equal(t, "topic", args.ExchangeType)
	assert.False(t, args.Durable)
	assert.False(t, args.Autodelete)
	assert.Equal(t, map[string]string{"x": "y"}, args.Args)
	assertEqualURL(t, "uri", args.AMQPURL)
}

func TestCliBindExchangeToExchangeWithBindingKey(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"exchange", "bind", "source", "to", "dest",
			"--bindingkey", "key", "--uri", "uri"})

	assert.NoError(t, err)
	assert.Equal(t, ExchangeBindToExchangeCmd, args.Cmd)
	assert.Equal(t, "source", args.ExchangeName)
	assert.Equal(t, "dest", args.DestExchangeName)
	assert.Equal(t, "key", args.BindingKey)
	assert.Equal(t, map[string]string{}, args.Args)
	assert.Equal(t, HeaderNone, args.HeaderMode)
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
	t.Setenv(key, "1")

	args, err := ParseCommandLineArgs(
		[]string{"info", "--api=uri"})

	assert.Nil(t, err)
	assert.Equal(t, InfoCmd, args.Cmd)
	assertEqualURL(t, "uri", args.APIURL)
	assert.True(t, args.NoColor)
	assert.False(t, args.ForceColor)
}

func TestParseForceColort(t *testing.T) {
	args, err := ParseCommandLineArgs(
		[]string{"info", "--api=uri", "--color"})

	assert.Nil(t, err)
	assert.Equal(t, InfoCmd, args.Cmd)
	assertEqualURL(t, "uri", args.APIURL)
	assert.False(t, args.NoColor)
	assert.True(t, args.ForceColor)
}
