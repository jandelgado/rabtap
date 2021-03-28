// Copyright (C) 2017 Jan Delgado

package rabtap

import (
	"context"
	"crypto/tls"
	"net/url"
	"testing"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
)

func TestDiscoveryUnknownExchange(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})

	_, err := DiscoverBindingsForExchange(context.TODO(), client, "/", "unknown")
	assert.NotNil(t, err)
}

func TestDiscoveryDirectExchange(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})

	result, err := DiscoverBindingsForExchange(context.TODO(), client, "/", "test-direct")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "direct-q1", result[0])
	assert.Equal(t, "direct-q2", result[1])
}

func TestDiscoveryTopicExchange(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})

	result, err := DiscoverBindingsForExchange(context.TODO(), client, "/", "test-topic")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "#", result[0])
}

func TestDiscoveryFanoutExchange(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})
	result, err := DiscoverBindingsForExchange(context.TODO(), client, "/", "test-fanout")

	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "", result[0])
}

func TestDiscoveryHeadersExchange(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})
	result, err := DiscoverBindingsForExchange(context.TODO(), client, "/", "test-headers")

	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "", result[0])
}
