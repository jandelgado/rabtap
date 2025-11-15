// Copyright (C) 2019-2021 Jan Delgado
//go:build integration

package rabtap

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jandelgado/rabtap/pkg/testcommon"
)

func TestSessionProvidesConnectionAndChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.DiscardHandler)
	sessions := redial(ctx, testcommon.IntegrationURIFromEnv(), &tls.Config{}, logger, FailEarly)

	sessionFactory := <-sessions
	session := <-sessionFactory
	assert.NotNil(t, session.Connection)
	assert.NotNil(t, session.Channel)
	assert.Nil(t, session.Connection.Close())
}

func TestSessionShutsDownProperlyWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	logger := slog.New(slog.DiscardHandler)
	sessions := redial(ctx, testcommon.IntegrationURIFromEnv(), &tls.Config{}, logger, FailEarly)

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

	logger := slog.New(slog.DiscardHandler)
	sessions := redial(ctx, testcommon.IntegrationURIFromEnv(), &tls.Config{}, logger, FailEarly)

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
	u, _ := url.Parse("amqp://localhost:1")
	logger := slog.New(slog.DiscardHandler)
	sessions := redial(ctx, u, &tls.Config{}, logger, FailEarly)

	sessionFactory, more := <-sessions
	assert.True(t, more)

	_, more = <-sessionFactory
	assert.False(t, more)

	_, more = <-sessions
	assert.False(t, more)
}

func TestSessionCanBeCancelledDuringRetryDelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	u, _ := url.Parse("amqp://localhost:1")
	logger := slog.New(slog.DiscardHandler)
	sessions := redial(ctx, u, &tls.Config{}, logger, !FailEarly)

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

	logger := slog.New(slog.DiscardHandler)
	sessions := redial(ctx, testcommon.IntegrationURIFromEnv(), &tls.Config{}, logger, FailEarly)

	sessionFactory := <-sessions
	session := <-sessionFactory

	assert.NotNil(t, session.Channel)
	chanOld := session.Channel
	session.NewChannel()
	assert.NotNil(t, session.Channel)
	assert.NotEqual(t, chanOld, session.Channel)
}
