# 首页重设计 (Homepage Redesign)

**日期**: 2026-04-28
**作者**: meitianwang + Claude
**状态**: Approved (待实现)

## 背景

现有首页 ([`frontend/src/views/HomeView.vue`](../../../frontend/src/views/HomeView.vue), 855 行) 视觉上"AI 感"很强 —— 紫色光晕 orbs、点阵背景、glassmorphism 头部、伪终端 perspective 3D 动画、3 卡特性栅格、"AI Gateway Infrastructure" 这种空洞 tagline，整体像 AI 模板批发出来的。需要重做成"开发者文档式"的克制风格，与新的 AIGateway logo 气质一致（几何严谨 + 三色玩心）。

## 目标 / 非目标

**目标**
- 视觉风格："克制有玩心" —— 中性色为主（白/灰/黑），品牌蓝紫绿三色仅做 accent
- 内容结构：文档式信息流，让产品自己说话（真实 curl + 真实价格表 > 营销文案）
- 主题感知：跟随用户切换亮/暗，与其他页一致（去掉首页"永远深色"的特殊化）
- 排版：等宽字体承载所有数据/代码（model_id、价格、curl、API 路径）

**非目标**
- 不做 hero 右侧 illustration / 产品截图 / 动画
- 不做 testimonials、stat 数字、对比表
- 不引入语法高亮库（纯文本配色已够）
- 不改后端

## 设计

### 整体调性

- 字体：`Inter`（沿用） + `ui-monospace, "SF Mono", Menlo, Consolas, monospace`（数据/代码）
- 颜色：基于中性色（zinc 阶或 gray 阶），亮/暗主题各一套：
  - 亮：`bg-white text-gray-900`，分隔线 `border-gray-200`
  - 暗：`bg-gray-950 text-white`，分隔线 `border-gray-800`
- 品牌色 `#3B82F6` (blue) / `#8B5CF6` (violet) / `#10B981` (green) 三色仅出现在 (1) 主 CTA hover 阴影、(2) 模型行 provider 徽章（复用 ProviderBrandIcon）、(3) 3 行价值的 bullet dot（一行一色循环）
- 装饰：**全砍** —— 删 `home-bg-orb-*` / `home-bg-dots` / glass header / pulsing dot / terminal perspective 3D / feature icon gradient

### Sections（自上而下）

#### 1. Header（沿用现有 nav）

不动 —— 已经在前面 commit 中重做为 `<img src="/logo.png"> + AIGateway`。

#### 2. Hero

```
One API key. All AI models.
Claude · GPT · Gemini · {N} models, ¥-per-token billing,
no subscription required.

[ Get Started → ]   View docs
```

- h1: `text-5xl sm:text-6xl font-semibold tracking-tight text-gray-900 dark:text-white`，**不渐变文字**
- 副标题: `text-base text-gray-600 dark:text-gray-400`，`{N}` 从 `/api/v1/settings/models` 实时取唯一 model_id 数量
- 主 CTA: 黑底白字 / 白底黑字（按主题反），rounded-md，hover 时 box-shadow 用品牌蓝色 8% 透明
  - 未登录跳 `/login`，已登录跳 `/dashboard`（admin 跳 `/admin/dashboard`，与现有 isAuthenticated/dashboardPath 逻辑一致）
- 次 CTA: plain underline link，跳 `/docs`
- 容器: `max-w-3xl` 居中，纵向 padding 充足让 hero 透气

#### 3. Curl 示例

```
Make your first call

┌───────────────────────────────────┐
│ POST /v1/messages       [Copy]    │
├───────────────────────────────────┤
│ curl https://aiinterface.store/.. │
│   -H "x-api-key: sk-xxxxxx" \     │
│   -H "anthropic-version: ..." \   │
│   -d '{...}'                      │
└───────────────────────────────────┘
```

