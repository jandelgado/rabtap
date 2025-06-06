// command line parsing for rabtap
// TODO split in per-command parsers
// TODO use docopt's bind feature to simplify mappings
// TODO consider using a different cli parser package
// Copyright (C) 2017-2024 Jan Delgado

package main

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	docopt "github.com/docopt/docopt-go"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// version and commit holds the application version and is set during link
// using "go -ldflags "-X main.version=a.b.c" (defaults used by goreleaser)
var (
	BuildVersion   = "(version)"
	BuildCommit    = "(commit)"
	BuildGoVersion = "(goversion)"
	BuildDate      = "(builddate)"
)

const (
	// note: usage is DSL interpreted by docopt - this is code. Change carefully.
	usage = `rabtap - RabbitMQ wire tap.                    github.com/jandelgado/rabtap

Usage:
  rabtap info [--api=APIURI] [--consumers] [--stats] [--filter=EXPR] [--omit-empty]
              [--show-default] [--mode=MODE] [--format=FORMAT] [-kncv]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap tap EXCHANGES [--uri=URI] [--saveto=DIR] [--format=FORMAT]  [--limit=NUM] 
	      [--idle-timeout=DURATION] [--filter=EXPR] [-jkncsv]
	      [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap (tap --uri=URI EXCHANGES)... [--saveto=DIR] [--format=FORMAT]  [--limit=NUM] 
	      [--idle-timeout=DURATION] [--filter=EXPR] [-jkncsv]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap sub QUEUE [--uri URI] [--saveto=DIR] [--format=FORMAT] [--limit=NUM]
              [--offset=OFFSET] [--args=KV]... [(--reject [--requeue])] [-jkcsvn]
              [--filter=EXPR] [--idle-timeout=DURATION]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap pub  [--uri=URI] [SOURCE] [--exchange=EXCHANGE] [--format=FORMAT]
              [--routingkey=KEY | (--header=KV)...] [ (--property=KV)... ]
              [--confirms] [--mandatory] [--delay=DELAY | --speed=FACTOR] [-jkv]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap exchange create EXCHANGE [--uri=URI] [--type=TYPE] [--args=KV]... [-kv]
              [--autodelete] [--durable]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap exchange bind EXCHANGE to DESTEXCHANGE [--uri=URI] [-kv]
              (--bindingkey=KEY | (--header=KV)... (--all|--any))
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap exchange rm EXCHANGE [--uri=URI] [-kv]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap queue create QUEUE [--uri=URI] [--queue-type=TYPE] [--args=KV]... [-kv]
              [--autodelete] [--durable] [--lazy]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap queue bind QUEUE to EXCHANGE [--uri=URI] [-kv]
              (--bindingkey=KEY | (--header=KV)... (--all|--any))
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap queue unbind QUEUE from EXCHANGE [--uri=URI] [-kv]
              (--bindingkey=KEY | (--header=KV)... (--all|--any))
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap queue rm QUEUE [--uri=URI] [-kv]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap queue purge QUEUE [--uri=URI] [-kv]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap conn close CONNECTION [--api=APIURI] [--reason=REASON] [-kv]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap --version
  rabtap (-h | --help | help) [properties]

Arguments and options:
 EXCHANGES            comma-separated list of exchanges and optional binding keys,
                      e.g. amq.topic:# or exchange1:key1,exchange2:key2.
 EXCHANGE             name of an exchange, e.g. amq.direct.
 DESTEXCHANGE         name of a a destination exchange in an exchange-to-exchange binding.
 SOURCE               file or directory to publish in pub mode. If omitted, stdin will be read.
 QUEUE                name of a queue.
 CONNECTION           name of a connection.
 DIR                  directory to read messages from.
 -a, --autodelete     create auto delete exchange/queue.
 --all                set x-match=all option in header based routing.
 --any                set x-match=any option in header based routing.
 --api=APIURI         connect to given API server. If APIURL is omitted,
                      the environment variable RABTAP_APIURI will be used.
 --args=KV            A key value pair in the form of "key=value" passed as
                      additional arguments. e.g. '--args=x-queue-type=quorum'
 -b, --bindingkey=KEY binding key to use in bind queue command.
 --by-connection      output of info command starts with connections.
 -c, --color          force colored output
 --confirms           enable publisher confirms and wait for confirmations.
 --consumers          include consumers and connections in output of info command.
 --delay=DELAY        Time to wait between sending messages during publish.
                      If not set then messages will be delayed as recorded.
                      The value must be suffixed with a time unit, e.g. ms, s etc.
 -d, --durable        create durable exchange/queue.
 --exchange=EXCHANGE  Optional exchange to publish to. If omitted, exchange will
                      be taken from message being published (see JSON message format).
 --filter=EXPR        Predicate for sub, tap, info command to filter the output [default: true]
 --format=FORMAT      * for tap, pub, sub command: format to write/read messages to console
                        and optionally to file (when --saveto DIR is given).
                        Valid options are: "raw", "json", "json-nopp". Default: raw
                      * for info command: controls generated output format. Valid
                        options are: "text", "dot". Default: text
 -h, --help           print this help
 --header=KV          A key value pair in the form of "key=value" used as a
                      routing- or binding-key. Can occur multiple times.
 --idle-timeout=DURATION end reading messages when no new message was received
                      for the given duration.  The value must be suffixed with 
                      a time unit, e.g. ms, s etc.
 -j, --json           deprecated. Use "--format json" instead.
 -k, --insecure       allow insecure TLS connections (no certificate check).
 --lazy               create a lazy queue.
 --limit=NUM          Stop afer NUM messages were received. When set to 0, will
                      run until terminated [default: 0].
 --mandatory          enable mandatory publishing (messages must be delivered to queue).
 --mode=MODE          mode for info command. One of "byConnection", "byExchange".
                      [default: byExchange].
 -n, --no-color       don't colorize output (see also environment variable NO_COLOR).
 --omit-empty         don't show echanges without bindings in info command.
 --offset=OFFSET      Offset when reading from a stream. Can be 'first', 'last',
                      'next', a duration like '10m', a RFC3339-Timestamp or
                      an integer index value. Basically it is an alias for
                      '--args=x-stream-offset=OFFSET'.
 --property=KV        A key value pair in the form of "key=value" to specify 
	              message properties like e.g. the content-type.
 --queue-type=TYPE    type of queue [default: classic].
 --reason=REASON      reason why the connection was closed [default: closed by rabtap].
 --reject             Reject messages. Default behaviour is to acknowledge messages.
 --requeue            Instruct broker to requeue rejected message
 -r, --routingkey=KEY routing key to use in publish mode. If omitted, routing key
                      will be taken from message being published (see JSON
                      message format).
 --saveto=DIR         also save messages and metadata to DIR.
 --show-default       include default exchange in output info command.
 -s, --silent         suppress message output to stdout.
 --speed=FACTOR       Speed factor to use during publish [default: 1.0].
 --stats              include statistics in output of info command.
 -t, --type=TYPE      type of exchange [default: fanout].
 --tls-cert-file=CERTFILE A Cert file to use for client authentication.
 --tls-key-file=KEYFILE   A Key file to use for client authentication.
 --tls-ca-file=CAFILE     A CA Cert file to use with TLS.
 --uri=URI            connect to given AQMP broker. If omitted, the
                      environment variable RABTAP_AMQPURI will be used.
 -v, --verbose        enable verbose mode.
 --version            show version information and exit.

Examples:
  rabtap tap --uri amqp://guest:guest@localhost/ amq.fanout:
  rabtap tap --uri amqp://guest:guest@localhost/ amq.topic:#,amq.fanout:
  rabtap pub --uri amqp://guest:guest@localhost/ --exchange amq.topic message.json --format=json
  rabtap info --api http://guest:guest@localhost:15672/api

  # use RABTAP_AMQPURI environment variable to specify broker instead of --uri
  export RABTAP_AMQPURI=amqp://guest:guest@localhost:5672/
  rabtap queue create JDQ
  rabtap queue bind JDQ to amq.topic --bindingkey=key
  echo "Hello"| gzip | rabtap pub --exchange amq.topic --routingkey "key" --property ContentType=gzip
  rabtap sub JDQ

  # print only messages that have ".Name == 'JAN'" in their JSON payload
  rabtap sub JDQ --filter="let b=fromJSON(r.toStr(r.body(r.msg))); b.Name == 'JAN'" 
  rabtap queue rm JDQ

  # use RABTAP_APIURI environment variable to specify mgmt api uri instead of --api
  export RABTAP_APIURI=http://guest:guest@localhost:15672/api
  rabtap info
  rabtap info --filter "r.binding.Source == 'amq.topic'" --omit-empty
  rabtap conn close "172.17.0.1:40874 -> 172.17.0.2:5672"

  # use RABTAP_TLS_CERTFILE | RABTAP_TLS_KEYFILE | RABTAP_TLS_CAFILE environments variables
  # instead of specifying --tls-cert-file=CERTFILE --tls-key-file=KEYFILE --tls-ca-file=CAFILE
`

	propertiesHelp = `
The following message properties are used with the '--property Key=Value'
option of the pub command. All keys are case-insensitve. Use multiple
'--property' options to set multiple properties at once.

DeliveryMode    - delivery mode: 'transient' or 'persistent' 
Priority        - message priority for priority queues
Expiration      - message TTL (ms)
ContentType     - application use - MIME content type
ContentEncoding - application use - MIME content encoding
CorrelationId   - application use - correlation identifier
ReplyTo         - application use - address to reply to
MessageId       - application use - message identifier
Timestamp       - application use - RFC3339 message timestamp
Type            - application use - message type name
AppId           - application use - creating application id
UserId          - user id, validated if set
`
)

