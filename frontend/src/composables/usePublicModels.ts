import { ref, computed, type Ref, type ComputedRef } from 'vue'
import { apiClient } from '@/api/client'
import type { ModelEntry } from '@/components/models/providerUtils'

const TTL_MS = 5 * 60 * 1000

const items = ref<ModelEntry[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
let lastFetched = 0
let inflight: Promise<void> | null = null

async function doFetch(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const res = await apiClient.get('/settings/models')
    items.value = (res.data as ModelEntry[]) || []
    lastFetched = Date.now()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'fetch failed'
    items.value = []
  } finally {
    loading.value = false
  }
}

export function usePublicModels(): {
  items: Ref<ModelEntry[]>
  loading: Ref<boolean>
  error: Ref<string | null>
  fetch: (force?: boolean) => Promise<void>
  uniqueModelCount: ComputedRef<number>
} {
  function fetch(force = false): Promise<void> {
    const fresh = Date.now() - lastFetched < TTL_MS && items.value.length > 0
    if (!force && fresh) return Promise.resolve()
    if (inflight) return inflight
    inflight = doFetch().finally(() => { inflight = null })
    return inflight
  }

  const uniqueModelCount = computed(() => new Set(items.value.map(i => i.model_id)).size)

  return { items, loading, error, fetch, uniqueModelCount }
}

// Test-only helper to reset the module-level cache between tests.
export function __resetPublicModelsCache(): void {
  items.value = []
  loading.value = false
  error.value = null
  lastFetched = 0
  inflight = null
}
