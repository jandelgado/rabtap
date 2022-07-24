// read persisted messages
package main

// MessageReaderFunc provides messages that can be published.
// returns the message to be published, xor an error. When no more
// messages are available, io.EOF must be returned.
type MessageReaderFunc func() (RabtapPersistentMessage, error)
