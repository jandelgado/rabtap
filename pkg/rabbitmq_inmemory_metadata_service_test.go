package rabtap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryMetadataServiceWorksWithEmptyBrokerInfo(t *testing.T) {
	brokerInfo := BrokerInfo{
		Overview:    RabbitOverview{},
		Connections: []RabbitConnection{},
		Exchanges:   []RabbitExchange{},
		Queues:      []RabbitQueue{},
		Consumers:   []RabbitConsumer{},
		Bindings:    []RabbitBinding{},
		Channels:    []RabbitChannel{},
		Vhosts:      []RabbitVhost{},
	}
	metadataService := NewInMemoryMetadataService(brokerInfo)

	assert.Equal(t, RabbitOverview{}, metadataService.Overview())
	assert.Equal(t, []RabbitConnection{}, metadataService.Connections())
	assert.Equal(t, []RabbitExchange{}, metadataService.Exchanges())
	assert.Equal(t, []RabbitQueue{}, metadataService.Queues())
	assert.Equal(t, []RabbitConsumer{}, metadataService.Consumers())
	assert.Equal(t, []RabbitBinding{}, metadataService.Bindings())
	assert.Equal(t, []RabbitChannel{}, metadataService.Channels())
	assert.Equal(t, []RabbitVhost{}, metadataService.Vhosts())

	assert.Equal(t, []*RabbitChannel(nil), metadataService.AllChannelsForConnection("vhost", "conn"))
	assert.Equal(t, []*RabbitConsumer(nil), metadataService.AllConsumersForChannel("vhost", "cons"))
	assert.Equal(t, []*RabbitBinding(nil), metadataService.AllBindingsForExchange("vhost", "exch"))

	assert.Nil(t, metadataService.FindQueueByName("vhost", "chan"))
	assert.Nil(t, metadataService.FindVhostByName("vhost"))
	assert.Nil(t, metadataService.FindExchangeByName("vhost", "exchange"))
	assert.Nil(t, metadataService.FindChannelByName("vhost", "chan"))
	assert.Nil(t, metadataService.FindConnectionByName("vhost", "conn"))

}