// ProgramCmd represents the mode of operation
type ProgramCmd int

const (
	// TapCmd sets mode to tapping mode
	TapCmd ProgramCmd = iota
	HelpCmd
	// PubCmd sets mode to message-publish
	PubCmd
	// SubCmd sets mode to message-subscribe
	SubCmd
	// InfoCmd shows info on exchanges and queues
	InfoCmd
	// ExchangeCreateCmd creates a new exchange
	ExchangeCreateCmd
	// ExchangeRemoveCmd remove an exchange
	ExchangeRemoveCmd
	// ExchangeBindToExchange binds an exchange to another exchange
	ExchangeBindToExchangeCmd
	// QueueCreateCmd creates a new queue
	QueueCreateCmd
	// QueueRemoveCmd removes a queue
	QueueRemoveCmd
	// QueueBindCmd binds a queue to an exchange
	QueueBindCmd
	// QueueUnbindCmd unbinds a queue from an exchange
	QueueUnbindCmd
	// QueuePurgeCmd purges a queue
	QueuePurgeCmd
	// ConnCloseCmd closes a connection
	ConnCloseCmd
)

type HelpTopic int

const (
	GeneralHelp HelpTopic = iota
	PropertiesHelp
)

type HeaderMode int

const (
	// match any headers (--any)
	HeaderMatchAny HeaderMode = iota
	// match all headers (-all)
	HeaderMatchAll
	// header based routing is not used
	HeaderNone
)

