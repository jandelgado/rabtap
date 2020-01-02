// Copyright (C) 2019 Jan Delgado

package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// a os.FileInfo implementation for tests
type fileInfoMock struct {
	name string
	mode os.FileMode
}

func (s fileInfoMock) Name() string {
	return s.name
}

func (s fileInfoMock) Size() int64 { return 0 }

func (s fileInfoMock) Mode() os.FileMode {
	return s.mode
}

func (s fileInfoMock) ModTime() time.Time {
	return time.Now()
}

func (s fileInfoMock) IsDir() bool {
	return false
}
func (s fileInfoMock) Sys() interface{} {
	return nil
}

func newFileInfoMock(name string, mode os.FileMode) fileInfoMock {
	return fileInfoMock{name, mode}
}

func mockDirFiles() []os.FileInfo {
	infos := []fileInfoMock{
		newFileInfoMock("somefile.txt", 0),
		newFileInfoMock("rabtap-1234.json", 0),
		newFileInfoMock("rabtap-9999.json", os.ModeDir),
		newFileInfoMock("rabtap-1235.json", 0),
		newFileInfoMock("xrabtap-1235.json", 0),
		newFileInfoMock("rabtap-1235.invalid", 0),
	}
	result := make([]os.FileInfo, len(infos))
	for i, v := range infos {
		result[i] = v
	}
	return result
}

func mockDirReader(dirname string) ([]os.FileInfo, error) {
	if dirname == "/rabtap" {
		return mockDirFiles(), nil
	}
	return nil, errors.New("invalid directory")
}

// writeTempFile creates a temporary file with data as it's content. The
// filename is returned.
func writeTempFile(t *testing.T, data string) string {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "")
	require.Nil(t, err)

	defer tmpFile.Close()

	_, err = tmpFile.Write([]byte(data))
	require.Nil(t, err)

	return tmpFile.Name()
}

func TestRabtapFileInfoPredicateFiltersExpectedFiles(t *testing.T) {
	p := newRabtapFileInfoPredicate()

	assert.True(t, p(newFileInfoMock("rabtap-1234.json", 0)))
	assert.True(t, p(newFileInfoMock("rabtap-1235.json", 0)))

	assert.False(t, p(newFileInfoMock("somefile.txt", 0)))
	assert.False(t, p(newFileInfoMock("rabtap-9999.jsonx", 0)))
	assert.False(t, p(newFileInfoMock("rabtap-9999.json", os.ModeDir)))
}

func TestFilterMetadataFilesAppliesFilterCorretly(t *testing.T) {
	pred := func(fi os.FileInfo) bool {
		return fi.Name() == "rabtap-1234.json"
	}
	files := filterMetadataFilenames(mockDirFiles(), pred)

	assert.Equal(t, 1, len(files))
	assert.Equal(t, "rabtap-1234.json", files[0])
}

func TestFindMetadataFilenamesFindsAllRabtapMetadataFiles(t *testing.T) {
	pred := func(fi os.FileInfo) bool {
		return fi.Name() == "rabtap-1234.json"
	}
	filenames, err := findMetadataFilenames("/rabtap", mockDirReader, pred)
	assert.Nil(t, err)
	assert.Equal(t, []string{"rabtap-1234.json"}, filenames)
}

func TestFindMetadataFilenamesReturnsErrorOnInvalidDirectory(t *testing.T) {
	pred := func(fi os.FileInfo) bool {
		return true
	}
	_, err := findMetadataFilenames("/invalid", mockDirReader, pred)
	assert.NotNil(t, err)
}

func TestReadMetadataFileReturnsErrorForNonExistingFile(t *testing.T) {
	_, err := readRabtapPersistenMessage("/this/file/should/not/exist")
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

	metadata, err := readRabtapPersistenMessage(filename)

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

	_, err := readMetadataOfFiles([]string{"/this/file/should/not/exist"})
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
	filename := writeTempFile(t, msg)
	defer os.Remove(filename)

	metadata, err := readMetadataOfFiles([]string{filename})

	assert.Nil(t, err)
	assert.Equal(t, 1, len(metadata))
	assert.Equal(t, filename, metadata[0].filename)
	assert.Equal(t, "amq.fanout", metadata[0].messageData.Exchange)
	assert.Equal(t, "key", metadata[0].messageData.RoutingKey)
	expectedTs, _ := time.Parse(time.RFC3339Nano, "2019-12-29T21:51:33.52288901+01:00")
	assert.Equal(t, expectedTs, metadata[0].messageData.XRabtapReceivedTimestamp)
	assert.Equal(t, []byte(""), metadata[0].messageData.Body)
	// etc
}

func TestReadMessageBodyFromFileReturnsFileAsIs(t *testing.T) {

	filename := writeTempFile(t, "Hello123")
	defer os.Remove(filename)

	contents, err := readMessageBodyFromFile(filename)

	assert.Nil(t, err)
	assert.Equal(t, []byte("Hello123"), contents)
}

func TestReadMessageBodyFromJSONReturnsBodyDecocded(t *testing.T) {
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

	contents, err := readMessageBodyFromJSON(filename)

	assert.Nil(t, err)
	assert.Equal(t, []byte("Hello"), contents)
}

func TestReadMessageBodyReadsMessageFromDatFileFirstIfPresent(t *testing.T) {
	metadata := `{ "Body": "" }`

	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	metadataFile := filepath.Join(dir, "rabtap.json")
	messageFile := filepath.Join(dir, "rabtap.dat")

	err = ioutil.WriteFile(metadataFile, []byte(metadata), 0666)
	require.Nil(t, err)
	err = ioutil.WriteFile(messageFile, []byte("Hello123"), 0666)
	require.Nil(t, err)

	body, err := readMessageBody(filepath.Join(dir, "rabtap"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("Hello123"), body)
}

func TestReadMessageBodyReadsMessageFromJSONIfNoDatFileIsPresent(t *testing.T) {
	msg := `{ "Body": "SGVsbG8=" } `

	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	messageFile := filepath.Join(dir, "rabtap.json")

	err = ioutil.WriteFile(messageFile, []byte(msg), 0666)
	require.Nil(t, err)

	body, err := readMessageBody(filepath.Join(dir, "rabtap"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("Hello"), body)
}

func TestReadMessageBodyReturnsErrorIfNoBodyIsPresent(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	_, err = readMessageBody(filepath.Join(dir, "rabtap"))
	assert.NotNil(t, err)
}
