<template>
  <div class="flex min-h-screen flex-col bg-gray-50 dark:bg-dark-950">
    <!-- Header -->
    <header class="sticky top-0 z-40 border-b border-gray-200/50 bg-white/80 px-4 py-3 backdrop-blur-md dark:border-dark-700/50 dark:bg-dark-900/80 sm:px-6">
      <nav class="mx-auto flex max-w-7xl items-center justify-between">
        <div class="flex items-center gap-3">
          <!-- Mobile sidebar toggle -->
          <button @click="sidebarOpen = !sidebarOpen" class="rounded-lg p-1.5 text-gray-500 hover:bg-gray-100 dark:text-dark-400 dark:hover:bg-dark-800 lg:hidden">
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" /></svg>
          </button>
          <router-link to="/home" class="flex items-center">
            <div class="h-9 w-9 overflow-hidden rounded-xl shadow-md">
              <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
            </div>
          </router-link>
        </div>
        <div class="hidden items-center gap-1 sm:flex">
          <router-link to="/home" class="nav-tab">{{ t('nav.home') }}</router-link>
          <router-link to="/models" class="nav-tab">{{ t('nav.models') }}</router-link>
          <router-link to="/docs" class="nav-tab nav-tab-active">{{ t('nav.docs') }}</router-link>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="nav-tab">{{ t('nav.console') }}</router-link>
        </div>
        <div class="flex items-center gap-2">
          <LocaleSwitcher />
          <button @click="toggleTheme" class="rounded-lg p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white">
            <svg v-if="isDark" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" /></svg>
            <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" /></svg>
          </button>
          <router-link v-if="isAuthenticated" :to="dashboardPath" class="flex h-7 w-7 items-center justify-center rounded-full bg-gradient-to-br from-primary-400 to-primary-600 text-[11px] font-semibold text-white">{{ userInitial }}</router-link>
          <router-link v-else to="/login" class="rounded-full bg-primary-500 px-3.5 py-1 text-xs font-medium text-white transition-colors hover:bg-primary-600">{{ t('home.login') }}</router-link>
        </div>
      </nav>
    </header>

    <div class="mx-auto flex w-full max-w-7xl flex-1">
      <!-- Mobile sidebar overlay -->
      <div v-if="sidebarOpen" class="fixed inset-0 z-30 bg-black/40 lg:hidden" @click="sidebarOpen = false" />

      <!-- Sidebar -->
      <aside :class="[
        'fixed inset-y-0 left-0 z-30 mt-[57px] w-72 overflow-y-auto border-r border-gray-200/50 bg-white/90 backdrop-blur-md transition-transform duration-200 dark:border-dark-700/50 dark:bg-dark-900/90 lg:sticky lg:top-[57px] lg:z-10 lg:h-[calc(100vh-57px)] lg:translate-x-0 lg:bg-transparent lg:backdrop-blur-none dark:lg:bg-transparent',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full'
      ]">
        <nav class="p-4">
          <div v-for="section in docsStructure" :key="section.id" class="mb-5">
            <h3 class="mb-1.5 flex items-center gap-2 px-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-dark-500">
              <span class="inline-flex h-5 w-5 items-center justify-center rounded text-[11px]" :class="sectionIconClass(section.icon)">{{ sectionIcon(section.icon) }}</span>
              {{ section.title }}
            </h3>
            <ul>
              <li v-for="page in section.pages" :key="page.id">
                <router-link
                  :to="`/docs/${section.id}/${page.id}`"
                  :class="[
                    'block rounded-lg px-3 py-1.5 text-[13px] transition-colors',
                    isActive(section.id, page.id)
                      ? 'bg-primary-50 font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300'
                      : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white'
                  ]"
                  @click="sidebarOpen = false"
                >
                  {{ page.title }}
                </router-link>
              </li>
            </ul>
          </div>
        </nav>
      </aside>

      <!-- Main content -->
      <main class="min-w-0 flex-1 px-6 py-8 sm:px-10 lg:px-12">
        <article v-if="renderedContent" class="docs-content mx-auto max-w-3xl">
          <div v-html="renderedContent" />

          <!-- Prev / Next navigation -->
          <div class="mt-12 flex items-center justify-between border-t border-gray-200 pt-6 dark:border-dark-700">
            <router-link
              v-if="adjacentPages.prev"
              :to="`/docs/${adjacentPages.prev.section}/${adjacentPages.prev.page}`"
              class="group flex items-center gap-2 text-sm text-gray-500 transition-colors hover:text-primary-600 dark:text-dark-400 dark:hover:text-primary-400"
            >
              <svg class="h-4 w-4 transition-transform group-hover:-translate-x-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
              <div class="text-right">
                <div class="text-[11px] text-gray-400 dark:text-dark-500">{{ adjacentPages.prev.sectionTitle }}</div>
                <div class="font-medium">{{ adjacentPages.prev.title }}</div>
              </div>
            </router-link>
            <div v-else />
            <router-link
              v-if="adjacentPages.next"
              :to="`/docs/${adjacentPages.next.section}/${adjacentPages.next.page}`"
              class="group flex items-center gap-2 text-sm text-gray-500 transition-colors hover:text-primary-600 dark:text-dark-400 dark:hover:text-primary-400"
            >
              <div>
                <div class="text-[11px] text-gray-400 dark:text-dark-500">{{ adjacentPages.next.sectionTitle }}</div>
                <div class="font-medium">{{ adjacentPages.next.title }}</div>
              </div>
              <svg class="h-4 w-4 transition-transform group-hover:translate-x-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </router-link>
            <div v-else />
          </div>
        </article>

        <!-- 404 -->
        <div v-else class="flex flex-col items-center justify-center py-20 text-gray-400 dark:text-dark-500">
          <svg class="mb-4 h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
          <p class="text-lg font-medium">页面未找到</p>
          <p class="mt-1 text-sm">请从左侧导航选择一个文档页面</p>
        </div>
      </main>

      <!-- TOC (table of contents) -->
      <aside v-if="tocItems.length > 1" class="hidden w-52 shrink-0 xl:block">
        <div class="sticky top-[73px] max-h-[calc(100vh-73px)] overflow-y-auto py-8 pr-4">
          <h4 class="mb-3 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-dark-500">目录</h4>
          <ul class="space-y-1 border-l border-gray-200 dark:border-dark-700">
            <li v-for="item in tocItems" :key="item.id">
              <a
                :href="`#${item.id}`"
                :class="[
                  'block border-l-2 py-1 text-[12px] leading-snug transition-colors',
                  item.level === 2 ? 'pl-3' : 'pl-5',
                  activeHeading === item.id
                    ? 'border-primary-500 font-medium text-primary-600 dark:text-primary-400'
                    : 'border-transparent text-gray-500 hover:text-gray-800 dark:text-dark-400 dark:hover:text-dark-200'
                ]"
              >{{ item.text }}</a>
            </li>
          </ul>
        </div>
      </aside>
    </div>

    <!-- Footer -->
    <footer class="border-t border-gray-200/50 px-6 py-6 dark:border-dark-800/50">
      <div class="mx-auto flex max-w-7xl items-center justify-center">
        <p class="text-xs text-gray-400 dark:text-dark-500">&copy; {{ currentYear }} {{ siteName }}. All rights reserved.</p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAppStore, useAuthStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import {
  docsStructure,
  defaultSection,
  defaultPage,
  getDocsContent,
  getAdjacentPages,
} from '@/content/docs'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'AIInterface')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const apiBaseUrl = computed(() => appStore.cachedPublicSettings?.api_base_url || window.location.origin)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => authStore.user?.role === 'admin' ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user) return ''
  if (user.username) return user.username[0].toUpperCase()
  if (user.email) return user.email[0].toUpperCase()
  return ''
})
const currentYear = new Date().getFullYear()

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}
onMounted(() => {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
})

