<template>
  <AuthLayout>
    <div class="font-mono text-xs text-gray-500 dark:text-gray-500">auth / register</div>
    <h1 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl">
      {{ t('auth.createAccount') }}
    </h1>
    <p class="mt-3 text-base text-gray-600 dark:text-gray-400">
      {{ t('auth.signUpToStart', { siteName }) }}
    </p>

    <div class="mt-10">
      <!-- LinuxDo OAuth -->
      <div v-if="linuxdoOAuthEnabled" class="mb-8">
        <LinuxDoOAuthSection :disabled="isLoading" />
      </div>

      <!-- Registration Disabled -->
      <p
        v-if="!registrationEnabled && settingsLoaded"
        class="border-l-2 border-amber-500 pl-3 text-sm text-amber-700 dark:text-amber-400"
      >
        {{ t('auth.registrationDisabled') }}
      </p>

      <!-- Form -->
      <form v-else @submit.prevent="handleRegister" class="space-y-6">
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
            autocomplete="new-password"
            :disabled="isLoading"
            :class="inputClass(!!errors.password)"
            :placeholder="t('auth.createPasswordPlaceholder')"
          />
          <p v-if="errors.password" class="mt-1.5 text-xs text-red-600 dark:text-red-400">
            {{ errors.password }}
          </p>
          <p v-else class="mt-1.5 text-xs text-gray-500 dark:text-gray-500">
            {{ t('auth.passwordHint') }}
          </p>
        </div>

        <!-- Invitation Code -->
        <div v-if="invitationCodeEnabled">
          <div class="mb-2 flex items-baseline justify-between">
            <label for="invitation_code" class="block font-mono text-[11px] uppercase tracking-[0.15em] text-gray-500 dark:text-gray-500">
              {{ t('auth.invitationCodeLabel') }}
            </label>
            <span v-if="invitationValidating" class="font-mono text-[11px] text-gray-500 dark:text-gray-500">
              ...
            </span>
            <span v-else-if="invitationValidation.valid" class="font-mono text-[11px] uppercase tracking-[0.1em] text-emerald-600 dark:text-emerald-500">
              ✓ {{ t('auth.invitationCodeValid') }}
            </span>
          </div>
          <input
            id="invitation_code"
            v-model="formData.invitation_code"
            type="text"
            :disabled="isLoading"
            :class="inputClass(invitationValidation.invalid || !!errors.invitation_code, invitationValidation.valid)"
            :placeholder="t('auth.invitationCodePlaceholder')"
            @input="handleInvitationCodeInput"
          />
          <p v-if="invitationValidation.invalid" class="mt-1.5 text-xs text-red-600 dark:text-red-400">
            {{ invitationValidation.message }}
          </p>
          <p v-else-if="errors.invitation_code" class="mt-1.5 text-xs text-red-600 dark:text-red-400">
            {{ errors.invitation_code }}
          </p>
        </div>

        <!-- Promo Code -->
        <div v-if="promoCodeEnabled">
          <div class="mb-2 flex items-baseline justify-between">
            <label for="promo_code" class="block font-mono text-[11px] uppercase tracking-[0.15em] text-gray-500 dark:text-gray-500">
              {{ t('auth.promoCodeLabel') }}
              <span class="ml-1 normal-case tracking-normal text-gray-400 dark:text-gray-600">({{ t('common.optional') }})</span>
            </label>
            <span v-if="promoValidating" class="font-mono text-[11px] text-gray-500 dark:text-gray-500">
              ...
            </span>
            <span v-else-if="promoValidation.valid" class="font-mono text-[11px] uppercase tracking-[0.1em] text-emerald-600 dark:text-emerald-500">
              ✓ +¥{{ promoValidation.bonusAmount?.toFixed(2) }}
            </span>
          </div>
          <input
            id="promo_code"
            v-model="formData.promo_code"
            type="text"
            :disabled="isLoading"
            :class="inputClass(promoValidation.invalid, promoValidation.valid)"
            :placeholder="t('auth.promoCodePlaceholder')"
            @input="handlePromoCodeInput"
          />
          <p v-if="promoValidation.invalid" class="mt-1.5 text-xs text-red-600 dark:text-red-400">
            {{ promoValidation.message }}
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
          <span>{{
            isLoading
              ? t('auth.processing')
              : emailVerifyEnabled
                ? t('auth.continue')
                : t('auth.createAccount')
          }}</span>
          <Icon v-if="!isLoading" name="arrowRight" size="sm" />
        </button>
      </form>
    </div>

    <!-- Footer -->
    <template #footer>
      <p>
        {{ t('auth.alreadyHaveAccount') }}
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
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import LinuxDoOAuthSection from '@/components/auth/LinuxDoOAuthSection.vue'
import Icon from '@/components/icons/Icon.vue'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import { useAuthStore, useAppStore } from '@/stores'
import { getPublicSettings, validatePromoCode, validateInvitationCode } from '@/api/auth'
import { buildAuthErrorMessage } from '@/utils/authError'
import { inputClass } from '@/components/auth/authStyles'
import {
  isRegistrationEmailSuffixAllowed,
  normalizeRegistrationEmailSuffixWhitelist
} from '@/utils/registrationEmailPolicy'