// parseKeyValue parses an expression of the form "key=value"
func parseKeyValue(expr string) (string, string, error) {
	re := regexp.MustCompile(`\s*([^= ]+)\s*=\s*([^= ]+)\s*`)
	all := re.FindStringSubmatch(expr)
	if all == nil {
		return "", "", fmt.Errorf("could not parse key-value expression %s", expr)
	}
	return all[1], all[2], nil
}

func parseKeyValueList(exprs []string) (map[string]string, error) {
	if exprs == nil {
		return map[string]string{}, nil
	}
	res := make(map[string]string, len(exprs))
	for _, expr := range exprs {
		k, v, err := parseKeyValue(expr)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", expr, err)
		}
		res[k] = v
	}
	return res, nil
}

type commonArgs struct {
	TLSCertFile string
	TLSKeyFile  string
	TLSCaFile   string
	Verbose     bool
	InsecureTLS bool
	NoColor     bool
	ForceColor  bool
	AMQPURL     *url.URL // pub, queue, exchange: amqp broker to use
}

const InfiniteMessages = int64(0)

// CommandLineArgs represents the parsed command line arguments
// TODO does not scale well - split in per-cmd structs
type CommandLineArgs struct {
	Cmd ProgramCmd
	commonArgs

	TapConfig []rabtap.TapConfiguration // configuration in tap mode
	APIURL    *url.URL

	PubExchange         *string        // pub: exchange to publish to
	PubRoutingKey       *string        // pub: routing key, defaults to ""
	Source              *string        // pub: file to send
	Speed               float64        // pub: speed factor
	Delay               *time.Duration // pub: fixed delay in ms
	Confirms            bool           // pub: wait for confirmations
	Mandatory           bool           // pub: set mandatory flag
	Properties          PropertiesOverride
	Limit               int64             // sub: optional limit
	Reject              bool              // sub: reject messages
	Requeue             bool              // sub: requeue rejectied messages
	IdleTimeout         time.Duration     // sub: idle timeout
	QueueName           string            // queue create, remove, bind, sub
	BindingKey          string            // a binding key
	ExchangeName        string            // exchange name  create, remove or queue bind
	DestExchangeName    string            // target exchange name  bind e2e
	ExchangeType        string            // exchange type create
	ShowConsumers       bool              // info: also show consumer
	InfoMode            string            // info: byExchange, byConnection
	ShowStats           bool              // info: also show statistics
	OmitEmptyExchanges  bool              // info: do not show exchanges wo/ bindings
	ShowDefaultExchange bool              // info: show default exchange
	Filter              string            // sub/tap/info: optional filter predicate
	Format              string            // output format, depends on command
	Durable             bool              // queue create, exchange create
	Autodelete          bool              // queue create, exchange create
	Args                map[string]string // optional additional arguments for pub, tap, queue
	SaveDir             *string           // save: optional directory to stores files to
	Silent              bool              // suppress message printing
	ConnName            string            // conn: name of connection
	CloseReason         string            // conn: reason of close
	HeaderMode          HeaderMode        // queue ceate, header based routing
	HelpTopic           HelpTopic
}

