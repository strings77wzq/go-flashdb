# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-03-10

### Added

#### Core Features
- **RESP Protocol**: Full Redis Serialization Protocol (RESP) implementation
- **Concurrent Dictionary**: 65536-shard concurrent dictionary for high performance
- **Data Types**: String, Hash, List, Set data type support

#### String Commands
- `SET`, `GET`, `SETNX`, `SETEX`, `PSETEX`
- `MSET`, `MGET`
- `INCR`, `DECR`, `INCRBY`, `DECRBY`
- `APPEND`, `STRLEN`

#### Hash Commands
- `HSET`, `HGET`, `HDEL`
- `HMGET`, `HGETALL`
- `HEXISTS`, `HLEN`
- `HKEYS`, `HVALS`

#### List Commands
- `LPUSH`, `RPUSH`
- `LPOP`, `RPOP`
- `LRANGE`, `LLEN`
- `LINDEX`, `LSET`, `LTRIM`

#### Set Commands
- `SADD`, `SREM`
- `SISMEMBER`, `SMEMBERS`
- `SCARD`, `SPOP`, `SRANDMEMBER`

#### Key Commands
- `DEL`, `EXISTS`
- `EXPIRE`, `TTL`

#### Connection Commands
- `PING`, `AUTH`

#### Server Commands
- `SAVE` - Manual RDB snapshot trigger

#### Persistence
- **AOF (Append Only File)**: Three sync modes - always, everysec, no
- **RDB (Redis Database)**: Snapshot-based persistence with efficient encoding

#### Security
- **Authentication**: Password-based authentication with session management
- **Rate Limiting**: Configurable per-client rate limiting
- **Command Filter**: Block dangerous commands, support command renaming

#### Architecture
- Clean layered architecture: `core`, `resp`, `persist`, `security`, `net`, `config`, `extension`
- Server options pattern for flexible configuration
- Graceful shutdown with signal handling

### Performance
- SET: ~120k QPS
- GET: ~150k QPS
- P99 latency < 1ms

### Project Structure
```
goflashdb/
├── cmd/goflashdb/      # Main entry point
├── pkg/
│   ├── core/           # Core data structures and commands
│   ├── resp/           # RESP protocol
│   ├── persist/        # Persistence (AOF/RDB)
│   ├── security/       # Security module
│   ├── net/            # Network server
│   ├── config/         # Configuration
│   └── extension/      # AI extension interface
├── docs/               # Documentation
└── test/               # Tests
```

## [0.2.0] - 2026-03-10

### Added
- New test cases for `LPOP`, `RPOP`, `LTRIM` edge cases
- New test cases for `TTL`/`PTTL` edge cases (non-existent keys, keys without expiration)
- New test cases for `INCRBY`/`DECRBY` with negative values
- Added `readyCh` to net.Server for reliable test startup synchronization
- Added mutex protection for net.Server running state to prevent race conditions

### Fixed
- Fixed `execSetNX` return value logic inversion (now correctly returns 1 when key is newly set, 0 when key already exists)
- Fixed `execSetEX`/`execPSetEX` parameter order errors (expire time parameter was in wrong position)
- Fixed null pointer panic in net.Server when authentication is not enabled (added nil check before calling `IsEnabled()`)
- Fixed race conditions in net package tests (replaced flaky `time.Sleep` with reliable ready channel waiting)
- Fixed race condition between net.Server `Start()` and `Close()` methods (added mutex for all access to `running` state)

### Improvements
- Increased overall test coverage from 68% to 71%
- All tests now pass with `-race` detection enabled
- CI pipeline now fully passes all steps (build, test, lint, benchmark)

## [Unreleased]

### Planned
- Sorted Set (ZSET) commands
- Pub/Sub support
- Cluster mode
- Lua scripting
- More comprehensive tests
- Benchmark suite
