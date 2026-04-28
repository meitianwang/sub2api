<template>
  <!-- Custom Home Content: Full Page Mode (admin override) -->
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page -->
  <div v-else class="flex min-h-screen flex-col bg-white text-gray-900 dark:bg-gray-950 dark:text-white">
    <PublicNav active="home" />

    <!-- Main -->
    <main class="flex-1">
      <div class="mx-auto max-w-5xl px-6">

        <!-- Hero -->
        <section class="py-20 sm:py-28">
          <h1 class="text-4xl font-semibold tracking-tight sm:text-6xl">
            {{ t('home.hero.title') }}
          </h1>
          <p class="mt-5 max-w-2xl text-base text-gray-600 dark:text-gray-400 sm:text-lg">
            {{ heroSubtitle }}
          </p>
          <div class="mt-8 flex items-center gap-5">
            <router-link
              :to="isAuthenticated ? dashboardPath : '/login'"
              class="inline-flex items-center gap-1.5 rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white transition-shadow hover:shadow-[0_4px_20px_-4px_rgba(59,130,246,0.5)] dark:bg-white dark:text-gray-900 dark:hover:shadow-[0_4px_20px_-4px_rgba(139,92,246,0.4)]"
            >
              {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
              <Icon name="arrowRight" size="sm" />
            </router-link>
            <router-link
              to="/docs"
              v-if="!docUrl"
              class="text-sm text-gray-600 underline-offset-4 hover:text-gray-900 hover:underline dark:text-gray-400 dark:hover:text-white"
            >{{ t('home.viewDocs') }}</router-link>
            <a
              v-else
              :href="docUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="text-sm text-gray-600 underline-offset-4 hover:text-gray-900 hover:underline dark:text-gray-400 dark:hover:text-white"
            >{{ t('home.viewDocs') }}</a>
          </div>
        </section>

        <!-- Curl example -->
        <section class="border-t border-gray-200 py-16 dark:border-gray-800">
          <div class="mb-5 flex items-baseline justify-between gap-4">
            <h2 class="text-xl font-semibold tracking-tight">{{ t('home.curl.title') }}</h2>
            <p class="hidden text-sm text-gray-500 dark:text-gray-500 sm:block">{{ t('home.curl.description') }}</p>
          </div>
          <div class="overflow-hidden rounded-lg border border-gray-200 bg-gray-50 dark:border-gray-800 dark:bg-gray-900/50">
            <div class="flex items-center justify-between border-b border-gray-200 px-4 py-2 dark:border-gray-800">
              <code class="font-mono text-xs text-gray-500 dark:text-gray-400">POST /v1/messages</code>
              <button
                @click="copyCurl"
                class="inline-flex items-center gap-1 rounded px-2 py-1 text-xs font-medium text-gray-600 transition-colors hover:bg-gray-200 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-gray-800 dark:hover:text-white"
              >
                <Icon :name="curlCopied ? 'check' : 'copy'" size="sm" />
                <span>{{ curlCopied ? t('home.curl.copied') : t('home.curl.copy') }}</span>
              </button>
            </div>
            <pre class="overflow-x-auto p-4 font-mono text-sm leading-relaxed text-gray-800 dark:text-gray-200"><code>{{ curlExample.code }}</code></pre>
          </div>
        </section>

        <!-- Pricing -->
        <section class="border-t border-gray-200 py-16 dark:border-gray-800">
          <div class="mb-5 flex items-baseline justify-between gap-4">
            <h2 class="text-xl font-semibold tracking-tight">{{ t('home.pricing.title') }}</h2>
            <router-link to="/models" class="text-sm text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white">
              {{ t('home.pricing.realtimeLink') }}
            </router-link>
          </div>

          <div v-if="pricingLoading" class="space-y-2">
            <div v-for="i in 6" :key="i" class="h-9 animate-pulse rounded bg-gray-100 dark:bg-gray-900" />
          </div>

          <div v-else-if="pricingError" class="rounded-lg border border-gray-200 px-4 py-6 text-center dark:border-gray-800">
            <p class="text-sm text-gray-500 dark:text-gray-500">{{ t('home.pricing.loadFailed') }}</p>
            <button @click="reloadPricing" class="mt-2 text-sm text-gray-700 underline underline-offset-4 hover:text-gray-900 dark:text-gray-300 dark:hover:text-white">{{ t('home.pricing.retry') }}</button>
          </div>

          <div v-else>
            <div class="grid grid-cols-[24px_1fr_auto] items-center gap-3 border-b border-gray-200 pb-2 text-xs font-medium uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:text-gray-500">
              <span></span>
              <span>Model</span>
              <span class="font-mono normal-case tracking-normal">{{ t('home.pricing.perMillion') }}</span>
            </div>
            <ul>
              <li v-for="m in pricingTopRows" :key="m.model_id + '-' + m.group_id">
                <router-link
                  to="/models"
                  class="grid grid-cols-[24px_1fr_auto] items-center gap-3 rounded px-2 py-2.5 -mx-2 transition-colors hover:bg-gray-50 dark:hover:bg-gray-900/50"
                >
                  <ProviderBrandIcon :provider="m.provider" circle class="h-5 w-5" />
                  <code class="truncate font-mono text-sm text-gray-800 dark:text-gray-200">{{ m.model_id }}</code>
                  <span class="font-mono text-sm tabular-nums text-gray-600 dark:text-gray-400">¥{{ fmtPrice(m.input_price) }} / ¥{{ fmtPrice(m.output_price) }}</span>
                </router-link>
              </li>
            </ul>
            <router-link
              v-if="uniqueModelCount > pricingTopRows.length"
              to="/models"
              class="mt-4 inline-block text-sm text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
            >{{ t('home.pricing.viewAll', { count: uniqueModelCount }) }}</router-link>
          </div>
        </section>

        <!-- 3 value lines -->
        <section class="border-t border-gray-200 py-16 dark:border-gray-800">
          <ul class="space-y-4">
            <li class="flex items-baseline gap-3">
              <span class="mt-2 h-2 w-2 flex-shrink-0 rounded-full bg-blue-500"></span>
              <span class="text-base text-gray-700 dark:text-gray-300">{{ t('home.values.routing') }}</span>
            </li>
            <li class="flex items-baseline gap-3">
              <span class="mt-2 h-2 w-2 flex-shrink-0 rounded-full bg-violet-500"></span>
              <span class="text-base text-gray-700 dark:text-gray-300">{{ t('home.values.billing') }}</span>
            </li>
            <li class="flex items-baseline gap-3">
              <span class="mt-2 h-2 w-2 flex-shrink-0 rounded-full bg-emerald-500"></span>
              <span class="text-base text-gray-700 dark:text-gray-300">{{ t('home.values.compat') }}</span>
            </li>
          </ul>
        </section>

      </div>
    </main>

    <!-- Footer -->
    <footer class="border-t border-gray-200 dark:border-gray-800">
      <div class="mx-auto flex max-w-5xl flex-col gap-3 px-6 py-8 text-xs text-gray-500 dark:text-gray-500 sm:flex-row sm:items-center sm:justify-between">
        <p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
        <div class="flex items-center gap-5">
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="hover:text-gray-900 dark:hover:text-white">{{ t('home.docs') }}</a>
          <router-link v-else to="/docs" class="hover:text-gray-900 dark:hover:text-white">{{ t('home.docs') }}</router-link>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import PublicNav from '@/components/common/PublicNav.vue'
import Icon from '@/components/icons/Icon.vue'
import ProviderBrandIcon from '@/components/icons/ProviderBrandIcon.vue'
import { usePublicModels } from '@/composables/usePublicModels'
import { fmtPrice } from '@/components/models/providerUtils'
import { buildCurlExample } from '@/components/models/buildCurlExample'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'AIInterface')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => authStore.isAdmin ? '/admin/dashboard' : '/dashboard')
const currentYear = computed(() => new Date().getFullYear())