// getAMQPURL returns the ith entry of amqpURLs array or the value
// of the RABTAP_AMQPURI environment variable if i is out of array
// bounds or the returned value would be empty.
func getAMQPURL(amqpURLs []string, i int) (*url.URL, error) {
	var u string
	if i >= len(amqpURLs) {
		u = os.Getenv("RABTAP_AMQPURI")
		if u == "" {
			return nil, fmt.Errorf("--uri omitted but RABTAP_AMQPURI not set in environment")
		}
	} else {
		u = amqpURLs[i]
	}
	return url.Parse(u)
}

func parseAMQPURL(args map[string]interface{}) (*url.URL, error) {
	amqpURLs := args["--uri"].([]string)
	return getAMQPURL(amqpURLs, 0)
}

func parseAPIURI(args map[string]interface{}) (*url.URL, error) {
	var apiURL string
	if args["--api"] != nil {
		apiURL = args["--api"].(string)
	} else {
		apiURL = os.Getenv("RABTAP_APIURI")
	}
	if apiURL == "" {
		return nil, fmt.Errorf("--api omitted but RABTAP_APIURI not set in environment")
	}
	return url.Parse(apiURL)
}

func parseCommonArgs(args map[string]interface{}) commonArgs {
	var tlsCertFile string
	var tlsKeyFile string
	var tlsCaFile string
	if args["--tls-cert-file"] != nil {
		tlsCertFile = args["--tls-cert-file"].(string)
	} else {
		tlsCertFile = os.Getenv("RABTAP_TLS_CERTFILE")
	}
	if args["--tls-key-file"] != nil {
		tlsKeyFile = args["--tls-key-file"].(string)
	} else {
		tlsKeyFile = os.Getenv("RABTAP_TLS_KEYFILE")
	}
	if args["--tls-ca-file"] != nil {
		tlsCaFile = args["--tls-ca-file"].(string)
	} else {
		tlsCaFile = os.Getenv("RABTAP_TLS_CAFILE")
	}
	return commonArgs{
		TLSCertFile: tlsCertFile,
		TLSKeyFile:  tlsKeyFile,
		TLSCaFile:   tlsCaFile,
		Verbose:     args["--verbose"].(bool),
		InsecureTLS: args["--insecure"].(bool),
		NoColor:     args["--no-color"].(bool) || (os.Getenv("NO_COLOR") != ""),
		ForceColor:  args["--color"].(bool),
	}
}

func parseInfoCmdArgs(args map[string]interface{}) (CommandLineArgs, error) {
	result := CommandLineArgs{
		Cmd:                 InfoCmd,
		commonArgs:          parseCommonArgs(args),
		Filter:              args["--filter"].(string),
		OmitEmptyExchanges:  args["--omit-empty"].(bool),
		ShowConsumers:       args["--consumers"].(bool),
		ShowStats:           args["--stats"].(bool),
		ShowDefaultExchange: args["--show-default"].(bool),
	}

	mode := args["--mode"].(string)
	if mode != "byExchange" && mode != "byConnection" {
		return result, errors.New("--mode MODE must be one of {byConnection, byExchange}")
	}
	result.InfoMode = mode

	format := "text"
	if args["--format"] != nil {
		format = args["--format"].(string)
	}
	if format != "text" && format != "dot" {
		return result, errors.New("--format=FORMAT must be one of {text, dot}")
	}
	result.Format = format

	var err error
	if result.APIURL, err = parseAPIURI(args); err != nil {
		return result, fmt.Errorf("failed to parse API URL: %w", err)
	}
	return result, nil
}

