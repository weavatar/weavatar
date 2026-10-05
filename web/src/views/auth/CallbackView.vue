<template>
  <div class="flex-1 flex items-center justify-center px-6 sm:px-10 py-20 min-h-[60vh]">
    <div class="w-full max-w-sm surface p-8 sm:p-10 text-center shadow-card">
      <template v-if="status === 'pending'">
        <div
          class="mx-auto w-12 h-12 rounded-full bg-brand/10 text-brand flex items-center justify-center"
        >
          <span class="i-lucide-loader-circle animate-spin text-2xl" />
        </div>
        <h1 class="mt-5 text-lg font-700 text-fg">正在登录</h1>
        <p class="mt-1.5 text-sm text-fg2">请稍候…</p>
      </template>

      <template v-else-if="status === 'success'">
        <div
          class="mx-auto w-12 h-12 rounded-full bg-green-500/10 text-green-500 flex items-center justify-center"
        >
          <span class="i-lucide-check text-2xl" />
        </div>
        <h1 class="mt-5 text-lg font-700 text-fg">登录成功</h1>
        <p class="mt-1.5 text-sm text-fg2">即将进入头像管理…</p>
      </template>

      <template v-else>
        <div
          class="mx-auto w-12 h-12 rounded-full bg-red-500/10 text-red-500 flex items-center justify-center"
        >
          <span class="i-lucide-x text-2xl" />
        </div>
        <h1 class="mt-5 text-lg font-700 text-fg">登录失败</h1>
        <p class="mt-1.5 text-sm text-fg2">授权信息无效或已过期，请重新登录。</p>
        <router-link :to="{ name: 'login' }" class="btn-primary mt-6">返回登录</router-link>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useRequest } from 'alova/client'
import { useUserStore } from '@/stores'
import auth from '@/api/auth'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const status = ref<'pending' | 'success' | 'error'>('pending')

const code = String(route.query.code || '')
const state = String(route.query.state || '')

if (!code || !state) {
  status.value = 'error'
} else {
  useRequest(auth.callback(code, state))
    .onSuccess(({ data }: any) => {
      status.value = 'success'
      userStore.updateToken(data.token)
      window.$message.success('登录成功')
      setTimeout(() => router.replace({ name: 'user-avatar' }), 800)
    })
    .onError(() => {
      status.value = 'error'
    })
}
</script>
