<template>
  <div class="flex min-h-screen flex-col bg-gray-50 dark:bg-dark-950">
    <!-- Header -->
    <header class="sticky top-0 z-30 border-b border-gray-200/50 bg-white/70 px-6 py-3 backdrop-blur-md dark:border-dark-700/50 dark:bg-dark-900/70">
      <nav class="mx-auto flex max-w-7xl items-center justify-between">
        <router-link to="/home" class="flex items-center">
          <div class="h-9 w-9 overflow-hidden rounded-xl shadow-md">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
        </router-link>
        <div class="hidden items-center gap-1 sm:flex">
          <router-link to="/home" class="nav-tab">{{ t('nav.home') }}</router-link>
          <router-link to="/models" class="nav-tab nav-tab-active">{{ t('nav.models') }}</router-link>
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="nav-tab">{{ t('nav.docs') }}</a>
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

    <div class="mx-auto flex w-full max-w-7xl flex-1 gap-0 px-4 py-6 lg:gap-6 lg:px-6">
      <!-- Sidebar -->
      <aside class="hidden w-56 flex-shrink-0 lg:block">
        <div class="sticky top-20 space-y-6">
          <!-- Provider -->
          <div>
            <h3 class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{{ t('models.sidebar.provider') }}</h3>
            <div class="flex flex-wrap gap-2">
              <button v-for="tab in providerTabs" :key="tab.value" @click="activeProvider = tab.value"
                :class="['filter-tag', activeProvider === tab.value ? 'filter-tag-active' : '']">
                <span v-if="tab.icon" :class="['flex h-4 w-4 items-center justify-center rounded', tab.iconClass]">
                  <span class="text-[8px] font-bold text-white">{{ tab.icon }}</span>
                </span>
                {{ tab.label }} <span class="opacity-50">{{ tab.count }}</span>
              </button>
            </div>
          </div>
          <!-- Group -->
          <div v-if="allGroups.length">
            <h3 class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{{ t('models.sidebar.group') }}</h3>
            <div class="flex flex-wrap gap-2">
              <button @click="activeGroup = ''" :class="['filter-tag', activeGroup === '' ? 'filter-tag-active' : '']">
                {{ t('models.filters.all') }} {{ allGroups.length }}
              </button>
              <button v-for="g in allGroups" :key="g" @click="activeGroup = g"
                :class="['filter-tag', activeGroup === g ? 'filter-tag-active' : '']">
                {{ g }}
              </button>
            </div>
          </div>
        </div>
      </aside>

      <!-- Main -->
      <main class="min-w-0 flex-1">
        <!-- Banner -->
        <div class="relative mb-5 overflow-hidden rounded-2xl bg-gradient-to-r from-indigo-600 via-blue-600 to-cyan-500 px-6 py-5">
          <div class="relative">
            <div class="mb-1 flex items-center gap-3">
              <h1 class="text-xl font-bold text-white">{{ bannerTitle }}</h1>
              <span class="rounded-full bg-white/20 px-2.5 py-0.5 text-xs font-medium text-white">
                {{ t('models.banner.count', { count: filtered.length }) }}
              </span>
            </div>
            <p class="text-sm text-white/70">{{ t('models.banner.description') }}</p>
          </div>
        </div>

        <!-- Mobile Filters + Search -->
        <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div class="flex flex-wrap items-center gap-1.5 lg:hidden">
            <button v-for="tab in providerTabs" :key="'m-'+tab.value" @click="activeProvider = tab.value"
              :class="['rounded-lg px-2.5 py-1 text-xs font-medium transition-all', activeProvider === tab.value ? 'bg-primary-500 text-white' : 'bg-white text-gray-600 dark:bg-dark-800 dark:text-dark-300']">
              {{ tab.label }} {{ tab.count }}
            </button>
          </div>
          <div class="flex items-center gap-2">
            <div class="relative w-full sm:w-64">
              <Icon name="search" size="sm" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
              <input v-model="searchQuery" type="text" :placeholder="t('models.searchPlaceholder')"
                class="w-full rounded-lg border border-gray-200 bg-white py-2 pl-9 pr-3 text-sm text-gray-900 placeholder-gray-400 transition-all focus:border-primary-300 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-800 dark:text-white dark:placeholder-dark-500" />
            </div>
            <button @click="toggleSort"
              :class="['flex items-center gap-1 rounded-lg border px-3 py-2 text-xs font-medium transition-all', sortBy === 'default' ? 'border-gray-200 bg-white text-gray-500 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-400' : 'border-primary-200 bg-primary-50 text-primary-700 dark:border-primary-800 dark:bg-primary-900/20 dark:text-primary-400']">
              <Icon :name="sortBy === 'price-desc' ? 'arrowDown' : 'arrowUp'" size="sm" />
              {{ t('models.sort.price') }}
            </button>
          </div>
        </div>

        <!-- Loading -->
        <div v-if="loading" class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          <div v-for="i in 9" :key="i" class="animate-pulse rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
            <div class="mb-3 flex items-center gap-3"><div class="h-9 w-9 rounded-full bg-gray-200 dark:bg-dark-700"></div><div class="h-4 w-32 rounded bg-gray-200 dark:bg-dark-700"></div></div>
            <div class="mb-2 h-3 w-48 rounded bg-gray-100 dark:bg-dark-700/60"></div>
            <div class="h-3 w-36 rounded bg-gray-100 dark:bg-dark-700/60"></div>
          </div>
        </div>

        <!-- Cards -->
        <div v-else-if="filtered.length" class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          <div v-for="item in filtered" :key="item.model_id + '-' + item.group_id"
            class="group rounded-xl border border-gray-200 bg-white p-4 transition-all duration-200 hover:border-gray-300 hover:shadow-md dark:border-dark-700 dark:bg-dark-800 dark:hover:border-dark-600">
            <!-- Top -->
            <div class="mb-2 flex items-start justify-between gap-3">
              <div class="flex items-center gap-3">
                <div :class="['flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-full', providerIconBg(item.provider)]">
                  <span class="text-xs font-bold text-white">{{ providerLetter(item.provider) }}</span>
                </div>
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ item.display_name }}</h3>
              </div>
              <button @click="copy(item.model_id)" class="rounded p-1 text-gray-400 opacity-0 transition-all hover:bg-gray-100 hover:text-gray-600 group-hover:opacity-100 dark:hover:bg-dark-700" :title="t('models.copyId')">
                <Icon :name="copiedId === item.model_id ? 'check' : 'copy'" size="sm" />
              </button>
            </div>
            <!-- Pricing -->
            <div class="mb-2.5 space-y-0.5 text-xs text-gray-600 dark:text-dark-400">
              <div>{{ t('models.pricing.input') }} <span class="font-medium text-gray-900 dark:text-white">${{ fmtPrice(item.input_price) }}/M</span></div>
              <div>{{ t('models.pricing.output') }} <span class="font-medium text-gray-900 dark:text-white">${{ fmtPrice(item.output_price) }}/M</span></div>
            </div>
            <!-- Tags -->
            <div class="flex flex-wrap items-center gap-1.5">
              <span :class="['rounded px-2 py-0.5 text-[11px] font-medium', providerBadge(item.provider)]">{{ providerLabel(item.provider) }}</span>
              <span class="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-500 dark:bg-dark-700 dark:text-dark-400">{{ item.group_name }}</span>
            </div>
          </div>
        </div>

        <!-- Empty -->
        <div v-else class="py-20 text-center">
          <Icon name="search" size="xl" class="mx-auto mb-4 text-gray-300 dark:text-dark-600" />
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('models.noResults') }}</p>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import { apiClient } from '@/api/client'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => authStore.isAdmin ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => authStore.user?.email?.charAt(0).toUpperCase() || '')