// Sidebar
const sidebarOpen = ref(false)

// Current section/page from route
const currentSection = computed(() => (route.params.section as string) || defaultSection)
const currentPage = computed(() => (route.params.page as string) || defaultPage)

function isActive(sectionId: string, pageId: string) {
  return currentSection.value === sectionId && currentPage.value === pageId
}

// Redirect bare /docs to default page
watch(
  () => route.path,
  (path) => {
    if (path === '/docs' || path === '/docs/') {
      router.replace(`/docs/${defaultSection}/${defaultPage}`)
    }
  },
  { immediate: true }
)

// Markdown rendering
marked.setOptions({ breaks: true, gfm: true })

const renderedContent = computed(() => {
  const raw = getDocsContent(currentSection.value, currentPage.value)
  if (!raw) return null
  const content = raw.replace(/\{\{BASE_URL\}\}/g, apiBaseUrl.value)
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
})

// Adjacent pages for prev/next nav
const adjacentPages = computed(() => getAdjacentPages(currentSection.value, currentPage.value))

// TOC extraction
interface TocItem { id: string; text: string; level: number }
const tocItems = ref<TocItem[]>([])
const activeHeading = ref('')

function extractToc() {
  nextTick(() => {
    const article = document.querySelector('.docs-content')
    if (!article) { tocItems.value = []; return }
    const headings = article.querySelectorAll('h2, h3')
    tocItems.value = Array.from(headings).map((h) => {
      const text = h.textContent?.trim() || ''
      const id = h.id || text.toLowerCase().replace(/[^a-z0-9\u4e00-\u9fff]+/g, '-').replace(/(^-|-$)/g, '')
      if (!h.id) h.id = id
      return { id, text, level: parseInt(h.tagName[1]) }
    })
  })
}

