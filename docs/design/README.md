# 设计理念

Go-FlashDB 的设计不是凭空而来，而是在**性能、可读性、可维护性**之间寻找最佳平衡点的结果。本章将深入讲解每个核心设计决策背后的思考。

## 为什么选择 Go 语言？

### 1. 原生并发支持

Go 的 goroutine + channel 模型是实现高并发服务器的理想选择：

```go
// 每个连接一个 goroutine
// 比线程更轻量（~2KB vs ~1MB）
// 比回调更易读
for {
    conn, err := listener.Accept()
    if err != nil {
        break
    }
    go handleConn(conn)  // 开启新 goroutine
}
```

### 2. 清晰的错误处理

虽然 `if err != nil` 被诟病，但它强制开发者思考每个错误场景：

```go
// 在 goflashdb 中，每个可能失败的操作都有明确的错误处理
reply := cmd.executor(db, args)
if reply.IsError() {
    // 处理错误
}
```

### 3. 单二进制部署

编译成单个二进制文件，无需运行时依赖，部署极其简单：

```bash
GOOS=linux go build  # 交叉编译到 Linux
```

## 架构设计哲学

### 分层架构：单一职责原则

```
┌─────────────────────────────────────┐
│  Extension Layer (扩展层)           │  ← AI 接口、插件系统
├─────────────────────────────────────┤
│  Security Layer (安全层)            │  ← 认证、限流、过滤
├─────────────────────────────────────┤
│  Network Layer (网络层)             │  ← TCP 服务器、TLS
├─────────────────────────────────────┤
│  Protocol Layer (协议层)            │  ← RESP 解析器
├─────────────────────────────────────┤
│  Core Layer (核心层)                │  ← 数据类型、命令、事务
├─────────────────────────────────────┤
│  Persistence Layer (持久化层)       │  ← AOF、RDB
└─────────────────────────────────────┘
```

**设计理由**：
- 每层只依赖下层，不依赖上层
- 可以单独测试每一层
- 易于替换某层实现（如把 TCP 换成 Unix Socket）

### 并发模型选择

#### 方案对比

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| sync.Map | 官方提供 | 写操作性能差 | 读多写少 |
| 分段锁 (Segment Lock) | 读写均衡 | 需要两次 hash | 通用场景 |
| 无锁数据结构 | 极致性能 | 实现复杂 | 专家级优化 |

#### go-flashdb 的选择：65536 分片

```go
const SegmentCount = 65536

type ConcurrentDict struct {
    segments []*Segment
}

type Segment struct {
    m  map[string]interface{}
    mu sync.RWMutex
}

func (dict *ConcurrentDict) Get(key string) (val interface{}, exists bool) {
    // 1. 计算 hash
    hashCode := fnv32(key)
    // 2. 定位分片
    segmentIndex := hashCode % SegmentCount
    segment := dict.segments[segmentIndex]
    // 3. 只锁住这个分片
    segment.mu.RLock()
    defer segment.mu.RUnlock()
    val, exists = segment.m[key]
    return
}
```

**为什么选 65536？**

1. **锁冲突概率低**：假设 1000 个并发连接，每个连接操作不同 key，冲突概率 < 0.001%
2. **内存占用合理**：每个分片一个 RWMutex (~24 bytes)，总共 ~1.5MB
3. **CPU 缓存友好**：分片数量是 2 的幂次，可用位运算代替取模

**对比 sync.Map**：

```go
// sync.Map 的问题：写操作需要复制整个 dirty map
// 当 map 很大时（如 100 万个 key），复制操作会阻塞所有协程

// go-flashdb 的优势：
// 写操作只锁住一个分片（~153 个 key），对其他分片无影响
```

### 协议解析：流式 vs 一次性

RESP 协议解析有两种思路：

#### 方案 A：流式解析（Streaming）

```go
// godis 使用的方式
func ParseStream(reader io.Reader) <-chan *Payload {
    ch := make(chan *Payload)
    go parse0(reader, ch)  // 后台 goroutine 持续解析
    return ch
}
```

优点：实时性好，适合高并发
缺点：需要额外的 goroutine，复杂度高

#### 方案 B：请求-响应解析（Request-Response）

```go
// go-flashdb 使用的方式
func (p *Parser) Parse() (Reply, error) {
    // 阻塞读取一个完整请求
    // 解析完成后立即返回
}
```

优点：简单直观，无额外 goroutine
缺点：解析大请求时会阻塞

**go-flashdb 的选择**：方案 B

理由：
1. 教学项目，简单比性能更重要
2. RESP 请求通常不大（< 1MB）
3. 一个连接一个 goroutine，解析阻塞不影响其他连接

### 持久化策略：AOF vs RDB

| 特性 | AOF | RDB |
|------|-----|-----|
| 恢复速度 | 慢（重放命令） | 快（直接加载） |
| 文件大小 | 大 | 小 |
| 数据安全 | 高（可配置每次 fsync） | 低（定期快照） |
| 实现复杂度 | 低 | 高 |

#### go-flashdb 的设计

**同时支持 AOF 和 RDB**：

```go
type PersistManager struct {
    aof *AOFWriter
    rdb *RDBWriter
}

// AOF：记录每个写命令
func (pm *PersistManager) AppendAOF(cmd []byte) {
    pm.aof.Write(cmd)
}

// RDB：定期全量快照
func (pm *PersistManager) SaveRDB(data map[string][]byte) {
    pm.rdb.Save(data)
}
```

**混合模式**（Redis 4.0+ 引入）：

