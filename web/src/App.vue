<template>
  <NConfigProvider :locale="zhCN" :date-locale="dateZhCN" :theme="theme" :hljs="hljs" :theme-overrides="themeOverrides">
    <NLoadingBarProvider>
      <NMessageProvider>
        <NNotificationProvider>
          <NDialogProvider>
            <SiteLayout />
          </NDialogProvider>
        </NNotificationProvider>
      </NMessageProvider>
    </NLoadingBarProvider>
  </NConfigProvider>
</template>

<script setup lang="ts">
import SiteLayout from '@/views/SiteLayout.vue'
import {
  darkTheme,
  dateZhCN,
  NConfigProvider,
  NDialogProvider,
  NLoadingBarProvider,
  NMessageProvider,
  NNotificationProvider,
  zhCN
} from 'naive-ui'
import hljs from 'highlight.js/lib/core'
import php from 'highlight.js/lib/languages/php'
import { computed } from 'vue'
import { useDark } from '@vueuse/core'

hljs.registerLanguage('php', php)

const isDark = useDark()

const theme = computed(() => (isDark.value ? darkTheme : null))

const themeOverrides = {
  common: {
    primaryColor: '#3A84F7'
  }
}
</script>