func parseConnCmdArgs(args map[string]interface{}) (CommandLineArgs, error) {
	result := CommandLineArgs{
		commonArgs: parseCommonArgs(args),
	}

	var err error
	if result.APIURL, err = parseAPIURI(args); err != nil {
		return result, fmt.Errorf("failed to parse API URL: %w", err)
	}
	if args["close"].(bool) {
		result.Cmd = ConnCloseCmd
		result.ConnName = args["CONNECTION"].(string)
		result.CloseReason = args["--reason"].(string)
	}
	return result, nil
}

// parsePubSubFormatArg parse --format=FORMAT option for pub, sub, tap command.
func parsePubSubFormatArg(args map[string]interface{}) (string, error) {
	format := "raw"

	if args["--format"] != nil {
		format = args["--format"].(string)
	}

	// deprecated --json option equals "--format=json"
	if args["--json"] != nil && args["--json"].(bool) {
		format = "json"
	}

	if format != "json" && format != "json-nopp" && format != "raw" {
		return "", errors.New("--format=FORMAT must be one of {raw,json,json-nopp}")
	}
	return format, nil
}

func parseSubCmdArgs(args map[string]interface{}) (CommandLineArgs, error) {
	result := CommandLineArgs{
		Cmd:         SubCmd,
		commonArgs:  parseCommonArgs(args),
		Reject:      args["--reject"].(bool),
		Requeue:     args["--requeue"].(bool),
		QueueName:   args["QUEUE"].(string),
		Filter:      args["--filter"].(string),
		Silent:      args["--silent"].(bool),
		IdleTimeout: time.Duration(math.MaxInt64),
	}

	format, err := parsePubSubFormatArg(args)
	if err != nil {
		return result, err
	}
	result.Format = format

	if timeout := args["--idle-timeout"]; timeout != nil {
		duration, err := time.ParseDuration(timeout.(string))
		if err != nil {
			return result, fmt.Errorf("failed to parse --idle-timeout: %w", err)
		}
		result.IdleTimeout = duration
	}
	if args["--limit"] != nil {
		limit, err := strconv.ParseInt(args["--limit"].(string), 10, 64)
		if err != nil {
			return result, fmt.Errorf("failed to parse --limit: %w", err)
		}
		result.Limit = limit
	}
	result.Args, err = parseKVListOption("--args", args)
	if err != nil {
		return result, fmt.Errorf("failed to parse --args: %w", err)
	}

	if args["--saveto"] != nil {
		saveDir := args["--saveto"].(string)
		result.SaveDir = &saveDir
	}
	if result.AMQPURL, err = parseAMQPURL(args); err != nil {
		return result, fmt.Errorf("failed to parse AMQP URL: %w", err)
	}
	if offset := args["--offset"]; offset != nil {
		result.Args["x-stream-offset"] = offset.(string)
	}
	return result, nil
}

func parseBindingKey(args map[string]interface{}) string {
	if key, ok := args["--bindingkey"].(string); ok {
		return key
	}
	return ""
}

func parseKVListOption(name string, args map[string]interface{}) (map[string]string, error) {
	if values, ok := args[name].([]string); ok {
		return parseKeyValueList(values)
	}
	return map[string]string{}, nil
}

func parseHeaderMode(args map[string]interface{}) HeaderMode {
	switch {
	case args["--any"].(bool):
		return HeaderMatchAny
	case args["--all"].(bool):
		return HeaderMatchAll
	default:
		return HeaderNone
	}
}

