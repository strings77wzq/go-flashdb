# goflashdb

[![Go Report Card](https://goreportcard.com/badge/github.com/strings77wzq/go-flashdb)](https://goreportcard.com/report/github.com/strings77wzq/go-flashdb)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

🚀 **goflashdb** - Go 语言实现的高性能、Redis 兼容的生产级键值存储系统

## ✨ 特性

- 🎯 **100% Redis 协议兼容** - 完全兼容 RESP 协议，可直接使用 redis-cli
- ⚡ **极致性能** - 65536 分片并发字典，充分发挥 Go 并发优势
- 🛡️ **生产级安全** - 认证、限流、危险命令拦截
- 💾 **持久化支持** - RDB 快照 + AOF 日志
- 🔌 **扩展接口** - OpenClaw/Skills/MCP 接口
- 📚 **完善文档** - 开发/学习/API 文档齐全

## 🚀 快速开始

```bash
# 编译
go build -o goflashdb ./cmd/goflashdb

# 启动
./goflashdb

# 测试
redis-cli ping
# PONG

redis-cli set hello world
# OK

redis-cli get hello
# "world"
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

## 🛠️ 支持的命令

| 类别 | 命令 |
|------|------|
| 字符串 | SET, GET, SETNX, SETEX, PSETEX, MSET, MGET, INCR, DECR, INCRBY, DECRBY, APPEND, STRLEN |
| 键 | DEL, EXISTS |
| 连接 | PING |
| 服务器 | BGSAVE |

## 🤝 贡献

欢迎贡献代码！

## 📄 许可证

[MIT License](LICENSE)

## 致谢

- [godis](https://github.com/HDT3213/godis) - Go 语言 Redis 实现参考
- [mini-redis](https://github.com/tokio-rs/mini-redis) - Rust 教学项目参考
