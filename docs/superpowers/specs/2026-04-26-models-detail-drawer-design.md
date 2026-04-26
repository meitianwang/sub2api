# 模型详情抽屉 (Models Detail Drawer)

**日期**: 2026-04-26
**作者**: meitianwang + Claude
**状态**: Approved (待实现)

## 背景

公开模型列表页 ([`frontend/src/views/ModelsView.vue`](../../../frontend/src/views/ModelsView.vue)) 当前展示模型卡片网格，访客无法直接看到"这个模型怎么调用"。需要一个面向访客的展示型组件，点击模型卡片后从右侧滑入抽屉，展示该模型的 curl 调用示例和基础信息。

定位：**访客（未登录）的模型市场展示**。不是开发者工具，curl 中 API key 用静态占位 `sk-xxxxxx`。

## 范围

**做：**
- 点击任意模型卡片打开右侧抽屉
- 抽屉展示模型 7 个已有字段 (model_id, display_name, provider, group_id, group_name, input_price, output_price)
- 单段 curl 示例，按 `provider` 选择对应原生格式（Anthropic / OpenAI / Gemini）
- "复制 curl" 按钮
- 桌面 480px 宽，移动端全宽，从右侧滑入
- 遮罩点击 / ESC / X 按钮关闭

**不做（YAGNI）：**
- 流式 (`stream: true`) 变体
- 多格式 Tab（每个模型只展示其原生格式）
- "用我的真实 API key"（访客定位）
- 后端补字段（context_length / capabilities / description）
- URL 同步 / 深链 / 浏览器后退键关闭（抽屉是纯前端组件状态）
- 语法高亮（避免引入 highlight.js / shiki，等宽字体配色已够）

## 架构

3 个文件改动 / 新增：

### 1. 新增 `frontend/src/components/models/buildCurlExample.ts`

纯函数，独立可测试。

```ts
export type CurlFormat = 'anthropic' | 'openai' | 'gemini'

export interface CurlExample {
  format: CurlFormat
  formatLabel: string  // 'Anthropic 格式' / 'OpenAI 格式' / 'Gemini 格式' (i18n key 在调用方)
  code: string         // 完整可复制的 curl 字符串
}

export function buildCurlExample(
  modelId: string,
  provider: string,    // 'claude' | 'openai' | 'gemini'
  baseUrl: string      // 已经处理过的根，不带尾部 /v1
): CurlExample
```

`provider → format` 映射：`claude → anthropic`, `openai → openai`, `gemini → gemini`。其他 provider 默认走 `openai` 格式。

