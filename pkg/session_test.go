// Copyright (C) 2019-2021 Jan Delgado

// +build integration

package rabtap

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
)

func TestSessionProvidesConnectionAndChannel(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := testcommon.NewTestLogger()
	sessions := redial(ctx, testcommon.IntegrationURIFromEnv().String(), &tls.Config{}, log, FailEarly)

	sessionFactory := <-sessions
	session := <-sessionFactory
	assert.NotNil(t, session.Connection)
	assert.NotNil(t, session.Channel)
	assert.Nil(t, session.Connection.Close())
}

func TestSessionShutsDownProperlyWhenCancelled(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	log := testcommon.NewTestLogger()
	sessions := redial(ctx, testcommon.IntegrationURIFromEnv().String(), &tls.Config{}, log, FailEarly)

	sessionFactory, more := <-sessions
	assert.True(t, more)
	time.Sleep(1 * time.Second)
	cancel()
	time.Sleep(1 * time.Second)

	<-sessionFactory
	_, more = <-sessions
	assert.False(t, more)
}

func TestSessionCanBeCancelledWhenSessionIsNotReadFromChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	log := testcommon.NewTestLogger()
	sessions := redial(ctx, testcommon.IntegrationURIFromEnv().String(), &tls.Config{}, log, FailEarly)

	sessionFactory, more := <-sessions
	assert.True(t, more)
	<-sessionFactory

	time.Sleep(1 * time.Second)
	cancel()
	time.Sleep(1 * time.Second)

	_, more = <-sessions
	assert.False(t, more)
}

func TestSessionFailsEarlyWhenNoConnectionIsPossible(t *testing.T) {

	ctx := context.Background()

	log := testcommon.NewTestLogger()
	sessions := redial(ctx, "amqp://localhost:1", &tls.Config{}, log, FailEarly)

	sessionFactory, more := <-sessions
	assert.True(t, more)

	_, more = <-sessionFactory
	assert.False(t, more)

	_, more = <-sessions
	assert.False(t, more)
}

func TestSessionCanBeCancelledDuringRetryDelay(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	log := testcommon.NewTestLogger()
	sessions := redial(ctx, "amqp://localhost:1", &tls.Config{}, log, !FailEarly)

	sessionFactory, more := <-sessions
	assert.True(t, more)

	time.Sleep(1 * time.Second)
	cancel()
	time.Sleep(1 * time.Second)

	_, more = <-sessionFactory
	assert.False(t, more)
}

func TestSessionNewChannelReturnsNewChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := testcommon.NewTestLogger()
	sessions := redial(ctx, testcommon.IntegrationURIFromEnv().String(), &tls.Config{}, log, FailEarly)

	sessionFactory := <-sessions
	session := <-sessionFactory

	assert.NotNil(t, session.Channel)
	chanOld := session.Channel
	session.NewChannel()
	assert.NotNil(t, session.Channel)
	assert.NotEqual(t, chanOld, session.Channel)
}
