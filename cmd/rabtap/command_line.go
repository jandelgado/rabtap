// Copyright (C) 2017 Jan Delgado

package main

import (
	"fmt"
	"os"

	docopt "github.com/docopt/docopt-go"
	rabtap "github.com/jandelgado/rabtap/pkg"
)

// RabtapAppVersion hold the application version and is set during link
// using "go -ldflags "-X main.RabtapAppVersion=a.b.c"
var RabtapAppVersion = "(version not specified)"

const (
	// note: usage is interpreted by docopt - this is code.
	usage = `rabtap - RabbitMQ wire tap.                  github.com/jandelgado/rabtap

Usage:
  rabtap -h|--help
  rabtap tap EXCHANGES [--uri URI] [--saveto=DIR] [-jknv]
  rabtap (tap --uri URI EXCHANGES)... [--saveto=DIR] [-jknv]
  rabtap info [--api APIURI] [--consumers] [--stats] 
              [--filter EXPR] 
              [--omit-empty] [--show-default] [--by-connection] [-knv]
  rabtap pub [--uri URI] EXCHANGE [FILE] [--routingkey=KEY] [-jkv]
  rabtap sub QUEUE [--uri URI] [--saveto=DIR] [-jkvn]
  rabtap exchange create EXCHANGE [--uri URI] [--type TYPE] [-adkv]
  rabtap exchange rm EXCHANGE [--uri URI] [-kv]
  rabtap queue create QUEUE [--uri URI] [-adkv]
  rabtap queue bind QUEUE to EXCHANGE --bindingkey=KEY [--uri URI] [-kv]
  rabtap queue unbind QUEUE from EXCHANGE --bindingkey=KEY [--uri URI] [-kv]
  rabtap queue rm QUEUE [--uri URI] [-kv]
  rabtap queue purge QUEUE [--uri URI] [-kv]
  rabtap conn close CONNECTION [--reason=REASON] [--api APIURI] [-kv]
  rabtap --version

Options:
 EXCHANGES            comma-separated list of exchanges and binding keys,
                      e.g. amq.topic:# or exchange1:key1,exchange2:key2.
 EXCHANGE             name of an exchange, e.g. amq.direct.
 FILE                 file to publish in pub mode. If omitted, stdin will
                      be read.
 QUEUE                name of a queue.
 CONNECTION           name of a connection.
 -a, --autodelete     create auto delete exchange/queue.
 --api APIURI         connect to given API server. If APIURI is omitted,
                      the environment variable RABTAP_APIURI will be used.
 -b, --bindingkey KEY binding key to use in bind queue command.
 --by-connection      output of info command starts with connections.
 --consumers          include consumers and connections in output of info command.
 -d, --durable        create durable exchange/queue.
 --filter EXPR        Predicate for info command to filter queues [default: true]
                      (see README.md for details)
 -h, --help           print this help.
 -j, --json           print/save/publish message metadata and body to a
                      single JSON file. JSON body is base64 encoded. Otherwise
                      metadata and body (as-is) are saved separately.
 -k, --insecure       allow insecure TLS connections (no certificate check).
 -n, --no-color       don't colorize output (also environment variable NO_COLOR)
 -o, --omit-empty     don't show echanges without bindings in info command.
 --reason=REASON      reason why the connection was closed 
                      [default: closed by rabtap].
 -r, --routingkey KEY routing key to use in publish mode.
 --saveto DIR         also save messages and metadata to DIR.
 --show-default       include default exchange in output info command.
 --stats              include statistics in output of info command.
 -t, --type TYPE      exchange type [default: fanout].
 --uri URI            connect to given AQMP broker. If omitted, the
                      environment variable RABTAP_AMQPURI will be used.
 -v, --verbose        enable verbose mode.
 --version            show version information and exit.

Examples:
  rabtap tap --uri amqp://guest:guest@localhost/ amq.fanout:
  rabtap tap --uri amqp://guest:guest@localhost/ amq.topic:#,amq.fanout:
  rabtap pub --uri amqp://guest:guest@localhost/ amq.topic message.json -j
  rabtap info --api http://guest:guest@localhost:15672/api

  # use RABTAP_AMQPURI environment variable to specify broker instead of --uri
  export RABTAP_AMQPURI=amqp://guest:guest@localhost:5672/
  echo "Hello" | rabtap pub amq.topic -r "some.key"
  rabtap sub JDQ
  rabtap queue create JDQ
  rabtap queue bind JDQ to amq.direct --bindingkey=key
  rabtap queue rm JDQ

  # use RABTAP_APIURI environment variable to specify mgmt api uri instead of --api
  export RABTAP_APIURI=http://guest:guest@localhost:15672/api
  rabtap info
  rabtap info --filter "binding.Source == 'amq.topic'" -o
  rabtap conn close "172.17.0.1:40874 -> 172.17.0.2:5672" 
`
)

