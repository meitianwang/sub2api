/**
 * Payment System Type Definitions
 */

// ==================== Status & Type Literals ====================

export type PaymentStatus =
  | 'pending'
  | 'paid'
  | 'completed'
  | 'failed'
  | 'cancelled'
  | 'expired'
  | 'refunded'
  | 'partially_refunded'
  | 'refund_pending'

export type PaymentType =
  | 'alipay'
  | 'wechat'
  | 'stripe'
  | 'usdt'
  | 'balance'

export type OrderType =
  | 'recharge'
  | 'balance'
  | 'subscription'

// ==================== Core Entities ====================

export interface PaymentOrder {
  id: number
  user_id: number
  user_email: string
  user_name: string
  amount: number
  pay_amount: number
  fee_rate: number
  recharge_code: string
  status: PaymentStatus
  payment_type: PaymentType
  payment_trade_no: string
  pay_url: string
  qr_code: string
  qr_code_img: string
  refund_amount: number
  refund_reason: string
  refund_at: string | null
  force_refund: boolean
  refund_requested_at: string | null
  refund_request_reason: string
  expires_at: string | null
  paid_at: string | null
  completed_at: string | null
  failed_at: string | null
  failed_reason: string
  client_ip: string
  src_host: string
  src_url: string
  order_type: OrderType
  plan_id: number | null
  subscription_group_id: number | null
  subscription_days: number | null
  provider_instance_id: number | null
  created_at: string
  updated_at: string
}

export interface MethodLimit {
  payment_type: string
  available: boolean
  daily_limit: string
  daily_used: string
  remaining: string
  single_min: string
  single_max: string
  fee_rate: string
}

export interface PaymentConfig {
  enabled_payment_types: string[]
  min_recharge_amount: string
  max_recharge_amount: string
  max_daily_recharge_amount: string
  balance_payment_disabled: boolean
  max_pending_orders: number
  pending_count: number
  method_limits: MethodLimit[]
}

export interface PaymentChannel {
  id: number
  group_id: number | null
  name: string
  platform: string
  rate_multiplier: number
  description: string
  models: string[]
  features: string[]
  sort_order: number
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface SubscriptionPlan {
  id: number
  group_id: number | null
  name: string
  description: string
  price: number
  original_price: number
  validity_days: number
  validity_unit: string
  features: string[]
  product_name: string
  for_sale: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

export interface PaymentAuditLog {
  id: number
  order_id: number
  action: string
  detail: string
  operator: string
  created_at: string
}

export interface ProviderInstance {
  id: number
  provider_key: string
  name: string
  config: Record<string, unknown>
  supported_types: PaymentType[]
  enabled: boolean
  sort_order: number
  limits: Record<string, unknown>
  refund_enabled: boolean
  created_at: string
  updated_at: string
}

// ==================== Dashboard ====================

export interface PaymentDashboardStats {
  today_amount: number
  today_order_count: number
  total_amount: number
  total_order_count: number
  daily_series: { date: string; amount: number; count: number }[]
  payment_methods: { type: PaymentType; amount: number; count: number }[]
  leaderboard: { user_id: number; user_email: string; total_amount: number; order_count: number }[]
}

// ==================== Request / Response ====================

export interface CreateOrderRequest {
  amount: string
  payment_type: string
  order_type: OrderType
  plan_id?: number
  return_url?: string
  src_host?: string
  src_url?: string
}

export interface CreateOrderResponse {
  order: PaymentOrder
  pay_url: string
  qr_code: string
  client_secret: string
  access_token: string
}
