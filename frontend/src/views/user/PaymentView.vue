<template>
  <AppLayout>
    <div class="mx-auto max-w-2xl space-y-6">
      <!-- Loading -->
      <div v-if="configLoading" class="flex items-center justify-center py-16">
        <LoadingSpinner size="lg" />
      </div>

      <template v-else>
        <!-- ==================== Step 1: Form ==================== -->
        <template v-if="step === 'form'">
          <!-- Account Info Card -->
          <div class="card overflow-hidden">
            <div class="bg-gradient-to-br from-primary-500 to-primary-600 px-6 py-8 text-center">
              <div
                class="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-white/20 backdrop-blur-sm"
              >
                <Icon name="creditCard" size="xl" class="text-white" />
              </div>
              <p class="text-sm font-medium text-primary-100">{{ t('user.payment.balance') }}</p>
              <p class="mt-2 text-4xl font-bold text-white">
                ¥{{ user?.balance?.toFixed(2) || '0.00' }}
              </p>
              <p v-if="user?.username" class="mt-1 text-sm text-primary-200">
                {{ user.username }}
              </p>
            </div>
          </div>

          <!-- Recharge Form -->
          <div class="card">
            <div class="p-6">
              <h2 class="mb-5 text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('user.payment.recharge') }}
              </h2>

              <!-- Pending Order Warning -->
              <div
                v-if="hasTooManyPendingOrders"
                class="mb-5 rounded-lg border border-amber-300 bg-amber-50 p-4 dark:border-amber-700 dark:bg-amber-900/20"
              >
                <div class="flex items-start gap-3">
                  <Icon
                    name="exclamationTriangle"
                    size="md"
                    class="mt-0.5 flex-shrink-0 text-amber-500"
                  />
                  <p class="text-sm text-amber-700 dark:text-amber-300">
                    {{ t('user.payment.tooManyPendingOrders') }}
                  </p>
                </div>
              </div>

              <form @submit.prevent="handleCreateOrder" class="space-y-5">
                <!-- Quick Amount Buttons -->
                <div>
                  <label class="input-label">{{ t('user.payment.amount') }}</label>
                  <div class="mt-2 grid grid-cols-4 gap-2">
                    <button
                      v-for="amt in quickAmounts"
                      :key="amt"
                      type="button"
                      @click="selectQuickAmount(amt)"
                      :class="[
                        'rounded-lg border-2 px-3 py-2.5 text-center text-sm font-semibold transition-all',
                        selectedQuickAmount === amt
                          ? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-900/30 dark:text-primary-300'
                          : 'border-gray-200 text-gray-700 hover:border-gray-300 hover:bg-gray-50 dark:border-dark-600 dark:text-gray-300 dark:hover:border-dark-500 dark:hover:bg-dark-700'
                      ]"
                    >
                      ¥{{ amt }}
                    </button>
                  </div>
                </div>

                <!-- Custom Amount Input -->
                <div>
                  <label class="input-label">{{ t('user.payment.customAmount') }}</label>
                  <div class="relative mt-1">
                    <div
                      class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4"
                    >
                      <span class="text-sm font-medium text-gray-400 dark:text-dark-500">¥</span>
                    </div>
                    <input
                      v-model="customAmountText"
                      type="text"
                      inputmode="decimal"
                      class="input pl-8"
                      :placeholder="amountRangeHint"
                      @input="handleCustomAmountInput"
                      @focus="selectedQuickAmount = null"
                    />
                  </div>
                  <p class="input-hint">
                    {{ t('user.payment.amountRange') }}: ¥{{ minAmount }} - ¥{{ maxAmount }}
                  </p>
                  <p v-if="amountError" class="mt-1 text-xs text-red-500">{{ amountError }}</p>
                </div>

                <!-- Payment Method Selector -->
                <div>
                  <label class="input-label">{{ t('user.payment.paymentType') }}</label>
                  <div class="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-3">
                    <button
                      v-for="method in availableMethods"
                      :key="method.paymentType"
                      type="button"
                      @click="selectPaymentMethod(method)"
                      :disabled="!method.available"
                      :class="[
                        'flex items-center gap-3 rounded-lg border-2 px-4 py-3 transition-all',
                        !method.available && 'cursor-not-allowed opacity-40',
                        selectedPaymentType === method.paymentType
                          ? paymentMethodSelectedClass(method.paymentType)
                          : 'border-gray-200 hover:border-gray-300 dark:border-dark-600 dark:hover:border-dark-500'
                      ]"
                    >
                      <!-- Payment Method Icon -->
                      <div
                        :class="[
                          'flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg',
                          paymentMethodIconBg(method.paymentType)
                        ]"
                      >
                        <span class="text-sm font-bold text-white">
                          {{ paymentMethodIconText(method.paymentType) }}
                        </span>
                      </div>
                      <div class="text-left">
                        <p class="text-sm font-medium text-gray-900 dark:text-white">
                          {{ getPaymentMethodLabel(method.paymentType) }}
                        </p>
                        <p
                          v-if="method.feeRate > 0"
                          class="text-xs text-gray-500 dark:text-gray-400"
                        >
                          {{ (method.feeRate * 100).toFixed(1) }}%
                          {{ t('user.payment.fee') }}
                        </p>
                        <p v-else class="text-xs text-green-500">
                          {{ t('user.payment.noFee') }}
                        </p>
                      </div>
                    </button>
                  </div>
                </div>

                <!-- Fee Breakdown -->
                <div
                  v-if="feeAmount > 0"
                  class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800"
                >
                  <div class="flex items-center justify-between text-sm">
                    <span class="text-gray-500 dark:text-gray-400">
                      {{ t('user.payment.amount') }}
                    </span>
                    <span class="font-medium text-gray-900 dark:text-white">
                      ¥{{ effectiveAmount.toFixed(2) }}
                    </span>
                  </div>
                  <div class="mt-1 flex items-center justify-between text-sm">
                    <span class="text-gray-500 dark:text-gray-400">
                      {{ t('user.payment.fee') }}
                      ({{ (currentFeeRate * 100).toFixed(1) }}%)
                    </span>
                    <span class="font-medium text-gray-900 dark:text-white">
                      + ¥{{ feeAmount.toFixed(2) }}
                    </span>
                  </div>
                  <div
                    class="mt-2 border-t border-gray-200 pt-2 dark:border-dark-600"
                  >
                    <div class="flex items-center justify-between text-sm">
                      <span class="font-medium text-gray-700 dark:text-gray-300">
                        {{ t('user.payment.payAmount') }}
                      </span>
                      <span class="text-base font-bold text-primary-600 dark:text-primary-400">
                        ¥{{ totalPayAmount.toFixed(2) }}
                      </span>
                    </div>
                  </div>
                </div>

                <!-- Submit Button -->
                <button
                  type="submit"
                  :disabled="creating || !canCreateOrder"
                  :class="[
                    'btn w-full py-3 text-white transition-all',
                    submitButtonColorClass
                  ]"
                >
                  <LoadingSpinner v-if="creating" size="sm" color="white" class="mr-2" />
                  {{ creating ? t('common.loading') : t('user.payment.createOrder') }}
                </button>
              </form>
            </div>
          </div>

          <!-- Order History -->
          <div class="card">
            <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
              <div class="flex items-center justify-between">
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                  {{ t('user.payment.orderHistory') }}
                </h2>
                <button
                  @click="loadOrders"
                  :disabled="ordersLoading"
                  class="btn btn-secondary"
                >
                  <Icon
                    name="refresh"
                    size="sm"
                    :class="ordersLoading ? 'animate-spin' : ''"
                  />
                </button>
              </div>
            </div>

            <div class="p-6">
              <!-- Loading -->
              <div v-if="ordersLoading && orders.length === 0" class="flex justify-center py-8">
                <LoadingSpinner />
              </div>

              <!-- Orders List -->
              <div v-else-if="orders.length > 0" class="space-y-3">
                <div
                  v-for="order in orders"
                  :key="order.id"
                  class="flex items-center justify-between rounded-xl bg-gray-50 p-4 dark:bg-dark-800"
                >
                  <div class="flex items-center gap-3">
                    <div
                      :class="[
                        'flex h-10 w-10 items-center justify-center rounded-xl',
                        paymentMethodIconBg(order.payment_type)
                      ]"
                    >
                      <span class="text-xs font-bold text-white">
                        {{ paymentMethodIconText(order.payment_type) }}
                      </span>
                    </div>
                    <div>
                      <div class="flex items-center gap-2">
                        <span class="text-sm font-semibold text-gray-900 dark:text-white">
                          ¥{{ Number(order.amount).toFixed(2) }}
                        </span>
                        <span
                          :class="['badge', getPaymentStatusBadgeClass(order.status)]"
                        >
                          {{ orderStatusLabel(order.status) }}
                        </span>
                      </div>
                      <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">
                        {{ formatDateTime(order.created_at) }}
                      </p>
                    </div>
                  </div>
                  <div class="flex items-center gap-2">
                    <button
                      v-if="order.status === 'pending'"
                      @click="handleCancelOrderFromHistory(order.id)"
                      class="text-xs text-red-500 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
                    >
                      {{ t('common.cancel') }}
                    </button>
                  </div>
                </div>
              </div>

              <!-- Empty State -->
              <div v-else class="empty-state py-8">
                <div
                  class="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gray-100 dark:bg-dark-800"
                >
                  <Icon name="clock" size="xl" class="text-gray-400 dark:text-dark-500" />
                </div>
                <p class="text-sm text-gray-500 dark:text-dark-400">
                  {{ t('user.payment.noOrders') }}
                </p>
              </div>
            </div>
          </div>
        </template>

        <!-- ==================== Step 2: Paying ==================== -->
        <template v-else-if="step === 'paying' && activeOrder">
          <!-- Amount Display -->
          <div class="card">
            <div class="p-6 text-center">
              <p class="text-sm text-gray-500 dark:text-gray-400">
                {{ t('user.payment.payAmount') }}
              </p>
              <p class="mt-2 text-4xl font-bold text-gray-900 dark:text-white">
                ¥{{ Number(activeOrder.order.pay_amount).toFixed(2) }}
              </p>
              <p
                v-if="Number(activeOrder.order.fee_rate) > 0"
                class="mt-1 text-sm text-gray-500 dark:text-gray-400"
              >
                {{ t('user.payment.creditedAmount') }}:
                ¥{{ Number(activeOrder.order.amount).toFixed(2) }}
              </p>

              <!-- QR Code Display -->
              <div v-if="qrDataUrl" class="mt-6 flex flex-col items-center">
                <p class="mb-3 text-sm text-gray-500 dark:text-gray-400">
                  {{ t('user.payment.scanQr') }}
                </p>
                <div
                  class="relative inline-block rounded-2xl border-2 border-gray-200 bg-white p-4 dark:border-dark-500"
                >
                  <img :src="qrDataUrl" alt="QR Code" class="h-56 w-56" />
                  <!-- Payment method icon overlay -->
                  <div
                    :class="[
                      'absolute left-1/2 top-1/2 flex h-10 w-10 -translate-x-1/2 -translate-y-1/2 items-center justify-center rounded-lg shadow-md',
                      paymentMethodIconBg(activeOrder.order.payment_type)
                    ]"
                  >
                    <span class="text-xs font-bold text-white">
                      {{ paymentMethodIconText(activeOrder.order.payment_type) }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- H5 Pay URL redirect (no QR code) -->
              <div
                v-else-if="activeOrder.pay_url && !activeOrder.qr_code"
                class="mt-6"
              >
                <a
                  :href="activeOrder.pay_url"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="btn btn-primary inline-flex items-center gap-2"
                >
                  {{ t('user.payment.goToPay') }}
                  <Icon name="externalLink" size="sm" />
                </a>
              </div>

              <!-- Countdown Timer -->
              <div v-if="countdownText" class="mt-5">
                <div
                  :class="[
                    'inline-flex items-center gap-2 rounded-full px-4 py-2 text-sm font-medium',
                    countdownUrgent
                      ? 'animate-pulse bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                      : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'
                  ]"
                >
                  <Icon name="clock" size="sm" />
                  {{ countdownText }}
                </div>
              </div>

              <!-- Polling Indicator -->
              <div v-if="polling" class="mt-4 flex items-center justify-center gap-2">
                <LoadingSpinner size="sm" color="secondary" />
                <span class="text-xs text-gray-400 dark:text-dark-400">
                  {{ t('user.payment.polling') }}
                </span>
              </div>
            </div>
          </div>

          <!-- Action Buttons -->
          <div class="flex gap-3">
            <button @click="goBackToForm" class="btn btn-secondary flex-1">
              {{ t('user.payment.backToForm') }}
            </button>
            <button
              @click="handleCancelActiveOrder"
              class="btn flex-1 border-red-300 text-red-600 hover:bg-red-50 dark:border-red-700 dark:text-red-400 dark:hover:bg-red-900/20"
            >
              {{ t('user.payment.cancelOrder') }}
            </button>
          </div>
        </template>

        <!-- ==================== Step 3: Result ==================== -->
        <template v-else-if="step === 'result'">
          <div class="card">
            <div class="p-8 text-center">
              <!-- Status Icon -->
              <div class="mb-4 flex justify-center">
                <!-- Completed -->
                <div
                  v-if="resultStatus === 'completed' || resultStatus === 'paid'"
                  class="flex h-20 w-20 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/30"
                >
                  <Icon
                    name="checkCircle"
                    size="xl"
                    class="!h-10 !w-10 text-green-500 dark:text-green-400"
                  />
                </div>
                <!-- Recharging -->
                <div
                  v-else-if="resultStatus === 'recharging'"
                  class="flex h-20 w-20 items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900/30"
                >
                  <Icon
                    name="sync"
                    size="xl"
                    class="!h-10 !w-10 animate-spin text-blue-500 dark:text-blue-400"
                  />
                </div>
                <!-- Failed -->
                <div
                  v-else-if="resultStatus === 'failed'"
                  class="flex h-20 w-20 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/30"
                >
                  <Icon
                    name="xCircle"
                    size="xl"
                    class="!h-10 !w-10 text-red-500 dark:text-red-400"
                  />
                </div>
                <!-- Cancelled -->
                <div
                  v-else-if="resultStatus === 'cancelled'"
                  class="flex h-20 w-20 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700"
                >
                  <Icon
                    name="xCircle"
                    size="xl"
                    class="!h-10 !w-10 text-gray-400 dark:text-dark-400"
                  />
                </div>
                <!-- Expired -->
                <div
                  v-else-if="resultStatus === 'expired'"
                  class="flex h-20 w-20 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700"
                >
                  <Icon
                    name="clock"
                    size="xl"
                    class="!h-10 !w-10 text-gray-400 dark:text-dark-400"
                  />
                </div>
                <!-- Fallback / Pending -->
                <div
                  v-else
                  class="flex h-20 w-20 items-center justify-center rounded-full bg-yellow-100 dark:bg-yellow-900/30"
                >
                  <Icon
                    name="exclamationCircle"
                    size="xl"
                    class="!h-10 !w-10 text-yellow-500 dark:text-yellow-400"
                  />
                </div>
              </div>

              <!-- Status Text -->
              <h3 class="text-xl font-bold text-gray-900 dark:text-white">
                {{ resultTitle }}
              </h3>
              <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
                {{ resultMessage }}
              </p>

              <!-- Recharging: still polling indicator -->
              <div
                v-if="resultStatus === 'recharging' && polling"
                class="mt-4 flex items-center justify-center gap-2"
              >
                <LoadingSpinner size="sm" color="secondary" />
                <span class="text-xs text-gray-400 dark:text-dark-400">
                  {{ t('user.payment.polling') }}
                </span>
              </div>

              <!-- Back Button -->
              <button @click="handleResultBack" class="btn btn-primary mt-6 w-full">
                {{ t('user.payment.backToForm') }}
              </button>
            </div>
          </div>
        </template>
      </template>

      <!-- Cancel Confirm Dialog -->
      <ConfirmDialog
        :show="showCancelConfirm"
        :title="t('user.payment.cancelOrder')"
        :message="t('user.payment.cancelConfirm')"
        :confirm-text="t('common.confirm')"
        :cancel-text="t('common.cancel')"
        danger
        @confirm="confirmCancelOrder"
        @cancel="showCancelConfirm = false"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { paymentAPI } from '@/api/payment'
