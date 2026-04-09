/**
 * Documentation structure and navigation configuration
 * Defines sidebar sections, pages, and markdown content mapping
 */

export interface DocsPage {
  id: string
  title: string
  /** Icon name (optional, for section headers) */
  icon?: string
}

export interface DocsSection {
  id: string
  title: string
  icon?: string
  pages: DocsPage[]
}

export const docsStructure: DocsSection[] = [
  {
    id: 'getting-started',
    title: '快速开始',
    icon: 'rocket',
    pages: [
      { id: 'register', title: '注册账号' },
      { id: 'recharge', title: '充值余额' },
      { id: 'create-key', title: '创建 API Key' },
      { id: 'first-call', title: '发起第一次调用' },
    ],
  },
  {
    id: 'models',
    title: '模型与分组',
    icon: 'cube',
    pages: [
      { id: 'groups', title: '分组介绍' },
      { id: 'pricing', title: '模型与定价' },
    ],
  },
  {
    id: 'cli',
    title: 'CLI 配置教程',
    icon: 'terminal',
    pages: [
      { id: 'claude-code', title: 'Claude Code' },
      { id: 'codex', title: 'Codex CLI' },
      { id: 'gemini', title: 'Gemini CLI' },
    ],
  },
  {
    id: 'api',
    title: 'API 接入文档',
    icon: 'code',
    pages: [
      { id: 'overview', title: '概述' },
      { id: 'openai', title: 'OpenAI 兼容格式' },
      { id: 'anthropic', title: 'Anthropic 原生格式' },
      { id: 'gemini', title: 'Gemini 原生格式' },
    ],
  },
  {
    id: 'tools',
    title: 'IDE 与工具',
    icon: 'puzzle',
    pages: [
      { id: 'cursor', title: 'Cursor' },
      { id: 'cline', title: 'Cline / Continue' },
      { id: 'opencode', title: 'OpenCode' },
    ],
  },
  {
    id: 'faq',
    title: '常见问题',
    icon: 'question',
    pages: [
      { id: 'claude-code', title: 'Claude Code' },
      { id: 'codex', title: 'Codex' },
      { id: 'gemini', title: 'Gemini' },
      { id: 'general', title: '通用问题' },
    ],
  },
]

/** Default page to show when visiting /docs */
export const defaultSection = 'getting-started'
export const defaultPage = 'register'

/**
 * Load all markdown files via Vite glob import
 * Keys are like: ./getting-started/register.md
 */
const markdownModules = import.meta.glob('./**/*.md', { query: '?raw', import: 'default', eager: true }) as Record<string, string>

/**
 * Get markdown content for a given section/page
 */
export function getDocsContent(section: string, page: string): string | null {
  const key = `./${section}/${page}.md`
  return (markdownModules[key] as string) ?? null
}

/**
 * Find the previous and next pages for navigation
 */
export function getAdjacentPages(section: string, page: string): {
  prev: { section: string; page: string; title: string; sectionTitle: string } | null
  next: { section: string; page: string; title: string; sectionTitle: string } | null
} {
  const allPages: { section: string; sectionTitle: string; page: string; title: string }[] = []
  for (const s of docsStructure) {
    for (const p of s.pages) {
      allPages.push({ section: s.id, sectionTitle: s.title, page: p.id, title: p.title })
    }
  }
  const idx = allPages.findIndex((p) => p.section === section && p.page === page)
  return {
    prev: idx > 0 ? allPages[idx - 1] : null,
    next: idx < allPages.length - 1 ? allPages[idx + 1] : null,
  }
}
