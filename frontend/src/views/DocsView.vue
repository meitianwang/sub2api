<template>
  <div class="flex min-h-screen flex-col bg-white text-gray-900 dark:bg-gray-950 dark:text-white">
    <!-- Top nav: PublicNav + mobile sidebar toggle prepended -->
    <div class="relative">
      <PublicNav active="docs" />
      <button
        @click="sidebarOpen = !sidebarOpen"
        class="absolute left-3 top-1/2 -translate-y-1/2 rounded-md p-1.5 text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800 lg:hidden"
        aria-label="Toggle docs sidebar"
      >
        <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" /></svg>
      </button>
    </div>

    <div class="mx-auto flex w-full max-w-7xl flex-1">
      <!-- Mobile sidebar overlay -->
      <div v-if="sidebarOpen" class="fixed inset-0 z-30 bg-black/40 lg:hidden" @click="sidebarOpen = false" />

      <!-- Sidebar -->
      <aside :class="[
        'fixed inset-y-0 left-0 z-30 mt-[57px] w-72 overflow-y-auto border-r border-gray-200 bg-white transition-transform duration-200 dark:border-gray-800 dark:bg-gray-950 lg:sticky lg:top-[57px] lg:z-10 lg:h-[calc(100vh-57px)] lg:translate-x-0',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full'
      ]">
        <nav class="p-4">
          <div v-for="section in docsStructure" :key="section.id" class="mb-5">
            <h3 class="mb-1.5 flex items-center gap-2 px-2 text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-500">
              <span class="inline-flex h-5 w-5 items-center justify-center rounded bg-gray-100 text-[11px] dark:bg-gray-800">{{ sectionIcon(section.icon) }}</span>
              {{ section.title }}
            </h3>
            <ul>
              <li v-for="page in section.pages" :key="page.id">
                <router-link
                  :to="`/docs/${section.id}/${page.id}`"
                  :class="[
                    'block rounded-md py-1.5 pl-3 pr-3 text-[13px] transition-colors',
                    isActive(section.id, page.id)
                      ? 'border-l-2 border-blue-500 bg-gray-50 pl-[10px] font-medium text-gray-900 dark:bg-gray-900/60 dark:text-white'
                      : 'border-l-2 border-transparent text-gray-600 hover:bg-gray-50 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-gray-900/40 dark:hover:text-white'
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
      <main class="min-w-0 flex-1 px-6 py-10 sm:px-10 lg:px-12">
        <article v-if="renderedContent" class="docs-content mx-auto max-w-3xl">
          <div v-html="renderedContent" />

          <!-- Prev / Next navigation -->
          <div class="mt-12 flex items-center justify-between border-t border-gray-200 pt-6 dark:border-gray-800">
            <router-link
              v-if="adjacentPages.prev"
              :to="`/docs/${adjacentPages.prev.section}/${adjacentPages.prev.page}`"
              class="group flex items-center gap-2 text-sm text-gray-500 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
            >
              <svg class="h-4 w-4 transition-transform group-hover:-translate-x-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
              <div class="text-right">
                <div class="text-[11px] text-gray-400 dark:text-gray-600">{{ adjacentPages.prev.sectionTitle }}</div>
                <div class="font-medium">{{ adjacentPages.prev.title }}</div>
              </div>
            </router-link>
            <div v-else />
            <router-link
              v-if="adjacentPages.next"
              :to="`/docs/${adjacentPages.next.section}/${adjacentPages.next.page}`"
              class="group flex items-center gap-2 text-sm text-gray-500 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
            >
              <div>
                <div class="text-[11px] text-gray-400 dark:text-gray-600">{{ adjacentPages.next.sectionTitle }}</div>
                <div class="font-medium">{{ adjacentPages.next.title }}</div>
              </div>
              <svg class="h-4 w-4 transition-transform group-hover:translate-x-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </router-link>
            <div v-else />
          </div>
        </article>

        <!-- 404 -->
        <div v-else class="flex flex-col items-center justify-center py-20 text-gray-400 dark:text-gray-600">
          <svg class="mb-4 h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
          <p class="text-lg font-medium">{{ t('docs.notFound.title') }}</p>
          <p class="mt-1 text-sm">{{ t('docs.notFound.hint') }}</p>
        </div>
      </main>

      <!-- TOC (table of contents) -->
      <aside v-if="tocItems.length > 1" class="hidden w-52 shrink-0 xl:block">
        <div class="sticky top-[73px] max-h-[calc(100vh-73px)] overflow-y-auto py-10 pr-4">
          <h4 class="mb-3 text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-500">{{ t('docs.toc') }}</h4>
          <ul class="space-y-1 border-l border-gray-200 dark:border-gray-800">
            <li v-for="item in tocItems" :key="item.id">
              <a
                :href="`#${item.id}`"
                :class="[
                  'block border-l-2 py-1 text-[12px] leading-snug transition-colors',
                  item.level === 2 ? 'pl-3' : 'pl-5',
                  activeHeading === item.id
                    ? 'border-blue-500 font-medium text-gray-900 dark:text-white'
                    : 'border-transparent text-gray-500 hover:text-gray-900 dark:text-gray-500 dark:hover:text-gray-200'
                ]"
              >{{ item.text }}</a>
            </li>
          </ul>
        </div>
      </aside>
    </div>

    <!-- Footer -->
    <footer class="border-t border-gray-200 px-6 py-6 dark:border-gray-800">
      <div class="mx-auto flex max-w-7xl items-center justify-center">
        <p class="text-xs text-gray-500 dark:text-gray-500">&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
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
import { useAppStore } from '@/stores'
import PublicNav from '@/components/common/PublicNav.vue'
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

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'AIInterface')
const apiBaseUrl = computed(() => appStore.cachedPublicSettings?.api_base_url || window.location.origin)
const currentYear = new Date().getFullYear()

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
</script>

