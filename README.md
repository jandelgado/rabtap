# rabtap - RabbitMQ wire tap

![Build Status](https://github.com/jandelgado/rabtap/workflows/run%20tests/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/jandelgado/rabtap/badge.svg?branch=master)](https://coveralls.io/github/jandelgado/rabtap?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/jandelgado/rabtap)](https://goreportcard.com/report/github.com/jandelgado/rabtap)

Swiss army knife for RabbitMQ. Tap/Pub/Sub messages, create/delete/bind queues
and exchanges, inspect broker.

## Contents

<!-- vim-markdown-toc GFM -->

* [Features](#features)
* [Screenshots](#screenshots)
  * [Show broker topology](#show-broker-topology)
  * [Visualize broker topology with graphviz](#visualize-broker-topology-with-graphviz)
  * [Tap messages](#tap-messages)
* [Installation](#installation)
  * [Docker image](#docker-image)
  * [Pre-compiled binaries](#pre-compiled-binaries)
  * [Arch Linux](#arch-linux)
  * [Installation from source](#installation-from-source)
* [Usage](#usage)
  * [Command overview](#command-overview)
  * [Broker URI specification](#broker-uri-specification)
    * [Authentication](#authentication)
  * [Environment variables](#environment-variables)
    * [Default RabbitMQ broker](#default-rabbitmq-broker)
    * [Default RabbitMQ management API endpoint](#default-rabbitmq-management-api-endpoint)
    * [Default RabbitMQ TLS config](#default-rabbitmq-tls-config)
    * [Colored output](#colored-output)
  * [Command reference and examples](#command-reference-and-examples)
    * [Broker info](#broker-info)
    * [Wire-tapping messages](#wire-tapping-messages)
      * [Tap all messages published or delivered (RabbitMQ FireHose)](#tap-all-messages-published-or-delivered-rabbitmq-firehose)
        * [Replaying messages from the FireHose exchange](#replaying-messages-from-the-firehose-exchange)
      * [Connect to multiple brokers](#connect-to-multiple-brokers)
      * [Message recorder](#message-recorder)
    * [Subscribe messages](#subscribe-messages)
    * [Publish messages](#publish-messages)
    * [Poor mans shovel](#poor-mans-shovel)
    * [Close connection](#close-connection)
    * [Exchange commands](#exchange-commands)
    * [Queue commands](#queue-commands)
  * [Format specification for tap and sub command](#format-specification-for-tap-and-sub-command)
  * [JSON message format](#json-message-format)
  * [Filtering output](#filtering-output)
    * [Filtering expressions](#filtering-expressions)
      * [Evaluation context](#evaluation-context)
      * [Examples](#examples)
    * [Type reference](#type-reference)
      * [Exchange type](#exchange-type)
      * [Queue type](#queue-type)
      * [Binding type](#binding-type)
      * [Connection type](#connection-type)
      * [Channel type](#channel-type)
      * [Message type](#message-type)
* [Build from source](#build-from-source)
  * [Download and build using go install](#download-and-build-using-go-install)
  * [Build using Makefile and tests](#build-using-makefile-and-tests)
  * [Experimental WASM/wasip1 port](#experimental-wasmwasip1-port)
* [Test data generator](#test-data-generator)
* [Contributing](#contributing)
* [Author](#author)
* [Copyright and license](#copyright-and-license)

<!-- vim-markdown-toc -->

## Features

* tap to messages being sent to exchanges using RabbitMQ
  exchange-to-exchange bindings without affecting actual message delivery (aka _tapping_)
* display broker related information using the
  [RabbitMQ REST management API](https://rawcdn.githack.com/rabbitmq/rabbitmq-management/rabbitmq_v3_6_14/priv/www/api/index.html)
* save messages and meta data for later analysis and replay
* publish messages to exchanges
* consume messages from queues and streams (subscribe)
* filter output using flexible expressions
* supports TLS
* no runtime dependencies (statically linked golang single file binary)
* simple to use command line tool
* runs on Linux, Windows, Mac and wherever you can compile go

## Screenshots

### Show broker topology

Output of `rabtap info` command:

![info mode](doc/images/info.png)

Output of `rabtap info --stats` command, showing additional statistics:

![info mode](doc/images/info-stats.png)

### Visualize broker topology with graphviz

Using the `--format=dot` option, the `info` command can generate output in the
[dot format](https://graphviz.org/doc/info/lang.html), which can be visualized
using graphviz, e.g. `rabtap info --show-default --format dot | dot -T svg >
mybroker.svg`. The resulting SVG file can be visualized with a web browser.

![info mode](doc/images/info-dot.png)

### Tap messages

Output of rabtap in `tap` mode, showing message meta data and the message body:

![info mode](doc/images/tap.png)

## Installation

### Docker image

A docker image is provided so rabtap can be used as a docker container, e.g.

```console
$ docker run --rm -ti ghcr.io/jandelgado/rabtap:latest
```

### Pre-compiled binaries

Pre-compiled binaries can be downloaded for multiple platforms from the
[releases page](https://github.com/jandelgado/rabtap/releases).

### Arch Linux

Rabtap can be installed from the Arch Linux User Repository (AUR):

```console
$ yay -S rabtap
```

### Installation from source

See the [build from source section](#build-from-source) if you prefer to
compile from source.

## Usage

```
rabtap - RabbitMQ wire tap.                    github.com/jandelgado/rabtap

Usage:
  rabtap -h|--help
  rabtap info [--api=APIURI] [--consumers] [--stats] [--filter=EXPR] [--omit-empty]
              [--show-default] [--mode=MODE] [--format=FORMAT] [-kncv]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap tap EXCHANGES [--uri=URI] [--saveto=DIR] [--format=FORMAT]  [--limit=NUM]
          [--idle-timeout=DURATION] [--filter=EXPR] [-jkncsv]
          [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap (tap --uri=URI EXCHANGES)... [--saveto=DIR] [--format=FORMAT]  [--limit=NUM]
          [--idle-timeout=DURATION] [--filter=EXPR] [-jkncsv]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap sub QUEUE [--uri URI] [--saveto=DIR] [--format=FORMAT] [--limit=NUM]
              [--offset=OFFSET] [--args=KV]... [(--reject [--requeue])] [-jkcsvn]
              [--filter=EXPR] [--idle-timeout=DURATION]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap pub  [--uri=URI] [SOURCE] [--exchange=EXCHANGE] [--format=FORMAT]
              [--routingkey=KEY | (--header=KV)...]
              [--confirms] [--mandatory] [--delay=DELAY | --speed=FACTOR] [-jkv]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap exchange create EXCHANGE [--uri=URI] [--type=TYPE] [--args=KV]... [-kv]
              [--autodelete] [--durable]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap exchange bind EXCHANGE to DESTEXCHANGE [--uri=URI] [-kv]
              (--bindingkey=KEY | (--header=KV)... (--all|--any))
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap exchange rm EXCHANGE [--uri=URI] [-kv]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap queue create QUEUE [--uri=URI] [--queue-type=TYPE] [--args=KV]... [-kv]
              [--autodelete] [--durable] [--lazy]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap queue bind QUEUE to EXCHANGE [--uri=URI] [-kv]
              (--bindingkey=KEY | (--header=KV)... (--all|--any))
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap queue unbind QUEUE from EXCHANGE [--uri=URI] [-kv]
              (--bindingkey=KEY | (--header=KV)... (--all|--any))
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap queue rm QUEUE [--uri=URI] [-kv]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap queue purge QUEUE [--uri=URI] [-kv]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap conn close CONNECTION [--api=APIURI] [--reason=REASON] [-kv]
              [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
  rabtap --version

Arguments and options:
 EXCHANGES            comma-separated list of exchanges and optional binding keys,
                      e.g. amq.topic:# or exchange1:key1,exchange2:key2.
 EXCHANGE             name of an exchange, e.g. amq.direct.
 DESTEXCHANGE         name of a a destination exchange in an exchange-to-exchange binding.
 SOURCE               file or directory to publish in pub mode. If omitted, stdin will be read.
 QUEUE                name of a queue.
 CONNECTION           name of a connection.
 DIR                  directory to read messages from.
 -a, --autodelete     create auto delete exchange/queue.
 --all                set x-match=all option in header based routing.
 --any                set x-match=any option in header based routing.
 --api=APIURI         connect to given API server. If APIURL is omitted,
                      the environment variable RABTAP_APIURI will be used.
 --args=KV            A key value pair in the form of "key=value" passed as
                      additional arguments. e.g. '--args=x-queue-type=quorum'
 -b, --bindingkey=KEY binding key to use in bind queue command.
 --by-connection      output of info command starts with connections.
 -c, --color          force colored output
 --confirms           enable publisher confirms and wait for confirmations.
 --consumers          include consumers and connections in output of info command.
 --delay=DELAY        Time to wait between sending messages during publish.
                      If not set then messages will be delayed as recorded.
                      The value must be suffixed with a time unit, e.g. ms, s etc.
 -d, --durable        create durable exchange/queue.
 --exchange=EXCHANGE  Optional exchange to publish to. If omitted, exchange will
                      be taken from message being published (see JSON message format).
 --filter=EXPR        Predicate for sub, tap, info command to filter the output [default: true]
 --format=FORMAT      * for tap, pub, sub command: format to write/read messages to console
                        and optionally to file (when --saveto DIR is given).
                        Valid options are: "raw", "json", "json-nopp". Default: raw
                      * for info command: controls generated output format. Valid
                        options are: "text", "dot". Default: text
 -h, --help           print this help.
 --header=KV          A key value pair in the form of "key=value" used as a
                      routing- or binding-key. Can occur multiple times.
 --idle-timeout=DURATION end reading messages when no new message was received
                      for the given duration.  The value must be suffixed with
                      a time unit, e.g. ms, s etc.
 -j, --json           deprecated. Use "--format json" instead.
 -k, --insecure       allow insecure TLS connections (no certificate check).
 --lazy               create a lazy queue.
 --limit=NUM          Stop afer NUM messages were received. When set to 0, will
                      run until terminated [default: 0].
 --mandatory          enable mandatory publishing (messages must be delivered to queue).
 --mode=MODE          mode for info command. One of "byConnection", "byExchange".
                      [default: byExchange].
 -n, --no-color       don't colorize output (see also environment variable NO_COLOR).
 --omit-empty         don't show echanges without bindings in info command.
 --offset=OFFSET      Offset when reading from a stream. Can be 'first', 'last',
                      'next', a duration like '10m', a RFC3339-Timestamp or
                      an integer index value. Basically it is an alias for
                      '--args=x-stream-offset=OFFSET'.
 --queue-type=TYPE    type of queue [default: classic].
 --reason=REASON      reason why the connection was closed [default: closed by rabtap].
 --reject             Reject messages. Default behaviour is to acknowledge messages.
 --requeue            Instruct broker to requeue rejected message
 -r, --routingkey=KEY routing key to use in publish mode. If omitted, routing key
                      will be taken from message being published (see JSON
                      message format).
 --saveto=DIR         also save messages and metadata to DIR.
 --show-default       include default exchange in output info command.
 -s, --silent         suppress message output to stdout.
 --speed=FACTOR       Speed factor to use during publish [default: 1.0].
 --stats              include statistics in output of info command.
 -t, --type=TYPE          type of exchange [default: fanout].
 --tls-cert-file=CERTFILE A Cert file to use for client authentication.
 --tls-key-file=KEYFILE   A Key file to use for client authentication.
 --tls-ca-file=CAFILE     A CA Cert file to use with TLS.
 --uri=URI            connect to given AQMP broker. If omitted, the
                      environment variable RABTAP_AMQPURI will be used.
 -v, --verbose        enable verbose mode.
 --version            show version information and exit.

Examples:
  rabtap tap --uri amqp://guest:guest@localhost/ amq.fanout:
  rabtap tap --uri amqp://guest:guest@localhost/ amq.topic:#,amq.fanout:
  rabtap pub --uri amqp://guest:guest@localhost/ --exchange amq.topic message.json --format=json
  rabtap info --api http://guest:guest@localhost:15672/api

  # use RABTAP_AMQPURI environment variable to specify broker instead of --uri
  export RABTAP_AMQPURI=amqp://guest:guest@localhost:5672/
  rabtap queue create JDQ
  rabtap queue bind JDQ to amq.topic --bindingkey=key
  echo "Hello" | rabtap pub --exchange amq.topic --routingkey "key"
  rabtap sub JDQ
  rabtap queue rm JDQ

  # use RABTAP_APIURI environment variable to specify mgmt api uri instead of --api
  export RABTAP_APIURI=http://guest:guest@localhost:15672/api
  rabtap info
  rabtap info --filter "binding.Source == 'amq.topic'" --omit-empty
  rabtap conn close "172.17.0.1:40874 -> 172.17.0.2:5672"

  # use RABTAP_TLS_CERTFILE | RABTAP_TLS_KEYFILE | RABTAP_TLS_CAFILE environments variables
  # instead of specifying --tls-cert-file=CERTFILE --tls-key-file=KEYFILE --tls-ca-file=CAFILE
```

### Command overview

Rabtap understands the following commands:

* `tap` - taps to an exchange and receives messages sent to the exchange,
   without affecting actual message delivery (using an exchange-to-exchange
   binding).
* `sub` - subscribes to a queue and consumes from the queue
* `pub` - publish messages to an exchange, optionally with the timing as recorded
* `info` - show broker related info (exchanges, queues, bindings, stats).
* `queue` - create,bind,unbind,remove or purge queues
* `exchange` - create or remove exchanges
* `conn` - close connections

See the below for detailed information.

### Broker URI specification

The specification of the RabbitMQ broker URI follows the [AMQP URI
specification](https://www.rabbitmq.com/uri-spec.html) as implemented by the
[go RabbitMQ client library](https://github.com/streadway/amqp).

Examples:
* `amqp://guest:guest@localhost:5672/`
* `amqps://guest:guest@my-broker.dev:5671/`
* `amqps://guest:guest@my-broker.dev:5671/vhost`

Note that according to [RFC3986](https://tools.ietf.org/html/rfc3986) it might be
necessary to escape certain characters like e.g. `?` (%3F) or `#` (%23) as otherwise
parsing of the URI may fail with an error.

#### Authentication

Authentication is either by the username and password provided in the broker
URI as desribed above (RabbitMQ `PLAIN` method), or by mTLS providing a client
certificate and key using the `--tls-key-file`, `--tls-cert-file` options (RabbitMQ
`EXTERNAL` method). If both mTLS and a username and password is provided, then
rabtap will use mTLS and `PLAIN` authentication with the given username and
password.

### Environment variables

Use environment variables to specify standard values for broker and api endpoint.

#### Default RabbitMQ broker

In cases where the URI argument is optional, e.g. `rabtap tap [-uri
URI] exchange ...`, the URI of the RabbitMQ broker can be set with the
environment variable `RABTAP_AMQPURI`.  Example:

```console
$ export RABTAP_AMQPURI=amqp://guest:guest@localhost:5672/
$ rabtap tap amq.fanout:
...
```

#### Default RabbitMQ management API endpoint

The default RabbitMQ management API URI can be set using the `RABTAP_APIURI`
environment variable. Example:

```console
$ export RABTAP_APIURI=http://guest:guest@localhost:15672/api
$ rabtap info
...
```

#### Default RabbitMQ TLS config

The default TLS certificates path can be set using the
`RABTAP_TLS_CERTFILE` and `RABTAP_TLS_KEYFILE` and `RABTAP_TLS_CAFILE`
environments variables. All certificate and key files are expected in PEM
format. Example:

```console
$ export RABTAP_TLS_CERTFILE=/path/to/certs/user.crt
$ export RABTAP_TLS_KEYFILE=/path/to/certs/user.key
$ export RABTAP_TLS_CAFILE =/path/to/certs/ca.crt
$ echo "Hello" | rabtap pub --exchange amq.topic --routingkey "key"
...
```

#### Colored output

Output is colored, when writing to a terminal. This behaviour can be changed:

* set environment variable `NO_COLOR` to disable color output (or set
  `--no-color` option)
* set `--color` option to force colored output

### Command reference and examples

The following examples assume a RabbitMQ broker running on `localhost:5672` and
the management API available on port `localhost:15672`. Easiest way to start
such an instance is by running `docker run -ti --rm -p 5672:5672 -p 15672:15672
rabbitmq:3-management` or similar command to start a RabbitMQ container.

#### Broker info

The `info` command uses the REST API of RabbitMQ to gather and display
topolgy related information from the broker.

The `--mode MODE` option controls how the output is structured. Valid options
for `MODE` are `byExchange` (default) or `byConnection`.

The `--format=FORMAT` option controls the format of generated output. Valid
options are `text` for console text format (default) or `dot` to output the
tree structure in dot format for visualization with graphviz.

The features of an exchange are displayed in square brackets with `D`
(durable), `AD` (auto delete) and `I` (internal). The features of a queue are
displayed in square brackets with `D` (durable), `AD` (auto delete) and `EX`
(exclusive).

If the `--statistics` option is enabled, basic statistics are included
in the output.

The `--filter` option allows to filter output. See
[filtering](#filtering-output) section for details. Use the
`--by-connection` to sort output by connection (implies `--consumers`)

Examples (assume that `RABTAP_APIURI` environment variable is set):

* `rabtap info --consumers` - shows virtual hosts exchanges, queues and
  consumers of given broker in a tree view (see [screenshot](#screenshots)).
* `rabtap info --mode=byConnection` - shows virtual hosts, connections,
  consumers and queues of given broker in an tree view.
* `rabtap info --format=dot | dot -T svg > broker.svg` - renders broker info
  into `dot` format and uses graphviz to render a SVG file for final
  visualization.

#### Wire-tapping messages

The `tap` command allows to tap to exchanges and transparently receive the messages
sent to the exchanges.  The general form of the tap command is either

```text
rabtap tap EXCHANGES [--uri=URI] [--saveto=DIR] [--format=FORMAT]  [--limit=NUM]
       [--idle-timeout=DURATION] [--filter=EXPR] [-jkncsv]
       [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
```

or, to connect to multiple brokers simultanously,

```text
rabtap (tap --uri=URI EXCHANGES)... [--saveto=DIR] [--format=FORMAT]  [--limit=NUM]
       [--idle-timeout=DURATION] [--filter=EXPR] [-jkncsv]
       [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
```

The `EXCHANGES` argument specifies the exchanges and binding keys to use.
The `EXCHANGES` argument is of the form `EXCHANGE:[KEY][,EXCHANGE:[KEY]]*`. If
the exchange name contains a colon, use `\\:` to escape it, e.g.
`myexchange\\:with\\:colons:KEY`.

The acutal format of the binding key depends on the exchange type (e.g.
direct, topic, headers) and is described in the [RabbitMQ
documentation](https://www.rabbitmq.com/tutorials/amqp-concepts.html).

When `--saveto=DIR` is set, received messages will be written to the specified
directory. The `--formate=FORMAT` option controls the format of output both on
the console as well as in the written files (see
[below](#format-specification-for-tap-and-sub-command) for details).

The `--filter EXPR` allows filtering of messages using an expression language.
See [Filtering](#filtering-output) for details and examples.

Use the `--limit=NUM` option to limit the number of received messages. If
specified, rabtap will terminate, after `NUM` messages were read and passed
the filter (if set).

When `--idle-timeout=DURATION` is set, the subscribe command will terminate
when no new messages were received in the given time period. Look for the
description of the `--delay` option for the format of the `DURATION` parameter.

Examples for binding keys used in `tap` command:

* `#` on  an exchange of type `topic` will make the tap receive all messages
  on the exchange.
* a valid queue name for an exchange of type `direct` binds exactly to messages
  destined for this queue
* an empty binding key for exchanges of type `fanout` or type `headers` will
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

##### Tap all messages published or delivered (RabbitMQ FireHose)

The [RabbitMQ Firehose Tracer](https://www.rabbitmq.com/firehose.html) allows
to "see" every message that is published or delivered. To use it, the FireHose
tracer has to be enabled first:

```text
$ rabbitmqctl trace_on
```

Afterwards, every message published or delivered will be CC'd to the topic
exhange `amq.rabbitmq.trace`.  The messages can now be tapped with rabtap:

```text
$ rabtap --uri amqp://guest:guest@localhost:5672/ tap amq.rabbitmq.trace:published.#
```

RabbitMQ sends all messages published or delivered to the FireHose exchange.
Published messages are sent with the routing key `publish.{exchangename}`, while
delivered messages are sent with the routing key `deliver.{queuename}`.
Depending on what you want to record, specify your binding accordingly.

###### Replaying messages from the FireHose exchange

When messages are tapped or subscribed from the FireHose tracer exchange, these
messages have the original meta data stored in the headers section of the
message. When published later, rabtap detects that these message was recorded
from the FireHose (by examining the `exchange` attribute, which will be set to
`amq.rabbitmq.trace` by RabbitMQ in that case) and automatically transform the
message so that the originally published messages are replayed again.

##### Connect to multiple brokers

Rabtap allows you also to connect simultaneously to multiple brokers and
exchanges:

```text
$ rabtap tap --uri amqp://broker1 amq.topic:# tap --uri amqp://broker2 amq.fanout:
```

The example connects to `broker1` and taps to the `amq.topic` exchange and to
the `amq.fanout` exchange on `broker2`.

##### Message recorder

All tapped messages can be also be saved for later analysis or replay. Rabtap
supports saving of messages in two formats: raw body and metadata in separate
files or [JSON message format](#json-message-format) with embedded metadata and
message the body base64 encode. Examples:

* `$ rabtap tap amq.topic:# --saveto /tmp` - saves messages as pair of
  files consisting of raw message body and JSON meta data file to `/tmp`
  directory.
* `$ rabtap tap amq.topic:# --saveto /tmp --format json` - saves messages as
  JSON files to `/tmp` directory.

Files are created with file name `rabtap-`+`<Unix-Nano-Timestamp>`+ `.` +
`<extension>`.

#### Subscribe messages

The `sub` command reads messages from a queue or a stream. The general form
of the `sub` command is:

```text
rabtap sub QUEUE [--uri URI] [--saveto=DIR] [--format=FORMAT] [--limit=NUM]
       [--offset=OFFSET] [--args=KV]... [(--reject [--requeue])] [-jkcsvn]
       [--filter=EXPR] [--idle-timeout=DURATION]
       [(--tls-cert-file=CERTFILE --tls-key-file=KEYFILE)] [--tls-ca-file=CAFILE]
```

Use the `--reject` option to 'nack' messages, which in turn will be discarded
by the broker or routed to a configured dead letter exchange (DLX). if
`--requeue` is also set, the message will be returned to the queue.

The `--offset=OFFSET` option is used when subscribing to streams. Streams are
append-only data structures with non-destructive semantics and were introduced
with RabbitMQ 3.9. The `OFFSET` parameter specifies where to start reading from the
stream and must be any of: `first`, `last`, `next`, a numerical offset, a
RFC3339-Timestamp or a duration specification like `10m`. Consult the RabbitMQ
documentation for more information on [streams](https://www.rabbitmq.com/streams.html).

When `--idle-timeout=DURATION` is set, the subscribe command will terminate
when no new messages were received in the given time period. Look for the
description of the `--delay` option for the format of the `DURATION` parameter.

Refer to the `tap` command for a description of the `--filter=EXPR`,
`--limit=NUM`, `--saveto=DIR` and `--format=FORMAT`  options.

Examples:

* `rabtap sub somequeue --format=json` - will consume messages from queue
  `somequeue` and print out messages in JSON format. The Example assumes that
  `RABTAP_AMQPURI` environment variable is set, as the `--uri=AMQPURI`
  parameter is omitted
* `rabtap sub somequeue --limit=1 --reject --requeue` - consume one message
  from the queue `somequeue`, then exit, and let the broker requeue the message
* `rabtap sub mystream --offset=first` - read all messages from the stream
  `mystream`
* `rabtap sub mystream --offset=50` - read messages from stream `mystream`
  starting with the 50th message
* `rabtap sub mystream --offset=10m` - read messages from stream `mystream`
  which are aged 10 minutes or less
* `rabtap sub somequeue --idle-timeout=5s` - read messages from queue `somequeue`
  and exit when there is no new message received for 5 seconds

#### Publish messages

The `pub` command is used to publish messages to an exchange.  The messages to
be published are either read from a file, or from a directory which contains
previously recorded messages (e.g. using the `--saveto` option of the `tap`
command).  The general form of the `pub` command is:

```text
rabtap pub  [--uri=URI] [SOURCE] [--exchange=EXCHANGE] [--format=FORMAT]
       [--routingkey=KEY | (--header=HEADERKV)...] [--confirms] [--mandatory]
       [--delay=DELAY | --speed=FACTOR] [-jkv]
```

Message routing is either specified with a routing key and the `--routingkey`
option or, when header based routing should be used, by specifying the headers
with the `--header` option. Each header is specified in the form `KEY=VALUE`.
Multiple headers can be specified by specifying multiple `--header` options.

Messages can be published either in raw format, in which they are sent as-is,
or in [JSON-format, as described here](#json-message-format), which includes
message metadata and the body in a single JSON document. When multiple messages
are published with metadata, rabtap will calculate the time elapsed of
consecutive recorded messages using the metadata, and delay publishing
accordingly.

To set the publishing delay to a fix value, use the `--delay` option. To
publish without delays, use `--delay=0s`. To modify publishing speed use the
`--speed` option, which allows to set a factor to apply to the delays. A delay
is a sequence of decimal numbers, each with optional fraction and a unit
suffix, such as `300ms`, `-1.5h` or `2h45m`. Valid time units are `ns`, `us`
(or `µs`), `ms`, `s`, `m`, `h`.

When the `--confirms` option is set, rabtap waits for publisher confirmations
from the server and logs an error if a confirmation is negative or not received
(slows down throughput),

When the `--mandatory` option is set, rabtap publishes message in mandatory
mode. If set and a message can not be delivered to a queue, the server returns
the message and rabtap will log an error.

Examples:

* `$ echo hello | rabtap pub --exchange amq.fanout` - publish "hello" to
  exchange amqp.fanout
* `echo "hello" | rabtap pub --exchange amq.header --header KEY=VAL --header X=Y` -
  publish `hello` to exchange `amq.header` and set given message headers
* `$ rabtap pub messages.json --format=json`  - messages are read from file
  `messages.json` in [rabtap JSON format](#json-message-format). Target
  exchange and routing keys are read from the messages meta data.  The
  `messages.json` file can contain multiple JSON documents as it is treated as
  a JSON stream.  Rabtap will honor the `XRabtapReceived` timestamps of the
  messages and by default will delay the messages as they were recorded. This
  behaviour can be overridden by the `--delay` and `--speed` options
* `$ rabtap pub --exchange amq.direct -r myKey --format=json messages.json --delay=0s` - as
  before, but publish messages always to exchange `amq.direct` with routing key
  `myKey` and without any delays
* `$ rabtap pub --exchange amq.direct -r myKey --format=raw somedir --delay=0s` - as
  before, but assuming that `somedir` is a directory, the messages are read
  from message files previously recorded to this directory and replayed in the
  order they were recorded

#### Poor mans shovel

Rabtap instances can be linked through a pipe and messages will be read on
one side and published to the other. Note that for publish to work in streaming
mode, the JSON mode (`--format json`) must be used on both sides, so that
messages are encapsulated in JSON messages.

The example taps messages on `broker1` and publishes the messages to the
`amq.direct` exchange on `broker2`

```console
$ rabtap tap --uri amqp://broker1 my-topic-exchange:# --format json | \
  rabtap pub --uri amqp://broker2 --exchange amq.direct -r routingKey --format json
```

#### Close connection

The `conn` command allows to close a connection. The name of the connection to
be closed is expected as parameter. Use the `info` command with the
`--consumers` option to find the connection associated with a queue. Example:

```console
$ rabtap info --consumers
http://localhost:15672/api (broker ver='3.6.9', mgmt ver='3.6.9', cluster='rabbit@ae1ad1477419')
└── Vhost /
    ├── amq.direct (exchange, type 'direct', [D])
    :
    └── test-topic (exchange, type 'topic', [AD])
        ├── test-q-test-topic-0 (queue, key='test-q-test-topic-0', running, [])
        │   └── __rabtap-consumer-4823a3c0 (consumer user='guest', chan='172.17.0.1:59228 -> 172.17.0.2:5672 (1)')
        │       └── '172.17.0.1:59228 -> 172.17.0.2:5672' (connection client='https://github.com/streadway/amqp', host='172.17.0.2:5672', peer='172.17.0.1:59228')
        ├── test-q-test-topic-1 (queue, key='test-q-test-topic-1', running, [])
        :
$ rabtap conn close '172.17.0.1:59228 -> 172.17.0.2:5672'
```

#### Exchange commands

The `exchange` command is used to create, remove and bind exchanges:

```text
$ rabtap exchange create myexchange --type topic
$ rabtap exchange rm myexchange
```

The `create` commands allows to specify additional arguments to be passed to
RabbitMQ using the `--args=key=value` syntax:

```text
$ rabtap exchange create myexchange --type topic --args=alternate-exchange=myae
```

The `bind` command creates an exchange-to-exchange binding (similar to a
queue-to-exchange binding):

```text
$ rabtap exchange bind myechange to destexchange --bindingkey=KEY
```

#### Queue commands

The `queue` command is used to create, remove, bind or unbind queues:

```text
$ rabtap queue create myqueue
$ rabtap info --show-default
http://localhost:15672/api (broker ver='3.7.8', mgmt ver='3.7.8', cluster='rabbit@b2fe3b3b6826')
└── Vhost /
    ├── (default) (exchange, type 'direct', [D])
    │   └── myqueue (queue, key='myqueue', idle since 2018-12-07 20:46:15, [])
    :
    └── amq.topic (exchange, type 'topic', [D])
$ rabtap queue bind myqueue to amq.topic --bindingkey hello
$ rabtap info --show-default
http://localhost:15672/api (broker ver='3.7.8', mgmt ver='3.7.8', cluster='rabbit@b2fe3b3b6826')
└── Vhost /
    ├── (default) (exchange, type 'direct', [D])
    │   └── myqueue (queue, key='myqueue', idle since 2018-12-07 20:46:15, [])
    :
    └── amq.topic (exchange, type 'topic', [D])
        └── myqueue (queue, key='hello', idle since 2018-12-07 20:46:15, [])
$ rabtap queue unbind myqueue from amq.topic --bindingkey hello
$ rabtap info --show-default
http://localhost:15672/api (broker ver='3.7.8', mgmt ver='3.7.8', cluster='rabbit@b2fe3b3b6826')
└── Vhost /
    ├── (default) (exchange, type 'direct', [D])
    │   └── myqueue (queue, key='myqueue', idle since 2018-12-07 20:46:15, [])
    :
    └── amq.topic (exchange, type 'topic', [D])
$ rabtap queue purge myqueue
$ rabtap queue rm myqueue
$ rabtap info
http://localhost:15672/api (broker ver='3.7.8', mgmt ver='3.7.8', cluster='rabbit@b2fe3b3b6826')
└── Vhost /
    :
    └── amq.topic (exchange, type 'topic', [D])
```

The `create` commands allows to specify additional arguments to be passed to
RabbitMQ using the `--args=key=value` syntax. This allows for example to specify
the queue type or mode:

* `rabtap queue create quorum_queue --args=x-queue-type=quorum --durable` -
  create a quorum queue named `quorum_queue`. The same can be achieved by using
  the `--queue-type` option, which is an alias for setting the arg `x-queue-type`:
  `rabtap queue create quorum --queue-type=quorum --durable`
* `rabtap queue create mystream --queue-type=stream --durable` - create a stream
* `rabtap queue create lazy_queue --lazy` - create a classic queue in lazy
  mode that is named `lazy_queue`. `--lazy` is an alias for setting the arg
  `x-queue-mode`

### Format specification for tap and sub command

The `--format=FORMAT` option controls the format of the `tap` and `sub`
commands when writing messages to the console and optionally to the filesystem
(i.e.  when `--saveto=DIR` is set). The `FORMAT` parameter has the following
effect on the output:

| `FORMAT`        | Format on console                            | Format of saved messages (`--saveto DIR`)    |
|-----------------|----------------------------------------------|----------------------------------------------|
| `raw` (default) | Pretty-printed metadata + raw Message body   | Metadata as JSON-File + Body as-is           |
| `json`          | Pretty-printed JSON wiht base64 encoded body | Pretty-printed JSON with base64 encoded body |
| `json-nopp`     | Single line JSON wiht base64 encoded body    | Pretty-printed JSON with base64 encoded body |

Notes:

* the `--json` option is now deprecated. Use `--format=json` instead
* `nopp` stands for `no pretty-print`

### JSON message format

When using the `--format json` option, messages are print/read as a stream of JSON
messages in the following format:

```json
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
  "XRabtapReceivedTimestamp": "2019-06-13T19:33:51.920711583+02:00",
  "Body": "dGhpcyB0ZXN0IG1lc3NhZ2U .... IGFuZCBoZWFkZXJzIGFtcXAuVGFibGV7fQ=="
}
```

Note that in JSON mode, the `Body` is base64 encoded.

### Filtering output

When your brokers topology is complex, the output of the `info` command can
become very bloated. The `--filter` helps you to narrow output to the desired
information. The same filtering mechanism can be applied to the `tap` and `sub`
commands to filter only messages of interest.

#### Filtering expressions

A filtering expression is a function that evaluates to `true` or `false` (i.e.
a _predicate_). When a filter is used, output will be supressed, if the predicate
evalautes to `false`.

Rabtap uses [Expr](https://expr-lang.org/) to evaluate predicates. This allows
for complex expressions. See the [official expr-lang
documentation](https://expr-lang.org/docs/language-definition) for further
information.

> Note: prior to version 1.40, rabtap used
> [govaluate](https://github.com/Knetic/govaluate) to evaluate expressions.
> With the switch to [Expr](https://expr-lang.org/), the syntax has changed in
> some aspects (e.g. `=~` vs `matches`  in regular expression matches). Consult
> the documentation for details.

##### Evaluation context

During evaluation, the context (i.e. the current exchange, queue, etc.)
is made available to the filter expression as variables. In the `info`
command, the following context is set:

When using `rabtap --info --mode=byExchange` (which is the default), the
following variables are bound:

* the current exchange is bound to the variable [r.exchange](#exchange-type)
* the current queue is bound to the variable [r.queue](#queue-type)
* the current binding is bound to the variable [r.binding](#binding-type)

When using `rabtap --info --mode=byConnection` , the following variables are bound:

* the current connection is bound to the variable [r.connection](#connection-type)
* the current channel is bound to the variable [r.connection](#channel-type)

In the `sub` and `tap` commands, the following context is set:

* the current received message is bound to the variable [r.msg](#message-type),
  which allows access to the message-metadata and the body
* the current count of messages received that passed the filter is bound to
 `r.count`
* Helper functions are provided to access the message body:
  * the `r.toStr` function converts a byte buffer into a string, e.g. `let
  b=toJSON(r.toStr(r.msg.Body))`
  * the `r.gunzip` function decompresses the given byte buffer, e.g. `let
  b=toJSON(r.toStr(r.gunzip(r.msg.Body)))`, allowing to inspect a compressed body
  * the `r.body` function returns the message body, decompressing if necessary (i.e.
    if `ContentType` is `gzip`), e.g.
    `let b=toJSON(r.toStr(r.body(r.msg))`

##### Examples

The examples assume that the `RABTAP_APIURI` environment variable points to the
broker to be used, e.g.  `http://guest:guest@localhost:15672/api`).

* `rabtap info --filter "exchange.Name == 'amq.direct'" --omit-empty` - print
  only queues bound to exchange `amq.direct` and skip all empty exchanges.
* `rabtap info --filter "queue.Name matches '.*test.*'" --omit-empty` - print all
  queues with `test` in their name
* `rabtap info --filter "queue.Name matches '.*test.*' && exchange.Type == 'topic'" --omit-empty` - like
  before, but consider only exchanges of type `topic`.
* `rabtap info --filter "queue.Consumers > 0" --omit --stats --consumers` - print
  all queues with at least one consumer
* `rabtap info --mode=byConnection --filter="r.channel.PrefetchCount > 1` - list
  all connection with channel that have a prefetch-count > 1
* `rabtap info --mode=byConnection --filter="r.connection.PeerCertSubject matches '.*CN=guest.*'"` - 
  list all connection that were authenticated using mTLS and which certificates
  subject contains `CN=guest`
* `rabtap sub JDQ --filter="r.msg.RoutingKey == 'test'"` - print only messages that
  were sent with the routing key `test`
* `rabtap sub JDQ --filter="let b=fromJSON(r.toStr(r.gunzip(r.msg.Body))); b.Name == 'JAN'"` -
  print only messages that have `.Name == "JAN"` in their gzipped payload,
  interpreted as `JSON`

#### Type reference

The types reflect more or less the JSON API objects of the [REST API of
RabbitMQ](https://rawcdn.githack.com/rabbitmq/rabbitmq-management/v3.7.7/priv/www/api/index.html)
transformed to golang types.

##### Exchange type

<details>
  <summary>Definition of the Exchange type</summary>

```go
type Exchange struct {
    Name       string
    Vhost      string
    Type       string
    Durable    bool
    AutoDelete bool
    Internal   bool
    MessageStats struct {
        PublishOut
        PublishOutDetails struct {
            Rate float64
        }
        PublishIn        int
        PublishInDetails struct {
            Rate float64
        }
    }
}
```
</details>

##### Queue type

<details>
  <summary>Definition of the Queue type</summary>

```go
type Queue struct {
    MessagesDetails struct {
        Rate float64
    }
    Messages
    MessagesUnacknowledgedDetails struct {
        Rate float64
    }
    MessagesUnacknowledged int
    MessagesReadyDetails   struct {
        Rate float64
    }
    MessagesReady     int
    ReductionsDetails struct {
        Rate float64
    }
    Reductions int
    Node       string
    Exclusive            bool
    AutoDelete           bool
    Durable              bool
    Vhost                string
    Name                 string
    MessageBytesPagedOut int
    MessagesPagedOut     int
    BackingQueueStatus   struct {
        Mode string
        Q1   int
        Q2   int
        Q3  int
        Q4  int
        Len int
        NextSeqID         int
        AvgIngressRate    float64
        AvgEgressRate     float64
        AvgAckIngressRate float64
        AvgAckEgressRate  float64
    }
    MessageBytesPersistent     int
    MessageBytesRAM            int
    MessageBytesUnacknowledged int
    MessageBytesReady          int
    MessageBytes               int
    MessagesPersistent         int
    MessagesUnacknowledgedRAM  int
    MessagesReadyRAM           int
    MessagesRAM                int
    GarbageCollection          struct {
        MinorGcs        int
        FullsweepAfter  int
        MinHeapSize     int
        MinBinVheapSize int
        MaxHeapSize     int
    }
    State string
    Consumers int
    IdleSince string
    Memory    int
}
```
</details>

##### Binding type

<details>
  <summary>Definition of the Binding type</summary>

```go
type Binding struct {
    Source          string
    Vhost           string
    Destination     string
    DestinationType string
    RoutingKey      string
    PropertiesKey string
}
```
</details>

##### Connection type

<details>
  <summary>Definition of the Connection type</summary>

```go
type Connection struct {
    ReductionsDetails struct {
        Rate float64
    }
    Reductions     int
    RecvOctDetails struct {
        Rate float64
    }
    RecvOct        int
    SendOctDetails struct {
        Rate float64
    }
    SendOct          int
    ConnectedAt      int64
    ClientProperties struct {
        Product        string
        Version        string
        ConnectionName string
        Capabilities   struct {
            ConnectionBlocked    bool
            ConsumerCancelNotify bool
        }
    }
    ChannelMax        int
    FrameMax          int
    Timeout           int
    Vhost             string
    User              string
    Protocol          string
    SslHash           string
    SslCipher         string
    SslKeyExchange    string
    SslProtocol       string
    AuthMechanism     string
    PeerCertValidity  string
    PeerCertIssuer    string
    PeerCertSubject   string
    Ssl               bool
    PeerHost          string
    Host              string
    PeerPort          int
    Port              int
    Name              string
    Node              string
    Type              string
    GarbageCollection struct {
        MinorGcs        int
        FullsweepAfter  int
        MinHeapSize     int
        MinBinVheapSize int
        MaxHeapSize     int
    }
    Channels int
    State    string
    SendPend int
    SendCnt  int
    RecvCnt  int
}
```

</details>

##### Channel type

<details>
  <summary>Definition of the Channel type</summary>

```go
type Channel struct {
    ReductionsDetails struct {
        Rate float64
    }
    Reductions   int
    MessageStats struct {
        ReturnUnroutableDetails struct {
            Rate float64
        }
        ReturnUnroutable int
        ConfirmDetails   struct {
            Rate float64
        }
        Confirm        int
        PublishDetails struct {
            Rate float64
        }
        Publish    int
        Ack        int
        AckDetails struct {
            Rate float64
        }
        Deliver        int
        DeliverDetails struct {
            Rate float64
        }
        DeliverGet        int
        DeliverGetDetails struct {
            Rate float64
        }
        DeliverNoAck        int
        DeliverNoAckDetails struct {
            Rate float64
        }
        Get        int
        GetDetails struct {
            Rate float64
        }
        GetEmpty        int
        GetEmptyDetails struct {
            Rate float64
        }
        GetNoAck        int
        GetNoAckDetails struct {
            Rate float64
        }
        Redeliver        int
        RedeliverDetails struct {
            Rate float64
        }
    }
    Vhost             string
    User              string
    Number            int
    Name              string
    Node              string
    ConnectionDetails ConnectionDetails // see below
    GarbageCollection struct {
        MinorGcs        int
        FullsweepAfter  int
        MinHeapSize     int
        MinBinVheapSize int
        MaxHeapSize     int
    }
    State                  string
    GlobalPrefetchCount    int
    PrefetchCount          int
    AcksUncommitted        int
    MessagesUncommitted    int
    MessagesUnconfirmed    int
    MessagesUnacknowledged int
    ConsumerCount          int
    Confirm                bool
    Transactional          bool
    IdleSince              string
}

type ConnectionDetails struct {
    PeerHost string
    PeerPort int
    Name     string
}
```

</details>

##### Message type

<details>
  <summary>Definition of the Message type</summary>

The currently received messages in the `sub` and `pub` commands is exposed to
the filter predicate under the name `r.msg`. The `Message` type is more or
less the same as the `amqp.Delivery` type from the [golang amqp
package](https://github.com/rabbitmq/amqp091-go):

```go
type Message struct {
    Headers Table

    // Properties
    ContentType     string    // MIME content type
    ContentEncoding string    // MIME content encoding
    DeliveryMode    uint8     // queue implementation use - non-persistent (1) or persistent (2)
    Priority        uint8     // queue implementation use - 0 to 9
    CorrelationId   string    // application use - correlation identifier
    ReplyTo         string    // application use - address to reply to (ex: RPC)
    Expiration      string    // implementation use - message expiration spec
    MessageId       string    // application use - message identifier
    Timestamp       time.Time // application use - message timestamp
    Type            string    // application use - message type name
    UserId          string    // application use - creating user - should be authenticated user
    AppId           string    // application use - creating application id

    // Valid only with Channel.Consume
    ConsumerTag string

    // Valid only with Channel.Get
    MessageCount uint32

    DeliveryTag uint64
    Redelivered bool
    Exchange    string // basic.publish exchange
    RoutingKey  string // basic.publish routing key

    Body []byte
}
```

</details>

## Build from source

### Download and build using go install

```text
$ go install github.com/jandelgado/rabtap/cmd/rabtap@latest
```

### Build using Makefile and tests

To build rabtap from source, you need [go](https://golang.org/) (version >= 1.18)
and [golangci-lint](https://github.com/golangci/golangci-lint) installed.

```text
$ git clone https://github.com/jandelgado/rabtap && cd rabtap
$ make test  -or- make short-test
$ make
```

In order to run all tests (`make test`) an instance of RabbitMQ is expected to
run on localhost. Easiest way to start one is running `make run-broker`, which
will start a RabbitMQ docker container (i.e.  `docker run -ti --rm -p 5672:5672
-p 15672:15672 rabbitmq:3-management`). Another target, `short-tests`  runs
only unit-tests that down reach out to RabbitMQ.

### Experimental WASM/wasip1 port

Rabtap can be compiled for Web Assembly (WASM) and the new `wasip1` `GOOS` and
run on the console using, e.g.
[wasirun](https://github.com/stealthrocket/wasi-go). Example:

```console
$ go version
go version go1.21.3 linux/amd64
$ make wasm-build
CGO_ENABLED=1 GOOS=wasip1 GOARCH=wasm go build -o ./bin/rabtap-wasm ./cmd/rabtap
$ wasirun --version
wasirun v0.6.5
$ wasirun  bin/rabtap-wasm -- --api "http://guest:password@localhost:15672/api" info --no-color
http://localhost:15672/api (broker ver='3.12.6', mgmt ver='3.12.6', cluster='rabbit@3ea5bf2bac2f')
└─ Vhost /
   ├─ amq.direct (exchange(direct), [D])
   ├─ amq.fanout (exchange(fanout), [D])
   ├─ amq.headers (exchange(headers), [D])
   ├─ amq.match (exchange(headers), [D])
   ├─ amq.rabbitmq.trace (exchange(topic), [D|I])
   └─ amq.topic (exchange(topic), [D])
```

Another example using `wasirun` and `wasmedge` to publish and subscribe
to a queue:

```console
$ URI="amqp://guest:password@localhost/"
$ rabtap queue create test
$ rabtap queue bind test to amq.topic --bindingkey=key
$ echo "hello" | wasmedge  bin/rabtap-wasm  --uri "$URI" pub --exchange amq.topic --routingkey=key
$ wasirun  bin/rabtap-wasm -- --uri "$URI" sub test --limit=1
------ message received on 2023-10-29T11:48:56Z ------
exchange.......: amq.topic
routingkey.....: key
hello
```

See [my blog](https://jandelgado.github.io/blog/posts/rabtap-wasm) for details.

Limitations:

* environment variables like `RABTAP_AMQPURI` not supported, must specify
  all options on the command line
* depending on the runtime, colors must explicitly set or disabled using
  `--color` and `--no-color`, since terminal detection not working.

## Test data generator

A simple [test data generator tool](cmd/testgen/README.md) for manual tests is
included in the `cmd/testgen` directory.

## Contributing

* fork this repository
* create your feature branch
* add code
* add tests and make sure test coverage does not fall (`make test`)
* add [documentation](README.md)
* commit changes
* submit a PR

## Author

Jan Delgado (jdelgado at gmx dot net)

## Copyright and license

Copyright (c) 2017-2024 Jan Delgado.
rabtap is licensed under the GPLv3 license.
