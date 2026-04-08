import{A as e,d as t,m as n,t as r}from"./plugin-vue_export-helper--IMkQmEh.js";var i=JSON.parse(`{"path":"/guide/01-quickstart.html","title":"快速开始","lang":"zh-CN","frontmatter":{},"git":{"updatedTime":1775618263000,"contributors":[{"name":"strings77wzq","username":"strings77wzq","email":"strings77wzq@github.com","commits":1,"url":"https://github.com/strings77wzq"}],"changelog":[{"hash":"3c536800d4175146327c71649c552297dd8add0c","time":1775618263000,"email":"strings77wzq@github.com","author":"strings77wzq","message":"feat: implement v0.3.0 features and documentation website","tag":"v0.3.0"}]},"filePathRelative":"guide/01-quickstart.md"}`),a={name:`01-quickstart.md`};function o(r,i,a,o,s,c){return e(),t(`div`,null,[...i[0]||=[n(`<h1 id="快速开始" tabindex="-1"><a class="header-anchor" href="#快速开始"><span>快速开始</span></a></h1><p>本文将带你快速上手 go-flashdb，体验一个 Go 语言实现的 Redis 服务器。</p><h2 id="安装" tabindex="-1"><a class="header-anchor" href="#安装"><span>安装</span></a></h2><h3 id="方式一-源码编译-推荐" tabindex="-1"><a class="header-anchor" href="#方式一-源码编译-推荐"><span>方式一：源码编译（推荐）</span></a></h3><div class="language-bash line-numbers-mode" data-highlighter="prismjs" data-ext="sh"><pre><code class="language-bash"><span class="line"><span class="token comment"># 1. 克隆项目</span></span>
<span class="line"><span class="token function">git</span> clone https://github.com/strings77wzq/go-flashdb.git</span>
<span class="line"><span class="token builtin class-name">cd</span> go-flashdb</span>
<span class="line"></span>
<span class="line"><span class="token comment"># 2. 编译（需要 Go 1.21+）</span></span>
<span class="line">go build <span class="token parameter variable">-o</span> goflashdb ./cmd/goflashdb</span>
<span class="line"></span>
<span class="line"><span class="token comment"># 3. 验证版本</span></span>
<span class="line">./goflashdb <span class="token parameter variable">--version</span></span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><h3 id="方式二-使用-go-install" tabindex="-1"><a class="header-anchor" href="#方式二-使用-go-install"><span>方式二：使用 go install</span></a></h3><div class="language-bash line-numbers-mode" data-highlighter="prismjs" data-ext="sh"><pre><code class="language-bash"><span class="line">go <span class="token function">install</span> github.com/strings77wzq/go-flashdb/cmd/goflashdb@latest</span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div></div></div><h2 id="启动服务器" tabindex="-1"><a class="header-anchor" href="#启动服务器"><span>启动服务器</span></a></h2><div class="language-bash line-numbers-mode" data-highlighter="prismjs" data-ext="sh"><pre><code class="language-bash"><span class="line"><span class="token comment"># 默认配置启动</span></span>
<span class="line">./goflashdb</span>
<span class="line"></span>
<span class="line"><span class="token comment"># 输出：</span></span>
<span class="line"><span class="token comment"># [INFO] 2026/01/08 10:30:00 goflashdb v0.3.0 starting...</span></span>
<span class="line"><span class="token comment"># [INFO] 2026/01/08 10:30:00 Listening on :6379</span></span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><p>服务器默认监听 6379 端口，与 Redis 默认端口一致。</p><h2 id="基本使用" tabindex="-1"><a class="header-anchor" href="#基本使用"><span>基本使用</span></a></h2><p>使用 redis-cli 或其他 Redis 客户端连接测试：</p><div class="language-bash line-numbers-mode" data-highlighter="prismjs" data-ext="sh"><pre><code class="language-bash"><span class="line"><span class="token comment"># 1. Ping 测试</span></span>
<span class="line">redis-cli <span class="token function">ping</span></span>
<span class="line"><span class="token comment"># PONG</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 2. String 操作</span></span>
<span class="line">redis-cli <span class="token builtin class-name">set</span> hello world</span>
<span class="line"><span class="token comment"># OK</span></span>
<span class="line"></span>
<span class="line">redis-cli get hello</span>
<span class="line"><span class="token comment"># &quot;world&quot;</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 3. Hash 操作</span></span>
<span class="line">redis-cli hset user:1 name <span class="token string">&quot;Alice&quot;</span> age <span class="token number">30</span></span>
<span class="line"><span class="token comment"># (integer) 2</span></span>
<span class="line"></span>
<span class="line">redis-cli hgetall user:1</span>
<span class="line"><span class="token comment"># 1) &quot;name&quot;</span></span>
<span class="line"><span class="token comment"># 2) &quot;Alice&quot;</span></span>
<span class="line"><span class="token comment"># 3) &quot;age&quot;</span></span>
<span class="line"><span class="token comment"># 4) &quot;30&quot;</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 4. List 操作</span></span>
<span class="line">redis-cli lpush mylist a b c</span>
<span class="line"><span class="token comment"># (integer) 3</span></span>
<span class="line"></span>
<span class="line">redis-cli lrange mylist <span class="token number">0</span> <span class="token parameter variable">-1</span></span>
<span class="line"><span class="token comment"># 1) &quot;c&quot;</span></span>
<span class="line"><span class="token comment"># 2) &quot;b&quot;</span></span>
<span class="line"><span class="token comment"># 3) &quot;a&quot;</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 5. Set 操作</span></span>
<span class="line">redis-cli sadd myset x y z</span>
<span class="line"><span class="token comment"># (integer) 3</span></span>
<span class="line"></span>
<span class="line">redis-cli smembers myset</span>
<span class="line"><span class="token comment"># 1) &quot;x&quot;</span></span>
<span class="line"><span class="token comment"># 2) &quot;y&quot;</span></span>
<span class="line"><span class="token comment"># 3) &quot;z&quot;</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 6. 事务操作</span></span>
<span class="line">redis-cli multi</span>
<span class="line"><span class="token comment"># OK</span></span>
<span class="line">redis-cli <span class="token builtin class-name">set</span> key1 value1</span>
<span class="line"><span class="token comment"># QUEUED</span></span>
<span class="line">redis-cli <span class="token builtin class-name">set</span> key2 value2</span>
<span class="line"><span class="token comment"># QUEUED</span></span>
<span class="line">redis-cli <span class="token builtin class-name">exec</span></span>
<span class="line"><span class="token comment"># 1) OK</span></span>
<span class="line"><span class="token comment"># 2) OK</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 7. 服务器信息</span></span>
<span class="line">redis-cli info server</span>
<span class="line"><span class="token comment"># # Server</span></span>
<span class="line"><span class="token comment"># goflashdb_version:0.3.0</span></span>
<span class="line"><span class="token comment"># goflashdb_mode:standalone</span></span>
<span class="line"><span class="token comment"># tcp_port:6379</span></span>
<span class="line"><span class="token comment"># uptime:3600</span></span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><h2 id="配置" tabindex="-1"><a class="header-anchor" href="#配置"><span>配置</span></a></h2><p>创建配置文件 <code>config.yaml</code>：</p><div class="language-yaml line-numbers-mode" data-highlighter="prismjs" data-ext="yml"><pre><code class="language-yaml"><span class="line"><span class="token comment"># 网络配置</span></span>
<span class="line"><span class="token key atrule">bind_addr</span><span class="token punctuation">:</span> <span class="token string">&quot;0.0.0.0:6379&quot;</span></span>
<span class="line"><span class="token key atrule">max_conn</span><span class="token punctuation">:</span> <span class="token number">10000</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 持久化配置</span></span>
<span class="line"><span class="token key atrule">append_only</span><span class="token punctuation">:</span> <span class="token boolean important">true</span></span>
<span class="line"><span class="token key atrule">append_filename</span><span class="token punctuation">:</span> <span class="token string">&quot;appendonly.aof&quot;</span></span>
<span class="line"><span class="token key atrule">append_fsync</span><span class="token punctuation">:</span> <span class="token string">&quot;everysec&quot;</span></span>
<span class="line"><span class="token key atrule">rdb_filename</span><span class="token punctuation">:</span> <span class="token string">&quot;dump.rdb&quot;</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 安全配置</span></span>
<span class="line"><span class="token key atrule">require_pass</span><span class="token punctuation">:</span> <span class="token string">&quot;your-password&quot;</span></span>
<span class="line"><span class="token key atrule">max_clients</span><span class="token punctuation">:</span> <span class="token number">10000</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 内存配置</span></span>
<span class="line"><span class="token key atrule">max_memory</span><span class="token punctuation">:</span> <span class="token number">0</span>  <span class="token comment"># 0 表示不限制</span></span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><p>启动时指定配置：</p><div class="language-bash line-numbers-mode" data-highlighter="prismjs" data-ext="sh"><pre><code class="language-bash"><span class="line">./goflashdb <span class="token parameter variable">-config</span> config.yaml</span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div></div></div><h2 id="性能测试" tabindex="-1"><a class="header-anchor" href="#性能测试"><span>性能测试</span></a></h2><p>使用内置的 benchmark 工具：</p><div class="language-bash line-numbers-mode" data-highlighter="prismjs" data-ext="sh"><pre><code class="language-bash"><span class="line"><span class="token comment"># 编译 benchmark</span></span>
<span class="line">go build <span class="token parameter variable">-o</span> goflashdb-bench ./cmd/benchmark</span>
<span class="line"></span>
<span class="line"><span class="token comment"># PING 测试</span></span>
<span class="line">./goflashdb-bench <span class="token parameter variable">-t</span> <span class="token function">ping</span> <span class="token parameter variable">-c</span> <span class="token number">50</span> <span class="token parameter variable">-n</span> <span class="token number">10000</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># SET 测试</span></span>
<span class="line">./goflashdb-bench <span class="token parameter variable">-t</span> <span class="token builtin class-name">set</span> <span class="token parameter variable">-c</span> <span class="token number">50</span> <span class="token parameter variable">-n</span> <span class="token number">10000</span> <span class="token parameter variable">-d</span> <span class="token number">10</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># GET 测试</span></span>
<span class="line">./goflashdb-bench <span class="token parameter variable">-t</span> get <span class="token parameter variable">-c</span> <span class="token number">50</span> <span class="token parameter variable">-n</span> <span class="token number">10000</span></span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><p>预期输出：</p><div class="language-text line-numbers-mode" data-highlighter="prismjs" data-ext="text"><pre><code class="language-text"><span class="line">Summary:</span>
<span class="line">  Requests completed: 10000</span>
<span class="line">  Requests failed: 0</span>
<span class="line">  Total duration: 0.08 seconds</span>
<span class="line">  QPS: 125000.00</span>
<span class="line">  Latency min: 0.01 ms</span>
<span class="line">  Latency max: 2.50 ms</span>
<span class="line">  Latency avg: 0.40 ms</span>
<span class="line">  Latency P50: 0.35 ms</span>
<span class="line">  Latency P95: 0.80 ms</span>
<span class="line">  Latency P99: 1.20 ms</span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><h2 id="与-redis-对比" tabindex="-1"><a class="header-anchor" href="#与-redis-对比"><span>与 Redis 对比</span></a></h2><table><thead><tr><th>特性</th><th>go-flashdb</th><th>Redis</th></tr></thead><tbody><tr><td>协议</td><td>RESP 2</td><td>RESP 2/3</td></tr><tr><td>数据类型</td><td>String/Hash/List/Set/ZSet/Bitmap/HLL</td><td>全支持</td></tr><tr><td>持久化</td><td>AOF + RDB</td><td>AOF + RDB</td></tr><tr><td>事务</td><td>WATCH/MULTI/EXEC</td><td>全支持</td></tr><tr><td>复制</td><td>基础实现</td><td>全功能</td></tr><tr><td>Cluster</td><td>协议兼容</td><td>完整实现</td></tr><tr><td>Lua</td><td>支持</td><td>全支持</td></tr></tbody></table><h2 id="项目结构" tabindex="-1"><a class="header-anchor" href="#项目结构"><span>项目结构</span></a></h2><div class="language-text line-numbers-mode" data-highlighter="prismjs" data-ext="text"><pre><code class="language-text"><span class="line">goflashdb/</span>
<span class="line">├── cmd/</span>
<span class="line">│   ├── goflashdb/      # 主程序入口</span>
<span class="line">│   └── benchmark/      # 性能测试工具</span>
<span class="line">├── pkg/</span>
<span class="line">│   ├── core/           # 核心：数据类型、命令、事务</span>
<span class="line">│   ├── resp/           # 协议层：RESP 解析器</span>
<span class="line">│   ├── persist/        # 持久化：AOF、RDB</span>
<span class="line">│   ├── net/            # 网络层：TCP 服务器</span>
<span class="line">│   ├── security/       # 安全：认证、限流、过滤</span>
<span class="line">│   ├── replication/    # 复制：主从同步</span>
<span class="line">│   ├── cluster/        # 集群：槽位管理</span>
<span class="line">│   ├── benchmark/      # 基准测试框架</span>
<span class="line">│   └── extension/      # 扩展：AI 接口预留</span>
<span class="line">└── docs/               # 文档</span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><h2 id="下一步" tabindex="-1"><a class="header-anchor" href="#下一步"><span>下一步</span></a></h2><ul><li><a href="/guide/02-tcp-server.html" target="_blank" rel="noopener noreferrer">TCP 服务器实现</a> - 了解 goroutine-per-connection 模型</li><li><a href="/guide/03-resp-protocol.html" target="_blank" rel="noopener noreferrer">RESP 协议解析</a> - 深入协议设计</li><li><a href="/design/" target="_blank" rel="noopener noreferrer">设计哲学</a> - 了解架构设计思路</li></ul><h2 id="常见问题" tabindex="-1"><a class="header-anchor" href="#常见问题"><span>常见问题</span></a></h2><h3 id="q-与-redis-的兼容性如何" tabindex="-1"><a class="header-anchor" href="#q-与-redis-的兼容性如何"><span>Q: 与 Redis 的兼容性如何？</span></a></h3><p>A: go-flashdb 实现了 RESP 2 协议，支持常用命令。可直接使用 redis-cli 或 Go-Redis 客户端连接。部分高级功能（如 Lua 脚本、Cluster 完整功能）正在开发中。</p><h3 id="q-性能如何" tabindex="-1"><a class="header-anchor" href="#q-性能如何"><span>Q: 性能如何？</span></a></h3><p>A: 本地测试 SET/GET 可达 10-15万 QPS。生产环境建议使用 Redis。</p><h3 id="q-可以替代-redis-吗" tabindex="-1"><a class="header-anchor" href="#q-可以替代-redis-吗"><span>Q: 可以替代 Redis 吗？</span></a></h3><p>A: 不建议。goflashdb 主要用于学习 Go 语言网络编程和 Redis 原理。生产环境请使用 Redis。</p><h3 id="q-如何贡献代码" tabindex="-1"><a class="header-anchor" href="#q-如何贡献代码"><span>Q: 如何贡献代码？</span></a></h3><p>A: 欢迎提交 Issue 和 PR！详见 <a href="https://github.com/strings77wzq/go-flashdb" target="_blank" rel="noopener noreferrer">GitHub</a></p><h2 id="参考" tabindex="-1"><a class="header-anchor" href="#参考"><span>参考</span></a></h2><ul><li><a href="https://github.com/strings77wzq/go-flashdb/blob/main/cmd/goflashdb/main.go" target="_blank" rel="noopener noreferrer">源码: cmd/goflashdb/main.go</a></li><li><a href="https://redis.io/topics/protocol" target="_blank" rel="noopener noreferrer">Redis 协议规范</a></li></ul>`,40)]])}var s=r(a,[[`render`,o]]);export{i as _pageData,s as default};