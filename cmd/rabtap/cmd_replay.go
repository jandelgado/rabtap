// replay a directory of previously recorded messages. Message metadata
// is read from JSON files. Message bodies are read from JSON or separatly
// stored message files.
// Copyright (C) 2019 Jan Delgado

package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"

	rabtap "github.com/jandelgado/rabtap/pkg"
	"golang.org/x/sync/errgroup"
)

const metadataFilePattern = `^rabtap-[0-9]+.json$`

// CmdReplayArg contains arguments for the replay command
type CmdReplayArg struct {
	amqpURI   string
	tlsConfig *tls.Config
	basedir   string
}

type replayMetadata struct {
	filename    string
	messageData RabtapPersistentMessage
}

type DirReader func(string) ([]os.FileInfo, error)

func filterMetadataFilenames(fileinfos []os.FileInfo) []string {
	filenameRegexp := regexp.MustCompile(metadataFilePattern)

	var filenames []string
	for _, fi := range fileinfos {
		if fi.Mode().IsRegular() && filenameRegexp.MatchString(fi.Name()) {
			filenames = append(filenames, fi.Name())
		}
	}
	return filenames
}

// findMetadataFilenames returns list of files looking like a rabtap
// persisted message/metadata files
func findMetadataFilenames(dirname string, dirReader DirReader) ([]string, error) {

	fileinfos, err := dirReader(dirname)
	if err != nil {
		return nil, err
	}
	return filterMetadataFilenames(fileinfos), nil
}

func readRabtapPersistenMessage(filename string) (RabtapPersistentMessage, error) {
	var data RabtapPersistentMessage

	file, err := os.Open(filename)
	if err != nil {
		return data, err
	}
	defer file.Close()

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(contents, &data)

	return data, err
}

// readMetadataOfFiles reads all metadata files from the given list of files
// Will return an error if any error occurs.
func readMetadataOfFiles(filenames []string) ([]replayMetadata, error) {

	metadata := make([]replayMetadata, len(filenames))
	for i, filename := range filenames {
		msg, err := readRabtapPersistenMessage(filename)
		if err != nil {
			return metadata, err
		}
		// we will load the body later when the message is published (from the
		// JSON or a separate message file). This approach read message bodies
		// twice, but this should not be a problem (for now...)
		msg.Body = []byte("")
		metadata[i] = replayMetadata{filename: filename, messageData: msg}
	}

	return metadata, nil
}

/***
// publishMessage publishes a single message on the given exchange with the
// provided routingkey
// TODO separate file ?
func publishMessage(publishChannel rabtap.PublishChannel,
	exchange, routingKey string,
	amqpPublishing amqp.Publishing) {

	log.Debugf("publishing message %+v to exchange %s with routing key %s",
		amqpPublishing, exchange, routingKey)

	publishChannel <- &rabtap.PublishMessage{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Publishing: &amqpPublishing}
}

// publishMessages reads messages with the provided readNextMessageFunc and
// publishes the messages to the given exchange. When done closes
// the publishChannel
func replayMessageStream(publishChannel rabtap.PublishChannel,
	exchange, routingKey string, readNextMessageFunc MessageReaderFunc) error {
	for {
		msg, more, err := readNextMessageFunc()
		switch err {
		case io.EOF:
			close(publishChannel)
			return nil
		case nil:
			publishMessage(publishChannel, exchange, routingKey, msg)
		default:
			close(publishChannel)
			return err
		}

		if !more {
			close(publishChannel)
			return nil
		}
	}
}
***/

// cmdPublish reads messages with the provied readNextMessageFunc and
// publishes the messages to the given exchange.
// Termination is a little bit tricky here, since we can not use "select"
// on a File object to stop a blocking read. There are 3 ways publishing
// can be stopped:
// * by an EOF or error on the input file
// * by ctx.Context() signaling cancellation (e.g. ctrl+c)
// * by an initial connection failure to the broker
func cmdReplay(ctx context.Context, cmd CmdReplayArg) error {

	g, ctx := errgroup.WithContext(ctx)

	publisher := rabtap.NewAmqpPublish(cmd.amqpURI, cmd.tlsConfig, log)
	publishChannel := make(rabtap.PublishChannel)

	reader := func() error {
		// TODO
		return nil
	}

	g.Go(reader)
	g.Go(func() error {
		return publisher.EstablishConnection(ctx, publishChannel)
	})

	return g.Wait()
}
