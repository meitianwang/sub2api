<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page -->
  <div v-else class="home-shell">
    <!-- Background -->
    <div class="home-bg" aria-hidden="true">
      <div class="home-bg-dots"></div>
      <div class="home-bg-orb home-bg-orb-1"></div>
      <div class="home-bg-orb home-bg-orb-2"></div>
    </div>

    <!-- Header -->
    <header class="home-header">
      <nav class="home-nav">
        <!-- Logo -->
        <div class="home-nav-logo">
          <div class="home-logo-img">
            <img :src="siteLogo || '/logo.png'" alt="Logo" />
          </div>
          <span class="home-logo-name">{{ siteName }}</span>
        </div>

        <!-- Center Nav -->
        <div class="home-nav-links">
          <router-link to="/home" class="nav-tab nav-tab-active">{{ t('nav.home') }}</router-link>
          <router-link to="/models" class="nav-tab">{{ t('nav.models') }}</router-link>
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="nav-tab">{{ t('nav.docs') }}</a>
          <router-link v-else to="/docs" class="nav-tab">{{ t('nav.docs') }}</router-link>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="nav-tab">{{ t('nav.console') }}</router-link>
        </div>

        <!-- Right Actions -->
        <div class="home-nav-actions">
          <LocaleSwitcher />
          <button @click="toggleTheme" class="home-theme-btn" :aria-label="isDark ? 'Light mode' : 'Dark mode'">
            <Icon v-if="isDark" name="sun" size="sm" />
            <Icon v-else name="moon" size="sm" />
          </button>
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="home-avatar"
          >{{ userInitial }}</router-link>
          <router-link
            v-else
            to="/login"
            class="home-cta-btn"
          >{{ t('home.login') }}</router-link>
        </div>
      </nav>
    </header>

    <!-- Main Content -->
    <main class="home-main">
      <div class="home-container">

        <!-- Hero Section -->
        <div class="home-hero">
          <!-- Left: Text Content -->
          <div class="home-hero-text">
            <div class="home-hero-badge">
              <span class="home-badge-dot"></span>
              AI Gateway Infrastructure
            </div>
            <h1 class="home-hero-title">{{ siteName }}</h1>
            <p class="home-hero-subtitle">{{ siteSubtitle }}</p>

            <!-- CTA -->
            <div class="home-hero-cta">
              <router-link
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="btn btn-primary home-cta-primary"
              >
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                <Icon name="arrowRight" size="md" :stroke-width="2" />
              </router-link>
            </div>

            <!-- Feature tags -->
            <div class="home-hero-tags">
              <div class="home-tag">
                <Icon name="swap" size="sm" class="text-primary-500" />
                <span>{{ t('home.tags.subscriptionToApi') }}</span>
              </div>
              <div class="home-tag">
                <Icon name="shield" size="sm" class="text-primary-500" />
                <span>{{ t('home.tags.stickySession') }}</span>
              </div>
              <div class="home-tag">
                <Icon name="chart" size="sm" class="text-primary-500" />
                <span>{{ t('home.tags.realtimeBilling') }}</span>
              </div>
            </div>
          </div>

          <!-- Right: Terminal Animation -->
          <div class="home-hero-terminal">
            <div class="terminal-window">
              <div class="terminal-header">
                <div class="terminal-dots">
                  <span class="t-dot t-close"></span>
                  <span class="t-dot t-min"></span>
                  <span class="t-dot t-max"></span>
                </div>
                <span class="terminal-label">terminal</span>
              </div>
              <div class="terminal-body">
                <div class="t-line t-line-1">
                  <span class="t-prompt">$</span>
                  <span class="t-cmd">curl</span>
                  <span class="t-flag">-X POST</span>
                  <span class="t-url">/v1/messages</span>
                </div>
                <div class="t-line t-line-2">
                  <span class="t-comment"># Routing to upstream...</span>
                </div>
                <div class="t-line t-line-3">
                  <span class="t-ok">200 OK</span>
                  <span class="t-json">&#123; "content": "Hello!" &#125;</span>
                </div>
                <div class="t-line t-line-4">
                  <span class="t-prompt">$</span>
                  <span class="t-cursor"></span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Features Grid -->
        <section class="home-features">
          <div class="home-section-header">
            <h2 class="home-section-title">{{ t('home.providers.title') }}</h2>
            <p class="home-section-desc">{{ t('home.providers.description') }}</p>
          </div>

          <div class="home-features-grid">
            <!-- Feature 1 -->
            <div class="home-feature-card">
              <div class="home-feature-icon home-feature-icon-blue">
                <Icon name="server" size="md" class="text-white" />
              </div>
              <h3 class="home-feature-name">{{ t('home.features.unifiedGateway') }}</h3>
              <p class="home-feature-desc">{{ t('home.features.unifiedGatewayDesc') }}</p>
            </div>

            <!-- Feature 2 -->
            <div class="home-feature-card">
              <div class="home-feature-icon home-feature-icon-violet">
                <svg class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72m.94 3.198l.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 016 18.719m12 0a5.971 5.971 0 00-.941-3.197m0 0A5.995 5.995 0 0012 12.75a5.995 5.995 0 00-5.058 2.772m0 0a3 3 0 00-4.681 2.72 8.986 8.986 0 003.74.477m.94-3.197a5.971 5.971 0 00-.94 3.197M15 6.75a3 3 0 11-6 0 3 3 0 016 0zm6 3a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0zm-13.5 0a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z" />
                </svg>
              </div>
              <h3 class="home-feature-name">{{ t('home.features.multiAccount') }}</h3>
              <p class="home-feature-desc">{{ t('home.features.multiAccountDesc') }}</p>
            </div>

            <!-- Feature 3 -->
            <div class="home-feature-card">
              <div class="home-feature-icon home-feature-icon-emerald">
                <svg class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z" />
                </svg>
              </div>
              <h3 class="home-feature-name">{{ t('home.features.balanceQuota') }}</h3>
              <p class="home-feature-desc">{{ t('home.features.balanceQuotaDesc') }}</p>
            </div>
          </div>
        </section>

        <!-- Providers -->
        <section class="home-providers">
          <div class="home-providers-list">
            <div class="home-provider home-provider-active">
              <div class="home-provider-icon" style="background: linear-gradient(135deg, #f97316, #ea580c);">
                <span>C</span>
              </div>
              <span>{{ t('home.providers.claude') }}</span>
              <span class="home-provider-badge">{{ t('home.providers.supported') }}</span>
            </div>
            <div class="home-provider home-provider-active">
              <div class="home-provider-icon" style="background: linear-gradient(135deg, #22c55e, #16a34a);">
                <span>G</span>
              </div>
              <span>GPT</span>
              <span class="home-provider-badge">{{ t('home.providers.supported') }}</span>
            </div>
            <div class="home-provider home-provider-active">
              <div class="home-provider-icon" style="background: linear-gradient(135deg, #3b82f6, #2563eb);">
                <span>G</span>
              </div>
              <span>{{ t('home.providers.gemini') }}</span>
              <span class="home-provider-badge">{{ t('home.providers.supported') }}</span>
            </div>
            <div class="home-provider home-provider-soon">
              <div class="home-provider-icon" style="background: linear-gradient(135deg, #6b7280, #4b5563);">
                <span>+</span>
              </div>
              <span>{{ t('home.providers.more') }}</span>
              <span class="home-provider-badge-soon">{{ t('home.providers.soon') }}</span>
            </div>
          </div>
        </section>

      </div>
    </main>

    <!-- Footer -->
    <footer class="home-footer">
      <div class="home-footer-inner">
        <p class="home-footer-copy">&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
        <div class="home-footer-links">
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="home-footer-link">
            {{ t('home.docs') }}
          </a>
          <router-link v-else to="/docs" class="home-footer-link">{{ t('home.docs') }}</router-link>
          <a :href="githubUrl" target="_blank" rel="noopener noreferrer" class="home-footer-link">GitHub</a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'AIInterface')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const githubUrl = 'https://github.com/Wei-Shaw/sub2api'
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})
const currentYear = computed(() => new Date().getFullYear())

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
/* ── Shell ── */
.home-shell {
  position: relative;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: #f7f7f9;
  overflow: hidden;
}
.dark .home-shell {
  background: #0c0c10;
}

