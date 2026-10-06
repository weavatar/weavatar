<template>
  <div class="flex-1 flex items-center justify-center px-6 sm:px-10 py-20 min-h-[60vh]">
    <div class="w-full max-w-sm surface p-8 sm:p-10 text-center shadow-card">
      <template v-if="status === 'pending'">
        <div
          class="mx-auto w-12 h-12 rounded-full bg-brand/10 text-brand flex items-center justify-center"
        >
          <span class="i-lucide-loader-circle animate-spin text-2xl" />
        </div>
        <h1 class="mt-5 text-lg font-700 text-fg">{{ deleting ? '正在确认身份' : '正在登录' }}</h1>
        <p class="mt-1.5 text-sm text-fg2">请稍候…</p>
      </template>

      <template v-else-if="status === 'success'">
        <div
          class="mx-auto w-12 h-12 rounded-full bg-green-500/10 text-green-500 flex items-center justify-center"
        >
          <span class="i-lucide-check text-2xl" />
        </div>
        <h1 class="mt-5 text-lg font-700 text-fg">{{ deleting ? '账号已注销' : '登录成功' }}</h1>
        <p class="mt-1.5 text-sm text-fg2">
          {{ deleting ? '即将回到首页…' : '即将跳转…' }}
        </p>
      </template>

      <template v-else>
        <div
          class="mx-auto w-12 h-12 rounded-full bg-red-500/10 text-red-500 flex items-center justify-center"
        >
          <span class="i-lucide-x text-2xl" />
        </div>
        <h1 class="mt-5 text-lg font-700 text-fg">{{ deleting ? '注销失败' : '登录失败' }}</h1>
        <p class="mt-1.5 text-sm text-fg2">
          {{
            deleting
              ? '身份确认未通过或已超时，账号未被注销。'
              : '授权信息无效或已过期，请重新登录。'
          }}
        </p>
        <router-link v-if="deleting" :to="{ name: 'user-info' }" class="btn-primary mt-6">
          返回我的资料
        </router-link>
        <router-link v-else :to="{ name: 'login' }" class="btn-primary mt-6">返回登录</router-link>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter, type RouteLocationRaw } from 'vue-router'
import { useTimeoutFn } from '@vueuse/core'
import { loginCallback } from '@/api/auth'
import { confirmDeletion } from '@/api/user'
import { useUserStore } from '@/stores'
import { safeRedirect } from '@/utils/redirect'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const status = ref<'pending' | 'success' | 'error'>('pending')

// 注销复用登录回调，意图读完即删，免得放弃的注销劫持下次登录
const deleting = sessionStorage.getItem('oauth_intent') === 'delete'
sessionStorage.removeItem('oauth_intent')

// 离开页面时自动取消
const { start: goLater } = useTimeoutFn((to: RouteLocationRaw) => router.replace(to), 800, {
  immediate: false
})

const run = async () => {
  const code = String(route.query.code || '')
  const state = String(route.query.state || '')
  if (!code || !state) {
    status.value = 'error'
    return
  }

  try {
    if (deleting) {
      await confirmDeletion(code, state)
      status.value = 'success'
      userStore.clearToken()
      window.$message.success('账号已注销')
      return goLater({ name: 'home' })
    }

    const { token } = await loginCallback(code, state)
    const redirect = safeRedirect(sessionStorage.getItem('login_redirect'))
    sessionStorage.removeItem('login_redirect')
    status.value = 'success'
    userStore.updateToken(token)
    window.$message.success('登录成功')
    goLater(redirect || { name: 'user-avatar' })
  } catch {
    status.value = 'error'
  }
}

run()
</script>
