<template>
  <Teleport to="body">
    <Transition name="drawer">
      <div
        v-if="open && model"
        class="drawer-overlay"
        role="dialog"
        aria-modal="true"
        @click.self="emit('close')"
      >
        <div class="drawer-panel" @click.stop>
          <header class="drawer-header">
            <ProviderBrandIcon :provider="model.provider" circle class="h-9 w-9" />
            <h3 class="drawer-title">{{ model.display_name }}</h3>
            <button
              @click="emit('close')"
              class="drawer-close"
              :aria-label="t('models.detail.close')"
            >
              <Icon name="x" size="md" />
            </button>
          </header>

          <div class="drawer-body">
            <section class="drawer-info">
              <div class="drawer-info-row">
                <span class="drawer-info-label">{{ t('models.detail.modelId') }}</span>
                <div class="drawer-info-value">
                  <code class="drawer-model-id">{{ model.model_id }}</code>
                  <button
                    @click="copyModelId"
                    class="drawer-inline-copy"
                    :aria-label="t('models.detail.copy')"
                    :title="t('models.detail.copy')"
                  >
                    <Icon :name="modelIdCopied ? 'check' : 'copy'" size="sm" />
                  </button>
                </div>
              </div>
              <div class="drawer-info-row">
                <span class="drawer-info-label">{{ t('models.detail.provider') }}</span>
                <span class="drawer-info-text">{{ providerLabel(model.provider) }}</span>
              </div>
              <div class="drawer-info-row">
                <span class="drawer-info-label">{{ t('models.detail.group') }}</span>
                <span class="drawer-info-badge">{{ model.group_name }}</span>
              </div>
              <div class="drawer-info-row">
                <span class="drawer-info-label">{{ t('models.detail.pricing') }}</span>
                <span class="drawer-info-text">
                  {{ t('models.pricing.input') }} ¥{{ fmtPrice(model.input_price) }}/M ·
                  {{ t('models.pricing.output') }} ¥{{ fmtPrice(model.output_price) }}/M
                </span>
              </div>
            </section>

            <section class="drawer-curl">
              <div class="drawer-curl-header">
                <h4 class="drawer-section-title">
                  {{ t('models.detail.example') }}
                  <span class="drawer-format-tag">{{ formatLabel }}</span>
                </h4>
                <button @click="copyCurl" class="drawer-curl-copy">
                  <Icon :name="curlCopied ? 'check' : 'copy'" size="sm" />
                  <span>{{ curlCopied ? t('models.detail.copied') : t('models.detail.copy') }}</span>
                </button>
              </div>
              <pre class="drawer-curl-code"><code>{{ curlExample.code }}</code></pre>
            </section>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'
import ProviderBrandIcon from '@/components/icons/ProviderBrandIcon.vue'
import { providerLabel, fmtPrice, type ModelEntry } from './providerUtils'
import { buildCurlExample } from './buildCurlExample'

const props = defineProps<{
  open: boolean
  model: ModelEntry | null
}>()

const emit = defineEmits<{ close: [] }>()

const { t } = useI18n()
const appStore = useAppStore()

const curlExample = computed(() => {
  if (!props.model) return { format: 'openai' as const, code: '' }
  const apiBaseUrl = appStore.cachedPublicSettings?.api_base_url || ''
  return buildCurlExample(props.model.model_id, props.model.provider, apiBaseUrl)
})

const formatLabel = computed(() => {
  const f = curlExample.value.format
  if (f === 'anthropic') return t('models.detail.formatAnthropic')
  if (f === 'gemini') return t('models.detail.formatGemini')
  return t('models.detail.formatOpenAI')
})

const modelIdCopied = ref(false)
const curlCopied = ref(false)

async function writeText(text: string): Promise<boolean> {
  if (navigator.clipboard && window.isSecureContext) {
    try {
      await navigator.clipboard.writeText(text)
      return true
    } catch {
      // fall through
    }
  }
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.style.cssText = 'position:fixed;left:-9999px;top:-9999px'
  document.body.appendChild(textarea)
  textarea.select()
  try {
    return document.execCommand('copy')
  } finally {
    document.body.removeChild(textarea)
  }
}

async function copyModelId() {
  if (!props.model) return
  if (await writeText(props.model.model_id)) {
    modelIdCopied.value = true
    setTimeout(() => { modelIdCopied.value = false }, 1500)
  }
}

async function copyCurl() {
  if (!curlExample.value.code) return
  if (await writeText(curlExample.value.code)) {
    curlCopied.value = true
    setTimeout(() => { curlCopied.value = false }, 1500)
  }
}

function handleEscape(event: KeyboardEvent) {
  if (props.open && event.key === 'Escape') emit('close')
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) document.body.classList.add('modal-open')
    else document.body.classList.remove('modal-open')
  },
  { immediate: true }
)

onMounted(() => document.addEventListener('keydown', handleEscape))
onUnmounted(() => {
  document.removeEventListener('keydown', handleEscape)
  document.body.classList.remove('modal-open')
})
</script>
