// read persisted metadata and messages from a directory
// Copyright (C) 2019 Jan Delgado

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"

	"github.com/streadway/amqp"
)

const metadataFilePattern = `^rabtap-[0-9]+.json$`

type DirReader func(string) ([]os.FileInfo, error)
type FileInfoPredicate func(fileinfo os.FileInfo) bool

type FilenameWithMetadata struct {
	filename string
	metadata RabtapPersistentMessage
}

// newRabtapFileInfoPredicate returns a FileInfoPredicate that matches
// rabtap metadata files
func NewRabtapFileInfoPredicate() FileInfoPredicate {
	filenameRe := regexp.MustCompile(metadataFilePattern)
	return func(fi os.FileInfo) bool {
		return fi.Mode().IsRegular() && filenameRe.MatchString(fi.Name())
	}
}

func filterMetadataFilenames(fileinfos []os.FileInfo, pred FileInfoPredicate) []string {
	var filenames []string
	for _, fi := range fileinfos {
		if pred(fi) {
			filenames = append(filenames, fi.Name())
		}
	}
	return filenames
}

// findMetadataFilenames returns list of absolute filenames looking like rabtap
// persisted message/metadata files
func findMetadataFilenames(dirname string, dirReader DirReader, pred FileInfoPredicate) ([]string, error) {
	fileinfos, err := dirReader(dirname)
	if err != nil {
		return nil, err
	}
	return filterMetadataFilenames(fileinfos, pred), nil
}

func readRabtapPersistentMessage(filename string) (RabtapPersistentMessage, error) {
	file, err := os.Open(filename)
	if err != nil {
		return RabtapPersistentMessage{}, err
	}
	defer file.Close()
	return readMessageFromJSON(file)
}

// readMetadataOfFiles reads all metadata files from the given list of files.
// returns an error if any error occurs.
func readMetadataOfFiles(dirname string, filenames []string) ([]FilenameWithMetadata, error) {

	data := make([]FilenameWithMetadata, len(filenames))
	for i, filename := range filenames {
		fullpath := path.Join(dirname, filename)
		msg, err := readRabtapPersistentMessage(fullpath)
		if err != nil {
			return data, err
		}
		// we will load the body later when the message is published (from the
		// JSON or a separate message file). This approach read message bodies
		// twice, but this should not be a problem (for now...)
		msg.Body = []byte("")
		data[i] = FilenameWithMetadata{filename: fullpath, metadata: msg}
	}

	return data, nil
}

// LoadMetadataFromDir loads all metadata files from the given directory
// passing the given predicate
func LoadMetadataFilesFromDir(dirname string, dirReader DirReader, pred FileInfoPredicate) ([]FilenameWithMetadata, error) {
	filenames, err := findMetadataFilenames(dirname, dirReader, pred)
	if err != nil {
		return nil, err
	}
	return readMetadataOfFiles(dirname, filenames)
}

// createMessageFromDirReaderFunc returns a MessageReaderFunc that reads
// messages from the given list of filenames in the given format.
func CreateMessageFromDirReaderFunc(format string, files []FilenameWithMetadata) (MessageReaderFunc, error) {

	i := 0

	switch format {
	case "json-nopp":
		fallthrough
	case "json":
		return func() (amqp.Publishing, bool, error) {
			fullMessage, err := readRabtapPersistentMessage(files[i].filename)
			i++
			return fullMessage.ToAmqpPublishing(), i < len(files), err
		}, nil
	case "raw":
		return func() (amqp.Publishing, bool, error) {
			body, err := ioutil.ReadFile(files[i].filename)
			message := files[i].metadata
			message.Body = body
			i++
			return message.ToAmqpPublishing(), i < len(files), err
		}, nil
	}
	return nil, fmt.Errorf("invaild format %s", format)
}
