// rabbitmq http api client
// Copyright (C) 2017-2022 Jan Delgado

package rabtap

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"

	"golang.org/x/net/context/ctxhttp"

	"golang.org/x/sync/errgroup"
)

const HTTP_DEFAULT_TIMEOUT = time.Duration(time.Second * 10)

// RabbitHTTPClient is a minimal client to the rabbitmq management REST api.
// It implements only functions needed by this tool (i.e. GET on some of the
// resources).  The messages structs were generated using json-to-go (
// https://mholt.github.io/json-to-go/).
type RabbitHTTPClient struct {
	url    *url.URL // base URL
	client *http.Client
}

// httpTimeout returns the HTTP timeout value to use. It's either the default
// HTTP_DEFAULT_TIMEOUT or specified by the environment variable RABTAP_HTTP_TIMEOUT
// as a time.Duration.
func httpTimeout() time.Duration {
	timeoutStr := os.Getenv("RABTAP_HTTP_TIMEOUT")
	if timeoutStr == "" {
		return HTTP_DEFAULT_TIMEOUT
	}
	if timeout, err := time.ParseDuration(timeoutStr); err != nil {
		return HTTP_DEFAULT_TIMEOUT
	} else {
		return timeout
	}
}

// NewRabbitHTTPClient returns a new instance of an RabbitHTTPClient. url
// is the base API URL of the REST server.
func NewRabbitHTTPClient(url *url.URL, tlsConfig *tls.Config) *RabbitHTTPClient {
	tr := &http.Transport{
		TLSClientConfig:    tlsConfig,
		DisableCompression: false,
		Dial:               Dialer,
	}
	client := &http.Client{Transport: tr, Timeout: httpTimeout()}
	return &RabbitHTTPClient{url, client}
}

type httpRequest struct {
	path string       // relative path
	t    reflect.Type // type of expected result
}

