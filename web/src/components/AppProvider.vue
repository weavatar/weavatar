<template>
  <slot />
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useDialog, useLoadingBar, useMessage, useNotification } from 'naive-ui'
import { useUserStore } from '@/stores'

// 暴露给 http 层等非组件上下文使用
window.$message = useMessage()
window.$dialog = useDialog()
window.$notification = useNotification()
window.$loadingBar = useLoadingBar()

const userStore = useUserStore()
watch(
  () => userStore.auth.login,
  (login) => {
    if (login) userStore.freshUserInfo()
  },
  { immediate: true }
)
</script>
