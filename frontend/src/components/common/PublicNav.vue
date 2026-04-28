<template>
  <header class="border-b border-gray-200 dark:border-gray-800">
    <nav class="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
      <router-link to="/home" class="flex items-center gap-2">
        <img src="/logo.png" alt="" class="h-7 w-7" />
        <span class="text-base font-semibold tracking-tight">AIGateway</span>
      </router-link>

      <div class="hidden items-center gap-1 sm:flex">
        <router-link to="/home" :class="linkClass(active === 'home')">{{ t('nav.home') }}</router-link>
        <router-link to="/models" :class="linkClass(active === 'models')">{{ t('nav.models') }}</router-link>
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" :class="linkClass(active === 'docs')">{{ t('nav.docs') }}</a>
        <router-link v-else to="/docs" :class="linkClass(active === 'docs')">{{ t('nav.docs') }}</router-link>
        <router-link :to="isAuthenticated ? dashboardPath : '/login'" :class="linkClass(false)">{{ t('nav.console') }}</router-link>
      </div>

      <div class="flex items-center gap-2">
        <LocaleSwitcher />
        <button
          @click="toggleTheme"
          class="rounded-md p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-gray-800 dark:hover:text-white"
          :aria-label="isDark ? t('home.switchToLight') : t('home.switchToDark')"
        >
          <Icon v-if="isDark" name="sun" size="sm" />
          <Icon v-else name="moon" size="sm" />
        </button>
        <router-link
          v-if="isAuthenticated"
          :to="dashboardPath"
          class="flex h-7 w-7 items-center justify-center rounded-full bg-gray-900 text-[11px] font-semibold text-white dark:bg-white dark:text-gray-900"
        >{{ userInitial }}</router-link>
        <router-link
          v-else
          to="/login"
          class="rounded-md bg-gray-900 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100"
        >{{ t('home.login') }}</router-link>
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

defineProps<{ active?: 'home' | 'models' | 'docs' }>()

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

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

function linkClass(isActive: boolean): string {
  const base = 'rounded-md px-3 py-1.5 text-sm font-medium transition-colors no-underline'
  if (isActive) {
    return `${base} text-gray-900 dark:text-white`
  }
  return `${base} text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-gray-800 dark:hover:text-white`
}
</script>
