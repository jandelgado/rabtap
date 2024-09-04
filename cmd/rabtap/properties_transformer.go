// Copyright (C) 2024 Jan Delgado

package main

// NewPropertiesTransformer creates a MessageTransformer that
// set/overrides message properties
func NewPropertiesTransformer(props PropertiesOverride) MessageTransformer {
	return func(m RabtapPersistentMessage) (RabtapPersistentMessage, error) {
		return *m.WithProperties(props), nil // TODO
	}
}
