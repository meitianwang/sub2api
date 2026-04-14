<template>
  <div class="stat-card">
    <!-- Top row: label + icon -->
    <div class="flex items-start justify-between gap-2">
      <p class="stat-label">{{ title }}</p>
      <div v-if="icon" :class="['stat-icon flex-shrink-0', iconClass]">
        <component :is="icon" class="h-4 w-4" aria-hidden="true" />
      </div>
    </div>
    <!-- Value -->
    <p class="stat-value" :title="String(formattedValue)">{{ formattedValue }}</p>
    <!-- Trend -->
    <span v-if="change !== undefined" :class="['stat-trend', trendClass]">
      <Icon
        v-if="changeType !== 'neutral'"
        name="arrowUp"
        size="xs"
        :class="changeType === 'down' && 'rotate-180'"
      />
      {{ formattedChange }}
    </span>
    <span v-else class="stat-trend opacity-0">—</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Component } from 'vue'
import Icon from '@/components/icons/Icon.vue'

type ChangeType = 'up' | 'down' | 'neutral'
type IconVariant = 'primary' | 'success' | 'warning' | 'danger'

interface Props {
  title: string
  value: number | string
  icon?: Component
  iconVariant?: IconVariant
  change?: number
  changeType?: ChangeType
  formatValue?: (value: number | string) => string
}

const props = withDefaults(defineProps<Props>(), {
  changeType: 'neutral',
  iconVariant: 'primary'
})

const formattedValue = computed(() => {
  if (props.formatValue) {
    return props.formatValue(props.value)
  }
  if (typeof props.value === 'number') {
    return props.value.toLocaleString()
  }
  return props.value
})

const formattedChange = computed(() => {
  if (props.change === undefined) return ''
  const absChange = Math.abs(props.change)
  return `${absChange}%`
})

const iconClass = computed(() => {
  const classes: Record<IconVariant, string> = {
    primary: 'stat-icon-primary',
    success: 'stat-icon-success',
    warning: 'stat-icon-warning',
    danger: 'stat-icon-danger'
  }
  return classes[props.iconVariant]
})

const trendClass = computed(() => {
  const classes: Record<ChangeType, string> = {
    up: 'stat-trend-up',
    down: 'stat-trend-down',
    neutral: 'text-gray-500 dark:text-dark-400'
  }
  return classes[props.changeType]
})
</script>