import { formatDateTime } from '@/utils/format'
import {
  getPaymentStatusBadgeClass,
  getPaymentMethodLabel,
  AMOUNT_TEXT_PATTERN
} from '@/utils/payment'
import type {
  PaymentOrder,
  PaymentConfig,
  MethodLimit,
  PaymentType,
  PaymentStatus,
  CreateOrderResponse
} from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import QRCode from 'qrcode'

// ==================== Composables ====================

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const user = computed(() => authStore.user)

// ==================== State Machine ====================

type Step = 'form' | 'paying' | 'result'
const step = ref<Step>('form')

// ==================== Config & Data ====================

const configLoading = ref(false)
const config = ref<PaymentConfig | null>(null)
const creating = ref(false)
const ordersLoading = ref(false)
const polling = ref(false)

const orders = ref<PaymentOrder[]>([])
const activeOrder = ref<CreateOrderResponse | null>(null)
const resultStatus = ref<string>('pending')

// ==================== Quick Amounts ====================

const quickAmounts = [10, 20, 50, 100, 200, 500, 1000, 2000]
const selectedQuickAmount = ref<number | null>(10)
const customAmountText = ref('')

// ==================== Payment Method ====================

const selectedPaymentType = ref<PaymentType | ''>('')

// ==================== Cancel Dialog ====================