func parseQueueCmdArgs(args map[string]interface{}) (CommandLineArgs, error) {
	result := CommandLineArgs{
		commonArgs: parseCommonArgs(args),
		QueueName:  args["QUEUE"].(string),
	}
	var err error
	if result.AMQPURL, err = parseAMQPURL(args); err != nil {
		return result, fmt.Errorf("failed to parse AMQP URL: %w", err)
	}
	switch {
	case args["create"].(bool):
		result.Cmd = QueueCreateCmd
		result.Durable = args["--durable"].(bool)
		result.Autodelete = args["--autodelete"].(bool)
		result.Args, err = parseKVListOption("--args", args)
		if err != nil {
			return result, fmt.Errorf("failed to parse --args: %w", err)
		}
		result.Args["x-queue-type"] = args["--queue-type"].(string)
		if args["--lazy"].(bool) {
			result.Args["x-queue-mode"] = "lazy"
		}
	case args["rm"].(bool):
		result.Cmd = QueueRemoveCmd
	case args["bind"].(bool):
		// bind QUEUE to EXCHANGE ([--bindingkey key] | (--header KEYVAL)* )
		var err error
		result.Cmd = QueueBindCmd
		result.BindingKey = parseBindingKey(args)

		result.Args, err = parseKVListOption("--header", args)
		if err != nil {
			return result, fmt.Errorf("failed to parse --header: %w", err)
		}
		result.HeaderMode = parseHeaderMode(args)
		result.ExchangeName = args["EXCHANGE"].(string)

	case args["unbind"].(bool):
		// unbind QUEUE from EXCHANGE [--bindingkey key]
		result.Cmd = QueueUnbindCmd
		result.BindingKey = parseBindingKey(args)
		result.Args, err = parseKVListOption("--header", args)
		if err != nil {
			return result, fmt.Errorf("failed to parse --header: %w", err)
		}
		result.HeaderMode = parseHeaderMode(args)
		result.ExchangeName = args["EXCHANGE"].(string)
	case args["purge"].(bool):
		result.Cmd = QueuePurgeCmd
	}
	return result, nil
}

func parseExchangeCmdArgs(args map[string]interface{}) (CommandLineArgs, error) {
	result := CommandLineArgs{
		commonArgs:   parseCommonArgs(args),
		ExchangeName: args["EXCHANGE"].(string),
	}

	var err error
	if result.AMQPURL, err = parseAMQPURL(args); err != nil {
		return result, err
	}
	switch {
	case args["create"].(bool):
		result.ExchangeType = args["--type"].(string)
		result.Cmd = ExchangeCreateCmd
		result.Durable = args["--durable"].(bool)
		result.Autodelete = args["--autodelete"].(bool)
		result.Args, err = parseKVListOption("--args", args)
		if err != nil {
			return result, fmt.Errorf("failed to parse --args: %w", err)
		}
	case args["rm"].(bool):
		result.Cmd = ExchangeRemoveCmd
	case args["bind"].(bool):
		// bind EXCHANGE to EXCHANGE ([--bindingkey key] | (--header KEYVAL)* )
		var err error
		result.Cmd = ExchangeBindToExchangeCmd
		result.BindingKey = parseBindingKey(args)

		result.Args, err = parseKVListOption("--header", args)
		if err != nil {
			return result, fmt.Errorf("failed to parse --header: %w", err)
		}
		result.HeaderMode = parseHeaderMode(args)
		result.DestExchangeName = args["DESTEXCHANGE"].(string)
	}
	return result, nil
}

func parsePublishCmdArgs(args map[string]interface{}) (CommandLineArgs, error) {
	result := CommandLineArgs{
		Cmd:        PubCmd,
		Confirms:   args["--confirms"].(bool),
		Mandatory:  args["--mandatory"].(bool),
		commonArgs: parseCommonArgs(args),
	}

	format, err := parsePubSubFormatArg(args)
	if err != nil {
		return result, err
	}
	result.Format = format

	if result.AMQPURL, err = parseAMQPURL(args); err != nil {
		return result, err
	}
	if args["--exchange"] != nil {
		exchange := args["--exchange"].(string)
		result.PubExchange = &exchange
	}
	result.Args, err = parseKVListOption("--header", args)
	if err != nil {
		return result, fmt.Errorf("failed to parse --header: %w", err)
	}
	if args["--routingkey"] != nil {
		routingKey := args["--routingkey"].(string)
		result.PubRoutingKey = &routingKey
	}
	if args["SOURCE"] != nil {
		file := args["SOURCE"].(string)
		result.Source = &file
	}
	if args["--delay"] != nil {
		delay, err := time.ParseDuration(args["--delay"].(string))
		if err != nil {
			return result, err
		}
		result.Delay = &delay
	}
	if args["--speed"] != nil {
		result.Speed, err = strconv.ParseFloat(args["--speed"].(string), 64)
		if err != nil {
			return result, fmt.Errorf("failed to parse --speed: %w", err)
		}
	}
	// multiple --property K=V allow to override message properties
	propsKV, err := parseKVListOption("--property", args)
	if err != nil {
		return result, fmt.Errorf("parse properties: %w", err)
	}
	props, err := parseMessageProperties(propsKV)
	if err != nil {
		return result, fmt.Errorf("parse properties: %w", err)
	}
	result.Properties = props

	return result, nil
}

