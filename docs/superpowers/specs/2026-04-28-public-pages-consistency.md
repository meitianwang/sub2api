# 公开页 / 控制台 UI 一致性 (Public Pages & Console Consistency)

**日期**: 2026-04-28
**作者**: meitianwang + Claude
**状态**: Approved (待实现)

## 背景

[`HomeView.vue`](../../../frontend/src/views/HomeView.vue) 已按 [`2026-04-28-homepage-redesign-design.md`](./2026-04-28-homepage-redesign-design.md) 重做为"开发者文档式"克制风格 —— 中性灰阶 + 蓝/紫/绿三色 8% accent，去掉了所有 glassmorphism / 紫色渐变装饰。但 [`ModelsView.vue`](../../../frontend/src/views/ModelsView.vue) / [`DocsView.vue`](../../../frontend/src/views/DocsView.vue) / 控制台 shell（[`AppLayout.vue`](../../../frontend/src/components/layout/AppLayout.vue) + [`AppHeader.vue`](../../../frontend/src/components/layout/AppHeader.vue) + [`AppSidebar.vue`](../../../frontend/src/components/layout/AppSidebar.vue)）还停留在旧风格 —— 紫色渐变 banner、`bg-white/70 backdrop-blur-md` glass nav、`primary-500` 紫色 active 状态、`bg-mesh-gradient` 紫色背景晕。

## 目标 / 非目标

**目标**
- 三个公开页（Home / Models / Docs）顶部 nav 完全一致 —— 抽 `<PublicNav>` 组件
- ModelsView：去紫色 banner、卡片改文档式（border-only、无 shadow/无 hover scale）
- DocsView：去 glass nav、sidebar active 改成中性 + 左侧细蓝条、保持文档可读性
- 控制台 shell：去 `bg-mesh-gradient` 装饰底、去 `primary-` 紫色 avatar 渐变、sidebar active 改成中性强调
- 视觉规则跟首页对齐：`bg-white dark:bg-gray-950`、`border-gray-200 dark:border-gray-800`、CTA 黑底白字 / 白底黑字