const showCancelConfirm = ref(false)
const cancellingOrderId = ref<number | null>(null)

// ==================== QR Code ====================

const qrDataUrl = ref<string>('')

// ==================== Countdown ====================

const countdownSeconds = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null
let pollTimer: ReturnType<typeof setInterval> | null = null

// ==================== Computed ====================

const availableMethods = computed<MethodLimit[]>(() => {
  return config.value?.limits ?? []
})

const minAmount = computed(() => {
  const val = config.value?.settings?.min_recharge_amount
  return val ? Number(val) : 1
})

const maxAmount = computed(() => {
  const val = config.value?.settings?.max_recharge_amount
  return val ? Number(val) : 10000
})

const maxPendingOrders = computed(() => {
  const val = config.value?.settings?.max_pending_orders
  return val ? Number(val) : 5
})

const pendingCount = computed(() => {
  const val = config.value?.settings?.pending_count
  return val ? Number(val) : 0
})

const hasTooManyPendingOrders = computed(() => {
  return pendingCount.value >= maxPendingOrders.value
})

const amountRangeHint = computed(() => {
  return `${minAmount.value} - ${maxAmount.value}`
})

const effectiveAmount = computed(() => {
  if (selectedQuickAmount.value !== null) {
    return selectedQuickAmount.value
  }
  const parsed = parseFloat(customAmountText.value)
  return isNaN(parsed) ? 0 : parsed
})

