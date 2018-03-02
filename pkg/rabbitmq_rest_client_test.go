// Copyright (C) 2017 Jan Delgado

package rabtap

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
)

// test invalid resource passed to getResource()
func TestGetResourceInvalidUri(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	err := client.getResource("xyz://abc", &RabbitOverview{})
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

	f := NewRabbitHTTPClient(ts.URL, &tls.Config{})
	err := f.getResource(ts.URL, &RabbitOverview{})
	assert.NotNil(t, err)
}

// test non invalid json returned
func TestGetResourceInvalidJSON(t *testing.T) {

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "non json response")
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	f := NewRabbitHTTPClient(ts.URL, &tls.Config{})
	err := f.getResource(ts.URL, &RabbitOverview{})
	assert.NotNil(t, err)
}

// test of GET /api/exchanges endpoint
func TestRabbitClientGetExchanges(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	result, err := client.GetExchanges()
	assert.Nil(t, err)
	assert.Equal(t, 12, len(result))
	assert.Equal(t, "", (result)[0].Name)
	assert.Equal(t, "/", (result)[0].Vhost)
	assert.Equal(t, "direct", (result)[0].Type)
	assert.Equal(t, "amq.direct", (result)[1].Name)
	assert.Equal(t, "/", (result)[1].Vhost)
	assert.Equal(t, "direct", (result)[1].Type)
	assert.Equal(t, "amq.fanout", (result)[2].Name)
	assert.Equal(t, "/", (result)[2].Vhost)
	assert.Equal(t, "fanout", (result)[2].Type)
	// etc ...
}

// test of GET /api/queues endpoint
func TestRabbitClientGetQueues(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	result, err := client.GetQueues()
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
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	result, err := client.GetOverview()
	assert.Nil(t, err)
	assert.Equal(t, "3.6.9", result.ManagementVersion)

}

// test of GET /api/bindings endpoint
func TestRabbitClientGetBindings(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	_, err := client.GetBindings()
	assert.Nil(t, err)
	// TODO

}

// test of GET /api/consumers endpoint
func TestRabbitClientGetConsumers(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	consumer, err := client.GetConsumers()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(consumer))
	assert.Equal(t, "some_consumer", consumer[0].ConsumerTag)
	assert.Equal(t, "another_consumer w/ faulty channel", consumer[1].ConsumerTag)

}

// test of GET /api/consumers endpoint workaround for empty channel_details
func TestRabbitClientGetConsumersChannelDetailsIsEmptyArray(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	consumer, err := client.GetConsumers()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(consumer))

	// the second channel_details were "[]" to test behaviour of RabbitMQ
	// api when [] is returned instead of a null object.
	assert.Equal(t, "another_consumer w/ faulty channel", consumer[1].ConsumerTag)
	assert.Equal(t, "", consumer[1].ChannelDetails.Name)
}