/* ── Background ── */
.home-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
}
.home-bg-dots {
  position: absolute;
  inset: 0;
  background-image: radial-gradient(circle, rgba(139, 92, 246, 0.12) 1px, transparent 1px);
  background-size: 40px 40px;
  mask-image: radial-gradient(ellipse 80% 60% at 50% 30%, black 20%, transparent 100%);
  -webkit-mask-image: radial-gradient(ellipse 80% 60% at 50% 30%, black 20%, transparent 100%);
}
.dark .home-bg-dots {
  background-image: radial-gradient(circle, rgba(139, 92, 246, 0.2) 1px, transparent 1px);
}
.home-bg-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(100px);
}
.home-bg-orb-1 {
  width: 500px;
  height: 500px;
  top: -150px;
  right: -100px;
  background: rgba(124, 58, 237, 0.12);
}
.home-bg-orb-2 {
  width: 400px;
  height: 400px;
  bottom: -100px;
  left: -80px;
  background: rgba(109, 40, 217, 0.08);
}
.dark .home-bg-orb-1 { background: rgba(124, 58, 237, 0.18); }
.dark .home-bg-orb-2 { background: rgba(109, 40, 217, 0.12); }

/* ── Header ── */
.home-header {
  position: relative;
  z-index: 20;
  border-bottom: 1px solid rgba(228, 228, 231, 0.8);
  background: rgba(247, 247, 249, 0.9);
  backdrop-filter: blur(12px);
  padding: 0 1.5rem;
}
.dark .home-header {
  border-bottom-color: rgba(42, 42, 53, 0.8);
  background: rgba(12, 12, 16, 0.85);
}

