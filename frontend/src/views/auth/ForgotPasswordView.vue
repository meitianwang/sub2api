<template>
  <AuthLayout>
    <div class="font-mono text-xs text-gray-500 dark:text-gray-500">auth / reset</div>
    <h1 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl">
      {{ t('auth.forgotPasswordTitle') }}
    </h1>
    <p class="mt-3 text-base text-gray-600 dark:text-gray-400">
      {{ t('auth.forgotPasswordHint') }}
    </p>

    <div class="mt-10">
      <!-- Success State -->
      <div v-if="isSubmitted">
        <div class="rounded-md border border-emerald-500/40 bg-emerald-50/40 px-4 py-4 dark:border-emerald-500/30 dark:bg-emerald-950/20">
          <p class="font-mono text-[11px] uppercase tracking-[0.15em] text-emerald-700 dark:text-emerald-500">
            ✓ {{ t('auth.resetEmailSent') }}
          </p>
          <p class="mt-2 text-sm text-gray-700 dark:text-gray-300">
            {{ t('auth.resetEmailSentHint') }}
          </p>
        </div>

        <div class="mt-6">
          <router-link
            to="/login"
            class="inline-flex items-center gap-1.5 text-sm text-gray-900 underline-offset-4 hover:underline dark:text-white"
          >
            <Icon name="arrowLeft" size="sm" />
            {{ t('auth.backToLogin') }}
          </router-link>
        </div>
      </div>

      <!-- Form -->
      <form v-else @submit.prevent="handleSubmit" class="space-y-6">
        <!-- Email -->
        <div>
          <label for="email" class="mb-2 block font-mono text-[11px] uppercase tracking-[0.15em] text-gray-500 dark:text-gray-500">
            {{ t('auth.emailLabel') }}
          </label>
          <input
            id="email"
            v-model="formData.email"
            type="email"
            required
            autofocus
            autocomplete="email"
            :disabled="isLoading"
            :class="inputClass(!!errors.email)"
            :placeholder="t('auth.emailPlaceholder')"
          />
          <p v-if="errors.email" class="mt-1.5 text-xs text-red-600 dark:text-red-400">
            {{ errors.email }}
          </p>
        </div>

        <!-- Turnstile -->
        <div v-if="turnstileEnabled && turnstileSiteKey">
          <TurnstileWidget
            ref="turnstileRef"
            :site-key="turnstileSiteKey"
            @verify="onTurnstileVerify"
            @expire="onTurnstileExpire"
            @error="onTurnstileError"
          />
          <p v-if="errors.turnstile" class="mt-1.5 text-xs text-red-600 dark:text-red-400">
            {{ errors.turnstile }}
          </p>
        </div>

        <!-- Error -->
        <transition name="auth-fade">
          <p
            v-if="errorMessage"
            class="border-l-2 border-red-500 pl-3 text-sm text-red-600 dark:text-red-400"
          >
            {{ errorMessage }}
          </p>
        </transition>

        <!-- Submit -->
        <button
          type="submit"
          :disabled="isLoading || (turnstileEnabled && !turnstileToken)"
          class="inline-flex w-full items-center justify-center gap-1.5 rounded-md bg-gray-900 px-4 py-2.5 text-sm font-medium text-white transition-shadow hover:shadow-[0_4px_20px_-4px_rgba(59,130,246,0.5)] disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:shadow-none dark:bg-white dark:text-gray-900 dark:hover:shadow-[0_4px_20px_-4px_rgba(139,92,246,0.4)]"
        >
          <span
            v-if="isLoading"
            class="inline-block h-3.5 w-3.5 animate-spin rounded-full border-[1.5px] border-current border-t-transparent"
          ></span>
          <span>{{ isLoading ? t('auth.sendingResetLink') : t('auth.sendResetLink') }}</span>
          <Icon v-if="!isLoading" name="arrowRight" size="sm" />
        </button>
      </form>
    </div>

    <!-- Footer -->
    <template #footer>
      <p>
        {{ t('auth.rememberedPassword') }}
        <router-link
          to="/login"
          class="font-medium text-gray-900 underline underline-offset-4 hover:text-gray-700 dark:text-white dark:hover:text-gray-300"
        >
          {{ t('auth.signIn') }}
        </router-link>
      </p>
    </template>
  </AuthLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import Icon from '@/components/icons/Icon.vue'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import { useAppStore } from '@/stores'
import { getPublicSettings, forgotPassword } from '@/api/auth'
import { inputClass } from '@/components/auth/authStyles'

const { t } = useI18n()

const appStore = useAppStore()

const isLoading = ref<boolean>(false)
const isSubmitted = ref<boolean>(false)
const errorMessage = ref<string>('')

const turnstileEnabled = ref<boolean>(false)
const turnstileSiteKey = ref<string>('')

const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref<string>('')

const formData = reactive({ email: '' })
const errors = reactive({ email: '', turnstile: '' })

onMounted(async () => {
  try {
    const settings = await getPublicSettings()
    turnstileEnabled.value = settings.turnstile_enabled
    turnstileSiteKey.value = settings.turnstile_site_key || ''
  } catch (error) {
    console.error('Failed to load public settings:', error)
  }
})

function onTurnstileVerify(token: string): void {
  turnstileToken.value = token
  errors.turnstile = ''
}

function onTurnstileExpire(): void {
  turnstileToken.value = ''
  errors.turnstile = t('auth.turnstileExpired')
}

function onTurnstileError(): void {
  turnstileToken.value = ''
  errors.turnstile = t('auth.turnstileFailed')
}

function validateForm(): boolean {
  errors.email = ''
  errors.turnstile = ''

  let isValid = true

  if (!formData.email.trim()) {
    errors.email = t('auth.emailRequired')
    isValid = false
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
    errors.email = t('auth.invalidEmail')
    isValid = false
  }

  if (turnstileEnabled.value && !turnstileToken.value) {
    errors.turnstile = t('auth.completeVerification')
    isValid = false
  }

  return isValid
}

async function handleSubmit(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) return

  isLoading.value = true

  try {
    await forgotPassword({
      email: formData.email,
      turnstile_token: turnstileEnabled.value ? turnstileToken.value : undefined
    })

    isSubmitted.value = true
    appStore.showSuccess(t('auth.resetEmailSent'))
  } catch (error: unknown) {
    if (turnstileRef.value) {
      turnstileRef.value.reset()
      turnstileToken.value = ''
    }

    const err = error as { message?: string; response?: { data?: { detail?: string } } }

    if (err.response?.data?.detail) {
      errorMessage.value = err.response.data.detail
    } else if (err.message) {
      errorMessage.value = err.message
    } else {
      errorMessage.value = t('auth.sendResetLinkFailed')
    }

    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}
</script>

<style scoped>
.auth-fade-enter-active,
.auth-fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.auth-fade-enter-from,
.auth-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
