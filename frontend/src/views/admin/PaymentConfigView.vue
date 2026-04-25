<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Tabs -->
      <div class="flex gap-1 overflow-x-auto border-b border-gray-200 dark:border-dark-700">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          :class="[
            'whitespace-nowrap border-b-2 px-4 py-2.5 text-sm font-medium transition-colors',
            activeTab === tab.key
              ? 'border-primary-500 text-primary-600 dark:border-primary-400 dark:text-primary-400'
              : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
          ]"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- Tab: Dashboard -->
      <div v-show="activeTab === 'dashboard'">
        <div v-if="dashboardLoading" class="flex items-center justify-center py-12">
          <LoadingSpinner />
        </div>
        <template v-else-if="dashboard">
          <!-- Stats Cards -->
          <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
            <div class="card"><div class="p-4">
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.payment.todayAmount') }}</p>
              <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ dashboard.today_amount }}</p>
            </div></div>
            <div class="card"><div class="p-4">
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.payment.todayOrders') }}</p>
              <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ dashboard.today_order_count }}</p>
            </div></div>
            <div class="card"><div class="p-4">
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.payment.totalAmount') }}</p>
              <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ dashboard.total_amount }}</p>
            </div></div>
            <div class="card"><div class="p-4">
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.payment.totalOrders') }}</p>
              <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ dashboard.total_order_count }}</p>
            </div></div>
          </div>

          <!-- Daily Chart -->
          <div class="card mt-6"><div class="p-6">
            <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.payment.dailyTrend') }}
            </h3>
            <div v-if="dashboard.daily_series.length > 0" class="overflow-x-auto">
              <canvas ref="chartCanvas" height="250"></canvas>
            </div>
            <p v-else class="py-8 text-center text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.payment.noData') }}
            </p>
          </div></div>

          <!-- Payment Methods -->
          <div v-if="dashboard.payment_methods.length > 0" class="card mt-6"><div class="p-6">
            <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.payment.paymentMethods') }}
            </h3>
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
              <div
                v-for="method in dashboard.payment_methods"
                :key="method.type"
                class="rounded-lg border border-gray-200 p-3 dark:border-dark-600"
              >
                <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ paymentTypeLabel(method.type) }}</p>
                <p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">{{ method.amount }}</p>
                <p class="text-xs text-gray-400">{{ method.count }} {{ t('admin.payment.orders') }}</p>
              </div>
            </div>
          </div></div>
        </template>
      </div>

      <!-- Tab: Settings -->
      <div v-show="activeTab === 'settings'">
        <div class="card"><div class="p-6">
          <div class="mb-6 flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.payment.configSettings') }}
            </h3>
            <button
              @click="loadConfig"
              :disabled="configLoading"
              class="btn btn-secondary"
            >
              <Icon name="refresh" size="sm" :class="configLoading ? 'animate-spin' : ''" />
            </button>
          </div>

          <div v-if="configLoading" class="flex items-center justify-center py-8">
            <LoadingSpinner />
          </div>
          <form v-else @submit.prevent="saveConfig" class="space-y-4">
            <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
              <div v-for="(_value, key) in configSettings" :key="key">
                <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ key }}</label>
                <input
                  v-model="configSettings[key]"
                  :type="isSensitiveKey(key) ? 'password' : 'text'"
                  class="input w-full font-mono text-sm"
                />
              </div>
            </div>
            <div class="flex justify-end pt-2">
              <button type="submit" :disabled="savingConfig" class="btn btn-primary">
                {{ savingConfig ? t('common.saving') : t('common.save') }}
              </button>
            </div>
          </form>
        </div></div>
      </div>

      <!-- Tab: Provider Instances -->
      <div v-show="activeTab === 'providers'">
        <div class="card">
          <div class="flex items-center justify-between border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.payment.providerInstances') }}
            </h3>
            <div class="flex gap-2">
              <button @click="loadProviderInstances" :disabled="providersLoading" class="btn btn-secondary">
                <Icon name="refresh" size="sm" :class="providersLoading ? 'animate-spin' : ''" />
              </button>
              <button @click="openProviderDialog()" class="btn btn-primary">
                <Icon name="plus" size="sm" class="mr-1" />
                {{ t('common.create') }}
              </button>
            </div>
          </div>

          <DataTable :columns="providerColumns" :data="providerInstances" :loading="providersLoading">
            <template #cell-provider_key="{ value }">
              <span :class="['badge', providerBadgeClass(value)]">{{ value }}</span>
            </template>

            <template #cell-supported_types="{ value }">
              <span class="text-sm text-gray-700 dark:text-gray-300">{{ value || '-' }}</span>
            </template>

            <template #cell-enabled="{ value }">
              <span :class="['badge', value ? 'badge-success' : 'badge-gray']">
                {{ value ? t('common.enabled') : t('common.disabled') }}
              </span>
            </template>

            <template #cell-refund_enabled="{ value }">
              <span :class="['badge', value ? 'badge-success' : 'badge-gray']">
                {{ value ? t('common.yes') : t('common.no') }}
              </span>
            </template>

            <template #cell-actions="{ row }">
              <div class="flex items-center space-x-2">
                <button
                  @click="openProviderDialog(row)"
                  class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-600 dark:hover:text-gray-300"
                >
                  <Icon name="edit" size="sm" />
                </button>
                <button
                  @click="handleDeleteProvider(row)"
                  class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                >
                  <Icon name="trash" size="sm" />
                </button>
              </div>
            </template>
          </DataTable>
        </div>
      </div>

      <!-- Tab: Channels -->
      <div v-show="activeTab === 'channels'">
        <div class="card">
          <div class="flex items-center justify-between border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.payment.channels') }}
            </h3>
            <div class="flex gap-2">
              <button @click="loadChannels" :disabled="channelsLoading" class="btn btn-secondary">
                <Icon name="refresh" size="sm" :class="channelsLoading ? 'animate-spin' : ''" />
              </button>
              <button @click="openChannelDialog()" class="btn btn-primary">
                <Icon name="plus" size="sm" class="mr-1" />
                {{ t('common.create') }}
              </button>
            </div>
          </div>

          <DataTable :columns="channelColumns" :data="channels" :loading="channelsLoading">
            <template #cell-group_id="{ value }">
              <span class="font-mono text-sm">{{ value ?? '-' }}</span>
            </template>

            <template #cell-rate_multiplier="{ value }">
              <span class="text-sm text-gray-700 dark:text-gray-300">{{ value }}x</span>
            </template>

            <template #cell-enabled="{ value }">
              <span :class="['badge', value ? 'badge-success' : 'badge-gray']">
                {{ value ? t('common.enabled') : t('common.disabled') }}
              </span>
            </template>

            <template #cell-actions="{ row }">
              <div class="flex items-center space-x-2">
                <button
                  @click="openChannelDialog(row)"
                  class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-600 dark:hover:text-gray-300"
                >
                  <Icon name="edit" size="sm" />
                </button>
                <button
                  @click="handleDeleteChannel(row)"
                  class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                >
                  <Icon name="trash" size="sm" />
                </button>
              </div>
            </template>
          </DataTable>
        </div>
      </div>

    </div>

    <!-- Provider Instance Dialog -->
    <BaseDialog
      :show="showProviderDialog"
      :title="editingProvider ? t('admin.payment.editProvider') : t('admin.payment.createProvider')"
      width="wide"
      @close="showProviderDialog = false"
    >
      <form id="provider-form" @submit.prevent="saveProvider" class="space-y-4">
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.payment.providerKey') }}</label>
            <Select
              v-model="providerForm.provider_key"
              :options="providerKeyOptions"
              :disabled="!!editingProvider"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.payment.name') }}</label>
            <input v-model="providerForm.name" type="text" required class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.payment.supportedTypes') }}</label>
            <input
              v-model="providerForm.supported_types"
              type="text"
              class="input"
              placeholder="alipay,wxpay"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.payment.sortOrder') }}</label>
            <input v-model.number="providerForm.sort_order" type="number" class="input" />
          </div>
          <div class="flex items-center gap-4">
            <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
              <Toggle v-model="providerForm.enabled" />
              {{ t('common.enabled') }}
            </label>
            <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
              <Toggle v-model="providerForm.refund_enabled" />
              {{ t('admin.payment.refundEnabled') }}
            </label>
          </div>
        </div>

        <!-- Config Fields -->
        <div>
          <label class="input-label">{{ t('admin.payment.configFields') }}</label>
          <div class="space-y-2">
            <div
              v-for="(_value, key) in providerForm.config"
              :key="key"
              class="flex items-center gap-2"
            >
              <input :value="key" disabled class="input w-40 bg-gray-50 font-mono text-xs dark:bg-dark-800" />
              <input
                v-model="providerForm.config[key]"
                :type="isSensitiveKey(String(key)) ? 'password' : 'text'"
                class="input flex-1 font-mono text-sm"
              />
              <button
                type="button"
                @click="removeConfigField(String(key))"
                class="rounded p-1 text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20"
              >
                <Icon name="x" size="sm" />
              </button>
            </div>
            <div class="flex items-center gap-2">
              <input
                v-model="newConfigKey"
                type="text"
                class="input w-40 font-mono text-xs"
                placeholder="key"
              />
              <input
                v-model="newConfigValue"
                type="text"
                class="input flex-1 font-mono text-sm"
                placeholder="value"
              />
              <button
                type="button"
                @click="addConfigField"
                class="btn btn-secondary"
              >
                <Icon name="plus" size="sm" />
              </button>
            </div>
          </div>
        </div>

        <!-- Limits JSON -->
        <div>
          <label class="input-label">
            {{ t('admin.payment.limits') }}
            <span class="ml-1 text-xs font-normal text-gray-400">(JSON, {{ t('common.optional') }})</span>
          </label>
          <textarea
            v-model="providerForm.limits"
            rows="3"
            class="input font-mono text-sm"
            placeholder='{"daily_amount":"10000","single_min":"1","single_max":"5000"}'
          ></textarea>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" @click="showProviderDialog = false" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button type="submit" form="provider-form" :disabled="savingProvider" class="btn btn-primary">
            {{ savingProvider ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Channel Dialog -->
    <BaseDialog
      :show="showChannelDialog"
      :title="editingChannel ? t('admin.payment.editChannel') : t('admin.payment.createChannel')"
      width="normal"
      @close="showChannelDialog = false"
    >
      <form id="channel-form" @submit.prevent="saveChannel" class="space-y-4">
        <div>
          <label class="input-label">{{ t('admin.payment.name') }}</label>
          <input v-model="channelForm.name" type="text" required class="input" />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="input-label">{{ t('admin.payment.groupId') }}</label>
            <input v-model.number="channelForm.group_id" type="number" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.payment.platform') }}</label>
            <input v-model="channelForm.platform" type="text" required class="input" placeholder="web / mobile" />
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="input-label">{{ t('admin.payment.rateMultiplier') }}</label>
            <input v-model="channelForm.rate_multiplier" type="text" class="input" placeholder="1.0" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.payment.sortOrder') }}</label>
            <input v-model.number="channelForm.sort_order" type="number" class="input" />
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('admin.payment.description') }}</label>
          <textarea v-model="channelForm.description" rows="2" class="input"></textarea>
        </div>
        <div>
          <label class="input-label">{{ t('admin.payment.models') }}</label>
          <input v-model="channelForm.models" type="text" class="input" placeholder="gpt-4,claude-3" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.payment.features') }}</label>
          <input v-model="channelForm.features" type="text" class="input" placeholder="unlimited,priority" />
        </div>
        <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
          <Toggle v-model="channelForm.enabled" />
          {{ t('common.enabled') }}
        </label>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" @click="showChannelDialog = false" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button type="submit" form="channel-form" :disabled="savingChannel" class="btn btn-primary">
            {{ savingChannel ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Delete Confirm -->
    <ConfirmDialog
      :show="showDeleteConfirm"
      :title="t('common.delete')"
      :message="deleteConfirmMessage"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDelete"
      @cancel="showDeleteConfirm = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { paymentAdminAPI } from '@/api/admin/payment'
import type { PaymentChannel, ProviderInstance, PaymentDashboardStats } from '@/types/payment'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

type TabKey = 'dashboard' | 'settings' | 'providers' | 'channels'
const activeTab = ref<TabKey>('dashboard')

const tabs = [
  { key: 'dashboard' as TabKey, label: t('admin.payment.dashboard') },
  { key: 'settings' as TabKey, label: t('admin.payment.settings') },
  { key: 'providers' as TabKey, label: t('admin.payment.providers') },
  { key: 'channels' as TabKey, label: t('admin.payment.channelsTab') }
]

// ==================== Dashboard ====================

const dashboardLoading = ref(false)
const dashboard = ref<PaymentDashboardStats | null>(null)
const chartCanvas = ref<HTMLCanvasElement | null>(null)
let chartInstance: any = null

async function loadDashboard() {
  dashboardLoading.value = true
  try {
    dashboard.value = await paymentAdminAPI.getDashboard(30)
    await nextTick()
    renderChart()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.loadDashboardFailed'))
  } finally {
    dashboardLoading.value = false
  }
}

async function renderChart() {
  if (!chartCanvas.value || !dashboard.value?.daily_series.length) return

  try {
    const { Chart, registerables } = await import('chart.js')
    Chart.register(...registerables)

    if (chartInstance) chartInstance.destroy()

    const series = dashboard.value.daily_series
    const isDark = document.documentElement.classList.contains('dark')
    const gridColor = isDark ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.1)'
    const textColor = isDark ? '#9ca3af' : '#6b7280'

    chartInstance = new Chart(chartCanvas.value, {
      type: 'bar',
      data: {
        labels: series.map(d => d.date),
        datasets: [
          {
            label: t('admin.payment.amount'),
            data: series.map(d => Number(d.amount)),
            backgroundColor: isDark ? 'rgba(139,92,246,0.6)' : 'rgba(139,92,246,0.7)',
            borderRadius: 4,
            yAxisID: 'y'
          },
          {
            label: t('admin.payment.orderCount'),
            data: series.map(d => d.count),
            type: 'line' as const,
            borderColor: isDark ? '#60a5fa' : '#3b82f6',
            backgroundColor: 'transparent',
            borderWidth: 2,
            pointRadius: 3,
            yAxisID: 'y1'
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: { mode: 'index', intersect: false },
        scales: {
          x: { ticks: { color: textColor }, grid: { color: gridColor } },
          y: {
            position: 'left',
            ticks: { color: textColor },
            grid: { color: gridColor },
            title: { display: true, text: t('admin.payment.amount'), color: textColor }
          },
          y1: {
            position: 'right',
            ticks: { color: textColor },
            grid: { drawOnChartArea: false },
            title: { display: true, text: t('admin.payment.orderCount'), color: textColor }
          }
        },
        plugins: {
          legend: { labels: { color: textColor } }
        }
      }
    })
  } catch {
    // chart.js not available, skip rendering
  }
}

// ==================== Config Settings ====================

const configLoading = ref(false)
const savingConfig = ref(false)
const configSettings = reactive<Record<string, string>>({})

async function loadConfig() {
  configLoading.value = true
  try {
    const data = await paymentAdminAPI.getConfig()
    // The admin GET config endpoint returns { configs: { key: value } }
    const configs = (data as any).configs || (data as any).settings || data
    Object.keys(configSettings).forEach(k => delete configSettings[k])
    if (typeof configs === 'object' && configs !== null) {
      Object.assign(configSettings, configs)
    }
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.loadConfigFailed'))
  } finally {
    configLoading.value = false
  }
}

async function saveConfig() {
  savingConfig.value = true
  try {
    await paymentAdminAPI.updateConfig(configSettings)
    appStore.showSuccess(t('admin.payment.configSaved'))
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.saveConfigFailed'))
  } finally {
    savingConfig.value = false
  }
}

function isSensitiveKey(key: string): boolean {
  const upper = key.toUpperCase()
  return upper.includes('KEY') || upper.includes('SECRET') || upper.includes('PASSWORD') || upper.includes('PRIVATE')
}

// ==================== Provider Instances ====================

const providersLoading = ref(false)
const savingProvider = ref(false)
const providerInstances = ref<ProviderInstance[]>([])
const showProviderDialog = ref(false)
const editingProvider = ref<ProviderInstance | null>(null)

const providerForm = reactive({
  provider_key: 'easypay',
  name: '',
  config: {} as Record<string, string>,
  supported_types: '',
  enabled: true,
  sort_order: 0,
  limits: '',
  refund_enabled: false
})

const newConfigKey = ref('')
const newConfigValue = ref('')

const providerKeyOptions = [
  { value: 'easypay', label: 'EasyPay' },
  { value: 'alipay', label: 'Alipay (直连)' },
  { value: 'wxpay', label: 'WeChat Pay (直连)' },
  { value: 'stripe', label: 'Stripe' }
]

const providerColumns = computed<Column[]>(() => [
  { key: 'id', label: 'ID', sortable: true },
  { key: 'name', label: t('admin.payment.name') },
  { key: 'provider_key', label: t('admin.payment.providerKey') },
  { key: 'supported_types', label: t('admin.payment.supportedTypes') },
  { key: 'enabled', label: t('common.enabled') },
  { key: 'refund_enabled', label: t('admin.payment.refundEnabled') },
  { key: 'actions', label: t('common.actions') }
])

function providerBadgeClass(key: string): string {
  const map: Record<string, string> = {
    easypay: 'badge-info',
    alipay: 'badge-primary',
    wxpay: 'badge-success',
    stripe: 'badge-warning'
  }
  return map[key] || 'badge-gray'
}

async function loadProviderInstances() {
  providersLoading.value = true
  try {
    providerInstances.value = await paymentAdminAPI.listProviderInstances()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.loadProvidersFailed'))
  } finally {
    providersLoading.value = false
  }
}

function openProviderDialog(instance?: ProviderInstance) {
  if (instance) {
    editingProvider.value = instance
    providerForm.provider_key = instance.provider_key
    providerForm.name = instance.name
    providerForm.config = { ...(instance.config as Record<string, string>) }
    providerForm.supported_types = Array.isArray(instance.supported_types)
      ? instance.supported_types.join(',')
      : String(instance.supported_types || '')
    providerForm.enabled = instance.enabled
    providerForm.sort_order = instance.sort_order
    providerForm.limits = instance.limits ? JSON.stringify(instance.limits) : ''
    providerForm.refund_enabled = instance.refund_enabled
  } else {
    editingProvider.value = null
    providerForm.provider_key = 'easypay'
    providerForm.name = ''
    providerForm.config = {}
    providerForm.supported_types = ''
    providerForm.enabled = true
    providerForm.sort_order = 0
    providerForm.limits = ''
    providerForm.refund_enabled = false
  }
  newConfigKey.value = ''
  newConfigValue.value = ''
  showProviderDialog.value = true
}

function addConfigField() {
  if (newConfigKey.value.trim()) {
    providerForm.config[newConfigKey.value.trim()] = newConfigValue.value
    newConfigKey.value = ''
    newConfigValue.value = ''
  }
}

function removeConfigField(key: string) {
  delete providerForm.config[key]
}

async function saveProvider() {
  savingProvider.value = true
  try {
    const payload: any = {
      provider_key: providerForm.provider_key,
      name: providerForm.name,
      config: providerForm.config,
      supported_types: providerForm.supported_types,
      enabled: providerForm.enabled,
      sort_order: providerForm.sort_order,
      refund_enabled: providerForm.refund_enabled
    }
    if (providerForm.limits.trim()) {
      payload.limits = providerForm.limits.trim()
    }

    if (editingProvider.value) {
      await paymentAdminAPI.updateProviderInstance(editingProvider.value.id, payload)
    } else {
      await paymentAdminAPI.createProviderInstance(payload)
    }
    appStore.showSuccess(t('admin.payment.providerSaved'))
    showProviderDialog.value = false
    loadProviderInstances()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.saveProviderFailed'))
  } finally {
    savingProvider.value = false
  }
}

// ==================== Channels ====================

const channelsLoading = ref(false)
const savingChannel = ref(false)
const channels = ref<PaymentChannel[]>([])
const showChannelDialog = ref(false)
const editingChannel = ref<PaymentChannel | null>(null)

const channelForm = reactive({
  group_id: null as number | null,
  name: '',
  platform: '',
  rate_multiplier: '1',
  description: '',
  models: '',
  features: '',
  sort_order: 0,
  enabled: true
})

const channelColumns = computed<Column[]>(() => [
  { key: 'id', label: 'ID', sortable: true },
  { key: 'name', label: t('admin.payment.name') },
  { key: 'group_id', label: t('admin.payment.groupId') },
  { key: 'platform', label: t('admin.payment.platform') },
  { key: 'rate_multiplier', label: t('admin.payment.rateMultiplier') },
  { key: 'enabled', label: t('common.enabled') },
  { key: 'actions', label: t('common.actions') }
])

async function loadChannels() {
  channelsLoading.value = true
  try {
    channels.value = await paymentAdminAPI.listChannels()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.loadChannelsFailed'))
  } finally {
    channelsLoading.value = false
  }
}

function openChannelDialog(channel?: PaymentChannel) {
  if (channel) {
    editingChannel.value = channel
    channelForm.group_id = channel.group_id
    channelForm.name = channel.name
    channelForm.platform = channel.platform
    channelForm.rate_multiplier = String(channel.rate_multiplier)
    channelForm.description = channel.description || ''
    channelForm.models = Array.isArray(channel.models) ? channel.models.join(',') : String(channel.models || '')
    channelForm.features = Array.isArray(channel.features) ? channel.features.join(',') : String(channel.features || '')
    channelForm.sort_order = channel.sort_order
    channelForm.enabled = channel.enabled
  } else {
    editingChannel.value = null
    channelForm.group_id = null
    channelForm.name = ''
    channelForm.platform = ''
    channelForm.rate_multiplier = '1'
    channelForm.description = ''
    channelForm.models = ''
    channelForm.features = ''
    channelForm.sort_order = 0
    channelForm.enabled = true
  }
  showChannelDialog.value = true
}

async function saveChannel() {
  savingChannel.value = true
  try {
    const payload: any = {
      name: channelForm.name,
      platform: channelForm.platform,
      rate_multiplier: channelForm.rate_multiplier,
      description: channelForm.description || undefined,
      models: channelForm.models || undefined,
      features: channelForm.features || undefined,
      sort_order: channelForm.sort_order,
      enabled: channelForm.enabled
    }
    if (channelForm.group_id !== null) payload.group_id = channelForm.group_id

    if (editingChannel.value) {
      await paymentAdminAPI.updateChannel(editingChannel.value.id, payload)
    } else {
      await paymentAdminAPI.createChannel(payload)
    }
    appStore.showSuccess(t('admin.payment.channelSaved'))
    showChannelDialog.value = false
    loadChannels()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.saveChannelFailed'))
  } finally {
    savingChannel.value = false
  }
}

// ==================== Delete ====================

const showDeleteConfirm = ref(false)
const deleteConfirmMessage = ref('')
let deleteAction: (() => Promise<void>) | null = null

function handleDeleteProvider(instance: ProviderInstance) {
  deleteConfirmMessage.value = t(
    'admin.payment.deleteProviderConfirm'
  )
  deleteAction = async () => {
    await paymentAdminAPI.deleteProviderInstance(instance.id)
    loadProviderInstances()
  }
  showDeleteConfirm.value = true
}

function handleDeleteChannel(channel: PaymentChannel) {
  deleteConfirmMessage.value = t(
    'admin.payment.deleteChannelConfirm'
  )
  deleteAction = async () => {
    await paymentAdminAPI.deleteChannel(channel.id)
    loadChannels()
  }
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (!deleteAction) return
  try {
    await deleteAction()
    appStore.showSuccess(t('admin.payment.deleted'))
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.deleteFailed'))
  } finally {
    showDeleteConfirm.value = false
    deleteAction = null
  }
}

// ==================== Helpers ====================

function paymentTypeLabel(type: string): string {
  const map: Record<string, string> = {
    alipay: '支付宝 / Alipay',
    wxpay: '微信 / WeChat',
    stripe: 'Stripe',
    usdt: 'USDT'
  }
  return map[type] || type
}

// ==================== Tab Loading ====================

watch(activeTab, (tab) => {
  if (tab === 'dashboard' && !dashboard.value) loadDashboard()
  if (tab === 'settings' && Object.keys(configSettings).length === 0) loadConfig()
  if (tab === 'providers' && providerInstances.value.length === 0) loadProviderInstances()
  if (tab === 'channels' && channels.value.length === 0) loadChannels()
})

onMounted(() => {
  loadDashboard()
})
</script>
