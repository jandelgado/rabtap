// Copyright (C) 2019-2022 Jan Delgado
package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
)

const metadataFilePattern = `^rabtap-[0-9]+.json$`

type (
	DirReader         func(string) ([]os.DirEntry, error)
	FileInfoPredicate func(entry os.DirEntry) bool
)

type FilenameWithMetadata struct {
	filename string
	metadata RabtapPersistentMessage
}

func filenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}

// NewRabtapFileInfoPredicate returns a FileInfoPredicate that matches
// rabtap metadata files
func NewRabtapFileInfoPredicate() FileInfoPredicate {
	filenameRe := regexp.MustCompile(metadataFilePattern)
	return func(entry os.DirEntry) bool {
		return entry.Type().IsRegular() && filenameRe.MatchString(entry.Name())
	}
}

func filterMetadataFilenames(fileinfos []os.DirEntry, pred FileInfoPredicate) []string {
	var filenames []string
	for _, entry := range fileinfos {
		if pred(entry) {
			filenames = append(filenames, entry.Name())
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
	contents, err := readMessageFromJSON(file)
	if err != nil {
		return RabtapPersistentMessage{}, fmt.Errorf("error reading %s: %v", filename, err)
	}
	return contents, nil
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
		// JSON or a separate message file). This approach reads message bodies
		// twice, but this should not be a problem
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

// NewReadFilesFromDirMessageSource returns a MessageProvicerFunc that reads
// messages from the given list of filenames in the given format.
func NewReadFilesFromDirMessageSource(format string, files []FilenameWithMetadata) (MessageSource, error) {
	curfile := 0

	switch format {
	case "json-nopp":
		fallthrough
	case "json":
		return func() (RabtapPersistentMessage, error) {
			var message RabtapPersistentMessage
			if curfile >= len(files) {
				return message, io.EOF
			}
			message, err := readRabtapPersistentMessage(files[curfile].filename)
			curfile++
			return message, err
		}, nil
	case "raw":
		return func() (RabtapPersistentMessage, error) {
			var message RabtapPersistentMessage
			if curfile >= len(files) {
				return message, io.EOF
			}
			rawFile := filenameWithoutExtension(files[curfile].filename) + ".dat"
			body, err := os.ReadFile(rawFile)
			message = files[curfile].metadata
			message.Body = body
			curfile++
			return message, err
		}, nil
	}
	return nil, fmt.Errorf("invaild format %s", format)
}