import{A as e,d as t,m as n,t as r}from"./plugin-vue_export-helper--IMkQmEh.js";var i=JSON.parse(`{"path":"/guide/deploy.html","title":"GitHub Actions 自动部署","lang":"zh-CN","frontmatter":{},"git":{"updatedTime":1775618263000,"contributors":[{"name":"strings77wzq","username":"strings77wzq","email":"strings77wzq@github.com","commits":1,"url":"https://github.com/strings77wzq"}],"changelog":[{"hash":"3c536800d4175146327c71649c552297dd8add0c","time":1775618263000,"email":"strings77wzq@github.com","author":"strings77wzq","message":"feat: implement v0.3.0 features and documentation website","tag":"v0.3.0"}]},"filePathRelative":"guide/deploy.md"}`),a={name:`deploy.md`};function o(r,i,a,o,s,c){return e(),t(`div`,null,[...i[0]||=[n(`<h1 id="github-actions-自动部署" tabindex="-1"><a class="header-anchor" href="#github-actions-自动部署"><span>GitHub Actions 自动部署</span></a></h1><p>配置 GitHub Actions，实现文档自动构建和部署到 GitHub Pages。</p><h2 id="配置步骤" tabindex="-1"><a class="header-anchor" href="#配置步骤"><span>配置步骤</span></a></h2><h3 id="_1-创建工作流文件" tabindex="-1"><a class="header-anchor" href="#_1-创建工作流文件"><span>1. 创建工作流文件</span></a></h3><div class="language-bash line-numbers-mode" data-highlighter="prismjs" data-ext="sh"><pre><code class="language-bash"><span class="line"><span class="token function">mkdir</span> <span class="token parameter variable">-p</span> .github/workflows</span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div></div></div><h3 id="_2-创建-deploy-docs-yml" tabindex="-1"><a class="header-anchor" href="#_2-创建-deploy-docs-yml"><span>2. 创建 deploy-docs.yml</span></a></h3><div class="language-yaml line-numbers-mode" data-highlighter="prismjs" data-ext="yml"><pre><code class="language-yaml"><span class="line"><span class="token key atrule">name</span><span class="token punctuation">:</span> Deploy Docs to GitHub Pages</span>
<span class="line"></span>
<span class="line"><span class="token key atrule">on</span><span class="token punctuation">:</span></span>
<span class="line">  <span class="token key atrule">push</span><span class="token punctuation">:</span></span>
<span class="line">    <span class="token key atrule">branches</span><span class="token punctuation">:</span></span>
<span class="line">      <span class="token punctuation">-</span> main</span>
<span class="line">    <span class="token key atrule">paths</span><span class="token punctuation">:</span></span>
<span class="line">      <span class="token punctuation">-</span> <span class="token string">&#39;docs/**&#39;</span></span>
<span class="line">      <span class="token punctuation">-</span> <span class="token string">&#39;package.json&#39;</span></span>
<span class="line">      <span class="token punctuation">-</span> <span class="token string">&#39;.github/workflows/docs.yml&#39;</span></span>
<span class="line">  </span>
<span class="line">  <span class="token key atrule">workflow_dispatch</span><span class="token punctuation">:</span></span>
<span class="line"></span>
<span class="line"><span class="token key atrule">permissions</span><span class="token punctuation">:</span></span>
<span class="line">  <span class="token key atrule">contents</span><span class="token punctuation">:</span> read</span>
<span class="line">  <span class="token key atrule">pages</span><span class="token punctuation">:</span> write</span>
<span class="line">  <span class="token key atrule">id-token</span><span class="token punctuation">:</span> write</span>
<span class="line"></span>
<span class="line"><span class="token key atrule">concurrency</span><span class="token punctuation">:</span></span>
<span class="line">  <span class="token key atrule">group</span><span class="token punctuation">:</span> pages</span>
<span class="line">  <span class="token key atrule">cancel-in-progress</span><span class="token punctuation">:</span> <span class="token boolean important">false</span></span>
<span class="line"></span>
<span class="line"><span class="token key atrule">jobs</span><span class="token punctuation">:</span></span>
<span class="line">  <span class="token key atrule">build</span><span class="token punctuation">:</span></span>
<span class="line">    <span class="token key atrule">runs-on</span><span class="token punctuation">:</span> ubuntu<span class="token punctuation">-</span>latest</span>
<span class="line">    <span class="token key atrule">steps</span><span class="token punctuation">:</span></span>
<span class="line">      <span class="token punctuation">-</span> <span class="token key atrule">name</span><span class="token punctuation">:</span> Checkout</span>
<span class="line">        <span class="token key atrule">uses</span><span class="token punctuation">:</span> actions/checkout@v4</span>
<span class="line">        </span>
<span class="line">      <span class="token punctuation">-</span> <span class="token key atrule">name</span><span class="token punctuation">:</span> Setup Node</span>
<span class="line">        <span class="token key atrule">uses</span><span class="token punctuation">:</span> actions/setup<span class="token punctuation">-</span>node@v4</span>
<span class="line">        <span class="token key atrule">with</span><span class="token punctuation">:</span></span>
<span class="line">          <span class="token key atrule">node-version</span><span class="token punctuation">:</span> <span class="token number">20</span></span>
<span class="line">          <span class="token key atrule">cache</span><span class="token punctuation">:</span> npm</span>
<span class="line">          </span>
<span class="line">      <span class="token punctuation">-</span> <span class="token key atrule">name</span><span class="token punctuation">:</span> Install dependencies</span>
<span class="line">        <span class="token key atrule">run</span><span class="token punctuation">:</span> npm ci</span>
<span class="line">        </span>
<span class="line">      <span class="token punctuation">-</span> <span class="token key atrule">name</span><span class="token punctuation">:</span> Build docs</span>
<span class="line">        <span class="token key atrule">run</span><span class="token punctuation">:</span> npm run docs<span class="token punctuation">:</span>build</span>
<span class="line">        </span>
<span class="line">      <span class="token punctuation">-</span> <span class="token key atrule">name</span><span class="token punctuation">:</span> Upload artifact</span>
<span class="line">        <span class="token key atrule">uses</span><span class="token punctuation">:</span> actions/upload<span class="token punctuation">-</span>pages<span class="token punctuation">-</span>artifact@v3</span>
<span class="line">        <span class="token key atrule">with</span><span class="token punctuation">:</span></span>
<span class="line">          <span class="token key atrule">path</span><span class="token punctuation">:</span> docs/.vuepress/dist</span>
<span class="line"></span>
<span class="line">  <span class="token key atrule">deploy</span><span class="token punctuation">:</span></span>
<span class="line">    <span class="token key atrule">environment</span><span class="token punctuation">:</span></span>
<span class="line">      <span class="token key atrule">name</span><span class="token punctuation">:</span> github<span class="token punctuation">-</span>pages</span>
<span class="line">      <span class="token key atrule">url</span><span class="token punctuation">:</span> $<span class="token punctuation">{</span><span class="token punctuation">{</span> steps.deployment.outputs.page_url <span class="token punctuation">}</span><span class="token punctuation">}</span></span>
<span class="line">    <span class="token key atrule">runs-on</span><span class="token punctuation">:</span> ubuntu<span class="token punctuation">-</span>latest</span>
<span class="line">    <span class="token key atrule">needs</span><span class="token punctuation">:</span> build</span>
<span class="line">    <span class="token key atrule">steps</span><span class="token punctuation">:</span></span>
<span class="line">      <span class="token punctuation">-</span> <span class="token key atrule">name</span><span class="token punctuation">:</span> Deploy to GitHub Pages</span>
<span class="line">        <span class="token key atrule">id</span><span class="token punctuation">:</span> deployment</span>
<span class="line">        <span class="token key atrule">uses</span><span class="token punctuation">:</span> actions/deploy<span class="token punctuation">-</span>pages@v4</span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><h3 id="_3-启用-github-pages" tabindex="-1"><a class="header-anchor" href="#_3-启用-github-pages"><span>3. 启用 GitHub Pages</span></a></h3><ol><li>进入 GitHub 仓库 Settings → Pages</li><li>Source 选择 &quot;GitHub Actions&quot;</li><li>保存</li></ol><h3 id="_4-部署完成" tabindex="-1"><a class="header-anchor" href="#_4-部署完成"><span>4. 部署完成</span></a></h3><p>访问：<code>https://strings77wzq.github.io/go-flashdb/</code></p><h2 id="本地预览" tabindex="-1"><a class="header-anchor" href="#本地预览"><span>本地预览</span></a></h2><div class="language-bash line-numbers-mode" data-highlighter="prismjs" data-ext="sh"><pre><code class="language-bash"><span class="line"><span class="token comment"># 安装依赖</span></span>
<span class="line"><span class="token function">npm</span> <span class="token function">install</span></span>
<span class="line"></span>
<span class="line"><span class="token comment"># 开发模式</span></span>
<span class="line"><span class="token function">npm</span> run docs:dev</span>
<span class="line"></span>
<span class="line"><span class="token comment"># 构建</span></span>
<span class="line"><span class="token function">npm</span> run docs:build</span>
<span class="line"></span></code></pre><div class="line-numbers" aria-hidden="true" style="counter-reset:line-number 0;"><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div><div class="line-number"></div></div></div><h2 id="自定义域名-可选" tabindex="-1"><a class="header-anchor" href="#自定义域名-可选"><span>自定义域名（可选）</span></a></h2><ol><li>添加 CNAME 文件到 docs/.vuepress/public/</li><li>在 DNS 配置 CNAME 记录</li><li>在 GitHub Pages 设置自定义域名</li></ol>`,15)]])}var s=r(a,[[`render`,o]]);export{i as _pageData,s as default};