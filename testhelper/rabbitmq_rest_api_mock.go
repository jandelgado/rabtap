package testhelper

// partial mock of rabbitmq http rest api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

// NewRabbitAPIMock returns a mock server for the rabbitmq http managemet
// API. It is used by the integration test. Only a very limited subset
// of resources is support (GET exchanges, bindings, queues, overviews).
// Usage:
//   mockServer := NewRabbitAPIMock()
//   defer mockServer.Close()
//   client f := NewRabbitHTTPClient(mockServe.URL)
//
func NewRabbitAPIMock() *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// TODO add test for GET
		result := ""
		switch r.URL.RequestURI() {
		case "/exchanges":
			result = exchangeResult
		case "/bindings":
			result = bindingResult
		case "/queues":
			result = queueResult
		case "/overview":
			result = overviewResult
		case "/consumers":
			result = consumerResult
		default:
			w.WriteHeader(http.StatusNotFound)
		}
		fmt.Fprint(w, result)
	}
	return httptest.NewServer(http.HandlerFunc(handler))
	//	defer ts.Close()
}

var bindingResult = `
[
    {
        "source": "",
        "vhost": "/",
        "destination": "direct-q1",
        "destination_type": "queue",
        "routing_key": "direct-q1",
        "arguments": {

        },
        "properties_key": "direct-q1"
    },
    {
        "source": "",
        "vhost": "/",
        "destination": "direct-q2",
        "destination_type": "queue",
        "routing_key": "direct-q2",
        "arguments": {

        },
        "properties_key": "direct-q2"
    },
    {
        "source": "",
        "vhost": "/",
        "destination": "fanout-q1",
        "destination_type": "queue",
        "routing_key": "fanout-q1",
        "arguments": {

        },
        "properties_key": "fanout-q1"
    },
    {
        "source": "",
        "vhost": "/",
        "destination": "fanout-q2",
        "destination_type": "queue",
        "routing_key": "fanout-q2",
        "arguments": {

        },
        "properties_key": "fanout-q2"
    },
    {
        "source": "",
        "vhost": "/",
        "destination": "header-q1",
        "destination_type": "queue",
        "routing_key": "header-q1",
        "arguments": {

        },
        "properties_key": "header-q1"
    },
    {
        "source": "",
        "vhost": "/",
        "destination": "header-q2",
        "destination_type": "queue",
        "routing_key": "header-q2",
        "arguments": {

        },
        "properties_key": "header-q2"
    },
    {
        "source": "",
        "vhost": "/",
        "destination": "topic-q1",
        "destination_type": "queue",
        "routing_key": "topic-q1",
        "arguments": {

        },
        "properties_key": "topic-q1"
    },
    {
        "source": "",
        "vhost": "/",
        "destination": "topic-q2",
        "destination_type": "queue",
        "routing_key": "topic-q2",
        "arguments": {

        },
        "properties_key": "topic-q2"
    },
    {
        "source": "test-direct",
        "vhost": "/",
        "destination": "direct-q1",
        "destination_type": "queue",
        "routing_key": "direct-q1",
        "arguments": {

        },
        "properties_key": "direct-q1"
    },
    {
        "source": "test-direct",
        "vhost": "/",
        "destination": "direct-q2",
        "destination_type": "queue",
        "routing_key": "direct-q2",
        "arguments": {

        },
        "properties_key": "direct-q2"
    },
    {
        "source": "test-fanout",
        "vhost": "/",
        "destination": "fanout-q1",
        "destination_type": "queue",
        "routing_key": "",
        "arguments": {

        },
        "properties_key": "~"
    },
    {
        "source": "test-fanout",
        "vhost": "/",
        "destination": "fanout-q2",
        "destination_type": "queue",
        "routing_key": "",
        "arguments": {

        },
        "properties_key": "~"
    },
    {
        "source": "test-headers",
        "vhost": "/",
        "destination": "header-q1",
        "destination_type": "queue",
        "routing_key": "headers-q1",
        "arguments": {

        },
        "properties_key": "headers-q1"
    },
    {
        "source": "test-headers",
        "vhost": "/",
        "destination": "header-q2",
        "destination_type": "queue",
        "routing_key": "headers-q2",
        "arguments": {

        },
        "properties_key": "headers-q2"
    },
    {
        "source": "test-topic",
        "vhost": "/",
        "destination": "topic-q1",
        "destination_type": "queue",
        "routing_key": "topic-q1",
        "arguments": {

        },
        "properties_key": "topic-q1"
    },
    {
        "source": "test-topic",
        "vhost": "/",
        "destination": "topic-q2",
        "destination_type": "queue",
        "routing_key": "topic-q2",
        "arguments": {

        },
        "properties_key": "topic-q2"
    }
]`