.home-nav {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
}

.home-nav-logo {
  display: flex;
  align-items: center;
  gap: 0.625rem;
}
.home-logo-img {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  overflow: hidden;
  flex-shrink: 0;
}
.home-logo-img img { width: 100%; height: 100%; object-fit: contain; }
.home-logo-name {
  font-size: 0.9375rem;
  font-weight: 700;
  color: #18181b;
  letter-spacing: -0.01em;
}
.dark .home-logo-name { color: #fafafa; }

.home-nav-links {
  display: none;
  align-items: center;
  gap: 0.125rem;
}
@media (min-width: 640px) {
  .home-nav-links { display: flex; }
}

.home-nav-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

/* Nav tabs */
.nav-tab {
  padding: 5px 14px;
  font-size: 13px;
  font-weight: 500;
  color: #71717a;
  border-radius: 6px;
  transition: all 0.15s;
  text-decoration: none;
}
.nav-tab:hover {
  color: #18181b;
  background: rgba(0,0,0,0.04);
}
.nav-tab-active {
  color: #7c3aed;
  background: rgba(124, 58, 237, 0.06);
}
.dark .nav-tab { color: #6b6b76; }
.dark .nav-tab:hover {
  color: #fafafa;
  background: rgba(255,255,255,0.05);
}
.dark .nav-tab-active {
  color: #a78bfa;
  background: rgba(139, 92, 246, 0.1);
}

/* Theme toggle */
.home-theme-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: #71717a;
  cursor: pointer;
  transition: all 0.15s;
}
.home-theme-btn:hover {
  background: rgba(0,0,0,0.05);
  color: #18181b;
}
.dark .home-theme-btn { color: #6b6b76; }
.dark .home-theme-btn:hover {
  background: rgba(255,255,255,0.05);
  color: #fafafa;
}

/* Auth avatar */
.home-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background: linear-gradient(135deg, #7c3aed, #6d28d9);
  color: white;
  font-size: 11px;
  font-weight: 700;
  text-decoration: none;
}

/* CTA button (header) */
.home-cta-btn {
  padding: 5px 14px;
  border-radius: 6px;
  background: #7c3aed;
  color: white;
  font-size: 12px;
  font-weight: 600;
  text-decoration: none;
  transition: background 0.15s;
}
.home-cta-btn:hover { background: #6d28d9; }

/* ── Main ── */
.home-main {
  position: relative;
  z-index: 10;
  flex: 1;
  padding: 5rem 1.5rem 3rem;
}

.home-container {
  max-width: 1200px;
  margin: 0 auto;
}

/* ── Hero ── */
.home-hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4rem;
  margin-bottom: 6rem;
}
@media (min-width: 1024px) {
  .home-hero {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }
}

.home-hero-text {
  flex: 1;
  text-align: center;
}
@media (min-width: 1024px) {
  .home-hero-text { text-align: left; }
}

/* Badge */
.home-hero-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 20px;
  background: rgba(124, 58, 237, 0.08);
  border: 1px solid rgba(124, 58, 237, 0.15);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.04em;
  color: #7c3aed;
  text-transform: uppercase;
  margin-bottom: 1.25rem;
}
.dark .home-hero-badge {
  background: rgba(139, 92, 246, 0.1);
  border-color: rgba(139, 92, 246, 0.2);
  color: #a78bfa;
}
.home-badge-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #7c3aed;
  animation: pulse 2s ease-in-out infinite;
}
.dark .home-badge-dot { background: #a78bfa; }

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(0.8); }
}

