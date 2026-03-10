# go-flashdb 开发指南

## 目录

1. [项目结构](#项目结构)
2. [架构设计](#架构设计)
3. [开发环境](#开发环境)
4. [添加新命令](#添加新命令)
5. [添加新数据类型](#添加新数据类型)
6. [测试指南](#测试指南)
7. [性能优化](#性能优化)

## 项目结构

```
go-flashdb/
├── cmd/goflashdb/        # 主程序入口
├── pkg/
│   ├── core/             # 核心数据库实现
│   │   ├── db.go         # 数据库核心
│   │   ├── string.go     # String 命令
│   │   ├── hash.go       # Hash 命令
│   │   ├── list.go       # List 命令
│   │   ├── set.go        # Set 命令
│   │   ├── zset.go       # ZSet 命令
│   │   ├── bitmap.go     # Bitmap 命令
│   │   ├── hyperloglog.go # HyperLogLog 命令
│   │   ├── pubsub.go     # Pub/Sub 实现
│   │   ├── lua.go        # Lua 脚本集成
│   │   └── replication.go # 复制集成
│   ├── net/              # 网络层
│   │   └── server.go     # TCP 服务器
│   ├── resp/             # RESP 协议
│   ├── persist/          # 持久化
│   │   ├── aof.go        # AOF 实现
│   │   ├── rdb.go        # RDB 实现
│   │   └── enhanced.go   # 增强持久化
│   ├── script/           # Lua 脚本引擎
│   ├── replication/      # 主从复制
│   ├── pool/             # 连接池
│   ├── security/         # 安全模块
│   └── config/           # 配置管理
├── docs/                 # 文档
│   ├── learning/         # 学习文档
│   ├── api/              # API 文档
│   └── development/      # 开发文档
└── openspec/             # 规范文档
```

## 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                      Client                              │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│                    net.Server                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │   Auth      │  │ RateLimiter │  │   Filter    │     │
│  └─────────────┘  └─────────────┘  └─────────────┘     │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│                     core.DB                              │
│  ┌─────────────────────────────────────────────────┐   │
│  │              ConcurrentDict (数据存储)           │   │
│  └─────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────┐   │
│  │              TTLDict (过期时间)                  │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
                            │
            ┌───────────────┼───────────────┐
            ▼               ▼               ▼
     ┌──────────┐    ┌──────────┐    ┌──────────┐
     │ persist  │    │ pubsub   │    │ script   │
     └──────────┘    └──────────┘    └──────────┘
```

### 核心模块

#### 1. 数据存储 (core.Dict)

使用分段锁实现的并发字典：

```go
type ConcurrentDict struct {
    shards []*shard
}

type shard struct {
    mu   sync.RWMutex
    data map[string]interface{}
}
```

#### 2. 命令注册

命令通过 init() 函数自动注册：

```go
func init() {
    RegisterCommand("mycommand", execMyCommand, prepareMyCommand, arity)
}
```

#### 3. RESP 协议

支持标准 RESP 协议：

- Simple String: `+OK\r\n`
- Error: `-ERR message\r\n`
- Integer: `:1\r\n`
- Bulk String: `$5\r\nhello\r\n`
- Array: `*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n`

## 开发环境

### 环境要求

- Go 1.21+
- Make (可选)

### 本地开发

```bash
# 克隆项目
git clone https://github.com/strings77wzq/go-flashdb.git
cd go-flashdb

# 安装依赖
go mod download

# 运行测试
go test ./... -race

# 运行服务器
go run cmd/goflashdb/main.go

# 代码检查
golangci-lint run
```

### 测试覆盖率

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 添加新命令

### 步骤 1: 创建执行函数

```go
func execMyCommand(db *DB, args [][]byte) resp.Reply {
    if len(args) < 1 {
        return resp.NewErrorReply("ERR wrong number of arguments")
    }
    
    // 实现命令逻辑
    key := string(args[0])
    value := db.data.Get(key)
    
    return &resp.BulkReply{Arg: value}
}
```

### 步骤 2: 创建预处理函数

```go
func prepareMyCommand(args [][]byte) ([]string, []string) {
    // 返回 (写键, 读键)
    if len(args) > 0 {
        return []string{string(args[0])}, nil
    }
    return nil, nil
}
```

### 步骤 3: 注册命令

```go
func init() {
    // arity: 正数表示精确参数数量，负数表示最小参数数量
    RegisterCommand("mycommand", execMyCommand, prepareMyCommand, 2)
}
```

### 步骤 4: 添加测试

```go
func TestMyCommand(t *testing.T) {
    db := newTestDB()
    
    reply := db.Exec("MYCOMMAND", [][]byte{[]byte("arg1")})
    
    bulkReply, ok := reply.(*resp.BulkReply)
    if !ok {
        t.Errorf("expected BulkReply, got %T", reply)
    }
    
    // 验证结果
}
```

## 添加新数据类型

### 步骤 1: 定义数据结构

```go
type MyType struct {
    field1 string
    field2 int
}
```

### 步骤 2: 添加到类型系统

```go
// 在 types.go 中添加类型常量
const (
    TypeString = iota
    TypeList
    TypeSet
    TypeZSet
    TypeHash
    TypeMyType  // 新类型
)
```

### 步骤 3: 实现命令

```go
func execMyTypeSet(db *DB, args [][]byte) resp.Reply {
    key := string(args[0])
    value := &MyType{...}
    db.data.Put(key, value)
    return resp.OkReply
}
```

## 测试指南

### 单元测试

```go
func TestMyFunction(t *testing.T) {
    // 准备
    db := newTestDB()
    
    // 执行
    reply := db.Exec("SET", [][]byte{[]byte("key"), []byte("value")})
    
    // 验证
    if !isOkReply(reply) {
        t.Errorf("expected OK reply")
    }
}
```

### 并发测试

```go
func TestConcurrent(t *testing.T) {
    db := newTestDB()
    var wg sync.WaitGroup
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            key := fmt.Sprintf("key%d", n)
            db.Exec("SET", [][]byte{[]byte(key), []byte("value")})
        }(i)
    }
    
    wg.Wait()
}
```

### 基准测试

```go
func BenchmarkSet(b *testing.B) {
    db := newTestDB()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        db.Exec("SET", [][]byte{[]byte("key"), []byte("value")})
    }
}
```

## 性能优化

### 1. 减少内存分配

```go
// 使用 sync.Pool 复用对象
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}
```

### 2. 批量操作

```go
// 使用 MSET 替代多次 SET
db.Exec("MSET", [][]byte{
    []byte("key1"), []byte("value1"),
    []byte("key2"), []byte("value2"),
})
```

### 3. 连接复用

使用连接池避免频繁创建连接。

### 4. 异步持久化

使用 BGSAVE 和 AOF 后台写入。

## 版本发布

### 版本号规则

遵循语义化版本：`MAJOR.MINOR.PATCH`

- MAJOR: 不兼容的 API 变更
- MINOR: 新功能，向后兼容
- PATCH: Bug 修复

### 发布流程

1. 更新 CHANGELOG.md
2. 创建 Git 标签
3. 推送到 GitHub
4. 构建 Release

```bash
git tag -a v0.3.0 -m "Release v0.3.0"
git push origin v0.3.0
```
