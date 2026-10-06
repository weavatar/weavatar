<template>
  <div
    class="relative flex-1 flex items-center justify-center px-6 sm:px-10 py-20 min-h-[70vh] overflow-x-clip"
  >
    <div
      class="absolute inset-0 -top-16 wa-grid wa-mask-radial pointer-events-none"
      aria-hidden="true"
    />
    <div
      class="absolute -top-40 left-1/2 -translate-x-1/2 w-[40rem] h-[20rem] rounded-full bg-brand/12 blur-[120px] pointer-events-none"
      aria-hidden="true"
    />

    <div v-reveal class="relative w-full max-w-sm surface p-8 sm:p-10 shadow-card">
      <brand-logo :height="28" />
      <h1 class="mt-8 display-text text-2xl text-fg">登录 WeAvatar</h1>
      <p class="mt-2 text-sm text-fg2 leading-relaxed">用树新峰通行证一键登录。</p>

      <button
        type="button"
        class="btn-primary btn-lg w-full mt-8"
        :disabled="loading"
        @click="handleLogin"
      >
        <span v-if="loading" class="i-lucide-loader-circle animate-spin text-base" />
        <span v-else class="i-lucide-key-round text-base" />
        {{ loading ? '正在跳转…' : '使用树新峰通行证登录' }}
      </button>

      <p class="mt-6 text-xs text-fg3 leading-relaxed">
        登录即表示你同意我们的
        <router-link :to="{ name: 'privacy' }" class="link">隐私政策</router-link>
        。树新峰通行证由
        <a :href="ACCOUNT_URL" target="_blank" rel="noreferrer" class="link"> account.haozi.net </a>
        提供。
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import BrandLogo from '@/components/ui/BrandLogo.vue'
import { fetchLoginUrl } from '@/api/auth'
import { ACCOUNT_URL } from '@/constants/links'
import { safeRedirect } from '@/utils/redirect'

const route = useRoute()
const loading = ref(false)

const handleLogin = async () => {
  loading.value = true
  try {
    const { url } = await fetchLoginUrl()
    // 放弃的注销不能劫持本次登录
    sessionStorage.removeItem('oauth_intent')
    const redirect = safeRedirect(route.query.redirect)
    if (redirect) sessionStorage.setItem('login_redirect', redirect)
    else sessionStorage.removeItem('login_redirect')
    window.location.href = url
  } catch {
    loading.value = false
  }
}
</script>