/* Title */
.home-hero-title {
  font-size: clamp(2.25rem, 5vw, 3.75rem);
  font-weight: 800;
  letter-spacing: -0.03em;
  line-height: 1.1;
  color: #09090b;
  margin-bottom: 1rem;
}
.dark .home-hero-title { color: #fafafa; }

/* Subtitle */
.home-hero-subtitle {
  font-size: 1.0625rem;
  color: #71717a;
  line-height: 1.7;
  margin-bottom: 2rem;
  max-width: 480px;
}
@media (min-width: 1024px) {
  .home-hero-subtitle { max-width: 440px; }
}
.dark .home-hero-subtitle { color: #6b6b76; }

/* CTA */
.home-hero-cta { margin-bottom: 2rem; }
.home-cta-primary {
  padding: 10px 24px !important;
  font-size: 14px !important;
  gap: 8px;
}

/* Hero tags */
.home-hero-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  justify-content: center;
}
@media (min-width: 1024px) {
  .home-hero-tags { justify-content: flex-start; }
}
.home-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 6px;
  border: 1px solid #e4e4e7;
  background: white;
  font-size: 12px;
  font-weight: 500;
  color: #52525b;
}
.dark .home-tag {
  border-color: #2a2a35;
  background: #1c1c22;
  color: #a1a1aa;
}

/* ── Terminal ── */
.home-hero-terminal {
  flex: 1;
  display: flex;
  justify-content: center;
}
@media (min-width: 1024px) {
  .home-hero-terminal { justify-content: flex-end; }
}

.terminal-window {
  width: 100%;
  max-width: 420px;
  background: #0c0c10;
  border-radius: 10px;
  border: 1px solid #2a2a35;
  box-shadow:
    0 32px 64px -16px rgba(0, 0, 0, 0.3),
    0 0 0 1px rgba(139, 92, 246, 0.1),
    0 0 48px rgba(124, 58, 237, 0.08);
  overflow: hidden;
  transform: perspective(1000px) rotateX(2deg) rotateY(-2deg);
  transition: transform 0.4s ease, box-shadow 0.4s ease;
}
.terminal-window:hover {
  transform: perspective(1000px) rotateX(0deg) rotateY(0deg) translateY(-6px);
  box-shadow:
    0 40px 80px -16px rgba(0, 0, 0, 0.4),
    0 0 0 1px rgba(139, 92, 246, 0.2),
    0 0 64px rgba(124, 58, 237, 0.12);
}

