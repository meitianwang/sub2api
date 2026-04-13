import type { Account } from '@/types'

export const buildOpenAIUsageRefreshKey = (_account: Pick<Account, 'id' | 'type' | 'updated_at' | 'last_used_at' | 'rate_limit_reset_at' | 'extra'>): string => {
  // Only Anthropic platform is supported; this function is no longer relevant
  return ''
}
