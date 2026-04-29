<template>
  <AuthLayout>
    <div class="font-mono text-xs text-gray-500 dark:text-gray-500">auth / reset</div>
    <h1 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl">
      {{ t('auth.resetPasswordTitle') }}
    </h1>
    <p class="mt-3 text-base text-gray-600 dark:text-gray-400">
      {{ t('auth.resetPasswordHint') }}
    </p>

    <div class="mt-10">
      <!-- Invalid Link -->
      <div v-if="isInvalidLink">
        <p class="border-l-2 border-red-500 pl-3 text-sm text-red-600 dark:text-red-400">
          <span class="font-medium">{{ t('auth.invalidResetLink') }}</span>
          <br />
          <span class="text-gray-600 dark:text-gray-400">{{ t('auth.invalidResetLinkHint') }}</span>
        </p>
        <div class="mt-6">
          <router-link
            to="/forgot-password"
            class="inline-flex items-center gap-1.5 text-sm text-gray-900 underline-offset-4 hover:underline dark:text-white"
          >
            {{ t('auth.requestNewResetLink') }}
            <Icon name="arrowRight" size="sm" />
          </router-link>
        </div>
      </div>

      <!-- Success -->
      <div v-else-if="isSuccess">
        <div class="rounded-md border border-emerald-500/40 bg-emerald-50/40 px-4 py-4 dark:border-emerald-500/30 dark:bg-emerald-950/20">
          <p class="font-mono text-[11px] uppercase tracking-[0.15em] text-emerald-700 dark:text-emerald-500">
            ✓ {{ t('auth.passwordResetSuccess') }}
          </p>
          <p class="mt-2 text-sm text-gray-700 dark:text-gray-300">
            {{ t('auth.passwordResetSuccessHint') }}
          </p>
        </div>

        <div class="mt-6">
          <router-link
            to="/login"
            class="inline-flex w-full items-center justify-center gap-1.5 rounded-md bg-gray-900 px-4 py-2.5 text-sm font-medium text-white transition-shadow hover:shadow-[0_4px_20px_-4px_rgba(59,130,246,0.5)] dark:bg-white dark:text-gray-900 dark:hover:shadow-[0_4px_20px_-4px_rgba(139,92,246,0.4)]"
          >
            {{ t('auth.signIn') }}
            <Icon name="arrowRight" size="sm" />
          </router-link>
        </div>
      </div>

      <!-- Form -->
      <form v-else @submit.prevent="handleSubmit" class="space-y-6">
        <!-- Email (readonly) -->
        <div>
          <label for="email" class="mb-2 block font-mono text-[11px] uppercase tracking-[0.15em] text-gray-500 dark:text-gray-500">
            {{ t('auth.emailLabel') }}
          </label>
          <input
            id="email"
            :value="email"
            type="email"
            readonly
            disabled
            :class="inputClass(false)"
          />
        </div>

        <!-- New Password -->
        <div>
          <div class="mb-2 flex items-baseline justify-between">
            <label for="password" class="block font-mono text-[11px] uppercase tracking-[0.15em] text-gray-500 dark:text-gray-500">
              {{ t('auth.newPassword') }}
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
            :placeholder="t('auth.newPasswordPlaceholder')"
          />
          <p v-if="errors.password" class="mt-1.5 text-xs text-red-600 dark:text-red-400">
            {{ errors.password }}
          </p>
        </div>

        <!-- Confirm Password -->
        <div>
          <div class="mb-2 flex items-baseline justify-between">
            <label for="confirmPassword" class="block font-mono text-[11px] uppercase tracking-[0.15em] text-gray-500 dark:text-gray-500">
              {{ t('auth.confirmPassword') }}
            </label>
            <button
              type="button"
              @click="showConfirmPassword = !showConfirmPassword"
              class="font-mono text-[11px] uppercase tracking-[0.1em] text-gray-500 transition-colors hover:text-gray-900 dark:text-gray-500 dark:hover:text-white"
            >
              {{ showConfirmPassword ? t('common.hide') : t('common.show') }}
            </button>
          </div>
          <input
            id="confirmPassword"
            v-model="formData.confirmPassword"
            :type="showConfirmPassword ? 'text' : 'password'"
            required
            autocomplete="new-password"
            :disabled="isLoading"
            :class="inputClass(!!errors.confirmPassword)"
            :placeholder="t('auth.confirmPasswordPlaceholder')"
          />
          <p v-if="errors.confirmPassword" class="mt-1.5 text-xs text-red-600 dark:text-red-400">
            {{ errors.confirmPassword }}
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
          :disabled="isLoading"
          class="inline-flex w-full items-center justify-center gap-1.5 rounded-md bg-gray-900 px-4 py-2.5 text-sm font-medium text-white transition-shadow hover:shadow-[0_4px_20px_-4px_rgba(59,130,246,0.5)] disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:shadow-none dark:bg-white dark:text-gray-900 dark:hover:shadow-[0_4px_20px_-4px_rgba(139,92,246,0.4)]"
        >
          <span
            v-if="isLoading"
            class="inline-block h-3.5 w-3.5 animate-spin rounded-full border-[1.5px] border-current border-t-transparent"
          ></span>
          <span>{{ isLoading ? t('auth.resettingPassword') : t('auth.resetPassword') }}</span>
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
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { resetPassword } from '@/api/auth'
import { inputClass } from '@/components/auth/authStyles'

const { t } = useI18n()

const route = useRoute()
const appStore = useAppStore()

const isLoading = ref<boolean>(false)
const isSuccess = ref<boolean>(false)
const errorMessage = ref<string>('')
const showPassword = ref<boolean>(false)
const showConfirmPassword = ref<boolean>(false)

const email = ref<string>('')
const token = ref<string>('')

const formData = reactive({
  password: '',
  confirmPassword: ''
})

const errors = reactive({
  password: '',
  confirmPassword: ''
})

const isInvalidLink = computed(() => !email.value || !token.value)

onMounted(() => {
  email.value = (route.query.email as string) || ''
  token.value = (route.query.token as string) || ''
})

function validateForm(): boolean {
  errors.password = ''
  errors.confirmPassword = ''

  let isValid = true

  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
    isValid = false
  }

  if (!formData.confirmPassword) {
    errors.confirmPassword = t('auth.confirmPasswordRequired')
    isValid = false
  } else if (formData.password !== formData.confirmPassword) {
    errors.confirmPassword = t('auth.passwordsDoNotMatch')
    isValid = false
  }

  return isValid
}

async function handleSubmit(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) return

  isLoading.value = true

  try {
    await resetPassword({
      email: email.value,
      token: token.value,
      new_password: formData.password
    })

    isSuccess.value = true
    appStore.showSuccess(t('auth.passwordResetSuccess'))
  } catch (error: unknown) {
    const err = error as { message?: string; response?: { data?: { detail?: string; code?: string } } }

    if (err.response?.data?.code === 'INVALID_RESET_TOKEN') {
      errorMessage.value = t('auth.invalidOrExpiredToken')
    } else if (err.response?.data?.detail) {
      errorMessage.value = err.response.data.detail
    } else if (err.message) {
      errorMessage.value = err.message
    } else {
      errorMessage.value = t('auth.resetPasswordFailed')
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
