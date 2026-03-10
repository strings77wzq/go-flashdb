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

## [Unreleased]

### Planned
- Sorted Set (ZSET) commands
- Pub/Sub support
- Cluster mode
- Lua scripting
- More comprehensive tests
- Benchmark suite