const { t, locale } = useI18n()

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const appStore = useAppStore()

const isLoading = ref<boolean>(false)
const settingsLoaded = ref<boolean>(false)
const errorMessage = ref<string>('')
const showPassword = ref<boolean>(false)

const registrationEnabled = ref<boolean>(true)
const emailVerifyEnabled = ref<boolean>(false)
const promoCodeEnabled = ref<boolean>(true)
const invitationCodeEnabled = ref<boolean>(false)
const turnstileEnabled = ref<boolean>(false)
const turnstileSiteKey = ref<string>('')
const siteName = ref<string>('AIInterface')
const linuxdoOAuthEnabled = ref<boolean>(false)
const registrationEmailSuffixWhitelist = ref<string[]>([])

const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref<string>('')

const promoValidating = ref<boolean>(false)
const promoValidation = reactive({
  valid: false,
  invalid: false,
  bonusAmount: null as number | null,
  message: ''
})
let promoValidateTimeout: ReturnType<typeof setTimeout> | null = null

const invitationValidating = ref<boolean>(false)
const invitationValidation = reactive({
  valid: false,
  invalid: false,
  message: ''
})
let invitationValidateTimeout: ReturnType<typeof setTimeout> | null = null

const formData = reactive({
  email: '',
  password: '',
  promo_code: '',
  invitation_code: ''
})

const errors = reactive({
  email: '',
  password: '',
  turnstile: '',
  invitation_code: ''
})

onMounted(async () => {
  try {
    const settings = await getPublicSettings()
    registrationEnabled.value = settings.registration_enabled
    emailVerifyEnabled.value = settings.email_verify_enabled
    promoCodeEnabled.value = settings.promo_code_enabled
    invitationCodeEnabled.value = settings.invitation_code_enabled
    turnstileEnabled.value = settings.turnstile_enabled
    turnstileSiteKey.value = settings.turnstile_site_key || ''
    siteName.value = settings.site_name || 'AIInterface'
    linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled
    registrationEmailSuffixWhitelist.value = normalizeRegistrationEmailSuffixWhitelist(
      settings.registration_email_suffix_whitelist || []
    )

    if (promoCodeEnabled.value) {
      const promoParam = route.query.promo as string
      if (promoParam) {
        formData.promo_code = promoParam
        await validatePromoCodeDebounced(promoParam)
      }
    }
  } catch (error) {
    console.error('Failed to load public settings:', error)
  } finally {
    settingsLoaded.value = true
  }
})

onUnmounted(() => {
  if (promoValidateTimeout) clearTimeout(promoValidateTimeout)
  if (invitationValidateTimeout) clearTimeout(invitationValidateTimeout)
})

function handlePromoCodeInput(): void {
  const code = formData.promo_code.trim()

  promoValidation.valid = false
  promoValidation.invalid = false
  promoValidation.bonusAmount = null
  promoValidation.message = ''

  if (!code) {
    promoValidating.value = false
    return
  }

  if (promoValidateTimeout) clearTimeout(promoValidateTimeout)
  promoValidateTimeout = setTimeout(() => {
    validatePromoCodeDebounced(code)
  }, 500)
}

async function validatePromoCodeDebounced(code: string): Promise<void> {
  if (!code.trim()) return

  promoValidating.value = true

  try {
    const result = await validatePromoCode(code)

    if (result.valid) {
      promoValidation.valid = true
      promoValidation.invalid = false
      promoValidation.bonusAmount = result.bonus_amount || 0
      promoValidation.message = ''
    } else {
      promoValidation.valid = false
      promoValidation.invalid = true
      promoValidation.bonusAmount = null
      promoValidation.message = getPromoErrorMessage(result.error_code)
    }
  } catch (error) {
    console.error('Failed to validate promo code:', error)
    promoValidation.valid = false
    promoValidation.invalid = true
    promoValidation.message = t('auth.promoCodeInvalid')
  } finally {
    promoValidating.value = false
  }
}

function getPromoErrorMessage(errorCode?: string): string {
  switch (errorCode) {
    case 'PROMO_CODE_NOT_FOUND':
      return t('auth.promoCodeNotFound')
    case 'PROMO_CODE_EXPIRED':
      return t('auth.promoCodeExpired')
    case 'PROMO_CODE_DISABLED':
      return t('auth.promoCodeDisabled')
    case 'PROMO_CODE_MAX_USED':
      return t('auth.promoCodeMaxUsed')
    case 'PROMO_CODE_ALREADY_USED':
      return t('auth.promoCodeAlreadyUsed')
    default:
      return t('auth.promoCodeInvalid')
  }
}

