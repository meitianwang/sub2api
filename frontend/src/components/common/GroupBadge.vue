<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-md bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-700 transition-colors dark:bg-emerald-900/30 dark:text-emerald-400"
  >
    <span class="truncate">{{ name }}</span>
  </span>
</template>

<script setup lang="ts">
interface Props {
  name: string
}

defineProps<Props>()
</script>
