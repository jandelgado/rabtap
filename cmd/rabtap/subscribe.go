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

type MessageReceiveFuncOptions struct {
	out              io.Writer
	format           string // currently: raw, json, json-nopp
	silent           bool
	optSaveDir       *string
	filenameProvider FilenameProvider
}

// MessageReceiveFunc processes receiced messages from a tap.
type MessageReceiveFunc func(rabtap.TapMessage) error

// var ErrMessageLoopEnded = errors.New("message loop ended")

// MessagePred is a predicate function on a message
type MessagePred func(rabtap.TapMessage) (bool, error)

type MessagePredEnv struct {
	msg    *amqp.Delivery
	count  int64
	toStr  func([]byte) string
	gunzip func([]byte) ([]byte, error)
}

func createMessagePredEnv(msg rabtap.TapMessage, count int64) map[string]interface{} {
	return map[string]interface{}{
		"rt_msg":   msg.AmqpMessage,
		"rt_count": count,
		"rt_toStr": func(b []byte) string { return string(b) },
		"rt_gunzip": func(b []byte) ([]byte, error) {
			return gunzip(bytes.NewReader(b))
		},
	}
}

// cerateMessagePredEnv returns an environment to evaluate predicates in the
// context of received messages
// func createMessagePredEnv(msg rabtap.TapMessage, count int64) map[string]interface{} {
/* func createMessagePredEnv(msg rabtap.TapMessage, count int64) MessagePredEnv {
	// expose the message and some helper function
	return MessagePredEnv{
		msg:   msg.AmqpMessage,
		count: count,
		toStr: func(b []byte) string { return string(b) },
		gunzip: func(b []byte) ([]byte, error) {
			return gunzip(bytes.NewReader(b))
		},
	}
} */

// createMessagePred creates a MessagePred predicate function that uses a
// PredicateExpression
/* func createMessagePred(expr Predicate) MessagePred {
	return func(m rabtap.TapMessage) (bool, error) {
		// expose the message and some helper function
		env := map[string]interface{}{
			"msg":   m.AmqpMessage,
			"toStr": func(b []byte) string { return string(b) },
			"gunzip": func(b []byte) ([]byte, error) {
				return gunzip(bytes.NewReader(b))
			},
		}
		return expr.Eval(env)
	}
} */

// createCountingMessageReceivePred creates the default message loop termination
// predicate (loop terminates when predicate is true). When limit is 0, loop
// will never terminate. Expectes a variable "rt_count" in the context, that
// holds the current number of messages received. The limit is provided by configuration.
// To unify predicate handling (see filter predicate), we use the same mechanism
// here. In later versions, the termination predicate may be defined by the
// user, so that rabtap quits if a certain condition is met.
func createCountingMessageReceivePred(limit int64) (Predicate, error) {
	env := map[string]interface{}{"rt_limit": limit}
	return NewExprPredicateWithEnv("(rt_limit > 0) && (rt_count >= rt_limit)", env)
}

/* func createCountingMessageReceivePred(num int64) MessagePred {

	if num == 0 {
		return func(_ rabtap.TapMessage) (bool, error) {
			return false, nil
		}
	}

	counter := int64(1)
	return func(_ rabtap.TapMessage) (bool, error) {
		counter++
		return counter > num, nil

	}
} */

// createAcknowledgeFunc returns the function used to acknowledge received
// functions, wich will either be ACKed or REJECTED with optional REQUEUE
// flag set.
func createAcknowledgeFunc(reject, requeue bool) AcknowledgeFunc {
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

// messageReceiveLoop passes received AMQP messages to messageReceiveFunc and
// handles errors received on the errorChan. AMQP messages are ascknowledged by
// the provides acknowleder function. Each message is passed to the predicate
// termPred function. If true is returned, processing is ended. Timeout
// specifies an idle timeout, which will end processing when for the given
// duration no new messages are received on messageChan.
// TODO pass in struct, limit number of arguments
func messageReceiveLoop(ctx context.Context,
	messageChan rabtap.TapChannel,
	errorChan rabtap.SubscribeErrorChannel,
	messageReceiveFunc MessageReceiveFunc,
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

			if err := messageReceiveFunc(message); err != nil {
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
func createMessageReceiveFuncPrintPretty(out io.Writer) MessageReceiveFunc {
	return func(message rabtap.TapMessage) error {
		return PrettyPrintMessage(out, message)
	}
}

func createMessageReceivePrintFunc(format string, out io.Writer, silent bool) (MessageReceiveFunc, error) {
	if silent {
		return NullMessageReceiveFunc, nil
	}

	switch format {
	case "json-nopp":
		return createMessageReceiveFuncPrintJSON(out, JSONMarshal), nil
	case "json":
		return createMessageReceiveFuncPrintJSON(out, JSONMarshalIndent), nil
	case "raw":
		return createMessageReceiveFuncPrintPretty(out), nil
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

	printFunc, err := createMessageReceivePrintFunc(opts.format, opts.out, opts.silent)
	if err != nil {
		return printFunc, err
	}
	saveFunc, err := createMessageReceiveSaveFunc(opts.format, opts.optSaveDir, opts.filenameProvider)
	return chainedMessageReceiveFunc(printFunc, saveFunc), err
}
