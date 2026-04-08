## Why

go-flashdb 目前虽然具备基础 Redis 协议兼容功能（String/Hash/List/Set/ZSET/Bitmap/HyperLogLog/PubSub/Lua），但在生产环境可用性方面与 godis 等成熟项目存在显著差距。缺乏 Redis Cluster 支持限制了水平扩展能力，缺少关键 Server 命令影响运维体验，主从复制未完成导致无高可用保障，事务功能不完整影响数据一致性，缺少基准测试无法量化性能优势。这些差距导致 go-flashdb 难以作为生产级解决方案，同时在与 godis (3.8k stars) 竞争社区认可时处于劣势。

## What Changes

### P0 - Redis Cluster 协议支持
- 实现 Cluster 模式下槽位（Slot）管理（0-16383 共 16384 槽）
- 实现 `CLUSTER SLOTS`、`CLUSTER INFO`、`CLUSTER NODES` 命令
- 实现 `MOVED` 和 `ASK` 重定向机制
- 支持 `CANCELLED` 和 `TRYAGAIN` 错误响应
- 实现客户端重试逻辑

### P1 - Server 命令
- 实现 `DBSIZE` - 返回当前数据库 key 数量
- 实现 `INFO` - 返回服务器信息和统计
- 实现 `CONFIG GET <pattern>` - 获取配置项
- 实现 `CONFIG SET <key> <value>` - 设置配置项
- 实现 `FLUSHDB` - 清空当前数据库
- 实现 `FLUSHALL` - 清空所有数据库
- 实现 `CLIENT LIST` - 列出客户端连接

### P1 - 主从复制 (Replication)
- 实现 Master 角色：复制偏移量、PSYNC 命令
- 实现 Slave 角色：连接管理、命令同步
- 实现复制缓冲区和复制积压
- 实现全量复制（RDB）和增量复制
- 支持 `ROLE`、`REPLICAOF`、`SLAVEOF` 命令

### P2 - 事务功能
- 实现 `WATCH` 命令 - 乐观锁
- 实现 `MULTI` / `EXEC` - 事务执行
- 实现 `DISCARD` - 取消事务
- 实现事务内命令排队和原子执行

### P2 - 基准测试
- 实现标准 redis-benchmark 兼容测试
- 实现 PING、SET、GET、LPUSH、HSET 等核心命令测试
- 实现延迟分布统计（P50/P95/P99）
- 添加持续压测模式和报告生成

## Capabilities

### New Capabilities
- `redis-cluster`: Redis Cluster 协议兼容和槽位管理
- `server-commands`: 运维相关 Server 命令
- `replication`: 主从复制和高可用支持
- `transactions`: 事务 WATCH/MULTI/EXEC 支持
- `benchmarking`: 性能基准测试套件

### Modified Capabilities
- (无现有 spec 需要修改，这是新增功能集)

## Impact

### 代码影响
- `pkg/core/` - 新增 Cluster 路由、事务执行器
- `pkg/core/` - 扩展 DB 结构支持复制状态
- `pkg/net/` - 支持新 Server 命令处理
- `pkg/replication/` - 完善现有 replication 模块
- 新增 `pkg/benchmark/` - 基准测试模块

### 依赖影响
- 无新外部依赖

### 系统影响
- 协议层需扩展错误类型支持 Cluster 重定向
- 持久化层需支持复制相关状态
- 配置层需支持运行时配置变更

### 用户影响
- 使用方式从单节点扩展到集群模式
- 支持主从部署提升可用性
- 支持事务保证数据一致性