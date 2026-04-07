<template>
  <div class="flex flex-col justify-center items-center h-[60vh]">
    <h1>已成功退出</h1>
    <p>页面即将自动跳转</p>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores'
import auth from '@/api/auth'
import { useRequest } from 'alova/client'

const userStore = useUserStore()
const router = useRouter()

useRequest(auth.logout(), { meta: { noAlert: true } }).onComplete(() => {
  userStore.clearToken()
  setTimeout(() => router.push({ name: 'login' }), 1000)
})
</script>
