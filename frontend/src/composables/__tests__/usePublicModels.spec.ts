import { describe, expect, it, vi, beforeEach } from 'vitest'

const mockGet = vi.fn()

vi.mock('@/api/client', () => ({
  apiClient: { get: (...args: unknown[]) => mockGet(...args) }
}))

import { usePublicModels, __resetPublicModelsCache } from '../usePublicModels'

const sample = [
  { model_id: 'a', display_name: 'A', provider: 'claude', group_id: 1, group_name: 'g', input_price: 1, output_price: 2 },
  { model_id: 'b', display_name: 'B', provider: 'openai', group_id: 1, group_name: 'g', input_price: 3, output_price: 4 },
  { model_id: 'a', display_name: 'A', provider: 'claude', group_id: 2, group_name: 'g2', input_price: 1, output_price: 2 }
]

describe('usePublicModels', () => {
  beforeEach(() => {
    mockGet.mockReset()
    __resetPublicModelsCache()
  })

  it('fetches and stores model list', async () => {
    mockGet.mockResolvedValueOnce({ data: sample })
    const { items, fetch, loading } = usePublicModels()
    await fetch()
    expect(loading.value).toBe(false)
    expect(items.value).toEqual(sample)
    expect(mockGet).toHaveBeenCalledTimes(1)
  })

  it('caches between calls within TTL', async () => {
    mockGet.mockResolvedValueOnce({ data: sample })
    const { fetch } = usePublicModels()
    await fetch()
    await fetch()
    await fetch()
    expect(mockGet).toHaveBeenCalledTimes(1)
  })

  it('force=true bypasses cache', async () => {
    mockGet.mockResolvedValue({ data: sample })
    const { fetch } = usePublicModels()
    await fetch()
    await fetch(true)
    expect(mockGet).toHaveBeenCalledTimes(2)
  })

  it('dedupes concurrent inflight requests', async () => {
    mockGet.mockResolvedValueOnce({ data: sample })
    const { fetch } = usePublicModels()
    await Promise.all([fetch(), fetch(), fetch()])
    expect(mockGet).toHaveBeenCalledTimes(1)
  })

  it('uniqueModelCount counts distinct model_id', async () => {
    mockGet.mockResolvedValueOnce({ data: sample })
    const { fetch, uniqueModelCount } = usePublicModels()
    await fetch()
    expect(uniqueModelCount.value).toBe(2)
  })

  it('handles fetch error and clears items', async () => {
    mockGet.mockRejectedValueOnce(new Error('boom'))
    const { fetch, items, error } = usePublicModels()
    await fetch()
    expect(items.value).toEqual([])
    expect(error.value).toBe('boom')
  })
})
