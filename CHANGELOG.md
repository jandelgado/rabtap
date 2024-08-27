# Changelog for rabtap

## v1.41 (2024-08-27)

* new: `--filter=FILTER` option for `tap` and `sub` commands to filter output
  of received messages, e.g. `rabtap sub JDQ --filter="r.msg.RoutingKey == 'test'"`
* breaking change: the `binding`, `queue`, `exchange`, `connection` and
  `channel` variables available in expressions of `rabtap info --filter=FILTER`
  are now all prefixed with `r.`  and are thus now to be referenced as `r.binding`,
  `r.queue` etc.

## v1.40 (2024-08-20)

* govaluate not being maintained since 2017, we switch to
  [Expr](https://expr-lang.org/) for use as the expression-evaluator of the
  `--filter <expr>`  option. The syntax of `Expr` is similar, but not the same,
  so this can be considered a breaking change
* dependency updates

## v1.39.3 (2024-06-22)

* simplify code (fanin) and reduce dependencies

## v1.39.2 (2024-06-08)

* fix consumer incorrectly displayed in info command
* update dependencies & go

## v1.39.1 (2023-10-29)

* experimental WASM support, use go 1.21, simplify code

## v1.38.2 (2023-02-24)

* update dependencies, use go 1.20

## v1.38 (2022-09-16)

* new: create exchange-to-exchange bindings with `rabtap exchange bind ...`
* new: show exchange-to-exchange bindings in `info` command
* fix: drastically improve performance of `info` command for large topologies
    with 1000's of queues/connections/channels
* chg: show channel information in `info` command with `--consumers` option
* chg: improve output of `info` command (attributes)
* chg: `dot` output of `info` command now shows separate vhosts

## v1.37 (2022-08-12)

* new: detect and replay messages subscribed from the RabbitMQ FireHose
       exchange (#74)

## v1.36 (2022-06-08)

* chg: use PLAIN auth when both a client certificate and username and
       password is provided (#73)
* switch to github.com/rabbitmq/amqp091-go amqp lib (#72)

## v1.35 (2021-11-27)

* new: new option `--idle-timeout=DURATION` added to `sub` and `tap` commands
* chg: always wait for server notifications after pubishing, to be informed
       on potential problems like publishing to a non existing exchange

## v1.34 (2021-11-24)

* new: specify multiple `--args=KEY=VALUE` options to pass additional arguments
       to the `sub` command.
* new: create lazy queue with `queue create ... --lazy`
* new: specify queue type to create with `queue create ... --queue-type=TYPE`
* new: specify offset with `--offset=OFFSET` when reading from streams

## v1.33 (2021-11-14)

* new: specify multiple `--args=KEY=VALUE` options to pass additional arguments
       to the `queue` and `exchange` commands.

## v1.32 (2021-11-13)

* new: `--limit=NUM` option in sub and tap command to limit the number
      of messages to receive.
* change: `--no-auto-ack` option in `sub` command was replaced by options
      `--reject` and `--requeue`. Prefetch count is now 1 and by default
      every message will be acknowledged by rabtap.

## v1.31 (2021-11-07)

* new: show queue utilisation and type (e.g. classic, quorum, stream) to info
  command
* new: build rabtap for more targets using goreleaser

## v1.30 (2021-10-13)

* new: support header based routing (pub, queue bind, queue unbind)

## v1.29 (2021-09-15)

* new: add a docker image (ghcr.io/jandelgado/rabtap)

## v1.28 (2021-08-26)

* fix: do not print messages to stdout in parallel, which can result in garbled
  output when the queue is filled up and messages are read at high frequency.
* new: listen for channel errors and print an error when a message is e.g.
  published to a non-existing exchange.
* new: `--confirms` option for `pub` command: wait for publisher confirmations
  from the server and log an error if a confirmation is negative or not
  received.
* new: `--mandatory` option for `pub` command: publish message with mandatory
  flag set. If set and a message can not be delivered to a queue, the server
  returns the message and rabtap will log error when the returned message
  is received (e.g. unroutable messages)
* new: listen for channel close events messages by the broker (e.g.  when
  publishing to an non-existant exchange)
* improved logging capabilities while reducing dependencies

## v1.27 (2021-03-28)

* new: `info` and `close` commands can now be cancelled by SIGTERM

## v1.26 (2021-03-26)

* fix: make client certificate auth work. This implements a workaround until
  https://github.com/streadway/amqp/pull/121 gets merged (#51)
* drop travis-ci, using github-actions now (#49)

## v1.25 (2020-10-30)

* fix: rabtap info: workaround for RabbitMQ API returning an `"undefined"`
  string where an integer was expected (#47)

## v1.24 (2020-09-28)

* new: support TLS client certificates (contributed by Francois Gouteroux)
* fix: make sure that headers in amqp.Publishing are always using amqp.Table
  structures, which could caused problems before.

## v1.23 (2020-04-09)

* fix: avoid endless recursion in info command (#42)

## v1.22 (2020-01-28)

* The `pub` command now allows ialso to replay messages from a direcory previously
  recorded. The pub command also honors the recorded timestamps and delays the
  messages during replay.  The signature of of the `pub` command was changed
  (see README.md). Note that the exchange is now optional and will be taken
  from the message metadata that is published.

## v1.21 (2019-12-14)

* new option: `--format FORMAT` which controls output format in `tap`,
 `subscribe` commands. `--format json` is equivalent to `--json`, which is
  now deprecated
* new output format: `--format json-nopp` which is not-pretty-printed JSON in
  `tap` and `subscribe` commands
* new option `--silent` for commands `tap` and `subscribe` which suppresses
  message output to stdout
* short `-o` option for the info command `--omit-empty` is no longer supported
* uniformly name test files `*_test.go` to improve external tool discoverbility

## v1.20 (2019-08-30)

* fix: avoid blocking write during tap, subscribe which can lead to ctrl+c
  to not work when e.g. ctrl+s is pressed during tap or subscribe.
* refactorings

## v1.19 (2019-08-18)

* new option `--no-auto-ack` for `sub` command disables auto acknowledge when
  messages are received by rabtap (#15)
* new: output of `info` command can now also be rendered to dot format, to
  create a visualization using graphviz. Set format with `--format=dot`, e.g.
  `rabtap info --format=dot`.
* fix: termination with ctrl+c in `tap`, `pub`, `sub` commands now works reliably
* change: option `--by-connection` of `info` command changed to `--mode=byConnection`
* heaviliy simplified code

## v1.18 (2019-07-05)

* fix: tap: allow colons in exchange names by escaping them (`exchange\\:with\\:colon`).
  Fixes #13.

## v1.17 (2019-06-13)

* Timestamp when message was received by rabtap now stored in JSON format
  in `XRabtapReceivedTimestamp` field.
* Simplified code

## v1.16 (2019-04-03)

* new option `--by-connection` for info command added, making `info` show
  connection >
* new: prefetch count added to output of `info` command (on consumer level)

## v1.15 (2019-03-01)

* new command `queue purge QUEUE` added

## v1.14 (2019-02-28)

* change: in subscribe mode, the consumer will use non-exclusive mode,
          allowing multiple consumers on the same queue.

## v1.13 (2019-02-26)

* updated go version to 1.12, dropping `dep` module manager
* fixed documentation

## v1.12 (2018-12-07)

### Added

* new command `queue unbind QUEUE from EXCHANGE` to unbind a queue from an
  exchange

### Fixes
* fix: when publishing (`rabtap pub` messages from stdin, a single EOF (ctrl+d)
  ends now the publishing process
* fix: `rabtap pub` fails with error message when publishing to unknown exchange
* fix: pub, sub, and tap now fail early when there is a connection problem on
  the initial connection to the broker

## v1.11 (2018-08-08)

### Changed

* fix: `--saveto=DIR` option had no effect in `sub` command

## v1.10 (2018-07-15)

### Added

* new options `--filter FILTER` to filter output of `info` command.

### Changed

* fix: bug in REST-client panicking when endpoint not available

## v1.9 (2018-05-15)

### Added

* info command accelerated by doing parallel REST requests to the RabbitMQ
  API endpoint

### Changed

* rabtap now terminates if the first connection attempt fails, instead
  of retrying to connect
* termination behaviour improved
* testgen tool adds a message count to generated messages

## v1.8 (2018-05-06)

### Added

* a changelog ;)
* new `--consumers` option of the `info` command prints also information on
  the connection.
* new command `conn` for connection related operations. Currently allows
  to close a connection with `rabtap conn close <connection-name>`.

### Changed

* minor changes to output of `info` command (i.e. some values are now quoted)