function handleInvitationCodeInput(): void {
  const code = formData.invitation_code.trim()

  invitationValidation.valid = false
  invitationValidation.invalid = false
  invitationValidation.message = ''
  errors.invitation_code = ''

  if (!code) return

  if (invitationValidateTimeout) clearTimeout(invitationValidateTimeout)
  invitationValidateTimeout = setTimeout(() => {
    validateInvitationCodeDebounced(code)
  }, 500)
}

async function validateInvitationCodeDebounced(code: string): Promise<void> {
  invitationValidating.value = true

  try {
    const result = await validateInvitationCode(code)

    if (result.valid) {
      invitationValidation.valid = true
      invitationValidation.invalid = false
      invitationValidation.message = ''
    } else {
      invitationValidation.valid = false
      invitationValidation.invalid = true
      invitationValidation.message = getInvitationErrorMessage(result.error_code)
    }
  } catch {
    invitationValidation.valid = false
    invitationValidation.invalid = true
    invitationValidation.message = t('auth.invitationCodeInvalid')
  } finally {
    invitationValidating.value = false
  }
}

function getInvitationErrorMessage(errorCode?: string): string {
  switch (errorCode) {
    case 'INVITATION_CODE_NOT_FOUND':
    case 'INVITATION_CODE_INVALID':
    case 'INVITATION_CODE_USED':
    case 'INVITATION_CODE_DISABLED':
    default:
      return t('auth.invitationCodeInvalid')
  }
}

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

function validateEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
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

function validateForm(): boolean {
  errors.email = ''
  errors.password = ''
  errors.turnstile = ''
  errors.invitation_code = ''

  let isValid = true

  if (!formData.email.trim()) {
    errors.email = t('auth.emailRequired')
    isValid = false
  } else if (!validateEmail(formData.email)) {
    errors.email = t('auth.invalidEmail')
    isValid = false
  } else if (
    !isRegistrationEmailSuffixAllowed(formData.email, registrationEmailSuffixWhitelist.value)
  ) {
    errors.email = buildEmailSuffixNotAllowedMessage()
    isValid = false
  }

  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
    isValid = false
  }

  if (invitationCodeEnabled.value && !formData.invitation_code.trim()) {
    errors.invitation_code = t('auth.invitationCodeRequired')
    isValid = false
  }

  if (turnstileEnabled.value && !turnstileToken.value) {
    errors.turnstile = t('auth.completeVerification')
    isValid = false
  }

  return isValid
}

async function handleRegister(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) return

  if (formData.promo_code.trim()) {
    if (promoValidating.value) {
      errorMessage.value = t('auth.promoCodeValidating')
      return
    }
    if (promoValidation.invalid) {
      errorMessage.value = t('auth.promoCodeInvalidCannotRegister')
      return
    }
  }

  if (invitationCodeEnabled.value) {
    if (invitationValidating.value) {
      errorMessage.value = t('auth.invitationCodeValidating')
      return
    }
    if (invitationValidation.invalid) {
      errorMessage.value = t('auth.invitationCodeInvalidCannotRegister')
      return
    }
    if (formData.invitation_code.trim() && !invitationValidation.valid) {
      errorMessage.value = t('auth.invitationCodeValidating')
      await validateInvitationCodeDebounced(formData.invitation_code.trim())
      if (!invitationValidation.valid) {
        errorMessage.value = t('auth.invitationCodeInvalidCannotRegister')
        return
      }
    }
  }

  isLoading.value = true

  try {
    if (emailVerifyEnabled.value) {
      sessionStorage.setItem(
        'register_data',
        JSON.stringify({
          email: formData.email,
          password: formData.password,
          turnstile_token: turnstileToken.value,
          promo_code: formData.promo_code || undefined,
          invitation_code: formData.invitation_code || undefined
        })
      )

      await router.push('/email-verify')
      return
    }

    await authStore.register({
      email: formData.email,
      password: formData.password,
      turnstile_token: turnstileEnabled.value ? turnstileToken.value : undefined,
      promo_code: formData.promo_code || undefined,
      invitation_code: formData.invitation_code || undefined
    })

    appStore.showSuccess(t('auth.accountCreatedSuccess', { siteName: siteName.value }))
    await router.push('/dashboard')
  } catch (error: unknown) {
    if (turnstileRef.value) {
      turnstileRef.value.reset()
      turnstileToken.value = ''
    }

    errorMessage.value = buildAuthErrorMessage(error, {
      fallback: t('auth.registrationFailed')
    })
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