const amountError = computed(() => {
  if (effectiveAmount.value <= 0) return ''
  if (effectiveAmount.value < minAmount.value) {
    return t('user.payment.amountTooLow', { min: minAmount.value })
  }
  if (effectiveAmount.value > maxAmount.value) {
    return t('user.payment.amountTooHigh', { max: maxAmount.value })
  }
  return ''
})

const currentFeeRate = computed(() => {
  if (!selectedPaymentType.value) return 0
  const method = availableMethods.value.find(
    (m) => m.paymentType === selectedPaymentType.value
  )
  return method?.feeRate ?? 0
})

const feeAmount = computed(() => {
  return effectiveAmount.value * currentFeeRate.value
})

const totalPayAmount = computed(() => {
  return effectiveAmount.value + feeAmount.value
})

const canCreateOrder = computed(() => {
  if (!selectedPaymentType.value) return false
  if (effectiveAmount.value <= 0) return false
  if (effectiveAmount.value < minAmount.value) return false
  if (effectiveAmount.value > maxAmount.value) return false
  if (amountError.value) return false
  if (hasTooManyPendingOrders.value) return false
  return true
})

const submitButtonColorClass = computed(() => {
  switch (selectedPaymentType.value) {
    case 'alipay':
      return 'bg-[#00AEEF] hover:bg-[#009ED8] disabled:bg-[#00AEEF]/50'
    case 'wechat':
      return 'bg-[#2BB741] hover:bg-[#249E38] disabled:bg-[#2BB741]/50'
    case 'stripe':
      return 'bg-[#635bff] hover:bg-[#5549e6] disabled:bg-[#635bff]/50'
    default:
      return 'bg-primary-600 hover:bg-primary-700 disabled:bg-primary-600/50'
  }
})

