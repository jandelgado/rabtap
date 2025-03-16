// Copyright (C) 2017 Jan Delgado

package rabtap

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jandelgado/rabtap/pkg/testcommon"
)

func TestHTTPTimeoutIsDefaultIfNotSetOrInvalid(t *testing.T) {
	t.Setenv("RABTAP_HTTP_TIMEOUT", "")
	assert.Equal(t, HTTP_DEFAULT_TIMEOUT, httpTimeout())

	t.Setenv("RABTAP_HTTP_TIMEOUT", "invalid")
	assert.Equal(t, HTTP_DEFAULT_TIMEOUT, httpTimeout())
}

func TestHTTPTimeoutCanBeConfigured(t *testing.T) {
	t.Setenv("RABTAP_HTTP_TIMEOUT", "3m")
	assert.Equal(t, time.Duration(3*time.Minute), httpTimeout())
}

func TestGetAllResources(t *testing.T) {
	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})

	all, err := client.BrokerInfo(context.TODO())

	assert.Nil(t, err)
	assert.Equal(t, "3.6.9", all.Overview.ManagementVersion)
	assert.Equal(t, 1, len(all.Connections))
	assert.Equal(t, 12, len(all.Exchanges))
	assert.Equal(t, 8, len(all.Queues))
	assert.Equal(t, 17, len(all.Bindings))
	assert.Equal(t, 2, len(all.Consumers))
}

func TestGetAllResourcesOnInvalidHostReturnErr(t *testing.T) {
	url, _ := url.Parse("localhost:1")
	client := NewRabbitHTTPClient(url, &tls.Config{})
	_, err := client.BrokerInfo(context.TODO())
	assert.NotNil(t, err)
}

// test invalid resource passed to getResource()
func TestGetResourceInvalidUriReturnsError(t *testing.T) {
	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})
	_, err := client.getResource(context.TODO(), httpRequest{"invalid", reflect.TypeOf(RabbitOverview{})})
	assert.NotNil(t, err)
}

// test non 200 status returned in getResource()
func TestGetResourceStatusNot200(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "500 internal server error")
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	url, _ := url.Parse(ts.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})
	_, err := client.getResource(context.TODO(), httpRequest{"overview", reflect.TypeOf(RabbitOverview{})})
	assert.NotNil(t, err) // TODO check error
}

// // test non invalid json returned
func TestGetResourceInvalidJSON(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "non json response")
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	url, _ := url.Parse(ts.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})
	_, err := client.getResource(context.TODO(), httpRequest{"overview", reflect.TypeOf(RabbitOverview{})})
	assert.NotNil(t, err) // TODO check error
}

// test of GET /api/exchanges endpoint
func TestRabbitClientGetExchanges(t *testing.T) {
	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})

	result, err := client.Exchanges(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 12, len(result))
	assert.Equal(t, "", (result)[0].Name)
	assert.Equal(t, "/", (result)[0].Vhost)
	assert.Equal(t, "direct", (result)[0].Type)
	assert.Equal(t, "amq.direct", (result)[1].Name)
	// etc ...
}

// test of GET /api/queues endpoint
func TestRabbitClientGetQueues(t *testing.T) {
	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})

	result, err := client.Queues(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 8, len(result))
	assert.Equal(t, "/", (result)[0].Vhost)
	assert.Equal(t, "direct-q1", (result)[0].Name)
	// etc ...
}

// test of GET /api/overview endpoint
func TestRabbitClientGetOverview(t *testing.T) {
	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})

	result, err := client.Overview(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, "3.6.9", result.ManagementVersion)
}

// test of GET /api/bindings endpoint
func TestRabbitClientGetBindings(t *testing.T) {
	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})

	_, err := client.Bindings(context.TODO())
	assert.Nil(t, err)
	// TODO
}

// test of GET /api/consumers endpoint
func TestRabbitClientGetConsumers(t *testing.T) {
	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})

	consumer, err := client.Consumers(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 2, len(consumer))
	assert.Equal(t, "some_consumer", consumer[0].ConsumerTag)
	assert.Equal(t, "another_consumer w/ faulty channel", consumer[1].ConsumerTag)
}

// test of GET /api/consumers endpoint
func TestRabbitClientGetConnections(t *testing.T) {
	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})

	conn, err := client.Connections(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(conn))
	assert.Equal(t, "172.17.0.1:40874 -> 172.17.0.2:5672", conn[0].Name)
}

