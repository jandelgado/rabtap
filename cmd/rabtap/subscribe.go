// subscribe to message producers
// Copyright (C) 2017-2021 Jan Delgado

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"time"

	rabtap "github.com/jandelgado/rabtap/pkg"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ErrIdleTimeout is returned by the message loop when the loop was terminated
// due to an timeout when no message was received
var ErrIdleTimeout = fmt.Errorf("idle timeout")

// FilenameProvider returns a filename to save a subscribed message to.
type FilenameProvider func() string

type AcknowledgeFunc func(rabtap.TapMessage) error

type MessageSinkOptions struct {
	out              io.Writer
	format           string // currently: raw, json, json-nopp
	silent           bool
	optSaveDir       *string
	filenameProvider FilenameProvider
}

// MessageSink processes received messages
type MessageSink func(rabtap.TapMessage) error

// var ErrMessageLoopEnded = errors.New("message loop ended")

func createMessagePredEnv(msg rabtap.TapMessage, count int64) map[string]interface{} {
	return map[string]interface{}{
		"msg":   msg.AmqpMessage,
		"count": count,
		"toStr": func(b []byte) string { return string(b) },
		"gunzip": func(b []byte) ([]byte, error) {
			return gunzip(bytes.NewReader(b))
		},
		"body": func(m *amqp.Delivery) ([]byte, error) {
			return body(m)
		},
	}
}

// loopCountPred creates is the default message loop
// termination predicate (loop terminates when predicate is true). When limit
// is 0, loop will never terminate. Expectes a variable "count" in the
// context, that holds the current number of messages received. The limit is
// provided by configuration. To unify predicate handling (see filter
// predicate), we use the same mechanism here. In later versions, the
// termination predicate may be defined by the user, so that rabtap quits if a
// certain condition is met.
type LoopCountPred struct {
	limit int64
}

func (s *LoopCountPred) Eval(env map[string]interface{}) (bool, error) {
	count := env["count"].(int64)
	return (s.limit > 0) && (count >= s.limit), nil
}

func NewLoopCountPred(limit int64) (*LoopCountPred, error) {
	return &LoopCountPred{limit}, nil
}

// CreateAcknowledgeFunc returns the function used to acknowledge received
// functions, wich will either be ACKed or REJECTED with optional REQUEUE
// flag set.
func CreateAcknowledgeFunc(reject, requeue bool) AcknowledgeFunc {
	return func(message rabtap.TapMessage) error {
		if reject {
			if err := message.AmqpMessage.Reject(requeue); err != nil {
				return fmt.Errorf("REJECT failed: %w", err)
			}
		} else {
			if err := message.AmqpMessage.Ack(false); err != nil {
				return fmt.Errorf("ACK failed: %w", err)
			}
		}
		return nil
	}
}

// MessageReceiveLoop passes received AMQP messages to the messageSink and
// handles errors received on the errorChan. AMQP messages are ascknowledged by
// the provides acknowleder function. Each message is passed to the predicate
// termPred function. If true is returned, processing is ended. Timeout
// specifies an idle timeout, which will end processing when for the given
// duration no new messages are received on messageChan.
// TODO pass in struct, limit number of arguments
func MessageReceiveLoop(ctx context.Context,
	messageChan rabtap.TapChannel,
	errorChan rabtap.SubscribeErrorChannel,
	messageSink MessageSink,
	filterPred Predicate,
	termPred Predicate,
	acknowledger AcknowledgeFunc,
	timeout time.Duration) error {

	timeoutTicker := time.NewTicker(timeout)
	defer timeoutTicker.Stop()

	count := int64(0) // counts not filtered messages
	for {
		select {

		case <-ctx.Done():
			log.Debugf("subscribe: cancel")
			return ctx.Err()

		case err, more := <-errorChan:
			if more {
				log.Errorf("subscribe: %v", err)
			}

		case message, more := <-messageChan:
			timeoutTicker.Reset(timeout)
			if !more {
				log.Debug("subscribe: messageReceiveLoop: channel closed.")
				return nil // ErrMessageLoopEnded
			}
			log.Debugf("subscribe: messageReceiveLoop: new message %+v", message)

			// acknowledge or reject the message
			if err := acknowledger(message); err != nil {
				log.Error(err)
			}

			env := createMessagePredEnv(message, count)
			passed, err := filterPred.Eval(env)
			if err != nil {
				log.Errorf("filter expression evaluation: %s", err.Error())
			}

			if !passed {
				log.Debugf("message with MessageId=%s was filtered out", message.AmqpMessage.MessageId)
				continue
			}
			count += 1

			if err := messageSink(message); err != nil {
				log.Error(err)
			}

			env = createMessagePredEnv(message, count)
			terminate, err := termPred.Eval(env)
			if err != nil {
				log.Errorf("terminate expression evaluation: %s", err.Error())
			}
			if terminate {
				return nil
			}

		case <-timeoutTicker.C:
			return ErrIdleTimeout
		}
	}
}

