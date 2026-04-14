<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Loading -->
      <div v-if="configLoading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <template v-else>
        <!-- Recharge Form -->
        <div class="card p-6">
          <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('user.payment.recharge', '充值 / Recharge') }}
          </h2>

          <form @submit.prevent="handleCreateOrder" class="space-y-5">
            <!-- Order Type -->
            <div>
              <label class="input-label">{{ t('user.payment.orderType', '订单类型 / Order Type') }}</label>
              <div class="mt-1 flex gap-4">
                <label class="flex cursor-pointer items-center gap-2">
                  <input
                    v-model="orderForm.order_type"
                    type="radio"
                    value="balance"
                    class="text-primary-600 focus:ring-primary-500"
                  />
                  <span class="text-sm text-gray-700 dark:text-gray-300">
                    {{ t('user.payment.balance', '余额充值 / Balance') }}
                  </span>
                </label>
                <label v-if="plans.length > 0" class="flex cursor-pointer items-center gap-2">
                  <input
                    v-model="orderForm.order_type"
                    type="radio"
                    value="subscription"
                    class="text-primary-600 focus:ring-primary-500"
                  />
                  <span class="text-sm text-gray-700 dark:text-gray-300">
                    {{ t('user.payment.subscription', '订阅 / Subscription') }}
                  </span>
                </label>
              </div>
            </div>

            <!-- Subscription Plan Selector (only when order_type === subscription) -->
            <div v-if="orderForm.order_type === 'subscription' && plans.length > 0">
              <label class="input-label">{{ t('user.payment.selectPlan', '选择套餐 / Select Plan') }}</label>
              <div class="mt-2 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
                <div
                  v-for="plan in plans"
                  :key="plan.id"
                  @click="selectPlan(plan)"
                  :class="[
                    'cursor-pointer rounded-xl border-2 p-4 transition-all',
                    orderForm.plan_id === plan.id
                      ? 'border-primary-500 bg-primary-50 dark:border-primary-400 dark:bg-primary-900/20'
                      : 'border-gray-200 hover:border-gray-300 dark:border-dark-600 dark:hover:border-dark-500'
                  ]"
                >
                  <div class="flex items-baseline justify-between">
                    <h4 class="font-semibold text-gray-900 dark:text-white">{{ plan.name }}</h4>
                    <div class="text-right">
                      <span class="text-lg font-bold text-primary-600 dark:text-primary-400">
                        {{ plan.price }}
                      </span>
                      <span
                        v-if="plan.original_price && Number(plan.original_price) > Number(plan.price)"
                        class="ml-1 text-sm text-gray-400 line-through"
                      >
                        {{ plan.original_price }}
                      </span>
                    </div>
                  </div>
                  <p v-if="plan.description" class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    {{ plan.description }}
                  </p>
                  <div class="mt-2 text-xs text-gray-400 dark:text-dark-400">
                    {{ plan.validity_days }}{{ validityUnitLabel(plan.validity_unit) }}
                  </div>
                  <div v-if="plan.features" class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                    {{ plan.features }}
                  </div>
                </div>
              </div>
            </div>

            <!-- Amount -->
            <div v-if="orderForm.order_type === 'balance'">
              <label class="input-label">{{ t('user.payment.amount', '金额 / Amount') }}</label>
              <input
                v-model.number="orderForm.amount"
                type="number"
                step="0.01"
                :min="config?.min_recharge_amount || 1"
                :max="config?.max_recharge_amount || 10000"
                required
                class="input"
                :placeholder="`${config?.min_recharge_amount || 1} - ${config?.max_recharge_amount || 10000}`"
              />
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                {{ t('user.payment.amountRange', '范围 / Range') }}:
                {{ config?.min_recharge_amount || 1 }} - {{ config?.max_recharge_amount || 10000 }}
              </p>
            </div>

            <!-- Payment Type -->
            <div>
              <label class="input-label">{{ t('user.payment.paymentType', '支付方式 / Payment Method') }}</label>
              <div class="mt-2 space-y-2">
                <label
                  v-for="method in availableMethods"
                  :key="method.payment_type"
                  :class="[
                    'flex cursor-pointer items-center justify-between rounded-lg border p-3 transition-all',
                    orderForm.payment_type === method.payment_type
                      ? 'border-primary-500 bg-primary-50 dark:border-primary-400 dark:bg-primary-900/20'
                      : 'border-gray-200 hover:border-gray-300 dark:border-dark-600 dark:hover:border-dark-500',
                    !method.available && 'cursor-not-allowed opacity-50'
                  ]"
                >
                  <div class="flex items-center gap-3">
                    <input
                      v-model="orderForm.payment_type"
                      type="radio"
                      :value="method.payment_type"
                      :disabled="!method.available"
                      class="text-primary-600 focus:ring-primary-500"
                    />
                    <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                      {{ paymentTypeLabel(method.payment_type) }}
                    </span>
                  </div>
                  <div class="text-right text-xs text-gray-500 dark:text-gray-400">
                    <span v-if="Number(method.fee_rate) > 0">
                      {{ t('user.payment.fee', '手续费 / Fee') }}: {{ (Number(method.fee_rate) * 100).toFixed(1) }}%
                    </span>
                    <span v-else class="text-green-500">
                      {{ t('user.payment.noFee', '免手续费 / No fee') }}
                    </span>
                  </div>
                </label>
              </div>
            </div>

            <button
              type="submit"
              :disabled="creating || !canCreateOrder"
              class="btn btn-primary w-full"
            >
              {{ creating ? t('common.loading', '处理中...') : t('user.payment.createOrder', '创建订单 / Create Order') }}
            </button>
          </form>
        </div>

        <!-- Active Order -->
        <div v-if="activeOrder" class="card p-6">
          <div class="mb-4 flex items-center justify-between">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('user.payment.activeOrder', '当前订单 / Active Order') }}
            </h2>
            <span :class="['badge', orderStatusClass(activeOrder.order.status)]">
              {{ orderStatusLabel(activeOrder.order.status) }}
            </span>
          </div>

          <div class="space-y-3">
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400">{{ t('user.payment.orderId', '订单号 / Order ID') }}</span>
              <span class="font-mono text-gray-900 dark:text-white">#{{ activeOrder.order.id }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400">{{ t('user.payment.amount', '金额 / Amount') }}</span>
              <span class="font-medium text-gray-900 dark:text-white">{{ activeOrder.order.amount }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400">{{ t('user.payment.payAmount', '实付 / Pay Amount') }}</span>
              <span class="font-medium text-gray-900 dark:text-white">{{ activeOrder.order.pay_amount }}</span>
            </div>
            <div v-if="Number(activeOrder.order.fee_rate) > 0" class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400">{{ t('user.payment.feeRate', '费率 / Fee Rate') }}</span>
              <span class="text-gray-700 dark:text-gray-300">{{ (Number(activeOrder.order.fee_rate) * 100).toFixed(1) }}%</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400">{{ t('user.payment.expiresAt', '过期时间 / Expires') }}</span>
              <span class="text-gray-700 dark:text-gray-300">{{ formatDateTime(activeOrder.order.expires_at) }}</span>
            </div>

            <!-- Pay URL -->
            <div v-if="activeOrder.pay_url" class="mt-4">
              <a
                :href="activeOrder.pay_url"
                target="_blank"
                rel="noopener noreferrer"
                class="btn btn-primary w-full"
              >
                {{ t('user.payment.goToPay', '前往支付 / Go to Pay') }}
                <Icon name="externalLink" size="sm" class="ml-2" />
              </a>
            </div>

            <!-- QR Code -->
            <div v-if="activeOrder.qr_code" class="mt-4 flex flex-col items-center">
              <p class="mb-2 text-sm text-gray-500 dark:text-gray-400">
                {{ t('user.payment.scanQr', '扫码支付 / Scan to Pay') }}
              </p>
              <div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600">
                <img
                  v-if="activeOrder.qr_code.startsWith('http')"
                  :src="activeOrder.qr_code"
                  alt="QR Code"
                  class="h-48 w-48"
                />
                <div v-else class="break-all font-mono text-xs text-gray-600">
                  {{ activeOrder.qr_code }}
                </div>
              </div>
            </div>

            <!-- Polling indicator -->
            <p v-if="polling" class="mt-3 text-center text-xs text-gray-400 dark:text-dark-400">
              {{ t('user.payment.polling', '正在检查支付状态... / Checking payment status...') }}
            </p>
          </div>
        </div>

        <!-- Order History -->
        <div class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <div class="flex items-center justify-between">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                {{ t('user.payment.orderHistory', '订单记录 / Order History') }}
              </h2>
              <button
                @click="loadOrders"
                :disabled="ordersLoading"
                class="btn btn-secondary btn-sm"
              >
                <Icon name="refresh" size="sm" :class="ordersLoading ? 'animate-spin' : ''" />
              </button>
            </div>
          </div>

          <DataTable :columns="orderColumns" :data="orders" :loading="ordersLoading">
            <template #cell-amount="{ value }">
              <span class="font-medium text-gray-900 dark:text-white">{{ value }}</span>
            </template>

            <template #cell-status="{ value }">
              <span :class="['badge', orderStatusClass(value)]">
                {{ orderStatusLabel(value) }}
              </span>
            </template>

            <template #cell-payment_type="{ value }">
              <span class="text-sm text-gray-700 dark:text-gray-300">
                {{ paymentTypeLabel(value) }}
              </span>
            </template>

            <template #cell-order_type="{ value }">
              <span class="text-sm text-gray-700 dark:text-gray-300">
                {{ value === 'balance' ? t('user.payment.balance', '余额') : t('user.payment.subscription', '订阅') }}
              </span>
            </template>

            <template #cell-created_at="{ value }">
              <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
            </template>

            <template #cell-actions="{ row }">
              <div class="flex items-center gap-1">
                <button
                  v-if="row.status === 'pending'"
                  @click="handleCancelOrder(row.id)"
                  class="btn btn-danger btn-sm text-xs"
                >
                  {{ t('common.cancel', '取消') }}
                </button>
                <button
                  v-if="row.status === 'completed'"
                  @click="openRefundDialog(row)"
                  class="btn btn-secondary btn-sm text-xs"
                >
                  {{ t('user.payment.refund', '退款') }}
                </button>
              </div>
            </template>
          </DataTable>

          <div class="px-6 py-3">
            <Pagination
              v-if="orderPagination.total > 0"
              :page="orderPagination.page"
              :total="orderPagination.total"
              :page-size="orderPagination.page_size"
              @update:page="handleOrderPageChange"
              @update:pageSize="handleOrderPageSizeChange"
            />
          </div>
        </div>
      </template>

      <!-- Refund Dialog -->
      <BaseDialog
        :show="showRefundDialog"
        :title="t('user.payment.requestRefund', '申请退款 / Request Refund')"
        width="normal"
        @close="showRefundDialog = false"
      >
        <form id="refund-form" @submit.prevent="handleRequestRefund" class="space-y-4">
          <div>
            <label class="input-label">{{ t('user.payment.refundAmount', '退款金额 / Refund Amount') }}</label>
            <input
              v-model.number="refundForm.amount"
              type="number"
              step="0.01"
              min="0.01"
              :max="refundingOrder?.amount"
              required
              class="input"
            />
          </div>
          <div>
            <label class="input-label">{{ t('user.payment.refundReason', '退款原因 / Reason') }}</label>
            <textarea
              v-model="refundForm.reason"
              rows="3"
              class="input"
              :placeholder="t('user.payment.refundReasonPlaceholder', '请说明退款原因 / Please describe the reason')"
            ></textarea>
          </div>
        </form>
        <template #footer>
          <div class="flex justify-end gap-3">
            <button type="button" @click="showRefundDialog = false" class="btn btn-secondary">
              {{ t('common.cancel', '取消') }}
            </button>
            <button type="submit" form="refund-form" :disabled="refunding" class="btn btn-primary">
              {{ refunding ? t('common.loading', '处理中...') : t('common.submit', '提交') }}
            </button>
          </div>
        </template>
      </BaseDialog>

      <!-- Cancel Confirm -->
      <ConfirmDialog
        :show="showCancelConfirm"
        :title="t('user.payment.cancelOrder', '取消订单 / Cancel Order')"
        :message="t('user.payment.cancelConfirm', '确定要取消此订单吗？/ Are you sure you want to cancel this order?')"
        :confirm-text="t('common.confirm', '确定')"
        :cancel-text="t('common.cancel', '取消')"
        danger
        @confirm="confirmCancelOrder"
        @cancel="showCancelConfirm = false"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { paymentAPI } from '@/api/payment'
import { formatDateTime } from '@/utils/format'
import type { PaymentOrder, SubscriptionPlan, CreateOrderResponse } from '@/types/payment'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

// ==================== State ====================

const configLoading = ref(false)
const creating = ref(false)
const ordersLoading = ref(false)
const refunding = ref(false)
const polling = ref(false)

const config = ref<{
  enabled_payment_types: string[]
  min_recharge_amount: string
  max_recharge_amount: string
  max_daily_recharge_amount: string
  balance_payment_disabled: boolean
  max_pending_orders: number
  pending_count: number
  method_limits: Array<{
    payment_type: string
    available: boolean
    daily_limit: string
    daily_used: string
    remaining: string
    single_min: string
    single_max: string
    fee_rate: string
  }>
} | null>(null)

const plans = ref<SubscriptionPlan[]>([])
const orders = ref<PaymentOrder[]>([])
const activeOrder = ref<CreateOrderResponse | null>(null)

const orderForm = reactive({
  amount: 10,
  payment_type: '',
  order_type: 'balance' as 'balance' | 'subscription',
  plan_id: null as number | null
})

const orderPagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0
})

