<template>
  <NLayout class="min-h-screen flex flex-col">
    <SiteHeader />
    <div class="flex-grow mt-16">
      <RouterView />
    </div>
    <SiteFooter />
  </NLayout>
</template>

<script setup lang="ts">
import { NLayout, useDialog, useLoadingBar, useMessage, useNotification } from 'naive-ui'
import SiteHeader from '@/components/layout/SiteHeader.vue'
import SiteFooter from '@/components/layout/SiteFooter.vue'

window.$message = useMessage()
window.$dialog = useDialog()
window.$notification = useNotification()
window.$loadingBar = useLoadingBar()
</script>

<style scoped>
:deep(.n-layout-scroll-container) {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
</style>
