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
        <!-- Center Nav -->
        <div class="hidden items-center gap-1 sm:flex">
          <router-link to="/home" class="nav-tab">{{ t('nav.home') }}</router-link>
          <router-link to="/models" class="nav-tab nav-tab-active">{{ t('nav.models') }}</router-link>
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="nav-tab">{{ t('nav.docs') }}</a>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="nav-tab">{{ t('nav.console') }}</router-link>
        </div>
        <!-- Right Actions -->
        <div class="flex items-center gap-2">
          <LocaleSwitcher />
          <button @click="toggleTheme" class="rounded-lg p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white">
            <Icon v-if="isDark" name="sun" size="sm" />
            <Icon v-else name="moon" size="sm" />
          </button>
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="flex h-7 w-7 items-center justify-center rounded-full bg-gradient-to-br from-primary-400 to-primary-600 text-[11px] font-semibold text-white"
          >{{ userInitial }}</router-link>
          <router-link v-else to="/login" class="rounded-full bg-primary-500 px-3.5 py-1 text-xs font-medium text-white transition-colors hover:bg-primary-600">{{ t('home.login') }}</router-link>
        </div>
      </nav>
    </header>

    <!-- Body -->
    <div class="mx-auto flex w-full max-w-7xl flex-1 gap-0 px-4 py-6 lg:gap-6 lg:px-6">
      <!-- Sidebar -->
      <aside class="hidden w-56 flex-shrink-0 lg:block">
        <div class="sticky top-20 space-y-6">
          <!-- Provider Filter -->
          <div>
            <h3 class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{{ t('models.sidebar.provider') }}</h3>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="tab in providerTabs"
                :key="tab.value"
                @click="activeProvider = tab.value"
                :class="[
                  'inline-flex items-center gap-1.5 rounded-lg border px-2.5 py-1.5 text-xs font-medium transition-all',
                  activeProvider === tab.value
                    ? 'border-primary-200 bg-primary-50 text-primary-700 dark:border-primary-800 dark:bg-primary-900/20 dark:text-primary-400'
                    : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300 hover:text-gray-900 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-300 dark:hover:border-dark-600 dark:hover:text-white'
                ]"
              >
                <span v-if="tab.icon" :class="['flex h-4 w-4 items-center justify-center rounded', tab.iconClass]">
                  <span class="text-[8px] font-bold text-white">{{ tab.icon }}</span>
                </span>
                <span>{{ tab.label }}</span>
                <span class="text-[10px] opacity-60">{{ tab.count }}</span>
              </button>
            </div>
          </div>
        </div>
      </aside>

      <!-- Main Content -->
      <main class="min-w-0 flex-1">
        <!-- Banner -->
        <div class="relative mb-5 overflow-hidden rounded-2xl bg-gradient-to-r from-indigo-600 via-blue-600 to-cyan-500 px-6 py-5">
          <div class="absolute right-4 top-1/2 -translate-y-1/2 opacity-10">
            <svg class="h-28 w-28 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="0.5">
              <path d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456z" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
          <div class="relative">
            <div class="mb-1 flex items-center gap-3">
              <h1 class="text-xl font-bold text-white">{{ bannerTitle }}</h1>
              <span class="rounded-full bg-white/20 px-2.5 py-0.5 text-xs font-medium text-white backdrop-blur-sm">
                {{ t('models.banner.count', { count: filteredModels.length }) }}
              </span>
            </div>
            <p class="text-sm text-white/70">{{ t('models.banner.description') }}</p>
          </div>
        </div>

        <!-- Mobile Filter + Search -->
        <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <!-- Mobile Provider Tabs -->
          <div class="flex flex-wrap items-center gap-1.5 lg:hidden">
            <button
              v-for="tab in providerTabs"
              :key="'m-' + tab.value"
              @click="activeProvider = tab.value"
              :class="[
                'rounded-lg px-2.5 py-1 text-xs font-medium transition-all',
                activeProvider === tab.value
                  ? 'bg-primary-500 text-white'
                  : 'bg-white text-gray-600 hover:text-gray-900 dark:bg-dark-800 dark:text-dark-300'
              ]"
            >{{ tab.label }} {{ tab.count }}</button>
          </div>

          <!-- Search -->
          <div class="relative w-full sm:max-w-xs">
            <Icon name="search" size="sm" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('models.searchPlaceholder')"
              class="w-full rounded-lg border border-gray-200 bg-white py-2 pl-9 pr-3 text-sm text-gray-900 placeholder-gray-400 transition-all focus:border-primary-300 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-800 dark:text-white dark:placeholder-dark-500"
            />
          </div>
        </div>

        <!-- Loading Skeleton -->
        <div v-if="loading" class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          <div v-for="i in 9" :key="i" class="animate-pulse rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
            <div class="mb-3 flex items-center gap-3">
              <div class="h-9 w-9 rounded-full bg-gray-200 dark:bg-dark-700"></div>
              <div class="h-4 w-32 rounded bg-gray-200 dark:bg-dark-700"></div>
            </div>
            <div class="mb-2 h-3 w-48 rounded bg-gray-100 dark:bg-dark-700/60"></div>
            <div class="h-3 w-36 rounded bg-gray-100 dark:bg-dark-700/60"></div>
          </div>
        </div>

        <!-- Model Grid -->
        <div v-else-if="filteredModels.length" class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          <div
            v-for="model in filteredModels"
            :key="model.id"
            class="group relative rounded-xl border border-gray-200 bg-white p-4 transition-all duration-200 hover:border-gray-300 hover:shadow-md dark:border-dark-700 dark:bg-dark-800 dark:hover:border-dark-600"
          >
            <!-- Top Row: Icon + Name + Actions -->
            <div class="mb-2.5 flex items-start justify-between gap-3">
              <div class="flex items-center gap-3">
                <div :class="['flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-full', providerIconBg(model.provider)]">
                  <span class="text-xs font-bold text-white">{{ providerIcon(model.provider) }}</span>
                </div>
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ model.display_name }}</h3>
              </div>
              <!-- Copy Buttons -->
              <div class="flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                <button
                  @click="copyModelId(model.id)"
                  class="rounded p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-dark-200"
                  :title="t('models.copyId')"
                >
                  <Icon :name="copiedId === model.id ? 'check' : 'copy'" size="sm" />
                </button>
              </div>
            </div>

            <!-- Model ID -->
            <div class="mb-3">
              <code class="text-xs text-gray-500 dark:text-dark-400">{{ model.id }}</code>
            </div>

            <!-- Provider Tag -->
            <div class="flex items-center justify-between">
              <span :class="['rounded px-2 py-0.5 text-[11px] font-medium', providerBadgeClass(model.provider)]">
                {{ providerLabel(model.provider) }}
              </span>
            </div>
          </div>
        </div>

        <!-- Empty State -->
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

