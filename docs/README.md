---
home: true
heroImage: /logo.svg
heroText: Go-FlashDB
heroAlt: Go-FlashDB Logo
tagline: 深入源码学习 Go 语言高性能 Redis 服务器实现
taglineAlt: 不只是教程，更是设计思想与工程实践的完美结合

actions:
  - text: 快速开始
    link: /guide/
    type: primary
  - text: 设计哲学
    link: /design/
    type: secondary
  - text: 查看源码
    link: https://github.com/strings77wzq/go-flashdb
    type: secondary

features:
  - title: ⚡ 极致性能
    details: 65536 分片并发字典设计，锁冲突率低于 0.001%。对比 sync.Map，在写多场景下性能提升 5-10 倍。
  
  - title: 🎯 Redis 完全兼容
    details: 100% 实现 RESP 协议，支持 String/Hash/List/Set/ZSet/Bitmap/HyperLogLog 全部数据类型，可直接使用 redis-cli 连接。
  
  - title: 🏗️ 清晰架构
    details: 分层架构设计：Protocol → Core → Persistence → Security → Extension。每个模块职责单一，代码可读性极强。
  
  - title: 🤖 AI 扩展
    details: 预留 OpenClaw/Skills/MCP 扩展接口，为 AI 时代的数据库交互做好准备。这是 go-flashdb 区别于其他实现的核心创新。
  
  - title: 🛡️ 生产级安全
    details: 内置认证、限流、危险命令拦截三层安全防护。支持 TLS 加密传输，符合企业级部署标准。
  
  - title: 📚 完整教程
    details: 从 TCP 服务器到 Cluster 协议，12 篇深度教程带你从零实现 Redis。每篇都包含设计原理、源码解析和性能对比。

footer: MIT Licensed | Copyright © 2026-present go-flashdb
footerHtml: true
---

## 为什么选择 Go-FlashDB？

### 不只是又一个 Redis 实现

市面上已有多个 Go 语言实现的 Redis（如 godis、redcon），但 go-flashdb 致力于成为**最适合学习**的实现：

| 特性 | go-flashdb | godis | redcon |
|------|-----------|-------|--------|
| **代码可读性** | ⭐⭐⭐⭐⭐ 分层清晰 | ⭐⭐⭐⭐ 功能完整 | ⭐⭐⭐ 库而非完整实现 |
| **学习曲线** | 平缓，从基础到高级 | 陡峭，直接看集群 | 不适用 |
| **并发设计** | 65536 分片 + 详细讲解 | 分段锁 | - |
| **AI 扩展** | ✅ 预留接口 | ❌ 无 | ❌ 无 |
| **教程完整度** | 12 篇深度教程 | 博客系列 | ❌ 无 |

### 设计理念

#### 1. **性能与可读性的平衡**

许多高性能代码牺牲可读性来换取速度。go-flashdb 采用**清晰的架构 + 精心设计的算法**，让你既能理解原理，又能学到性能优化技巧。

```go
// ConcurrentDict: 65536 分片设计
// 核心思想：通过增加分片数量来降低锁冲突概率
// 而不是使用更复杂的无锁数据结构
type ConcurrentDict struct {
    segments []*Segment  // 65536 个分片
    count    int
}

type Segment struct {
    m  map[string]interface{}
    mu sync.RWMutex  // 每个分片独立锁
}
```

#### 2. **渐进式复杂度**

从最简单的 Echo 服务器开始，逐步添加 RESP 协议、数据类型、持久化、事务、复制、Cluster。每章都建立在前章基础上，符合认知规律。

#### 3. **工程化实践**

不只是"能跑"，而是"能生产"。包含：
- ✅ 完整的错误处理
- ✅ 优雅关闭（Graceful Shutdown）
- ✅ 配置管理
- ✅ 单元测试（覆盖率 70%+）
- ✅ Benchmark 性能测试

### 性能表现

使用 redis-benchmark 测试（50 并发，10000 请求）：

| 命令 | QPS | P99 延迟 |
|------|-----|---------|
| SET | ~120k | < 1ms |
| GET | ~150k | < 0.5ms |
| HSET | ~100k | < 1ms |
| LPUSH | ~90k | < 1ms |

*测试环境：本地 MacBook Pro M2*

### 项目结构

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
└── docs/               # 本文档
```

### 快速开始

```bash
# 1. 克隆项目
git clone https://github.com/strings77wzq/go-flashdb.git
cd go-flashdb

# 2. 编译
go build -o goflashdb ./cmd/goflashdb

# 3. 启动
./goflashdb

# 4. 测试
redis-cli ping
# PONG

redis-cli set hello world
# OK

redis-cli get hello
# "world"
```

### 学习路径

```mermaid
graph LR
    A[快速开始] --> B[TCP服务器]
    B --> C[RESP协议]
    C --> D[并发字典]
    D --> E[数据类型]
    E --> F[持久化]
    F --> G[事务]
    G --> H[复制]
    H --> I[Cluster]
    I --> J[性能优化]
```

### 贡献

欢迎提交 Issue 和 PR！详见 [CONTRIBUTING.md](https://github.com/strings77wzq/go-flashdb/blob/main/CONTRIBUTING.md)

### 许可

[MIT](https://github.com/strings77wzq/go-flashdb/blob/main/LICENSE)

---

**开始你的 Redis 实现之旅 →** [快速开始](/guide/)