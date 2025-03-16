// Copyright (C) 2019 Jan Delgado

package main

import (
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// a os.DirEntry implementation for tests
type dirEntryMock struct {
	name string
	mode os.FileMode
}

func (s dirEntryMock) Name() string {
	return s.name
}

func (s dirEntryMock) IsDir() bool {
	return false
}

func (s dirEntryMock) Type() os.FileMode {
	return s.mode
}

func (s dirEntryMock) Info() (os.FileInfo, error) {
	panic("not supported")
}

func newDirEntryMock(name string, mode os.FileMode) dirEntryMock {
	return dirEntryMock{name, mode}
}

func mockDirFiles() []os.DirEntry {
	infos := []dirEntryMock{
		newDirEntryMock("somefile.txt", 0),
		newDirEntryMock("rabtap-1234.json", 0),
		newDirEntryMock("rabtap-9999.json", os.ModeDir),
		newDirEntryMock("rabtap-1235.json", 0),
		newDirEntryMock("xrabtap-1235.json", 0),
		newDirEntryMock("rabtap-1235.invalid", 0),
	}
	result := make([]os.DirEntry, len(infos))
	for i, v := range infos {
		result[i] = v
	}
	return result
}

func mockDirReader(dirname string) ([]os.DirEntry, error) {
	if dirname == "/rabtap" {
		return mockDirFiles(), nil
	}
	return nil, errors.New("invalid directory")
}

// writeTempFile creates a temporary file with data as it's content. The
// filename is returned.
func writeTempFile(t *testing.T, data string) string {
	tmpFile, err := os.CreateTemp(os.TempDir(), "")
	require.Nil(t, err)

	defer tmpFile.Close()

	_, err = tmpFile.Write([]byte(data))
	require.Nil(t, err)

	return tmpFile.Name()
}

func TestFilenameWithoutExtensionReturnsExpectedResult(t *testing.T) {
	assert.Equal(t, "/some/file", filenameWithoutExtension("/some/file.ext"))
	assert.Equal(t, "/some/file", filenameWithoutExtension("/some/file"))
}

func TestRabtapFileInfoPredicateFiltersExpectedFiles(t *testing.T) {
	p := NewRabtapFileInfoPredicate()

	assert.True(t, p(newDirEntryMock("rabtap-1234.json", 0)))
	assert.True(t, p(newDirEntryMock("rabtap-1235.json", 0)))

	assert.False(t, p(newDirEntryMock("somefile.txt", 0)))
	assert.False(t, p(newDirEntryMock("rabtap-9999.jsonx", 0)))
	assert.False(t, p(newDirEntryMock("rabtap-9999.json", os.ModeDir)))
}

func TestFilterMetadataFilesAppliesFilterCorretly(t *testing.T) {
	pred := func(e os.DirEntry) bool {
		return e.Name() == "rabtap-1234.json"
	}
	files := filterMetadataFilenames(mockDirFiles(), pred)

	assert.Equal(t, 1, len(files))
	assert.Equal(t, "rabtap-1234.json", files[0])
}

func TestFindMetadataFilenamesFindsAllRabtapMetadataFiles(t *testing.T) {
	pred := func(e os.DirEntry) bool {
		return e.Name() == "rabtap-1234.json"
	}
	filenames, err := findMetadataFilenames("/rabtap", mockDirReader, pred)
	assert.Nil(t, err)
	assert.Equal(t, []string{"rabtap-1234.json"}, filenames)
}

func TestFindMetadataFilenamesReturnsErrorOnInvalidDirectory(t *testing.T) {
	pred := func(e os.DirEntry) bool {
		return true
	}
	_, err := findMetadataFilenames("/invalid", mockDirReader, pred)
	assert.NotNil(t, err)
}

func TestReadMetadataFileReturnsErrorForNonExistingFile(t *testing.T) {
	_, err := readRabtapPersistentMessage("/this/file/should/not/exist")
	assert.NotNil(t, err)
}

func TestLoadMetadataFilesFromDirReturnsExpectedMetadata(t *testing.T) {
	pred := func(e os.DirEntry) bool {
		return e.Name() == "rabtap.json"
	}
	msg := `{ "Exchange": "exchange", "Body": "" }`

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	metadataFile := filepath.Join(dir, "rabtap.json")
	messageFile := filepath.Join(dir, "rabtap.dat")

	err = os.WriteFile(metadataFile, []byte(msg), 0o666)
	require.Nil(t, err)
	err = os.WriteFile(messageFile, []byte("Hello123"), 0o666)
	require.Nil(t, err)

	metadata, err := LoadMetadataFilesFromDir(dir, os.ReadDir, pred)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(metadata))
	assert.Equal(t, path.Join(dir, "rabtap.json"), metadata[0].filename)
	assert.Equal(t, "exchange", metadata[0].metadata.Exchange)
}

