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
            <div class="card p-4">
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.payment.todayAmount', '今日金额') }}</p>
              <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ dashboard.today_amount }}</p>
            </div>
            <div class="card p-4">
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.payment.todayOrders', '今日订单') }}</p>
              <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ dashboard.today_order_count }}</p>
            </div>
            <div class="card p-4">
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.payment.totalAmount', '总金额') }}</p>
              <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ dashboard.total_amount }}</p>
            </div>
            <div class="card p-4">
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.payment.totalOrders', '总订单') }}</p>
              <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ dashboard.total_order_count }}</p>
            </div>
          </div>

          <!-- Daily Chart -->
          <div class="card mt-6 p-6">
            <h3 class="mb-4 text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.payment.dailyTrend', '每日趋势 / Daily Trend') }}
            </h3>
            <div v-if="dashboard.daily_series.length > 0" class="overflow-x-auto">
              <canvas ref="chartCanvas" height="250"></canvas>
            </div>
            <p v-else class="py-8 text-center text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.payment.noData', '暂无数据') }}
            </p>
          </div>

          <!-- Payment Methods -->
          <div v-if="dashboard.payment_methods.length > 0" class="card mt-6 p-6">
            <h3 class="mb-4 text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.payment.paymentMethods', '支付方式统计') }}
            </h3>
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
              <div
                v-for="method in dashboard.payment_methods"
                :key="method.type"
                class="rounded-lg border border-gray-200 p-3 dark:border-dark-600"
              >
                <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ paymentTypeLabel(method.type) }}</p>
                <p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">{{ method.amount }}</p>
                <p class="text-xs text-gray-400">{{ method.count }} {{ t('admin.payment.orders', '笔') }}</p>
              </div>
            </div>
          </div>
        </template>
      </div>

      <!-- Tab: Settings -->
      <div v-show="activeTab === 'settings'">
        <div class="card p-6">
          <div class="mb-6 flex items-center justify-between">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.payment.configSettings', '支付配置 / Payment Settings') }}
            </h3>
            <button
              @click="loadConfig"
              :disabled="configLoading"
              class="btn btn-secondary btn-sm"
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
                {{ savingConfig ? t('common.saving', '保存中...') : t('common.save', '保存') }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Tab: Provider Instances -->
      <div v-show="activeTab === 'providers'">
        <div class="card">
          <div class="flex items-center justify-between border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.payment.providerInstances', '支付渠道实例 / Provider Instances') }}
            </h3>
            <div class="flex gap-2">
              <button @click="loadProviderInstances" :disabled="providersLoading" class="btn btn-secondary btn-sm">
                <Icon name="refresh" size="sm" :class="providersLoading ? 'animate-spin' : ''" />
              </button>
              <button @click="openProviderDialog()" class="btn btn-primary btn-sm">
                <Icon name="plus" size="sm" class="mr-1" />
                {{ t('common.create', '创建') }}
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
                {{ value ? t('common.enabled', '启用') : t('common.disabled', '禁用') }}
              </span>
            </template>

            <template #cell-refund_enabled="{ value }">
              <span :class="['badge', value ? 'badge-success' : 'badge-gray']">
                {{ value ? t('common.yes', '是') : t('common.no', '否') }}
              </span>
            </template>

            <template #cell-actions="{ row }">
              <div class="flex items-center gap-1">
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
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.payment.channels', '支付渠道 / Channels') }}
            </h3>
            <div class="flex gap-2">
              <button @click="loadChannels" :disabled="channelsLoading" class="btn btn-secondary btn-sm">
                <Icon name="refresh" size="sm" :class="channelsLoading ? 'animate-spin' : ''" />
              </button>
              <button @click="openChannelDialog()" class="btn btn-primary btn-sm">
                <Icon name="plus" size="sm" class="mr-1" />
                {{ t('common.create', '创建') }}
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
                {{ value ? t('common.enabled', '启用') : t('common.disabled', '禁用') }}
              </span>
            </template>

            <template #cell-actions="{ row }">
              <div class="flex items-center gap-1">
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

      <!-- Tab: Subscription Plans -->
      <div v-show="activeTab === 'plans'">
        <div class="card">
          <div class="flex items-center justify-between border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.payment.subscriptionPlans', '订阅套餐 / Subscription Plans') }}
            </h3>
            <div class="flex gap-2">
              <button @click="loadPlans" :disabled="plansLoading" class="btn btn-secondary btn-sm">
                <Icon name="refresh" size="sm" :class="plansLoading ? 'animate-spin' : ''" />
              </button>
              <button @click="openPlanDialog()" class="btn btn-primary btn-sm">
                <Icon name="plus" size="sm" class="mr-1" />
                {{ t('common.create', '创建') }}
              </button>
            </div>
          </div>

          <DataTable :columns="planColumns" :data="plansList" :loading="plansLoading">
            <template #cell-group_id="{ value }">
              <span class="font-mono text-sm">{{ value ?? '-' }}</span>
            </template>

            <template #cell-price="{ value, row }">
              <div>
                <span class="font-medium text-gray-900 dark:text-white">{{ value }}</span>
                <span
                  v-if="row.original_price && Number(row.original_price) > Number(value)"
                  class="ml-1 text-xs text-gray-400 line-through"
                >
                  {{ row.original_price }}
                </span>
              </div>
            </template>

            <template #cell-validity="{ row }">
              <span class="text-sm text-gray-700 dark:text-gray-300">
                {{ row.validity_days }} {{ validityUnitLabel(row.validity_unit) }}
              </span>
            </template>

            <template #cell-for_sale="{ value }">
              <span :class="['badge', value ? 'badge-success' : 'badge-gray']">
                {{ value ? t('common.yes', '是') : t('common.no', '否') }}
              </span>
            </template>

            <template #cell-actions="{ row }">
              <div class="flex items-center gap-1">
                <button
                  @click="openPlanDialog(row)"
                  class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-600 dark:hover:text-gray-300"
                >
                  <Icon name="edit" size="sm" />
                </button>
                <button
                  @click="handleDeletePlan(row)"
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
      :title="editingProvider ? t('admin.payment.editProvider', '编辑实例') : t('admin.payment.createProvider', '创建实例')"
      width="wide"
      @close="showProviderDialog = false"
    >
      <form id="provider-form" @submit.prevent="saveProvider" class="space-y-4">
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.payment.providerKey', '提供商 / Provider') }}</label>
            <Select
              v-model="providerForm.provider_key"
              :options="providerKeyOptions"
              :disabled="!!editingProvider"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.payment.name', '名称 / Name') }}</label>
            <input v-model="providerForm.name" type="text" required class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.payment.supportedTypes', '支持类型 / Supported Types') }}</label>
            <input
              v-model="providerForm.supported_types"
              type="text"
              class="input"
              placeholder="alipay,wxpay"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.payment.sortOrder', '排序 / Sort Order') }}</label>
            <input v-model.number="providerForm.sort_order" type="number" class="input" />
          </div>
          <div class="flex items-center gap-4">
            <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
              <Toggle v-model="providerForm.enabled" />
              {{ t('common.enabled', '启用') }}
            </label>
            <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
              <Toggle v-model="providerForm.refund_enabled" />
              {{ t('admin.payment.refundEnabled', '允许退款') }}
            </label>
          </div>
        </div>

        <!-- Config Fields -->
        <div>
          <label class="input-label">{{ t('admin.payment.configFields', '配置字段 / Config') }}</label>
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
                class="btn btn-secondary btn-sm"
              >
                <Icon name="plus" size="sm" />
              </button>
            </div>
          </div>
        </div>

        <!-- Limits JSON -->
        <div>
          <label class="input-label">
            {{ t('admin.payment.limits', '限额 / Limits') }}
            <span class="ml-1 text-xs font-normal text-gray-400">(JSON, {{ t('common.optional', '可选') }})</span>
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
            {{ t('common.cancel', '取消') }}
          </button>
          <button type="submit" form="provider-form" :disabled="savingProvider" class="btn btn-primary">
            {{ savingProvider ? t('common.saving', '保存中...') : t('common.save', '保存') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Channel Dialog -->
    <BaseDialog
      :show="showChannelDialog"
      :title="editingChannel ? t('admin.payment.editChannel', '编辑渠道') : t('admin.payment.createChannel', '创建渠道')"
      width="normal"
      @close="showChannelDialog = false"
    >
      <form id="channel-form" @submit.prevent="saveChannel" class="space-y-4">
        <div>
          <label class="input-label">{{ t('admin.payment.name', '名称 / Name') }}</label>
          <input v-model="channelForm.name" type="text" required class="input" />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="input-label">{{ t('admin.payment.groupId', '分组 ID / Group ID') }}</label>
            <input v-model.number="channelForm.group_id" type="number" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.payment.platform', '平台 / Platform') }}</label>
            <input v-model="channelForm.platform" type="text" required class="input" placeholder="web / mobile" />
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="input-label">{{ t('admin.payment.rateMultiplier', '倍率 / Rate Multiplier') }}</label>
            <input v-model="channelForm.rate_multiplier" type="text" class="input" placeholder="1.0" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.payment.sortOrder', '排序 / Sort Order') }}</label>
            <input v-model.number="channelForm.sort_order" type="number" class="input" />
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('admin.payment.description', '描述 / Description') }}</label>
          <textarea v-model="channelForm.description" rows="2" class="input"></textarea>
        </div>
        <div>
          <label class="input-label">{{ t('admin.payment.models', '模型 / Models') }}</label>
          <input v-model="channelForm.models" type="text" class="input" placeholder="gpt-4,claude-3" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.payment.features', '特性 / Features') }}</label>
          <input v-model="channelForm.features" type="text" class="input" placeholder="unlimited,priority" />
        </div>
        <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
          <Toggle v-model="channelForm.enabled" />
          {{ t('common.enabled', '启用') }}
        </label>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" @click="showChannelDialog = false" class="btn btn-secondary">
            {{ t('common.cancel', '取消') }}
          </button>
          <button type="submit" form="channel-form" :disabled="savingChannel" class="btn btn-primary">
            {{ savingChannel ? t('common.saving', '保存中...') : t('common.save', '保存') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Plan Dialog -->
    <BaseDialog
      :show="showPlanDialog"
      :title="editingPlan ? t('admin.payment.editPlan', '编辑套餐') : t('admin.payment.createPlan', '创建套餐')"
      width="normal"
      @close="showPlanDialog = false"
    >
      <form id="plan-form" @submit.prevent="savePlan" class="space-y-4">
        <div>
          <label class="input-label">{{ t('admin.payment.name', '名称 / Name') }}</label>
          <input v-model="planForm.name" type="text" required class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.payment.description', '描述 / Description') }}</label>
          <textarea v-model="planForm.description" rows="2" class="input"></textarea>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="input-label">{{ t('admin.payment.groupId', '分组 ID / Group ID') }}</label>
            <input v-model.number="planForm.group_id" type="number" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.payment.price', '价格 / Price') }}</label>
            <input v-model="planForm.price" type="text" required class="input" />
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="input-label">{{ t('admin.payment.originalPrice', '原价 / Original Price') }}</label>
            <input v-model="planForm.original_price" type="text" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.payment.productName', '产品名 / Product Name') }}</label>
            <input v-model="planForm.product_name" type="text" class="input" />
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="input-label">{{ t('admin.payment.validityDays', '有效期 / Validity') }}</label>
            <input v-model.number="planForm.validity_days" type="number" min="1" required class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.payment.validityUnit', '单位 / Unit') }}</label>
            <Select v-model="planForm.validity_unit" :options="validityUnitOptions" />
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('admin.payment.features', '特性 / Features') }}</label>
          <textarea v-model="planForm.features" rows="2" class="input" placeholder="Feature 1, Feature 2"></textarea>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="input-label">{{ t('admin.payment.sortOrder', '排序 / Sort Order') }}</label>
            <input v-model.number="planForm.sort_order" type="number" class="input" />
          </div>
          <label class="flex items-center gap-2 self-end text-sm text-gray-700 dark:text-gray-300">
            <Toggle v-model="planForm.for_sale" />
            {{ t('admin.payment.forSale', '上架销售') }}
          </label>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" @click="showPlanDialog = false" class="btn btn-secondary">
            {{ t('common.cancel', '取消') }}
          </button>
          <button type="submit" form="plan-form" :disabled="savingPlan" class="btn btn-primary">
            {{ savingPlan ? t('common.saving', '保存中...') : t('common.save', '保存') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Delete Confirm -->
    <ConfirmDialog
      :show="showDeleteConfirm"
      :title="t('common.delete', '删除')"
      :message="deleteConfirmMessage"
      :confirm-text="t('common.delete', '删除')"
      :cancel-text="t('common.cancel', '取消')"
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
import type { PaymentChannel, SubscriptionPlan, ProviderInstance, PaymentDashboardStats } from '@/types/payment'
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

type TabKey = 'dashboard' | 'settings' | 'providers' | 'channels' | 'plans'
const activeTab = ref<TabKey>('dashboard')

const tabs = [
  { key: 'dashboard' as TabKey, label: t('admin.payment.dashboard', '数据概览 / Dashboard') },
  { key: 'settings' as TabKey, label: t('admin.payment.settings', '基础配置 / Settings') },
  { key: 'providers' as TabKey, label: t('admin.payment.providers', '支付实例 / Providers') },
  { key: 'channels' as TabKey, label: t('admin.payment.channelsTab', '渠道 / Channels') },
  { key: 'plans' as TabKey, label: t('admin.payment.plansTab', '套餐 / Plans') }
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
    appStore.showError(error.response?.data?.detail || t('admin.payment.loadDashboardFailed', '加载仪表盘失败'))
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
            label: t('admin.payment.amount', '金额'),
            data: series.map(d => Number(d.amount)),
            backgroundColor: isDark ? 'rgba(139,92,246,0.6)' : 'rgba(139,92,246,0.7)',
            borderRadius: 4,
            yAxisID: 'y'
          },
          {
            label: t('admin.payment.orderCount', '订单数'),
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
            title: { display: true, text: t('admin.payment.amount', '金额'), color: textColor }
          },
          y1: {
            position: 'right',
            ticks: { color: textColor },
            grid: { drawOnChartArea: false },
            title: { display: true, text: t('admin.payment.orderCount', '订单数'), color: textColor }
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
    appStore.showError(error.response?.data?.detail || t('admin.payment.loadConfigFailed', '加载配置失败'))
  } finally {
    configLoading.value = false
  }
}

async function saveConfig() {
  savingConfig.value = true
  try {
    await paymentAdminAPI.updateConfig(configSettings)
    appStore.showSuccess(t('admin.payment.configSaved', '配置已保存'))
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.saveConfigFailed', '保存配置失败'))
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
  { key: 'name', label: t('admin.payment.name', '名称') },
  { key: 'provider_key', label: t('admin.payment.providerKey', '提供商') },
  { key: 'supported_types', label: t('admin.payment.supportedTypes', '类型') },
  { key: 'enabled', label: t('common.enabled', '启用') },
  { key: 'refund_enabled', label: t('admin.payment.refundEnabled', '退款') },
  { key: 'actions', label: t('common.actions', '操作') }
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
    appStore.showError(error.response?.data?.detail || t('admin.payment.loadProvidersFailed', '加载失败'))
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
    appStore.showSuccess(t('admin.payment.providerSaved', '已保存'))
    showProviderDialog.value = false
    loadProviderInstances()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.saveProviderFailed', '保存失败'))
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
  { key: 'name', label: t('admin.payment.name', '名称') },
  { key: 'group_id', label: t('admin.payment.groupId', '分组') },
  { key: 'platform', label: t('admin.payment.platform', '平台') },
  { key: 'rate_multiplier', label: t('admin.payment.rateMultiplier', '倍率') },
  { key: 'enabled', label: t('common.enabled', '启用') },
  { key: 'actions', label: t('common.actions', '操作') }
])

async function loadChannels() {
  channelsLoading.value = true
  try {
    channels.value = await paymentAdminAPI.listChannels()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.loadChannelsFailed', '加载失败'))
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
    appStore.showSuccess(t('admin.payment.channelSaved', '已保存'))
    showChannelDialog.value = false
    loadChannels()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.saveChannelFailed', '保存失败'))
  } finally {
    savingChannel.value = false
  }
}

// ==================== Subscription Plans ====================

const plansLoading = ref(false)
const savingPlan = ref(false)
const plansList = ref<SubscriptionPlan[]>([])
const showPlanDialog = ref(false)
const editingPlan = ref<SubscriptionPlan | null>(null)

const planForm = reactive({
  group_id: null as number | null,
  name: '',
  description: '',
  price: '',
  original_price: '',
  validity_days: 30,
  validity_unit: 'day',
  features: '',
  product_name: '',
  for_sale: false,
  sort_order: 0
})

const validityUnitOptions = [
  { value: 'day', label: '天 / Day' },
  { value: 'week', label: '周 / Week' },
  { value: 'month', label: '月 / Month' }
]

const planColumns = computed<Column[]>(() => [
  { key: 'id', label: 'ID', sortable: true },
  { key: 'name', label: t('admin.payment.name', '名称') },
  { key: 'group_id', label: t('admin.payment.groupId', '分组') },
  { key: 'price', label: t('admin.payment.price', '价格'), sortable: true },
  { key: 'validity', label: t('admin.payment.validity', '有效期') },
  { key: 'for_sale', label: t('admin.payment.forSale', '上架'), sortable: true },
  { key: 'sort_order', label: t('admin.payment.sortOrder', '排序'), sortable: true },
  { key: 'actions', label: t('common.actions', '操作') }
])

function validityUnitLabel(unit: string): string {
  const map: Record<string, string> = { day: '天', week: '周', month: '月' }
  return map[unit] || unit
}

async function loadPlans() {
  plansLoading.value = true
  try {
    plansList.value = await paymentAdminAPI.listPlans()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.loadPlansFailed', '加载失败'))
  } finally {
    plansLoading.value = false
  }
}

function openPlanDialog(plan?: SubscriptionPlan) {
  if (plan) {
    editingPlan.value = plan
    planForm.group_id = plan.group_id
    planForm.name = plan.name
    planForm.description = plan.description || ''
    planForm.price = String(plan.price)
    planForm.original_price = plan.original_price ? String(plan.original_price) : ''
    planForm.validity_days = plan.validity_days
    planForm.validity_unit = plan.validity_unit
    planForm.features = Array.isArray(plan.features) ? plan.features.join(',') : String(plan.features || '')
    planForm.product_name = plan.product_name || ''
    planForm.for_sale = plan.for_sale
    planForm.sort_order = plan.sort_order
  } else {
    editingPlan.value = null
    planForm.group_id = null
    planForm.name = ''
    planForm.description = ''
    planForm.price = ''
    planForm.original_price = ''
    planForm.validity_days = 30
    planForm.validity_unit = 'day'
    planForm.features = ''
    planForm.product_name = ''
    planForm.for_sale = false
    planForm.sort_order = 0
  }
  showPlanDialog.value = true
}

async function savePlan() {
  savingPlan.value = true
  try {
    const payload: any = {
      name: planForm.name,
      price: planForm.price,
      validity_days: planForm.validity_days,
      validity_unit: planForm.validity_unit,
      for_sale: planForm.for_sale,
      sort_order: planForm.sort_order
    }
    if (planForm.group_id !== null) payload.group_id = planForm.group_id
    if (planForm.description) payload.description = planForm.description
    if (planForm.original_price) payload.original_price = planForm.original_price
    if (planForm.features) payload.features = planForm.features
    if (planForm.product_name) payload.product_name = planForm.product_name

    if (editingPlan.value) {
      await paymentAdminAPI.updatePlan(editingPlan.value.id, payload)
    } else {
      await paymentAdminAPI.createPlan(payload)
    }
    appStore.showSuccess(t('admin.payment.planSaved', '已保存'))
    showPlanDialog.value = false
    loadPlans()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.savePlanFailed', '保存失败'))
  } finally {
    savingPlan.value = false
  }
}

// ==================== Delete ====================

const showDeleteConfirm = ref(false)
const deleteConfirmMessage = ref('')
let deleteAction: (() => Promise<void>) | null = null

function handleDeleteProvider(instance: ProviderInstance) {
  deleteConfirmMessage.value = t(
    'admin.payment.deleteProviderConfirm',
    `确定要删除支付实例 "${instance.name}" 吗？/ Delete provider "${instance.name}"?`
  )
  deleteAction = async () => {
    await paymentAdminAPI.deleteProviderInstance(instance.id)
    loadProviderInstances()
  }
  showDeleteConfirm.value = true
}

function handleDeleteChannel(channel: PaymentChannel) {
  deleteConfirmMessage.value = t(
    'admin.payment.deleteChannelConfirm',
    `确定要删除渠道 "${channel.name}" 吗？/ Delete channel "${channel.name}"?`
  )
  deleteAction = async () => {
    await paymentAdminAPI.deleteChannel(channel.id)
    loadChannels()
  }
  showDeleteConfirm.value = true
}

function handleDeletePlan(plan: SubscriptionPlan) {
  deleteConfirmMessage.value = t(
    'admin.payment.deletePlanConfirm',
    `确定要删除套餐 "${plan.name}" 吗？/ Delete plan "${plan.name}"?`
  )
  deleteAction = async () => {
    await paymentAdminAPI.deletePlan(plan.id)
    loadPlans()
  }
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (!deleteAction) return
  try {
    await deleteAction()
    appStore.showSuccess(t('admin.payment.deleted', '已删除'))
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.deleteFailed', '删除失败'))
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
  if (tab === 'plans' && plansList.value.length === 0) loadPlans()
})

onMounted(() => {
  loadDashboard()
})
</script>
