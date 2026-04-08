import { defaultTheme } from '@vuepress/theme-default'
import { defineUserConfig } from 'vuepress'
import { searchPlugin } from '@vuepress/plugin-search'

export default defineUserConfig({
  lang: 'zh-CN',
  title: 'Go-FlashDB',
  description: 'Go 语言实现的高性能 Redis 服务器 - 深入源码学习',
  
  head: [
    ['link', { rel: 'icon', href: '/favicon.ico' }],
    ['meta', { name: 'keywords', content: 'redis, go, golang, 源码, 教程, 数据库' }],
  ],

  theme: defaultTheme({
    logo: '/logo.svg',
    
    navbar: [
      { text: '首页', link: '/' },
      { text: '快速开始', link: '/guide/' },
      { 
        text: '源码学习',
        children: [
          { text: '设计理念', link: '/design/' },
          { text: 'TCP 服务器', link: '/guide/02-tcp-server.html' },
          { text: 'RESP 协议', link: '/guide/03-resp-protocol.html' },
          { text: '并发字典', link: '/guide/04-concurrent-dict.html' },
          { text: '数据类型', link: '/guide/05-data-types.html' },
          { text: '持久化', link: '/guide/07-aof-persistence.html' },
          { text: '事务', link: '/guide/09-transactions.html' },
          { text: '主从复制', link: '/guide/10-replication.html' },
          { text: 'Cluster', link: '/guide/11-cluster.html' },
        ]
      },
      { text: '架构设计', link: '/architecture/' },
      { text: 'API', link: '/api/commands.html' },
      { text: 'GitHub', link: 'https://github.com/strings77wzq/go-flashdb' },
    ],

    sidebar: {
      '/guide/': [
        {
          text: '入门',
          collapsible: false,
          children: [
            '/guide/README.md',
            '/guide/01-quickstart.md',
          ]
        },
        {
          text: '核心模块',
          collapsible: false,
          children: [
            '/guide/02-tcp-server.md',
            '/guide/03-resp-protocol.md',
            '/guide/04-concurrent-dict.md',
            '/guide/05-data-types.md',
            '/guide/06-ttl-expiration.md',
          ]
        },
        {
          text: '高级特性',
          collapsible: false,
          children: [
            '/guide/07-aof-persistence.md',
            '/guide/08-rdb-snapshot.md',
            '/guide/09-transactions.md',
            '/guide/10-replication.md',
            '/guide/11-cluster.md',
            '/guide/12-performance.md',
          ]
        },
      ],
      '/design/': [
        {
          text: '设计理念',
          collapsible: false,
          children: [
            '/design/README.md',
            '/design/why-golang.md',
            '/design/concurrency-model.md',
            '/design/memory-management.md',
            '/design/performance-optimization.md',
          ]
        },
      ],
      '/architecture/': [
        {
          text: '架构设计',
          collapsible: false,
          children: [
            '/architecture/README.md',
            '/architecture/modules.md',
            '/architecture/dataflow.md',
            '/architecture/extension.md',
          ]
        },
      ],
      '/api/': [
        {
          text: 'API 参考',
          collapsible: false,
          children: [
            '/api/commands.md',
            '/api/errors.md',
          ]
        },
      ],
    },

    editLink: true,
    editLinkText: '在 GitHub 上编辑此页',
    lastUpdated: true,
    lastUpdatedText: '最后更新',
    contributors: true,
    contributorsText: '贡献者',
    
    backToHome: '返回首页',
    openInNewWindow: '在新窗口打开',
    toggleColorMode: '切换颜色模式',
    toggleSidebar: '切换侧边栏',
  }),

  plugins: [
    searchPlugin({
      locales: {
        '/': {
          placeholder: '搜索文档',
        },
      },
    }),
  ],

  base: '/go-flashdb/',
})