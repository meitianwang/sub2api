import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { nextTick } from 'vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    cachedPublicSettings: { api_base_url: 'https://api.example.com' }
  })
}))

import ModelDetailDrawer from '../ModelDetailDrawer.vue'
import type { ModelEntry } from '../providerUtils'

const claudeModel: ModelEntry = {
  model_id: 'claude-opus-4-6',
  display_name: 'Claude Opus 4.6',
  provider: 'claude',
  group_id: 1,
  group_name: 'bat-claude',
  input_price: 16,
  output_price: 80
}

function mountDrawer(props: { open: boolean; model: ModelEntry | null }) {
  return mount(ModelDetailDrawer, {
    props,
    attachTo: document.body,
    global: {
      stubs: {
        ProviderBrandIcon: { template: '<span class="brand-stub" />' },
        Icon: { template: '<span class="icon-stub" />' }
      }
    }
  })
}

describe('ModelDetailDrawer', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
    document.body.classList.remove('modal-open')
  })

  it('renders nothing when closed', () => {
    mountDrawer({ open: false, model: claudeModel })
    expect(document.querySelector('.drawer-overlay')).toBeNull()
  })

  it('renders nothing when model is null', () => {
    mountDrawer({ open: true, model: null })
    expect(document.querySelector('.drawer-overlay')).toBeNull()
  })

  it('renders model fields and curl when open with model', async () => {
    mountDrawer({ open: true, model: claudeModel })
    await nextTick()
    const overlay = document.querySelector('.drawer-overlay')
    expect(overlay).not.toBeNull()
    const text = overlay!.textContent || ''
    expect(text).toContain('Claude Opus 4.6')
    expect(text).toContain('claude-opus-4-6')
    expect(text).toContain('Anthropic')
    expect(text).toContain('bat-claude')
    const code = document.querySelector('.drawer-curl-code')!.textContent || ''
    expect(code).toContain('https://api.example.com/v1/messages')
    expect(code).toContain('"model": "claude-opus-4-6"')
    expect(code).toContain('x-api-key: sk-xxxxxx')
  })

  it('emits close when ESC pressed while open', async () => {
    const wrapper = mountDrawer({ open: true, model: claudeModel })
    await nextTick()
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await flushPromises()
    expect(wrapper.emitted('close')).toBeTruthy()
  })

  it('does not emit close on ESC when closed', async () => {
    const wrapper = mountDrawer({ open: false, model: claudeModel })
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await flushPromises()
    expect(wrapper.emitted('close')).toBeFalsy()
  })

  it('emits close when overlay backdrop clicked', async () => {
    const wrapper = mountDrawer({ open: true, model: claudeModel })
    await nextTick()
    const overlay = document.querySelector('.drawer-overlay') as HTMLElement
    overlay.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await flushPromises()
    expect(wrapper.emitted('close')).toBeTruthy()
  })

  it('locks body scroll while open', async () => {
    mountDrawer({ open: true, model: claudeModel })
    await nextTick()
    expect(document.body.classList.contains('modal-open')).toBe(true)
  })

  it('uses OpenAI format curl for openai provider', async () => {
    const openaiModel: ModelEntry = { ...claudeModel, model_id: 'gpt-5', provider: 'openai' }
    mountDrawer({ open: true, model: openaiModel })
    await nextTick()
    const code = document.querySelector('.drawer-curl-code')!.textContent || ''
    expect(code).toContain('/v1/chat/completions')
    expect(code).toContain('Authorization: Bearer sk-xxxxxx')
  })

  it('uses Gemini format curl for gemini provider', async () => {
    const geminiModel: ModelEntry = { ...claudeModel, model_id: 'gemini-2.0-flash', provider: 'gemini' }
    mountDrawer({ open: true, model: geminiModel })
    await nextTick()
    const code = document.querySelector('.drawer-curl-code')!.textContent || ''
    expect(code).toContain('/v1beta/models/gemini-2.0-flash:generateContent')
    expect(code).toContain('x-goog-api-key: sk-xxxxxx')
  })
})
