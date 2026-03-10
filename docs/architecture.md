# GoSwiftKV 架构设计

## 分层架构

```
┌─────────────────────────────────────────────────────┐
│                 Application Layer                    │
│                  cmd/goswiftkv/                      │
└─────────────────────────────────────────────────────┘
                          │
┌─────────────────────────────────────────────────────┐
│                 Extension Layer                      │
│              pkg/extension/                          │
│         OpenClaw / Skills / MCP / Plugins           │
└─────────────────────────────────────────────────────┘
                          │
┌─────────────────────────────────────────────────────┐
│                 Security Layer                       │
│              pkg/security/                           │
│        Authenticator / RateLimiter / Filter         │
└─────────────────────────────────────────────────────┘
                          │
┌─────────────────────────────────────────────────────┐
│                 Network Layer                        │
│              pkg/net/                                │
│            TCP Server / Connection                   │
└─────────────────────────────────────────────────────┘
                          │
┌─────────────────────────────────────────────────────┐
│                 Protocol Layer                       │
│              pkg/resp/                               │
│          RESP Parser / Serializer                    │
└─────────────────────────────────────────────────────┘
                          │
┌─────────────────────────────────────────────────────┐
│                 Core Layer                           │
│              pkg/core/                               │
│     DB / ConcurrentDict / Commands / Types          │
└─────────────────────────────────────────────────────┘
                          │
┌─────────────────────────────────────────────────────┐
│                 Persistence Layer                    │
│              pkg/persist/                            │
│             AOF / RDB                                │
└─────────────────────────────────────────────────────┘
                          │
┌─────────────────────────────────────────────────────┐
│                 Infrastructure Layer                 │
│              pkg/lib/                                │
│          Logger / Utils / Config                     │
└─────────────────────────────────────────────────────┘
```

## 核心模块

### 1. 存储引擎 (pkg/core)

**ConcurrentDict**: 65536 分片并发字典

```go
type ConcurrentDict struct {
    segments []*Segment  // 65536 个独立分片
    count    int
}

type Segment struct {
    m  map[string]interface{}
    mu sync.RWMutex  // 每个分片独立锁
}
```

**优势**:
- 锁冲突概率 < 0.001%
- 读操作完全并发
- 写操作仅锁定相关分片

### 2. 协议层 (pkg/resp)

**RESP 协议支持**:
- Simple String: `+OK\r\n`
- Error: `-ERR message\r\n`
- Integer: `:1\r\n`
- Bulk String: `$6\r\nfoobar\r\n`
- Array: `*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n`

### 3. 持久化 (pkg/persist)

**AOF (Append Only File)**:
- 三种同步策略: Always / Everysec / No
- 异步写入，不阻塞主线程
- 支持 AOF Rewrite 压缩

**RDB (Redis Database)**:
- 二进制格式快照
- 支持过期时间
- 快速加载

### 4. 安全模块 (pkg/security)

**认证**: 密码 + Session 管理
**限流**: 滑动窗口算法
**命令过滤**: 危险命令拦截 + 重命名

### 5. AI 扩展 (pkg/extension)

**OpenClaw**: AI 助手直接操作接口
**Skills**: 自定义命令扩展
**MCP**: 模型上下文协议
**Plugins**: 插件系统

## 并发模型

```
                    ┌─────────────┐
                    │  Listener   │
                    └──────┬──────┘
                           │ Accept
              ┌────────────┼────────────┐
              │            │            │
        ┌─────▼─────┐ ┌────▼────┐ ┌────▼────┐
        │ Goroutine │ │Goroutine│ │Goroutine│
        │  Conn 1   │ │ Conn 2  │ │ Conn N  │
        └─────┬─────┘ └────┬────┘ └────┬────┘
              │            │            │
              └────────────┼────────────┘
                           │
                    ┌──────▼──────┐
                    │ Concurrent  │
                    │    Dict     │
                    │ (65536 shards)
                    └─────────────┘
```

## 性能优化

### 1. 锁优化
- 分片锁替代全局锁
- 读写分离 (RLock/RWMutex)
- 短临界区

### 2. 内存优化
- []byte 避免字符串转换
- sync.Pool 复用对象
- 预分配缓冲区

### 3. IO 优化
- bufio 缓冲读写
- 批量写入 AOF
- 异步 fsync

## 数据流

```
Client Request
      │
      ▼
┌─────────────┐
│ TCP Server  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ RESP Parser │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Router    │
└──────┬──────┘
       │
       ▼
┌─────────────┐     ┌─────────────┐
│   Command   │────▶│    AOF      │
│  Executor   │     └─────────────┘
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Concurrent  │
│    Dict     │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Response   │
└─────────────┘
```

## 扩展点

### 添加新命令

```go
func init() {
    RegisterCommand("mycommand", execMyCommand, prepareMyCommand, 2)
}

func execMyCommand(db *DB, args [][]byte) resp.Reply {
    // 实现逻辑
    return resp.OkReply
}

func prepareMyCommand(args [][]byte) (write, read []string) {
    return []string{string(args[0])}, nil
}
```

### 添加新数据类型

```go
type MyTypeData struct {
    data     interface{}
    expireAt int64
}

func (db *DB) GetMyTypeData(key string) (*MyTypeData, bool) {
    // 实现
}
```

### 添加插件

```go
type MyPlugin struct{}

func (p *MyPlugin) Init(config map[string]interface{}) error { return nil }
func (p *MyPlugin) Start() error { return nil }
func (p *MyPlugin) Stop() error { return nil }
func (p *MyPlugin) Info() PluginInfo {
    return PluginInfo{Name: "my-plugin", Version: "1.0"}
}
```
