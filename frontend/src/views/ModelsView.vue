<template>
  <div class="flex min-h-screen flex-col bg-white text-gray-900 dark:bg-gray-950 dark:text-white">
    <PublicNav active="models" />

    <div class="mx-auto flex w-full max-w-5xl flex-1 gap-0 px-6 py-10 lg:gap-8">
      <!-- Sidebar -->
      <aside class="hidden w-48 flex-shrink-0 lg:block">
        <div class="sticky top-6 space-y-7">
          <!-- Provider -->
          <div>
            <h3 class="mb-3 text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-500">{{ t('models.sidebar.provider') }}</h3>
            <div class="flex flex-wrap gap-1.5">
              <button v-for="tab in providerTabs" :key="tab.value" @click="activeProvider = tab.value"
                :class="['filter-tag', activeProvider === tab.value ? 'filter-tag-active' : '']">
                <ProviderBrandIcon v-if="tab.brand" :provider="tab.brand" class="h-3.5 w-3.5" />
                {{ tab.label }} <span class="opacity-60">{{ tab.count }}</span>
              </button>
            </div>
          </div>
          <!-- Group -->
          <div v-if="allGroups.length">
            <h3 class="mb-3 text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-500">{{ t('models.sidebar.group') }}</h3>
            <div class="flex flex-wrap gap-1.5">
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
        <!-- Page header (replaces purple banner) -->
        <div class="mb-6 border-b border-gray-200 pb-5 dark:border-gray-800">
          <h1 class="text-2xl font-semibold tracking-tight">{{ bannerTitle }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-500">
            {{ t('models.banner.count', { count: filtered.length }) }} · {{ t('models.banner.description') }}
          </p>
        </div>

        <!-- Mobile Filters + Search -->
        <div class="mb-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div class="flex flex-wrap items-center gap-1.5 lg:hidden">
            <button v-for="tab in providerTabs" :key="'m-'+tab.value" @click="activeProvider = tab.value"
              :class="['rounded-md px-2.5 py-1 text-xs font-medium transition-colors', activeProvider === tab.value ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900' : 'border border-gray-200 bg-white text-gray-600 dark:border-gray-800 dark:bg-gray-950 dark:text-gray-400']">
              {{ tab.label }} {{ tab.count }}
            </button>
          </div>
          <div class="flex items-center gap-2">
            <div class="relative w-full sm:w-64">
              <Icon name="search" size="sm" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-600" />
              <input v-model="searchQuery" type="text" :placeholder="t('models.searchPlaceholder')"
                class="w-full rounded-md border border-gray-200 bg-white py-2 pl-9 pr-3 font-mono text-sm text-gray-900 placeholder-gray-400 focus:border-gray-400 focus:outline-none focus:ring-0 dark:border-gray-800 dark:bg-gray-950 dark:text-white dark:placeholder-gray-600 dark:focus:border-gray-600" />
            </div>
            <button @click="toggleSort"
              :class="['flex items-center gap-1 rounded-md border px-3 py-2 text-xs font-medium transition-colors', sortBy === 'default' ? 'border-gray-200 bg-white text-gray-500 dark:border-gray-800 dark:bg-gray-950 dark:text-gray-500' : 'border-gray-900 bg-gray-900 text-white dark:border-white dark:bg-white dark:text-gray-900']">
              <Icon :name="sortBy === 'price-desc' ? 'arrowDown' : 'arrowUp'" size="sm" />
              {{ t('models.sort.price') }}
            </button>
          </div>
        </div>

        <!-- Loading -->
        <div v-if="loading" class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
          <div v-for="i in 9" :key="i" class="animate-pulse rounded-lg border border-gray-200 p-4 dark:border-gray-800">
            <div class="mb-3 flex items-center gap-3"><div class="h-8 w-8 rounded-full bg-gray-100 dark:bg-gray-800"></div><div class="h-4 w-32 rounded bg-gray-100 dark:bg-gray-800"></div></div>
            <div class="mb-2 h-3 w-48 rounded bg-gray-100 dark:bg-gray-900"></div>
            <div class="h-3 w-36 rounded bg-gray-100 dark:bg-gray-900"></div>
          </div>
        </div>

        <!-- Cards -->
        <div v-else-if="filtered.length" class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
          <div v-for="item in filtered" :key="item.model_id + '-' + item.group_id"
            @click="openDrawer(item)"
            class="group cursor-pointer rounded-lg border border-gray-200 p-4 transition-colors hover:bg-gray-50 dark:border-gray-800 dark:hover:bg-gray-900/50">
            <!-- Top -->
            <div class="mb-3 flex items-start justify-between gap-3">
              <div class="flex min-w-0 items-center gap-2.5">
                <ProviderBrandIcon :provider="item.provider" circle class="h-8 w-8 flex-shrink-0" />
                <h3 class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ item.display_name }}</h3>
              </div>
              <button @click.stop="copy(item.model_id)" class="rounded p-1 text-gray-400 opacity-0 transition-all hover:bg-gray-100 hover:text-gray-700 group-hover:opacity-100 dark:hover:bg-gray-800 dark:hover:text-gray-300" :title="t('models.copyId')">
                <Icon :name="copiedId === item.model_id ? 'check' : 'copy'" size="sm" />
              </button>
            </div>
            <!-- Pricing (mono, tabular) -->
            <div class="mb-3 space-y-0.5 font-mono text-xs tabular-nums text-gray-600 dark:text-gray-400">
              <div class="flex items-baseline gap-2"><span class="w-12 text-gray-400 dark:text-gray-600">{{ t('models.pricing.input') }}</span><span class="text-gray-900 dark:text-white">¥{{ fmtPrice(item.input_price) }}/M</span></div>
              <div class="flex items-baseline gap-2"><span class="w-12 text-gray-400 dark:text-gray-600">{{ t('models.pricing.output') }}</span><span class="text-gray-900 dark:text-white">¥{{ fmtPrice(item.output_price) }}/M</span></div>
            </div>
            <!-- Tags -->
            <div class="flex flex-wrap items-center gap-1.5">
              <span :class="['rounded px-1.5 py-0.5 text-[10px] font-medium', providerBadge(item.provider)]">{{ providerLabel(item.provider) }}</span>
              <span class="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-500 dark:bg-gray-800 dark:text-gray-400">{{ item.group_name }}</span>
            </div>
          </div>
        </div>

        <!-- Empty -->
        <div v-else class="py-20 text-center">
          <Icon name="search" size="xl" class="mx-auto mb-4 text-gray-300 dark:text-gray-700" />
          <p class="text-sm text-gray-500 dark:text-gray-500">{{ t('models.noResults') }}</p>
        </div>
      </main>
    </div>

    <ModelDetailDrawer :open="drawerOpen" :model="drawerModel" @close="closeDrawer" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import { apiClient } from '@/api/client'