// Refund
const showRefundDialog = ref(false)
const refundingOrder = ref<PaymentOrder | null>(null)
const refundForm = reactive({ amount: 0, reason: '' })

// Cancel
const showCancelConfirm = ref(false)
const cancellingOrderId = ref<number | null>(null)

let pollTimer: ReturnType<typeof setInterval> | null = null

// ==================== Computed ====================

const availableMethods = computed(() => {
  return config.value?.method_limits || []
})

const canCreateOrder = computed(() => {
  if (!orderForm.payment_type) return false
  if (orderForm.order_type === 'balance') {
    const amt = Number(orderForm.amount)
    if (!amt || amt <= 0) return false
    const min = Number(config.value?.min_recharge_amount || 0)
    const max = Number(config.value?.max_recharge_amount || Infinity)
    if (amt < min || amt > max) return false
  }
  if (orderForm.order_type === 'subscription' && !orderForm.plan_id) return false
  if (config.value && config.value.pending_count >= config.value.max_pending_orders) return false
  return true
})

const orderColumns = computed<Column[]>(() => [
  { key: 'id', label: 'ID' },
  { key: 'amount', label: t('user.payment.amount', '金额') },
  { key: 'status', label: t('user.payment.status', '状态') },
  { key: 'payment_type', label: t('user.payment.paymentType', '支付方式') },
  { key: 'order_type', label: t('user.payment.orderType', '类型') },
  { key: 'created_at', label: t('user.payment.createdAt', '创建时间'), sortable: true },
  { key: 'actions', label: t('common.actions', '操作') }
])