```
启动时：加载 RDB（快速恢复大部分数据）
       ↓
      重放 AOF（恢复 RDB 之后的数据）
       ↓
     完成恢复
```

### 事务实现：乐观锁 vs 悲观锁

#### 方案对比

| 方案 | 实现复杂度 | 冲突处理 | 适用场景 |
|------|-----------|----------|----------|
| 悲观锁 | 高 | 阻塞等待 | 高冲突场景 |
| 乐观锁 (WATCH) | 低 | 失败重试 | 低冲突场景 |

#### go-flashdb 的实现：乐观锁

```go
// WATCH：监控 key
func (db *DB) Watch(keys []string) {
    for _, key := range keys {
        db.watchMap[key] = struct{}{}
    }
}

// EXEC：执行事务前检查
func (db *DB) ExecTransaction() Reply {
    // 检查被 WATCH 的 key 是否被修改
    for key := range db.watchMap {
        if db.isKeyModified(key) {
            return EmptyReply  // 事务取消
        }
    }
    // 执行事务中的命令
    ...
}
```

**优点**：
1. 实现简单，无需死锁检测
2. 无阻塞，性能更好
3. 与 Redis 行为一致

**缺点**：
1. 高冲突场景下失败率高
2. 需要客户端重试逻辑

### AI 扩展接口设计

这是 go-flashdb 区别于其他实现的**核心创新**。

#### 为什么需要 AI 扩展？

传统 Redis 通过命令行或客户端 API 交互。在 AI 时代，我们希望：
- AI 助手能直接操作数据库
- 自然语言查询（"查询最近 10 分钟登录的用户"）
- 智能数据迁移、优化建议

#### 接口设计

```go
// Extension 接口定义
type Extension interface {
    Init(config map[string]interface{}) error
    Execute(ctx context.Context, cmd string, args []string) (Reply, error)
    Close() error
}

// OpenClaw 扩展：AI 助手直接操作
func (e *OpenClawExtension) Execute(ctx context.Context, cmd string, args []string) (Reply, error) {
    // AI 助手通过自然语言描述操作
    // 内部转换为 Redis 命令
}

// MCP 扩展：模型上下文协议
func (e *MCPExtension) Execute(ctx context.Context, cmd string, args []string) (Reply, error) {
    // 实现 MCP 协议
    // 与大模型服务通信
}
```

**设计亮点**：
1. **插件化**：新增 AI 能力无需修改核心代码
2. **向后兼容**：传统客户端完全不受影响
3. **安全**：AI 操作同样经过权限验证

## 性能优化策略

### 1. 内存分配优化

```go
// 不好的做法：频繁分配小对象
func bad() []byte {
    return []byte("some string")  // 每次调用都分配
}

// go-flashdb 的做法：复用对象
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

func good() []byte {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf)
    // 使用 buf
    return buf
}
```

### 2. 减少锁粒度

```go
// 全局锁：性能差
type BadDB struct {
    mu   sync.RWMutex
    data map[string]interface{}
}

// 分片锁：性能好
type GoodDB struct {
    segments [65536]*Segment  // 每个分片独立锁
}
```

### 3. 零拷贝优化

RESP 协议解析时，尽量复用原始字节，避免复制：

```go
// 不好的做法：复制字节
func bad(data []byte) string {
    return string(data)  // 复制到新的字符串
}

// go-flashdb 的做法：使用 []byte
func good(data []byte) []byte {
    return data  // 复用原始切片
}
```

## 工程化实践

### 1. 错误处理策略

```go
// 分层错误处理
// 底层（pkg/core）：返回详细错误
func (db *DB) Get(key string) ([]byte, error) {
    if !db.Exists(key) {
        return nil, ErrKeyNotFound  // 具体错误类型
    }
    ...
}

// 协议层：转换为 Redis 错误回复
func handleGet(args [][]byte) resp.Reply {
    val, err := db.Get(string(args[0]))
    if err != nil {
        return resp.NewErrorReply("ERR " + err.Error())
    }
    return resp.NewBulkReply(val)
}
```

### 2. 测试策略

```go
// 单元测试：测试单个函数
func TestDictGet(t *testing.T) {
    dict := NewConcurrentDict()
    dict.Set("key", "value")
    val, ok := dict.Get("key")
    assert.True(t, ok)
    assert.Equal(t, "value", val)
}

// 集成测试：测试完整流程
func TestSetGet(t *testing.T) {
    db := NewDB(0)
    db.Exec("set", [][]byte{[]byte("key"), []byte("value")})
    reply := db.Exec("get", [][]byte{[]byte("key")})
    assert.Equal(t, "value", reply.String())
}

// 基准测试：测试性能
func BenchmarkSet(b *testing.B) {
    db := NewDB(0)
    for i := 0; i < b.N; i++ {
        db.Exec("set", [][]byte{[]byte("key"), []byte("value")})
    }
}
```

### 3. 文档策略

**代码即文档**：
- 清晰的变量命名
- 关键逻辑添加注释
- 复杂算法附上伪代码

**示例优于描述**：
- 每个功能都有使用示例
- 示例可直接运行
- 示例展示最佳实践

## 总结

Go-FlashDB 的设计遵循以下原则：

1. **简单优于复杂**：选择易于理解的算法和数据结构
2. **性能与可读性平衡**：清晰的代码 + 精心设计的算法
3. **渐进式复杂度**：从简单到复杂，循序渐进
4. **工程化实践**：错误处理、测试、文档一样不少
5. **面向未来**：预留 AI 扩展接口

下一章：[为什么选 65536 分片？](/design/concurrency-model.html)