func parseTapCmdArgs(args map[string]interface{}) (CommandLineArgs, error) {
	result := CommandLineArgs{
		Cmd:         TapCmd,
		commonArgs:  parseCommonArgs(args),
		Filter:      args["--filter"].(string),
		Silent:      args["--silent"].(bool),
		TapConfig:   []rabtap.TapConfiguration{},
		IdleTimeout: time.Duration(math.MaxInt64),
	}

	format, err := parsePubSubFormatArg(args)
	if err != nil {
		return result, err
	}
	result.Format = format

	if timeout := args["--idle-timeout"]; timeout != nil {
		duration, err := time.ParseDuration(timeout.(string))
		if err != nil {
			return result, fmt.Errorf("failed to parse --idle-timeout: %w", err)
		}
		result.IdleTimeout = duration
	}
	if args["--limit"] != nil {
		limit, err := strconv.ParseInt(args["--limit"].(string), 10, 64)
		if err != nil {
			return result, fmt.Errorf("failed to parse --limit: %w", err)
		}
		result.Limit = limit
	}

	if args["--saveto"] != nil {
		saveDir := args["--saveto"].(string)
		result.SaveDir = &saveDir
	}
	amqpURLs := args["--uri"].([]string)
	exchanges := args["EXCHANGES"].([]string)
	for i, exchange := range exchanges {
		// eihter the amqp uri is provided with --uri URI or the value
		// is used from the RABTAP_AMQPURI environment variable.
		amqpURL, err := getAMQPURL(amqpURLs, i)
		if err != nil {
			return result, fmt.Errorf("failed to parse AMQP URL: %w", err)
		}
		tapConfig, err := rabtap.NewTapConfiguration(amqpURL, exchange)
		if err != nil {
			return result, fmt.Errorf("failed to parse tap configuration: %w", err)
		}
		result.TapConfig = append(result.TapConfig, *tapConfig)
	}
	return result, nil
}

func parseHelpCmdArgs(args map[string]interface{}) (CommandLineArgs, error) {
	result := CommandLineArgs{Cmd: HelpCmd}

	if args["properties"].(bool) {
		result.HelpTopic = PropertiesHelp
	} else {
		result.HelpTopic = GeneralHelp
	}
	return result, nil
}

func formatVersion() string {
	return fmt.Sprintf("version: %s\ncommit: %s\nbuilt with: %s\nbuild date: %s",
		BuildVersion, BuildCommit, BuildGoVersion, BuildDate)
}

func parseCommandLineArgsWithSpec(spec string, cliArgs []string) (CommandLineArgs, error) {
	parser := docopt.Parser{SkipHelpFlags: true}
	args, err := parser.ParseArgs(spec, cliArgs, formatVersion())
	if err != nil {
		return CommandLineArgs{}, err
	}
	switch {
	case args["tap"].(int) > 0:
		return parseTapCmdArgs(args)
	case args["info"].(bool):
		return parseInfoCmdArgs(args)
	case args["pub"].(bool):
		return parsePublishCmdArgs(args)
	case args["sub"].(bool):
		return parseSubCmdArgs(args)
	case args["queue"].(bool):
		return parseQueueCmdArgs(args)
	case args["exchange"].(bool):
		return parseExchangeCmdArgs(args)
	case args["conn"].(bool):
		return parseConnCmdArgs(args)
	case args["--help"].(bool):
		fallthrough
	case args["help"].(bool):
		return parseHelpCmdArgs(args)
	}
	return CommandLineArgs{}, fmt.Errorf("command missing")
}

// PrintHelp is explicitly called when the "help" command is given. On -h or
// --help docopt internally prints help
func PrintHelp(topic HelpTopic) {
	switch topic {
	case GeneralHelp:
		docopt.PrintHelpOnly(nil, usage)
	case PropertiesHelp:
		fmt.Print(propertiesHelp)
	}
}

// ParseCommandLineArgs parses command line arguments into an object of
// type CommandLineArgs.
func ParseCommandLineArgs(cliArgs []string) (CommandLineArgs, error) {
	return parseCommandLineArgsWithSpec(usage, cliArgs)
}