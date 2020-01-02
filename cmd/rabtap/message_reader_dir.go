// read persisted metadata and messages from a directory
// Copyright (C) 2019 Jan Delgado

package main

import (
	"os"
	"regexp"
)

const metadataFilePattern = `^rabtap-[0-9]+.json$`

type DirReader func(string) ([]os.FileInfo, error)
type FileInfoPredicate func(fileinfo os.FileInfo) bool

type filenameWithMetadata struct {
	filename string
	metadata RabtapPersistentMessage
}

// newRabtapFileInfoPredicate returns a FileInfoPredicate that matches
// rabtap metadata files
func newRabtapFileInfoPredicate() FileInfoPredicate {
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

// findMetadataFilenames returns list of filenames looking like a rabtap
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
func readMetadataOfFiles(filenames []string) ([]filenameWithMetadata, error) {

	data := make([]filenameWithMetadata, len(filenames))
	for i, filename := range filenames {
		msg, err := readRabtapPersistentMessage(filename)
		if err != nil {
			return data, err
		}
		// we will load the body later when the message is published (from the
		// JSON or a separate message file). This approach read message bodies
		// twice, but this should not be a problem (for now...)
		msg.Body = []byte("")
		data[i] = filenameWithMetadata{filename: filename, metadata: msg}
	}

	return data, nil
}

// loadMetadataFromDir loads all metadata files from the given directory
// passing the given predicate
func loadMetadataFilesFromDir(dirname string, dirReader DirReader, pred FileInfoPredicate) ([]filenameWithMetadata, error) {
	filenames, err := findMetadataFilenames(dirname, dirReader, pred)
	if err != nil {
		return nil, err
	}
	return readMetadataOfFiles(filenames)
}