<style scoped>
/* ── Markdown Content ── */
.docs-content { @apply text-gray-800 dark:text-gray-200 leading-relaxed; }
.docs-content :deep(h1) { @apply text-2xl font-semibold text-gray-900 dark:text-white mb-2 pb-3 border-b border-gray-200 dark:border-gray-800; }
.docs-content :deep(h2) { @apply text-xl font-semibold text-gray-900 dark:text-white mt-10 mb-4 pb-2 border-b border-gray-100 dark:border-gray-900; }
.docs-content :deep(h3) { @apply text-lg font-semibold text-gray-900 dark:text-white mt-8 mb-3; }
.docs-content :deep(h4) { @apply text-base font-semibold text-gray-900 dark:text-white mt-6 mb-2; }
.docs-content :deep(p) { @apply mb-4 leading-7; }
.docs-content :deep(a) { @apply text-blue-600 dark:text-blue-400 hover:underline; }
.docs-content :deep(code) { @apply bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-1.5 py-0.5 rounded text-sm font-mono; }
.docs-content :deep(pre) { @apply bg-gray-50 dark:bg-gray-900/50 rounded-lg p-4 mb-4 overflow-x-auto border border-gray-200 dark:border-gray-800; }
.docs-content :deep(pre code) { @apply bg-transparent text-gray-800 dark:text-gray-200 p-0 text-sm leading-6; }
.docs-content :deep(blockquote) { @apply border-l-2 border-blue-500 bg-blue-50/50 dark:bg-blue-950/20 px-4 py-3 mb-4 rounded-r text-gray-700 dark:text-gray-300; }
.docs-content :deep(blockquote p) { @apply mb-0; }
.docs-content :deep(ul) { @apply list-disc pl-6 mb-4 space-y-1; }
.docs-content :deep(ol) { @apply list-decimal pl-6 mb-4 space-y-1; }
.docs-content :deep(li) { @apply leading-7; }
.docs-content :deep(table) { @apply w-full mb-4 border-collapse text-sm; }
.docs-content :deep(thead) { @apply border-b border-gray-200 dark:border-gray-800; }
.docs-content :deep(th) { @apply px-4 py-2.5 text-left font-medium text-gray-900 dark:text-white; }
.docs-content :deep(td) { @apply px-4 py-2.5 border-b border-gray-100 dark:border-gray-900; }
.docs-content :deep(tr:hover td) { @apply bg-gray-50 dark:bg-gray-900/50; }
.docs-content :deep(hr) { @apply my-8 border-gray-200 dark:border-gray-800; }
.docs-content :deep(strong) { @apply font-semibold text-gray-900 dark:text-white; }
</style>
