<template>
  <div class="flex min-h-screen flex-col bg-white text-gray-900 dark:bg-gray-950 dark:text-white">
    <PublicNav />

    <main class="flex-1">
      <article class="policy mx-auto max-w-3xl px-6 py-12 sm:py-16">
        <header class="mb-10 border-b border-gray-200 pb-6 dark:border-gray-800">
          <h1 class="text-3xl font-semibold tracking-tight">{{ t('usagePolicy.title') }}</h1>
          <p class="mt-3 text-sm text-gray-500 dark:text-gray-500">{{ t('usagePolicy.effectiveDate') }}</p>
        </header>

        <section>
          <p>{{ t('usagePolicy.intro.p1') }}</p>
          <p>{{ t('usagePolicy.intro.p2') }}</p>
        </section>

        <h2>{{ t('usagePolicy.general.title') }}</h2>
        <p class="text-gray-600 dark:text-gray-400">{{ t('usagePolicy.general.intro') }}</p>

        <template v-for="key in generalSimpleItems" :key="key">
          <h3>{{ t(`usagePolicy.general.items.${key}.title`) }}</h3>
          <p>{{ t(`usagePolicy.general.items.${key}.body`) }}</p>
        </template>

        <h3>{{ t('usagePolicy.general.items.fraud.title') }}</h3>
        <p>{{ t('usagePolicy.general.items.fraud.intro') }}</p>
        <ul>
          <li v-for="k in fraudList" :key="k">{{ t(`usagePolicy.general.items.fraud.list.${k}`) }}</li>
        </ul>

        <h3>{{ t('usagePolicy.general.items.platformAbuse.title') }}</h3>
        <p>{{ t('usagePolicy.general.items.platformAbuse.intro') }}</p>
        <ul>
          <li v-for="k in platformAbuseList" :key="k">{{ t(`usagePolicy.general.items.platformAbuse.list.${k}`) }}</li>
        </ul>

        <h3>{{ t('usagePolicy.general.items.sexual.title') }}</h3>
        <p>{{ t('usagePolicy.general.items.sexual.body') }}</p>

        <h2>{{ t('usagePolicy.highRisk.title') }}</h2>
        <p>{{ t('usagePolicy.highRisk.intro') }}</p>
        <ul>
          <li>
            <strong>{{ t('usagePolicy.highRisk.humanInLoop.label') }}</strong>：{{ t('usagePolicy.highRisk.humanInLoop.body') }}
          </li>
          <li>
            <strong>{{ t('usagePolicy.highRisk.disclosure.label') }}</strong>：{{ t('usagePolicy.highRisk.disclosure.body') }}
          </li>
        </ul>
        <p>{{ t('usagePolicy.highRisk.domainsTitle') }}</p>
        <ul>
          <li v-for="k in highRiskDomains" :key="k">
            <strong>{{ t(`usagePolicy.highRisk.domains.${k}.label`) }}</strong>：{{ t(`usagePolicy.highRisk.domains.${k}.body`) }}
          </li>
        </ul>

        <h2>{{ t('usagePolicy.transparency.title') }}</h2>
        <template v-for="k in transparencyItems" :key="k">
          <h3>{{ t(`usagePolicy.transparency.${k}.title`) }}</h3>
          <p>{{ t(`usagePolicy.transparency.${k}.body`) }}</p>
        </template>

        <h2>{{ t('usagePolicy.upstream.title') }}</h2>
        <p>{{ t('usagePolicy.upstream.p1') }}</p>
        <p>{{ t('usagePolicy.upstream.p2') }}</p>

        <h2>{{ t('usagePolicy.consequences.title') }}</h2>
        <p>{{ t('usagePolicy.consequences.p1') }}</p>
        <ul>
          <li v-for="k in consequencesList" :key="k">{{ t(`usagePolicy.consequences.list.${k}`) }}</li>
        </ul>
        <p>{{ t('usagePolicy.consequences.p2') }}</p>

        <h2>{{ t('usagePolicy.revisions.title') }}</h2>
        <p>{{ t('usagePolicy.revisions.p1') }}</p>
      </article>
    </main>

    <footer class="border-t border-gray-200 dark:border-gray-800">
      <div class="mx-auto flex max-w-5xl flex-col gap-3 px-6 py-8 text-xs text-gray-500 dark:text-gray-500 sm:flex-row sm:items-center sm:justify-between">
        <p>&copy; {{ currentYear }} AIInterface. {{ t('usagePolicy.footer.copyright') }}</p>
        <div class="flex items-center gap-5">
          <router-link to="/home" class="hover:text-gray-900 dark:hover:text-white">{{ t('usagePolicy.footer.home') }}</router-link>
          <router-link to="/docs" class="hover:text-gray-900 dark:hover:text-white">{{ t('usagePolicy.footer.docs') }}</router-link>
          <router-link to="/terms" class="hover:text-gray-900 dark:hover:text-white">{{ t('usagePolicy.footer.terms') }}</router-link>
          <router-link to="/supported-regions" class="hover:text-gray-900 dark:hover:text-white">{{ t('home.supportedRegions') }}</router-link>
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

const generalSimpleItems = [
  'laws', 'criticalInfra', 'cyberSecurity', 'weapons', 'violence',
  'privacy', 'childSafety', 'psychological', 'misinformation', 'politics',
  'surveillance',
] as const
const fraudList = ['fraudulent', 'predatory', 'spam', 'deception'] as const
const platformAbuseList = ['bypass', 'multiAccount', 'shareKey', 'scrape', 'extract'] as const
const highRiskDomains = [
  'legal', 'healthcare', 'insurance', 'financial', 'employment', 'academic', 'media',
] as const
const transparencyItems = ['chatbots', 'agents', 'minors'] as const
const consequencesList = [
  'warning', 'suspend', 'terminate', 'refundDeny', 'lawEnforcement', 'damages',
] as const
</script>

<style scoped>
.policy { @apply text-gray-800 dark:text-gray-200 leading-relaxed; }
.policy h2 {
  @apply text-xl font-semibold text-gray-900 dark:text-white mt-12 mb-4 pb-2;
  @apply border-b border-gray-100 dark:border-gray-900;
}
.policy h3 {
  @apply text-base font-semibold text-gray-900 dark:text-white mt-6 mb-2;
}
.policy p { @apply mb-4 leading-7; }
.policy ul { @apply mb-4 list-disc pl-6 space-y-2; }
.policy li { @apply leading-7; }
.policy strong { @apply font-semibold text-gray-900 dark:text-white; }
</style>
