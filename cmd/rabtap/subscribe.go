// Copyright (C) 2017-2019 Jan Delgado

package main

// common functionality to subscribe to queues.

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

type MessageReceiveFuncOptions struct {
	out        io.Writer
	format     string // currently: raw, json, json-nopp
	noColor    bool
	silent     bool
	optSaveDir *string
}

// MessageReceiveFunc processes receiced messages from a tap.
type MessageReceiveFunc func(rabtap.TapMessage) error

func messageReceiveLoop(ctx context.Context, messageChan rabtap.TapChannel,
	messageReceiveFunc MessageReceiveFunc) error {

	for {
		select {
		case <-ctx.Done():
			log.Debugf("subscribe: cancel")
			return nil

		case message, more := <-messageChan:
			if !more {
				log.Debug("subscribe: messageReceiveLoop: channel closed.")
				return nil
			}
			log.Debugf("subscribe: messageReceiveLoop: new message %#+v", message)
			tmpCh := make(rabtap.TapChannel)
			go func() {
				m := <-tmpCh
				// let the receiveFunc do the actual message processing
				if err := messageReceiveFunc(m); err != nil {
					log.Error(err)
				}
			}()
			select {
			case tmpCh <- message:
			case <-ctx.Done():
				log.Debugf("subscribe: cancel (messageReceiveFunc)")
				return nil
			}
		}
	}
}

// NullMessageReceiveFunc is used a sentinel to terminal a chain of
// MessageReceiveFuncs
func NullMessageReceiveFunc(rabtap.TapMessage) error {
	return nil
}

func chainedMessageReceiveFunc(first MessageReceiveFunc, second MessageReceiveFunc) MessageReceiveFunc {
	return func(message rabtap.TapMessage) error {
		if err := first(message); err != nil {
			return err
		}
		return second(message)
	}
}

// createMessageReceiveFuncWriteToJSONFile return receive func that writes the
// message and metadata to separate files in the provided directory using the
// provided marshaller.
// TODO make testable (filename creation) and write test
func createMessageReceiveFuncWriteToRawFiles(dir string, marshaller marshalFunc) MessageReceiveFunc {
	return func(message rabtap.TapMessage) error {
		basename := path.Join(dir,
			fmt.Sprintf("rabtap-%d", time.Now().UnixNano()))
		return SaveMessageToRawFiles(basename, message, marshaller)
	}
}

// createMessageReceiveFuncWriteToJSONFile return receive func that writes the
// message to a file in the provided directory using the provided marshaller.
// TODO make testable (filename creation) and write test
func createMessageReceiveFuncWriteToJSONFile(dir string, marshaller marshalFunc) MessageReceiveFunc {
	return func(message rabtap.TapMessage) error {
		filename := path.Join(dir,
			fmt.Sprintf("rabtap-%d.json", time.Now().UnixNano()))
		return SaveMessageToJSONFile(filename, message, marshaller)
	}
}

// createMessageReceiveFuncJSON returns a function that prints messages as
// JSON to the provided writer
// messages as JSON messages
func createMessageReceiveFuncJSON(out io.Writer, marshaller marshalFunc) MessageReceiveFunc {
	return func(message rabtap.TapMessage) error {
		return WriteMessage(out, message, marshaller)
	}
}

// createMessageReceiveFuncPrettyPrint returns a function that pretty prints
// received messaged to the provided writer
func createMessageReceiveFuncPrettyPrint(out io.Writer, noColor bool) MessageReceiveFunc {
	return func(message rabtap.TapMessage) error {
		return PrettyPrintMessage(out, message, noColor)
	}
}

func createMessageReceivePrintFunc(format string, out io.Writer, noColor bool, silent bool) (MessageReceiveFunc, error) {
	if silent {
		return NullMessageReceiveFunc, nil
	}

	switch format {
	case "json-nopp":
		return createMessageReceiveFuncJSON(out, JSONMarshal), nil
	case "json":
		return createMessageReceiveFuncJSON(out, JSONMarshalIndent), nil
	case "raw":
		return createMessageReceiveFuncPrettyPrint(out, noColor), nil
	default:
		return nil, fmt.Errorf("invalid format %s", format)
	}
}

func createMessageReceiveSaveFunc(format string, optSaveDir *string) (MessageReceiveFunc, error) {
	if optSaveDir == nil {
		return NullMessageReceiveFunc, nil
	}

	switch format {
	case "json-nopp":
		fallthrough
	case "json":
		return createMessageReceiveFuncWriteToJSONFile(*optSaveDir, JSONMarshalIndent), nil
	case "raw":
		return createMessageReceiveFuncWriteToRawFiles(*optSaveDir, JSONMarshalIndent), nil
	default:
		return nil, fmt.Errorf("invalid format %s", format)
	}
}

// createMessageReceiveFunc returns a MessageReceiveFunc which is invoked on
// receival of a message during tap and subscribe. Depending on the options
// set, function that optionally prints to the proviced io.Writer and
// optionally to the provided directory is returned.
func createMessageReceiveFunc(opts MessageReceiveFuncOptions) (MessageReceiveFunc, error) {

	printFunc, err := createMessageReceivePrintFunc(opts.format, opts.out, opts.noColor, opts.silent)
	if err != nil {
		return printFunc, err
	}
	saveFunc, err := createMessageReceiveSaveFunc(opts.format, opts.optSaveDir)
	return chainedMessageReceiveFunc(printFunc, saveFunc), err
}
