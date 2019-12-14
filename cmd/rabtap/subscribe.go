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
	format     string // currently: raw, json
	noColor    bool
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

func createMessageReceiveFunc(out io.Writer, opts MessageReceiveFuncOptions) (MessageReceiveFunc, error) {

	var printFunc MessageReceiveFunc
	saveFunc := NullMessageReceiveFunc

	switch opts.format {
	case "json":
		printFunc = createMessageReceiveFuncJSON(out, JSONMarshalIndent)
		if opts.optSaveDir != nil {
			saveFunc = createMessageReceiveFuncWriteToJSONFile(*opts.optSaveDir, JSONMarshalIndent)
		}
	case "raw":
		printFunc = createMessageReceiveFuncPrettyPrint(out, opts.noColor)
		if opts.optSaveDir != nil {
			saveFunc = createMessageReceiveFuncWriteToRawFiles(*opts.optSaveDir, JSONMarshalIndent)
		}
	default:
		return nil, fmt.Errorf("invalid format %s", opts.format)
	}
	return chainedMessageReceiveFunc(printFunc, saveFunc), nil
}
