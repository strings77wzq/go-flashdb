## Why

goflashdb 目前仅支持 String 数据类型，功能完整性严重不足。用户需要 Hash/List/Set 等 Redis 核心数据类型，以及持久化集成和安全模块集成才能正常使用。当前状态无法满足基本的 Redis 兼容需求，亟需完善核心功能。

## What Changes

1. **集成持久化到 DB**: 将 AOF/RDB 持久化模块与 DB 核心逻辑集成，实现数据自动持久化
2. **实现 Hash 数据类型**: 支持 HSET/HGET/HDEL/HMGET/HGETALL/HEXISTS 命令
3. **实现 List 数据类型**: 支持 LPUSH/RPUSH/LPOP/RPOP/LRANGE/LLEN 命令
4. **实现 Set 数据类型**: 支持 SADD/SREM/SISMEMBER/SMEMBERS/SCARD 命令
5. **集成安全模块到 Server**: 将认证、限流、命令过滤集成到 TCP 服务器

## Capabilities

### New Capabilities
- `persistence-integration`: 持久化模块与 DB 核心集成，实现自动数据持久化
- `hash-data-type`: Redis Hash 数据类型完整支持
- `list-data-type`: Redis List 数据类型完整支持  
- `set-data-type`: Redis Set 数据类型完整支持
- `security-integration`: 安全模块与服务器集成

### Modified Capabilities
- 无

## Impact

- 修改 `pkg/core/db.go`: 添加持久化集成和新的命令注册
- 修改 `pkg/core/string.go`: 集成 TTL 过期逻辑
- 新增 `pkg/core/hash.go`: Hash 类型命令实现
- 新增 `pkg/core/list.go`: List 类型命令实现
- 新增 `pkg/core/set.go`: Set 类型命令实现
- 修改 `pkg/net/server.go`: 集成安全模块
