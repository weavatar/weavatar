<template>
  <slot />
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDialog, useLoadingBar, useMessage, useNotification } from 'naive-ui'
import { useUserStore } from '@/stores'

// 暴露给 http 层等非组件上下文使用
window.$message = useMessage()
window.$dialog = useDialog()
window.$notification = useNotification()
window.$loadingBar = useLoadingBar()

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

watch(
  () => userStore.isLogin,
  (login) => {
    if (login) return userStore.freshUserInfo()
    // 登录态失效（如接口 401）时离开需要登录的页面
    if (route.matched.some((r) => r.meta.requiresAuth)) {
      router.replace({ name: 'login', query: { redirect: route.fullPath } })
    }
  },
  { immediate: true }
)
</script>
