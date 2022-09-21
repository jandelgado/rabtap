// an implementation of MetadataService which uses in memory lookups
package rabtap

import "fmt"

type InMemoryMetadataService struct {
	brokerInfo           BrokerInfo
	vhostByName          map[string]*RabbitVhost
	exchangeByName       map[string]*RabbitExchange
	queueByName          map[string]*RabbitQueue
	channelByName        map[string]*RabbitChannel
	connectionByName     map[string]*RabbitConnection
	channelsByConnection map[string][]*RabbitChannel
	consumersByChannel   map[string][]*RabbitConsumer
	bindingsByExchange   map[string][]*RabbitBinding
}

func scopedName(vhost, name string) string {
	return fmt.Sprintf("%s-%s", vhost, name)
}

// entityByName returns a map[string]*T where all entities are referenced
// by their name, which is provided by nameFunc(T)
func entityByName[T any](entities []T, nameFunc func(T) string) map[string]*T {
	res := make(map[string]*T, len(entities))
	for _, ent := range entities {
		e := ent
		res[nameFunc(e)] = &e
	}
	return res
}

func findByName[T any](m map[string]*T, vhost, name string) *T {
	return m[scopedName(vhost, name)]
}

func channelsByConnection(channels []RabbitChannel) map[string][]*RabbitChannel {
	chanByConn := map[string][]*RabbitChannel{}
	for _, channel := range channels {
		connName := scopedName(channel.Vhost, channel.ConnectionDetails.Name)
		c := channel
		chans := chanByConn[connName]
		chanByConn[connName] = append(chans, &c)
	}
	return chanByConn
}

func consumersByChannel(consumers []RabbitConsumer) map[string][]*RabbitConsumer {
	consumerByChan := map[string][]*RabbitConsumer{}
	for _, consumer := range consumers {
		chanName := scopedName(consumer.Queue.Vhost, consumer.ChannelDetails.Name)
		c := consumer
		consumers := consumerByChan[chanName]
		consumerByChan[chanName] = append(consumers, &c)
	}
	return consumerByChan
}

func bindingsByExchange(bindings []RabbitBinding) map[string][]*RabbitBinding {
	bindingsByExchange := map[string][]*RabbitBinding{}
	for _, binding := range bindings {
		exchangeName := scopedName(binding.Vhost, binding.Source)
		b := binding
		bindings := bindingsByExchange[exchangeName]
		bindingsByExchange[exchangeName] = append(bindings, &b)
	}
	return bindingsByExchange
}

func NewInMemoryMetadataService(brokerInfo BrokerInfo) *InMemoryMetadataService {

	return &InMemoryMetadataService{
		brokerInfo:           brokerInfo,
		vhostByName:          entityByName(brokerInfo.Vhosts, func(v RabbitVhost) string { return scopedName(v.Name, "") }),
		exchangeByName:       entityByName(brokerInfo.Exchanges, func(e RabbitExchange) string { return scopedName(e.Vhost, e.Name) }),
		queueByName:          entityByName(brokerInfo.Queues, func(q RabbitQueue) string { return scopedName(q.Vhost, q.Name) }),
		channelByName:        entityByName(brokerInfo.Channels, func(c RabbitChannel) string { return scopedName(c.Vhost, c.Name) }),
		connectionByName:     entityByName(brokerInfo.Connections, func(c RabbitConnection) string { return scopedName(c.Vhost, c.Name) }),
		channelsByConnection: channelsByConnection(brokerInfo.Channels),
		consumersByChannel:   consumersByChannel(brokerInfo.Consumers),
		bindingsByExchange:   bindingsByExchange(brokerInfo.Bindings),
	}
}

func (s InMemoryMetadataService) Overview() RabbitOverview {
	return s.brokerInfo.Overview
}

func (s InMemoryMetadataService) Connections() []RabbitConnection {
	return s.brokerInfo.Connections
}

func (s InMemoryMetadataService) Exchanges() []RabbitExchange {
	return s.brokerInfo.Exchanges
}

func (s InMemoryMetadataService) Queues() []RabbitQueue {
	return s.brokerInfo.Queues
}

func (s InMemoryMetadataService) Consumers() []RabbitConsumer {
	return s.brokerInfo.Consumers
}

func (s InMemoryMetadataService) Bindings() []RabbitBinding {
	return s.brokerInfo.Bindings
}

func (s InMemoryMetadataService) Channels() []RabbitChannel {
	return s.brokerInfo.Channels
}

func (s InMemoryMetadataService) Vhosts() []RabbitVhost {
	return s.brokerInfo.Vhosts
}

func (s InMemoryMetadataService) FindQueueByName(vhost, name string) *RabbitQueue {
	return findByName(s.queueByName, vhost, name)
}

func (s InMemoryMetadataService) FindVhostByName(vhost string) *RabbitVhost {
	return findByName(s.vhostByName, vhost, "")
}

func (s InMemoryMetadataService) FindExchangeByName(vhost, name string) *RabbitExchange {
	return findByName(s.exchangeByName, vhost, name)
}

func (s InMemoryMetadataService) FindChannelByName(vhost, name string) *RabbitChannel {
	return findByName(s.channelByName, vhost, name)
}

func (s InMemoryMetadataService) FindConnectionByName(vhost, name string) *RabbitConnection {
	return findByName(s.connectionByName, vhost, name)
}

func (s InMemoryMetadataService) AllChannelsForConnection(vhost, name string) []*RabbitChannel {
	return s.channelsByConnection[scopedName(vhost, name)]
}

func (s InMemoryMetadataService) AllConsumersForChannel(vhost, name string) []*RabbitConsumer {
	return s.consumersByChannel[scopedName(vhost, name)]
}

func (s InMemoryMetadataService) AllBindingsForExchange(vhost, name string) []*RabbitBinding {
	return s.bindingsByExchange[scopedName(vhost, name)]
}

var _ MetadataService = (*InMemoryMetadataService)(nil)
