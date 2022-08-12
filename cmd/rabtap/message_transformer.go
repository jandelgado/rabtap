// compose message providers
package main

// MessageTransformer transforms the given message
type MessageTransformer func(m RabtapPersistentMessage) (RabtapPersistentMessage, error)

// NewTransformingMessageProvider returns a new message provider that computes
// m = t(f()), i.e. that applies the transformer to the message provided by f.
func NewTransformingMessageProvider(t MessageTransformer, f MessageProviderFunc) MessageProviderFunc {
	return func() (RabtapPersistentMessage, error) {
		m, err := f()
		if err == nil {
			return t(m)
		}
		return RabtapPersistentMessage{}, err
	}
}