func TestLoadMetadataFilesFromDirFailsOnInvalidDir(t *testing.T) {
	pred := func(e os.DirEntry) bool {
		return true
	}
	dirReader := func(string) ([]os.DirEntry, error) {
		return nil, errors.New("invalid dir")
	}
	_, err := LoadMetadataFilesFromDir("unused", dirReader, pred)
	assert.NotNil(t, err)
}

func TestReadPersistentRabtapMessageReturnsCorrectObject(t *testing.T) {
	msg := `{
  "ContentType": "",
  "ContentEncoding": "",
  "DeliveryMode": 0,
  "Priority": 0,
  "CorrelationID": "",
  "ReplyTo": "",
  "Expiration": "",
  "MessageID": "",
  "Timestamp": "0001-01-01T00:00:00Z",
  "Type": "",
  "UserID": "",
  "AppID": "",
  "DeliveryTag": 1,
  "Redelivered": false,
  "Exchange": "amq.fanout",
  "RoutingKey": "key",
  "XRabtapReceivedTimestamp": "2019-12-29T21:51:33.52288901+01:00",
  "Body": "SGVsbG8="
}
`
	filename := writeTempFile(t, msg)
	defer os.Remove(filename)

	metadata, err := readRabtapPersistentMessage(filename)

	assert.Nil(t, err)
	assert.Equal(t, "amq.fanout", metadata.Exchange)
	assert.Equal(t, "key", metadata.RoutingKey)
	expectedTs, _ := time.Parse(time.RFC3339Nano, "2019-12-29T21:51:33.52288901+01:00")
	assert.Equal(t, expectedTs, metadata.XRabtapReceivedTimestamp)
	// base64dec("SGVsbG=") == "Hello"
	assert.Equal(t, []byte("Hello"), metadata.Body)
	// etc
}

func TestReadMetadataOfFilesFailsWithErrorIfAnyFileCouldNotBeRead(t *testing.T) {
	_, err := readMetadataOfFiles("/base", []string{"/this/file/should/not/exist"})
	assert.NotNil(t, err)
}

func TestReadMetadataOfFilesReturnsExpectedMetadata(t *testing.T) {
	msg := `{
  "ContentType": "",
  "ContentEncoding": "",
  "DeliveryMode": 0,
  "Priority": 0,
  "CorrelationID": "",
  "ReplyTo": "",
  "Expiration": "",
  "MessageID": "",
  "Timestamp": "0001-01-01T00:00:00Z",
  "Type": "",
  "UserID": "",
  "AppID": "",
  "DeliveryTag": 1,
  "Redelivered": false,
  "Exchange": "amq.fanout",
  "RoutingKey": "key",
  "XRabtapReceivedTimestamp": "2019-12-29T21:51:33.52288901+01:00",
  "Body": "SGVsbG8="
}
`
	dir, filename := path.Split(writeTempFile(t, msg))
	defer os.Remove(filename)

	data, err := readMetadataOfFiles(dir, []string{filename})

	assert.Nil(t, err)
	assert.Equal(t, 1, len(data))
	assert.Equal(t, path.Join(dir, filename), data[0].filename)
	assert.Equal(t, "amq.fanout", data[0].metadata.Exchange)
	assert.Equal(t, "key", data[0].metadata.RoutingKey)
	expectedTs, _ := time.Parse(time.RFC3339Nano, "2019-12-29T21:51:33.52288901+01:00")
	assert.Equal(t, expectedTs, data[0].metadata.XRabtapReceivedTimestamp)
	assert.Equal(t, []byte(""), data[0].metadata.Body)
	// etc
}

func TestCreateMessageFromDirReaderFuncReturnsErrorForUnknownFormat(t *testing.T) {
	_, err := NewReadFilesFromDirMessageSource("invalid", []FilenameWithMetadata{})
	assert.NotNil(t, err)
}

func TestCreateMessageFromDirReaderFuncReturnsCorrectReaderForJSONFormat(t *testing.T) {
	// TODO complete test

	formats := []string{"json", "json-nopp"}
	for _, format := range formats {
		reader, err := NewReadFilesFromDirMessageSource(format, []FilenameWithMetadata{})
		assert.Nil(t, err)
		assert.NotNil(t, reader)

		// TODO test the reader func with actual data
		_, err = reader()
		assert.Equal(t, err, io.EOF)
	}
}

func TestCreateMessageFromDirReaderFuncReturnsCorrectReaderForRawFormat(t *testing.T) {
	// TODO complete test

	reader, err := NewReadFilesFromDirMessageSource("raw", []FilenameWithMetadata{})
	assert.Nil(t, err)
	assert.NotNil(t, reader)

	// TODO test the reader func with actual data
	_, err = reader()
	assert.Equal(t, err, io.EOF)
}