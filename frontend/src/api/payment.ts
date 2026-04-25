/**
 * Payment API endpoints (user-facing)
 * Handles order creation, listing, and payment config retrieval
 */

import { apiClient } from './client'
import type { BasePaginationResponse } from '@/types'
import type {
  PaymentOrder,
  PaymentConfig,
  PaymentChannel,
  CreateOrderRequest,
  CreateOrderResponse
} from '@/types/payment'

/**
 * Create a new payment order
 */
export async function createOrder(req: CreateOrderRequest): Promise<CreateOrderResponse> {
  const { data } = await apiClient.post<CreateOrderResponse>('/pay/orders', req)
  return data
}

/**
 * List current user's payment orders
 */
export async function listOrders(
  page: number = 1,
  pageSize: number = 20
): Promise<BasePaginationResponse<PaymentOrder>> {
  const { data } = await apiClient.get<BasePaginationResponse<PaymentOrder>>('/pay/orders', {
    params: { page, page_size: pageSize }
  })
  return data
}

/**
 * Get a single order by ID
 */
export async function getOrder(id: number): Promise<PaymentOrder> {
  const { data } = await apiClient.get<PaymentOrder>(`/pay/orders/${id}`)
  return data
}

/**
 * Cancel a pending order
 */
export async function cancelOrder(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>(`/pay/orders/${id}/cancel`)
  return data
}

/**
 * Request a refund for a completed order
 */
export async function requestRefund(
  id: number,
  amount: number,
  reason: string
): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>(`/pay/orders/${id}/refund-request`, {
    amount,
    reason
  })
  return data
}

/**
 * Get payment configuration and method limits
 */
export async function getConfig(): Promise<PaymentConfig> {
  const { data } = await apiClient.get<PaymentConfig>('/pay/config')
  return data
}

/**
 * List available payment channels
 */
export async function listChannels(): Promise<PaymentChannel[]> {
  const { data } = await apiClient.get<PaymentChannel[]>('/pay/channels')
  return data
}

export const paymentAPI = {
  createOrder,
  listOrders,
  getOrder,
  cancelOrder,
  requestRefund,
  getConfig,
  listChannels
}

export default paymentAPI
