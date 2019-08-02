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
			if message.Error != nil {
				// unrecoverable error received -> log and exit
				log.Error(message.Error)
				return message.Error
			}
			// let the receiveFunc do the actual message processing
			if err := messageReceiveFunc(message); err != nil {
				log.Error(err)
			}
		}
	}
}

// createMessageReceiveFuncJSON returns a function that processes received
// messages as JSON messages
// TODO make easier testable (filename creation) and write test
func createMessageReceiveFuncJSON(out io.Writer, optSaveDir *string,
	_ /* noColor */ bool) MessageReceiveFunc {
	return func(message rabtap.TapMessage) error {
		err := WriteMessageJSON(out, message)
		if err != nil || optSaveDir == nil {
			return err
		}
		filename := path.Join(*optSaveDir,
			fmt.Sprintf("rabtap-%d.json", time.Now().UnixNano()))
		return SaveMessageToJSONFile(filename, message)
	}
}

// createMessageReceiveFuncRaw returns a function that processes received
// messages as "raw" messages
// TODO make easier testable (filename creation) and write test
func createMessageReceiveFuncRaw(out io.Writer, optSaveDir *string,
	noColor bool) MessageReceiveFunc {

	return func(message rabtap.TapMessage) error {
		err := PrettyPrintMessage(out, message, noColor)
		if err != nil || optSaveDir == nil {
			return err
		}
		basename := path.Join(*optSaveDir,
			fmt.Sprintf("rabtap-%d", time.Now().UnixNano()))
		return SaveMessageToRawFile(basename, message)
	}
}

func createMessageReceiveFunc(out io.Writer, jsonFormat bool,
	optSaveDir *string, noColor bool) MessageReceiveFunc {

	if jsonFormat {
		return createMessageReceiveFuncJSON(out, optSaveDir, noColor)
	}
	return createMessageReceiveFuncRaw(out, optSaveDir, noColor)
}