.terminal-header {
  display: flex;
  align-items: center;
  padding: 10px 14px;
  background: #141418;
  border-bottom: 1px solid #2a2a35;
}
.terminal-dots { display: flex; gap: 6px; }
.t-dot { width: 11px; height: 11px; border-radius: 50%; }
.t-close { background: #ef4444; }
.t-min { background: #eab308; }
.t-max { background: #22c55e; }
.terminal-label {
  flex: 1;
  text-align: center;
  font-size: 11px;
  font-family: 'JetBrains Mono', monospace;
  color: #3d3d48;
  margin-right: 50px;
  letter-spacing: 0.02em;
}

.terminal-body {
  padding: 18px 20px;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 13px;
  line-height: 2;
}
.t-line {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  opacity: 0;
  animation: t-appear 0.4s ease forwards;
}
.t-line-1 { animation-delay: 0.3s; }
.t-line-2 { animation-delay: 1.0s; }
.t-line-3 { animation-delay: 1.8s; }
.t-line-4 { animation-delay: 2.6s; }

@keyframes t-appear {
  from { opacity: 0; transform: translateX(-6px); }
  to   { opacity: 1; transform: translateX(0); }
}

.t-prompt  { color: #22c55e; font-weight: 700; }
.t-cmd     { color: #38bdf8; }
.t-flag    { color: #a78bfa; }
.t-url     { color: #8b5cf6; }
.t-comment { color: #3d3d48; font-style: italic; }
.t-ok {
  color: #22c55e;
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.2);
  padding: 1px 8px;
  border-radius: 4px;
  font-weight: 600;
  font-size: 12px;
}
.t-json { color: #fbbf24; }
.t-cursor {
  display: inline-block;
  width: 7px;
  height: 15px;
  background: #22c55e;
  border-radius: 1px;
  animation: t-blink 1s step-end infinite;
}
@keyframes t-blink {
  0%, 50%  { opacity: 1; }
  51%, 100% { opacity: 0; }
}

/* ── Features ── */
.home-features {
  margin-bottom: 4rem;
}
.home-section-header {
  text-align: center;
  margin-bottom: 2.5rem;
}
.home-section-title {
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: #09090b;
  margin-bottom: 0.5rem;
}
.dark .home-section-title { color: #fafafa; }
.home-section-desc {
  font-size: 14px;
  color: #71717a;
}
.dark .home-section-desc { color: #6b6b76; }

.home-features-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: 1fr;
}
@media (min-width: 768px) {
  .home-features-grid { grid-template-columns: repeat(3, 1fr); }
}

.home-feature-card {
  padding: 1.5rem;
  border-radius: 10px;
  border: 1px solid #e4e4e7;
  background: white;
  transition: all 0.2s;
}
.home-feature-card:hover {
  border-color: rgba(124, 58, 237, 0.3);
  box-shadow: 0 0 0 1px rgba(124, 58, 237, 0.1), 0 8px 24px rgba(0,0,0,0.06);
}
.dark .home-feature-card {
  border-color: #2a2a35;
  background: #141418;
}
.dark .home-feature-card:hover {
  border-color: rgba(139, 92, 246, 0.3);
  box-shadow: 0 0 0 1px rgba(139, 92, 246, 0.1), 0 8px 24px rgba(0,0,0,0.2);
}

.home-feature-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
}
.home-feature-icon-blue    { background: linear-gradient(135deg, #3b82f6, #2563eb); }
.home-feature-icon-violet  { background: linear-gradient(135deg, #8b5cf6, #7c3aed); }
.home-feature-icon-emerald { background: linear-gradient(135deg, #10b981, #059669); }

.home-feature-name {
  font-size: 15px;
  font-weight: 600;
  color: #09090b;
  margin-bottom: 0.5rem;
  letter-spacing: -0.01em;
}
.dark .home-feature-name { color: #fafafa; }
.home-feature-desc {
  font-size: 13px;
  color: #71717a;
  line-height: 1.6;
}
.dark .home-feature-desc { color: #6b6b76; }

/* ── Providers ── */
.home-providers { margin-bottom: 4rem; }
.home-providers-list {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.75rem;
}

.home-provider {
  display: inline-flex;
  align-items: center;
  gap: 0.625rem;
  padding: 8px 16px;
  border-radius: 8px;
  border: 1px solid #e4e4e7;
  background: white;
  font-size: 13px;
  font-weight: 500;
  color: #3f3f46;
  transition: all 0.15s;
}
.dark .home-provider {
  border-color: #2a2a35;
  background: #141418;
  color: #a1a1aa;
}
.home-provider-active:hover {
  border-color: rgba(124, 58, 237, 0.3);
  background: rgba(124, 58, 237, 0.03);
}
.dark .home-provider-active:hover {
  border-color: rgba(139, 92, 246, 0.3);
  background: rgba(139, 92, 246, 0.05);
}
.home-provider-soon { opacity: 0.5; }

.home-provider-icon {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.home-provider-icon span {
  font-size: 11px;
  font-weight: 800;
  color: white;
}

.home-provider-badge {
  padding: 1px 7px;
  border-radius: 4px;
  background: rgba(124, 58, 237, 0.08);
  color: #7c3aed;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.02em;
}
.dark .home-provider-badge {
  background: rgba(139, 92, 246, 0.1);
  color: #a78bfa;
}
.home-provider-badge-soon {
  padding: 1px 7px;
  border-radius: 4px;
  background: #f4f4f5;
  color: #71717a;
  font-size: 10px;
  font-weight: 600;
}
.dark .home-provider-badge-soon {
  background: #2a2a35;
  color: #6b6b76;
}

/* ── Footer ── */
.home-footer {
  position: relative;
  z-index: 10;
  border-top: 1px solid #e4e4e7;
  padding: 1.5rem;
}
.dark .home-footer { border-top-color: #1c1c22; }
.home-footer-inner {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  text-align: center;
}
@media (min-width: 640px) {
  .home-footer-inner {
    flex-direction: row;
    justify-content: space-between;
    text-align: left;
  }
}
.home-footer-copy {
  font-size: 12px;
  color: #a1a1aa;
}
.dark .home-footer-copy { color: #3d3d48; }
.home-footer-links {
  display: flex;
  gap: 1.25rem;
}
.home-footer-link {
  font-size: 12px;
  color: #a1a1aa;
  text-decoration: none;
  transition: color 0.15s;
}
.home-footer-link:hover { color: #18181b; }
.dark .home-footer-link { color: #3d3d48; }
.dark .home-footer-link:hover { color: #fafafa; }
</style>