var exchangeResult = `
[
    {
        "name": "",
        "vhost": "/",
        "type": "direct",
        "durable": true,
        "auto_delete": false,
        "internal": false,
        "arguments": {

        }
    },
    {
        "name": "amq.direct",
        "vhost": "/",
        "type": "direct",
        "durable": true,
        "auto_delete": false,
        "internal": false,
        "arguments": {

        }
    },
    {
        "name": "amq.fanout",
        "vhost": "/",
        "type": "fanout",
        "durable": true,
        "auto_delete": false,
        "internal": false,
        "arguments": {

        }
    },
    {
        "name": "amq.headers",
        "vhost": "/",
        "type": "headers",
        "durable": true,
        "auto_delete": false,
        "internal": false,
        "arguments": {

        }
    },
    {
        "name": "amq.match",
        "vhost": "/",
        "type": "headers",
        "durable": true,
        "auto_delete": false,
        "internal": false,
        "arguments": {

        }
    },
    {
        "name": "amq.rabbitmq.log",
        "vhost": "/",
        "type": "topic",
        "durable": true,
        "auto_delete": false,
        "internal": true,
        "arguments": {

        }
    },
    {
        "name": "amq.rabbitmq.trace",
        "vhost": "/",
        "type": "topic",
        "durable": true,
        "auto_delete": false,
        "internal": true,
        "arguments": {

        }
    },
    {
        "name": "amq.topic",
        "vhost": "/",
        "type": "topic",
        "durable": true,
        "auto_delete": false,
        "internal": false,
        "arguments": {

        }
    },
    {
        "name": "test-direct",
        "vhost": "/",
        "type": "direct",
        "durable": true,
        "auto_delete": true,
        "internal": true,
        "arguments": {

        }
    },
    {
        "name": "test-fanout",
        "vhost": "/",
        "type": "fanout",
        "durable": true,
        "auto_delete": false,
        "internal": false,
        "arguments": {

        }
    },
    {
        "name": "test-headers",
        "vhost": "/",
        "type": "headers",
        "durable": true,
        "auto_delete": true,
        "internal": false,
        "arguments": {

        }
    },
    {
        "name": "test-topic",
        "vhost": "/",
        "type": "topic",
        "durable": true,
        "auto_delete": false,
        "internal": false,
        "arguments": {

        }
    }
]
`