const countdownText = computed(() => {
  if (countdownSeconds.value <= 0) return ''
  const mins = Math.floor(countdownSeconds.value / 60)
  const secs = countdownSeconds.value % 60
  return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`
})

const countdownUrgent = computed(() => {
  return countdownSeconds.value > 0 && countdownSeconds.value < 60
})

const resultTitle = computed(() => {
  switch (resultStatus.value) {
    case 'completed':
    case 'paid':
      return t('user.payment.paymentSuccess')
    case 'recharging':
      return t('user.payment.recharging')
    case 'failed':
      return t('user.payment.paymentFailed')
    case 'cancelled':
      return t('user.payment.orderCancelled')
    case 'expired':
      return t('user.payment.orderExpired')
    default:
      return t('user.payment.statusUnknown')
  }
})

const resultMessage = computed(() => {
  switch (resultStatus.value) {
    case 'completed':
    case 'paid':
      return t('user.payment.paymentSuccessDesc')
    case 'recharging':
      return t('user.payment.rechargingDesc')
    case 'failed':
      return t('user.payment.paymentFailedDesc')
    case 'cancelled':
      return t('user.payment.orderCancelledDesc')
    case 'expired':
      return t('user.payment.orderExpiredDesc')
    default:
      return ''
  }
})

// ==================== Payment Method Helpers ====================

function paymentMethodIconBg(type: PaymentType | string): string {
  switch (type) {
    case 'alipay':
      return 'bg-[#00AEEF]'
    case 'wechat':
      return 'bg-[#2BB741]'
    case 'stripe':
      return 'bg-[#635bff]'
    case 'usdt':
      return 'bg-emerald-500'
    default:
      return 'bg-gray-500'
  }
}

function paymentMethodIconText(type: PaymentType | string): string {
  switch (type) {
    case 'alipay':
      return 'A'
    case 'wechat':
      return 'W'
    case 'stripe':
      return 'S'
    case 'usdt':
      return 'U'
    default:
      return '?'
  }
}

function paymentMethodSelectedClass(type: PaymentType | string): string {
  switch (type) {
    case 'alipay':
      return 'border-cyan-400 bg-cyan-50 dark:border-cyan-500 dark:bg-cyan-900/20'
    case 'wechat':
      return 'border-green-500 bg-green-50 dark:border-green-500 dark:bg-green-900/20'
    case 'stripe':
      return 'border-[#635bff] bg-violet-50 dark:border-[#635bff] dark:bg-violet-900/20'
    case 'usdt':
      return 'border-emerald-500 bg-emerald-50 dark:border-emerald-500 dark:bg-emerald-900/20'
    default:
      return 'border-primary-500 bg-primary-50 dark:border-primary-400 dark:bg-primary-900/20'
  }
}

function orderStatusLabel(status: PaymentStatus | string): string {
  const key = `user.payment.status.${status}`
  const translated = t(key)
  // If the key doesn't exist, t() returns the key itself - fall back to the status string
  return translated === key ? status : translated
}

// ==================== Amount Handling ====================

function selectQuickAmount(amt: number) {
  selectedQuickAmount.value = amt
  customAmountText.value = ''
}

function handleCustomAmountInput() {
  // Clear quick amount selection when typing custom amount
  selectedQuickAmount.value = null
  // Validate: allow only numbers and up to 2 decimal places
  const raw = customAmountText.value
  if (raw && !AMOUNT_TEXT_PATTERN.test(raw)) {
    // Strip invalid chars but keep a valid partial input
    customAmountText.value = raw.replace(/[^\d.]/g, '')
    // Ensure only one decimal point
    const parts = customAmountText.value.split('.')
    if (parts.length > 2) {
      customAmountText.value = parts[0] + '.' + parts.slice(1).join('')
    }
    // Truncate to 2 decimal places
    if (parts.length === 2 && parts[1].length > 2) {
      customAmountText.value = parts[0] + '.' + parts[1].slice(0, 2)
    }
  }
}

function selectPaymentMethod(method: MethodLimit) {
  if (!method.available) return
  selectedPaymentType.value = method.paymentType
}

// ==================== QR Code Generation ====================

async function generateQrCode(qrString: string) {
  if (!qrString) {
    qrDataUrl.value = ''
    return
  }
  try {
    qrDataUrl.value = await QRCode.toDataURL(qrString, {
      width: 224,
      margin: 1,
      errorCorrectionLevel: 'M'
    })
  } catch (err) {
    console.error('Failed to generate QR code:', err)
    qrDataUrl.value = ''
  }
}

// ==================== Countdown ====================

function startCountdown(expiresAt: string | null) {
  stopCountdown()
  if (!expiresAt) return

  const updateCountdown = () => {
    const now = Date.now()
    const expires = new Date(expiresAt).getTime()
    const diff = Math.max(0, Math.floor((expires - now) / 1000))
    countdownSeconds.value = diff
    if (diff <= 0) {
      stopCountdown()
    }
  }

  updateCountdown()
  countdownTimer = setInterval(updateCountdown, 1000)
}

function stopCountdown() {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
  countdownSeconds.value = 0
}

// ==================== Polling ====================

function startPolling(orderId: number) {
  stopPolling()
  polling.value = true
  pollTimer = setInterval(async () => {
    try {
      const order = await paymentAPI.getOrder(orderId)
      const status = order.status as string
      const terminalStatuses = ['completed', 'paid', 'failed', 'cancelled', 'expired']

      if (terminalStatuses.includes(status)) {
        stopPolling()
        stopCountdown()
        resultStatus.value = status
        step.value = 'result'

        if (status === 'completed' || status === 'paid') {
          appStore.showSuccess(t('user.payment.paymentSuccess'))
          authStore.refreshUser()
        }
      } else if (status === 'recharging') {
        // Move to result but keep polling
        resultStatus.value = status
        step.value = 'result'
      }
    } catch {
      // Silently retry
    }
  }, 2000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  polling.value = false
}

// ==================== API Calls ====================

async function loadConfig() {
  configLoading.value = true
  try {
    config.value = await paymentAPI.getConfig()
    // Auto-select first available payment method
    const firstAvailable = availableMethods.value.find((m) => m.available)
    if (firstAvailable) {
      selectedPaymentType.value = firstAvailable.paymentType
    }
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('user.payment.loadFailed'))
    console.error('Failed to load payment config:', error)
  } finally {
    configLoading.value = false
  }
}

async function loadOrders() {
  ordersLoading.value = true
  try {
    const response = await paymentAPI.listOrders(1, 10)
    orders.value = response.items
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.detail || t('user.payment.loadOrdersFailed')
    )
    console.error('Failed to load orders:', error)
  } finally {
    ordersLoading.value = false
  }
}

async function handleCreateOrder() {
  if (!canCreateOrder.value || creating.value) return

  creating.value = true
  try {
    const result = await paymentAPI.createOrder({
      amount: effectiveAmount.value,
      payment_type: selectedPaymentType.value as PaymentType,
      order_type: 'balance',
      src_host: window.location.hostname,
      return_url: window.location.href
    })

    activeOrder.value = result
    appStore.showSuccess(t('user.payment.orderCreated'))

    // Generate QR code if qr_code is present
    if (result.qr_code) {
      await generateQrCode(result.qr_code)
    }

    // Start countdown
    startCountdown(result.order.expires_at)

    // Start polling
    startPolling(result.order.id)

    // Move to paying step
    step.value = 'paying'

    // Refresh orders list in background
    loadOrders()
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.detail || t('user.payment.createFailed')
    )
    console.error('Failed to create order:', error)
  } finally {
    creating.value = false
  }
}

// ==================== Cancel Handling ====================

function handleCancelActiveOrder() {
  if (!activeOrder.value) return
  cancellingOrderId.value = activeOrder.value.order.id
  showCancelConfirm.value = true
}

function handleCancelOrderFromHistory(orderId: number) {
  cancellingOrderId.value = orderId
  showCancelConfirm.value = true
}

async function confirmCancelOrder() {
  if (!cancellingOrderId.value) return
  try {
    await paymentAPI.cancelOrder(cancellingOrderId.value)
    appStore.showSuccess(t('user.payment.orderCancelled'))

    // If this was the active order, go to result
    if (activeOrder.value?.order.id === cancellingOrderId.value) {
      stopPolling()
      stopCountdown()
      resultStatus.value = 'cancelled'
      step.value = 'result'
      activeOrder.value = null
    }

    showCancelConfirm.value = false
    cancellingOrderId.value = null
    loadOrders()
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.detail || t('user.payment.cancelFailed')
    )
  }
}

// ==================== Navigation ====================

function goBackToForm() {
  stopPolling()
  stopCountdown()
  activeOrder.value = null
  qrDataUrl.value = ''
  step.value = 'form'
  loadOrders()
}

function handleResultBack() {
  stopPolling()
  stopCountdown()
  activeOrder.value = null
  qrDataUrl.value = ''
  resultStatus.value = 'pending'
  step.value = 'form'
  // Refresh balance and orders
  authStore.refreshUser()
  loadOrders()
  loadConfig()
}

// ==================== Lifecycle ====================

onMounted(() => {
  loadConfig()
  loadOrders()
})

onUnmounted(() => {
  stopPolling()
  stopCountdown()
})
</script>