// ==================== Helpers ====================

function paymentTypeLabel(type: string): string {
  const map: Record<string, string> = {
    alipay: '支付宝 / Alipay',
    wxpay: '微信支付 / WeChat Pay',
    wechat: '微信支付 / WeChat Pay',
    stripe: 'Stripe',
    usdt: 'USDT'
  }
  return map[type] || type
}

function orderStatusLabel(status: string): string {
  const map: Record<string, string> = {
    pending: '待支付 / Pending',
    paid: '已支付 / Paid',
    recharging: '充值中 / Recharging',
    completed: '已完成 / Completed',
    expired: '已过期 / Expired',
    cancelled: '已取消 / Cancelled',
    failed: '失败 / Failed',
    refund_requested: '退款申请中 / Refund Requested',
    refunding: '退款中 / Refunding',
    partially_refunded: '部分退款 / Partial Refund',
    refunded: '已退款 / Refunded',
    refund_failed: '退款失败 / Refund Failed'
  }
  return map[status] || status
}

function orderStatusClass(status: string): string {
  const map: Record<string, string> = {
    pending: 'badge-warning',
    paid: 'badge-info',
    recharging: 'badge-info',
    completed: 'badge-success',
    expired: 'badge-gray',
    cancelled: 'badge-gray',
    failed: 'badge-danger',
    refund_requested: 'badge-warning',
    refunding: 'badge-warning',
    partially_refunded: 'badge-info',
    refunded: 'badge-gray',
    refund_failed: 'badge-danger'
  }
  return map[status] || 'badge-gray'
}

