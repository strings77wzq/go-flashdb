export const redirects = JSON.parse("{}")

export const routes = Object.fromEntries([
  ["/", { loader: () => import(/* webpackChunkName: "index.html" */"/home/strin/go/src/devLearn/aiLab/goflashdb/docs/README.md"), meta: {"title":""} }],
  ["/api-reference.html", { loader: () => import(/* webpackChunkName: "api-reference.html" */"/home/strin/go/src/devLearn/aiLab/goflashdb/docs/api-reference.md"), meta: {"title":"GoSwiftKV API 参考手册"} }],
  ["/architecture.html", { loader: () => import(/* webpackChunkName: "architecture.html" */"/home/strin/go/src/devLearn/aiLab/goflashdb/docs/architecture.md"), meta: {"title":"GoSwiftKV 架构设计"} }],
  ["/development.html", { loader: () => import(/* webpackChunkName: "development.html" */"/home/strin/go/src/devLearn/aiLab/goflashdb/docs/development.md"), meta: {"title":"GoSwiftKV 开发指南"} }],
  ["/getting-started.html", { loader: () => import(/* webpackChunkName: "getting-started.html" */"/home/strin/go/src/devLearn/aiLab/goflashdb/docs/getting-started.md"), meta: {"title":"GoSwiftKV 入门指南"} }],
  ["/api/", { loader: () => import(/* webpackChunkName: "api_index.html" */"/home/strin/go/src/devLearn/aiLab/goflashdb/docs/api/README.md"), meta: {"title":"go-flashdb API 文档"} }],
  ["/design/", { loader: () => import(/* webpackChunkName: "design_index.html" */"/home/strin/go/src/devLearn/aiLab/goflashdb/docs/design/README.md"), meta: {"title":"设计理念"} }],
  ["/development/", { loader: () => import(/* webpackChunkName: "development_index.html" */"/home/strin/go/src/devLearn/aiLab/goflashdb/docs/development/README.md"), meta: {"title":"go-flashdb 开发指南"} }],
  ["/development/architecture.html", { loader: () => import(/* webpackChunkName: "development_architecture.html" */"/home/strin/go/src/devLearn/aiLab/goflashdb/docs/development/architecture.md"), meta: {"title":"go-flashdb 架构设计"} }],
  ["/guide/01-quickstart.html", { loader: () => import(/* webpackChunkName: "guide_01-quickstart.html" */"/home/strin/go/src/devLearn/aiLab/goflashdb/docs/guide/01-quickstart.md"), meta: {"title":"快速开始"} }],
  ["/guide/02-tcp-server.html", { loader: () => import(/* webpackChunkName: "guide_02-tcp-server.html" */"/home/strin/go/src/devLearn/aiLab/goflashdb/docs/guide/02-tcp-server.md"), meta: {"title":"TCP 服务器：goroutine-per-connection 的艺术"} }],
  ["/guide/04-concurrent-dict.html", { loader: () => import(/* webpackChunkName: "guide_04-concurrent-dict.html" */"/home/strin/go/src/devLearn/aiLab/goflashdb/docs/guide/04-concurrent-dict.md"), meta: {"title":"并发字典：为什么选 65536 分片？"} }],
  ["/guide/deploy.html", { loader: () => import(/* webpackChunkName: "guide_deploy.html" */"/home/strin/go/src/devLearn/aiLab/goflashdb/docs/guide/deploy.md"), meta: {"title":"GitHub Actions 自动部署"} }],
  ["/learning/", { loader: () => import(/* webpackChunkName: "learning_index.html" */"/home/strin/go/src/devLearn/aiLab/goflashdb/docs/learning/README.md"), meta: {"title":"go-flashdb 学习指南"} }],
  ["/404.html", { loader: () => import(/* webpackChunkName: "404.html" */"/home/strin/go/src/devLearn/aiLab/goflashdb/docs/.vuepress/.temp/pages/404.html.vue"), meta: {"title":""} }],
]);
