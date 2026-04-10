<template>
    <div class="space-y-6">
      <div class="card p-6">
        <div class="text-center text-sm text-gray-500 dark:text-gray-400">
          {{ t('common.noData') }}
        </div>
      </div>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>