func TestRabbitClientDeserializePeerPortInConsumerToInt(t *testing.T) {
	msg := `
[
  {
    "arguments": {},
    "ack_required": true,
    "active": true,
    "activity_status": "up",
    "channel_details": {
      "connection_name": "XXX",
      "name": "XXX (1)",
      "node": "XXX",
      "number": 1,
      "peer_host": "undefined",
      "peer_port": 1234,
      "user": "none"
    },
    "consumer_tag": "amq.ctag-InRAvLn4GW3j2mRwPmWJxA",
    "exclusive": false,
    "prefetch_count": 20,
    "queue": {
      "name": "logstream",
      "vhost": "/"
    }
  }
]
`
	var consumer []RabbitConsumer
	err := json.Unmarshal([]byte(msg), &consumer)
	assert.NoError(t, err)
	assert.Equal(t, OptInt(1234), consumer[0].ChannelDetails.PeerPort)
}

func TestRabbitClientDeserializePeerPortInConsumerAsStringWithoutError(t *testing.T) {
	// RabbitMQ sometimes returns "undefined" for the peer_port attribute,
	// but we expect an integer.
	msg := `
[
  {
    "arguments": {},
    "ack_required": true,
    "active": true,
    "activity_status": "up",
    "channel_details": {
      "connection_name": "XXX",
      "name": "XXX (1)",
      "node": "XXX",
      "number": 1,
      "peer_host": "undefined",
      "peer_port": "undefined",
      "user": "none"
    },
    "consumer_tag": "amq.ctag-InRAvLn4GW3j2mRwPmWJxA",
    "exclusive": false,
    "prefetch_count": 20,
    "queue": {
      "name": "logstream",
      "vhost": "/"
    }
  }
]
`
	var consumer []RabbitConsumer
	err := json.Unmarshal([]byte(msg), &consumer)
	assert.NoError(t, err)
	assert.Equal(t, OptInt(0), consumer[0].ChannelDetails.PeerPort)
}

// we use a custom unmarshaler as a WORKAROUND for RabbitMQ API
// returning "[]" instead of null.  To make sure deserialization does not
// break, we catch this case, and return an empty ChannelDetails struct.
// see e.g. https://github.com/rabbitmq/rabbitmq-management/issues/424
func TestChannelDetailsIsDetectedAsNull(t *testing.T) {
	msg := `
[
  {
    "arguments": {},
    "ack_required": true,
    "active": true,
    "activity_status": "up",
    "channel_details": null,
    "consumer_tag": "amq.ctag-InRAvLn4GW3j2mRwPmWJxA",
    "exclusive": false,
    "prefetch_count": 20,
    "queue": {
      "name": "logstream",
      "vhost": "/"
    }
  }
]
`
	var consumer []RabbitConsumer
	err := json.Unmarshal([]byte(msg), &consumer)
	assert.NoError(t, err)
	assert.Equal(t, ChannelDetails{}, consumer[0].ChannelDetails)
}

// we use a custom unmarshaler as a WORKAROUND for RabbitMQ API
// returning "[]" instead of null.  To make sure deserialization does not
// break, we catch this case, and return an empty ChannelDetails struct.
// see e.g. https://github.com/rabbitmq/rabbitmq-management/issues/424
func TestChannelDetailsIsDetectedAsEmptyArray(t *testing.T) {
	msg := `
[
  {
    "arguments": {},
    "ack_required": true,
    "active": true,
    "activity_status": "up",
    "channel_details": [],
    "consumer_tag": "amq.ctag-InRAvLn4GW3j2mRwPmWJxA",
    "exclusive": false,
    "prefetch_count": 20,
    "queue": {
      "name": "logstream",
      "vhost": "/"
    }
  }
]
`
	var consumer []RabbitConsumer
	err := json.Unmarshal([]byte(msg), &consumer)
	assert.NoError(t, err)
	assert.Equal(t, ChannelDetails{}, consumer[0].ChannelDetails)
}

// test of DELETE /connections/conn to close a connection
func TestRabbitClientCloseExistingConnection(t *testing.T) {
	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})

	err := client.CloseConnection(context.TODO(),
		"172.17.0.1:40874 -> 172.17.0.2:5672", "reason")
	assert.Nil(t, err)
}

// test of DELETE /connections/conn to close a connection
func TestRabbitClientCloseNonExistingConnectionRaisesError(t *testing.T) {
	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	url, _ := url.Parse(mock.URL)
	client := NewRabbitHTTPClient(url, &tls.Config{})

	err := client.CloseConnection(context.TODO(), "DOES NOT EXIST", "reason")
	assert.NotNil(t, err)
}