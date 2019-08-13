// Copyright (C) 2019 Jan Delgado

// +build integration

package rabtap

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
)

func TestSessionProvidesConnectionAndChannel(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessions := redial(ctx, testcommon.IntegrationURIFromEnv(), &tls.Config{})

	sessionFactory := <-sessions
	session := <-sessionFactory
	assert.NotNil(t, session.Connection)
	assert.NotNil(t, session.Channel)
	assert.Nil(t, session.Connection.Close())
}

func TestSessionShutsDownProperlyWhenCancelled(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	sessions := redial(ctx, testcommon.IntegrationURIFromEnv(), &tls.Config{})

	sessionFactory, more := <-sessions
	assert.True(t, more)
	cancel()
	<-sessionFactory
	_, more = <-sessions
	assert.False(t, more)
}

func TestSessionFailsEarlyWhenNoConnectionIsPossible(t *testing.T) {

	ctx := context.Background()

	sessions := redial(ctx, "amqp://localhost:1", &tls.Config{})

	sessionFactory, more := <-sessions
	assert.True(t, more)

	_, more = <-sessionFactory
	assert.False(t, more)

	_, more = <-sessions
	assert.False(t, more)
}

func TestSessionNewChannelReturnsNewChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessions := redial(ctx, testcommon.IntegrationURIFromEnv(), &tls.Config{})

	sessionFactory := <-sessions
	session := <-sessionFactory

	assert.NotNil(t, session.Channel)
	chanOld := session.Channel
	session.NewChannel()
	assert.NotNil(t, session.Channel)
	assert.NotEqual(t, chanOld, session.Channel)
}
