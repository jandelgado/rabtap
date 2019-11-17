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

// createMessageReceiveFuncJSON returns a function that processes received
// messages as JSON messages
// TODO make testable (filename creation) and write test
func createMessageReceiveFuncJSON(out io.Writer, opts MessageReceiveFuncOptions) MessageReceiveFunc {
	return func(message rabtap.TapMessage) error {
		err := WriteMessageJSON(out, message)
		if err != nil || opts.optSaveDir == nil {
			return err
		}
		filename := path.Join(*opts.optSaveDir,
			fmt.Sprintf("rabtap-%d.json", time.Now().UnixNano()))
		return SaveMessageToJSONFile(filename, message)
	}
}

// createMessageReceiveFuncRaw returns a function that processes received
// messages as "raw" messages
// TODO make testable (filename creation) and write test
func createMessageReceiveFuncRaw(out io.Writer, opts MessageReceiveFuncOptions) MessageReceiveFunc {

	return func(message rabtap.TapMessage) error {
		err := PrettyPrintMessage(out, message, opts.noColor)
		if err != nil || opts.optSaveDir == nil {
			return err
		}
		basename := path.Join(*opts.optSaveDir,
			fmt.Sprintf("rabtap-%d", time.Now().UnixNano()))
		return SaveMessageToRawFile(basename, message)
	}
}

func createMessageReceiveFunc(out io.Writer, opts MessageReceiveFuncOptions) (MessageReceiveFunc, error) {

	switch opts.format {
	case "json":
		return createMessageReceiveFuncJSON(out, opts), nil
	case "raw":
		return createMessageReceiveFuncRaw(out, opts), nil
	}
	return nil, fmt.Errorf("invalid format %s", opts.format)
}
