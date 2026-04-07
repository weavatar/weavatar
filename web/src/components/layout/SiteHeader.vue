<template>
  <NLayoutHeader bordered class="fixed top-0 z-999 w-full bg-white dark:bg-dark">
    <div class="px-10 h-16 flex items-center justify-between">
      <!-- 左侧：Logo + 导航菜单 -->
      <div class="flex items-center">
        <div class="flex items-center cursor-pointer" @click="router.push({ name: 'home' })">
          <NImage :src="currentLogo" alt="WeAvatar" :height="32" preview-disabled class="mr-2" />
        </div>
        <div class="hidden md:block">
          <NMenu
            v-model:value="activeKey"
            mode="horizontal"
            :options="menuOptions"
            :watch-props="['defaultValue']"
          />
        </div>
      </div>

      <!-- 右侧：深色模式 + 用户区域 -->
      <div class="flex items-center gap-4">
        <NButton text @click="toggleDark()">
          <template #icon>
            <NIcon size="18">
              <SunIcon v-if="isDark" />
              <MoonIcon v-else />
            </NIcon>
          </template>
        </NButton>

        <NDropdown
          v-if="userStore.auth.login"
          trigger="hover"
          :options="userOptions"
          @select="handleSelect"
        >
          <div class="flex items-center gap-2 cursor-pointer">
            <NAvatar size="small" :src="userStore.info.avatar" />
            <NText class="hidden sm:block">{{ userStore.info.nickname }}</NText>
          </div>
        </NDropdown>

        <NButton v-else type="info" @click="router.push({ name: 'login' })"> 开始使用 </NButton>
      </div>
    </div>
  </NLayoutHeader>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { computed, h, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NAvatar, NButton, NDropdown, NIcon, NImage, NLayoutHeader, NMenu, NText } from 'naive-ui'
import {
  DocumentTextOutline as DocumentTextIcon,
  HomeOutline as HomeIcon,
  InformationCircleOutline as InformationCircleIcon,
  PersonCircleOutline as PersonCircleIcon,
  SunnyOutline as SunIcon,
  MoonOutline as MoonIcon
} from '@vicons/ionicons5'
import { useDark, useToggle } from '@vueuse/core'

import logo from '@/assets/logo.png'
import logoWhite from '@/assets/logo-white.png'
import { useUserStore } from '@/stores'

const isDark = useDark()
const toggleDark = useToggle(isDark)
const currentLogo = computed(() => (isDark.value ? logoWhite : logo))

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const renderIcon = (icon: Component) => () => h(NIcon, null, { default: () => h(icon) })

const activeKey = computed({
  get: () => route.name as string,
  set: (value) => router.push({ name: value })
})

const menuOptions = computed(() =>
  [
    { label: '首页', key: 'home', icon: renderIcon(HomeIcon) },
    {
      label: '头像',
      key: 'user-avatar',
      icon: renderIcon(PersonCircleIcon),
      show: userStore.auth.login
    },
    { label: '文档', key: 'doc', icon: renderIcon(DocumentTextIcon) },
    { label: '关于', key: 'about', icon: renderIcon(InformationCircleIcon) }
  ].filter((item) => item.show !== false)
)

const userOptions = computed(() => [
  { label: '头像管理', key: 'user-avatar' },
  { label: '我的资料', key: 'user-info' },
  { label: '登出', key: 'logout' }
])

watch(
  () => userStore.auth.login,
  (login) => {
    if (login) {
      userStore.freshUserInfo()
    }
  },
  { immediate: true }
)

const handleSelect = (key: string) => router.push({ name: key })
</script>
