// Copyright (C) 2017 Jan Delgado

package main

// common functionality to subscribe to queues.

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/jandelgado/rabtap"
	"github.com/streadway/amqp"
)

// MessageReceiveFunc processes receiced messages from a tap.
type MessageReceiveFunc func(*amqp.Delivery) error

func messageReceiveLoop(messageChan rabtap.TapChannel,
	messageReceiveFunc MessageReceiveFunc, signalChannel chan os.Signal) {

ReceiveLoop:
	for {
		select {
		case message := <-messageChan:
			log.Debugf("subscribe: messageReceiveLoop: new message %#+v", message)
			if message.Error != nil {
				// unrecoverable error received -> log and exit
				log.Error(message.Error)
				break ReceiveLoop
			}
			// let the receiveFunc do the actual message processing
			if err := messageReceiveFunc(message.AmqpMessage); err != nil {
				log.Error(err)
			}
		case <-signalChannel:
			break ReceiveLoop
		}
	}
}

// createMessageReceiveFuncJSON returns a function that processes received
// messages as JSON messages
// TODO make easier testable (filename creation) and write test
func createMessageReceiveFuncJSON(out io.Writer, optSaveDir *string,
	_ /* noColor */ bool) MessageReceiveFunc {
	return func(message *amqp.Delivery) error {
		err := WriteMessageJSON(out, true, message)
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

	return func(message *amqp.Delivery) error {
		err := PrettyPrintMessage(out, message,
			fmt.Sprintf("message received on %s",
				time.Now().Format(time.RFC3339)),
			noColor,
		)
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