function validityUnitLabel(unit: string): string {
  const map: Record<string, string> = {
    day: '天 / days',
    week: '周 / weeks',
    month: '月 / months'
  }
  return map[unit] || unit
}

function selectPlan(plan: SubscriptionPlan) {
  orderForm.plan_id = plan.id
  orderForm.amount = Number(plan.price)
}

// ==================== API Calls ====================

async function loadConfig() {
  configLoading.value = true
  try {
    const [configData, plansData] = await Promise.all([
      paymentAPI.getConfig(),
      paymentAPI.listPlans()
    ])
    config.value = configData as any
    plans.value = plansData
    // Auto-select first available payment type
    const firstAvailable = config.value?.method_limits?.find(m => m.available)
    if (firstAvailable) {
      orderForm.payment_type = firstAvailable.payment_type
    }
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('user.payment.loadFailed', '加载配置失败'))
    console.error('Failed to load payment config:', error)
  } finally {
    configLoading.value = false
  }
}

async function loadOrders() {
  ordersLoading.value = true
  try {
    const response = await paymentAPI.listOrders(orderPagination.page, orderPagination.page_size)
    orders.value = response.items
    orderPagination.total = response.total
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('user.payment.loadOrdersFailed', '加载订单失败'))
    console.error('Failed to load orders:', error)
  } finally {
    ordersLoading.value = false
  }
}