// getResource gets resource constructed from s.url and equest.url and
// deserialized the resource into an request.t type, which is returned.
func (s *RabbitHTTPClient) getResource(ctx context.Context, request httpRequest) (interface{}, error) {
	r := reflect.New(request.t).Interface()
	url := s.url.String() + "/" + request.path
	resp, err := ctxhttp.Get(ctx, s.client, url)
	if err != nil {
		return r, err
	}

	if resp.StatusCode != 200 {
		return r, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(r)
	return r, err
}

// delResource make DELETE request to given relative path
func (s *RabbitHTTPClient) delResource(ctx context.Context, path string) error {
	url := s.url.String() + "/" + path
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := ctxhttp.Do(ctx, s.client, req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return errors.New(resp.Status)
	}
	defer resp.Body.Close()
	return nil
}

// BrokerInfo represents the state of various RabbitMQ ressources as
// returned by the RabbitMQ REST API
type BrokerInfo struct {
	Overview    RabbitOverview
	Connections []RabbitConnection
	Exchanges   []RabbitExchange
	Queues      []RabbitQueue
	Consumers   []RabbitConsumer
	Bindings    []RabbitBinding
	Channels    []RabbitChannel
	Vhosts      []RabbitVhost
}

// Overview returns the /overview resource of the RabbitMQ REST API
func (s *RabbitHTTPClient) Overview(ctx context.Context) (RabbitOverview, error) {
	res, err := s.getResource(ctx, httpRequest{"overview", reflect.TypeOf(RabbitOverview{})})
	return *res.(*RabbitOverview), err
}

// Connections returns the /connections resource of the RabbitMQ REST API
func (s *RabbitHTTPClient) Connections(ctx context.Context) ([]RabbitConnection, error) {
	res, err := s.getResource(ctx, httpRequest{"connections", reflect.TypeOf([]RabbitConnection{})})
	return *res.(*[]RabbitConnection), err
}

// Channels returns the /channels resource of the RabbitMQ REST API
func (s *RabbitHTTPClient) Channels(ctx context.Context) ([]RabbitChannel, error) {
	res, err := s.getResource(ctx, httpRequest{"channels", reflect.TypeOf([]RabbitChannel{})})
	return *res.(*[]RabbitChannel), err
}

// Exchanges returns the /exchanges resource of the RabbitMQ REST API
func (s *RabbitHTTPClient) Exchanges(ctx context.Context) ([]RabbitExchange, error) {
	res, err := s.getResource(ctx, httpRequest{"exchanges", reflect.TypeOf([]RabbitExchange{})})
	return *res.(*[]RabbitExchange), err
}

// Queues returns the /queues resource of the RabbitMQ REST API
func (s *RabbitHTTPClient) Queues(ctx context.Context) ([]RabbitQueue, error) {
	res, err := s.getResource(ctx, httpRequest{"queues", reflect.TypeOf([]RabbitQueue{})})
	return *res.(*[]RabbitQueue), err
}

// Consumers returns the /consumers resource of the RabbitMQ REST API
func (s *RabbitHTTPClient) Consumers(ctx context.Context) ([]RabbitConsumer, error) {
	res, err := s.getResource(ctx, httpRequest{"consumers", reflect.TypeOf([]RabbitConsumer{})})
	return *res.(*[]RabbitConsumer), err
}

// Bindings returns the /bindings resource of the RabbitMQ REST API
func (s *RabbitHTTPClient) Bindings(ctx context.Context) ([]RabbitBinding, error) {
	res, err := s.getResource(ctx, httpRequest{"bindings", reflect.TypeOf([]RabbitBinding{})})
	return *res.(*[]RabbitBinding), err
}

// Vhosts returns the /vhosts resource of the RabbitMQ REST API
func (s *RabbitHTTPClient) Vhosts(ctx context.Context) ([]RabbitVhost, error) {
	res, err := s.getResource(ctx, httpRequest{"vhosts", reflect.TypeOf([]RabbitVhost{})})
	return *res.(*[]RabbitVhost), err
}

// BrokerInfo gets all resources of the broker in parallel
func (s *RabbitHTTPClient) BrokerInfo(ctx context.Context) (BrokerInfo, error) {
	g, ctx := errgroup.WithContext(ctx)

	var r BrokerInfo
	g.Go(func() (err error) { r.Overview, err = s.Overview(ctx); return })
	g.Go(func() (err error) { r.Connections, err = s.Connections(ctx); return })
	g.Go(func() (err error) { r.Exchanges, err = s.Exchanges(ctx); return })
	g.Go(func() (err error) { r.Queues, err = s.Queues(ctx); return })
	g.Go(func() (err error) { r.Consumers, err = s.Consumers(ctx); return })
	g.Go(func() (err error) { r.Bindings, err = s.Bindings(ctx); return })
	g.Go(func() (err error) { r.Channels, err = s.Channels(ctx); return })
	g.Go(func() (err error) { r.Vhosts, err = s.Vhosts(ctx); return })
	return r, g.Wait()
}

// CloseConnection closes a connection by DELETING the associated resource
func (s *RabbitHTTPClient) CloseConnection(ctx context.Context, conn, reason string) error {
	return s.delResource(ctx, "connections/"+conn)
}

// UnmarshalJSON is a workaround to deserialize int attributes in the
// RabbitMQ API which are sometimes returned as strings, (i.e. the
// value "undefined").
func (d *OptInt) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		return nil
	}
	type Alias int
	aux := (*Alias)(d)
	return json.Unmarshal(data, aux)
}

// helper for UnmarshalJSON. Unfortunately we can not use generic here and define
// an "type Alias T" here (see://go101.org/generics/888-the-status-quo-of-go-custom-generics.html)
// So some boiler plate is left in the UnmarshalJSON functions.
func unmarshalEmptyArrayOrObject(data []byte, v any) error {
	if data[0] == '[' {
		// JSON array detected
		return nil
	}
	return json.Unmarshal(data, v)
}

// UnmarshalJSON is a custom unmarshaler as a WORKAROUND for RabbitMQ API
// returning "[]" instead of null.  To make sure deserialization does not
// break, we catch this case, and return an empty ChannelDetails struct.
// see e.g. https://github.com/rabbitmq/rabbitmq-management/issues/424
func (d *ChannelDetails) UnmarshalJSON(data []byte) error {
	// alias ChannelDetails to avoid recursion when calling Unmarshal
	type Alias ChannelDetails
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(d),
	}
	return unmarshalEmptyArrayOrObject(data, &aux)
}

// UnmarshalJSON is a custom unmarshaler as a WORKAROUND for RabbitMQ API
// returning "[]" instead of null.  To make sure deserialization does not
// break, we catch this case, and return an empty ChannelDetails struct.
// see e.g. https://github.com/rabbitmq/rabbitmq-management/issues/424
func (d *ConnectionDetails) UnmarshalJSON(data []byte) error {
	type Alias ConnectionDetails
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(d),
	}
	return unmarshalEmptyArrayOrObject(data, &aux)
}