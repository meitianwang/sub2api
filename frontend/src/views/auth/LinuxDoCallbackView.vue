<template>
  <AuthLayout>
    <div class="space-y-6">
      <div class="text-center">
        <h2 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">
          {{ t('auth.linuxdo.callbackTitle') }}
        </h2>
        <p class="mt-1.5 text-sm text-gray-500 dark:text-dark-400">
          {{ isProcessing ? t('auth.linuxdo.callbackProcessing') : t('auth.linuxdo.callbackHint') }}
        </p>
      </div>

      <!-- Spinner during processing -->
      <div v-if="isProcessing" class="flex justify-center py-4">
        <div class="relative">
          <div class="h-10 w-10 animate-spin rounded-full border-2 border-gray-200 border-t-blue-600 dark:border-dark-700 dark:border-t-blue-400"></div>
        </div>
      </div>

      <!-- Invitation Required -->
      <transition name="fade">
        <div v-if="needsInvitation" class="space-y-5">
          <p class="text-sm text-gray-700 dark:text-gray-300">
            {{ t('auth.linuxdo.invitationRequired') }}
          </p>

          <div>
            <label for="invitation_code" class="input-label">{{ t('auth.invitationCodeLabel') }}</label>
            <div class="relative">
              <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
                <Icon name="key" size="md" class="text-gray-400 dark:text-dark-500" />
              </div>
              <input
                id="invitation_code"
                v-model="invitationCode"
                type="text"
                class="input pl-11"
                :class="{ 'input-error': !!invitationError }"
                :placeholder="t('auth.invitationCodePlaceholder')"
                :disabled="isSubmitting"
                @keyup.enter="handleSubmitInvitation"
              />
            </div>
            <transition name="fade">
              <p v-if="invitationError" class="input-error-text">{{ invitationError }}</p>
            </transition>
          </div>

          <button
            class="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-blue-600 via-blue-600 to-violet-600 px-4 py-2.5 text-sm font-semibold text-white shadow-lg shadow-blue-500/30 transition-all hover:shadow-xl hover:shadow-violet-500/40 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:ring-offset-2 active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:shadow-blue-500/30 dark:focus:ring-offset-gray-900"
            :disabled="isSubmitting || !invitationCode.trim()"
            @click="handleSubmitInvitation"
          >
            <svg
              v-if="isSubmitting"
              class="-ml-1 mr-1 h-4 w-4 animate-spin text-white"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <Icon v-else name="checkCircle" size="md" />
            {{ isSubmitting ? t('auth.linuxdo.completing') : t('auth.linuxdo.completeRegistration') }}
          </button>
        </div>
      </transition>

      <!-- Error -->
      <transition name="fade">
        <div v-if="errorMessage" class="space-y-4">
          <div class="rounded-xl border border-red-200 bg-red-50 p-3.5 dark:border-red-800/50 dark:bg-red-900/20">
            <div class="flex items-start gap-3">
              <Icon name="exclamationCircle" size="md" class="flex-shrink-0 text-red-500" />
              <p class="text-sm text-red-700 dark:text-red-400">{{ errorMessage }}</p>
            </div>
          </div>
          <div class="text-center">
            <router-link
              to="/login"
              class="inline-flex items-center gap-2 font-semibold text-blue-600 transition-colors hover:text-blue-500 dark:text-blue-400 dark:hover:text-blue-300"
            >
              <Icon name="arrowLeft" size="sm" />
              {{ t('auth.linuxdo.backToLogin') }}
            </router-link>
          </div>
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
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