// result of GET /api/queues
var queueResult = `
[
    {
        "messages_details": {
            "rate": 100.0
        },
        "messages": 999,
        "messages_unacknowledged_details": {
            "rate": 0.0
        },
        "messages_unacknowledged": 0,
        "messages_ready_details": {
            "rate": 0.0
        },
        "messages_ready": 0,
        "reductions_details": {
            "rate": 0.0
        },
        "reductions": 4298,
        "node": "rabbit@08f57d1fe8ab",
        "arguments": {

        },
        "exclusive": false,
        "auto_delete": false,
        "durable": true,
        "vhost": "/",
        "name": "direct-q1",
        "message_bytes_paged_out": 0,
        "messages_paged_out": 0,
        "backing_queue_status": {
            "mode": "default",
            "q1": 0,
            "q2": 0,
            "delta": [
                "delta",
                "undefined",
                0,
                0,
                "undefined"
            ],
            "q3": 0,
            "q4": 0,
            "len": 0,
            "target_ram_count": "infinity",
            "next_seq_id": 0,
            "avg_ingress_rate": 0.0,
            "avg_egress_rate": 0.0,
            "avg_ack_ingress_rate": 0.0,
            "avg_ack_egress_rate": 0.0
        },
        "head_message_timestamp": null,
        "message_bytes_persistent": 0,
        "message_bytes_ram": 0,
        "message_bytes_unacknowledged": 0,
        "message_bytes_ready": 0,
        "message_bytes": 0,
        "messages_persistent": 0,
        "messages_unacknowledged_ram": 0,
        "messages_ready_ram": 0,
        "messages_ram": 0,
        "garbage_collection": {
            "minor_gcs": 4,
            "fullsweep_after": 65535,
            "min_heap_size": 233,
            "min_bin_vheap_size": 46422,
            "max_heap_size": 0
        },
        "state": "running",
        "recoverable_slaves": null,
        "consumers": 4,
        "exclusive_consumer_tag": null,
        "policy": null,
        "consumer_utilisation": null,
        "memory": 29840
    },
    {
        "messages_details": {
            "rate": 0.0
        },
        "messages": 0,
        "messages_unacknowledged_details": {
            "rate": 0.0
        },
        "messages_unacknowledged": 0,
        "messages_ready_details": {
            "rate": 0.0
        },
        "messages_ready": 0,
        "reductions_details": {
            "rate": 0.0
        },
        "reductions": 4298,
        "node": "rabbit@08f57d1fe8ab",
        "arguments": {

        },
        "exclusive": false,
        "auto_delete": false,
        "durable": true,
        "vhost": "/",
        "name": "direct-q2",
        "message_bytes_paged_out": 0,
        "messages_paged_out": 0,
        "backing_queue_status": {
            "mode": "default",
            "q1": 0,
            "q2": 0,
            "delta": [
                "delta",
                "undefined",
                0,
                0,
                "undefined"
            ],
            "q3": 0,
            "q4": 0,
            "len": 0,
            "target_ram_count": "infinity",
            "next_seq_id": 0,
            "avg_ingress_rate": 0.0,
            "avg_egress_rate": 0.0,
            "avg_ack_ingress_rate": 0.0,
            "avg_ack_egress_rate": 0.0
        },
        "head_message_timestamp": null,
        "message_bytes_persistent": 0,
        "message_bytes_ram": 0,
        "message_bytes_unacknowledged": 0,
        "message_bytes_ready": 0,
        "message_bytes": 0,
        "messages_persistent": 0,
        "messages_unacknowledged_ram": 0,
        "messages_ready_ram": 0,
        "messages_ram": 0,
        "garbage_collection": {
            "minor_gcs": 4,
            "fullsweep_after": 65535,
            "min_heap_size": 233,
            "min_bin_vheap_size": 46422,
            "max_heap_size": 0
        },
        "state": "running",
        "recoverable_slaves": null,
        "consumers": 0,
        "exclusive_consumer_tag": null,
        "policy": null,
        "consumer_utilisation": null,
        "memory": 29840
    },
    {
        "messages_details": {
            "rate": 0.0
        },
        "messages": 0,
        "messages_unacknowledged_details": {
            "rate": 0.0
        },
        "messages_unacknowledged": 0,
        "messages_ready_details": {
            "rate": 0.0
        },
        "messages_ready": 0,
        "reductions_details": {
            "rate": 0.0
        },
        "reductions": 5012,
        "node": "rabbit@08f57d1fe8ab",
        "arguments": {

        },
        "exclusive": false,
        "auto_delete": false,
        "durable": true,
        "vhost": "/",
        "name": "fanout-q1",
        "message_bytes_paged_out": 0,
        "messages_paged_out": 0,
        "backing_queue_status": {
            "mode": "default",
            "q1": 0,
            "q2": 0,
            "delta": [
                "delta",
                "undefined",
                0,
                0,
                "undefined"
            ],
            "q3": 0,
            "q4": 0,
            "len": 0,
            "target_ram_count": "infinity",
            "next_seq_id": 0,
            "avg_ingress_rate": 0.0,
            "avg_egress_rate": 0.0,
            "avg_ack_ingress_rate": 0.0,
            "avg_ack_egress_rate": 0.0
        },
        "head_message_timestamp": null,
        "message_bytes_persistent": 0,
        "message_bytes_ram": 0,
        "message_bytes_unacknowledged": 0,
        "message_bytes_ready": 0,
        "message_bytes": 0,
        "messages_persistent": 0,
        "messages_unacknowledged_ram": 0,
        "messages_ready_ram": 0,
        "messages_ram": 0,
        "garbage_collection": {
            "minor_gcs": 5,
            "fullsweep_after": 65535,
            "min_heap_size": 233,
            "min_bin_vheap_size": 46422,
            "max_heap_size": 0
        },
        "state": "running",
        "recoverable_slaves": null,
        "consumers": 0,
        "exclusive_consumer_tag": null,
        "policy": null,
        "consumer_utilisation": null,
        "idle_since": "2017-05-25 19:14:32",
        "memory": 29840
    },
    {
        "messages_details": {
            "rate": 0.0
        },
        "messages": 0,
        "messages_unacknowledged_details": {
            "rate": 0.0
        },
        "messages_unacknowledged": 0,
        "messages_ready_details": {
            "rate": 0.0
        },
        "messages_ready": 0,
        "reductions_details": {
            "rate": 0.0
        },
        "reductions": 4298,
        "node": "rabbit@08f57d1fe8ab",
        "arguments": {

        },
        "exclusive": false,
        "auto_delete": false,
        "durable": true,
        "vhost": "/",
        "name": "fanout-q2",
        "message_bytes_paged_out": 0,
        "messages_paged_out": 0,
        "backing_queue_status": {
            "mode": "default",
            "q1": 0,
            "q2": 0,
            "delta": [
                "delta",
                "undefined",
                0,
                0,
                "undefined"
            ],
            "q3": 0,
            "q4": 0,
            "len": 0,
            "target_ram_count": "infinity",
            "next_seq_id": 0,
            "avg_ingress_rate": 0.0,
            "avg_egress_rate": 0.0,
            "avg_ack_ingress_rate": 0.0,
            "avg_ack_egress_rate": 0.0
        },
        "head_message_timestamp": null,
        "message_bytes_persistent": 0,
        "message_bytes_ram": 0,
        "message_bytes_unacknowledged": 0,
        "message_bytes_ready": 0,
        "message_bytes": 0,
        "messages_persistent": 0,
        "messages_unacknowledged_ram": 0,
        "messages_ready_ram": 0,
        "messages_ram": 0,
        "garbage_collection": {
            "minor_gcs": 4,
            "fullsweep_after": 65535,
            "min_heap_size": 233,
            "min_bin_vheap_size": 46422,
            "max_heap_size": 0
        },
        "state": "running",
        "recoverable_slaves": null,
        "consumers": 0,
        "exclusive_consumer_tag": null,
        "policy": null,
        "consumer_utilisation": null,
        "idle_since": "2017-05-25 19:14:32",
        "memory": 29840
    },
    {
        "messages_details": {
            "rate": 0.0
        },
        "messages": 0,
        "messages_unacknowledged_details": {
            "rate": 0.0
        },
        "messages_unacknowledged": 0,
        "messages_ready_details": {
            "rate": 0.0
        },
        "messages_ready": 0,
        "reductions_details": {
            "rate": 0.0
        },
        "reductions": 6440,
        "node": "rabbit@08f57d1fe8ab",
        "arguments": {

        },
        "exclusive": false,
        "auto_delete": false,
        "durable": true,
        "vhost": "/",
        "name": "header-q1",
        "message_bytes_paged_out": 0,
        "messages_paged_out": 0,
        "backing_queue_status": {
            "mode": "default",
            "q1": 0,
            "q2": 0,
            "delta": [
                "delta",
                "undefined",
                0,
                0,
                "undefined"
            ],
            "q3": 0,
            "q4": 0,
            "len": 0,
            "target_ram_count": "infinity",
            "next_seq_id": 0,
            "avg_ingress_rate": 0.0,
            "avg_egress_rate": 0.0,
            "avg_ack_ingress_rate": 0.0,
            "avg_ack_egress_rate": 0.0
        },
        "head_message_timestamp": null,
        "message_bytes_persistent": 0,
        "message_bytes_ram": 0,
        "message_bytes_unacknowledged": 0,
        "message_bytes_ready": 0,
        "message_bytes": 0,
        "messages_persistent": 0,
        "messages_unacknowledged_ram": 0,
        "messages_ready_ram": 0,
        "messages_ram": 0,
        "garbage_collection": {
            "minor_gcs": 7,
            "fullsweep_after": 65535,
            "min_heap_size": 233,
            "min_bin_vheap_size": 46422,
            "max_heap_size": 0
        },
        "state": "running",
        "recoverable_slaves": null,
        "consumers": 0,
        "exclusive_consumer_tag": null,
        "policy": null,
        "consumer_utilisation": null,
        "idle_since": "2017-05-25 19:14:53",
        "memory": 29840
    },
    {
        "messages_details": {
            "rate": 0.0
        },
        "messages": 0,
        "messages_unacknowledged_details": {
            "rate": 0.0
        },
        "messages_unacknowledged": 0,
        "messages_ready_details": {
            "rate": 0.0
        },
        "messages_ready": 0,
        "reductions_details": {
            "rate": 0.0
        },
        "reductions": 5012,
        "node": "rabbit@08f57d1fe8ab",
        "arguments": {

        },
        "exclusive": false,
        "auto_delete": false,
        "durable": true,
        "vhost": "/",
        "name": "header-q2",
        "message_bytes_paged_out": 0,
        "messages_paged_out": 0,
        "backing_queue_status": {
            "mode": "default",
            "q1": 0,
            "q2": 0,
            "delta": [
                "delta",
                "undefined",
                0,
                0,
                "undefined"
            ],
            "q3": 0,
            "q4": 0,
            "len": 0,
            "target_ram_count": "infinity",
            "next_seq_id": 0,
            "avg_ingress_rate": 0.0,
            "avg_egress_rate": 0.0,
            "avg_ack_ingress_rate": 0.0,
            "avg_ack_egress_rate": 0.0
        },
        "head_message_timestamp": null,
        "message_bytes_persistent": 0,
        "message_bytes_ram": 0,
        "message_bytes_unacknowledged": 0,
        "message_bytes_ready": 0,
        "message_bytes": 0,
        "messages_persistent": 0,
        "messages_unacknowledged_ram": 0,
        "messages_ready_ram": 0,
        "messages_ram": 0,
        "garbage_collection": {
            "minor_gcs": 5,
            "fullsweep_after": 65535,
            "min_heap_size": 233,
            "min_bin_vheap_size": 46422,
            "max_heap_size": 0
        },
        "state": "running",
        "recoverable_slaves": null,
        "consumers": 0,
        "exclusive_consumer_tag": null,
        "policy": null,
        "consumer_utilisation": null,
        "idle_since": "2017-05-25 19:14:47",
        "memory": 29840
    },
    {
        "messages_details": {
            "rate": 0.0
        },
        "messages": 0,
        "messages_unacknowledged_details": {
            "rate": 0.0
        },
        "messages_unacknowledged": 0,
        "messages_ready_details": {
            "rate": 0.0
        },
        "messages_ready": 0,
        "reductions_details": {
            "rate": 0.0
        },
        "reductions": 4297,
        "node": "rabbit@08f57d1fe8ab",
        "arguments": {

        },
        "exclusive": true,
        "auto_delete": true,
        "durable": true,
        "vhost": "/",
        "name": "topic-q1",
        "message_bytes_paged_out": 0,
        "messages_paged_out": 0,
        "backing_queue_status": {
            "mode": "default",
            "q1": 0,
            "q2": 0,
            "delta": [
                "delta",
                "undefined",
                0,
                0,
                "undefined"
            ],
            "q3": 0,
            "q4": 0,
            "len": 0,
            "target_ram_count": "infinity",
            "next_seq_id": 0,
            "avg_ingress_rate": 0.0,
            "avg_egress_rate": 0.0,
            "avg_ack_ingress_rate": 0.0,
            "avg_ack_egress_rate": 0.0
        },
        "head_message_timestamp": null,
        "message_bytes_persistent": 0,
        "message_bytes_ram": 0,
        "message_bytes_unacknowledged": 0,
        "message_bytes_ready": 0,
        "message_bytes": 0,
        "messages_persistent": 0,
        "messages_unacknowledged_ram": 0,
        "messages_ready_ram": 0,
        "messages_ram": 0,
        "garbage_collection": {
            "minor_gcs": 4,
            "fullsweep_after": 65535,
            "min_heap_size": 233,
            "min_bin_vheap_size": 46422,
            "max_heap_size": 0
        },
        "state": "running",
        "recoverable_slaves": null,
        "consumers": 0,
        "exclusive_consumer_tag": null,
        "policy": null,
        "consumer_utilisation": null,
        "idle_since": "2017-05-25 19:14:17",
        "memory": 29840
    },
    {
        "messages_details": {
            "rate": 0.0
        },
        "messages": 0,
        "messages_unacknowledged_details": {
            "rate": 0.0
        },
        "messages_unacknowledged": 0,
        "messages_ready_details": {
            "rate": 0.0
        },
        "messages_ready": 0,
        "reductions_details": {
            "rate": 0.0
        },
        "reductions": 4297,
        "node": "rabbit@08f57d1fe8ab",
        "arguments": {

        },
        "exclusive": false,
        "auto_delete": false,
        "durable": true,
        "vhost": "/",
        "name": "topic-q2",
        "message_bytes_paged_out": 0,
        "messages_paged_out": 0,
        "backing_queue_status": {
            "mode": "default",
            "q1": 0,
            "q2": 0,
            "delta": [
                "delta",
                "undefined",
                0,
                0,
                "undefined"
            ],
            "q3": 0,
            "q4": 0,
            "len": 0,
            "target_ram_count": "infinity",
            "next_seq_id": 0,
            "avg_ingress_rate": 0.0,
            "avg_egress_rate": 0.0,
            "avg_ack_ingress_rate": 0.0,
            "avg_ack_egress_rate": 0.0
        },
        "head_message_timestamp": null,
        "message_bytes_persistent": 0,
        "message_bytes_ram": 0,
        "message_bytes_unacknowledged": 0,
        "message_bytes_ready": 0,
        "message_bytes": 0,
        "messages_persistent": 0,
        "messages_unacknowledged_ram": 0,
        "messages_ready_ram": 0,
        "messages_ram": 0,
        "garbage_collection": {
            "minor_gcs": 4,
            "fullsweep_after": 65535,
            "min_heap_size": 233,
            "min_bin_vheap_size": 46422,
            "max_heap_size": 0
        },
        "state": "running",
        "recoverable_slaves": null,
        "consumers": 0,
        "exclusive_consumer_tag": null,
        "policy": null,
        "consumer_utilisation": null,
        "idle_since": "2017-05-25 19:14:21",
        "memory": 29840
    }
]
`

