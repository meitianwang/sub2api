<template>
  <div class="space-y-3">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.groups.modelPricing.title') }}</h4>
    </div>

    <!-- Pricing Table -->
    <div v-if="entries.length > 0" class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-600">
      <table class="min-w-full text-sm">
        <thead class="bg-gray-50 dark:bg-dark-700">
          <tr>
            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.groups.modelPricing.modelName') }}</th>
            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.groups.modelPricing.sellInput') }}</th>
            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.groups.modelPricing.sellOutput') }}</th>
            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.groups.modelPricing.costInput') }}</th>
            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.groups.modelPricing.costOutput') }}</th>
            <th class="px-3 py-2 text-center text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.groups.modelPricing.actions') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 dark:divide-dark-600">
          <tr v-for="(entry, index) in entries" :key="index">
            <td class="px-3 py-1.5">
              <input
                v-model="entry.model"
                type="text"
                class="input text-sm w-full min-w-[140px]"
                placeholder="claude-sonnet-4-20250514"
              />
            </td>
            <td class="px-3 py-1.5">
              <input
                v-model.number="entry.sell_input_price"
                type="number"
                step="0.01"
                min="0"
                class="input text-sm w-full min-w-[80px]"
                placeholder="3.00"
                @blur="clampPrice(entry, 'sell_input_price')"
              />
            </td>
            <td class="px-3 py-1.5">
              <input
                v-model.number="entry.sell_output_price"
                type="number"
                step="0.01"
                min="0"
                class="input text-sm w-full min-w-[80px]"
                placeholder="15.00"
                @blur="clampPrice(entry, 'sell_output_price')"
              />
            </td>
            <td class="px-3 py-1.5">
              <input
                v-model.number="entry.cost_input_price"
                type="number"
                step="0.01"
                min="0"
                class="input text-sm w-full min-w-[80px]"
                placeholder="3.00"
                @blur="clampPrice(entry, 'cost_input_price')"
              />
            </td>
            <td class="px-3 py-1.5">
              <input
                v-model.number="entry.cost_output_price"
                type="number"
                step="0.01"
                min="0"
                class="input text-sm w-full min-w-[80px]"
                placeholder="15.00"
                @blur="clampPrice(entry, 'cost_output_price')"
              />
            </td>
            <td class="px-3 py-1.5 text-center">
              <button
                type="button"
                @click="removeEntry(index)"
                class="p-1.5 text-gray-400 hover:text-red-500 transition-colors"
                :title="t('admin.groups.modelPricing.delete')"
              >
                <Icon name="trash" size="sm" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <!-- 空模型名警告 -->
    <div v-if="hasEmptyModelNames" class="rounded-lg border border-amber-200 bg-amber-50 p-3 dark:border-amber-800 dark:bg-amber-900/20">
      <p class="text-sm text-amber-700 dark:text-amber-300">
        {{ t('admin.groups.modelPricing.emptyModelWarning') }}
      </p>
    </div>
    <!-- 亏损警告 -->
    <div v-if="hasUnprofitableModels" class="rounded-lg border border-red-200 bg-red-50 p-3 dark:border-red-800 dark:bg-red-900/20">
      <p class="text-sm text-red-700 dark:text-red-300">
        {{ t('admin.groups.modelPricing.unprofitableWarning') }}
      </p>
    </div>

    <p v-else-if="entries.length === 0" class="text-xs text-gray-500 dark:text-gray-400">
      {{ t('admin.groups.modelPricing.emptyHint') }}
    </p>

    <!-- Add Button -->
    <button
      type="button"
      @click="addEntry"
      class="flex items-center gap-1.5 text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
    >
      <Icon name="plus" size="sm" />
      {{ t('admin.groups.modelPricing.addModel') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { AdminGroup } from '@/types'

const { t } = useI18n()

interface PricingEntry {
  model: string
  sell_input_price: number
  sell_output_price: number
  cost_input_price: number
  cost_output_price: number
}

type ModelPricingMap = NonNullable<AdminGroup['model_pricing']>

const props = defineProps<{
  modelValue: ModelPricingMap | undefined | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: ModelPricingMap | undefined]
}>()

const entries = ref<PricingEntry[]>([])
let updatingFromProp = false

// Initialize from prop
const initFromProp = () => {
  if (props.modelValue && Object.keys(props.modelValue).length > 0) {
    updatingFromProp = true
    entries.value = Object.entries(props.modelValue).map(([model, pricing]) => ({
      model,
      sell_input_price: pricing.sell_input_price,
      sell_output_price: pricing.sell_output_price,
      cost_input_price: pricing.cost_input_price,
      cost_output_price: pricing.cost_output_price,
    }))
    updatingFromProp = false
  } else if (entries.value.length > 0) {
    updatingFromProp = true
    entries.value = []
    updatingFromProp = false
  }
}

initFromProp()

// Watch for external changes to modelValue (only on mount/reset, not on self-emit)
watch(() => props.modelValue, (newVal, oldVal) => {
  if (newVal === oldVal) return
  initFromProp()
})

// Emit changes when entries change
const emitUpdate = () => {
  if (updatingFromProp) return
  if (entries.value.length === 0) {
    emit('update:modelValue', undefined)
    return
  }
  const result: ModelPricingMap = {}
  for (const entry of entries.value) {
    const model = entry.model.trim()
    if (model) {
      result[model] = {
        sell_input_price: Math.max(0, entry.sell_input_price || 0),
        sell_output_price: Math.max(0, entry.sell_output_price || 0),
        cost_input_price: Math.max(0, entry.cost_input_price || 0),
        cost_output_price: Math.max(0, entry.cost_output_price || 0),
      }
    }
  }
  emit('update:modelValue', Object.keys(result).length > 0 ? result : undefined)
}

watch(entries, emitUpdate, { deep: true })

// 检查是否有空模型名
const hasEmptyModelNames = computed(() => {
  return entries.value.some(e => !e.model.trim())
})

// 检查是否有卖价低于成本的模型
const hasUnprofitableModels = computed(() => {
  return entries.value.some(e =>
    (e.sell_input_price > 0 && e.cost_input_price > 0 && e.sell_input_price < e.cost_input_price) ||
    (e.sell_output_price > 0 && e.cost_output_price > 0 && e.sell_output_price < e.cost_output_price)
  )
})

const clampPrice = (entry: PricingEntry, field: keyof Omit<PricingEntry, 'model'>) => {
  const val = entry[field]
  if (typeof val !== 'number' || isNaN(val) || val < 0) {
    entry[field] = 0
  }
}

const addEntry = () => {
  entries.value.push({
    model: '',
    sell_input_price: 0,
    sell_output_price: 0,
    cost_input_price: 0,
    cost_output_price: 0,
  })
}

const removeEntry = (index: number) => {
  entries.value.splice(index, 1)
}
</script>
