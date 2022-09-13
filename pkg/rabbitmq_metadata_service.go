// a service providing RabbitMQ metadata (queues, exchanges, connections...).
package rabtap

type MetadataService interface {
	Overview() RabbitOverview
	Connections() []RabbitConnection
	Exchanges() []RabbitExchange
	Queues() []RabbitQueue
	Consumers() []RabbitConsumer
	Bindings() []RabbitBinding
	Channels() []RabbitChannel
	Vhosts() []RabbitVhost

	FindQueueByName(vhost, name string) *RabbitQueue
	FindExchangeByName(vhost, name string) *RabbitExchange
	FindChannelByName(vhost, name string) *RabbitChannel
	FindConnectionByName(vhost, name string) *RabbitConnection
	FindVhostByName(vhost string) *RabbitVhost

	AllChannelsForConnection(vhost, name string) []*RabbitChannel
	AllConsumersForChannel(vhost, name string) []*RabbitConsumer
	AllBindingsForExchange(vhost, name string) []*RabbitBinding
}
