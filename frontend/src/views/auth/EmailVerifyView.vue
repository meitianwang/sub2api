<template>
  <AuthLayout>
    <div class="font-mono text-xs text-gray-500 dark:text-gray-500">auth / verify</div>
    <h1 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl">
      {{ t('auth.verifyYourEmail') }}
    </h1>
    <p class="mt-3 text-base text-gray-600 dark:text-gray-400">
      {{ t('auth.sendCodeDesc') }}
      <span class="font-mono text-gray-900 dark:text-white">{{ email }}</span>
    </p>

    <div class="mt-10">
      <!-- Session Expired -->
      <div v-if="!hasRegisterData">
        <p class="border-l-2 border-amber-500 pl-3 text-sm text-amber-700 dark:text-amber-400">
          <span class="font-medium">{{ t('auth.sessionExpired') }}</span>
          <br />
          <span class="text-gray-600 dark:text-gray-400">{{ t('auth.sessionExpiredDesc') }}</span>
        </p>
        <div class="mt-6">
          <router-link
            to="/register"
            class="inline-flex items-center gap-1.5 text-sm text-gray-900 underline-offset-4 hover:underline dark:text-white"
          >
            <Icon name="arrowLeft" size="sm" />
            {{ t('auth.backToRegistration') }}
          </router-link>
        </div>
      </div>

      <!-- Verify Form -->
      <form v-else @submit.prevent="handleVerify" class="space-y-6">
        <!-- Verification Code -->
        <div>
          <label for="code" class="mb-2 block font-mono text-[11px] uppercase tracking-[0.15em] text-gray-500 dark:text-gray-500">
            {{ t('auth.verificationCode') }}
          </label>
          <input
            id="code"
            v-model="verifyCode"
            type="text"
            required
            autocomplete="one-time-code"
            inputmode="numeric"
            maxlength="6"
            :disabled="isLoading"
            :class="codeInputClass"
            placeholder="000000"
          />
          <p v-if="errors.code" class="mt-1.5 text-xs text-red-600 dark:text-red-400">
            {{ errors.code }}
          </p>
          <p v-else class="mt-1.5 text-xs text-gray-500 dark:text-gray-500">{{ t('auth.verificationCodeHint') }}</p>
        </div>

        <!-- Code Sent Notice -->
        <p
          v-if="codeSent"
          class="border-l-2 border-emerald-500 pl-3 text-sm text-emerald-700 dark:text-emerald-500"
        >
          ✓ {{ t('auth.codeSentSuccess') }}
        </p>

        <!-- Turnstile for resend -->
        <div v-if="turnstileEnabled && turnstileSiteKey && showResendTurnstile">
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
          :disabled="isLoading || !verifyCode"
          class="inline-flex w-full items-center justify-center gap-1.5 rounded-md bg-gray-900 px-4 py-2.5 text-sm font-medium text-white transition-shadow hover:shadow-[0_4px_20px_-4px_rgba(59,130,246,0.5)] disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:shadow-none dark:bg-white dark:text-gray-900 dark:hover:shadow-[0_4px_20px_-4px_rgba(139,92,246,0.4)]"
        >
          <span
            v-if="isLoading"
            class="inline-block h-3.5 w-3.5 animate-spin rounded-full border-[1.5px] border-current border-t-transparent"
          ></span>
          <span>{{ isLoading ? t('auth.verifying') : t('auth.verifyAndCreate') }}</span>
          <Icon v-if="!isLoading" name="arrowRight" size="sm" />
        </button>

        <!-- Resend -->
        <div class="text-center">
          <button
            v-if="countdown > 0"
            type="button"
            disabled
            class="cursor-not-allowed font-mono text-xs text-gray-400 dark:text-gray-600"
          >
            {{ t('auth.resendCountdown', { countdown }) }}
          </button>
          <button
            v-else
            type="button"
            :disabled="isSendingCode || (turnstileEnabled && showResendTurnstile && !resendTurnstileToken)"
            class="text-xs text-gray-600 underline-offset-4 hover:text-gray-900 hover:underline disabled:cursor-not-allowed disabled:opacity-50 dark:text-gray-400 dark:hover:text-white"
            @click="handleResendCode"
          >
            <span v-if="isSendingCode">{{ t('auth.sendingCode') }}</span>
            <span v-else-if="turnstileEnabled && !showResendTurnstile">
              {{ t('auth.clickToResend') }}
            </span>
            <span v-else>{{ t('auth.resendCode') }}</span>
          </button>
        </div>
      </form>
    </div>

    <!-- Footer -->
    <template #footer>
      <button
        @click="handleBack"
        class="inline-flex items-center gap-1.5 text-gray-600 underline-offset-4 hover:text-gray-900 hover:underline dark:text-gray-400 dark:hover:text-white"
      >
        <Icon name="arrowLeft" size="sm" />
        {{ t('auth.backToRegistration') }}
      </button>
    </template>
  </AuthLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import Icon from '@/components/icons/Icon.vue'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import { useAuthStore, useAppStore } from '@/stores'
import { getPublicSettings, sendVerifyCode } from '@/api/auth'
import { buildAuthErrorMessage } from '@/utils/authError'
import { inputClass } from '@/components/auth/authStyles'
import {
  isRegistrationEmailSuffixAllowed,
  normalizeRegistrationEmailSuffixWhitelist
} from '@/utils/registrationEmailPolicy'

const { t, locale } = useI18n()

const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

