## 1. Server Commands (P1 - 高优先级)

### 1.1 DBSIZE Command
- [x] 1.1.1 在 pkg/core/db.go 添加 DBSIZE 命令实现
- [x] 1.1.2 添加单元测试覆盖空数据库和正常情况

### 1.2 INFO Command
- [x] 1.2.1 实现 INFO server 部分（版本、运行时间、内存）
- [x] 1.2.2 实现 INFO memory 部分（内存统计）
- [x] 1.2.3 实现 INFO stats 部分（命令统计）
- [x] 1.2.4 添加单元测试

### 1.3 CONFIG GET/SET Command
- [x] 1.3.1 实现 CONFIG GET 支持通配符匹配
- [x] 1.3.2 实现 CONFIG SET 支持运行时配置变更
- [x] 1.3.3 添加保护（不允许修改关键配置）
- [x] 1.3.4 添加单元测试

### 1.4 FLUSHDB/FLUSHALL Command
- [x] 1.4.1 实现 FLUSHDB（清空当前数据库）
- [x] 1.4.2 实现 FLUSHDB ASYNC（异步删除）
- [x] 1.4.3 实现 FLUSHALL（清空所有数据库）
- [x] 1.4.4 添加单元测试

### 1.5 CLIENT LIST Command
- [x] 1.5.1 实现 CLIENT LIST 返回连接信息
- [x] 1.5.2 添加单元测试

## 2. Transactions (P2 - 中优先级)

### 2.1 Watch Mechanism
- [x] 2.1.1 在 DB 结构中添加 watchMap
- [x] 2.1.2 实现 WATCH 命令（注册 key 监视）
- [x] 2.1.3 实现 UNWATCH 命令（取消监视）
- [x] 2.1.4 实现修改时检查 watch 状态

### 2.2 Multi/Exec Transaction
- [x] 2.2.1 实现 MULTI 命令（进入事务模式）
- [x] 2.2.2 实现命令队列（将命令加入队列）
- [x] 2.2.3 实现 EXEC 命令（执行队列中所有命令）
- [x] 2.2.4 实现 DISCARD 命令（取消事务）
- [x] 2.2.5 处理事务中的错误（继续执行或回滚）

### 2.3 Transaction Tests
- [ ] 2.3.1 基础事务测试（MULTI/EXEC）
- [ ] 2.3.2 WATCH 冲突测试
- [ ] 2.3.3 DISCARD 测试
- [ ] 2.3.4 事务内错误处理测试

## 3. Replication (P1 - 高优先级)

### 3.1 Master Implementation
- [x] 3.1.1 扩展 DB 结构添加 replication 相关字段
- [x] 3.1.2 实现 replication ID 生成
- [x] 3.1.3 实现复制偏移量追踪
- [x] 3.1.4 实现 slave 连接管理
- [x] 3.1.5 实现命令传播到所有 slave

### 3.2 Slave Implementation
- [ ] 3.2.1 实现 slave 连接逻辑
- [ ] 3.2.2 实现 PSYNC 命令（同步请求）
- [ ] 3.2.3 实现 RDB 全量同步接收
- [ ] 3.2.4 实现增量命令同步
- [ ] 3.2.5 实现 replication 状态转换

### 3.3 Replication Commands
- [x] 3.3.1 实现 ROLE 命令
- [x] 3.3.2 实现 REPLICAOF NO ONE 命令
- [x] 3.3.3 实现 REPLICAOF host port 命令
- [x] 3.3.4 配置中支持指定 master

### 3.4 Replication Tests
- [ ] 3.4.1 Master 启动测试
- [ ] 3.4.2 Slave 同步测试
- [ ] 3.4.3 增量同步测试
- [ ] 3.4.4 ROLE 命令测试

## 4. Benchmark (P2 - 中优先级)

### 4.1 Benchmark Framework
- [x] 4.1.1 创建 pkg/benchmark 模块
- [x] 4.1.2 实现标准参数解析（-c 并发数，-n 请求数）
- [x] 4.1.3 实现基准测试运行器

### 4.2 Performance Tests
- [x] 4.2.1 实现 PING 基准测试
- [x] 4.2.2 实现 SET/GET 基准测试
- [x] 4.2.3 实现 LPUSH/LPOP 基准测试
- [x] 4.2.4 实现 HSET/HGET 基准测试

### 4.3 Statistics
- [x] 4.3.1 实现延迟收集
- [x] 4.3.2 实现 P50/P95/P99 统计
- [x] 4.3.3 实现 QPS 计算
- [x] 4.3.4 添加结果报告输出

## 5. Redis Cluster (P0 - 最高优先级)

### 5.1 Slot Management
- [x] 5.1.1 实现 CRC16 槽位计算
- [x] 5.1.2 实现 16384 槽位管理
- [x] 5.1.3 实现槽位查询接口

### 5.2 Cluster Commands
- [x] 5.2.1 实现 CLUSTER SLOTS 命令
- [x] 5.2.2 实现 CLUSTER INFO 命令
- [x] 5.2.3 实现 CLUSTER NODES 命令
- [x] 5.2.4 实现 CLUSTER ADDSLOTS 命令

### 5.3 Redirection
- [ ] 5.3.1 实现 MOVED 错误响应
- [ ] 5.3.2 实现 ASK 错误响应
- [ ] 5.3.3 添加客户端重试提示

### 5.4 Cluster Mode
- [x] 5.4.1 添加 cluster 配置项
- [x] 5.4.2 实现 Cluster 模式启动
- [x] 5.4.3 路由到正确的槽位处理

### 5.5 Cluster Tests
- [x] 5.5.1 槽位计算测试
- [x] 5.5.2 CLUSTER 命令测试
- [ ] 5.5.3 重定向测试

## 6. Integration & Polish

### 6.1 CI/CD
- [ ] 6.1.1 添加新功能测试到 CI
- [ ] 6.1.2 添加 race condition 检测
- [ ] 6.1.3 添加 benchmark 到 CI

### 6.2 Documentation
- [ ] 6.2.1 更新 README 添加新功能说明
- [ ] 6.2.2 更新 API 参考文档
- [ ] 6.2.3 添加 Replication 配置说明
- [ ] 6.2.4 添加 Cluster 使用说明

### 6.3 Release
- [ ] 6.3.1 更新 CHANGELOG
- [ ] 6.3.2 版本号更新到 v0.3.0
- [ ] 6.3.3 创建 Release