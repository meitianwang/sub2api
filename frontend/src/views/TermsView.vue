<template>
  <div class="flex min-h-screen flex-col bg-white text-gray-900 dark:bg-gray-950 dark:text-white">
    <PublicNav />

    <main class="flex-1">
      <article class="terms mx-auto max-w-3xl px-6 py-12 sm:py-16">
        <header class="mb-10 border-b border-gray-200 pb-6 dark:border-gray-800">
          <h1 class="text-3xl font-semibold tracking-tight">{{ t('terms.title') }}</h1>
          <p class="mt-3 text-sm text-gray-500 dark:text-gray-500">
            {{ t('terms.effectiveDate') }}
          </p>
        </header>

        <section>
          <p>{{ t('terms.intro.welcome') }}</p>
          <p>{{ t('terms.intro.notTargeted') }}</p>
        </section>

        <h2>{{ t('terms.A.title') }}</h2>
        <p>{{ t('terms.A.p1') }}</p>
        <p>{{ t('terms.A.p2') }}</p>

        <h2>{{ t('terms.B.title') }}</h2>
        <p>{{ t('terms.B.p1') }}</p>
        <p>{{ t('terms.B.p2') }}</p>

        <h2>{{ t('terms.C.title') }}</h2>
        <p>{{ t('terms.C.p1') }}</p>
        <p>{{ t('terms.C.p2') }}</p>

        <h2>{{ t('terms.D.title') }}</h2>
        <ul>
          <li v-for="key in refundItems" :key="key">
            <strong>{{ t(`terms.D.items.${key}.label`) }}</strong>：{{ t(`terms.D.items.${key}.body`) }}
          </li>
        </ul>

        <h2>{{ t('terms.E.title') }}</h2>
        <p>{{ t('terms.E.intro') }}</p>
        <ul>
          <li v-for="key in prohibitedActs" :key="key">{{ t(`terms.E.list.${key}`) }}</li>
        </ul>
        <p>{{ t('terms.E.outputDisclaimer') }}</p>

        <h2>{{ t('terms.F.title') }}</h2>
        <p>{{ t('terms.F.p1') }}</p>
        <p>{{ t('terms.F.p2') }}</p>

        <h2>{{ t('terms.G.title') }}</h2>
        <p>{{ t('terms.G.p1') }}</p>

        <h2>{{ t('terms.H.title') }}</h2>
        <p>{{ t('terms.H.p1') }}</p>
        <p>{{ t('terms.H.p2') }}</p>

        <h2>{{ t('terms.I.title') }}</h2>
        <p>{{ t('terms.I.p1') }}</p>

        <h2>{{ t('terms.J.title') }}</h2>
        <p>{{ t('terms.J.p1') }}</p>
        <p>{{ t('terms.J.p2') }}</p>

        <h2>{{ t('terms.K.title') }}</h2>
        <p>{{ t('terms.K.p1') }}</p>

        <h2>{{ t('terms.L.title') }}</h2>
        <p>{{ t('terms.L.p1') }}</p>
        <p>{{ t('terms.L.p2') }}</p>
        <p>{{ t('terms.L.p3') }}</p>

        <h2>{{ t('terms.M.title') }}</h2>
        <p>{{ t('terms.M.p1') }}</p>
        <ul>
          <li>{{ t('terms.M.email') }}: <code class="font-mono text-sm">support@aiinterface.store</code></li>
        </ul>
      </article>
    </main>

    <footer class="border-t border-gray-200 dark:border-gray-800">
      <div class="mx-auto flex max-w-5xl flex-col gap-3 px-6 py-8 text-xs text-gray-500 dark:text-gray-500 sm:flex-row sm:items-center sm:justify-between">
        <p>&copy; {{ currentYear }} AIInterface. {{ t('terms.footer.copyright') }}</p>
        <div class="flex items-center gap-5">
          <router-link to="/home" class="hover:text-gray-900 dark:hover:text-white">{{ t('terms.footer.home') }}</router-link>
          <router-link to="/docs" class="hover:text-gray-900 dark:hover:text-white">{{ t('terms.footer.docs') }}</router-link>
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

const refundItems = ['scope', 'method', 'review', 'antiFraud', 'finalInterpretation'] as const
const prohibitedActs = ['laws', 'harmfulContent', 'reverseEngineering', 'compete', 'attack', 'ip', 'other'] as const
</script>

<style scoped>
.terms { @apply text-gray-800 dark:text-gray-200 leading-relaxed; }
.terms h2 {
  @apply text-xl font-semibold text-gray-900 dark:text-white mt-10 mb-4 pb-2;
  @apply border-b border-gray-100 dark:border-gray-900;
}
.terms p { @apply mb-4 leading-7; }
.terms ul { @apply mb-4 list-disc pl-6 space-y-2; }
.terms li { @apply leading-7; }
.terms strong { @apply font-semibold text-gray-900 dark:text-white; }
.terms code { @apply bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-1.5 py-0.5 rounded text-sm; }
</style>