const isLoading = ref<boolean>(false)
const isSendingCode = ref<boolean>(false)
const errorMessage = ref<string>('')
const codeSent = ref<boolean>(false)
const verifyCode = ref<string>('')
const countdown = ref<number>(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

const email = ref<string>('')
const password = ref<string>('')
const initialTurnstileToken = ref<string>('')
const promoCode = ref<string>('')
const invitationCode = ref<string>('')
const hasRegisterData = ref<boolean>(false)

const turnstileEnabled = ref<boolean>(false)
const turnstileSiteKey = ref<string>('')
const siteName = ref<string>('AIInterface')
const registrationEmailSuffixWhitelist = ref<string[]>([])

const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const resendTurnstileToken = ref<string>('')
const showResendTurnstile = ref<boolean>(false)

const errors = ref({
  code: '',
  turnstile: ''
})

const codeInputClass = computed(() => {
  return `${inputClass(!!errors.value.code)} text-center font-mono text-xl tracking-[0.5em] py-3`
})

onMounted(async () => {
  const registerDataStr = sessionStorage.getItem('register_data')
  if (registerDataStr) {
    try {
      const registerData = JSON.parse(registerDataStr)
      email.value = registerData.email || ''
      password.value = registerData.password || ''
      initialTurnstileToken.value = registerData.turnstile_token || ''
      promoCode.value = registerData.promo_code || ''
      invitationCode.value = registerData.invitation_code || ''
      hasRegisterData.value = !!(email.value && password.value)
    } catch {
      hasRegisterData.value = false
    }
  }

  try {
    const settings = await getPublicSettings()
    turnstileEnabled.value = settings.turnstile_enabled
    turnstileSiteKey.value = settings.turnstile_site_key || ''
    siteName.value = settings.site_name || 'AIInterface'
    registrationEmailSuffixWhitelist.value = normalizeRegistrationEmailSuffixWhitelist(
      settings.registration_email_suffix_whitelist || []
    )
  } catch (error) {
    console.error('Failed to load public settings:', error)
  }

  if (hasRegisterData.value) {
    await sendCode()
  }
})

onUnmounted(() => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
})

function startCountdown(seconds: number): void {
  countdown.value = seconds

  if (countdownTimer) clearInterval(countdownTimer)

  countdownTimer = setInterval(() => {
    if (countdown.value > 0) {
      countdown.value--
    } else if (countdownTimer) {
      clearInterval(countdownTimer)
      countdownTimer = null
    }
  }, 1000)
}

function onTurnstileVerify(token: string): void {
  resendTurnstileToken.value = token
  errors.value.turnstile = ''
}

function onTurnstileExpire(): void {
  resendTurnstileToken.value = ''
  errors.value.turnstile = t('auth.turnstileExpired')
}

function onTurnstileError(): void {
  resendTurnstileToken.value = ''
  errors.value.turnstile = t('auth.turnstileFailed')
}

async function sendCode(): Promise<void> {
  isSendingCode.value = true
  errorMessage.value = ''

  try {
    if (!isRegistrationEmailSuffixAllowed(email.value, registrationEmailSuffixWhitelist.value)) {
      errorMessage.value = buildEmailSuffixNotAllowedMessage()
      appStore.showError(errorMessage.value)
      return
    }

    const response = await sendVerifyCode({
      email: email.value,
      turnstile_token: resendTurnstileToken.value || initialTurnstileToken.value || undefined
    })

    codeSent.value = true
    startCountdown(response.countdown)

    initialTurnstileToken.value = ''
    showResendTurnstile.value = false
    resendTurnstileToken.value = ''
  } catch (error: unknown) {
    errorMessage.value = buildAuthErrorMessage(error, {
      fallback: t('auth.sendCodeFailed')
    })
    appStore.showError(errorMessage.value)
  } finally {
    isSendingCode.value = false
  }
}

async function handleResendCode(): Promise<void> {
  if (turnstileEnabled.value && !showResendTurnstile.value) {
    showResendTurnstile.value = true
    return
  }

  if (turnstileEnabled.value && !resendTurnstileToken.value) {
    errors.value.turnstile = t('auth.completeVerification')
    return
  }

  await sendCode()
}

function validateForm(): boolean {
  errors.value.code = ''

  if (!verifyCode.value.trim()) {
    errors.value.code = t('auth.codeRequired')
    return false
  }

  if (!/^\d{6}$/.test(verifyCode.value.trim())) {
    errors.value.code = t('auth.invalidCode')
    return false
  }

  return true
}

async function handleVerify(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) return

  isLoading.value = true

  try {
    if (!isRegistrationEmailSuffixAllowed(email.value, registrationEmailSuffixWhitelist.value)) {
      errorMessage.value = buildEmailSuffixNotAllowedMessage()
      appStore.showError(errorMessage.value)
      return
    }

    await authStore.register({
      email: email.value,
      password: password.value,
      verify_code: verifyCode.value.trim(),
      turnstile_token: initialTurnstileToken.value || undefined,
      promo_code: promoCode.value || undefined,
      invitation_code: invitationCode.value || undefined
    })

    sessionStorage.removeItem('register_data')

    appStore.showSuccess(t('auth.accountCreatedSuccess', { siteName: siteName.value }))
    await router.push('/dashboard')
  } catch (error: unknown) {
    errorMessage.value = buildAuthErrorMessage(error, {
      fallback: t('auth.verifyFailed')
    })
    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}

function handleBack(): void {
  sessionStorage.removeItem('register_data')
  router.push('/register')
}

function buildEmailSuffixNotAllowedMessage(): string {
  const normalizedWhitelist = normalizeRegistrationEmailSuffixWhitelist(
    registrationEmailSuffixWhitelist.value
  )
  if (normalizedWhitelist.length === 0) return t('auth.emailSuffixNotAllowed')
  const separator = String(locale.value || '').toLowerCase().startsWith('zh') ? '、' : ', '
  return t('auth.emailSuffixNotAllowedWithAllowed', {
    suffixes: normalizedWhitelist.join(separator)
  })
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