const isDark = ref(document.documentElement.classList.contains('dark'))
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Data
interface ModelEntry {
  model_id: string; display_name: string; provider: string
  group_id: number; group_name: string
  input_price: number; output_price: number
}

const items = ref<ModelEntry[]>([])
const loading = ref(true)
const searchQuery = ref('')
const activeProvider = ref('all')
const activeGroup = ref('')
const sortBy = ref<'default' | 'price-asc' | 'price-desc'>('default')
const copiedId = ref('')

// Derived
const allGroups = computed(() => [...new Set(items.value.map(i => i.group_name))].sort())

const providerTabs = computed(() => {
  const counts: Record<string, number> = {}
  // 按唯一 model_id 计数（去重）
  const seen: Record<string, Set<string>> = {}
  items.value.forEach(i => {
    if (!seen[i.provider]) seen[i.provider] = new Set()
    seen[i.provider].add(i.model_id)
  })
  for (const [p, s] of Object.entries(seen)) counts[p] = s.size
  const total = new Set(items.value.map(i => i.model_id)).size
  return [
    { value: 'all', label: t('models.filters.all'), count: total, icon: null, iconClass: '' },
    { value: 'claude', label: 'Claude', count: counts.claude || 0, icon: 'C', iconClass: 'bg-gradient-to-br from-orange-400 to-orange-500' },
    { value: 'openai', label: 'OpenAI', count: counts.openai || 0, icon: 'G', iconClass: 'bg-gradient-to-br from-green-500 to-green-600' },
    { value: 'gemini', label: 'Google', count: counts.gemini || 0, icon: 'G', iconClass: 'bg-gradient-to-br from-blue-500 to-blue-600' },
  ]
})

