<template>
  <div class="relative flex min-h-screen items-center justify-center overflow-hidden p-4 sm:p-6">
    <!-- Atmospheric background -->
    <div
      class="absolute inset-0 bg-gradient-to-br from-slate-50 via-blue-50/40 to-violet-50/30 dark:from-gray-950 dark:via-gray-950 dark:to-gray-950"
    ></div>

    <!-- Tri-color orbs (echo logo's blue/violet/green circuit) -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <!-- Blue orb top-right -->
      <div
        class="absolute -right-32 -top-32 h-[28rem] w-[28rem] rounded-full bg-blue-400/25 blur-3xl dark:bg-blue-500/15"
      ></div>
      <!-- Violet orb bottom-left -->
      <div
        class="absolute -bottom-40 -left-32 h-[28rem] w-[28rem] rounded-full bg-violet-400/25 blur-3xl dark:bg-violet-500/15"
      ></div>
      <!-- Emerald accent mid-right (smaller, brand "accent" tier) -->
      <div
        class="absolute right-[15%] top-[40%] h-64 w-64 rounded-full bg-emerald-400/15 blur-3xl dark:bg-emerald-500/10"
      ></div>

      <!-- Subtle grid -->
      <div
        class="absolute inset-0 bg-[linear-gradient(rgba(99,102,241,0.04)_1px,transparent_1px),linear-gradient(90deg,rgba(99,102,241,0.04)_1px,transparent_1px)] bg-[size:48px_48px] dark:bg-[linear-gradient(rgba(139,92,246,0.05)_1px,transparent_1px),linear-gradient(90deg,rgba(139,92,246,0.05)_1px,transparent_1px)]"
      ></div>
    </div>

    <!-- Content -->
    <div class="relative z-10 w-full max-w-md">
      <!-- Brand mark -->
      <div class="mb-7 text-center">
        <RouterLink
          to="/home"
          class="group inline-flex flex-col items-center gap-3 rounded-2xl px-3 py-2 transition focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 focus-visible:ring-offset-2"
          :aria-label="siteName"
        >
          <div class="relative">
            <!-- soft glow under logo -->
            <div class="absolute inset-0 rounded-2xl bg-gradient-to-br from-blue-500/20 via-violet-500/20 to-emerald-400/15 blur-xl transition-opacity group-hover:opacity-80"></div>
            <img
              src="/logo.png"
              alt=""
              class="relative h-16 w-16 transition-transform group-hover:scale-[1.03]"
            />
          </div>
          <div class="flex flex-col items-center gap-1">
            <span class="bg-gradient-to-r from-blue-600 via-violet-600 to-violet-700 bg-clip-text text-2xl font-bold tracking-tight text-transparent dark:from-blue-400 dark:via-violet-400 dark:to-violet-300">
              {{ siteName }}
            </span>
            <span v-if="siteSubtitle" class="text-xs text-gray-500 dark:text-gray-400">
              {{ siteSubtitle }}
            </span>
          </div>
        </RouterLink>
      </div>

      <!-- Glass form card -->
      <div
        class="relative rounded-2xl border border-white/60 bg-white/70 p-7 shadow-[0_8px_32px_rgba(15,23,42,0.08)] backdrop-blur-xl dark:border-gray-800/60 dark:bg-gray-900/60 dark:shadow-[0_8px_32px_rgba(0,0,0,0.35)] sm:p-8"
      >
        <slot />
      </div>

      <!-- Footer slot -->
      <div v-if="$slots.footer" class="mt-6 text-center text-sm">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <p class="mt-6 text-center text-xs text-gray-400 dark:text-gray-600">
        &copy; {{ currentYear }} {{ siteName }}.
        <RouterLink to="/terms" class="ml-1 hover:text-gray-600 dark:hover:text-gray-400">{{ t('home.terms') }}</RouterLink>
        <span class="mx-1.5 text-gray-300 dark:text-gray-700">·</span>
        <RouterLink to="/usage-policy" class="hover:text-gray-600 dark:hover:text-gray-400">{{ t('home.usagePolicy') }}</RouterLink>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'AIGateway')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Precision infrastructure for the AI era')
const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  if (!appStore.publicSettingsLoaded) appStore.fetchPublicSettings()
})
</script>
