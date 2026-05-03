# go-redis-clone

A minimal Redis-compatible server written in Go from scratch — no third-party dependencies. It speaks the [RESP2 protocol](https://redis.io/docs/latest/develop/reference/protocol-spec/) over TCP on port `6379`, so any Redis client (`redis-cli`, `go-redis`, `ioredis`, etc.) can connect to it.

This is a learning project: building Redis piece by piece to understand the wire protocol, the request/response loop, concurrent in-memory storage, and TTLs.

## Status

Implemented commands:

| Command | Notes |
|---------|-------|
| `PING [message]` | Replies `PONG` or echoes `message`. |
| `SET key value` | Stores a string value. No options yet (`EX`, `NX`, etc. are not supported). |
| `GET key` | Returns the value, or nil if missing or expired. |
| `DEL key` | Returns `1` if deleted, `0` if the key did not exist. Single-key only. |
| `EXPIRE key seconds` | Sets a TTL on an existing key (second resolution). |
| `COMMAND` | Stub reply so clients that probe on connect (e.g. `redis-cli`) work. |

Command names are case-insensitive. Unsupported commands return `-ERR unknown command 'X'`.

## Running the server

Requires Go `1.26+`.

```bash
make run        # go run ./cmd/redis-server/main.go
# or
make build      # builds ./bin/redis-server
make start      # runs ./bin/redis-server
```

Then in another terminal:

```bash
redis-cli
127.0.0.1:6379> SET hello world
OK
127.0.0.1:6379> GET hello
"world"
127.0.0.1:6379> EXPIRE hello 5
(integer) 1
127.0.0.1:6379> GET hello   # after 5+ seconds
(nil)
```

## Project layout

```
cmd/redis-server/         entrypoint — calls server.Start()
internal/
  core/                   in-memory key/value store with TTL
    database.go             GET / SET / DEL / EXPIRE — singleton, RWMutex-guarded
    entry.go                stored value + expiresAt
    store.go                store struct
  resp/                   RESP2 wire protocol
    message.go              Message struct + ToRaw() serializer
    message_type.go         RESP type tags (+, -, :, $, *)
    sender.go               write helpers (SendInteger, SendBulkString, ...)
    compare.go              case-insensitive command matching
    parser/                 parses inbound RESP frames; tolerates TCP fragmentation
  server/                 TCP server + command dispatcher
    tcp_server.go           Start() (port 6379) and Serve(listener) for tests
    handler.go              command routing on RESP arrays
    supported_commands.go   command name registry
```

Tests live next to the package they cover under a `test/` subdirectory (`internal/core/test`, `internal/resp/test`, `internal/server/test`).

## Development

```bash
make test           # go test ./...
make test-verbose   # go test -v ./...
make bench          # go test -bench=. -benchmem -run=^$ ./...
make dev            # hot-reload with air (go install github.com/air-verse/air@latest)
```

## Performance

Benchmarks live alongside the tests and cover three layers:

- `internal/core/test/database_bench_test.go` — direct calls to `SET`/`GET`/`DEL`/`EXPIRE`, with `*Parallel` variants exercising the `RWMutex`.
- `internal/resp/test/parser_bench_test.go` — RESP frame parsing (PING, SET, large bulk strings, pipelined batches, fragmented inputs). Reports MB/s via `b.SetBytes`.
- `internal/server/test/server_bench_test.go` — full TCP roundtrips against a server bound to `127.0.0.1:0`, including a pipelined batch and a concurrent-client variant.

### Environment
- **CPU**: Apple M4 (10 cores)
- **OS**: MacOS (Darwin arm64)
- **Go Version**: 1.26.2
- **Date**: 3 May, 2026

### Database

| Operation | exec | sec/op |  B/op | alloc/op |
| :--- | :--- | :--- | :--- | :--- |
|SetSameKey-10 |                 42714399 |                27.75 ns/op |           48 B/op |          2 allocs/op |
|SetUniqueKeys-10 |               4933201 |               274.7 ns/op |           178 B/op |          4 allocs/op |
|GetHit-10 |                     174766974 |                6.866 ns/op |           0 B/op |          0 allocs/op |
|GetMiss-10 |                    232038394 |                5.137 ns/op |           0 B/op |          0 allocs/op |
|GetWithTTL-10 |                 35092335 |                34.44 ns/op |            0 B/op |          0 allocs/op |
|Del-10 |                        27259174 |                47.64 ns/op |            0 B/op |          0 allocs/op |
|SetParallel-10 |                 3374941 |               479.8 ns/op |           236 B/op |          4 allocs/op |
|GetParallel-10 |                12876037 |                91.91 ns/op |            0 B/op |          0 allocs/op |
|MixedParallel-10 |               4801905 |               247.3 ns/op |            20 B/op |          1 allocs/op |

### RESP parser

| Operation | iterations | sec/op | performance |  B/op | alloc/op |
| :--- | :--- | :--- | :--- | :--- | :--- |
|Ping-10|          17133364 |                69.62 ns/op |      201.08 MB/s |         256 B/op |          4 allocs/op |
|Set-10|            9319818 |               128.0 ns/op |       273.53 MB/s |         496 B/op |          6 allocs/op |
|LargeBulkString-10|               30269110 |                38.97 ns/op |     105344.57 MB/s |       128 B/op |          2 allocs/op |
|Pipelined10-10|                     840963 |              1377 ns/op |         225.13 MB/s |        6192 B/op |         55 allocs/op |
|Fragmented-10|                    13797358 |                88.59 ns/op |      361.21 MB/s |         304 B/op |          3 allocs/op |

### Server (Integration)

| Operation | iterations | sec/op | performance |  B/op | alloc/op |
| :--- | :--- | :--- | :--- | :--- | :--- |
|Ping-10 |                 80653 |            12774 ns/op |           1.10 MB/s |         312 B/op |          9 allocs/op
|Set-10 |                  94320 |            12960 ns/op |           2.70 MB/s |         672 B/op |         17 allocs/op
|Get-10 |                  85202 |            13123 ns/op |           2.13 MB/s |         507 B/op |         11 allocs/op
|Pipelined10-10 |          54717 |            22209 ns/op |           6.30 MB/s |        4338 B/op |         76 allocs/op
|PingParallel-10 |        177940 |             6934 ns/op |           ? |  312 B/op |          9 allocs/op |


## Not implemented (yet)

`SET` options (`EX`, `PX`, `NX`, `XX`), `MGET`/`MSET`, `INCR`/`DECR`, `EXISTS`, `TTL`/`PTTL`, `PERSIST`, hashes, lists, sets, sorted sets, pub/sub, persistence, authentication, replication, clustering. Contributions welcome — the architecture is small enough to read in one sitting.
