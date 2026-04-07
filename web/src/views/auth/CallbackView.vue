<template>
  <div class="w-100 max-w-md mx-auto mt-48">
    <NCard title="登录中">
      <NSpin :size="20" />
      <NText>请稍后...</NText>
    </NCard>
  </div>
</template>

<script setup lang="ts">
import { NCard, NSpin, NText } from 'naive-ui'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores'
import auth from '@/api/auth'
import { useRequest } from 'alova/client'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const code = String(route.query.code || '')
const state = String(route.query.state || '')

useRequest(auth.callback(code, state)).onSuccess(({ data }: any) => {
  window.$message.success('登录成功')
  userStore.updateToken(data.token)
  setTimeout(() => router.push({ name: 'home' }), 1000)
})
</script>
