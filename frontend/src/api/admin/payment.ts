/**
 * Admin Payment API endpoints
 * Handles order management, provider instances, channels, plans, and dashboard
 */

import { apiClient } from '../client'
import type { BasePaginationResponse } from '@/types'
import type {
  PaymentOrder,
  PaymentConfig,
  PaymentChannel,
  SubscriptionPlan,
  PaymentAuditLog,
  ProviderInstance,
  PaymentDashboardStats,
  PaymentStatus,
  PaymentType,
  OrderType
} from '@/types/payment'

// ==================== Orders ====================

/**
 * List all payment orders with filters
 */
export async function listOrders(
  filters?: {
    page?: number
    page_size?: number
    status?: PaymentStatus
    payment_type?: PaymentType
    order_type?: OrderType
    user_id?: number
    search?: string
  },
  options?: { signal?: AbortSignal }
): Promise<BasePaginationResponse<PaymentOrder>> {
  const { data } = await apiClient.get<BasePaginationResponse<PaymentOrder>>('/admin/pay/orders', {
    params: filters,
    signal: options?.signal
  })
  return data
}

/**
 * Get order detail including audit logs
 */
export async function getOrderDetail(id: number): Promise<{
  order: PaymentOrder
  audit_logs: PaymentAuditLog[]
}> {
  const { data } = await apiClient.get<{
    order: PaymentOrder
    audit_logs: PaymentAuditLog[]
  }>(`/admin/pay/orders/${id}`)
  return data
}

/**
 * Cancel a pending order
 */
export async function cancelOrder(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>(`/admin/pay/orders/${id}/cancel`)
  return data
}

/**
 * Retry recharge for a paid order that failed to credit
 */
export async function retryRecharge(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>(`/admin/pay/orders/${id}/retry`)
  return data
}

/**
 * Process a refund for an order
 */
export async function processRefund(
  orderId: number,
  amount: number,
  reason: string
): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>('/admin/pay/refund', {
    order_id: orderId,
    amount,
    reason
  })
  return data
}

// ==================== Config ====================

/**
 * Get payment system configuration
 */
export async function getConfig(): Promise<PaymentConfig> {
  const { data } = await apiClient.get<PaymentConfig>('/admin/pay/config')
  return data
}

/**
 * Update payment system configuration
 */
export async function updateConfig(settings: Record<string, string>): Promise<{ message: string }> {
  const { data } = await apiClient.put<{ message: string }>('/admin/pay/config', { settings })
  return data
}

// ==================== Provider Instances ====================

/**
 * List payment provider instances
 */
export async function listProviderInstances(
  providerKey?: string
): Promise<ProviderInstance[]> {
  const { data } = await apiClient.get<ProviderInstance[]>('/admin/pay/provider-instances', {
    params: providerKey ? { provider_key: providerKey } : undefined
  })
  return data
}

/**
 * Create a new provider instance
 */
export async function createProviderInstance(
  instance: Partial<ProviderInstance>
): Promise<ProviderInstance> {
  const { data } = await apiClient.post<ProviderInstance>('/admin/pay/provider-instances', instance)
  return data
}

/**
 * Update a provider instance
 */
export async function updateProviderInstance(
  id: number,
  instance: Partial<ProviderInstance>
): Promise<ProviderInstance> {
  const { data } = await apiClient.put<ProviderInstance>(`/admin/pay/provider-instances/${id}`, instance)
  return data
}

/**
 * Delete a provider instance
 */
export async function deleteProviderInstance(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/admin/pay/provider-instances/${id}`)
  return data
}

// ==================== Channels ====================

/**
 * List payment channels
 */
export async function listChannels(): Promise<PaymentChannel[]> {
  const { data } = await apiClient.get<PaymentChannel[]>('/admin/pay/channels')
  return data
}

/**
 * Create a new payment channel
 */
export async function createChannel(channel: Partial<PaymentChannel>): Promise<PaymentChannel> {
  const { data } = await apiClient.post<PaymentChannel>('/admin/pay/channels', channel)
  return data
}

/**
 * Update a payment channel
 */
export async function updateChannel(
  id: number,
  channel: Partial<PaymentChannel>
): Promise<PaymentChannel> {
  const { data } = await apiClient.put<PaymentChannel>(`/admin/pay/channels/${id}`, channel)
  return data
}

/**
 * Delete a payment channel
 */
export async function deleteChannel(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/admin/pay/channels/${id}`)
  return data
}

// ==================== Subscription Plans ====================

/**
 * List subscription plans
 */
export async function listPlans(): Promise<SubscriptionPlan[]> {
  const { data } = await apiClient.get<SubscriptionPlan[]>('/admin/pay/subscription-plans')
  return data
}

/**
 * Create a new subscription plan
 */
export async function createPlan(plan: Partial<SubscriptionPlan>): Promise<SubscriptionPlan> {
  const { data } = await apiClient.post<SubscriptionPlan>('/admin/pay/subscription-plans', plan)
  return data
}

/**
 * Update a subscription plan
 */
export async function updatePlan(
  id: number,
  plan: Partial<SubscriptionPlan>
): Promise<SubscriptionPlan> {
  const { data } = await apiClient.put<SubscriptionPlan>(`/admin/pay/subscription-plans/${id}`, plan)
  return data
}

/**
 * Delete a subscription plan
 */
export async function deletePlan(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/admin/pay/subscription-plans/${id}`)
  return data
}

// ==================== Dashboard ====================

/**
 * Get payment dashboard statistics
 */
export async function getDashboard(days?: number): Promise<PaymentDashboardStats> {
  const { data } = await apiClient.get<PaymentDashboardStats>('/admin/pay/dashboard', {
    params: days ? { days } : undefined
  })
  return data
}

export const paymentAdminAPI = {
  listOrders,
  getOrderDetail,
  cancelOrder,
  retryRecharge,
  processRefund,
  getConfig,
  updateConfig,
  listProviderInstances,
  createProviderInstance,
  updateProviderInstance,
  deleteProviderInstance,
  listChannels,
  createChannel,
  updateChannel,
  deleteChannel,
  listPlans,
  createPlan,
  updatePlan,
  deletePlan,
  getDashboard
}

export default paymentAdminAPI
