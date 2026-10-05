<template>
  <div class="flex-1 flex items-center justify-center px-6 sm:px-10 py-20 min-h-[60vh]">
    <div class="w-full max-w-sm surface p-8 sm:p-10 text-center shadow-card">
      <div
        class="mx-auto w-12 h-12 rounded-full bg-brand/10 text-brand flex items-center justify-center"
      >
        <span v-if="done" class="i-lucide-check text-2xl" />
        <span v-else class="i-lucide-loader-circle animate-spin text-2xl" />
      </div>
      <h1 class="mt-5 text-lg font-700 text-fg">{{ done ? '已成功退出' : '正在退出…' }}</h1>
      <p class="mt-1.5 text-sm text-fg2">
        {{ done ? '页面即将自动跳转。' : '正在清理登录状态。' }}
      </p>
      <router-link v-if="done" :to="{ name: 'home' }" class="btn-secondary mt-6"
        >回到首页</router-link
      >
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useRequest } from 'alova/client'
import { useUserStore } from '@/stores'
import auth from '@/api/auth'

const userStore = useUserStore()
const router = useRouter()
const done = ref(false)

useRequest(auth.logout(), { meta: { noAlert: true } }).onComplete(() => {
  userStore.clearToken()
  done.value = true
  setTimeout(() => router.replace({ name: 'login' }), 1000)
})
</script>