BASE_URL 处理逻辑（来自 [UseKeyModal.vue:360-367](../../../frontend/src/components/keys/UseKeyModal.vue#L360) 现有模式）：
```ts
const baseRoot = (apiBaseUrl || window.location.origin)
  .replace(/\/v1\/?$/, '')
  .replace(/\/+$/, '')
```

### 2. 新增 `frontend/src/components/models/ModelDetailDrawer.vue`

```vue
<script setup lang="ts">
defineProps<{
  open: boolean
  model: ModelEntry | null
}>()
defineEmits<{ close: [] }>()
</script>
```

**结构：**
- `<Teleport to="body">` 挂到 body
- `<Transition name="drawer-fade">` 遮罩 200ms opacity
- `<Transition name="drawer-slide">` 抽屉 200ms translate-x
- 抽屉打开时给 `body` 加 `overflow: hidden`，关闭时恢复
- ESC 监听 keydown 全局事件，open 时绑定，close 时解绑

**内容区（top → bottom）：**

1. **Header**（sticky 顶部）
   - `<ProviderBrandIcon :provider circle class="h-9 w-9">`
   - `display_name` 大字
   - 关闭 X 按钮

2. **模型信息**
   - 模型 ID：等宽字体 + 行内复制图标，点击复制
   - 供应商：文字（`Anthropic` / `OpenAI` / `Google`），用 `providerLabel()` 复用 ModelsView 现有逻辑（从 ModelsView 抽出到 [shared util](../../../frontend/src/components/models/) 或留在 ModelsView 内复制一份——见"重构机会"段）
   - 令牌分组：`group_name` 徽章
   - 价格：`输入 ¥X.XX/M · 输出 ¥X.XX/M`，`fmtPrice()` 同上

3. **调用示例**
   - 小标题：`调用示例 · {formatLabel}`
   - `<pre>` 暗色背景代码块（`bg-gray-900 text-gray-100`），等宽字体，移动端 `overflow-x-auto`
   - 右上角"复制"按钮 → 点击后变"已复制 ✓" 1.5s

### 3. 改 `frontend/src/views/ModelsView.vue`

```ts
// 新增 state
const drawerModel = ref<ModelEntry | null>(null)
const drawerOpen = computed(() => drawerModel.value !== null)

function openDrawer(item: ModelEntry) { drawerModel.value = item }
function closeDrawer() { drawerModel.value = null }
```

**模板改动：**
- 卡片外层 `<div>` 加 `@click="openDrawer(item)"` + `cursor-pointer`
- 已有"复制 model_id"按钮加 `@click.stop`，避免触发卡片点击
- 网格末尾挂 `<ModelDetailDrawer :open="drawerOpen" :model="drawerModel" @close="closeDrawer" />`

## Curl 模板

`${BASE_URL}` 是处理后的根（不带 /v1）。`${MODEL_ID}` 是该模型的 `model_id`。占位 key 固定 `sk-xxxxxx`。

**Anthropic 格式**（`provider === 'claude'`）
```bash
curl ${BASE_URL}/v1/messages \
  -H "x-api-key: sk-xxxxxx" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{
    "model": "${MODEL_ID}",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

**OpenAI 格式**（`provider === 'openai'` 或其他）
```bash
curl ${BASE_URL}/v1/chat/completions \
  -H "Authorization: Bearer sk-xxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "${MODEL_ID}",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

**Gemini 格式**（`provider === 'gemini'`）
```bash
curl "${BASE_URL}/v1beta/models/${MODEL_ID}:generateContent" \
  -H "x-goog-api-key: sk-xxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [{"parts": [{"text": "Hello"}]}]
  }'
```

## i18n keys

新增至 [`frontend/src/i18n/locales/en.ts`](../../../frontend/src/i18n/locales/en.ts) 和 [`zh.ts`](../../../frontend/src/i18n/locales/zh.ts) 的 `models` 命名空间下：

```ts
models: {
  // ... existing keys
  detail: {
    modelId: 'Model ID' / '模型 ID'
    provider: 'Provider' / '供应商'
    group: 'Token Group' / '令牌分组'
    pricing: 'Pricing' / '价格'
    example: 'Example Call' / '调用示例'
    formatAnthropic: 'Anthropic Format' / 'Anthropic 格式'
    formatOpenAI: 'OpenAI Format' / 'OpenAI 格式'
    formatGemini: 'Gemini Format' / 'Gemini 格式'
    copy: 'Copy' / '复制'
    copied: 'Copied' / '已复制'
    close: 'Close' / '关闭'
  }
}
```

## 错误处理

- 抽屉接收 `model: null` 时不渲染内容（`v-if`），避免空指针
- `navigator.clipboard.writeText` 在不安全上下文 (HTTP) 失败时 catch 静默；按钮不变状态
- BASE_URL 缺失时回退 `window.location.origin`（永远有值）

## 测试

按照仓库现有模式（[`frontend/src/components/keys/__tests__/UseKeyModal.spec.ts`](../../../frontend/src/components/keys/__tests__/UseKeyModal.spec.ts)）：

- `buildCurlExample` 单测：3 种 provider + 2 种 baseUrl 形态（带 /v1 vs 不带）+ 1 种未知 provider 默认走 OpenAI 格式 = 7 case
- `ModelDetailDrawer.spec.ts` 组件测：
  - `open=true model=...` 渲染所有字段 + curl 块
  - 点击 X 触发 `close` 事件
  - ESC 触发 `close`
  - 复制按钮调用 `navigator.clipboard.writeText`

ModelsView 不需要新增集成测试（点击行为是单行 `@click`）。

## 重构机会

ModelsView 的 `providerLabel(p)` / `fmtPrice(p)` 在抽屉里要重用。两种处理：

(a) 抽屉自己复制一份这两个 4 行函数 —— 最小改动
(b) 抽到 `frontend/src/components/models/providerUtils.ts` —— ModelsView 也用，避免重复

推荐 **(b)**，对应 5 行新模块、ModelsView 删 5 行——零回归风险，也让 `providerUtils.ts` 后续承载更多模型相关纯函数。

## 实施顺序

1. 创建 `providerUtils.ts`（抽 `providerLabel` + `fmtPrice`），ModelsView 改成 import
2. 创建 `buildCurlExample.ts` + 单测
3. 创建 `ModelDetailDrawer.vue` + 组件测
4. ModelsView 加 state、卡片 `@click`、复制按钮 `@click.stop`、挂载抽屉
5. i18n keys
6. `npm run build` + 浏览器手动验证（点击/ESC/复制/移动端）

每步独立可提交。