**非目标**
- 不动控制台**内部**业务页（DashboardView / KeysView / UsageView / 所有 admin/* 等）—— 它们是功能型表格/表单页，跟营销页是两套设计语言，单独迭代
- 不动 auth 页（Login/Register/Forgot/Reset）—— 用 `AuthLayout` 自成体系，单独迭代
- 不改后端、不改 i18n 结构（除非现有 key 不够用）
- 不引入新依赖

## 设计

### 共享视觉规则（与首页一致）

| 项 | 亮色 | 暗色 |
|---|---|---|
| 页面底 | `bg-white text-gray-900` | `bg-gray-950 text-white` |
| 分隔线 | `border-gray-200` | `border-gray-800` |
| 次要文字 | `text-gray-600` / `text-gray-500` | `text-gray-400` / `text-gray-500` |
| 卡片底 | 透明 / `bg-gray-50` | 透明 / `bg-gray-900/50` |
| Hover | `hover:bg-gray-50` | `hover:bg-gray-900/50` |
| 主 CTA | `bg-gray-900 text-white` | `bg-white text-gray-900` |
| Mono 数据 | `font-mono text-sm` | 同 |
| 三色 accent | `#3B82F6` / `#8B5CF6` / `#10B981`（仅 dot/icon/hover shadow） | 同 |

**禁用**：`backdrop-blur`、`bg-mesh-gradient`、`from-primary-* to-primary-*` 渐变、`bg-violet-* via-purple-*` banner、`shadow-card-hover` 之类的紫色 glow。

### 1. `<PublicNav>` 组件（新增）

**`frontend/src/components/common/PublicNav.vue`**

抽出 HomeView/ModelsView/DocsView 三处重复的 nav header。

```vue
<template>
  <header class="border-b border-gray-200 dark:border-gray-800">
    <nav class="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
      <router-link to="/home" class="flex items-center gap-2">
        <img src="/logo.png" alt="" class="h-7 w-7" />
        <span class="text-base font-semibold tracking-tight">AIGateway</span>
      </router-link>

      <div class="hidden items-center gap-1 sm:flex">
        <router-link to="/home" :class="navLinkClass('home')">{{ t('nav.home') }}</router-link>
        <router-link to="/models" :class="navLinkClass('models')">{{ t('nav.models') }}</router-link>
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="nav-link">{{ t('nav.docs') }}</a>
        <router-link v-else to="/docs" :class="navLinkClass('docs')">{{ t('nav.docs') }}</router-link>
        <router-link :to="consoleHref" class="nav-link">{{ t('nav.console') }}</router-link>
      </div>

      <div class="flex items-center gap-2">
        <LocaleSwitcher />
        <button @click="toggleTheme" class="theme-toggle"><Icon ... /></button>
        <UserAvatarOrLogin />
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
const props = defineProps<{ active?: 'home' | 'models' | 'docs' }>()
// active prop 决定哪个 link 加 nav-link-active class
</script>
```

- 容器宽度：`max-w-5xl`（Models/Docs 之前是 `max-w-7xl`，统一收窄到首页的 5xl）
- nav-link active：`text-gray-900 dark:text-white`，**不要紫色背景或下划线**
- nav-link inactive：`text-gray-600 hover:bg-gray-100 hover:text-gray-900` + dark variants
- 头像：`bg-gray-900 dark:bg-white` 圆形（**去 `from-primary-400 to-primary-600` 渐变**）
- 登录按钮：黑底白字 rounded-md（**去 `bg-primary-500 rounded-full`**）
- 主题切换 button：跟首页一致（rounded-md p-1.5）

### 2. ModelsView

**改动**
- nav header → 替换为 `<PublicNav active="models">`
- **删紫色 banner**（`bg-gradient-to-r from-violet-600 via-purple-600 to-fuchsia-500`）→ 改为：
  ```
  <h1 text-2xl font-semibold>Models</h1>
  <p text-sm text-gray-500>{{ count }} models · {{ description }}</p>
  ```
  `border-b` 分隔后是过滤器和卡片。
- **卡片栅格保留 3 列**，但去装饰：
  - `rounded-xl` → `rounded-lg`
  - 删 `hover:shadow-md`、`hover:border-gray-300` 软装饰 → 改 `hover:bg-gray-50 dark:hover:bg-gray-900/50`
  - `display_name` 字号从 `text-sm font-semibold` 提到 `text-sm font-medium`（不要太重）
  - 价格行用 `font-mono text-xs tabular-nums`，跟首页 pricing 表对齐
  - provider badge 仍用 `providerBadge()`（橙/灰/蓝），但底色降到 `bg-orange-50/50` 之类（不抢戏）
- **filter tags**：active 状态从 `bg-primary-50 text-primary-700` 改为 `bg-gray-900 text-white dark:bg-white dark:text-gray-900`
- **sort 按钮**同上去紫色
- **背景色**：`bg-gray-50 dark:bg-dark-950` → `bg-white dark:bg-gray-950`

### 3. DocsView

**改动**
- nav header → `<PublicNav active="docs">`
- **sidebar 去 glass**：`bg-white/90 backdrop-blur-md` → `bg-white dark:bg-gray-950`
- **sidebar active 链接**：`bg-primary-50 text-primary-700` → `text-gray-900 dark:text-white font-medium` + 左侧 `border-l-2 border-blue-500`（蓝色细条做 accent，三色规则的"蓝"）
- **section icon 配色**：原本是 `bg-primary-50 text-primary-600` → 改为中性 `bg-gray-100 text-gray-500`
- **prev/next 链接 hover**：`hover:text-primary-600` → `hover:text-gray-900 dark:hover:text-white`
- 背景：`bg-gray-50 dark:bg-dark-950` → `bg-white dark:bg-gray-950`
- 文档正文（`docs-content`）不动 —— 是 v-html 渲染的 markdown，prose 样式独立

### 4. 控制台 Shell

**[`AppLayout.vue`](../../../frontend/src/components/layout/AppLayout.vue)**
- 删 `<div bg-mesh-gradient opacity-60>` 紫色装饰底
- `bg-accent-50 dark:bg-dark-950` → `bg-white dark:bg-gray-950`

**[`AppHeader.vue`](../../../frontend/src/components/layout/AppHeader.vue)**
- 用户头像：`bg-gradient-to-br from-primary-600 to-primary-700` → `bg-gray-900 dark:bg-white text-white dark:text-gray-900`，去 `shadow-sm`
- 余额 pill：`border-primary-100 bg-primary-50/80 text-primary-700` → `border-gray-200 bg-gray-50 text-gray-700` + 暗色对应（钱币图标保留中性色）
- header 底色：`bg-white/95 backdrop-blur-sm` → `bg-white dark:bg-gray-950`（去 backdrop-blur）

**[`AppSidebar.vue`](../../../frontend/src/components/layout/AppSidebar.vue)** + style.css
- `sidebar-link-active` 当前是 `linear-gradient(135deg, #7c3aed, #6d28d9) + box-shadow purple` —— 改为：
  ```css
  .sidebar-link-active {
    @apply bg-gray-100 text-gray-900 dark:bg-gray-800 dark:text-white;
    @apply font-medium;
    /* 左侧 2px 蓝条做 accent */
    box-shadow: inset 2px 0 0 #3B82F6;
  }
  ```
- sidebar header logo：跟 PublicNav 同款（`h-7 w-7` + `font-semibold tracking-tight`，**不要 font-bold**）
- sidebar 底色：跟 AppHeader 一致

**保持**
- 各功能页内部内容（卡片、表格、表单）不动 —— 它们用了 `card`/`btn-primary` 等 utility class，独立批次再处理
- onboarding tour 不受影响（选择器都没变）

### 5. i18n

新增 keys（需要时）：
- 暂无 —— `nav.*` / `home.*` 现有 key 已覆盖。`models.banner` 文案保留（去 banner 视觉但文案复用）。

### 6. 删除/清理

style.css 中保留 `.sidebar-link` / `.app-header` 等结构 class（避免大改），仅修改 active state 颜色和 header 底色相关行。

## 实施顺序

1. **新增 `PublicNav.vue`**（先单独写、单独验证 props 与 active 状态切换）
2. **HomeView 改用 `<PublicNav active="home">`**（最小改动，删除内联 nav）
3. **ModelsView**：换 PublicNav → 删 banner → 改卡片样式 → 改 filter tag
4. **DocsView**：换 PublicNav → 改 sidebar active → 去 glass
5. **AppLayout / AppHeader / AppSidebar / style.css**：改 chrome 配色
6. `vue-tsc --noEmit` + `npm run build` + 浏览器手验亮/暗主题
7. 每步独立 commit（5 个 commit，对应 1-5）

## 风险 / 取舍

- **PublicNav 收窄到 max-w-5xl** —— ModelsView 内容区原本是 max-w-7xl，整页变窄 30%。收益是三页 nav 完全一致；代价是 Models 卡片栅格在大屏少一列（xl:grid-cols-3 → 仍然 3 列，因为 max-w-5xl ≈ 1024px，3 列依然合适）。**可接受**。
- **Sidebar active 改中性** —— 失去现有"紫色高亮一眼定位"的视觉权重。用左侧 2px 蓝条 + `font-medium` + `bg-gray-100` 三个信号叠加补偿，定位仍清晰。
- **控制台 shell 改但内部不改** —— 短期会有"chrome 中性 + 内部紫色按钮"轻微割裂感。这是有意的"先改外、后改内"，避免一次性 PR 太大不好 review。
- **不抽 `<ConsoleShell>` 组件** —— AppLayout 已经是 console shell，不重复抽。

## 测试

- `vue-tsc --noEmit` 全过
- `npm run build` 全过
- 浏览器手验：5 个页面（Home / Models / Docs / /dashboard / /admin/dashboard）× 亮/暗主题 = 10 组合
- onboarding tour 选择器（`#sidebar-channel-manage` 等）必须仍可命中