async function handleCreateOrder() {
  creating.value = true
  try {
    const payload: any = {
      amount: String(orderForm.amount),
      payment_type: orderForm.payment_type,
      order_type: orderForm.order_type,
      src_host: window.location.hostname,
      return_url: window.location.href
    }
    if (orderForm.order_type === 'subscription' && orderForm.plan_id) {
      payload.plan_id = orderForm.plan_id
    }
    const result = await paymentAPI.createOrder(payload)
    activeOrder.value = result
    appStore.showSuccess(t('user.payment.orderCreated', '订单已创建'))
    startPolling(result.order.id)
    loadOrders()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('user.payment.createFailed', '创建订单失败'))
    console.error('Failed to create order:', error)
  } finally {
    creating.value = false
  }
}

function startPolling(orderId: number) {
  stopPolling()
  polling.value = true
  pollTimer = setInterval(async () => {
    try {
      const order = await paymentAPI.getOrder(orderId)
      if (order.status === 'completed' || order.status === 'paid') {
        stopPolling()
        activeOrder.value = null
        appStore.showSuccess(t('user.payment.paymentSuccess', '支付成功'))
        loadOrders()
      } else if (
        order.status === 'expired' ||
        order.status === 'cancelled' ||
        order.status === 'failed'
      ) {
        stopPolling()
        activeOrder.value = null
        loadOrders()
      }
    } catch {
      // Silently retry
    }
  }, 5000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  polling.value = false
}

// Cancel
function handleCancelOrder(orderId: number) {
  cancellingOrderId.value = orderId
  showCancelConfirm.value = true
}

async function confirmCancelOrder() {
  if (!cancellingOrderId.value) return
  try {
    await paymentAPI.cancelOrder(cancellingOrderId.value)
    appStore.showSuccess(t('user.payment.orderCancelled', '订单已取消'))
    showCancelConfirm.value = false
    cancellingOrderId.value = null
    if (activeOrder.value?.order.id === cancellingOrderId.value) {
      stopPolling()
      activeOrder.value = null
    }
    loadOrders()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('user.payment.cancelFailed', '取消失败'))
  }
}

// Refund
function openRefundDialog(order: PaymentOrder) {
  refundingOrder.value = order
  refundForm.amount = Number(order.amount)
  refundForm.reason = ''
  showRefundDialog.value = true
}

async function handleRequestRefund() {
  if (!refundingOrder.value) return
  refunding.value = true
  try {
    await paymentAPI.requestRefund(refundingOrder.value.id, refundForm.amount, refundForm.reason)
    appStore.showSuccess(t('user.payment.refundRequested', '退款申请已提交'))
    showRefundDialog.value = false
    refundingOrder.value = null
    loadOrders()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('user.payment.refundFailed', '退款申请失败'))
  } finally {
    refunding.value = false
  }
}

// Pagination
function handleOrderPageChange(page: number) {
  orderPagination.page = page
  loadOrders()
}

function handleOrderPageSizeChange(pageSize: number) {
  orderPagination.page_size = pageSize
  orderPagination.page = 1
  loadOrders()
}

// ==================== Lifecycle ====================

onMounted(() => {
  loadConfig()
  loadOrders()
})

onUnmounted(() => {
  stopPolling()
})
</script>
