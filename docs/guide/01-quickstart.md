# 快速开始

本文将带你快速上手 go-flashdb，体验一个 Go 语言实现的 Redis 服务器。

## 安装

### 方式一：源码编译（推荐）

```bash
# 1. 克隆项目
git clone https://github.com/strings77wzq/go-flashdb.git
cd go-flashdb

# 2. 编译（需要 Go 1.21+）
go build -o goflashdb ./cmd/goflashdb

# 3. 验证版本
./goflashdb --version
```

### 方式二：使用 go install

```bash
go install github.com/strings77wzq/go-flashdb/cmd/goflashdb@latest
```

## 启动服务器

```bash
# 默认配置启动
./goflashdb

# 输出：
# [INFO] 2026/01/08 10:30:00 goflashdb v0.3.0 starting...
# [INFO] 2026/01/08 10:30:00 Listening on :6379
```

服务器默认监听 6379 端口，与 Redis 默认端口一致。

## 基本使用

使用 redis-cli 或其他 Redis 客户端连接测试：

```bash
# 1. Ping 测试
redis-cli ping
# PONG

# 2. String 操作
redis-cli set hello world
# OK

redis-cli get hello
# "world"

# 3. Hash 操作
redis-cli hset user:1 name "Alice" age 30
# (integer) 2

redis-cli hgetall user:1
# 1) "name"
# 2) "Alice"
# 3) "age"
# 4) "30"

# 4. List 操作
redis-cli lpush mylist a b c
# (integer) 3

redis-cli lrange mylist 0 -1
# 1) "c"
# 2) "b"
# 3) "a"

# 5. Set 操作
redis-cli sadd myset x y z
# (integer) 3

redis-cli smembers myset
# 1) "x"
# 2) "y"
# 3) "z"

# 6. 事务操作
redis-cli multi
# OK
redis-cli set key1 value1
# QUEUED
redis-cli set key2 value2
# QUEUED
redis-cli exec
# 1) OK
# 2) OK

# 7. 服务器信息
redis-cli info server
# # Server
# goflashdb_version:0.3.0
# goflashdb_mode:standalone
# tcp_port:6379
# uptime:3600
```

## 配置

创建配置文件 `config.yaml`：

```yaml
# 网络配置
bind_addr: "0.0.0.0:6379"
max_conn: 10000

# 持久化配置
append_only: true
append_filename: "appendonly.aof"
append_fsync: "everysec"
rdb_filename: "dump.rdb"

# 安全配置
require_pass: "your-password"
max_clients: 10000

# 内存配置
max_memory: 0  # 0 表示不限制
```

启动时指定配置：

```bash
./goflashdb -config config.yaml
```

## 性能测试

使用内置的 benchmark 工具：

```bash
# 编译 benchmark
go build -o goflashdb-bench ./cmd/benchmark

# PING 测试
./goflashdb-bench -t ping -c 50 -n 10000

# SET 测试
./goflashdb-bench -t set -c 50 -n 10000 -d 10

# GET 测试
./goflashdb-bench -t get -c 50 -n 10000
```

预期输出：

```
Summary:
  Requests completed: 10000
  Requests failed: 0
  Total duration: 0.08 seconds
  QPS: 125000.00
  Latency min: 0.01 ms
  Latency max: 2.50 ms
  Latency avg: 0.40 ms
  Latency P50: 0.35 ms
  Latency P95: 0.80 ms
  Latency P99: 1.20 ms
```

## 与 Redis 对比

| 特性 | go-flashdb | Redis |
|------|-----------|-------|
| 协议 | RESP 2 | RESP 2/3 |
| 数据类型 | String/Hash/List/Set/ZSet/Bitmap/HLL | 全支持 |
| 持久化 | AOF + RDB | AOF + RDB |
| 事务 | WATCH/MULTI/EXEC | 全支持 |
| 复制 | 基础实现 | 全功能 |
| Cluster | 协议兼容 | 完整实现 |
| Lua | 支持 | 全支持 |

## 项目结构

```
goflashdb/
├── cmd/
│   ├── goflashdb/      # 主程序入口
│   └── benchmark/      # 性能测试工具
├── pkg/
│   ├── core/           # 核心：数据类型、命令、事务
│   ├── resp/           # 协议层：RESP 解析器
│   ├── persist/        # 持久化：AOF、RDB
│   ├── net/            # 网络层：TCP 服务器
│   ├── security/       # 安全：认证、限流、过滤
│   ├── replication/    # 复制：主从同步
│   ├── cluster/        # 集群：槽位管理
│   ├── benchmark/      # 基准测试框架
│   └── extension/      # 扩展：AI 接口预留
└── docs/               # 文档
```

## 下一步

- [TCP 服务器实现](/guide/02-tcp-server.html) - 了解 goroutine-per-connection 模型
- [RESP 协议解析](/guide/03-resp-protocol.html) - 深入协议设计
- [设计哲学](/design/) - 了解架构设计思路

## 常见问题

### Q: 与 Redis 的兼容性如何？

A: go-flashdb 实现了 RESP 2 协议，支持常用命令。可直接使用 redis-cli 或 Go-Redis 客户端连接。部分高级功能（如 Lua 脚本、Cluster 完整功能）正在开发中。

### Q: 性能如何？

A: 本地测试 SET/GET 可达 10-15万 QPS。生产环境建议使用 Redis。

### Q: 可以替代 Redis 吗？

A: 不建议。goflashdb 主要用于学习 Go 语言网络编程和 Redis 原理。生产环境请使用 Redis。

### Q: 如何贡献代码？

A: 欢迎提交 Issue 和 PR！详见 [GitHub](https://github.com/strings77wzq/go-flashdb)

## 参考

- [源码: cmd/goflashdb/main.go](https://github.com/strings77wzq/go-flashdb/blob/main/cmd/goflashdb/main.go)
- [Redis 协议规范](https://redis.io/topics/protocol)