# go-flashdb 架构设计

## 整体架构

go-flashdb 采用分层架构设计：

```
┌─────────────────────────────────────────────────────────────┐
│                        客户端层                              │
│  redis-cli / redis-benchmark / 自定义客户端                  │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ TCP/RESP 协议
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        网络层 (net)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Server    │  │  Connection │  │   Handler   │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        安全层 (security)                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │    Auth     │  │ RateLimiter │  │   Filter    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        核心层 (core)                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                      DB                              │   │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐       │   │
│  │  │   Dict    │  │   TTL     │  │    TX     │       │   │
│  │  └───────────┘  └───────────┘  └───────────┘       │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                   命令处理器                          │   │
│  │  String │ Hash │ List │ Set │ ZSet │ Bitmap │ HLL  │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            ▼                 ▼                 ▼
┌───────────────────┐ ┌───────────────┐ ┌───────────────┐
│    持久化层        │ │   功能层      │ │   扩展层      │
│   (persist)       │ │ (pubsub,      │ │  (script,     │
│                   │ │ replication)  │ │   pool)       │
│  ┌─────────────┐  │ │               │ │               │
│  │     AOF     │  │ │               │ │               │
│  └─────────────┘  │ │               │ │               │
│  ┌─────────────┐  │ │               │ │               │
│  │     RDB     │  │ │               │ │               │
│  └─────────────┘  │ │               │ │               │
└───────────────────┘ └───────────────┘ └───────────────┘
```

## 核心组件

### 1. 数据存储 (ConcurrentDict)

采用分段锁设计，提高并发性能：

```go
type ConcurrentDict struct {
    shardCount int
    shards     []*shard
}

type shard struct {
    mu   sync.RWMutex
    data map[string]interface{}
}

func (d *ConcurrentDict) Get(key string) interface{} {
    shard := d.getShard(key)
    shard.mu.RLock()
    defer shard.mu.RUnlock()
    return shard.data[key]
}
```

**优势**：
- 减少锁竞争
- 提高并发读写性能
- 支持水平扩展

### 2. 命令注册系统

使用函数注册模式，支持动态扩展：

```go
type command struct {
    executor ExecFunc
    prepare  PreFunc
    arity    int
}

var cmdTable = make(map[string]*command)

func RegisterCommand(name string, executor ExecFunc, prepare PreFunc, arity int) {
    cmdTable[name] = &command{executor, prepare, arity}
}
```

### 3. RESP 协议解析

支持完整的 RESP 协议：

```
Simple String: +OK\r\n
Error:         -ERR message\r\n
Integer:       :1000\r\n
Bulk String:   $6\r\nfoobar\r\n
Array:         *2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n
```

### 4. 过期机制

使用惰性删除 + 定期清理：

```go
func (db *DB) checkExpire(key string) bool {
    if expireAt, ok := db.ttlDict.Get(key); ok {
        if time.Now().Unix() > expireAt.(int64) {
            db.data.Remove(key)
            db.ttlDict.Remove(key)
            return true
        }
    }
    return false
}
```

## 数据类型实现

### String

直接存储 []byte：

```go
type StringObject struct {
    value []byte
}
```

### Hash

使用 Go map 存储：

```go
type HashObject struct {
    fields map[string][]byte
}
```

### List

使用双向链表：

```go
type ListObject struct {
    elements *list.List
}
```

### Set

使用 map 实现集合：

```go
type SetObject struct {
    members map[string]struct{}
}
```

### ZSet

使用跳表实现有序集合：

```go
type ZSetObject struct {
    dict    map[string]float64
    skiplist *skiplist
}
```

## 高级特性

### Pub/Sub 实现

```go
type PubSubManager struct {
    mu       sync.RWMutex
    channels map[string]map[*Subscriber]bool
    patterns map[string]map[*Subscriber]bool
}

func (pm *PubSubManager) Publish(channel string, message []byte) int {
    pm.mu.RLock()
    defer pm.mu.RUnlock()
    
    count := 0
    if subs, ok := pm.channels[channel]; ok {
        for sub := range subs {
            sub.MsgCh <- &PubSubMessage{...}
            count++
        }
    }
    return count
}
```

### Lua 脚本沙箱

```go
func (e *LuaEngine) newSandboxedState() *lua.LState {
    L := lua.NewState(lua.Options{
        SkipOpenLibs: true,
    })
    
    // 禁用危险模块
    for _, name := range []string{"io", "os", "file", "net"} {
        L.SetGlobal(name, lua.LNil)
    }
    
    return L
}
```

### 主从复制

```
┌────────┐         ┌────────┐
│ Master │───────▶│ Slave  │
│        │  SYNC  │        │
└────────┘         └────────┘
    │
    │ Commands
    ▼
┌────────────────┐
│ Command Buffer │
│  (环形缓冲区)   │
└────────────────┘
```

### 持久化策略

#### AOF

```
┌─────────┐     ┌─────────┐     ┌─────────┐
│ Command │────▶│ AOF Buf │────▶│   Disk  │
└─────────┘     └─────────┘     └─────────┘
                     │
                     │ fsync
                     ▼
              ┌─────────────┐
              │ Always/     │
              │ Everysec/   │
              │ No          │
              └─────────────┘
```

#### RDB

```
┌─────────┐     ┌─────────┐     ┌─────────┐
│ Memory  │────▶│ Encode  │────▶│   RDB   │
└─────────┘     └─────────┘     └─────────┘
```

## 性能优化

### 1. 内存池

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}
```

### 2. 批量操作

```go
func execMSet(db *DB, args [][]byte) resp.Reply {
    for i := 0; i < len(args); i += 2 {
        db.data.Put(string(args[i]), args[i+1])
    }
    return resp.OkReply
}
```

### 3. 异步 I/O

```go
go func() {
    for cmd := range aofChan {
        writeCmd(cmd)
    }
}()
```

## 扩展性

### 添加新命令

1. 实现执行函数
2. 实现预处理函数
3. 注册命令
4. 添加测试

### 添加新数据类型

1. 定义数据结构
2. 实现类型检查
3. 实现相关命令
4. 更新持久化支持

## 安全考虑

### 认证

```go
if !authenticated {
    return resp.NewErrorReply("NOAUTH Authentication required")
}
```

### 限流

```go
if !rateLimiter.Allow(clientAddr) {
    return resp.NewErrorReply("ERR rate limit exceeded")
}
```

### 命令过滤

```go
if filter.IsBlocked(cmdName) {
    return resp.NewErrorReply("ERR command is blocked")
}
```
