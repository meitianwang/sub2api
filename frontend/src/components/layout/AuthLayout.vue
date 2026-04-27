<template>
  <div class="auth-shell">
    <!-- Left panel: decorative -->
    <div class="auth-panel-left" aria-hidden="true">
      <!-- Geometric background -->
      <div class="auth-panel-bg">
        <!-- Dot grid -->
        <div class="auth-dot-grid"></div>
        <!-- Floating orbs -->
        <div class="auth-orb auth-orb-1"></div>
        <div class="auth-orb auth-orb-2"></div>
        <!-- Corner accent lines -->
        <div class="auth-corner-line auth-corner-tl"></div>
        <div class="auth-corner-line auth-corner-br"></div>
      </div>
      <!-- Brand mark -->
      <div class="auth-brand-mark" v-if="settingsLoaded">
        <div class="auth-brand-row">
          <img src="/logo.png" alt="" class="auth-brand-icon" />
          <span class="auth-brand-name">AIGateway</span>
        </div>
        <span class="auth-brand-tagline">{{ siteSubtitle }}</span>
      </div>
      <!-- Bottom quote -->
      <div class="auth-panel-footer">
        <p class="auth-panel-quote">Precision infrastructure<br/>for the AI era.</p>
      </div>
    </div>

    <!-- Right panel: form -->
    <div class="auth-panel-right">
      <div class="auth-form-container">
        <!-- Mobile logo (hidden on desktop) -->
        <div class="auth-mobile-logo" v-if="settingsLoaded">
          <img src="/logo.png" alt="" class="h-8 w-8" />
          <span class="text-base font-bold tracking-tight text-gray-900 dark:text-white">AIGateway</span>
        </div>

        <!-- Form card -->
        <div class="auth-form-card">
          <slot />
        </div>

        <!-- Footer links -->
        <div class="auth-form-footer">
          <slot name="footer" />
        </div>

        <!-- Copyright -->
        <p class="auth-copyright">
          &copy; {{ currentYear }} {{ siteName }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'AIInterface')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Subscription to API Conversion Platform')
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
/* ── Shell ── */
.auth-shell {
  display: flex;
  min-height: 100vh;
  background: #f7f7f9;
}
.dark .auth-shell {
  background: #0c0c10;
}

/* ── Left decorative panel ── */
.auth-panel-left {
  display: none;
  position: relative;
  overflow: hidden;
  background: #0c0c10;
}
@media (min-width: 1024px) {
  .auth-panel-left {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    width: 44%;
    flex-shrink: 0;
    padding: 3rem;
  }
}

.auth-panel-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

/* Dot grid pattern */
.auth-dot-grid {
  position: absolute;
  inset: 0;
  background-image: radial-gradient(circle, rgba(139, 92, 246, 0.25) 1px, transparent 1px);
  background-size: 32px 32px;
  mask-image: radial-gradient(ellipse at center, black 30%, transparent 80%);
  -webkit-mask-image: radial-gradient(ellipse at center, black 30%, transparent 80%);
}

/* Floating orbs */
.auth-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
}
.auth-orb-1 {
  width: 360px;
  height: 360px;
  top: -80px;
  right: -80px;
  background: rgba(109, 40, 217, 0.25);
}
.auth-orb-2 {
  width: 280px;
  height: 280px;
  bottom: -60px;
  left: -40px;
  background: rgba(139, 92, 246, 0.15);
}

/* Corner accent lines */
.auth-corner-line {
  position: absolute;
  width: 120px;
  height: 120px;
}
.auth-corner-tl {
  top: 2rem;
  left: 2rem;
  border-top: 1px solid rgba(139, 92, 246, 0.3);
  border-left: 1px solid rgba(139, 92, 246, 0.3);
}
.auth-corner-br {
  bottom: 2rem;
  right: 2rem;
  border-bottom: 1px solid rgba(139, 92, 246, 0.3);
  border-right: 1px solid rgba(139, 92, 246, 0.3);
}

/* Brand mark */
.auth-brand-mark {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.5rem;
}

.auth-brand-row {
  display: flex;
  align-items: center;
  gap: 0.625rem;
}

.auth-brand-icon {
  width: 40px;
  height: 40px;
  display: block;
  object-fit: contain;
}

.auth-brand-name {
  font-size: 1.25rem;
  font-weight: 700;
  color: #fafafa;
  letter-spacing: -0.01em;
}

.auth-brand-tagline {
  font-size: 0.75rem;
  color: #6b6b76;
  font-weight: 400;
}

/* Bottom quote */
.auth-panel-footer {
  position: relative;
  z-index: 1;
}
.auth-panel-quote {
  font-size: 1.25rem;
  font-weight: 600;
  color: #fafafa;
  line-height: 1.5;
  letter-spacing: -0.02em;
}

/* ── Right form panel ── */
.auth-panel-right {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem 1.5rem;
}

.auth-form-container {
  width: 100%;
  max-width: 400px;
}

/* Mobile logo */
.auth-mobile-logo {
  display: flex;
  align-items: center;
  margin-bottom: 2rem;
}
@media (min-width: 1024px) {
  .auth-mobile-logo {
    display: none;
  }
}

/* Form card */
.auth-form-card {
  background: white;
  border: 1px solid #e4e4e7;
  border-radius: 12px;
  padding: 2rem;
}
.dark .auth-form-card {
  background: #1c1c22;
  border-color: #2a2a35;
}

/* Footer links */
.auth-form-footer {
  margin-top: 1.25rem;
  text-align: center;
  font-size: 0.875rem;
  color: #71717a;
}
.dark .auth-form-footer {
  color: #6b6b76;
}

/* Copyright */
.auth-copyright {
  margin-top: 1.5rem;
  text-align: center;
  font-size: 0.75rem;
  color: #a1a1aa;
}
.dark .auth-copyright {
  color: #3d3d48;
}
</style>
