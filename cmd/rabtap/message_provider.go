// provide RabtapPersistentMessages to be published
package main

// MessageProviderFunc provides messages that can be published.
// returns the message to be published, xor an error. When no more
// messages are available, io.EOF must be returned.
type MessageProviderFunc func() (RabtapPersistentMessage, error)
