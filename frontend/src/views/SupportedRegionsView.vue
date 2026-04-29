<template>
  <div class="flex min-h-screen flex-col bg-white text-gray-900 dark:bg-gray-950 dark:text-white">
    <PublicNav />

    <main class="flex-1">
      <article class="regions mx-auto max-w-3xl px-6 py-12 sm:py-16">
        <header class="mb-10 border-b border-gray-200 pb-6 dark:border-gray-800">
          <h1 class="text-3xl font-semibold tracking-tight">{{ t('supportedRegions.title') }}</h1>
          <p class="mt-3 text-sm text-gray-500 dark:text-gray-500">{{ t('supportedRegions.effectiveDate') }}</p>
        </header>

        <section>
          <p>{{ t('supportedRegions.intro') }}</p>
        </section>

        <h2>{{ t('supportedRegions.general.title') }}</h2>
        <p>{{ t('supportedRegions.general.body') }}</p>

        <h2>{{ t('supportedRegions.restricted.title') }}</h2>
        <p>{{ t('supportedRegions.restricted.intro') }}</p>
        <ul>
          <li v-for="k in restrictedList" :key="k">{{ t(`supportedRegions.restricted.list.${k}`) }}</li>
        </ul>
        <p>{{ t('supportedRegions.restricted.footer') }}</p>

        <h2>{{ t('supportedRegions.upstream.title') }}</h2>
        <p>{{ t('supportedRegions.upstream.p1') }}</p>
        <p>{{ t('supportedRegions.upstream.p2') }}</p>
        <ul>
          <li v-for="k in upstreamList" :key="k">{{ t(`supportedRegions.upstream.list.${k}`) }}</li>
        </ul>
        <p>{{ t('supportedRegions.upstream.p3') }}</p>

        <h2>{{ t('supportedRegions.obligation.title') }}</h2>
        <p>{{ t('supportedRegions.obligation.body') }}</p>

        <h2>{{ t('supportedRegions.updates.title') }}</h2>
        <p>{{ t('supportedRegions.updates.body') }}</p>

        <h2>{{ t('supportedRegions.contact.title') }}</h2>
        <p>
          {{ t('supportedRegions.contact.body') }}
          <code class="font-mono text-sm">support@aiinterface.store</code>
        </p>
      </article>
    </main>

    <footer class="border-t border-gray-200 dark:border-gray-800">
      <div class="mx-auto flex max-w-5xl flex-col gap-3 px-6 py-8 text-xs text-gray-500 dark:text-gray-500 sm:flex-row sm:items-center sm:justify-between">
        <p>&copy; {{ currentYear }} AIInterface. {{ t('supportedRegions.footer.copyright') }}</p>
        <div class="flex flex-wrap items-center gap-5">
          <router-link to="/home" class="hover:text-gray-900 dark:hover:text-white">{{ t('supportedRegions.footer.home') }}</router-link>
          <router-link to="/docs" class="hover:text-gray-900 dark:hover:text-white">{{ t('supportedRegions.footer.docs') }}</router-link>
          <router-link to="/terms" class="hover:text-gray-900 dark:hover:text-white">{{ t('supportedRegions.footer.terms') }}</router-link>
          <router-link to="/usage-policy" class="hover:text-gray-900 dark:hover:text-white">{{ t('supportedRegions.footer.usagePolicy') }}</router-link>
          <router-link to="/service-specific-terms" class="hover:text-gray-900 dark:hover:text-white">{{ t('home.serviceTerms') }}</router-link>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import PublicNav from '@/components/common/PublicNav.vue'

const { t } = useI18n()
const currentYear = computed(() => new Date().getFullYear())

const restrictedList = ['singapore', 'un', 'local', 'upstream'] as const
const upstreamList = ['ipBlock', 'rateLimit', 'contentReview'] as const
</script>

<style scoped>
.regions { @apply text-gray-800 dark:text-gray-200 leading-relaxed; }
.regions h2 {
  @apply text-xl font-semibold text-gray-900 dark:text-white mt-10 mb-4 pb-2;
  @apply border-b border-gray-100 dark:border-gray-900;
}
.regions p { @apply mb-4 leading-7; }
.regions ul { @apply mb-4 list-disc pl-6 space-y-2; }
.regions li { @apply leading-7; }
.regions code { @apply bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-1.5 py-0.5 rounded text-sm; }
</style>
