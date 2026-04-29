<template>
  <AuthLayout>
    <!-- Header: doc-style crumb + title -->
    <div class="font-mono text-xs text-gray-500 dark:text-gray-500">auth / login</div>
    <h1 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl">
      {{ t('auth.welcomeBack') }}
    </h1>
    <p class="mt-3 text-base text-gray-600 dark:text-gray-400">
      {{ t('auth.signInToAccount') }}
    </p>

    <div class="mt-10">
      <!-- LinuxDo OAuth -->
      <div v-if="linuxdoOAuthEnabled && !backendModeEnabled" class="mb-8">
        <LinuxDoOAuthSection :disabled="isLoading" />
      </div>

      <!-- Login Form -->
      <form @submit.prevent="handleLogin" class="space-y-6">
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

        <!-- Password -->
        <div>
          <div class="mb-2 flex items-baseline justify-between">
            <label for="password" class="block font-mono text-[11px] uppercase tracking-[0.15em] text-gray-500 dark:text-gray-500">
              {{ t('auth.passwordLabel') }}
            </label>
            <button
              type="button"
              @click="showPassword = !showPassword"
              class="font-mono text-[11px] uppercase tracking-[0.1em] text-gray-500 transition-colors hover:text-gray-900 dark:text-gray-500 dark:hover:text-white"
            >
              {{ showPassword ? t('common.hide') : t('common.show') }}
            </button>
          </div>
          <input
            id="password"
            v-model="formData.password"
            :type="showPassword ? 'text' : 'password'"
            required
            autocomplete="current-password"
            :disabled="isLoading"
            :class="inputClass(!!errors.password)"
            :placeholder="t('auth.passwordPlaceholder')"
          />
          <div class="mt-1.5 flex items-baseline justify-between gap-3">
            <p v-if="errors.password" class="text-xs text-red-600 dark:text-red-400">
              {{ errors.password }}
            </p>
            <span v-else></span>
            <router-link
              v-if="passwordResetEnabled && !backendModeEnabled"
              to="/forgot-password"
              class="text-xs text-gray-600 underline-offset-4 hover:text-gray-900 hover:underline dark:text-gray-400 dark:hover:text-white"
            >
              {{ t('auth.forgotPassword') }}
            </router-link>
          </div>
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
          <span>{{ isLoading ? t('auth.signingIn') : t('auth.signIn') }}</span>
          <Icon v-if="!isLoading" name="arrowRight" size="sm" />
        </button>
      </form>
    </div>

    <!-- Footer -->
    <template v-if="!backendModeEnabled" #footer>
      <p>
        {{ t('auth.dontHaveAccount') }}
        <router-link
          to="/register"
          class="font-medium text-gray-900 underline underline-offset-4 hover:text-gray-700 dark:text-white dark:hover:text-gray-300"
        >
          {{ t('auth.signUp') }}
        </router-link>
      </p>
    </template>
  </AuthLayout>

  <!-- 2FA Modal -->
  <TotpLoginModal
    v-if="show2FAModal"
    ref="totpModalRef"
    :temp-token="totpTempToken"
    :user-email-masked="totpUserEmailMasked"
    @verify="handle2FAVerify"
    @cancel="handle2FACancel"
  />
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import LinuxDoOAuthSection from '@/components/auth/LinuxDoOAuthSection.vue'
import TotpLoginModal from '@/components/auth/TotpLoginModal.vue'
import Icon from '@/components/icons/Icon.vue'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import { useAuthStore, useAppStore } from '@/stores'
import { getPublicSettings, isTotp2FARequired } from '@/api/auth'
import { inputClass } from '@/components/auth/authStyles'
import type { TotpLoginResponse } from '@/types'

const { t } = useI18n()

const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

const isLoading = ref<boolean>(false)
const errorMessage = ref<string>('')
const showPassword = ref<boolean>(false)

const turnstileEnabled = ref<boolean>(false)
const turnstileSiteKey = ref<string>('')
const linuxdoOAuthEnabled = ref<boolean>(false)
const backendModeEnabled = ref<boolean>(false)
const passwordResetEnabled = ref<boolean>(false)

const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref<string>('')

const show2FAModal = ref<boolean>(false)
const totpTempToken = ref<string>('')
const totpUserEmailMasked = ref<string>('')
const totpModalRef = ref<InstanceType<typeof TotpLoginModal> | null>(null)

const formData = reactive({
  email: '',
  password: ''
})

const errors = reactive({
  email: '',
  password: '',
  turnstile: ''
})

onMounted(async () => {
  const expiredFlag = sessionStorage.getItem('auth_expired')
  if (expiredFlag) {
    sessionStorage.removeItem('auth_expired')
    const message = t('auth.reloginRequired')
    errorMessage.value = message
    appStore.showWarning(message)
  }

  try {
    const settings = await getPublicSettings()
    turnstileEnabled.value = settings.turnstile_enabled
    turnstileSiteKey.value = settings.turnstile_site_key || ''
    linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled
    backendModeEnabled.value = settings.backend_mode_enabled
    passwordResetEnabled.value = settings.password_reset_enabled
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
  errors.password = ''
  errors.turnstile = ''

  let isValid = true

  if (!formData.email.trim()) {
    errors.email = t('auth.emailRequired')
    isValid = false
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
    errors.email = t('auth.invalidEmail')
    isValid = false
  }

  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
    isValid = false
  }

  if (turnstileEnabled.value && !turnstileToken.value) {
    errors.turnstile = t('auth.completeVerification')
    isValid = false
  }

  return isValid
}

async function handleLogin(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    const response = await authStore.login({
      email: formData.email,
      password: formData.password,
      turnstile_token: turnstileEnabled.value ? turnstileToken.value : undefined
    })

    if (isTotp2FARequired(response)) {
      const totpResponse = response as TotpLoginResponse
      totpTempToken.value = totpResponse.temp_token || ''
      totpUserEmailMasked.value = totpResponse.user_email_masked || ''
      show2FAModal.value = true
      isLoading.value = false
      return
    }

    appStore.showSuccess(t('auth.loginSuccess'))

    const redirectTo = (router.currentRoute.value.query.redirect as string) || '/dashboard'
    await router.push(redirectTo)
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
      errorMessage.value = t('auth.loginFailed')
    }

    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}

async function handle2FAVerify(code: string): Promise<void> {
  if (totpModalRef.value) {
    totpModalRef.value.setVerifying(true)
  }

  try {
    await authStore.login2FA(totpTempToken.value, code)

    show2FAModal.value = false
    appStore.showSuccess(t('auth.loginSuccess'))

    const redirectTo = (router.currentRoute.value.query.redirect as string) || '/dashboard'
    await router.push(redirectTo)
  } catch (error: unknown) {
    const err = error as { message?: string; response?: { data?: { message?: string } } }
    const message = err.response?.data?.message || err.message || t('profile.totp.loginFailed')

    if (totpModalRef.value) {
      totpModalRef.value.setError(message)
      totpModalRef.value.setVerifying(false)
    }
  }
}

function handle2FACancel(): void {
  show2FAModal.value = false
  totpTempToken.value = ''
  totpUserEmailMasked.value = ''
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
