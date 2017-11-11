# rabtap - RabbitMQ wire tap

[![Build Status](https://travis-ci.org/jandelgado/rabtap.svg?branch=master)](https://travis-ci.org/jandelgado/rabtap)
[![Coverage Status](https://coveralls.io/repos/github/jandelgado/rabtap/badge.svg?branch=master)](https://coveralls.io/github/jandelgado/rabtap?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/jandelgado/rabtap)](https://goreportcard.com/report/github.com/jandelgado/rabtap)

Rabtap helps you understand what's going on in your RabbitMQ message broker and
your distributed apps.

<!-- vim-markdown-toc GFM -->

* [Features](#features)
* [Screenshots](#screenshots)
* [Installation](#installation)
* [Usage](#usage)
    * [Broker URI specification](#broker-uri-specification)
    * [Environment variables](#environment-variables)
        * [Default RabbitMQ broker](#default-rabbitmq-broker)
        * [Default RabbitMQ management API endpoint](#default-rabbitmq-management-api-endpoint)
    * [Examples](#examples)
        * [Broker info](#broker-info)
        * [Wire-tapping messages](#wire-tapping-messages)
        * [Message recorder](#message-recorder)
        * [Send Messages](#send-messages)
        * [Poor mans shovel](#poor-mans-shovel)
* [JSON message format](#json-message-format)
* [Build from source](#build-from-source)
* [Test data generator](#test-data-generator)
* [Author](#author)
* [Copyright and license](#copyright-and-license)

<!-- vim-markdown-toc -->

## Features

* display messages being sent to exchanges using RabbitMQ
  exchange-to-exchange bindings without affecting actual message delivery (aka _tapping_)
* display broker related information using the
  [RabbitMQ REST management API](https://rawcdn.githack.com/rabbitmq/rabbitmq-management/rabbitmq_v3_6_14/priv/www/api/index.html)
* save messages and meta data for later analysis and replay
* send messages to exchanges
* TLS support
* no runtime dependencies (statically linked go single file binary)
* simple to use command line tool
* runs on Linux, Windows, Mac and wherever you can compile go

## Screenshots

Output of `rabtap info` command:

![info mode](doc/images/info.png)

Output of rabtap running in `tap` mode, showing message meta data
with unset attributes filtered out and the message body:

![info mode](doc/images/tap.png)

## Installation

Pre-compiled binaries can be downloaded for multiple platforms from the
[releases page](releases/).

See [below](#build-from-source) if you prefer to compile from source.

## Usage

```
rabtap - RabbitMQ message tap.

Usage:
  rabtap tap [--uri URI] EXCHANGES [--saveto=DIR] [-jkvn]
  rabtap (tap --uri URI EXCHANGES)... [--saveto=DIR] [-jkvn]
  rabtap send [--uri URI] EXCHANGE [FILE] [--routingkey KEY] [-jkv]
  rabtap info [--api APIURI] [--consumers] [--stats] [--show-default] [-kvn]
  rabtap -h|--help

Examples:
  rabtap tap --uri amqp://guest:guest@localhost/ amq.fanout:
  rabtap tap --uri amqp://guest:guest@localhost/ amq.topic:#,amq.fanout:
  rabtap send --uri amqp://guest:guest@localhost/ amq.topic message.JSON -j
  rabtap info --api http://guest:guest@localhost:15672/api

Options:
 -h, --help           print this help.
 --uri URI            connect to given AQMP broker. If omitted, the
                      environment variable RABTAP_AMQPURI will be used.
 EXCHANGES            comma-separated list of exchanges and routing keys,
                      e.g. amq.topic:# or exchange1:key1,exchange2:key2.
 EXCHANGE             name of an exchange, e.g. amq.direct.
 FILE                 file to send with in send mode. If omitted, stdin will
                      be read.
 --saveto DIR         also save messages and metadata to DIR.
 -j, --json           print/save/send message metadata and body to a
                      single JSON file. JSON body is base64 encoded. Otherwise
                      metadata and body (as-is) are saved separately.
 -r, --routingkey KEY routing key to use in send mode.
 --api APIURI         connect to given API server. If APIURI is omitted,
                      the environment variable RABTAP_APIURI will be used.
 -n, --no-color       don't color output.
 --consumers          include consumers in output of info command.
 --stats              include statistics in output of info command.
 --show-default       include default exchange in output info command.
 -k, --insecure       allow insecure TLS connections (no certificate check).
 -v, --verbose        enable verbose mode.
```

### Broker URI specification

The specification of the RabbitMQ broker URI follows the [AMQP URI
specification](https://www.rabbitmq.com/uri-spec.html) as implemented by the
[go RabbitMQ client library](https://github.com/streadway/amqp).

### Environment variables

#### Default RabbitMQ broker

In cases where the URI argument is optional, e.g. `rabtap tap [-uri
URI] exchange ...`, the URI of the RabbitMQ broker can be set with the
environment variable `RABTAP_AMQPURI`.  Example:

```
$ export RABTAP_AMQPURI=amqp://guest:guest@localhost:5672/
$ rabtap tap amq.fanout:
...
```

#### Default RabbitMQ management API endpoint

The default RabbitMQ management API URI can be set using the `RABTAP_APIURI`
environment variable. Example:

```
$ export RABTAP_APIURI=http://guest:guest@localhost:15672/
$ rabtap info
...
```

### Examples

The following examples expect a RabbitMQ broker running on localhost:5672 and
the management API available on port 15672. Easiest way to start such an
instance is by running `docker run -ti --rm -p 5672:5672 -p 15672:15672
rabbitmq:3-management` or similar command to start a RabbitMQ container.

#### Broker info

* `$ rabtap info --api http://guest:guest@localhost:15672/api --consumers` -
  shows exchanges, queues and consumers of given broker in an tree view (see
  [screenshot](#screenshots)).

#### Wire-tapping messages

The `tap` command allows to tap to multiple exchanges, with optionally
specifying binding keys. The syntax of the `tap` command is `rabtap tap [--uri
URI] EXCHANGES` where the `EXCHANGES` argument specifies the exchanges and
binding keys to use. The `EXCHANGES` argument is of the form
`EXCHANGE:[KEY][,EXCHANGE:[KEY]]*`.

The acutal format of the binding key depends on the exchange type (e.g.
direct, topic, headers) and is described in the [RabbitMQ
documentation](https://www.rabbitmq.com/tutorials/amqp-concepts.html).

Some examples:

* `#` on  an exchange of `type topic` will make the tap receive all messages
  on the exchange.
* a valid queue name for an exchange of `type direct` binds exactly to messages
  destined for this queue
* an empty binding key for exchanges of `type  fanout` or `type headers` will
  receive all messages published to these exchanges

Note: on exchanges of type `headers` the binding key is currently ignored and
all messages are received by the tap.

The following examples assume that the `RABTAP_AMQPURI` environment variable is
set, otherwise you have to pass the additional `--uri URI` parameter to the
commands below.

* `$ rabtap tap my-topic-exchange:#`
* `$ rabtap tap my-fanout-exchange:`
* `$ rabtap tap my-headers-exchange:`
* `$ rabtap tap my-direct-exchange:binding-key`

The following example connects to multiple exchanges:

* `$ rabtap tap my-fanout-exchange:,my-topic-exchange:#,my-other-exchange:binding-key`

Rabtap allows you to connect simultaneously to multiple brokers and
exchanges:

* `$ raptap tap --uri amqp://broker1 amq.topic:# tap --uri amqp://broker2 amq.fanout:`

#### Message recorder

All tapped messages can be also be saved for later analysis or replay.

* `$ rabtap tap amq.topic:# --saveto /tmp` - saves messages as pair of
  files consisting of raw message body and JSON meta data file to `/tmp`
  directory.
* `$ rabtap tap amq.topic:# --saveto /tmp --json` - saves messages as JSON
  files to `/tmp` directory.

Files are created with file name `rabtap-`+`<Unix-Nano-Timestamp>`+ `.` + `<extension>`.

#### Send Messages

* `$ rabtap send amq.direct -r routingKey message.json --json`  - Send
  message(s) in JSON format to exchange `amq.direct` with routing key
  `routingKey`.
* `$ cat message.json | rabtap send amqp.direct -r routingKey --json` - same
  as above, but read message(s) from stdin.

#### Poor mans shovel

Rabtap instances can be connected through a pipe and messages will be read on
one side and send to the other. Note that for send to work in streaming mode,
the JSON mode (`--json`) must be used on both sides, so that messages are
encapsulated in JSON messages.

```
$ rabtap tap --uri amqp://broker1 my-topic-exchange:# --json | \
  rabtap send --uri amqp://broker2 amq.direct -r routingKey --json
```

## JSON message format

When using the `--json` option, messages are print/read as a stream of JSON
messages in the following format:

```
...
{
  "ContentType": "text/plain",
  "ContentEncoding": "",
  "DeliveryMode": 0,
  "Priority": 0,
  "CorrelationID": "",
  "ReplyTo": "",
  "Expiration": "",
  "MessageID": "",
  "Timestamp": "2017-11-10T00:13:38+01:00",
  "Type": "",
  "UserID": "",
  "AppID": "rabtap.testgen",
  "DeliveryTag": 27,
  "Redelivered": false,
  "Exchange": "amq.topic",
  "RoutingKey": "test-q-amq.topic-0",
  "Body": "dGhpcyB0ZXN0IG1lc3NhZ2U .... IGFuZCBoZWFkZXJzIGFtcXAuVGFibGV7fQ=="
}
...
```

Note that in JSON mode, the `Body` is base64 encoded.

## Build from source

To build rabtap from source, you need [go]() and the following tools installed:

Build dependencies:

* [dep](https://github.com/golang/dep)
* [ineffassign](https://github.com/gordonklaus/ineffassign)
* [misspell](https://github.com/client9/misspell/cmd/misspell)
* [golint](https://github.com/golang/lint/golint)
* [gocyclo](https://github.com/fzipp/gocyclo)

TODO

## Test data generator

A simple [test data generator tool](app/testgen/README.md) for manual tests is
included in the `app/testgen` directory.

## Author

Jan Delgado (jdelgado at gmx dot net)

## Copyright and license

Copyright (c) 2017 Jan Delgado.
rabtap is licensed under the GPLv3 license.

