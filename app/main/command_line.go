// Copyright (C) 2017 Jan Delgado

package main

import (
	"fmt"
	"os"

	"github.com/docopt/docopt-go"
	"github.com/jandelgado/rabtap"
)

const (
	usage = `rabtap - RabbitMQ message tap.

Usage:
  rabtap tap [--uri URI] EXCHANGES [--saveto=DIR] [-jkvn]
  rabtap (tap --uri URI EXCHANGES)... [--saveto=DIR] [-jkvn]
  rabtap pub [--uri URI] EXCHANGE [FILE] [--routingkey KEY] [-jkv]
  rabtap info [--api APIURI] [--consumers] [--stats] [--show-default] [-kvn]
  rabtap -h|--help

Examples:
  rabtap tap --uri amqp://guest:guest@localhost/ amq.fanout:
  rabtap tap --uri amqp://guest:guest@localhost/ amq.topic:#,amq.fanout:
  rabtap pub --uri amqp://guest:guest@localhost/ amq.topic message.json -j
  rabtap info --api http://guest:guest@localhost:15672/api

Options:
 -h, --help           print this help.
 --uri URI            connect to given AQMP broker. If omitted, the 
                      environment variable RABTAP_AMQPURI will be used.
 EXCHANGES            comma-separated list of exchanges and routing keys, 
                      e.g. amq.topic:# or exchange1:key1,exchange2:key2.
 EXCHANGE             name of an exchange, e.g. amq.direct.
 FILE                 file to publish in pub mode. If omitted, stdin will 
                      be read.
 --saveto DIR         also save messages and metadata to DIR.
 -j, --json           print/save/publish message metadata and body to a 
                      single JSON file. JSON body is base64 encoded. Otherwise 
                      metadata and body (as-is) are saved separately. 
 -r, --routingkey KEY routing key to use in publish mode.
 --api APIURI         connect to given API server. If APIURI is omitted, 
                      the environment variable RABTAP_APIURI will be used.
 -n, --no-color       don't colorize output.
 --consumers          include consumers in output of info command.
 --stats              include statistics in output of info command.
 --show-default       include default exchange in output info command.
 -k, --insecure       allow insecure TLS connections (no certificate check).
 -v, --verbose        enable verbose mode.
`
)

// ProgramMode represents the mode of operation
type ProgramMode int

const (
	// TapMode sets mode to tapping mode
	TapMode ProgramMode = iota
	// PubMode sets mode to publish messages
	PubMode
	// InfoMode shows info on exchanges and queues
	InfoMode
)

// CommandLineArgs represents the parsed command line arguments
type CommandLineArgs struct {
	Mode                ProgramMode
	TapConfig           []rabtap.TapConfiguration
	PubAmqpURI          string  // pub mode: broker to use
	PubExchange         string  // pub mode: exchange to publish to
	PubRoutingKey       string  // pub mode: routing key, defaults to ""
	PubFile             *string // pub mode: file to send
	APIURI              string
	Verbose             bool
	InsecureTLS         bool
	ShowConsumers       bool // info mode: also show consumer
	ShowStats           bool // info mode: also show statistics
	NoColor             bool
	SaveDir             *string // save mode: optional directory to stores files to
	JSONFormat          bool    // save/print meta + body to single JSON-file
	ShowDefaultExchange bool
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

func parseInfoArgs(args map[string]interface{}) (CommandLineArgs, error) {

	result := CommandLineArgs{
		Verbose:             args["--verbose"].(bool),
		InsecureTLS:         args["--insecure"].(bool),
		ShowConsumers:       args["--consumers"].(bool),
		ShowStats:           args["--stats"].(bool),
		NoColor:             args["--no-color"].(bool),
		ShowDefaultExchange: args["--show-default"].(bool)}

	result.Mode = InfoMode
	var apiURI string
	if args["--api"] != nil {
		apiURI = args["--api"].(string)
	} else {
		apiURI = os.Getenv("RABTAP_APIURI")
	}
	if apiURI == "" {
		return CommandLineArgs{}, fmt.Errorf("--api omitted but RABTAP_APIURI not set in environment")
	}
	result.APIURI = apiURI
	return result, nil
}

func parsePublishArgs(args map[string]interface{}) (CommandLineArgs, error) {

	result := CommandLineArgs{
		Verbose:     args["--verbose"].(bool),
		InsecureTLS: args["--insecure"].(bool),
		JSONFormat:  args["--json"].(bool)}

	result.Mode = PubMode
	amqpURIs := args["--uri"].([]string)
	amqpURI := getAmqpURI(amqpURIs, 0)
	if amqpURI == "" {
		return CommandLineArgs{}, fmt.Errorf("--uri omitted but RABTAP_AMQPURI not set in environment")
	}
	result.PubAmqpURI = amqpURI
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

func parseTapArgs(args map[string]interface{}) (CommandLineArgs, error) {

	result := CommandLineArgs{
		TapConfig:   []rabtap.TapConfiguration{},
		Verbose:     args["--verbose"].(bool),
		NoColor:     args["--no-color"].(bool),
		InsecureTLS: args["--insecure"].(bool),
		JSONFormat:  args["--json"].(bool)}

	result.Mode = TapMode
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
	args, err := docopt.Parse(usage, cliArgs, true, "", false)
	// fmt.Printf("%#+v", args)
	if err != nil {
		return CommandLineArgs{}, err
	}
	if args["tap"].(int) > 0 {
		return parseTapArgs(args)
	} else if args["info"].(bool) {
		return parseInfoArgs(args)
	} else if args["pub"].(bool) {
		return parsePublishArgs(args)
	}
	return CommandLineArgs{}, fmt.Errorf("command missing")
}
