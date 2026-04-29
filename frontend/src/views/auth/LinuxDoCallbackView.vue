<template>
  <AuthLayout>
    <div class="font-mono text-xs text-gray-500 dark:text-gray-500">auth / oauth / linux.do</div>
    <h1 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl">
      {{ t('auth.linuxdo.callbackTitle') }}
    </h1>
    <p class="mt-3 flex items-center gap-2 text-base text-gray-600 dark:text-gray-400">
      <span
        v-if="isProcessing"
        class="inline-block h-3 w-3 animate-spin rounded-full border-[1.5px] border-current border-t-transparent"
      ></span>
      <span>{{ isProcessing ? t('auth.linuxdo.callbackProcessing') : t('auth.linuxdo.callbackHint') }}</span>
    </p>

    <div class="mt-10">
      <!-- Invitation Code Required -->
      <transition name="auth-fade">
        <div v-if="needsInvitation" class="space-y-6">
          <p class="text-sm text-gray-700 dark:text-gray-300">
            {{ t('auth.linuxdo.invitationRequired') }}
          </p>

          <div>
            <label for="invitation_code" class="mb-2 block font-mono text-[11px] uppercase tracking-[0.15em] text-gray-500 dark:text-gray-500">
              {{ t('auth.invitationCodeLabel') }}
            </label>
            <input
              id="invitation_code"
              v-model="invitationCode"
              type="text"
              :class="inputClass(!!invitationError)"
              :placeholder="t('auth.invitationCodePlaceholder')"
              :disabled="isSubmitting"
              @keyup.enter="handleSubmitInvitation"
            />
            <transition name="auth-fade">
              <p v-if="invitationError" class="mt-1.5 text-xs text-red-600 dark:text-red-400">
                {{ invitationError }}
              </p>
            </transition>
          </div>

          <button
            class="inline-flex w-full items-center justify-center gap-1.5 rounded-md bg-gray-900 px-4 py-2.5 text-sm font-medium text-white transition-shadow hover:shadow-[0_4px_20px_-4px_rgba(59,130,246,0.5)] disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:shadow-none dark:bg-white dark:text-gray-900 dark:hover:shadow-[0_4px_20px_-4px_rgba(139,92,246,0.4)]"
            :disabled="isSubmitting || !invitationCode.trim()"
            @click="handleSubmitInvitation"
          >
            <span
              v-if="isSubmitting"
              class="inline-block h-3.5 w-3.5 animate-spin rounded-full border-[1.5px] border-current border-t-transparent"
            ></span>
            <span>{{ isSubmitting ? t('auth.linuxdo.completing') : t('auth.linuxdo.completeRegistration') }}</span>
            <Icon v-if="!isSubmitting" name="arrowRight" size="sm" />
          </button>
        </div>
      </transition>

      <!-- Error -->
      <transition name="auth-fade">
        <div v-if="errorMessage" class="space-y-4">
          <p class="border-l-2 border-red-500 pl-3 text-sm text-red-600 dark:text-red-400">
            {{ errorMessage }}
          </p>
          <router-link
            to="/login"
            class="inline-flex items-center gap-1.5 text-sm text-gray-900 underline-offset-4 hover:underline dark:text-white"
          >
            <Icon name="arrowLeft" size="sm" />
            {{ t('auth.linuxdo.backToLogin') }}
          </router-link>
        </div>
      </transition>
    </div>
  </AuthLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import Icon from '@/components/icons/Icon.vue'
import { useAuthStore, useAppStore } from '@/stores'
import { completeLinuxDoOAuthRegistration } from '@/api/auth'
import { inputClass } from '@/components/auth/authStyles'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

const isProcessing = ref(true)
const errorMessage = ref('')

const needsInvitation = ref(false)
const pendingOAuthToken = ref('')
const invitationCode = ref('')
const isSubmitting = ref(false)
const invitationError = ref('')
const redirectTo = ref('/dashboard')

function parseFragmentParams(): URLSearchParams {
  const raw = typeof window !== 'undefined' ? window.location.hash : ''
  const hash = raw.startsWith('#') ? raw.slice(1) : raw
  return new URLSearchParams(hash)
}

function sanitizeRedirectPath(path: string | null | undefined): string {
  if (!path) return '/dashboard'
  if (!path.startsWith('/')) return '/dashboard'
  if (path.startsWith('//')) return '/dashboard'
  if (path.includes('://')) return '/dashboard'
  if (path.includes('\n') || path.includes('\r')) return '/dashboard'
  return path
}

async function handleSubmitInvitation() {
  invitationError.value = ''
  if (!invitationCode.value.trim()) return

  isSubmitting.value = true
  try {
    const tokenData = await completeLinuxDoOAuthRegistration(
      pendingOAuthToken.value,
      invitationCode.value.trim()
    )
    if (tokenData.refresh_token) {
      localStorage.setItem('refresh_token', tokenData.refresh_token)
    }
    if (tokenData.expires_in) {
      localStorage.setItem('token_expires_at', String(Date.now() + tokenData.expires_in * 1000))
    }
    await authStore.setToken(tokenData.access_token)
    appStore.showSuccess(t('auth.loginSuccess'))
    await router.replace(redirectTo.value)
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { message?: string } } }
    invitationError.value =
      err.response?.data?.message || err.message || t('auth.linuxdo.completeRegistrationFailed')
  } finally {
    isSubmitting.value = false
  }
}

onMounted(async () => {
  const params = parseFragmentParams()

  const token = params.get('access_token') || ''
  const refreshToken = params.get('refresh_token') || ''
  const expiresInStr = params.get('expires_in') || ''
  const redirect = sanitizeRedirectPath(
    params.get('redirect') || (route.query.redirect as string | undefined) || '/dashboard'
  )
  const error = params.get('error')
  const errorDesc = params.get('error_description') || params.get('error_message') || ''

  if (error) {
    if (error === 'invitation_required') {
      pendingOAuthToken.value = params.get('pending_oauth_token') || ''
      redirectTo.value = sanitizeRedirectPath(params.get('redirect'))
      if (!pendingOAuthToken.value) {
        errorMessage.value = t('auth.linuxdo.invalidPendingToken')
        appStore.showError(errorMessage.value)
        isProcessing.value = false
        return
      }
      needsInvitation.value = true
      isProcessing.value = false
      return
    }
    errorMessage.value = errorDesc || error
    appStore.showError(errorMessage.value)
    isProcessing.value = false
    return
  }

  if (!token) {
    errorMessage.value = t('auth.linuxdo.callbackMissingToken')
    appStore.showError(errorMessage.value)
    isProcessing.value = false
    return
  }

  try {
    if (refreshToken) {
      localStorage.setItem('refresh_token', refreshToken)
    }
    if (expiresInStr) {
      const expiresIn = parseInt(expiresInStr, 10)
      if (!isNaN(expiresIn)) {
        localStorage.setItem('token_expires_at', String(Date.now() + expiresIn * 1000))
      }
    }

    await authStore.setToken(token)
    appStore.showSuccess(t('auth.loginSuccess'))
    await router.replace(redirect)
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { detail?: string } } }
    errorMessage.value = err.response?.data?.detail || err.message || t('auth.loginFailed')
    appStore.showError(errorMessage.value)
    isProcessing.value = false
  }
})
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
