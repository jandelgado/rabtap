// Copyright (C) 2019-2021 Jan Delgado

// +build integration

package rabtap

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"testing"
	"time"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// relative location of certificates used during the integration tests
const certDir = "../inttest/pki/certs"

func TestSessionProvidesConnectionAndChannel(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := testcommon.NewTestLogger()
	sessions := redial(ctx, testcommon.IntegrationURIFromEnv(), &tls.Config{}, log, FailEarly)

	sessionFactory := <-sessions
	session := <-sessionFactory
	assert.NotNil(t, session.Connection)
	assert.NotNil(t, session.Channel)
	assert.Nil(t, session.Connection.Close())
}

func TestSessionShutsDownProperlyWhenCancelled(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	log := testcommon.NewTestLogger()
	sessions := redial(ctx, testcommon.IntegrationURIFromEnv(), &tls.Config{}, log, FailEarly)

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
	sessions := redial(ctx, testcommon.IntegrationURIFromEnv(), &tls.Config{}, log, FailEarly)

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
	sessions := redial(ctx, testcommon.IntegrationURIFromEnv(), &tls.Config{}, log, FailEarly)

	sessionFactory := <-sessions
	session := <-sessionFactory

	assert.NotNil(t, session.Channel)
	chanOld := session.Channel
	session.NewChannel()
	assert.NotNil(t, session.Channel)
	assert.NotEqual(t, chanOld, session.Channel)
}

func TestSessionNewChannelReturnsNewChannelWithTLS(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := testcommon.NewTestLogger()
	url := "amqps://localhost:5671/"
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	// certificates created for by testdata/pki/mkcerts.sh
	// server certificates issued for localhost.
	// client certificate uses CN testuser.

	// load client certificate, key and ca certificate
	cert, err := tls.LoadX509KeyPair(certDir+"/testuser.crt", certDir+"/testuser.key")
	require.NoError(t, err)

	tlsConfig.Certificates = []tls.Certificate{cert}
	tlsConfig.BuildNameToCertificate()

	// Load CA cert
	caCert, err := ioutil.ReadFile(certDir + "/ca.crt")
	require.NoError(t, err)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool
	tlsConfig.BuildNameToCertificate()

	sessions := redial(ctx, url, tlsConfig, log, FailEarly)

	sessionFactory := <-sessions
	session, ok := <-sessionFactory
	require.True(t, ok)
	assert.NotNil(t, session.Channel)
	chanOld := session.Channel
	session.NewChannel()
	assert.NotNil(t, session.Channel)
	assert.NotEqual(t, chanOld, session.Channel)
}

func TestSessionNewChannelFailsWithCertificateWithUnknownUser(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := testcommon.NewTestLogger()
	url := "amqps://localhost:5671/"
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	// load client certificate, key and ca certificate
	cert, err := tls.LoadX509KeyPair(certDir+"/unknown.crt", certDir+"/unknown.key")
	require.NoError(t, err)

	tlsConfig.Certificates = []tls.Certificate{cert}
	tlsConfig.BuildNameToCertificate()

	// Load CA cert
	caCert, err := ioutil.ReadFile(certDir + "/ca.crt")
	require.NoError(t, err)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool
	tlsConfig.BuildNameToCertificate()

	sessions := redial(ctx, url, tlsConfig, log, FailEarly)

	sessionFactory := <-sessions
	_, ok := <-sessionFactory
	assert.False(t, ok)
}
