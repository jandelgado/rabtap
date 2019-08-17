
# Changelog for rabtap

## v1.19 (2019-08-03)

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



