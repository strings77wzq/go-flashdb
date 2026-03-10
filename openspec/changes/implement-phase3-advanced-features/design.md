# go-flashdb v0.3.0 技术设计文档

## 目录

1. [Pub/Sub 消息队列设计](#1-pubsub-消息队列设计)
2. [Lua 脚本支持设计](#2-lua-脚本支持设计)
3. [主从复制设计](#3-主从复制设计)
4. [连接池优化设计](#4-连接池优化设计)
5. [持久化增强设计](#5-持久化增强设计)

---

## 1. Pub/Sub 消息队列设计

### 1.1 概述

Pub/Sub (发布/订阅) 是 Redis 的核心功能之一，允许消息生产者发布消息到频道，订阅者接收消息。本实现将支持以下命令：

- `SUBSCRIBE channel [channel ...]`
- `PSUBSCRIBE pattern [pattern ...]`
- `PUBLISH channel message`
- `PUBSUB subcommand [argument [argument ...]]`
- `UNSUBSCRIBE [channel [channel ...]]`
- `PUNSUBSCRIBE [pattern [pattern ...]]`

### 1.2 架构设计

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Publisher │────▶│   Server   │◀────│ Subscriber │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
                    ┌──────┴──────┐
                    │ PubSubManager │
                    │  - channels  │
                    │  - patterns  │
                    │  - messages  │
                    └─────────────┘
```

### 1.3 核心数据结构

```go
type PubSubManager struct {
    mu        sync.RWMutex
    channels  map[string]map[*Client]bool
    patterns  []PatternSubscription
    messages  *ring.Buffer
}

type PatternSubscription struct {
    pattern   string
    client    *Client
}

type Message struct {
    channel   string
    pattern   string
    payload   []byte
    timestamp time.Time
}
```

### 1.4 实现细节

1. **频道管理**
   - 使用 `map[string]map[*Client]bool` 存储频道和订阅者
   - 支持模式匹配订阅 (`PSUBSCRIBE`)
   - 消息通过 channel 广播

2. **订阅/退订**
   - `SUBSCRIBE`: 添加 client 到频道订阅列表
   - `UNSUBSCRIBE`: 从频道移除 client
   - 支持正则模式匹配 (`PSUBSCRIBE`)

3. **消息发布**
   - `PUBLISH`: 将消息写入所有订阅者的 message queue
   - 支持阻塞和非阻塞模式

4. **消息传递**
   - 订阅者的 connection handler 读取 message queue
   - 使用 RESP 数组格式发送消息

### 1.5 API 设计

```go
// pkg/pubsub/pubsub.go
type PubSubManager struct {
    channels map[string]map[*Client]bool
    patterns []PatternSubscription
}

func NewPubSubManager() *PubSubManager
func (p *PubSubManager) Subscribe(client *Client, channels []string) error
func (p *PubSubManager) Unsubscribe(client *Client, channels []string) error
func (p *PubSubManager) Publish(channel string, message []byte) int
func (p *PubSubManager) PSubscribe(client *Client, patterns []string) error
func (p *PubSubManager) Punubscribe(client *Client, patterns []string) error
```

---

## 2. Lua 脚本支持设计

### 2.1 概述

Lua 脚本允许在服务器端执行原子操作，增强 Redis 的编程能力。本实现将支持：

- `EVAL script numkeys [key [key ...]] [arg [arg ...]]`
- `EVALSHA sha1 numkeys [key [key ...]] [arg [arg ...]]`
- `SCRIPT LOAD script`
- `SCRIPT EXISTS sha1 [sha1 ...]`
- `SCRIPT FLUSH`
- `SCRIPT KILL`

### 2.2 架构设计

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Client    │────▶│  Eval Command│────▶│ Lua Engine  │
└──────────────┘     └──────────────┘     └──────────────┘
                           │                    │
                           ▼                    ▼
                    ┌──────────────┐     ┌──────────────┐
                    │ Script Cache │     │   Redis API  │
                    └──────────────┘     └──────────────┘
```

### 2.3 核心数据结构

```go
type LuaEngine struct {
    state      *lua.LState
    cache      *ScriptCache
    timeout    time.Duration
}

type ScriptCache struct {
    mu    sync.RWMutex
    cache map[string]*CompiledScript
}

type CompiledScript struct {
    source string
    func   *lua.LFunction
    sha1   string
}
```

### 2.4 安全设计

1. **沙箱限制**
   - 禁用文件系统访问 (`io`, `os`, `file`)
   - 禁用网络访问 (`socket`, `net`)
   - 限制执行时间 (默认 5ms)
   - 限制内存使用 (默认 4MB)

2. **API 白名单**
   - 只暴露安全的 Redis 命令
   - 禁止 `SHUTDOWN`, `FLUSHALL`, `CONFIG` 等危险命令

3. **超时控制**
   - 脚本执行超时自动终止
   - 死循环检测

### 2.5 API 设计

```go
// pkg/script/script.go
type LuaEngine struct {
 *ScriptCache
    timeout time.Duration
       cache   maxMem  int64
}

func NewLuaEngine() *LuaEngine
func (e *LuaEngine) Eval(script string, keys []string, args [][]byte) (resp.Reply, error)
func (e *LuaEngine) EvalSHA(sha1 string, keys []string, args [][]byte) (resp.Reply, error)
func (e *LuaEngine) LoadScript(source string) (string, error)
func (e *LuaEngine) Exists(sha1 string) bool
func (e *LuaEngine) Flush()
```

---

## 3. 主从复制设计

### 3.1 概述

主从复制 (Replication) 允许创建 Redis 的只读副本，实现读写分离和数据备份。本实现将支持：

- `REPLICAOF host port`
- `SLAVEOF host port`
- `ROLE`
- `WAIT numreplicas timeout`

### 3.2 架构设计

```
┌────────┐         ┌────────┐         ┌────────┐
│ Client │────────▶│ Master │────────▶│ Slave  │
│  (W)  │         │        │──sync──▶│ (R)   │
└────────┘         └────────┘         └────────┘
                          │
                     ┌────┴────┐
                     │ Replication
                     │  - PSYNC
                     │  - Commands
                     │  - Timeout
                     └──────────┘
```

### 3.3 复制协议

1. **全量同步 (Full Sync)**
   - Master 执行 `BGSAVE`
   - Master 发送 RDB 文件给 Slave
   - Master 发送缓冲区命令

2. **增量同步 (Partial Sync/PSYNC)**
   - 使用 replication ID 和 offset
   - 失败时尝试 PSYNC
   - 失败则回退到全量同步

### 3.4 核心数据结构

```go
type ReplicationManager struct {
    mu           sync.RWMutex
    role         Role
    master       *MasterInfo
    slave        *SlaveInfo
    replID       string
    replOffset   int64
    buffer       *CommandBuffer
}

type MasterInfo struct {
    connectedSlaves map[*SlaveConn]bool
    replOffset     int64
}

type SlaveInfo struct {
    masterHost    string
    masterPort    int
    masterConn    net.Conn
    replOffset    int64
    status        SlaveStatus
}
```

### 3.5 复制流程

1. **Slave 发起复制**
   - 发送 `PSYNC replicationID offset`
   - Master 返回 `+CONTINUE` 或 `+FULLRESYNC`

2. **命令传播**
   - Master 写命令同步到所有 Slave
   - 使用命令缓冲区 (Command Buffer)
   - 支持管道化

3. **心跳和超时**
   - Master 每 10s 发送 PING
   - Slave 超时 60s 判定连接断开

### 3.6 API 设计

```go
// pkg/replication/replication.go
type ReplicationManager struct {
    role       Role
    masterID   string
    offset     int64
}

type Master struct {
    slaves    map[string]*SlaveConn
    backlog   *CommandBacklog
}

func NewReplicationManager() *ReplicationManager
func (r *ReplicationManager) ReplicaOf(host string, port int) error
func (r *ReplicationManager) SlaveOf(host string, port int) error
func (r *ReplicationManager) AddSlave(conn net.Conn) error
func (r *ReplicationManager) GetRole() RoleInfo
```

---

## 4. 连接池优化设计

### 4.1 概述

连接池管理多客户端连接，提升并发性能。本实现将支持：

- 连接复用和回收
- 连接超时控制
- 最大连接数限制
- 连接健康检查

### 4.2 架构设计

```
┌────────────┐     ┌────────────┐     ┌────────────┐
│  Client 1 │     │  Client 2  │     │  Client N  │
└─────┬──────┘     └─────┬──────┘     └─────┬──────┘
      │                   │                   │
      ▼                   ▼                   ▼
┌─────────────────────────────────────────────┐
│              Connection Pool                │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐    │
│  │   Conn  │ │   Conn  │ │   Conn  │    │
│  └─────────┘ └─────────┘ └─────────┘    │
│                                             │
│  - Get()    - Put()    - Close()         │
│  - MaxOpen  - MaxIdle  - ConnMaxLifetime │
└─────────────────────────────────────────────┘
```

### 4.3 核心数据结构

```go
type ConnPool struct {
    mu         sync.Mutex
    freeConn   []*Conn
    openConns  int
    maxOpen    int
    maxIdle    int
    
    conns      map[*Conn]bool
    waitQueue  chan getRequest
    
    dial       func() (net.Conn, error)
    dialTimeout time.Duration
    maxLifetime time.Duration
}

type Conn struct {
    conn       net.Conn
    createdAt  time.Time
    usedAt     time.Time
    closed     bool
}
```

### 4.4 连接池配置

| 参数 | 默认值 | 说明 |
|------|--------|------|
| MaxOpen | 100 | 最大打开连接数 |
| MaxIdle | 10 | 最大空闲连接数 |
| MaxLifetime | 5min | 连接最大生命周期 |
| DialTimeout | 5s | 拨号超时 |
| IdleTimeout | 1min | 空闲超时 |

### 4.5 批量操作优化

```go
// Pipeline 支持
func (p *ConnPool) Pipeline(ctx context.Context, commands []Command) ([]Reply, error)

// Transaction 支持
func (p *ConnPool) ExecuteTx(ctx context.Context, commands []Command) ([]Reply, error)

// MGet/MSet 优化
func (p *ConnPool) MGet(ctx context.Context, keys []string) ([]string, error)
func (p *ConnPool) MSet(ctx context.Context, pairs map[string]string) error
```

### 4.6 API 设计

```go
// pkg/pool/pool.go
type ConnPool struct {
    maxOpen    int
    maxIdle    int
    maxLifetime time.Duration
    maxIdleTime time.Duration
}

func NewConnPool(opts ...PoolOption) *ConnPool
func (p *ConnPool) Get() (*Conn, error)
func (p *ConnPool) Put(conn *Conn)
func (p *ConnPool) Close() error
func (p *ConnPool) Stats() PoolStats
func (p *ConnPool) Pipeline(commands []Command) ([]Reply, error)
```

---

## 5. 持久化增强设计

### 5.1 概述

持久化是 Redis 数据安全的核心。本版本将优化：

- AOF 持久化性能
- 混合持久化模式
- 快照压缩
- 持久化监控

### 5.2 混合持久化模式

```
┌─────────────────────────────────────────────┐
│              Hybrid Persistence             │
│                                             │
│  ┌─────────┐      ┌─────────┐             │
│  │  RDB    │─────▶│  AOF    │             │
│  │ (Base)  │      │ (Delta) │             │
│  └─────────┘      └─────────┘             │
│       │                │                   │
│       └───────┬────────┘                   │
│               ▼                            │
│  ┌─────────────────────────┐             │
│  │   Loaded at startup    │             │
│  └─────────────────────────┘             │
└─────────────────────────────────────────────┘
```

### 5.3 AOF 优化

1. **AOF Rewrite**
   - 后台自动 AOF 重写
   - 增量重写 (避免全量重写)
   - 重写过程中命令合并

2. **fsync 策略**
   - `always`: 每次写同步 (最安全)
   - `everysec`: 每秒同步 (推荐)
   - `no`: 由系统决定 (最快)

3. **命令缓冲**
   - 使用环形缓冲区
   - 批量写入减少 I/O

### 5.4 核心数据结构

```go
type AOFPersister struct {
    file       *os.File
    mu         sync.Mutex
    bgMu       sync.Mutex
    rewriteCh  chan struct{}
    
    fsync      FsyncStrategy
    buf        *bufio.Writer
    appendOnly bool
    
    size       int64
    rewrites   int
}

type RDBSaver struct {
    file       *os.File
    buf        *bufio.Writer
    version    int
    checksum   bool
}
```

### 5.5 API 设计

```go
// pkg/persist/manager.go
type PersistManager interface {
    Append(cmd Command) error
    SaveRDB() error
    Load() error
    Rewrite() error
    Close() error
}

// 混合持久化
type HybridManager struct {
    rdb    *RDBLoader
    aof    *AOFPersister
    mode   PersistMode
}

func NewHybridManager(opts ...PersistOption) (*HybridManager, error)
func (m *HybridManager) Load() error
func (m *HybridManager) Append(cmd Command) error
func (m *HybridManager) RewriteAOF() error
```

---

## 6. 测试策略

### 6.1 单元测试

每个模块需要达到 80%+ 测试覆盖率：

- PubSub: 频道订阅、消息发布、模式匹配
- Lua: 脚本执行、沙箱限制、API 调用
- Replication: 主从同步、故障恢复
- Pool: 连接管理、并发安全
- Persist: AOF/RDB、混合模式

### 6.2 集成测试

- 主从复制端到端测试
- 连接池压力测试
- Lua 脚本功能测试
- Pub/Sub 消息延迟测试

---

## 7. 文档清单

| 文档 | 位置 | 说明 |
|------|------|------|
| Pub/Sub 教程 | docs/pubsub/README.md | 消息队列使用指南 |
| Lua 脚本指南 | docs/script/README.md | 脚本开发文档 |
| 复制指南 | docs/replication/README.md | 主从配置手册 |
| 连接池配置 | docs/pool/README.md | 连接池最佳实践 |
| 持久化配置 | docs/persist/README.md | 持久化选项说明 |
| API 参考 | docs/api/reference.md | 完整 API 文档 |
| 架构设计 | docs/arch/design.md | 技术架构详解 |