const bannerTitle = computed(() => {
  if (activeProvider.value === 'all') return t('models.banner.allProviders')
  return providerTabs.value.find(t => t.value === activeProvider.value)?.label || ''
})

const filtered = computed(() => {
  let r = items.value
  if (activeProvider.value !== 'all') r = r.filter(i => i.provider === activeProvider.value)
  if (activeGroup.value) r = r.filter(i => i.group_name === activeGroup.value)
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    r = r.filter(i => i.model_id.includes(q) || i.display_name.toLowerCase().includes(q))
  }
  if (sortBy.value === 'price-asc') {
    r = [...r].sort((a, b) => a.input_price - b.input_price)
  } else if (sortBy.value === 'price-desc') {
    r = [...r].sort((a, b) => b.input_price - a.input_price)
  }
  return r
})

function toggleSort() {
  if (sortBy.value === 'default') sortBy.value = 'price-asc'
  else if (sortBy.value === 'price-asc') sortBy.value = 'price-desc'
  else sortBy.value = 'default'
}

// Helpers
function fmtPrice(p: number): string {
  if (p < 0.01) return p.toFixed(4)
  if (p < 1) return p.toFixed(3)
  return p.toFixed(2)
}
function providerIconBg(p: string) {
  return p === 'claude' ? 'bg-gradient-to-br from-orange-400 to-orange-500'
    : p === 'openai' ? 'bg-gradient-to-br from-gray-700 to-gray-900 dark:from-gray-500 dark:to-gray-700'
    : p === 'gemini' ? 'bg-gradient-to-br from-blue-500 to-blue-600'
    : 'bg-gradient-to-br from-gray-400 to-gray-500'
}
function providerLetter(p: string) { return p === 'claude' ? 'C' : p === 'openai' ? 'G' : p === 'gemini' ? 'G' : '?' }
function providerLabel(p: string) { return p === 'claude' ? 'Anthropic' : p === 'openai' ? 'OpenAI' : p === 'gemini' ? 'Google' : p }
function providerBadge(p: string) {
  return p === 'claude' ? 'bg-orange-50 text-orange-600 dark:bg-orange-900/20 dark:text-orange-400'
    : p === 'openai' ? 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-dark-300'
    : p === 'gemini' ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/20 dark:text-blue-400'
    : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-400'
}
async function copy(id: string) {
  try { await navigator.clipboard.writeText(id) } catch { return }
  copiedId.value = id
  setTimeout(() => { copiedId.value = '' }, 1500)
}

async function fetchModels() {
  try {
    const res = await apiClient.get('/settings/models')
    items.value = (res.data as ModelEntry[]) || []
  } catch { items.value = [] }
  finally { loading.value = false }
}

onMounted(() => {
  const saved = localStorage.getItem('theme')
  if (saved === 'dark' || (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    isDark.value = true; document.documentElement.classList.add('dark')
  }
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) appStore.fetchPublicSettings()
  fetchModels()
})
</script>

<style scoped>
.nav-tab { padding: 6px 16px; font-size: 14px; font-weight: 500; color: #6b7280; border-radius: 8px; transition: all 0.2s; text-decoration: none; }
.nav-tab:hover { color: #111827; background: rgba(0, 0, 0, 0.04); }
.nav-tab-active { color: #14b8a6; background: rgba(20, 184, 166, 0.08); }
.filter-tag { @apply inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs font-medium text-gray-600 transition-all hover:border-gray-300 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-300; }
.filter-tag-active { @apply border-primary-200 bg-primary-50 text-primary-700 dark:border-primary-800 dark:bg-primary-900/20 dark:text-primary-400; }
:deep(.dark) .nav-tab { color: #9ca3af; }
:deep(.dark) .nav-tab:hover { color: #f3f4f6; background: rgba(255, 255, 255, 0.06); }
:deep(.dark) .nav-tab-active { color: #2dd4bf; background: rgba(20, 184, 166, 0.12); }
</style>
