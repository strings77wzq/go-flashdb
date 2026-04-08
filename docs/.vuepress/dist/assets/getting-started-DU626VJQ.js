import{A as e,H as t,P as n,c as r,d as i,g as a,h as o,m as s,t as c}from"./plugin-vue_export-helper--IMkQmEh.js";var l=JSON.parse(`{"path":"/getting-started.html","title":"GoSwiftKV 入门指南","lang":"zh-CN","frontmatter":{},"git":{"updatedTime":1773129064000,"contributors":[{"name":"strings77wzq","username":"strings77wzq","email":"strings77wzq@github.com","commits":1,"url":"https://github.com/strings77wzq"}],"changelog":[{"hash":"8ffa3c030f4e1a89621269471e8a91cab0166582","time":1773129064000,"email":"strings77wzq@github.com","author":"strings77wzq","message":"feat: initial commit - Go high-performance Redis-compatible KV store"}]},"filePathRelative":"getting-started.md"}`),u={name:`getting-started.md`};function d(c,l,u,d,f,p){let m=n(`RouteLink`);return e(),i(`div`,null,[l[3]||=s(`<h1 id="goswiftkv-入门指南" tabindex="-1"><a class="header-anchor" href="#goswiftkv-入门指南"><span>GoSwiftKV 入门指南</span></a></h1><h2 id="快速开始" tabindex="-1"><a class="header-anchor" href="#快速开始"><span>快速开始</span></a></h2><h3 id="安装" tabindex="-1"><a class="header-anchor" href="#安装"><span>安装</span></a></h3><div class="language-bash line-numbers-mode" data-highlighter="prismjs" data-ext="sh"><pre><code class="language-bash"><span class="line">go <span class="token function">install</span> github.com/goswiftkv/goswiftkv@latest</span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div></div></div><h3 id="启动服务器" tabindex="-1"><a class="header-anchor" href="#启动服务器"><span>启动服务器</span></a></h3><div class="language-bash line-numbers-mode" data-highlighter="prismjs" data-ext="sh"><pre><code class="language-bash"><span class="line">goswiftkv</span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div></div></div><p>默认监听 <code>0.0.0.0:6379</code>，可直接使用 <code>redis-cli</code> 连接：</p><div class="language-bash line-numbers-mode" data-highlighter="prismjs" data-ext="sh"><pre><code class="language-bash"><span class="line">redis-cli <span class="token function">ping</span></span>
<span class="line"><span class="token comment"># PONG</span></span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div></div></div><h2 id="基本使用" tabindex="-1"><a class="header-anchor" href="#基本使用"><span>基本使用</span></a></h2><h3 id="字符串操作" tabindex="-1"><a class="header-anchor" href="#字符串操作"><span>字符串操作</span></a></h3><div class="language-bash line-numbers-mode" data-highlighter="prismjs" data-ext="sh"><pre><code class="language-bash"><span class="line"><span class="token comment"># 设置值</span></span>
<span class="line">redis-cli <span class="token builtin class-name">set</span> mykey <span class="token string">&quot;Hello GoSwiftKV&quot;</span></span>
<span class="line"><span class="token comment"># OK</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 获取值</span></span>
<span class="line">redis-cli get mykey</span>
<span class="line"><span class="token comment"># &quot;Hello GoSwiftKV&quot;</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 设置过期时间（秒）</span></span>
<span class="line">redis-cli setex mykey <span class="token number">60</span> <span class="token string">&quot;expires in 60s&quot;</span></span>
<span class="line"><span class="token comment"># OK</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 批量设置</span></span>
<span class="line">redis-cli mset key1 value1 key2 value2</span>
<span class="line"><span class="token comment"># OK</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 批量获取</span></span>
<span class="line">redis-cli mget key1 key2</span>
<span class="line"><span class="token comment"># 1) &quot;value1&quot;</span></span>
<span class="line"><span class="token comment"># 2) &quot;value2&quot;</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 数值操作</span></span>
<span class="line">redis-cli <span class="token builtin class-name">set</span> counter <span class="token number">10</span></span>
<span class="line">redis-cli incr counter</span>
<span class="line"><span class="token comment"># (integer) 11</span></span>
<span class="line">redis-cli incrby counter <span class="token number">5</span></span>
<span class="line"><span class="token comment"># (integer) 16</span></span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><h3 id="键管理" tabindex="-1"><a class="header-anchor" href="#键管理"><span>键管理</span></a></h3><div class="language-bash line-numbers-mode" data-highlighter="prismjs" data-ext="sh"><pre><code class="language-bash"><span class="line"><span class="token comment"># 检查键是否存在</span></span>
<span class="line">redis-cli exists mykey</span>
<span class="line"><span class="token comment"># (integer) 1</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 删除键</span></span>
<span class="line">redis-cli del mykey</span>
<span class="line"><span class="token comment"># (integer) 1</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 查看键类型</span></span>
<span class="line">redis-cli <span class="token builtin class-name">type</span> mykey</span>
<span class="line"><span class="token comment"># string</span></span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><h2 id="配置" tabindex="-1"><a class="header-anchor" href="#配置"><span>配置</span></a></h2><h3 id="配置文件" tabindex="-1"><a class="header-anchor" href="#配置文件"><span>配置文件</span></a></h3><p>创建 <code>config.yaml</code>:</p><div class="language-yaml line-numbers-mode" data-highlighter="prismjs" data-ext="yml"><pre><code class="language-yaml"><span class="line"><span class="token comment"># 网络配置</span></span>
<span class="line"><span class="token key atrule">bind_addr</span><span class="token punctuation">:</span> <span class="token string">&quot;0.0.0.0:6379&quot;</span></span>
<span class="line"><span class="token key atrule">max_conn</span><span class="token punctuation">:</span> <span class="token number">10000</span></span>
<span class="line"><span class="token key atrule">timeout</span><span class="token punctuation">:</span> <span class="token number">300</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 内存配置</span></span>
<span class="line"><span class="token key atrule">max_memory</span><span class="token punctuation">:</span> <span class="token number">0</span>  <span class="token comment"># 0 表示不限制</span></span>
<span class="line"><span class="token key atrule">eviction_policy</span><span class="token punctuation">:</span> <span class="token string">&quot;volatile-lru&quot;</span></span>
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
<span class="line"><span class="token comment"># AI 扩展配置</span></span>
<span class="line"><span class="token key atrule">enable_ai</span><span class="token punctuation">:</span> <span class="token boolean important">false</span></span>
<span class="line"><span class="token key atrule">openclaw_endpoint</span><span class="token punctuation">:</span> <span class="token string">&quot;&quot;</span></span>
<span class="line"><span class="token key atrule">mcp_server_addr</span><span class="token punctuation">:</span> <span class="token string">&quot;&quot;</span></span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><h3 id="启动时指定配置" tabindex="-1"><a class="header-anchor" href="#启动时指定配置"><span>启动时指定配置</span></a></h3><div class="language-bash line-numbers-mode" data-highlighter="prismjs" data-ext="sh"><pre><code class="language-bash"><span class="line">goswiftkv <span class="token parameter variable">-c</span> /path/to/config.yaml</span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div></div></div><h2 id="持久化" tabindex="-1"><a class="header-anchor" href="#持久化"><span>持久化</span></a></h2><h3 id="rdb-快照" tabindex="-1"><a class="header-anchor" href="#rdb-快照"><span>RDB 快照</span></a></h3><div class="language-bash line-numbers-mode" data-highlighter="prismjs" data-ext="sh"><pre><code class="language-bash"><span class="line"><span class="token comment"># 手动触发快照</span></span>
<span class="line">redis-cli bgsave</span>
<span class="line"><span class="token comment"># Background saving started</span></span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><h3 id="aof-日志" tabindex="-1"><a class="header-anchor" href="#aof-日志"><span>AOF 日志</span></a></h3><p>在配置文件中启用：</p><div class="language-yaml line-numbers-mode" data-highlighter="prismjs" data-ext="yml"><pre><code class="language-yaml"><span class="line"><span class="token key atrule">append_only</span><span class="token punctuation">:</span> <span class="token boolean important">true</span></span>
<span class="line"><span class="token key atrule">append_fsync</span><span class="token punctuation">:</span> <span class="token string">&quot;everysec&quot;</span>  <span class="token comment"># always / everysec / no</span></span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div></div></div><h2 id="性能优化建议" tabindex="-1"><a class="header-anchor" href="#性能优化建议"><span>性能优化建议</span></a></h2><h3 id="_1-内存优化" tabindex="-1"><a class="header-anchor" href="#_1-内存优化"><span>1. 内存优化</span></a></h3><ul><li>合理设置 <code>max_memory</code></li><li>选择合适的淘汰策略</li><li>避免大 Key（&gt; 10KB）</li></ul><h3 id="_2-并发优化" tabindex="-1"><a class="header-anchor" href="#_2-并发优化"><span>2. 并发优化</span></a></h3><ul><li>使用连接池</li><li>批量操作使用 MSET/MGET</li><li>避免阻塞命令</li></ul><h3 id="_3-持久化优化" tabindex="-1"><a class="header-anchor" href="#_3-持久化优化"><span>3. 持久化优化</span></a></h3><ul><li>生产环境建议 AOF + RDB 混合</li><li><code>append_fsync</code> 选择 <code>everysec</code></li><li>定期执行 BGREWRITEAOF</li></ul><h2 id="下一步" tabindex="-1"><a class="header-anchor" href="#下一步"><span>下一步</span></a></h2>`,33),r(`ul`,null,[r(`li`,null,[a(m,{to:`/api-reference.html`},{default:t(()=>[...l[0]||=[o(`API 参考`,-1)]]),_:1})]),r(`li`,null,[a(m,{to:`/architecture.html`},{default:t(()=>[...l[1]||=[o(`架构设计`,-1)]]),_:1})]),r(`li`,null,[a(m,{to:`/deployment.html`},{default:t(()=>[...l[2]||=[o(`部署指南`,-1)]]),_:1})])])])}var f=c(u,[[`render`,d]]);export{l as _pageData,f as default};