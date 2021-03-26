// subscribe to message producers
// Copyright (C) 2017-2021 Jan Delgado

package main

import (
	"context"
	"fmt"
	"io"
	"path"

	rabtap "github.com/jandelgado/rabtap/pkg"
)

// FilenameProvider returns a filename to save a subscribed message to.
type FilenameProvider func() string

type MessageReceiveFuncOptions struct {
	out              io.Writer
	format           string // currently: raw, json, json-nopp
	noColor          bool
	silent           bool
	optSaveDir       *string
	filenameProvider FilenameProvider
}

// MessageReceiveFunc processes receiced messages from a tap.
type MessageReceiveFunc func(rabtap.TapMessage) error

//var ErrMessageLoopEnded = errors.New("message loop ended")

// messageReceiveLoopPred is called once for each a message that was received.
// If it returns true, the subscriber loop continues, otherwise the loop
// terminates.
type MessageReceiveLoopPred func(rabtap.TapMessage) bool

// createCountingMessageReceivePred returns a (stateful) predicate that will
// return false after num messages where successfully received, thus limiting
// the number of messages received
func createCountingMessageReceivePred(num int) MessageReceiveLoopPred {
	counter := 0
	return func(_ rabtap.TapMessage) bool {
		counter++
		return counter <= num
	}
}

// continueMessageReceivePred will never return false, resulting in an unlimited
// receival of messages.
func continueMessageReceivePred(msg rabtap.TapMessage) bool {
	return true
}

func messageReceiveLoop(ctx context.Context,
	messageChan rabtap.TapChannel,
	messageReceiveFunc MessageReceiveFunc,
	pred MessageReceiveLoopPred) error {

	//wg := new(sync.WaitGroup)

	for {
		// if pred(count) {
		//     //return ErrMessageLoopEnded
		//     return nil
		// }

		select {
		case <-ctx.Done():
			log.Debugf("subscribe: cancel")
			return nil

		case message, more := <-messageChan:
			if !more {
				log.Debug("subscribe: messageReceiveLoop: channel closed.")
				return nil //ErrMessageLoopEnded
			}
			log.Debugf("subscribe: messageReceiveLoop: new message %+v", message)

			if err := messageReceiveFunc(message); err != nil {
				log.Error(err)
			}
			if !pred(message) {
				return nil
			}

			/*
				// do actual output in a go-routine so we can terminate using
				// ctrl+c/ctx.Done() even if the terminal output is stopped with ctrl+s
				tmpCh := make(rabtap.TapChannel)
				go func() {
					wg.Wait()
					wg.Add(1)
					m := <-tmpCh
					// let the receiveFunc do the actual message processing
					var err error
					if err = messageReceiveFunc(m); err != nil {
						log.Error(err)
					}
					wg.Done()
				}()
				select {
				case tmpCh <- message:
				case <-ctx.Done():
					log.Debugf("subscribe: cancel (messageReceiveFunc)")
					return nil
				}*/
		}
	}
}

// NullMessageReceiveFunc is used a sentinel to terminate a chain of
// MessageReceiveFuncs
func NullMessageReceiveFunc(rabtap.TapMessage) error {
	return nil
}

func chainedMessageReceiveFunc(first, second MessageReceiveFunc) MessageReceiveFunc {
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
func createMessageReceiveFuncWriteToRawFiles(dir string, marshaller marshalFunc, filenameProvider FilenameProvider) MessageReceiveFunc {
	return func(message rabtap.TapMessage) error {
		basename := path.Join(dir, filenameProvider())
		return SaveMessageToRawFiles(basename, message, marshaller)
	}
}

// createMessageReceiveFuncWriteToJSONFile return receive func that writes the
// message to a file in the provided directory using the provided marshaller.
func createMessageReceiveFuncWriteToJSONFile(dir string, marshaller marshalFunc, filenameProvider FilenameProvider) MessageReceiveFunc {
	return func(message rabtap.TapMessage) error {
		filename := path.Join(dir, filenameProvider()+".json")
		return SaveMessageToJSONFile(filename, message, marshaller)
	}
}

// createMessageReceiveFuncPrintJSON returns a function that prints messages as
// JSON to the provided writer
// messages as JSON messages
func createMessageReceiveFuncPrintJSON(out io.Writer, marshaller marshalFunc) MessageReceiveFunc {
	return func(message rabtap.TapMessage) error {
		return WriteMessage(out, message, marshaller)
	}
}

// createMessageReceiveFuncPrintPretty returns a function that pretty prints
// received messaged to the provided writer
func createMessageReceiveFuncPrintPretty(out io.Writer, noColor bool) MessageReceiveFunc {
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
		return createMessageReceiveFuncPrintJSON(out, JSONMarshal), nil
	case "json":
		return createMessageReceiveFuncPrintJSON(out, JSONMarshalIndent), nil
	case "raw":
		return createMessageReceiveFuncPrintPretty(out, noColor), nil
	default:
		return nil, fmt.Errorf("invalid format %s", format)
	}
}

func createMessageReceiveSaveFunc(format string, optSaveDir *string, filenameProvider FilenameProvider) (MessageReceiveFunc, error) {
	if optSaveDir == nil {
		return NullMessageReceiveFunc, nil
	}

	switch format {
	case "json-nopp":
		fallthrough
	case "json":
		return createMessageReceiveFuncWriteToJSONFile(*optSaveDir, JSONMarshalIndent, filenameProvider), nil
	case "raw":
		return createMessageReceiveFuncWriteToRawFiles(*optSaveDir, JSONMarshalIndent, filenameProvider), nil
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
	saveFunc, err := createMessageReceiveSaveFunc(opts.format, opts.optSaveDir, opts.filenameProvider)
	return chainedMessageReceiveFunc(printFunc, saveFunc), err
}
