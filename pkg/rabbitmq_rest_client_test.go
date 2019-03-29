// Copyright (C) 2017 Jan Delgado

package rabtap

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/jandelgado/rabtap/pkg/testcommon"
	"github.com/stretchr/testify/assert"
)

func TestGetAllResources(t *testing.T) {
	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	all, err := client.BrokerInfo()

	assert.Nil(t, err)
	assert.Equal(t, "3.6.9", all.Overview.ManagementVersion)
	assert.Equal(t, 1, len(all.Connections))
	assert.Equal(t, 12, len(all.Exchanges))
	assert.Equal(t, 8, len(all.Queues))
	assert.Equal(t, 16, len(all.Bindings))
	assert.Equal(t, 2, len(all.Consumers))
}

func TestGetAllResourcesOnInvalidHostReturnErr(t *testing.T) {
	client := NewRabbitHTTPClient("localhost:1", &tls.Config{})
	_, err := client.BrokerInfo()
	assert.NotNil(t, err)
}

// test invalid resource passed to getResource()
func TestGetResourceInvalidUriReturnsError(t *testing.T) {
	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	resCh := client.getResource(httpRequest{"invalid", reflect.TypeOf(RabbitOverview{})})

	select {
	case res := <-resCh:
		assert.NotNil(t, res.err)
	case <-time.After(time.Second * 1):
		assert.Fail(t, "result not received in expected time frame")
	}
}

// test non 200 status returned in getResource()
func TestGetResourceStatusNot200(t *testing.T) {

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "500 internal server error")
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := NewRabbitHTTPClient(ts.URL, &tls.Config{})
	resCh := client.getResource(httpRequest{ts.URL + "/overview", reflect.TypeOf(RabbitOverview{})})

	select {
	case res := <-resCh:
		assert.NotNil(t, res.err) // TODO check error
	case <-time.After(time.Second * 1):
		assert.Fail(t, "result not received in expected time frame")
	}
}

// // test non invalid json returned
func TestGetResourceInvalidJSON(t *testing.T) {

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "non json response")
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := NewRabbitHTTPClient(ts.URL, &tls.Config{})
	resCh := client.getResource(httpRequest{ts.URL + "/overview", reflect.TypeOf(RabbitOverview{})})

	select {
	case res := <-resCh:
		assert.NotNil(t, res.err) // TODO check error
	case <-time.After(time.Second * 1):
		assert.Fail(t, "result not received in expected time frame")
	}
}

// test of GET /api/exchanges endpoint
func TestRabbitClientGetExchanges(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	result, err := client.Exchanges()
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
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	result, err := client.Queues()
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

	result, err := client.Overview()
	assert.Nil(t, err)
	assert.Equal(t, "3.6.9", result.ManagementVersion)
}

// test of GET /api/bindings endpoint
func TestRabbitClientGetBindings(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	_, err := client.Bindings()
	assert.Nil(t, err)
	// TODO

}

// test of GET /api/consumers endpoint
func TestRabbitClientGetConsumers(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	consumer, err := client.Consumers()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(consumer))
	assert.Equal(t, "some_consumer", consumer[0].ConsumerTag)
	assert.Equal(t, "another_consumer w/ faulty channel", consumer[1].ConsumerTag)

}

// test of GET /api/consumers endpoint
func TestRabbitClientGetConnections(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	conn, err := client.Connections()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(conn))
	assert.Equal(t, "172.17.0.1:40874 -> 172.17.0.2:5672", conn[0].Name)
}

// test of GET /api/consumers endpoint workaround for empty channel_details
func TestRabbitClientGetConsumersChannelDetailsIsEmptyArray(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	consumer, err := client.Consumers()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(consumer))

	// the second channel_details were "[]" to test behaviour of RabbitMQ
	// api when [] is returned instead of a null object.
	assert.Equal(t, "another_consumer w/ faulty channel", consumer[1].ConsumerTag)
	assert.Equal(t, "", consumer[1].ChannelDetails.Name)
}

// test of DELETE /connections/conn to close a connection
func TestRabbitClientCloseExistingConnection(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	err := client.CloseConnection("172.17.0.1:40874 -> 172.17.0.2:5672", "reason")
	assert.Nil(t, err)
}

// test of DELETE /connections/conn to close a connection
func TestRabbitClientCloseNonExistingConnectionRaisesError(t *testing.T) {

	mock := testcommon.NewRabbitAPIMock(testcommon.MockModeStd)
	defer mock.Close()
	client := NewRabbitHTTPClient(mock.URL, &tls.Config{})

	err := client.CloseConnection("DOES NOT EXIST", "reason")
	assert.NotNil(t, err)
}

func TestFindExchangeByName(t *testing.T) {
	exchanges := []RabbitExchange{
		{Name: "exchange1", Vhost: "vhost"},
		{Name: "exchange2", Vhost: "vhost"},
	}
	assert.Equal(t, 1, FindExchangeByName(exchanges, "vhost", "exchange2"))
}

func TestFindExchangeByNameNotFound(t *testing.T) {
	exchanges := []RabbitExchange{
		{Name: "exchange1", Vhost: "vhost"},
	}
	assert.Equal(t, -1, FindExchangeByName(exchanges, "/", "not-available"))
}

func TestFindQueueByName(t *testing.T) {
	queues := []RabbitQueue{
		{Name: "q1", Vhost: "vhost"},
		{Name: "q2", Vhost: "vhost"},
	}
	assert.Equal(t, 1, FindQueueByName(queues, "vhost", "q2"))
}

func TestFindQueueByNameNotFound(t *testing.T) {
	queues := []RabbitQueue{
		{Name: "q1", Vhost: "vhost"},
		{Name: "q2", Vhost: "vhost"},
	}
	assert.Equal(t, -1, FindQueueByName(queues, "/", "not-available"))
}

func TestFindConnectionByName(t *testing.T) {
	conns := []RabbitConnection{
		{Name: "c1", Vhost: "vhost"},
		{Name: "c2", Vhost: "vhost"},
	}
	assert.Equal(t, 1, FindConnectionByName(conns, "vhost", "c2"))
}

func TestFindConnectionByNameNotFoundReturnsCorrectValue(t *testing.T) {
	assert.Equal(t, -1, FindConnectionByName([]RabbitConnection{}, "vhost", "c2"))
}

func TestFindConsumerByQueue(t *testing.T) {
	var con1, con2, con3 RabbitConsumer
	con1.Queue.Name = "q1"
	con1.Queue.Vhost = "vhost"
	con2.Queue.Name = "q2"
	con2.Queue.Vhost = "vhost"
	con3.Queue.Name = "q3"
	con3.Queue.Vhost = "vhost"
	cons := []RabbitConsumer{con1, con2, con3}
	assert.Equal(t, 1, FindConsumerByQueue(cons, "vhost", "q2"))
}

func TestFindConsumerByQueueNotFoundReturnsCorrectValue(t *testing.T) {
	assert.Equal(t, -1, FindConsumerByQueue([]RabbitConsumer{}, "vhost", "q1"))
}
