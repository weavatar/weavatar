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
          {{ deleting ? '即将回到首页…' : '即将进入头像管理…' }}
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
            deleting ? '身份确认未通过或已超时，账号未被注销。' : '授权信息无效或已过期，请重新登录。'
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
import { useRoute, useRouter } from 'vue-router'
import { useRequest } from 'alova/client'
import { useUserStore } from '@/stores'
import auth from '@/api/auth'
import userApi from '@/api/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const status = ref<'pending' | 'success' | 'error'>('pending')

// account deletion reuses the login callback; the intent is consumed at once
// so an abandoned deletion cannot hijack the next login
const deleting = sessionStorage.getItem('oauth_intent') === 'delete'
sessionStorage.removeItem('oauth_intent')

const code = String(route.query.code || '')
const state = String(route.query.state || '')

if (!code || !state) {
  status.value = 'error'
} else if (deleting) {
  useRequest(userApi.deletionConfirm(code, state))
    .onSuccess(() => {
      status.value = 'success'
      userStore.clearToken()
      window.$message.success('账号已注销')
      setTimeout(() => router.replace({ name: 'home' }), 800)
    })
    .onError(() => {
      status.value = 'error'
      setTimeout(() => router.replace({ name: 'user-info' }), 1500)
    })
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
