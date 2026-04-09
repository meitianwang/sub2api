<template>
  <div class="flex min-h-screen flex-col bg-gray-50 dark:bg-dark-950">
    <!-- Header -->
    <header class="sticky top-0 z-30 border-b border-gray-200/50 bg-white/70 px-6 py-3 backdrop-blur-md dark:border-dark-700/50 dark:bg-dark-900/70">
      <nav class="mx-auto flex max-w-6xl items-center justify-between">
        <router-link to="/home" class="flex items-center">
          <div class="h-9 w-9 overflow-hidden rounded-xl shadow-md">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
        </router-link>
        <div class="hidden items-center gap-1 sm:flex">
          <router-link to="/home" class="nav-tab">{{ t('nav.home') }}</router-link>
          <router-link to="/models" class="nav-tab">{{ t('nav.models') }}</router-link>
          <router-link to="/docs" class="nav-tab nav-tab-active">{{ t('nav.docs') }}</router-link>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="nav-tab">{{ t('nav.console') }}</router-link>
        </div>
        <div class="flex items-center gap-2">
          <LocaleSwitcher />
          <button @click="toggleTheme" class="rounded-lg p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white">
            <Icon v-if="isDark" name="sun" size="sm" />
            <Icon v-else name="moon" size="sm" />
          </button>
          <router-link v-if="isAuthenticated" :to="dashboardPath" class="flex h-7 w-7 items-center justify-center rounded-full bg-gradient-to-br from-primary-400 to-primary-600 text-[11px] font-semibold text-white">{{ userInitial }}</router-link>
          <router-link v-else to="/login" class="rounded-full bg-primary-500 px-3.5 py-1 text-xs font-medium text-white transition-colors hover:bg-primary-600">{{ t('home.login') }}</router-link>
        </div>
      </nav>
    </header>

    <!-- Content -->
    <main class="mx-auto w-full max-w-4xl flex-1 px-6 py-10">
      <div class="docs-content" v-html="renderedContent"></div>
    </main>

    <!-- Footer -->
    <footer class="border-t border-gray-200/50 px-6 py-8 dark:border-dark-800/50">
      <div class="mx-auto flex max-w-6xl flex-col items-center justify-center gap-4 text-center sm:flex-row sm:text-left">
        <p class="text-sm text-gray-500 dark:text-dark-400">
          &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
        </p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAppStore, useAuthStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import docsRaw from '@/content/user-guide.md?raw'

const { t } = useI18n()
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

const isDark = ref(document.documentElement.classList.contains('dark'))

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

onMounted(() => {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
})

marked.setOptions({ breaks: true, gfm: true })

const renderedContent = computed(() => {
  const content = docsRaw.replace(/https:\/\/你的域名/g, apiBaseUrl.value)
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
})
</script>

<style scoped>
.docs-content {
  @apply text-gray-800 dark:text-gray-200 leading-relaxed;
}

.docs-content :deep(h1) {
  @apply text-3xl font-bold text-gray-900 dark:text-white mb-2 pb-3 border-b border-gray-200 dark:border-dark-700;
}

.docs-content :deep(h2) {
  @apply text-xl font-semibold text-gray-900 dark:text-white mt-10 mb-4 pb-2 border-b border-gray-100 dark:border-dark-800;
}

.docs-content :deep(h3) {
  @apply text-lg font-semibold text-gray-900 dark:text-white mt-8 mb-3;
}

.docs-content :deep(h4) {
  @apply text-base font-semibold text-gray-900 dark:text-white mt-6 mb-2;
}

.docs-content :deep(p) {
  @apply mb-4 leading-7;
}

.docs-content :deep(a) {
  @apply text-primary-600 dark:text-primary-400 hover:underline;
}

.docs-content :deep(code) {
  @apply bg-gray-100 dark:bg-dark-800 text-pink-600 dark:text-pink-400 px-1.5 py-0.5 rounded text-sm font-mono;
}

.docs-content :deep(pre) {
  @apply bg-gray-900 dark:bg-dark-900 rounded-xl p-4 mb-4 overflow-x-auto border border-gray-200 dark:border-dark-700;
}

.docs-content :deep(pre code) {
  @apply bg-transparent text-gray-100 p-0 text-sm leading-6;
}

.docs-content :deep(blockquote) {
  @apply border-l-4 border-primary-400 dark:border-primary-600 bg-primary-50 dark:bg-primary-900/20 px-4 py-3 mb-4 rounded-r-lg text-gray-700 dark:text-gray-300;
}

.docs-content :deep(blockquote p) {
  @apply mb-0;
}

.docs-content :deep(ul) {
  @apply list-disc list-inside mb-4 space-y-1;
}

.docs-content :deep(ol) {
  @apply list-decimal list-inside mb-4 space-y-1;
}

.docs-content :deep(li) {
  @apply leading-7;
}

.docs-content :deep(table) {
  @apply w-full mb-4 border-collapse rounded-lg overflow-hidden;
}

.docs-content :deep(thead) {
  @apply bg-gray-100 dark:bg-dark-800;
}

.docs-content :deep(th) {
  @apply px-4 py-2.5 text-left text-sm font-semibold text-gray-900 dark:text-white border-b border-gray-200 dark:border-dark-700;
}

.docs-content :deep(td) {
  @apply px-4 py-2.5 text-sm border-b border-gray-100 dark:border-dark-800;
}

.docs-content :deep(tr:hover td) {
  @apply bg-gray-50 dark:bg-dark-900/50;
}

.docs-content :deep(hr) {
  @apply my-8 border-gray-200 dark:border-dark-700;
}

.docs-content :deep(strong) {
  @apply font-semibold text-gray-900 dark:text-white;
}

/* Nav Tabs */
.nav-tab {
  padding: 6px 16px;
  font-size: 14px;
  font-weight: 500;
  color: #6b7280;
  border-radius: 8px;
  transition: all 0.2s;
  text-decoration: none;
}
.nav-tab:hover {
  color: #111827;
  background: rgba(0, 0, 0, 0.04);
}
.nav-tab-active {
  color: #14b8a6;
  background: rgba(20, 184, 166, 0.08);
}
:deep(.dark) .nav-tab {
  color: #9ca3af;
}
:deep(.dark) .nav-tab:hover {
  color: #f3f4f6;
  background: rgba(255, 255, 255, 0.06);
}
:deep(.dark) .nav-tab-active {
  color: #2dd4bf;
  background: rgba(20, 184, 166, 0.12);
}
</style>
