package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessageTransformerProvidesAMessageProviderThatTransformsAMessageOnSuccess(t *testing.T) {

	// given
	provider := func() (RabtapPersistentMessage, error) {
		return RabtapPersistentMessage{UserID: "user1"}, nil
	}
	transformer := func(m RabtapPersistentMessage) (RabtapPersistentMessage, error) {
		newmsg := m
		newmsg.UserID = "transformedUser"
		return newmsg, nil
	}

	// when
	provider = NewTransformingMessageProvider(transformer, provider)
	msg, err := provider()

	// then
	assert.NoError(t, err)
	assert.Equal(t, "transformedUser", msg.UserID)
}

func TestNewMessageTransformerProvidesAMessageProviderThatPropagtesErrors(t *testing.T) {

	// given
	provider := func() (RabtapPersistentMessage, error) {
		return RabtapPersistentMessage{}, fmt.Errorf("some error")
	}
	transformer := func(m RabtapPersistentMessage) (RabtapPersistentMessage, error) {
		newmsg := m
		newmsg.UserID = "transformedUser"
		return newmsg, nil
	}

	// when
	provider = NewTransformingMessageProvider(transformer, provider)
	_, err := provider()

	// then
	assert.Error(t, err, "some error")
}
