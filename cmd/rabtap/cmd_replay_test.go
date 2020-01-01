// Copyright (C) 2019 Jan Delgado

package main

import (
	"io/ioutil"
	"os"
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

func mockDirReader(_ string) ([]os.FileInfo, error) {
	return mockDirFiles(), nil
}

func writeTempFile(t *testing.T, data string) string {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "")
	require.Nil(t, err)

	defer tmpFile.Close()

	_, err = tmpFile.Write([]byte(data))
	require.Nil(t, err)

	return tmpFile.Name()
}

func TestFilterMetadataFilesReturnsOnlyRabtapMetadataFiles(t *testing.T) {
	files := filterMetadataFilenames(mockDirFiles())

	assert.Equal(t, 2, len(files))
	assert.Equal(t, "rabtap-1234.json", files[0])
	assert.Equal(t, "rabtap-1235.json", files[1])
}

func TestFindMetadataFilesReturnsOnlyRabtapMetadataFiles(t *testing.T) {
	files, err := findMetadataFilenames("testdir", mockDirReader)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(files))
	assert.Equal(t, "rabtap-1234.json", files[0])
	assert.Equal(t, "rabtap-1235.json", files[1])
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
  "Body": "Body"
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