import PublicNav from '@/components/common/PublicNav.vue'
import Icon from '@/components/icons/Icon.vue'
import ProviderBrandIcon from '@/components/icons/ProviderBrandIcon.vue'
import ModelDetailDrawer from '@/components/models/ModelDetailDrawer.vue'
import { providerLabel, fmtPrice, type ModelEntry } from '@/components/models/providerUtils'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const items = ref<ModelEntry[]>([])
const loading = ref(true)
const searchQuery = ref('')
const activeProvider = ref('all')
const activeGroup = ref('')
const sortBy = ref<'default' | 'price-asc' | 'price-desc'>('default')
const copiedId = ref('')

const drawerModel = ref<ModelEntry | null>(null)
const drawerOpen = computed(() => drawerModel.value !== null)
function openDrawer(item: ModelEntry) { drawerModel.value = item }
function closeDrawer() { drawerModel.value = null }

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
    { value: 'all', label: t('models.filters.all'), count: total, brand: '' },
    { value: 'claude', label: 'Claude', count: counts.claude || 0, brand: 'claude' },
    { value: 'openai', label: 'OpenAI', count: counts.openai || 0, brand: 'openai' },
    { value: 'gemini', label: 'Google', count: counts.gemini || 0, brand: 'gemini' },
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

function providerBadge(p: string) {
  return p === 'claude' ? 'bg-orange-50 text-orange-600 dark:bg-orange-950/40 dark:text-orange-400'
    : p === 'openai' ? 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300'
    : p === 'gemini' ? 'bg-blue-50 text-blue-600 dark:bg-blue-950/40 dark:text-blue-400'
    : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
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
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) appStore.fetchPublicSettings()
  fetchModels()
})
</script>

<style scoped>
.filter-tag {
  @apply inline-flex items-center gap-1.5 rounded-md border border-gray-200 bg-white px-2 py-1 text-xs font-medium text-gray-600 transition-colors;
  @apply hover:border-gray-300 hover:text-gray-900;
  @apply dark:border-gray-800 dark:bg-gray-950 dark:text-gray-400;
  @apply dark:hover:border-gray-700 dark:hover:text-white;
}
.filter-tag-active {
  @apply border-gray-900 bg-gray-900 text-white;
  @apply hover:border-gray-900 hover:text-white;
  @apply dark:border-white dark:bg-white dark:text-gray-900;
  @apply dark:hover:border-white dark:hover:text-gray-900;
}
</style>
