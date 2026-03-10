## Context

goflashdb 是一个 Go 语言实现的 Redis 兼容键值存储系统。当前版本仅支持 String 类型，缺少持久化集成和其他核心数据类型。项目需要完善 Phase 1 核心功能才能成为可用的 Redis 替代品。

当前状态：
- 已有完整 RESP 协议解析
- 已有 65536 分片并发字典
- 已有 String 类型命令 (15个)
- 已有 AOF/RDB 持久化模块 (未集成)
- 已有安全模块 (Auth/RateLimiter/Filter，未集成)

## Goals / Non-Goals

**Goals:**
1. 集成持久化模块到 DB，实现数据自动保存
2. 实现 Hash/List/Set 三个核心数据类型
3. 集成安全模块到 TCP 服务器
4. 更新所有相关文档

**Non-Goals:**
- 不实现 ZSet (Phase 3)
- 不实现 Pub/Sub (Phase 3)
- 不实现事务 (Phase 3)
- 不实现主从复制 (Phase 3)
- 不添加集群支持

## Decisions

### D1: 持久化集成方式
**选择**: 同步执行 + 后台异步保存
**理由**: 简单可靠，避免复杂异步逻辑带来的 bug

### D2: 数据类型存储结构
**选择**: 继续使用已定义的 types.go 中的结构体
**理由**: 代码复用，减少重复定义

### D3: 命令注册模式
**选择**: 保持现有 init() 注册模式
**理由**: 与现有代码风格一致

### D4: 安全模块集成
**选择**: 中间件模式，在命令执行前检查
**理由**: 解耦安全逻辑和业务逻辑

## Risks / Trade-offs

- [Risk] 持久化可能影响性能 → [Mitigation] 使用后台 goroutine 异步写入
- [Risk] 新数据类型可能有类型冲突 → [Mitigation] 在命令执行前检查 key 类型
- [Risk] 安全模块集成可能引入 bug → [Mitigation] 先集成最基础的认证，其他功能逐步添加

## Migration Plan

1. 首先集成持久化 (最核心)
2. 然后添加 Hash/List/Set 命令
3. 最后集成安全模块
4. 更新文档和 CHANGELOG

## Open Questions

- 是否需要支持 EXPIRE 命令对所有数据类型生效? (暂不实现，后续版本)
- 安全模块是否需要配置文件支持? (暂不实现，使用代码配置)