var overviewResult = `
{
    "management_version": "3.6.9",
    "rates_mode": "basic",
    "exchange_types": [
        {
            "name": "headers",
            "description": "AMQP headers exchange, as per the AMQP specification",
            "enabled": true
        },
        {
            "name": "topic",
            "description": "AMQP topic exchange, as per the AMQP specification",
            "enabled": true
        },
        {
            "name": "fanout",
            "description": "AMQP fanout exchange, as per the AMQP specification",
            "enabled": true
        },
        {
            "name": "direct",
            "description": "AMQP direct exchange, as per the AMQP specification",
            "enabled": true
        }
    ],
    "rabbitmq_version": "3.6.9",
    "cluster_name": "rabbit@08f57d1fe8ab",
    "erlang_version": "19.3",
    "erlang_full_version": "Erlang/OTP 19 [erts-8.3] [source] [64-bit] [smp:4:4] [async-threads:64] [hipe] [kernel-poll:true]",
    "message_stats": {
        "publish": 1200,
        "publish_details": {
            "rate": 0.0
        },
        "confirm": 0,
        "confirm_details": {
            "rate": 0.0
        },
        "return_unroutable": 0,
        "return_unroutable_details": {
            "rate": 0.0
        },
        "disk_reads": 0,
        "disk_reads_details": {
            "rate": 0.0
        },
        "disk_writes": 0,
        "disk_writes_details": {
            "rate": 0.0
        },
        "get": 0,
        "get_details": {
            "rate": 0.0
        },
        "get_no_ack": 0,
        "get_no_ack_details": {
            "rate": 0.0
        },
        "deliver": 1200,
        "deliver_details": {
            "rate": 0.0
        },
        "deliver_no_ack": 960,
        "deliver_no_ack_details": {
            "rate": 0.0
        },
        "redeliver": 0,
        "redeliver_details": {
            "rate": 0.0
        },
        "ack": 0,
        "ack_details": {
            "rate": 0.0
        },
        "deliver_get": 2160,
        "deliver_get_details": {
            "rate": 0.0
        }
    },
    "queue_totals": {
        "messages_ready": 0,
        "messages_ready_details": {
            "rate": 0.0
        },
        "messages_unacknowledged": 0,
        "messages_unacknowledged_details": {
            "rate": 0.0
        },
        "messages": 0,
        "messages_details": {
            "rate": 0.0
        }
    },
    "object_totals": {
        "consumers": 0,
        "queues": 8,
        "exchanges": 12,
        "connections": 0,
        "channels": 0
    },
    "statistics_db_event_queue": 0,
    "node": "rabbit@08f57d1fe8ab",
    "listeners": [
        {
            "node": "rabbit@08f57d1fe8ab",
            "protocol": "amqp",
            "ip_address": "::",
            "port": 5672,
            "socket_opts": {
                "backlog": 128,
                "nodelay": true,
                "linger": [
                    true,
                    0
                ],
                "exit_on_close": false
            }
        },
        {
            "node": "rabbit@08f57d1fe8ab",
            "protocol": "clustering",
            "ip_address": "::",
            "port": 25672,
            "socket_opts": [

            ]
        },
        {
            "node": "rabbit@08f57d1fe8ab",
            "protocol": "http",
            "ip_address": "::",
            "port": 15672,
            "socket_opts": {
                "port": 15672,
                "ssl": false
            }
        }
    ],
    "contexts": [
        {
            "node": "rabbit@08f57d1fe8ab",
            "description": "RabbitMQ Management",
            "path": "/",
            "port": "15672",
            "ssl": "false"
        }
    ]
}`

var consumerResult = ` 
[
    {
        "arguments": [

        ],
        "prefetch_count": 0,
        "ack_required": false,
        "exclusive": true,
        "consumer_tag": "some_consumer",
        "channel_details": {
            "peer_host": "172.17.0.1",
            "peer_port": 58938,
            "connection_name": "172.17.0.1:58938 -> 172.17.0.2:5672",
            "user": "guest",
            "number": 2,
            "node": "rabbit@35b655845dfd",
            "name": "172.17.0.1:58938 -> 172.17.0.2:5672 (2)"
        },
        "queue": {
            "vhost": "/",
            "name": "direct-q1"
        }
    }
]
`
