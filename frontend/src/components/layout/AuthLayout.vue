<template>
  <div class="flex min-h-screen flex-col bg-white text-gray-900 dark:bg-gray-950 dark:text-white">
    <PublicNav />

    <main class="flex-1">
      <div class="mx-auto w-full max-w-md px-6 py-16 sm:py-24">
        <slot />

        <div
          v-if="$slots.footer"
          class="mt-10 border-t border-gray-200 pt-6 text-sm text-gray-600 dark:border-gray-800 dark:text-gray-400"
        >
          <slot name="footer" />
        </div>
      </div>
    </main>

    <footer class="border-t border-gray-200 dark:border-gray-800">
      <div
        class="mx-auto flex max-w-5xl flex-col gap-3 px-6 py-8 text-xs text-gray-500 dark:text-gray-500 sm:flex-row sm:items-center sm:justify-between"
      >
        <p>&copy; {{ currentYear }} {{ siteName }}.</p>
        <div class="flex items-center gap-5">
          <router-link to="/terms" class="hover:text-gray-900 dark:hover:text-white">{{ t('home.terms') }}</router-link>
          <router-link to="/usage-policy" class="hover:text-gray-900 dark:hover:text-white">{{ t('home.usagePolicy') }}</router-link>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import PublicNav from '@/components/common/PublicNav.vue'

const { t } = useI18n()
const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'AIInterface')
const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  if (!appStore.publicSettingsLoaded) appStore.fetchPublicSettings()
})
</script>