// ProgramCmd represents the mode of operation
type ProgramCmd int

const (
	// TapCmd sets mode to tapping mode
	TapCmd ProgramCmd = iota
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

type commonArgs struct {
	Verbose     bool
	InsecureTLS bool
	NoColor     bool
	JSONFormat  bool   // output in JSON
	AmqpURI     string // pub, queue, exchange: amqp broker to use
}

// CommandLineArgs represents the parsed command line arguments
// TODO does not scale well - split in per-cmd structs?
type CommandLineArgs struct {
	Cmd ProgramCmd
	commonArgs

	TapConfig []rabtap.TapConfiguration // configuration in tap mode
	APIURI    string

	PubExchange         string  // pub mode: exchange to publish to
	PubRoutingKey       string  // pub mode: routing key, defaults to ""
	PubFile             *string // pub mode: file to send
	QueueName           string  // queue create, remove, bind, sub
	QueueBindingKey     string  // queue bind
	ExchangeName        string  // exchange name  create, remove or queue bind
	ExchangeType        string  // exchange type create, remove or queue bind
	ShowConsumers       bool    // info mode: also show consumer
	ShowByConnection    bool    // info mode: show by connection
	ShowStats           bool    // info mode: also show statistics
	QueueFilter         string  // info mode: optional filter predicate
	OmitEmptyExchanges  bool    // info mode: do not show exchanges wo/ bindings
	Durable             bool    // queue create, exchange create
	Autodelete          bool    // queue create, exchange create
	SaveDir             *string // save mode: optional directory to stores files to
	ShowDefaultExchange bool

	ConnName    string // conn mode: name of connection
	CloseReason string // conn mode: reason of close
}

// getAmqpURI returns the ith entry of amqpURIs array or the value
// of the RABTAP_AMQPURI environment variable if i is out of array
// bounds or the returned value would be empty.
func getAmqpURI(amqpURIs []string, i int) string {
	if i >= len(amqpURIs) {
		return os.Getenv("RABTAP_AMQPURI")
	}
	amqpURI := amqpURIs[i]
	if amqpURI == "" {
		return os.Getenv("RABTAP_AMQPURI")
	}
	return amqpURI
}

func parseAmqpURI(args map[string]interface{}) (string, error) {
	amqpURIs := args["--uri"].([]string)
	uri := getAmqpURI(amqpURIs, 0)
	if uri == "" {
		return "", fmt.Errorf("--uri omitted but RABTAP_AMQPURI not set in environment")
	}
	return uri, nil
}

func parseAPIURI(args map[string]interface{}) (string, error) {
	var apiURI string
	if args["--api"] != nil {
		apiURI = args["--api"].(string)
	} else {
		apiURI = os.Getenv("RABTAP_APIURI")
	}
	if apiURI == "" {
		return "", fmt.Errorf("--api omitted but RABTAP_APIURI not set in environment")
	}
	return apiURI, nil
}

func parseCommonArgs(args map[string]interface{}) commonArgs {
	return commonArgs{
		Verbose:     args["--verbose"].(bool),
		InsecureTLS: args["--insecure"].(bool),
		NoColor:     args["--no-color"].(bool) || (os.Getenv("NO_COLOR") != ""),
		JSONFormat:  args["--json"].(bool)}
}

func parseInfoCmdArgs(args map[string]interface{}) (CommandLineArgs, error) {
	result := CommandLineArgs{
		Cmd:                 InfoCmd,
		commonArgs:          parseCommonArgs(args),
		QueueFilter:         args["--filter"].(string),
		OmitEmptyExchanges:  args["--omit-empty"].(bool),
		ShowConsumers:       args["--consumers"].(bool),
		ShowStats:           args["--stats"].(bool),
		ShowDefaultExchange: args["--show-default"].(bool),
		ShowByConnection:    args["--by-connection"].(bool)}

	var err error
	if result.APIURI, err = parseAPIURI(args); err != nil {
		return result, err
	}
	return result, nil
}

func parseConnCmdArgs(args map[string]interface{}) (CommandLineArgs, error) {
	result := CommandLineArgs{
		commonArgs: parseCommonArgs(args)}

	var err error
	if result.APIURI, err = parseAPIURI(args); err != nil {
		return result, err
	}
	if args["close"].(bool) {
		result.Cmd = ConnCloseCmd
		result.ConnName = args["CONNECTION"].(string)
		result.CloseReason = args["--reason"].(string)
	}
	return result, nil
}

func parseSubCmdArgs(args map[string]interface{}) (CommandLineArgs, error) {
	result := CommandLineArgs{
		Cmd:        SubCmd,
		commonArgs: parseCommonArgs(args),
		QueueName:  args["QUEUE"].(string),
	}
	var err error
	if args["--saveto"] != nil {
		saveDir := args["--saveto"].(string)
		result.SaveDir = &saveDir
	}
	if result.AmqpURI, err = parseAmqpURI(args); err != nil {
		return result, err
	}
	return result, nil
}

func parseQueueCmdArgs(args map[string]interface{}) (CommandLineArgs, error) {
	result := CommandLineArgs{
		commonArgs: parseCommonArgs(args),
		QueueName:  args["QUEUE"].(string),
	}
	var err error
	if result.AmqpURI, err = parseAmqpURI(args); err != nil {
		return result, err
	}
	switch {
	case args["create"].(bool):
		result.Cmd = QueueCreateCmd
		result.Durable = args["--durable"].(bool)
		result.Autodelete = args["--autodelete"].(bool)
	case args["rm"].(bool):
		result.Cmd = QueueRemoveCmd
	case args["bind"].(bool):
		// bind QUEUE to EXCHANGE [--bindingkey key]
		result.Cmd = QueueBindCmd
		result.QueueBindingKey = args["--bindingkey"].(string)
		result.ExchangeName = args["EXCHANGE"].(string)
	case args["unbind"].(bool):
		// unbind QUEUE from EXCHANGE [--bindingkey key]
		result.Cmd = QueueUnbindCmd
		result.QueueBindingKey = args["--bindingkey"].(string)
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
		ExchangeType: args["--type"].(string)}

	var err error
	if result.AmqpURI, err = parseAmqpURI(args); err != nil {
		return result, err
	}
	switch {
	case args["create"].(bool):
		result.Cmd = ExchangeCreateCmd
		result.Durable = args["--durable"].(bool)
		result.Autodelete = args["--autodelete"].(bool)
	case args["rm"].(bool):
		result.Cmd = ExchangeRemoveCmd
	}
	return result, nil
}

func parsePublishCmdArgs(args map[string]interface{}) (CommandLineArgs, error) {
	result := CommandLineArgs{
		Cmd:        PubCmd,
		commonArgs: parseCommonArgs(args)}

	var err error
	if result.AmqpURI, err = parseAmqpURI(args); err != nil {
		return result, err
	}
	result.PubExchange = args["EXCHANGE"].(string)
	if args["--routingkey"] != nil {
		result.PubRoutingKey = args["--routingkey"].(string)
	}
	if args["FILE"] != nil {
		file := args["FILE"].(string)
		result.PubFile = &file
	}
	return result, nil
}

func parseTapCmdArgs(args map[string]interface{}) (CommandLineArgs, error) {
	result := CommandLineArgs{
		Cmd:        TapCmd,
		commonArgs: parseCommonArgs(args),
		TapConfig:  []rabtap.TapConfiguration{}}

	if args["--saveto"] != nil {
		saveDir := args["--saveto"].(string)
		result.SaveDir = &saveDir
	}
	amqpURIs := args["--uri"].([]string)
	exchanges := args["EXCHANGES"].([]string)
	for i, exchange := range exchanges {
		// eihter the amqp uri is provided with --uri URI or the value
		// is used from the RABTAP_AMQPURI environment variable.
		amqpURI := getAmqpURI(amqpURIs, i)
		if amqpURI == "" {
			return result, fmt.Errorf("--uri omitted but RABTAP_AMQPURI not set in environment")
		}
		tapConfig, err := rabtap.NewTapConfiguration(amqpURI, exchange)
		if err != nil {
			return result, err
		}
		result.TapConfig = append(result.TapConfig, *tapConfig)
	}
	return result, nil
}

// ParseCommandLineArgs parses command line arguments into an object of
// type CommandLineArgs.
func ParseCommandLineArgs(cliArgs []string) (CommandLineArgs, error) {
	args, err := docopt.ParseArgs(usage, cliArgs, RabtapAppVersion)
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
	}
	return CommandLineArgs{}, fmt.Errorf("command missing")
}
