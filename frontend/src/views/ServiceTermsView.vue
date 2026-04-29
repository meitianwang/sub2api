<template>
  <div class="flex min-h-screen flex-col bg-white text-gray-900 dark:bg-gray-950 dark:text-white">
    <PublicNav />

    <main class="flex-1">
      <article class="service-terms mx-auto max-w-3xl px-6 py-12 sm:py-16">
        <header class="mb-10 border-b border-gray-200 pb-6 dark:border-gray-800">
          <h1 class="text-3xl font-semibold tracking-tight">{{ t('serviceTerms.title') }}</h1>
          <p class="mt-3 text-sm text-gray-500 dark:text-gray-500">{{ t('serviceTerms.effectiveDate') }}</p>
        </header>

        <section>
          <p>{{ t('serviceTerms.intro.p1') }}</p>
          <p>{{ t('serviceTerms.intro.p2') }}</p>
        </section>

        <h2>{{ t('serviceTerms.A.title') }}</h2>
        <p>{{ t('serviceTerms.A.intro') }}</p>
        <ul>
          <li v-for="k in apiLimitItems" :key="k">
            <strong>{{ t(`serviceTerms.A.items.${k}.label`) }}</strong>：{{ t(`serviceTerms.A.items.${k}.body`) }}
          </li>
        </ul>

        <h2>{{ t('serviceTerms.B.title') }}</h2>
        <p>{{ t('serviceTerms.B.p1') }}</p>
        <ul>
          <li v-for="k in routingItems" :key="k">
            <strong>{{ t(`serviceTerms.B.items.${k}.label`) }}</strong>：{{ t(`serviceTerms.B.items.${k}.body`) }}
          </li>
        </ul>

        <h2>{{ t('serviceTerms.C.title') }}</h2>
        <ul>
          <li v-for="k in keyItems" :key="k">
            <strong>{{ t(`serviceTerms.C.items.${k}.label`) }}</strong>：{{ t(`serviceTerms.C.items.${k}.body`) }}
          </li>
        </ul>

        <h2>{{ t('serviceTerms.D.title') }}</h2>
        <ul>
          <li v-for="k in groupItems" :key="k">
            <strong>{{ t(`serviceTerms.D.items.${k}.label`) }}</strong>：{{ t(`serviceTerms.D.items.${k}.body`) }}
          </li>
        </ul>

        <h2>{{ t('serviceTerms.E.title') }}</h2>
        <ul>
          <li v-for="k in billingItems" :key="k">
            <strong>{{ t(`serviceTerms.E.items.${k}.label`) }}</strong>：{{ t(`serviceTerms.E.items.${k}.body`) }}
          </li>
        </ul>

        <h2>{{ t('serviceTerms.F.title') }}</h2>
        <p>{{ t('serviceTerms.F.p1') }}</p>
        <ul>
          <li v-for="k in betaItems" :key="k">{{ t(`serviceTerms.F.items.${k}`) }}</li>
        </ul>
        <p>{{ t('serviceTerms.F.p2') }}</p>

        <h2>{{ t('serviceTerms.G.title') }}</h2>
        <ul>
          <li v-for="k in retentionItems" :key="k">
            <strong>{{ t(`serviceTerms.G.items.${k}.label`) }}</strong>：{{ t(`serviceTerms.G.items.${k}.body`) }}
          </li>
        </ul>

        <h2>{{ t('serviceTerms.H.title') }}</h2>
        <p>{{ t('serviceTerms.H.intro') }}</p>
        <ul>
          <li v-for="k in notProvidedItems" :key="k">
            <strong>{{ t(`serviceTerms.H.items.${k}.label`) }}</strong>：{{ t(`serviceTerms.H.items.${k}.body`) }}
          </li>
        </ul>

        <h2>{{ t('serviceTerms.I.title') }}</h2>
        <p>{{ t('serviceTerms.I.p1') }}</p>
      </article>
    </main>

    <footer class="border-t border-gray-200 dark:border-gray-800">
      <div class="mx-auto flex max-w-5xl flex-col gap-3 px-6 py-8 text-xs text-gray-500 dark:text-gray-500 sm:flex-row sm:items-center sm:justify-between">
        <p>&copy; {{ currentYear }} AIInterface. {{ t('serviceTerms.footer.copyright') }}</p>
        <div class="flex flex-wrap items-center gap-5">
          <router-link to="/home" class="hover:text-gray-900 dark:hover:text-white">{{ t('serviceTerms.footer.home') }}</router-link>
          <router-link to="/docs" class="hover:text-gray-900 dark:hover:text-white">{{ t('serviceTerms.footer.docs') }}</router-link>
          <router-link to="/terms" class="hover:text-gray-900 dark:hover:text-white">{{ t('serviceTerms.footer.terms') }}</router-link>
          <router-link to="/usage-policy" class="hover:text-gray-900 dark:hover:text-white">{{ t('serviceTerms.footer.usagePolicy') }}</router-link>
          <router-link to="/supported-regions" class="hover:text-gray-900 dark:hover:text-white">{{ t('serviceTerms.footer.supportedRegions') }}</router-link>
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

const apiLimitItems = ['rateLimit', 'quota', 'cache', 'monitoring'] as const
const routingItems = ['sharing', 'failover', 'opacity', 'nonDiscrimination'] as const
const keyItems = ['lifecycle', 'propagation', 'rotation'] as const
const groupItems = ['groupBased', 'denied', 'adminChange'] as const
const billingItems = ['priority', 'expiry', 'nonRefundable', 'trial'] as const
const betaItems = ['asIs', 'priceVolatile', 'sunset'] as const
const retentionItems = ['prompts', 'metadata', 'account', 'legalHold', 'noTraining'] as const
const notProvidedItems = ['finetuning', 'selfHosted', 'dataPartner', 'formalSla'] as const
</script>

<style scoped>
.service-terms { @apply text-gray-800 dark:text-gray-200 leading-relaxed; }
.service-terms h2 {
  @apply text-xl font-semibold text-gray-900 dark:text-white mt-10 mb-4 pb-2;
  @apply border-b border-gray-100 dark:border-gray-900;
}
.service-terms p { @apply mb-4 leading-7; }
.service-terms ul { @apply mb-4 list-disc pl-6 space-y-2; }
.service-terms li { @apply leading-7; }
.service-terms strong { @apply font-semibold text-gray-900 dark:text-white; }
</style>
