// transform RabtapPersistentMessages
package main

// MessageTransformer transforms a RabtapPersistentMessage
type MessageTransformer func(RabtapPersistentMessage) (RabtapPersistentMessage, error)

// NewComposingMessageTransformer returns a new message transformer that computes
// m' = f(g(m))
func NewComposingMessageTransformer(f, g MessageTransformer) MessageTransformer {
	return func(m RabtapPersistentMessage) (RabtapPersistentMessage, error) {
		m, err := g(m)
		if err == nil {
			return f(m)
		}
		return RabtapPersistentMessage{}, err
	}
}
