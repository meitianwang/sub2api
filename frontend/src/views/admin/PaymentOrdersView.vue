<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <!-- Filters -->
          <div class="flex-1 sm:max-w-48">
            <input
              v-model="filters.user_id"
              type="text"
              :placeholder="t('admin.payment.userIdSearch', 'User ID')"
              class="input"
              @input="debouncedLoad"
            />
          </div>
          <Select
            v-model="filters.status"
            :options="statusOptions"
            class="w-40"
            @change="resetAndLoad"
          />
          <Select
            v-model="filters.payment_type"
            :options="paymentTypeOptions"
            class="w-40"
            @change="resetAndLoad"
          />
          <Select
            v-model="filters.order_type"
            :options="orderTypeOptions"
            class="w-36"
            @change="resetAndLoad"
          />

          <!-- Date Range -->
          <input
            v-model="filters.date_from"
            type="date"
            class="input w-36"
            @change="resetAndLoad"
          />
          <span class="text-gray-400">-</span>
          <input
            v-model="filters.date_to"
            type="date"
            class="input w-36"
            @change="resetAndLoad"
          />

          <!-- Actions -->
          <div class="flex flex-1 items-center justify-end gap-2">
            <button
              @click="loadOrders"
              :disabled="loading"
              class="btn btn-secondary"
              :title="t('common.refresh', '刷新')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="orders" :loading="loading">
          <template #cell-id="{ value }">
            <button
              @click="openDetail(value)"
              class="font-mono text-sm text-primary-600 hover:underline dark:text-primary-400"
            >
              #{{ value }}
            </button>
          </template>

          <template #cell-user_email="{ value, row }">
            <div class="text-sm">
              <div class="text-gray-900 dark:text-white">{{ value || '-' }}</div>
              <div class="text-xs text-gray-400">ID: {{ row.user_id }}</div>
            </div>
          </template>

          <template #cell-amount="{ value }">
            <span class="font-medium text-gray-900 dark:text-white">{{ value }}</span>
          </template>

          <template #cell-pay_amount="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ value || '-' }}</span>
          </template>

          <template #cell-status="{ value }">
            <span :class="['badge', orderStatusClass(value)]">
              {{ orderStatusLabel(value) }}
            </span>
          </template>

          <template #cell-payment_type="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ paymentTypeLabel(value) }}</span>
          </template>

          <template #cell-order_type="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">
              {{ value === 'balance' ? '余额 / Balance' : '订阅 / Subscription' }}
            </span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <button
                @click="openDetail(row.id)"
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-600 dark:hover:text-gray-300"
                :title="t('common.details', '详情')"
              >
                <Icon name="eye" size="sm" />
              </button>
              <button
                v-if="row.status === 'pending'"
                @click="handleCancel(row.id)"
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                :title="t('common.cancel', '取消')"
              >
                <Icon name="x" size="sm" />
              </button>
              <button
                v-if="row.status === 'failed'"
                @click="handleRetry(row.id)"
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-blue-50 hover:text-blue-600 dark:hover:bg-blue-900/20 dark:hover:text-blue-400"
                :title="t('admin.payment.retry', '重试')"
              >
                <Icon name="refresh" size="sm" />
              </button>
              <button
                v-if="row.status === 'completed' || row.status === 'paid'"
                @click="openRefundDialog(row)"
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-yellow-50 hover:text-yellow-600 dark:hover:bg-yellow-900/20 dark:hover:text-yellow-400"
                :title="t('admin.payment.refund', '退款')"
              >
                <Icon name="arrowLeft" size="sm" />
              </button>
            </div>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <!-- Order Detail Dialog -->
    <BaseDialog
      :show="showDetailDialog"
      :title="t('admin.payment.orderDetail', '订单详情 / Order Detail')"
      width="wide"
      @close="showDetailDialog = false"
    >
      <div v-if="detailLoading" class="flex items-center justify-center py-8">
        <Icon name="refresh" size="lg" class="animate-spin text-gray-400" />
      </div>
      <div v-else-if="orderDetail" class="space-y-6">
        <!-- Order Info -->
        <div class="grid grid-cols-2 gap-3 text-sm">
          <div>
            <span class="text-gray-500 dark:text-gray-400">ID</span>
            <p class="font-mono font-medium text-gray-900 dark:text-white">#{{ orderDetail.order.id }}</p>
          </div>
          <div>
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.user', '用户') }}</span>
            <p class="text-gray-900 dark:text-white">{{ orderDetail.order.user_email || orderDetail.order.user_id }}</p>
          </div>
          <div>
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.amount', '金额') }}</span>
            <p class="font-medium text-gray-900 dark:text-white">{{ orderDetail.order.amount }}</p>
          </div>
          <div>
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.payAmount', '实付') }}</span>
            <p class="text-gray-900 dark:text-white">{{ orderDetail.order.pay_amount || '-' }}</p>
          </div>
          <div>
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.status', '状态') }}</span>
            <p><span :class="['badge', orderStatusClass(orderDetail.order.status)]">{{ orderStatusLabel(orderDetail.order.status) }}</span></p>
          </div>
          <div>
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.paymentType', '支付方式') }}</span>
            <p class="text-gray-900 dark:text-white">{{ paymentTypeLabel(orderDetail.order.payment_type) }}</p>
          </div>
          <div>
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.orderType', '类型') }}</span>
            <p class="text-gray-900 dark:text-white">{{ orderDetail.order.order_type === 'balance' ? '余额' : '订阅' }}</p>
          </div>
          <div>
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.rechargeCode', '充值码') }}</span>
            <p class="font-mono text-xs text-gray-700 dark:text-gray-300">{{ orderDetail.order.recharge_code || '-' }}</p>
          </div>
          <div v-if="orderDetail.order.payment_trade_no">
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.tradeNo', '交易号') }}</span>
            <p class="font-mono text-xs text-gray-700 dark:text-gray-300">{{ orderDetail.order.payment_trade_no }}</p>
          </div>
          <div v-if="orderDetail.order.fee_rate">
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.feeRate', '费率') }}</span>
            <p class="text-gray-900 dark:text-white">{{ (Number(orderDetail.order.fee_rate) * 100).toFixed(1) }}%</p>
          </div>
          <div v-if="orderDetail.order.refund_amount">
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.refundAmount', '退款金额') }}</span>
            <p class="text-red-600 dark:text-red-400">{{ orderDetail.order.refund_amount }}</p>
          </div>
          <div v-if="orderDetail.order.refund_reason">
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.refundReason', '退款原因') }}</span>
            <p class="text-gray-700 dark:text-gray-300">{{ orderDetail.order.refund_reason }}</p>
          </div>
          <div v-if="orderDetail.order.failed_reason">
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.failedReason', '失败原因') }}</span>
            <p class="text-red-600 dark:text-red-400">{{ orderDetail.order.failed_reason }}</p>
          </div>
          <div>
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.createdAt', '创建时间') }}</span>
            <p class="text-gray-900 dark:text-white">{{ formatDateTime(orderDetail.order.created_at) }}</p>
          </div>
          <div v-if="orderDetail.order.paid_at">
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.paidAt', '支付时间') }}</span>
            <p class="text-gray-900 dark:text-white">{{ formatDateTime(orderDetail.order.paid_at) }}</p>
          </div>
          <div v-if="orderDetail.order.completed_at">
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.completedAt', '完成时间') }}</span>
            <p class="text-gray-900 dark:text-white">{{ formatDateTime(orderDetail.order.completed_at) }}</p>
          </div>
          <div v-if="orderDetail.order.src_host">
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.payment.srcHost', '来源') }}</span>
            <p class="text-gray-700 dark:text-gray-300">{{ orderDetail.order.src_host }}</p>
          </div>
        </div>

        <!-- Audit Logs -->
        <div v-if="orderDetail.audit_logs.length > 0">
          <h4 class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('admin.payment.auditLogs', '审计日志 / Audit Logs') }}
          </h4>
          <div class="space-y-2">
            <div
              v-for="log in orderDetail.audit_logs"
              :key="log.id"
              class="flex items-start gap-3 rounded-lg border border-gray-100 p-3 dark:border-dark-700"
            >
              <div class="mt-0.5 h-2 w-2 flex-shrink-0 rounded-full bg-gray-400"></div>
              <div class="flex-1">
                <div class="flex items-center justify-between">
                  <span class="text-sm font-medium text-gray-900 dark:text-white">{{ log.action }}</span>
                  <span class="text-xs text-gray-400">{{ formatDateTime(log.created_at) }}</span>
                </div>
                <p v-if="log.detail" class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ log.detail }}</p>
                <p v-if="log.operator" class="mt-0.5 text-xs text-gray-400">{{ log.operator }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end">
          <button type="button" @click="showDetailDialog = false" class="btn btn-secondary">
            {{ t('common.close', '关闭') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Refund Dialog -->
    <BaseDialog
      :show="showRefundDialog"
      :title="t('admin.payment.processRefund', '处理退款 / Process Refund')"
      width="normal"
      @close="showRefundDialog = false"
    >
      <form id="admin-refund-form" @submit.prevent="handleRefund" class="space-y-4">
        <div>
          <label class="input-label">{{ t('admin.payment.orderId', '订单 ID') }}</label>
          <input :value="refundingOrder?.id" disabled class="input bg-gray-50 dark:bg-dark-800" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.payment.refundAmount', '退款金额 / Refund Amount') }}</label>
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
          <label class="input-label">{{ t('admin.payment.refundReason', '退款原因 / Reason') }}</label>
          <textarea
            v-model="refundForm.reason"
            rows="3"
            class="input"
            :placeholder="t('admin.payment.refundReasonPlaceholder', '请说明退款原因')"
          ></textarea>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" @click="showRefundDialog = false" class="btn btn-secondary">
            {{ t('common.cancel', '取消') }}
          </button>
          <button type="submit" form="admin-refund-form" :disabled="processingRefund" class="btn btn-danger">
            {{ processingRefund ? t('common.loading', '处理中...') : t('admin.payment.confirmRefund', '确认退款') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Cancel Confirm -->
    <ConfirmDialog
      :show="showCancelConfirm"
      :title="t('admin.payment.cancelOrder', '取消订单 / Cancel Order')"
      :message="t('admin.payment.cancelConfirm', '确定要取消此订单吗？')"
      :confirm-text="t('common.confirm', '确定')"
      :cancel-text="t('common.cancel', '取消')"
      danger
      @confirm="confirmCancel"
      @cancel="showCancelConfirm = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { paymentAdminAPI } from '@/api/admin/payment'
import { formatDateTime } from '@/utils/format'
import type { PaymentOrder, PaymentAuditLog } from '@/types/payment'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

// ==================== State ====================

const loading = ref(false)
const orders = ref<PaymentOrder[]>([])
const detailLoading = ref(false)
const processingRefund = ref(false)

const filters = reactive({
  status: '',
  payment_type: '',
  order_type: '',
  user_id: '',
  date_from: '',
  date_to: ''
})

const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0
})

// Detail
const showDetailDialog = ref(false)
const orderDetail = ref<{ order: PaymentOrder; audit_logs: PaymentAuditLog[] } | null>(null)

// Refund
const showRefundDialog = ref(false)
const refundingOrder = ref<PaymentOrder | null>(null)
const refundForm = reactive({ amount: 0, reason: '' })

// Cancel
const showCancelConfirm = ref(false)
const cancellingOrderId = ref<number | null>(null)

let abortController: AbortController | null = null

// ==================== Options ====================

const statusOptions = computed(() => [
  { value: '', label: t('admin.payment.allStatus', '全部状态 / All Status') },
  { value: 'pending', label: '待支付 / Pending' },
  { value: 'paid', label: '已支付 / Paid' },
  { value: 'recharging', label: '充值中 / Recharging' },
  { value: 'completed', label: '已完成 / Completed' },
  { value: 'expired', label: '已过期 / Expired' },
  { value: 'cancelled', label: '已取消 / Cancelled' },
  { value: 'failed', label: '失败 / Failed' },
  { value: 'refund_requested', label: '退款申请 / Refund Req.' },
  { value: 'refunding', label: '退款中 / Refunding' },
  { value: 'partially_refunded', label: '部分退款 / Partial' },
  { value: 'refunded', label: '已退款 / Refunded' },
  { value: 'refund_failed', label: '退款失败 / Refund Failed' }
])

const paymentTypeOptions = computed(() => [
  { value: '', label: t('admin.payment.allPaymentTypes', '全部方式 / All Types') },
  { value: 'alipay', label: '支付宝 / Alipay' },
  { value: 'wxpay', label: '微信 / WeChat' },
  { value: 'stripe', label: 'Stripe' },
  { value: 'usdt', label: 'USDT' }
])

const orderTypeOptions = computed(() => [
  { value: '', label: t('admin.payment.allOrderTypes', '全部类型 / All') },
  { value: 'balance', label: '余额 / Balance' },
  { value: 'subscription', label: '订阅 / Subscription' }
])

const columns = computed<Column[]>(() => [
  { key: 'id', label: 'ID', sortable: true },
  { key: 'user_email', label: t('admin.payment.user', '用户') },
  { key: 'amount', label: t('admin.payment.amount', '金额'), sortable: true },
  { key: 'pay_amount', label: t('admin.payment.payAmount', '实付') },
  { key: 'status', label: t('admin.payment.status', '状态'), sortable: true },
  { key: 'payment_type', label: t('admin.payment.paymentType', '方式') },
  { key: 'order_type', label: t('admin.payment.orderType', '类型') },
  { key: 'created_at', label: t('admin.payment.createdAt', '创建时间'), sortable: true },
  { key: 'actions', label: t('common.actions', '操作') }
])

// ==================== Helpers ====================

function paymentTypeLabel(type: string): string {
  const map: Record<string, string> = {
    alipay: '支付宝',
    wxpay: '微信',
    wechat: '微信',
    stripe: 'Stripe',
    usdt: 'USDT'
  }
  return map[type] || type
}

function orderStatusLabel(status: string): string {
  const map: Record<string, string> = {
    pending: '待支付',
    paid: '已支付',
    recharging: '充值中',
    completed: '已完成',
    expired: '已过期',
    cancelled: '已取消',
    failed: '失败',
    refund_requested: '退款申请',
    refunding: '退款中',
    partially_refunded: '部分退款',
    refunded: '已退款',
    refund_failed: '退款失败'
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

// ==================== API Calls ====================

async function loadOrders() {
  if (abortController) abortController.abort()
  const currentController = new AbortController()
  abortController = currentController
  loading.value = true

  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.page_size
    }
    if (filters.status) params.status = filters.status
    if (filters.payment_type) params.payment_type = filters.payment_type
    if (filters.order_type) params.order_type = filters.order_type
    if (filters.user_id) params.user_id = Number(filters.user_id)
    if (filters.date_from) params.date_from = new Date(filters.date_from).toISOString()
    if (filters.date_to) {
      const d = new Date(filters.date_to)
      d.setHours(23, 59, 59, 999)
      params.date_to = d.toISOString()
    }

    const response = await paymentAdminAPI.listOrders(params, { signal: currentController.signal })
    if (currentController.signal.aborted) return

    orders.value = response.items
    pagination.total = response.total
  } catch (error: any) {
    if (currentController.signal.aborted || error?.name === 'AbortError') return
    appStore.showError(error.response?.data?.detail || t('admin.payment.loadFailed', '加载订单失败'))
    console.error('Failed to load orders:', error)
  } finally {
    if (abortController === currentController && !currentController.signal.aborted) {
      loading.value = false
      abortController = null
    }
  }
}

let searchTimeout: ReturnType<typeof setTimeout>
function debouncedLoad() {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.page = 1
    loadOrders()
  }, 300)
}

function resetAndLoad() {
  pagination.page = 1
  loadOrders()
}

async function openDetail(orderId: number) {
  showDetailDialog.value = true
  detailLoading.value = true
  orderDetail.value = null

  try {
    const data = await paymentAdminAPI.getOrderDetail(orderId)
    orderDetail.value = data
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.loadDetailFailed', '加载详情失败'))
  } finally {
    detailLoading.value = false
  }
}

// Cancel
function handleCancel(orderId: number) {
  cancellingOrderId.value = orderId
  showCancelConfirm.value = true
}

async function confirmCancel() {
  if (!cancellingOrderId.value) return
  try {
    await paymentAdminAPI.cancelOrder(cancellingOrderId.value)
    appStore.showSuccess(t('admin.payment.orderCancelled', '订单已取消'))
    showCancelConfirm.value = false
    cancellingOrderId.value = null
    loadOrders()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.cancelFailed', '取消失败'))
  }
}

// Retry
async function handleRetry(orderId: number) {
  try {
    await paymentAdminAPI.retryRecharge(orderId)
    appStore.showSuccess(t('admin.payment.retrySuccess', '重试已发起'))
    loadOrders()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.retryFailed', '重试失败'))
  }
}

// Refund
function openRefundDialog(order: PaymentOrder) {
  refundingOrder.value = order
  refundForm.amount = Number(order.amount)
  refundForm.reason = ''
  showRefundDialog.value = true
}

async function handleRefund() {
  if (!refundingOrder.value) return
  processingRefund.value = true
  try {
    await paymentAdminAPI.processRefund(refundingOrder.value.id, refundForm.amount, refundForm.reason)
    appStore.showSuccess(t('admin.payment.refundSuccess', '退款已处理'))
    showRefundDialog.value = false
    refundingOrder.value = null
    loadOrders()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.refundFailed', '退款失败'))
  } finally {
    processingRefund.value = false
  }
}

// Pagination
function handlePageChange(page: number) {
  pagination.page = page
  loadOrders()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  loadOrders()
}

// ==================== Lifecycle ====================

onMounted(() => {
  loadOrders()
})

onUnmounted(() => {
  clearTimeout(searchTimeout)
  abortController?.abort()
})
</script>
