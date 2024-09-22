package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessageTransformerProvidesAMessageSourceThatTransformsAMessageOnSuccess(t *testing.T) {

	// given
	provider := func() (RabtapPersistentMessage, error) {
		return RabtapPersistentMessage{MessageID: "123", UserID: "user1"}, nil
	}
	transformer1 := func(m RabtapPersistentMessage) (RabtapPersistentMessage, error) {
		newmsg := m
		newmsg.AppID = "appID"
		newmsg.UserID = ""
		return newmsg, nil
	}
	transformer2 := func(m RabtapPersistentMessage) (RabtapPersistentMessage, error) {
		newmsg := m
		newmsg.UserID = "transformedUser"
		return newmsg, nil
	}

	// when
	provider = NewTransformingMessageSource(provider, transformer1, transformer2)
	msg, err := provider()

	// then
	assert.NoError(t, err)
	assert.Equal(t, "transformedUser", msg.UserID)
	assert.Equal(t, "appID", msg.AppID)
	assert.Equal(t, "123", msg.MessageID)
}

func TestNewMessageTransformerProvidesAMessageSourceThatPropagtesErrors(t *testing.T) {

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
	provider = NewTransformingMessageSource(provider, transformer)
	_, err := provider()

	// then
	assert.Error(t, err, "some error")
}
