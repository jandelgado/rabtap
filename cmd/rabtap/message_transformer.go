package main

// MessageTransformer transforms the given message
type MessageTransformer func(m RabtapPersistentMessage) (RabtapPersistentMessage, error)

// NewTransformingMessageSource returns a new message source that computes
// m = tn(...t1(f()), i.e. that applies the transformer to the message provided by f.
func NewTransformingMessageSource(f MessageSource, transformer ...MessageTransformer) MessageSource {
	return func() (RabtapPersistentMessage, error) {
		m, err := f()
		if err != nil {
			return RabtapPersistentMessage{}, err
		}
		for _, t := range transformer {
			m, err = t(m)
			if err != nil {
				return RabtapPersistentMessage{}, err
			}
		}
		return m, nil
	}
}