// Site settings
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')

// Auth
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => authStore.isAdmin ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Models data
interface PublicModel {
  id: string
  display_name: string
  provider: string
  created_at?: string
}

const models = ref<PublicModel[]>([])
const loading = ref(true)
const searchQuery = ref('')
const activeProvider = ref('all')
const copiedId = ref('')

// Provider counts
const providerCounts = computed(() => {
  const counts: Record<string, number> = {}
  models.value.forEach(m => { counts[m.provider] = (counts[m.provider] || 0) + 1 })
  return counts
})

// Provider tabs
const providerTabs = computed(() => [
  { value: 'all', label: t('models.filters.all'), count: models.value.length, icon: null, iconClass: '' },
  { value: 'claude', label: 'Claude', count: providerCounts.value.claude || 0, icon: 'C', iconClass: 'bg-gradient-to-br from-orange-400 to-orange-500' },
  { value: 'openai', label: 'OpenAI', count: providerCounts.value.openai || 0, icon: 'G', iconClass: 'bg-gradient-to-br from-green-500 to-green-600' },
  { value: 'gemini', label: 'Google', count: providerCounts.value.gemini || 0, icon: 'G', iconClass: 'bg-gradient-to-br from-blue-500 to-blue-600' },
])

// Banner title
const bannerTitle = computed(() => {
  const tab = providerTabs.value.find(t => t.value === activeProvider.value)
  if (activeProvider.value === 'all') return t('models.banner.allProviders')
  return tab?.label || ''
})

// Filtered models
const filteredModels = computed(() => {
  let result = models.value
  if (activeProvider.value !== 'all') {
    result = result.filter(m => m.provider === activeProvider.value)
  }
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(m => m.id.toLowerCase().includes(q) || m.display_name.toLowerCase().includes(q))
  }
  return result
})

// Provider styling
function providerIconBg(provider: string): string {
  switch (provider) {
    case 'claude': return 'bg-gradient-to-br from-orange-400 to-orange-500'
    case 'openai': return 'bg-gradient-to-br from-gray-700 to-gray-900 dark:from-gray-500 dark:to-gray-700'
    case 'gemini': return 'bg-gradient-to-br from-blue-500 to-blue-600'
    default: return 'bg-gradient-to-br from-gray-400 to-gray-500'
  }
}

function providerIcon(provider: string): string {
  switch (provider) {
    case 'claude': return 'C'
    case 'openai': return 'G'
    case 'gemini': return 'G'
    default: return '?'
  }
}

function providerLabel(provider: string): string {
  switch (provider) {
    case 'claude': return 'Anthropic'
    case 'openai': return 'OpenAI'
    case 'gemini': return 'Google'
    default: return provider
  }
}

function providerBadgeClass(provider: string): string {
  switch (provider) {
    case 'claude': return 'bg-orange-50 text-orange-600 dark:bg-orange-900/20 dark:text-orange-400'
    case 'openai': return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-dark-300'
    case 'gemini': return 'bg-blue-50 text-blue-600 dark:bg-blue-900/20 dark:text-blue-400'
    default: return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-400'
  }
}

// Copy
async function copyModelId(id: string) {
  try {
    await navigator.clipboard.writeText(id)
  } catch {
    return
  }
  copiedId.value = id
  setTimeout(() => { copiedId.value = '' }, 1500)
}

// Fetch
async function fetchModels() {
  try {
    const res = await apiClient.get('/settings/models')
    models.value = res.data || []
  } catch {
    models.value = []
  } finally {
    loading.value = false
  }
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
  fetchModels()
})
</script>

<style scoped>
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
