# go-flashdb API 文档

## 目录

1. [服务器 API](#服务器-api)
2. [核心数据库 API](#核心数据库-api)
3. [Pub/Sub API](#pubsub-api)
4. [Lua 脚本 API](#lua-脚本-api)
5. [复制 API](#复制-api)
6. [连接池 API](#连接池-api)
7. [持久化 API](#持久化-api)

## 服务器 API

### net.Server

服务器实例，处理客户端连接和命令执行。

```go
import "goflashdb/pkg/net"

server, err := net.NewServer(":6379",
    net.WithAuth("password"),
    net.WithRateLimit(1000, time.Minute),
    net.WithFilter(renamedCommands),
    net.WithPersist("appendonly.aof", "dump.rdb", true, time.Minute),
    net.WithTLS("cert.pem", "key.pem"),
)
```

#### 配置选项

| 选项 | 类型 | 说明 |
|------|------|------|
| WithAuth | string | 设置密码认证 |
| WithRateLimit | (int, time.Duration) | 限流配置 |
| WithFilter | map[string]string | 命令重命名/禁用 |
| WithPersist | (string, string, bool, time.Duration) | 持久化配置 |
| WithTLS | (string, string) | TLS 证书配置 |

#### 方法

```go
func (s *Server) Start() error    // 启动服务器
func (s *Server) Close()           // 关闭服务器
func (s *Server) GetDB() *core.DB  // 获取数据库实例
```

## 核心数据库 API

### core.DB

数据库实例，提供数据存储和命令执行。

```go
import "goflashdb/pkg/core"

db := core.NewDB(0)  // 创建数据库，参数为数据库索引
```

#### 方法

```go
func (db *DB) Exec(cmd string, args [][]byte) resp.Reply  // 执行命令
func (db *DB) LoadFromPersist() error                      // 从持久化加载数据
```

#### 命令执行示例

```go
reply := db.Exec("SET", [][]byte{[]byte("key"), []byte("value")})

switch r := reply.(type) {
case *resp.BulkReply:
    fmt.Println(string(r.Arg))
case *resp.IntegerReply:
    fmt.Println(r.Num)
case *resp.ErrorReply:
    fmt.Println("Error:", r.Error())
}
```

## Pub/Sub API

### core.PubSubManager

发布订阅管理器。

```go
import "goflashdb/pkg/core"

pm := core.NewPubSubManager()
```

#### 方法

```go
func (pm *PubSubManager) Subscribe(sub *Subscriber, channels []string) []resp.Reply
func (pm *PubSubManager) Unsubscribe(sub *Subscriber, channels []string) []resp.Reply
func (pm *PubSubManager) PSubscribe(sub *Subscriber, patterns []string) []resp.Reply
func (pm *PubSubManager) PUnsubscribe(sub *Subscriber, patterns []string) []resp.Reply
func (pm *PubSubManager) Publish(channel string, message []byte) int
func (pm *PubSubManager) Channels(pattern string) []string
func (pm *PubSubManager) NumSub(channels []string) map[string]int
func (pm *PubSubManager) NumPat() int
func (pm *PubSubManager) RemoveSubscriber(sub *Subscriber)
```

#### 订阅者

```go
sub := core.NewSubscriber(clientID)

go func() {
    for msg := range sub.MsgCh {
        fmt.Printf("Received on %s: %s\n", msg.Channel, string(msg.Payload))
    }
}()

pm.Subscribe(sub, []string{"channel1", "channel2"})
```

## Lua 脚本 API

### script.LuaEngine

Lua 脚本引擎。

```go
import "goflashdb/pkg/script"

engine := script.NewLuaEngine()
```

#### 方法

```go
func (e *LuaEngine) Eval(source string, keys []string, args [][]byte) (resp.Reply, error)
func (e *LuaEngine) EvalSHA(sha1 string, keys []string, args [][]byte) (resp.Reply, error)
func (e *LuaEngine) LoadScript(source string) (string, error)
func (e *LuaEngine) Exists(sha1Hashes []string) []int
func (e *LuaEngine) Flush()
```

#### 脚本执行示例

```go
script := `return KEYS[1] .. ' ' .. ARGV[1]`
reply, err := engine.Eval(script, []string{"hello"}, [][]byte{[]byte("world")})
// reply = "hello world"
```

#### 全局变量

脚本中可用的全局变量：

| 变量 | 类型 | 说明 |
|------|------|------|
| KEYS | table | 键名数组 |
| ARGV | table | 参数数组 |

## 复制 API

### replication.ReplicationManager

复制管理器。

```go
import "goflashdb/pkg/replication"

rm := replication.NewReplicationManager()
```

#### 方法

```go
func (rm *ReplicationManager) Role() Role
func (rm *ReplicationManager) ReplicaOf(host string, port int) error
func (rm *ReplicationManager) GetReplID() string
func (rm *ReplicationManager) GetReplOffset() int64
func (rm *ReplicationManager) AddCommand(command []byte)
func (rm *ReplicationManager) GetSlaveCount() int
func (rm *ReplicationManager) GetInfo() map[string]interface{}
```

#### 角色管理

```go
// 设置为从节点
rm.ReplicaOf("master-host", 6379)

// 设置为主节点
rm.ReplicaOf("no", 0)

// 获取角色信息
role := rm.Role()  // RoleMaster 或 RoleSlave
```

## 连接池 API

### pool.ConnPool

连接池管理器。

```go
import "goflashdb/pkg/pool"

config := pool.PoolConfig{
    MaxOpen:     100,              // 最大连接数
    MaxIdle:     10,               // 最大空闲连接数
    MaxLifetime: 30 * time.Minute, // 连接最大生命周期
    MaxIdleTime: 5 * time.Minute,  // 空闲连接超时时间
}

connFunc := func() (pool.Conn, error) {
    return &MyConn{}, nil
}

pool := pool.NewConnPool(config, connFunc)
```

#### 方法

```go
func (p *ConnPool) Get() (Conn, error)     // 获取连接
func (p *ConnPool) Put(conn Conn) error    // 归还连接
func (p *ConnPool) Close() error           // 关闭连接池
func (p *ConnPool) Stats() PoolStats       // 获取统计信息
func (p *ConnPool) PruneIdle() int         // 清理空闲连接
```

#### 连接接口

```go
type Conn interface {
    Close() error
    IsClosed() bool
    LastUsed() time.Time
    SetLastUsed(t time.Time)
}
```

## 持久化 API

### persist.HybridPersistence

混合持久化管理器。

```go
import "goflashdb/pkg/persist"

config := persist.HybridConfig{
    AOFile:     "appendonly.aof",
    RDBFile:    "dump.rdb",
    AOFMode:    persist.AOFEverysec,
    AOFEnabled: true,
    RDBEnabled: true,
}

hp, err := persist.NewHybridPersistence(config)
```

#### 方法

```go
func (h *HybridPersistence) AppendAOF(cmd []byte)
func (h *HybridPersistence) Save(data map[string][]byte, expireTimes map[string]int64) error
func (h *HybridPersistence) BGSAVE(data map[string][]byte, expireTimes map[string]int64) error
func (h *HybridPersistence) BGREWRITE(data map[string][]byte) error
func (h *HybridPersistence) Load() (map[string][]byte, map[string]int64, error)
func (h *HybridPersistence) Close()
func (h *HybridPersistence) Info() map[string]interface{}
```

#### AOF 模式

| 模式 | 说明 |
|------|------|
| AOFAlways | 每次写入都 fsync |
| AOFEverysec | 每秒 fsync |
| AOFNo | 不主动 fsync |

## 响应类型

### resp.Reply

所有命令执行返回 resp.Reply 接口。

```go
type Reply interface {
    ToBytes() []byte
}

// 具体类型
type BulkReply struct { Arg []byte }
type IntegerReply struct { Num int64 }
type ArrayReply struct { Replies []Reply }
type ErrorReply struct { Error string }
```

## 错误处理

```go
reply := db.Exec("GET", [][]byte{[]byte("key")})

if errReply, ok := reply.(*resp.ErrorReply); ok {
    log.Printf("Command error: %s", errReply.Error())
    return
}
```
