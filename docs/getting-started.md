# GoSwiftKV 入门指南

## 快速开始

### 安装

```bash
go install github.com/goswiftkv/goswiftkv@latest
```

### 启动服务器

```bash
goswiftkv
```

默认监听 `0.0.0.0:6379`，可直接使用 `redis-cli` 连接：

```bash
redis-cli ping
# PONG
```

## 基本使用

### 字符串操作

```bash
# 设置值
redis-cli set mykey "Hello GoSwiftKV"
# OK

# 获取值
redis-cli get mykey
# "Hello GoSwiftKV"

# 设置过期时间（秒）
redis-cli setex mykey 60 "expires in 60s"
# OK

# 批量设置
redis-cli mset key1 value1 key2 value2
# OK

# 批量获取
redis-cli mget key1 key2
# 1) "value1"
# 2) "value2"

# 数值操作
redis-cli set counter 10
redis-cli incr counter
# (integer) 11
redis-cli incrby counter 5
# (integer) 16
```

### 键管理

```bash
# 检查键是否存在
redis-cli exists mykey
# (integer) 1

# 删除键
redis-cli del mykey
# (integer) 1

# 查看键类型
redis-cli type mykey
# string
```

## 配置

### 配置文件

创建 `config.yaml`:

```yaml
# 网络配置
bind_addr: "0.0.0.0:6379"
max_conn: 10000
timeout: 300

# 内存配置
max_memory: 0  # 0 表示不限制
eviction_policy: "volatile-lru"

# 持久化配置
append_only: true
append_filename: "appendonly.aof"
append_fsync: "everysec"
rdb_filename: "dump.rdb"

# 安全配置
require_pass: "your-password"
max_clients: 10000

# AI 扩展配置
enable_ai: false
openclaw_endpoint: ""
mcp_server_addr: ""
```

### 启动时指定配置

```bash
goswiftkv -c /path/to/config.yaml
```

## 持久化

### RDB 快照

```bash
# 手动触发快照
redis-cli bgsave
# Background saving started
```

### AOF 日志

在配置文件中启用：

```yaml
append_only: true
append_fsync: "everysec"  # always / everysec / no
```

## 性能优化建议

### 1. 内存优化

- 合理设置 `max_memory`
- 选择合适的淘汰策略
- 避免大 Key（> 10KB）

### 2. 并发优化

- 使用连接池
- 批量操作使用 MSET/MGET
- 避免阻塞命令

### 3. 持久化优化

- 生产环境建议 AOF + RDB 混合
- `append_fsync` 选择 `everysec`
- 定期执行 BGREWRITEAOF

## 下一步

- [API 参考](./api-reference.md)
- [架构设计](./architecture.md)
- [部署指南](./deployment.md)
