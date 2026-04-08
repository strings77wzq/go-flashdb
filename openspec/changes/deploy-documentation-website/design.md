## Context

go-flashdb 已实现 v0.3.0 功能，需要参考 claude-code-Go 和 godis 的文档模式，建立一个完整的文档网站。

参考网站特点 (claude-code-Go)：
1. VuePress 构建的现代化文档站点
2. 清晰的首页 Hero + Features
3. 侧边栏导航 organized by topic
4. 代码高亮和 copy 按钮
5. 中英文切换
6. GitHub 集成

godis 文档特点：
1. 系列教程形式（11篇）
2. 原理+设计+代码的结构
3. 详细的代码注释
4. 架构图和流程图

## Goals / Non-Goals

**Goals:**
- 使用 VuePress 2 构建文档网站
- 实现类似 claude-code-Go 的首页设计
- 编写 12 篇系列教程（参考 godis 模式）
- 部署到 GitHub Pages
- 配置自动 CI/CD

**Non-Goals:**
- 视频教程
- 在线 playground（超出当前范围）
- 多语言翻译（优先中文）

## Decisions

### 1. 静态站点生成器选择

**决策**: 使用 VuePress 2

**理由**:
- VuePress 是 Vue 官方文档工具，成熟稳定
- 支持 Markdown，易于维护
- 内置代码高亮、搜索、导航
- 大量中文文档项目使用
- claude-code-Go 也是类似技术栈

**替代方案**:
- Docusaurus (Facebook) → 学习成本稍高
- GitBook → 商业产品，不够灵活
- 自建 → 工作量太大

### 2. 文档组织结构

**决策**: 参考 godis 系列，按实现顺序组织

**结构**:
```
指南 (Guide)
├── 01-快速开始
├── 02-TCP服务器
├── 03-RESP协议
├── 04-并发字典
├── 05-数据类型
├── 06-TTL过期
├── 07-AOF持久化
├── 08-RDB快照
├── 09-事务
├── 10-主从复制
├── 11-Cluster协议
└── 12-性能优化

架构 (Architecture)
├── 概述
├── 模块设计
└── 性能分析

API
├── 命令列表
└── 错误码
```

### 3. 部署方式

**决策**: GitHub Actions 自动部署到 GitHub Pages

**流程**:
1. Push to main → 触发 Action
2. Build VuePress site
3. Deploy to gh-pages branch
4. GitHub Pages 自动发布

### 4. 内容风格

**决策**: 每篇教程遵循统一模板

**模板**:
```markdown
# 标题

## 概述
- 本文目标
- 前置知识

## 原理/设计
- 架构图
- 核心概念

## 实现
- 分步骤讲解
- 完整代码
- 关键代码解析

## 测试/验证
- 如何测试
- 预期结果

## 总结
- 要点回顾
- 下篇预告

## 参考
- 源码链接
- 相关文档
```

## Risks / Trade-offs

### Risk 1: 内容质量
**风险**: 12篇教程工作量大，可能质量参差不齐
**缓解**: 使用 AI 辅助生成初稿，人工审核修改

### Risk 2: 维护成本
**风险**: 代码更新后文档不同步
**缓解**: 在代码注释中添加文档链接，定期 review

### Risk 3: 网站性能
**风险**: 大量图片和代码块影响加载
**缓解**: 使用懒加载、代码折叠、图片压缩