- 标题: `text-xl font-semibold` + 灰色描述 "Use any supported model with the same key."
- 代码块容器: `border` + `rounded-lg`，亮色 `bg-gray-50 border-gray-200`，暗色 `bg-gray-950 border-gray-800`
- 顶部小 toolbar: 左侧灰色路径标签 `POST /v1/messages`，右侧"Copy"按钮（复用 [`ModelDetailDrawer.vue`](../../../frontend/src/components/models/ModelDetailDrawer.vue) 中已有的复制实现）
- `<pre>`: `font-mono text-sm leading-relaxed`，`overflow-x-auto`，padding `p-4`
- 内容: 复用 [`buildCurlExample('claude-sonnet-4-6', 'claude', baseUrl)`](../../../frontend/src/components/models/buildCurlExample.ts) 输出
- 占位 key: `sk-xxxxxx`（已是默认）

#### 4. Pricing 表

```
Pricing                   Real-time pricing →

  ●  claude-opus-4-6        ¥16.00 / ¥80.00
  ●  claude-sonnet-4-6      ¥8.40  / ¥50.00
  ●  claude-haiku-4-5       ¥3.20  / ¥18.00
  ●  gpt-5-2025...          ¥X.XX  / ¥X.XX
  ●  gemini-2.0-flash...    ¥X.XX  / ¥X.XX
  ...

  View all {N} models →
```

- 标题行: 左 `text-xl font-semibold "Pricing"`，右一个小 link "Real-time pricing →" 跳 `/models`（强调价格是动态的）
- 数据源: 抽出 `frontend/src/composables/usePublicModels.ts`，封装 `GET /api/v1/settings/models` + 缓存。HomeView 和 ModelsView 共用
- 排序: 按 `input_price desc` 取 Top 8。同 model_id 多分组的去重（按 ModelsView 现有逻辑）
- 行结构: `display: grid; grid-template-columns: 24px 1fr auto;`，gap-3
  - 第 1 列: `<ProviderBrandIcon :provider circle class="h-5 w-5">` (logo)
  - 第 2 列: `<code class="font-mono text-sm">{{ model_id }}</code>`（mono，可复制选中）
  - 第 3 列: `<span class="font-mono text-sm tabular-nums text-gray-600 dark:text-gray-400">¥{{ fmtPrice(input) }} / ¥{{ fmtPrice(output) }}</span>`（mono + tabular-nums 让数字列对齐）
- 行 hover: `bg-gray-50 dark:bg-gray-900/50`，cursor-pointer，整行点击跳 ModelsView 并打开该 model 的 detail drawer（复用刚做好的 drawer）—— 这点是 nice to have，如果实现复杂可降级为整行链接到 /模型 不带 query
- 表底: "View all {N} models →" link 跳 /模型

#### 5. 3 行价值

```
●  Multi-account routing with automatic failover.
●  Token-level billing — pay only for what you use.
●  Drop-in compatible with Claude Code, OpenAI SDK, Gemini CLI.
```

- `<ul class="space-y-3">`，每个 `<li class="flex gap-3 items-baseline">`
- bullet: 8×8 圆点，三个颜色循环：
  - 第 1 行: `bg-blue-500`（logo 蓝）
  - 第 2 行: `bg-violet-500`（logo 紫）
  - 第 3 行: `bg-emerald-500`（logo 绿）
- 文字: `text-base text-gray-700 dark:text-gray-300`
- **不做卡片、不做图标、不做 hover 动画**
- 中文版（i18n zh）：
  - "多账号路由，故障自动切换。"
  - "按 token 计费 —— 只为使用付费。"
  - "兼容 Claude Code、OpenAI SDK、Gemini CLI。"

#### 6. Footer

保留现有 copyright + Docs/GitHub 链接，结构不动，但删掉装饰：
- 顶部 `border-t border-gray-200 dark:border-gray-800`
- 文字 `text-xs text-gray-500 dark:text-gray-500`
- 间距 `py-8`

## 技术架构

### 新增

