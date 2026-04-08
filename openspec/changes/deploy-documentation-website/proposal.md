## Why

go-flashdb 目前缺少一个集中式的学习和文档中心。虽然项目已实现丰富的功能（Redis 协议兼容、事务、复制、Cluster、Benchmark），但用户难以快速上手和深入学习。

参考成功案例：
- claude-code-Go (用户自己的项目): https://strings77wzq.github.io/claude-code-Go/zh/
- godis 博客系列 (11篇教程，带来 3.8k stars)

目标是在 GitHub Pages 上部署一个类似 claude-code-Go 的文档官网，提供：
1. 清晰的上手指南
2. 详细的架构教程（类似 godis 博客系列）
3. 完整的 API 参考
4. 交互式示例

## What Changes

### P0 - 文档网站框架
使用 VuePress 构建静态文档网站：
- 首页：项目介绍、功能亮点、快速开始
- 指南系列（11篇教程）：
  1. 快速开始 - 安装和基本使用
  2. TCP 服务器 - 网络层实现
  3. RESP 协议 - 协议解析器
  4. 并发字典 - 分片锁设计
  5. 数据类型 - String/Hash/List/Set/ZSet
  6. TTL 过期 - 时间轮实现
  7. AOF 持久化 - 日志和重写
  8. RDB 快照 - 二进制编码
  9. 事务 - WATCH/MULTI/EXEC
  10. 主从复制 - 复制协议
  11. Cluster 协议 - 槽位和重定向
  12. 性能优化 - Benchmark
- API 参考：所有命令文档
- 架构设计：系统架构图

### P1 - GitHub Pages 部署
- 配置 GitHub Actions 自动部署
- 绑定自定义域名（可选）
- SEO 优化

### P2 - 内容完善
- 每篇教程包含：原理讲解、代码示例、架构图
- 添加交互式代码 playground（可选）
- 中英双语支持

## Capabilities

### New Capabilities
- `documentation-website`: VuePress 文档网站
- `github-pages-deployment`: GitHub Pages 自动部署

### Modified Capabilities
- (无)

## Impact

### 技术栈
- VuePress v2 (Vue 3 + Vite)
- Markdown 文档
- GitHub Actions CI/CD

### 文件结构
```
docs/
├── .vuepress/          # VuePress 配置
│   ├── config.js       # 站点配置
│   ├── navbar.js       # 导航栏
│   ├── sidebar.js      # 侧边栏
│   └── styles/         # 自定义样式
├── guide/              # 指南
│   ├── 01-quickstart.md
│   ├── 02-tcp-server.md
│   └── ...
├── api/                # API 参考
├── architecture/       # 架构文档
└── README.md           # 首页
```

### 发布地址
- https://strings77wzq.github.io/go-flashdb/