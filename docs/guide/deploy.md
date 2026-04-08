# GitHub Actions 自动部署

配置 GitHub Actions，实现文档自动构建和部署到 GitHub Pages。

## 配置步骤

### 1. 创建工作流文件

```bash
mkdir -p .github/workflows
```

### 2. 创建 deploy-docs.yml

```yaml
name: Deploy Docs to GitHub Pages

on:
  push:
    branches:
      - main
    paths:
      - 'docs/**'
      - 'package.json'
      - '.github/workflows/docs.yml'
  
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: false

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        
      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm
          
      - name: Install dependencies
        run: npm ci
        
      - name: Build docs
        run: npm run docs:build
        
      - name: Upload artifact
        uses: actions/upload-pages-artifact@v3
        with:
          path: docs/.vuepress/dist

  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    runs-on: ubuntu-latest
    needs: build
    steps:
      - name: Deploy to GitHub Pages
        id: deployment
        uses: actions/deploy-pages@v4
```

### 3. 启用 GitHub Pages

1. 进入 GitHub 仓库 Settings → Pages
2. Source 选择 "GitHub Actions"
3. 保存

### 4. 部署完成

访问：`https://strings77wzq.github.io/go-flashdb/`

## 本地预览

```bash
# 安装依赖
npm install

# 开发模式
npm run docs:dev

# 构建
npm run docs:build
```

## 自定义域名（可选）

1. 添加 CNAME 文件到 docs/.vuepress/public/
2. 在 DNS 配置 CNAME 记录
3. 在 GitHub Pages 设置自定义域名