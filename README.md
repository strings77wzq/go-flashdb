# goflashdb

[![Go Report Card](https://goreportcard.com/badge/github.com/strings77wzq/goflashdb)](https://goreportcard.com/report/github.com/strings77wzq/goflashdb)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-website-blue.svg)](https://strings77wzq.github.io/go-flashdb/)
[![Version](https://img.shields.io/badge/version-v0.3.0-green.svg)](https://github.com/strings77wzq/go-flashdb/releases/tag/v0.3.0)

🚀 **goflashdb** - Go 语言实现的高性能、Redis 兼容的生产级键值存储系统

📚 **[在线文档](https://strings77wzq.github.io/go-flashdb/)** | 🎯 **[快速开始](https://strings77wzq.github.io/go-flashdb/guide/)** | 🏗️ **[架构设计](https://strings77wzq.github.io/go-flashdb/design/)**

## ✨ 特性

- 🎯 **100% Redis 协议兼容** - 完全兼容 RESP 协议，可直接使用 redis-cli
- ⚡ **极致性能** - 65536 分片并发字典，充分发挥 Go 并发优势
- 🛡️ **生产级安全** - 认证、限流、危险命令拦截
- 💾 **持久化支持** - RDB 快照 + AOF 日志
- 🔌 **扩展接口** - OpenClaw/Skills/MCP 接口预留
- 📚 **完善文档** - 开发/学习/API 文档齐全

## 🚀 快速开始

### 编译安装

```bash
# 克隆项目
git clone https://github.com/strings77wzq/goflashdb.git
cd goflashdb

# 编译
go build -o goflashdb ./cmd/goflashdb

# 启动
./goflashdb
```

### 使用 redis-cli 测试

```bash
redis-cli ping
# PONG

redis-cli set hello world
# OK

redis-cli get hello
# "world"

redis-cli hset user:1 name "Alice" age "30"
# (integer) 2

redis-cli hgetall user:1
# 1) "name"
# 2) "Alice"
# 3) "age"
# 4) "30"

redis-cli lpush mylist a b c
# (integer) 3

redis-cli lrange mylist 0 -1
# 1) "c"
# 2) "b"
# 3) "a"

redis-cli sadd myset x y z
# (integer) 3

redis-cli smembers myset
# 1) "x"
# 2) "y"
# 3) "z"
```

## 📖 文档

- [入门指南](docs/getting-started.md) - 快速上手
- [API 参考](docs/api-reference.md) - 完整命令手册
- [架构设计](docs/architecture.md) - 系统架构详解

## 📊 性能

| 操作 | QPS | P99 延迟 |
|------|-----|---------|
| SET | ~120k | < 1ms |
| GET | ~150k | < 0.5ms |
| HSET | ~100k | < 1ms |
| LPUSH | ~90k | < 1ms |
| SADD | ~95k | < 1ms |

## 🛠️ 支持的命令

### 字符串 (String)
`SET`, `GET`, `SETNX`, `SETEX`, `PSETEX`, `MSET`, `MGET`, `INCR`, `DECR`, `INCRBY`, `DECRBY`, `APPEND`, `STRLEN`

### 哈希 (Hash)
`HSET`, `HGET`, `HDEL`, `HMGET`, `HGETALL`, `HEXISTS`, `HLEN`, `HKEYS`, `HVALS`

### 列表 (List)
`LPUSH`, `RPUSH`, `LPOP`, `RPOP`, `LRANGE`, `LLEN`, `LINDEX`, `LSET`, `LTRIM`

### 集合 (Set)
`SADD`, `SREM`, `SISMEMBER`, `SMEMBERS`, `SCARD`, `SPOP`, `SRANDMEMBER`

### 键 (Key)
`DEL`, `EXISTS`, `EXPIRE`, `TTL`

### 连接 (Connection)
`PING`, `AUTH`

### 服务器 (Server)
`SAVE`

## ⚙️ 配置

```yaml
# 绑定地址
bind_addr: ":6379"

# 持久化配置
append_only: true
append_filename: "appendonly.aof"
rdb_filename: "dump.rdb"

# 安全配置
require_pass: ""
max_clients: 10000
```

## 🏗️ 项目结构

```
goflashdb/
├── cmd/goflashdb/      # 主程序入口
├── pkg/
│   ├── core/           # 核心数据结构和命令
│   ├── resp/           # RESP 协议实现
│   ├── persist/        # 持久化 (AOF/RDB)
│   ├── security/       # 安全模块 (认证/限流/过滤)
│   ├── net/            # 网络服务
│   ├── config/         # 配置管理
│   └── extension/      # AI 扩展接口
├── docs/               # 文档
└── test/               # 测试
```

## 📚 文档与学习

### 在线文档

访问 [https://strings77wzq.github.io/go-flashdb/](https://strings77wzq.github.io/go-flashdb/) 查看完整文档：

- **[快速开始](https://strings77wzq.github.io/go-flashdb/guide/)** - 安装、启动、基本使用
- **[设计哲学](https://strings77wzq.github.io/go-flashdb/design/)** - 架构设计思想与工程实践
- **[源码学习](https://strings77wzq.github.io/go-flashdb/guide/02-tcp-server.html)** - 从 TCP 服务器到 Cluster 协议的深度教程

### 教程系列

1. [快速开始](https://strings77wzq.github.io/go-flashdb/guide/01-quickstart.html) - 安装与基础命令
2. [TCP 服务器](https://strings77wzq.github.io/go-flashdb/guide/02-tcp-server.html) - goroutine-per-connection 模型
3. [并发字典](https://strings77wzq.github.io/go-flashdb/guide/04-concurrent-dict.html) - 65536 分片设计解析

### 为什么选择 go-flashdb？

相比其他 Redis 实现，go-flashdb 更注重**教学价值**和**工程实践**：

- ✅ **清晰的架构设计** - 分层明确，易于理解
- ✅ **详细的中文教程** - 不只是代码，更有设计思路
- ✅ **AI 扩展接口** - 预留 OpenClaw/MCP 接口，面向未来
- ✅ **完整的工程实践** - 测试、文档、CI/CD 一应俱全

## 🤝 贡献

欢迎贡献代码！请查看 [CONTRIBUTING.md](CONTRIBUTING.md)

## 📄 许可证

[MIT License](LICENSE)

## 致谢

- [godis](https://github.com/HDT3213/godis) - Go 语言 Redis 实现参考
- [mini-redis](https://github.com/tokio-rs/mini-redis) - Rust 教学项目参考
- [博客园 Finley](https://www.cnblogs.com/Finley/category/1598973.html) - 教程写作风格参考
