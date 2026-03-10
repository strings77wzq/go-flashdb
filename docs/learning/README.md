# go-flashdb 学习指南

## 目录

1. [快速开始](#快速开始)
2. [基本概念](#基本概念)
3. [数据类型](#数据类型)
4. [命令参考](#命令参考)
5. [高级特性](#高级特性)

## 快速开始

### 安装

```bash
go get github.com/strings77wzq/go-flashdb
```

### 启动服务器

```go
package main

import (
    "log"
    "goflashdb/pkg/net"
)

func main() {
    server, err := net.NewServer(":6379")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("Starting go-flashdb server on :6379")
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### 使用 redis-cli 连接

```bash
redis-cli -p 6379
```

## 基本概念

go-flashdb 是一个 Redis 兼容的键值存储引擎，支持：

- **内存存储**: 所有数据存储在内存中，提供极速访问
- **持久化**: 支持 AOF 和 RDB 两种持久化方式
- **主从复制**: 支持读写分离，提升系统可用性
- **Lua 脚本**: 支持服务器端脚本执行

## 数据类型

### String (字符串)

最基本的数据类型，可以存储字符串、整数或浮点数。

```redis
SET key value
GET key
INCR counter
```

### Hash (哈希)

用于存储字段-值对的集合。

```redis
HSET user:1 name "Alice"
HSET user:1 age 30
HGET user:1 name
HGETALL user:1
```

### List (列表)

有序的字符串列表。

```redis
LPUSH mylist "world"
LPUSH mylist "hello"
LRANGE mylist 0 -1
```

### Set (集合)

无序的唯一字符串集合。

```redis
SADD myset "member1"
SADD myset "member2"
SMEMBERS myset
```

### ZSet (有序集合)

带分数的有序集合。

```redis
ZADD leaderboard 100 "player1"
ZADD leaderboard 200 "player2"
ZRANGE leaderboard 0 -1 WITHSCORES
```

### Bitmap (位图)

用于位操作。

```redis
SETBIT mybitmap 0 1
GETBIT mybitmap 0
BITCOUNT mybitmap
```

### HyperLogLog

用于基数估算。

```redis
PFADD hll "element1" "element2"
PFCOUNT hll
```

## 命令参考

### 键操作

| 命令 | 说明 |
|------|------|
| SET key value | 设置键值 |
| GET key | 获取键值 |
| DEL key | 删除键 |
| EXISTS key | 检查键是否存在 |
| EXPIRE key seconds | 设置过期时间 |
| TTL key | 获取剩余过期时间 |

### 数值操作

| 命令 | 说明 |
|------|------|
| INCR key | 自增 1 |
| DECR key | 自减 1 |
| INCRBY key increment | 增加指定值 |

## 高级特性

### Pub/Sub 消息队列

发布订阅模式用于实时消息传递。

```redis
SUBSCRIBE channel1
PUBLISH channel1 "hello"
```

### Lua 脚本

在服务器端执行原子操作。

```redis
EVAL "return KEYS[1] .. ' ' .. ARGV[1]" 1 hello world
```

### 主从复制

配置从节点复制主节点数据。

```redis
REPLICAOF localhost 6380
ROLE
```

### 持久化

#### AOF (Append Only File)

记录所有写操作命令。

```go
server, _ := net.NewServer(":6379",
    net.WithPersist("appendonly.aof", "dump.rdb", true, time.Minute),
)
```

#### RDB (Redis Database)

定期保存数据快照。

```redis
SAVE
BGSAVE
```

## 性能优化

### 连接池

使用连接池管理客户端连接。

```go
config := pool.PoolConfig{
    MaxOpen:     100,
    MaxIdle:     10,
    MaxLifetime: 30 * time.Minute,
    MaxIdleTime: 5 * time.Minute,
}
p := pool.NewConnPool(config, connFunc)
```

### 批量操作

使用 MSET/MGET 进行批量操作。

```redis
MSET key1 value1 key2 value2
MGET key1 key2
```

## 下一步

- 阅读 [API 文档](../api/README.md)
- 查看 [开发指南](../development/README.md)
- 了解 [架构设计](../development/architecture.md)