watch([currentSection, currentPage], extractToc)
onMounted(extractToc)

// Scroll spy for TOC
let observer: IntersectionObserver | null = null
function setupScrollSpy() {
  nextTick(() => {
    observer?.disconnect()
    const headings = document.querySelectorAll('.docs-content h2, .docs-content h3')
    if (!headings.length) return
    observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            activeHeading.value = entry.target.id
            break
          }
        }
      },
      { rootMargin: '-80px 0px -70% 0px', threshold: 0 }
    )
    headings.forEach((h) => observer!.observe(h))
  })
}
watch([currentSection, currentPage], setupScrollSpy)
onMounted(setupScrollSpy)
onUnmounted(() => observer?.disconnect())

// Section icons
function sectionIcon(icon?: string): string {
  const icons: Record<string, string> = {
    rocket: '\u{1F680}',
    cube: '\u{1F4E6}',
    terminal: '\u{1F4BB}',
    code: '\u{1F4C4}',
    puzzle: '\u{1F9E9}',
    question: '\u{2753}',
  }
  return icon ? icons[icon] || '' : ''
}
function sectionIconClass(_icon?: string): string {
  return ''
}
</script>

<style scoped>
/* ── Nav Tabs ── */
.nav-tab { padding: 6px 16px; font-size: 14px; font-weight: 500; color: #6b7280; border-radius: 8px; transition: all 0.2s; text-decoration: none; }
.nav-tab:hover { color: #111827; background: rgba(0,0,0,0.04); }
.nav-tab-active { color: #7c3aed; background: rgba(124,58,237,0.08); }
:deep(.dark) .nav-tab { color: #9ca3af; }
:deep(.dark) .nav-tab:hover { color: #f3f4f6; background: rgba(255,255,255,0.06); }
:deep(.dark) .nav-tab-active { color: #a78bfa; background: rgba(124,58,237,0.12); }

/* ── Markdown Content ── */
.docs-content { @apply text-gray-800 dark:text-gray-200 leading-relaxed; }
.docs-content :deep(h1) { @apply text-2xl font-bold text-gray-900 dark:text-white mb-2 pb-3 border-b border-gray-200 dark:border-dark-700; }
.docs-content :deep(h2) { @apply text-xl font-semibold text-gray-900 dark:text-white mt-10 mb-4 pb-2 border-b border-gray-100 dark:border-dark-800; }
.docs-content :deep(h3) { @apply text-lg font-semibold text-gray-900 dark:text-white mt-8 mb-3; }
.docs-content :deep(h4) { @apply text-base font-semibold text-gray-900 dark:text-white mt-6 mb-2; }
.docs-content :deep(p) { @apply mb-4 leading-7; }
.docs-content :deep(a) { @apply text-primary-600 dark:text-primary-400 hover:underline; }
.docs-content :deep(code) { @apply bg-gray-100 dark:bg-dark-800 text-pink-600 dark:text-pink-400 px-1.5 py-0.5 rounded text-sm font-mono; }
.docs-content :deep(pre) { @apply bg-gray-900 dark:bg-dark-900 rounded-xl p-4 mb-4 overflow-x-auto border border-gray-200 dark:border-dark-700; }
.docs-content :deep(pre code) { @apply bg-transparent text-gray-100 p-0 text-sm leading-6; }
.docs-content :deep(blockquote) { @apply border-l-4 border-primary-400 dark:border-primary-600 bg-primary-50 dark:bg-primary-900/20 px-4 py-3 mb-4 rounded-r-lg text-gray-700 dark:text-gray-300; }
.docs-content :deep(blockquote p) { @apply mb-0; }
.docs-content :deep(ul) { @apply list-disc pl-6 mb-4 space-y-1; }
.docs-content :deep(ol) { @apply list-decimal pl-6 mb-4 space-y-1; }
.docs-content :deep(li) { @apply leading-7; }
.docs-content :deep(table) { @apply w-full mb-4 border-collapse rounded-lg overflow-hidden text-sm; }
.docs-content :deep(thead) { @apply bg-gray-100 dark:bg-dark-800; }
.docs-content :deep(th) { @apply px-4 py-2.5 text-left font-semibold text-gray-900 dark:text-white border-b border-gray-200 dark:border-dark-700; }
.docs-content :deep(td) { @apply px-4 py-2.5 border-b border-gray-100 dark:border-dark-800; }
.docs-content :deep(tr:hover td) { @apply bg-gray-50 dark:bg-dark-900/50; }
.docs-content :deep(hr) { @apply my-8 border-gray-200 dark:border-dark-700; }
.docs-content :deep(strong) { @apply font-semibold text-gray-900 dark:text-white; }

/* Tip/Warning callout blocks via blockquote with specific prefixes */
.docs-content :deep(blockquote:has(> p:first-child > strong:first-child)) {
  @apply border-l-4;
}
</style>
