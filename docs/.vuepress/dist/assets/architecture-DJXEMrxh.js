import{A as e,d as t,m as n,t as r}from"./plugin-vue_export-helper--IMkQmEh.js";var i=JSON.parse(`{"path":"/architecture.html","title":"GoSwiftKV 架构设计","lang":"zh-CN","frontmatter":{},"git":{"updatedTime":1773129064000,"contributors":[{"name":"strings77wzq","username":"strings77wzq","email":"strings77wzq@github.com","commits":1,"url":"https://github.com/strings77wzq"}],"changelog":[{"hash":"8ffa3c030f4e1a89621269471e8a91cab0166582","time":1773129064000,"email":"strings77wzq@github.com","author":"strings77wzq","message":"feat: initial commit - Go high-performance Redis-compatible KV store"}]},"filePathRelative":"architecture.md"}`),a={name:`architecture.md`};function o(r,i,a,o,s,c){return e(),t(`div`,null,[...i[0]||=[n(`<h1 id="goswiftkv-架构设计" tabindex="-1"><a class="header-anchor" href="#goswiftkv-架构设计"><span>GoSwiftKV 架构设计</span></a></h1><h2 id="分层架构" tabindex="-1"><a class="header-anchor" href="#分层架构"><span>分层架构</span></a></h2><div class="language-text line-numbers-mode" data-highlighter="prismjs" data-ext="text"><pre><code class="language-text"><span class="line">┌─────────────────────────────────────────────────────┐</span>
<span class="line">│                 Application Layer                    │</span>
<span class="line">│                  cmd/goswiftkv/                      │</span>
<span class="line">└─────────────────────────────────────────────────────┘</span>
<span class="line">                          │</span>
<span class="line">┌─────────────────────────────────────────────────────┐</span>
<span class="line">│                 Extension Layer                      │</span>
<span class="line">│              pkg/extension/                          │</span>
<span class="line">│         OpenClaw / Skills / MCP / Plugins           │</span>
<span class="line">└─────────────────────────────────────────────────────┘</span>
<span class="line">                          │</span>
<span class="line">┌─────────────────────────────────────────────────────┐</span>
<span class="line">│                 Security Layer                       │</span>
<span class="line">│              pkg/security/                           │</span>
<span class="line">│        Authenticator / RateLimiter / Filter         │</span>
<span class="line">└─────────────────────────────────────────────────────┘</span>
<span class="line">                          │</span>
<span class="line">┌─────────────────────────────────────────────────────┐</span>
<span class="line">│                 Network Layer                        │</span>
<span class="line">│              pkg/net/                                │</span>
<span class="line">│            TCP Server / Connection                   │</span>
<span class="line">└─────────────────────────────────────────────────────┘</span>
<span class="line">                          │</span>
<span class="line">┌─────────────────────────────────────────────────────┐</span>
<span class="line">│                 Protocol Layer                       │</span>
<span class="line">│              pkg/resp/                               │</span>
<span class="line">│          RESP Parser / Serializer                    │</span>
<span class="line">└─────────────────────────────────────────────────────┘</span>
<span class="line">                          │</span>
<span class="line">┌─────────────────────────────────────────────────────┐</span>
<span class="line">│                 Core Layer                           │</span>
<span class="line">│              pkg/core/                               │</span>
<span class="line">│     DB / ConcurrentDict / Commands / Types          │</span>
<span class="line">└─────────────────────────────────────────────────────┘</span>
<span class="line">                          │</span>
<span class="line">┌─────────────────────────────────────────────────────┐</span>
<span class="line">│                 Persistence Layer                    │</span>
<span class="line">│              pkg/persist/                            │</span>
<span class="line">│             AOF / RDB                                │</span>
<span class="line">└─────────────────────────────────────────────────────┘</span>
<span class="line">                          │</span>
<span class="line">┌─────────────────────────────────────────────────────┐</span>
<span class="line">│                 Infrastructure Layer                 │</span>
<span class="line">│              pkg/lib/                                │</span>
<span class="line">│          Logger / Utils / Config                     │</span>
<span class="line">└─────────────────────────────────────────────────────┘</span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><h2 id="核心模块" tabindex="-1"><a class="header-anchor" href="#核心模块"><span>核心模块</span></a></h2><h3 id="_1-存储引擎-pkg-core" tabindex="-1"><a class="header-anchor" href="#_1-存储引擎-pkg-core"><span>1. 存储引擎 (pkg/core)</span></a></h3><p><strong>ConcurrentDict</strong>: 65536 分片并发字典</p><div class="language-go line-numbers-mode" data-highlighter="prismjs" data-ext="go"><pre><code class="language-go"><span class="line"><span class="token keyword">type</span> ConcurrentDict <span class="token keyword">struct</span> <span class="token punctuation">{</span></span>
<span class="line">    segments <span class="token punctuation">[</span><span class="token punctuation">]</span><span class="token operator">*</span>Segment  <span class="token comment">// 65536 个独立分片</span></span>
<span class="line">    count    <span class="token builtin">int</span></span>
<span class="line"><span class="token punctuation">}</span></span>
<span class="line"></span>
<span class="line"><span class="token keyword">type</span> Segment <span class="token keyword">struct</span> <span class="token punctuation">{</span></span>
<span class="line">    m  <span class="token keyword">map</span><span class="token punctuation">[</span><span class="token builtin">string</span><span class="token punctuation">]</span><span class="token keyword">interface</span><span class="token punctuation">{</span><span class="token punctuation">}</span></span>
<span class="line">    mu sync<span class="token punctuation">.</span>RWMutex  <span class="token comment">// 每个分片独立锁</span></span>
<span class="line"><span class="token punctuation">}</span></span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><p><strong>优势</strong>:</p><ul><li>锁冲突概率 &lt; 0.001%</li><li>读操作完全并发</li><li>写操作仅锁定相关分片</li></ul><h3 id="_2-协议层-pkg-resp" tabindex="-1"><a class="header-anchor" href="#_2-协议层-pkg-resp"><span>2. 协议层 (pkg/resp)</span></a></h3><p><strong>RESP 协议支持</strong>:</p><ul><li>Simple String: <code>+OK\\r\\n</code></li><li>Error: <code>-ERR message\\r\\n</code></li><li>Integer: <code>:1\\r\\n</code></li><li>Bulk String: <code>$6\\r\\nfoobar\\r\\n</code></li><li>Array: <code>*2\\r\\n$3\\r\\nfoo\\r\\n$3\\r\\nbar\\r\\n</code></li></ul><h3 id="_3-持久化-pkg-persist" tabindex="-1"><a class="header-anchor" href="#_3-持久化-pkg-persist"><span>3. 持久化 (pkg/persist)</span></a></h3><p><strong>AOF (Append Only File)</strong>:</p><ul><li>三种同步策略: Always / Everysec / No</li><li>异步写入，不阻塞主线程</li><li>支持 AOF Rewrite 压缩</li></ul><p><strong>RDB (Redis Database)</strong>:</p><ul><li>二进制格式快照</li><li>支持过期时间</li><li>快速加载</li></ul><h3 id="_4-安全模块-pkg-security" tabindex="-1"><a class="header-anchor" href="#_4-安全模块-pkg-security"><span>4. 安全模块 (pkg/security)</span></a></h3><p><strong>认证</strong>: 密码 + Session 管理 <strong>限流</strong>: 滑动窗口算法 <strong>命令过滤</strong>: 危险命令拦截 + 重命名</p><h3 id="_5-ai-扩展-pkg-extension" tabindex="-1"><a class="header-anchor" href="#_5-ai-扩展-pkg-extension"><span>5. AI 扩展 (pkg/extension)</span></a></h3><p><strong>OpenClaw</strong>: AI 助手直接操作接口 <strong>Skills</strong>: 自定义命令扩展 <strong>MCP</strong>: 模型上下文协议 <strong>Plugins</strong>: 插件系统</p><h2 id="并发模型" tabindex="-1"><a class="header-anchor" href="#并发模型"><span>并发模型</span></a></h2><div class="language-text line-numbers-mode" data-highlighter="prismjs" data-ext="text"><pre><code class="language-text"><span class="line">                    ┌─────────────┐</span>
<span class="line">                    │  Listener   │</span>
<span class="line">                    └──────┬──────┘</span>
<span class="line">                           │ Accept</span>
<span class="line">              ┌────────────┼────────────┐</span>
<span class="line">              │            │            │</span>
<span class="line">        ┌─────▼─────┐ ┌────▼────┐ ┌────▼────┐</span>
<span class="line">        │ Goroutine │ │Goroutine│ │Goroutine│</span>
<span class="line">        │  Conn 1   │ │ Conn 2  │ │ Conn N  │</span>
<span class="line">        └─────┬─────┘ └────┬────┘ └────┬────┘</span>
<span class="line">              │            │            │</span>
<span class="line">              └────────────┼────────────┘</span>
<span class="line">                           │</span>
<span class="line">                    ┌──────▼──────┐</span>
<span class="line">                    │ Concurrent  │</span>
<span class="line">                    │    Dict     │</span>
<span class="line">                    │ (65536 shards)</span>
<span class="line">                    └─────────────┘</span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><h2 id="性能优化" tabindex="-1"><a class="header-anchor" href="#性能优化"><span>性能优化</span></a></h2><h3 id="_1-锁优化" tabindex="-1"><a class="header-anchor" href="#_1-锁优化"><span>1. 锁优化</span></a></h3><ul><li>分片锁替代全局锁</li><li>读写分离 (RLock/RWMutex)</li><li>短临界区</li></ul><h3 id="_2-内存优化" tabindex="-1"><a class="header-anchor" href="#_2-内存优化"><span>2. 内存优化</span></a></h3><ul><li>[]byte 避免字符串转换</li><li>sync.Pool 复用对象</li><li>预分配缓冲区</li></ul><h3 id="_3-io-优化" tabindex="-1"><a class="header-anchor" href="#_3-io-优化"><span>3. IO 优化</span></a></h3><ul><li>bufio 缓冲读写</li><li>批量写入 AOF</li><li>异步 fsync</li></ul><h2 id="数据流" tabindex="-1"><a class="header-anchor" href="#数据流"><span>数据流</span></a></h2><div class="language-text line-numbers-mode" data-highlighter="prismjs" data-ext="text"><pre><code class="language-text"><span class="line">Client Request</span>
<span class="line">      │</span>
<span class="line">      ▼</span>
<span class="line">┌─────────────┐</span>
<span class="line">│ TCP Server  │</span>
<span class="line">└──────┬──────┘</span>
<span class="line">       │</span>
<span class="line">       ▼</span>
<span class="line">┌─────────────┐</span>
<span class="line">│ RESP Parser │</span>
<span class="line">└──────┬──────┘</span>
<span class="line">       │</span>
<span class="line">       ▼</span>
<span class="line">┌─────────────┐</span>
<span class="line">│   Router    │</span>
<span class="line">└──────┬──────┘</span>
<span class="line">       │</span>
<span class="line">       ▼</span>
<span class="line">┌─────────────┐     ┌─────────────┐</span>
<span class="line">│   Command   │────▶│    AOF      │</span>
<span class="line">│  Executor   │     └─────────────┘</span>
<span class="line">└──────┬──────┘</span>
<span class="line">       │</span>
<span class="line">       ▼</span>
<span class="line">┌─────────────┐</span>
<span class="line">│ Concurrent  │</span>
<span class="line">│    Dict     │</span>
<span class="line">└──────┬──────┘</span>
<span class="line">       │</span>
<span class="line">       ▼</span>
<span class="line">┌─────────────┐</span>
<span class="line">│  Response   │</span>
<span class="line">└─────────────┘</span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><h2 id="扩展点" tabindex="-1"><a class="header-anchor" href="#扩展点"><span>扩展点</span></a></h2><h3 id="添加新命令" tabindex="-1"><a class="header-anchor" href="#添加新命令"><span>添加新命令</span></a></h3><div class="language-go line-numbers-mode" data-highlighter="prismjs" data-ext="go"><pre><code class="language-go"><span class="line"><span class="token keyword">func</span> <span class="token function">init</span><span class="token punctuation">(</span><span class="token punctuation">)</span> <span class="token punctuation">{</span></span>
<span class="line">    <span class="token function">RegisterCommand</span><span class="token punctuation">(</span><span class="token string">&quot;mycommand&quot;</span><span class="token punctuation">,</span> execMyCommand<span class="token punctuation">,</span> prepareMyCommand<span class="token punctuation">,</span> <span class="token number">2</span><span class="token punctuation">)</span></span>
<span class="line"><span class="token punctuation">}</span></span>
<span class="line"></span>
<span class="line"><span class="token keyword">func</span> <span class="token function">execMyCommand</span><span class="token punctuation">(</span>db <span class="token operator">*</span>DB<span class="token punctuation">,</span> args <span class="token punctuation">[</span><span class="token punctuation">]</span><span class="token punctuation">[</span><span class="token punctuation">]</span><span class="token builtin">byte</span><span class="token punctuation">)</span> resp<span class="token punctuation">.</span>Reply <span class="token punctuation">{</span></span>
<span class="line">    <span class="token comment">// 实现逻辑</span></span>
<span class="line">    <span class="token keyword">return</span> resp<span class="token punctuation">.</span>OkReply</span>
<span class="line"><span class="token punctuation">}</span></span>
<span class="line"></span>
<span class="line"><span class="token keyword">func</span> <span class="token function">prepareMyCommand</span><span class="token punctuation">(</span>args <span class="token punctuation">[</span><span class="token punctuation">]</span><span class="token punctuation">[</span><span class="token punctuation">]</span><span class="token builtin">byte</span><span class="token punctuation">)</span> <span class="token punctuation">(</span>write<span class="token punctuation">,</span> read <span class="token punctuation">[</span><span class="token punctuation">]</span><span class="token builtin">string</span><span class="token punctuation">)</span> <span class="token punctuation">{</span></span>
<span class="line">    <span class="token keyword">return</span> <span class="token punctuation">[</span><span class="token punctuation">]</span><span class="token builtin">string</span><span class="token punctuation">{</span><span class="token function">string</span><span class="token punctuation">(</span>args<span class="token punctuation">[</span><span class="token number">0</span><span class="token punctuation">]</span><span class="token punctuation">)</span><span class="token punctuation">}</span><span class="token punctuation">,</span> <span class="token boolean">nil</span></span>
<span class="line"><span class="token punctuation">}</span></span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><h3 id="添加新数据类型" tabindex="-1"><a class="header-anchor" href="#添加新数据类型"><span>添加新数据类型</span></a></h3><div class="language-go line-numbers-mode" data-highlighter="prismjs" data-ext="go"><pre><code class="language-go"><span class="line"><span class="token keyword">type</span> MyTypeData <span class="token keyword">struct</span> <span class="token punctuation">{</span></span>
<span class="line">    data     <span class="token keyword">interface</span><span class="token punctuation">{</span><span class="token punctuation">}</span></span>
<span class="line">    expireAt <span class="token builtin">int64</span></span>
<span class="line"><span class="token punctuation">}</span></span>
<span class="line"></span>
<span class="line"><span class="token keyword">func</span> <span class="token punctuation">(</span>db <span class="token operator">*</span>DB<span class="token punctuation">)</span> <span class="token function">GetMyTypeData</span><span class="token punctuation">(</span>key <span class="token builtin">string</span><span class="token punctuation">)</span> <span class="token punctuation">(</span><span class="token operator">*</span>MyTypeData<span class="token punctuation">,</span> <span class="token builtin">bool</span><span class="token punctuation">)</span> <span class="token punctuation">{</span></span>
<span class="line">    <span class="token comment">// 实现</span></span>
<span class="line"><span class="token punctuation">}</span></span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><h3 id="添加插件" tabindex="-1"><a class="header-anchor" href="#添加插件"><span>添加插件</span></a></h3><div class="language-go line-numbers-mode" data-highlighter="prismjs" data-ext="go"><pre><code class="language-go"><span class="line"><span class="token keyword">type</span> MyPlugin <span class="token keyword">struct</span><span class="token punctuation">{</span><span class="token punctuation">}</span></span>
<span class="line"></span>
<span class="line"><span class="token keyword">func</span> <span class="token punctuation">(</span>p <span class="token operator">*</span>MyPlugin<span class="token punctuation">)</span> <span class="token function">Init</span><span class="token punctuation">(</span>config <span class="token keyword">map</span><span class="token punctuation">[</span><span class="token builtin">string</span><span class="token punctuation">]</span><span class="token keyword">interface</span><span class="token punctuation">{</span><span class="token punctuation">}</span><span class="token punctuation">)</span> <span class="token builtin">error</span> <span class="token punctuation">{</span> <span class="token keyword">return</span> <span class="token boolean">nil</span> <span class="token punctuation">}</span></span>
<span class="line"><span class="token keyword">func</span> <span class="token punctuation">(</span>p <span class="token operator">*</span>MyPlugin<span class="token punctuation">)</span> <span class="token function">Start</span><span class="token punctuation">(</span><span class="token punctuation">)</span> <span class="token builtin">error</span> <span class="token punctuation">{</span> <span class="token keyword">return</span> <span class="token boolean">nil</span> <span class="token punctuation">}</span></span>
<span class="line"><span class="token keyword">func</span> <span class="token punctuation">(</span>p <span class="token operator">*</span>MyPlugin<span class="token punctuation">)</span> <span class="token function">Stop</span><span class="token punctuation">(</span><span class="token punctuation">)</span> <span class="token builtin">error</span> <span class="token punctuation">{</span> <span class="token keyword">return</span> <span class="token boolean">nil</span> <span class="token punctuation">}</span></span>
<span class="line"><span class="token keyword">func</span> <span class="token punctuation">(</span>p <span class="token operator">*</span>MyPlugin<span class="token punctuation">)</span> <span class="token function">Info</span><span class="token punctuation">(</span><span class="token punctuation">)</span> PluginInfo <span class="token punctuation">{</span></span>
<span class="line">    <span class="token keyword">return</span> PluginInfo<span class="token punctuation">{</span>Name<span class="token punctuation">:</span> <span class="token string">&quot;my-plugin&quot;</span><span class="token punctuation">,</span> Version<span class="token punctuation">:</span> <span class="token string">&quot;1.0&quot;</span><span class="token punctuation">}</span></span>
<span class="line"><span class="token punctuation">}</span></span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div>`,39)]])}var s=r(a,[[`render`,o]]);export{i as _pageData,s as default};