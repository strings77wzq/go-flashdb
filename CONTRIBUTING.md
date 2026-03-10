# 贡献指南

欢迎贡献 goflashdb！

## 行为准则

请阅读并遵守我们的 [行为准则](CODE_OF_CONDUCT.md)。

## 如何贡献

### 报告 Bug

1. 搜索现有 issues 确认没有重复
2. 使用 Bug 模板创建 issue
3. 提供复现步骤和环境信息

### 提出新功能

1. 搜索现有 issues 和 PRs
2. 使用 Feature Request 模板
3. 详细描述功能需求和使用场景

### 提交代码

#### 开发环境设置

```bash
# 克隆项目
git clone https://github.com/strings77wzq/go-flashdb.git
cd go-flashdb

# 安装依赖
go mod download

# 运行测试
go test ./...

# 构建
go build -o goflashdb ./cmd/goflashdb
```

#### 代码规范

- 遵循 Go 代码规范
- 使用 `gofmt` 格式化代码
- 添加必要的注释
- 确保新增代码有测试覆盖

#### 提交信息格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

类型 (type):
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具

示例:
```
feat(hash): add HCLEAR command

Add HCLEAR command to delete all fields in a hash.
Fixes #123
```

#### Pull Request 流程

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改
4. 推送分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 测试要求

- 所有功能必须有对应测试
- 测试覆盖率不低于 80%
- 运行测试: `go test -cover ./...`

### 代码审查标准

- 代码必须通过 `go vet` 和 `golint`
- 测试必须通过
- 必须有适当的注释

## 开发指南

### 项目结构

```
goflashdb/
├── cmd/goflashdb/    # 主程序
├── pkg/
│   ├── core/         # 核心命令实现
│   ├── resp/         # RESP 协议
│   ├── persist/       # 持久化
│   ├── security/     # 安全模块
│   ├── net/          # 网络层
│   └── config/       # 配置
└── docs/             # 文档
```

### 添加新命令

1. 在 `pkg/core/` 创建新文件或添加到现有文件
2. 实现命令函数
3. 在 `init()` 中注册命令
4. 添加对应测试

### 文档

- API 文档在 `docs/api-reference.md`
- 架构文档在 `docs/architecture.md`

## 许可证

贡献代码即表示你同意 MIT 许可证。
