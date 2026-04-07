<template>
  <div class="w-100 max-w-md mx-auto mt-24">
    <NCard title="登录">
      <NTabs default-value="haozi-login" size="large" justify-content="space-evenly">
        <NTabPane name="haozi-login" tab="耗子通行证">
          <NButton type="primary" block :loading="loading" @click="handleLogin">
            耗子通行证 登录
          </NButton>
        </NTabPane>
      </NTabs>
    </NCard>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { NButton, NCard, NTabPane, NTabs } from 'naive-ui'
import auth from '@/api/auth'
import { useRequest } from 'alova/client'

const loading = ref(false)

const handleLogin = () => {
  loading.value = true
  useRequest(auth.login())
    .onSuccess(({ data }: any) => {
      window.location.href = data.url
    })
    .onComplete(() => {
      loading.value = false
    })
}
</script>