**`frontend/src/composables/usePublicModels.ts`**
```ts
export interface ModelEntry { /* 复用 providerUtils.ts 的类型 */ }
export function usePublicModels(): {
  items: Ref<ModelEntry[]>
  loading: Ref<boolean>
  fetch: () => Promise<void>
  uniqueModelCount: ComputedRef<number>
}
```
- 单例 / module-level 缓存，避免 Home 和 Models 双重请求
- 5 分钟 TTL（model 列表是配置型数据，变化频率低）

### 改写

**`frontend/src/views/HomeView.vue`**
- 大幅瘦身：从 855 行 → 预计 ~300 行
- 删除：所有 `home-bg-*` / `home-feature-*` / `home-provider-*` / `terminal-*` / `home-cta-*` / `t-line-*` / `home-section-*` 等约 40 个 CSS 类（不再使用）
- 保留：nav header（已重做）、`home-shell` 顶层 wrapper（瘦身后）、footer 基础结构
- 重构：sections 全部用 Tailwind utility class 直接写，不再走 scoped CSS 命名

**`frontend/src/i18n/locales/{en,zh}.ts`**
- `home` namespace 重写
- 删除（不再使用）: `tags.*` / `features.*` / `providers.*` / `cta.*` / `painPoints.*` / `solutions.*` / `comparison.*` 等
- 新增 keys:
  ```
  home: {
    hero: { title, subtitle, getStarted, viewDocs }
    curl: { title, description, copy, copied }
    pricing: { title, realtimeLink, viewAll }
    values: { routing, billing, compat }
    footer: { allRightsReserved }
  }
  ```

### 复用

- `ProviderBrandIcon`（已存在）
- `buildCurlExample`（已存在）
- `providerUtils.ts` 的 `fmtPrice` / `providerLabel` / `ModelEntry`
- `LocaleSwitcher`、`Icon`（已存在）
- `useClipboard` composable（已存在，用于 Copy 按钮）

## 错误处理

- `/settings/models` 请求失败 → Pricing 表显示 4 行 skeleton 占位 + 重试按钮
- 离线/慢网 → Hero 副标题里 `{N} models` 显示 "many"（fallback 文案）
- 移动端 < 640px：所有 section padding 减半，hero 字号降至 `text-4xl`，pricing 表第 3 列价格换行到第二行（保持 mono 对齐）

## 测试

参照仓库现有 vitest 模式：

- `usePublicModels.spec.ts` 单测：fetch 成功 / 失败 / 缓存命中 / TTL 过期重取
- HomeView 不单测整页渲染（页面级测试维护成本高，且大部分组件已有自己的单测）—— 仅做 type check + 浏览器手验
- 现有 `ModelDetailDrawer.spec.ts`、`buildCurlExample.spec.ts` 不受影响

## 实施顺序

1. 抽 `usePublicModels.ts` composable + 单测
2. 重写 i18n keys（新增 + 标记/删除旧）
3. 重写 HomeView.vue：先 hero + footer + 主题感知，再 curl block，再 pricing 表，再 3 行价值
4. 删除全部不再用的 `home-*` / `terminal-*` 等 scoped CSS
5. 浏览器手验：亮/暗主题 + 移动端 + Copy 按钮 + Pricing 行 hover/click
6. `vue-tsc` + `npm run build` + `vitest run`

每步独立可提交。

## 风险 / 取舍

- **国际化文案需要重写** —— 旧 home.* keys 大半作废。我会用 git diff 标记哪些 key 删除，便于 review
- **Pricing 行点击打开 drawer 的跨页协作** —— 我标记为 nice to have；如果实现需要 router state + ModelsView 暴露 API，太复杂就降级为整行 link 到 /模型 不带 query
- **数据源延迟** —— Hero 副标题里 `{N} models` 需要等 `/settings/models` 加载。期间显示 fallback 字符串"28+ models"或单纯"many models"避免空格抖动
