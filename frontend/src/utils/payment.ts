/**
 * Payment utility functions
 * Status labels, badge classes, validation, and constants
 */

import type { PaymentStatus, PaymentType } from '@/types/payment'
import { formatDate } from '@/utils/format'

// ==================== Constants ====================

export const PAYMENT_STATUSES: PaymentStatus[] = [
  'pending',
  'paid',
  'completed',
  'failed',
  'cancelled',
  'expired',
  'refunded',
  'partially_refunded',
  'refund_pending'
]

export const PAYMENT_TYPES: PaymentType[] = [
  'alipay',
  'wechat',
  'stripe',
  'usdt',
  'balance'
]

/** Regex for decimal amounts with up to 2 decimal places */
export const AMOUNT_TEXT_PATTERN = /^\d+(\.\d{0,2})?$/

// ==================== Status Helpers ====================

/**
 * Get Tailwind badge classes for a payment status
 */
export function getPaymentStatusBadgeClass(status: PaymentStatus): string {
  switch (status) {
    case 'completed':
      return 'badge-success'
    case 'paid':
      return 'badge-success'
    case 'pending':
      return 'badge-warning'
    case 'refund_pending':
      return 'badge-warning'
    case 'failed':
      return 'badge-error'
    case 'cancelled':
      return 'badge-error'
    case 'expired':
      return 'badge-error'
    case 'refunded':
      return 'badge-info'
    case 'partially_refunded':
      return 'badge-info'
    default:
      return 'badge-default'
  }
}

/**
 * Get human-readable label for a payment status
 */
export function getPaymentStatusLabel(status: PaymentStatus): string {
  const labels: Record<PaymentStatus, string> = {
    pending: '待支付',
    paid: '已支付',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消',
    expired: '已过期',
    refunded: '已退款',
    partially_refunded: '部分退款',
    refund_pending: '退款中'
  }
  return labels[status] ?? status
}

// ==================== Payment Method Helpers ====================

/**
 * Get human-readable label for a payment method type
 */
export function getPaymentMethodLabel(type: PaymentType): string {
  const labels: Record<PaymentType, string> = {
    alipay: '支付宝',
    wechat: '微信支付',
    stripe: 'Stripe',
    usdt: 'USDT',
    balance: '余额支付'
  }
  return labels[type] ?? type
}

// ==================== Date Formatting ====================

/**
 * Format an ISO date string for payment display
 */
export function formatPaymentDate(iso: string | null | undefined): string {
  return formatDate(iso)
}

// ==================== URL Validation ====================

/**
 * Check whether a payment URL is safe to open
 * Rejects non-http(s) protocols and URLs with embedded credentials
 */
export function isSafePaymentUrl(url: string | null | undefined): boolean {
  if (!url) return false

  try {
    const parsed = new URL(url)
    // Only allow http and https
    if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
      return false
    }
    // Reject embedded credentials (user:pass@host)
    if (parsed.username || parsed.password) {
      return false
    }
    return true
  } catch {
    return false
  }
}
