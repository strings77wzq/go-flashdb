## 1. 持久化集成

- [ ] 1.1 初始化 PersistManager 并传入 DB 构造函数
- [ ] 1.2 修改 DB.Exec() 方法，在写命令后调用 AOF Append
- [ ] 1.3 实现 DB.LoadFromPersist() 启动时加载数据
- [ ] 1.4 添加 SAVE 命令支持手动触发快照

## 2. Hash 数据类型实现

- [ ] 2.1 在 types.go 中确保 HashData 结构完整
- [ ] 2.2 在 core 包添加 HSET/HGET/HDEL/HMGET/HGETALL/HEXISTS 命令
- [ ] 2.3 添加 Hash 类型检查逻辑
- [ ] 2.4 测试 Hash 命令基本功能

## 3. List 数据类型实现

- [ ] 3.1 在 types.go 中确保 ListData 结构完整
- [ ] 3.2 在 core 包添加 LPUSH/RPUSH/LPOP/RPOP/LRANGE/LLEN 命令
- [ ] 3.3 添加 List 类型检查逻辑
- [ ] 3.4 测试 List 命令基本功能

## 4. Set 数据类型实现

- [ ] 4.1 在 types.go 中确保 SetData 结构完整
- [ ] 4.2 在 core 包添加 SADD/SREM/SISMEMBER/SMEMBERS/SCARD 命令
- [ ] 4.3 添加 Set 类型检查逻辑
- [ ] 4.4 测试 Set 命令基本功能

## 5. 安全模块集成

- [ ] 5.1 在 Server 初始化时创建 Authenticator
- [ ] 5.2 添加 AUTH 命令支持密码认证
- [ ] 5.3 在 Server 接入 RateLimiter 中间件
- [ ] 5.4 在命令执行前检查危险命令

## 6. 文档更新

- [ ] 6.1 更新 README.md 支持的命令列表
- [ ] 6.2 更新 api-reference.md 添加 Hash/List/Set 命令文档
- [ ] 6.3 更新 architecture.md 架构图
- [ ] 6.4 创建 CHANGELOG.md 记录版本变更