// Public models (shared with /models page via composable)
const { items: models, loading: pricingLoading, error: pricingError, fetch: fetchModels, uniqueModelCount } = usePublicModels()

const pricingTopRows = computed(() => {
  const seen = new Set<string>()
  const rows = []
  for (const m of [...models.value].sort((a, b) => b.input_price - a.input_price)) {
    if (seen.has(m.model_id)) continue
    seen.add(m.model_id)
    rows.push(m)
    if (rows.length >= 8) break
  }
  return rows
})

const heroSubtitle = computed(() => {
  if (uniqueModelCount.value > 0) {
    return t('home.hero.subtitle', { count: uniqueModelCount.value })
  }
  return t('home.hero.subtitleFallback')
})

// Curl example: default to a Claude Sonnet call (Anthropic format)
const curlExample = computed(() => {
  const apiBaseUrl = appStore.cachedPublicSettings?.api_base_url || ''
  return buildCurlExample('claude-sonnet-4-6', 'claude', apiBaseUrl)
})

const curlCopied = ref(false)

async function writeText(text: string): Promise<boolean> {
  if (navigator.clipboard && window.isSecureContext) {
    try { await navigator.clipboard.writeText(text); return true } catch { /* fall through */ }
  }
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.style.cssText = 'position:fixed;left:-9999px;top:-9999px'
  document.body.appendChild(textarea)
  textarea.select()
  try { return document.execCommand('copy') } finally { document.body.removeChild(textarea) }
}

async function copyCurl() {
  if (!curlExample.value.code) return
  if (await writeText(curlExample.value.code)) {
    curlCopied.value = true
    setTimeout(() => { curlCopied.value = false }, 1500)
  }
}

function reloadPricing() {
  void fetchModels(true)
}

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) appStore.fetchPublicSettings()
  void fetchModels()
})
</script>