// nopMessageSink is used a sentinel to terminate a chain of
// MessageReceiveFuncs
func nopMessageSink(rabtap.TapMessage) error {
	return nil
}

func messageSinkTee(first, second MessageSink) MessageSink {
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
func newWriteToRawFileMessageSink(dir string, marshaller marshalFunc, filenameProvider FilenameProvider) MessageSink {
	return func(message rabtap.TapMessage) error {
		basename := path.Join(dir, filenameProvider())
		return SaveMessageToRawFiles(basename, message, marshaller)
	}
}

// creatmMessageReceiveFuncWriteToJSONFile return receive func that writes the
// message to a file in the provided directory using the provided marshaller.
func newWriteToJSONFileMessageSink(dir string, marshaller marshalFunc, filenameProvider FilenameProvider) MessageSink {
	return func(message rabtap.TapMessage) error {
		filename := path.Join(dir, filenameProvider()+".json")
		return SaveMessageToJSONFile(filename, message, marshaller)
	}
}

// newPrintJSONMessageSink returns a function that prints messages as
// JSON to the provided writer
// messages as JSON messages
func newPrintJSONMessageSink(out io.Writer, marshaller marshalFunc) MessageSink {
	return func(message rabtap.TapMessage) error {
		return WriteMessage(out, message, marshaller)
	}
}

// newPrettyPrintJSONMessageSink returns a function that pretty prints
// received messaged to the provided writer
func newPrettyPrintJSONMessageSink(out io.Writer) MessageSink {
	return func(message rabtap.TapMessage) error {
		return PrettyPrintMessage(out, message)
	}
}

func newPrintMessageMessageSink(format string, out io.Writer, silent bool) (MessageSink, error) {
	if silent {
		return nopMessageSink, nil
	}

	switch format {
	case "json-nopp":
		return newPrintJSONMessageSink(out, JSONMarshal), nil
	case "json":
		return newPrintJSONMessageSink(out, JSONMarshalIndent), nil
	case "raw":
		return newPrettyPrintJSONMessageSink(out), nil
	default:
		return nil, fmt.Errorf("invalid format %s", format)
	}
}

func newSaveFileMessageSink(format string, optSaveDir *string, filenameProvider FilenameProvider) (MessageSink, error) {
	if optSaveDir == nil {
		return nopMessageSink, nil
	}

	switch format {
	case "json-nopp":
		fallthrough
	case "json":
		return newWriteToJSONFileMessageSink(*optSaveDir, JSONMarshalIndent, filenameProvider), nil
	case "raw":
		return newWriteToRawFileMessageSink(*optSaveDir, JSONMarshalIndent, filenameProvider), nil
	default:
		return nil, fmt.Errorf("invalid format %s", format)
	}
}

// NewMessageSink returns a MessageReceiveFunc which is invoked on
// receival of a message during tap and subscribe. Depending on the options
// set, function that optionally prints to the proviced io.Writer and
// optionally to the provided directory is returned.
func NewMessageSink(opts MessageSinkOptions) (MessageSink, error) {
	printFunc, err := newPrintMessageMessageSink(opts.format, opts.out, opts.silent)
	if err != nil {
		return printFunc, err
	}
	saveFunc, err := newSaveFileMessageSink(opts.format, opts.optSaveDir, opts.filenameProvider)
	return messageSinkTee(printFunc, saveFunc), err
}
