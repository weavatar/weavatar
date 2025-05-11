<template>
  <NLayoutHeader bordered class="fixed top-0 z-999 w-full bg-white dark:bg-dark">
    <div class="max-w-7xl px-4 mx-auto h-16 flex items-center justify-between">
      <!-- Logo区域 -->
      <div class="flex items-center cursor-pointer" @click="router.push({ name: 'home' })">
        <NImage
          :src="currentLogo"
          alt="WeAvatar"
          :height="32"
          preview-disabled
          class="mr-2"
        />
      </div>

      <!-- 导航区域 -->
      <div class="flex items-center">
        <!-- PC端菜单 -->
        <div class="hidden md:block">
          <NMenu
            v-model:value="activeKey"
            mode="horizontal"
            :options="menuOptions"
            :watch-props="['defaultValue']"
          />
        </div>
        
        <!-- 右侧用户区域 -->
        <div class="ml-auto flex items-center gap-4">
          <!-- 登录用户下拉菜单 -->
          <NDropdown v-if="userStore.auth.login" trigger="hover" :options="userOptions" @select="handleSelect">
            <div class="flex items-center gap-2 cursor-pointer">
              <NAvatar size="small" :src="userStore.info.avatar" />
              <NText class="hidden sm:block">{{ userStore.info.nickname }}</NText>
            </div>
          </NDropdown>
          
          <!-- 未登录状态的按钮 -->
          <div v-else>
            <NButton 
              type="info"
              @click="router.push({ name: 'login' })"
            >
              开始使用
            </NButton>
          </div>

          <!-- 深色模式切换 -->
          <NButton text @click="toggleDark()">
            <template #icon>
              <NIcon size="18">
                <SunIcon v-if="isDark" />
                <MoonIcon v-else />
              </NIcon>
            </template>
          </NButton>
        </div>
      </div>
    </div>
  </NLayoutHeader>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { computed, h, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NAvatar, NButton, NDropdown, NIcon, NImage, NLayoutHeader, NMenu, NText } from 'naive-ui'
import {
  DocumentTextOutline as DocumentTextIcon,
  HomeOutline as HomeIcon,
  InformationCircleOutline as InformationCircleIcon,
  LogInOutline as LoginIcon,
  Menu as MenuIcon,
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

const currentLogo = computed(() => isDark.value ? logoWhite : logo)

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

// 主菜单
const menuOptions = ref()
// 用户菜单
const userOptions = ref()

const activeKey = computed({
  get: () => {
    return route.name as string
  },
  set: (value) => {
    router.push({
      name: value
    })
  }
})

// 监听登录状态
watch(
  () => userStore.auth.login,
  (login) => {
    if (login) {
      useUserStore().freshUserInfo()
    } else {
      useUserStore().resetUserInfo()
    }
    menuOptions.value = [
      {
        label: '首页',
        key: 'home',
        icon: renderIcon(HomeIcon),
        show: true
      },
      {
        label: '头像',
        key: 'user-avatar',
        icon: renderIcon(PersonCircleIcon),
        show: login
      },
      {
        label: '文档',
        key: 'doc',
        icon: renderIcon(DocumentTextIcon),
        show: true
      },
      {
        label: '关于',
        key: 'about',
        icon: renderIcon(InformationCircleIcon),
        show: true
      }
    ].filter(item => item.show)
    
    userOptions.value = [
      {
        label: '头像管理',
        key: 'user-avatar',
        show: login
      },
      {
        label: '我的资料',
        key: 'user-info',
        show: login
      },
      {
        label: '登出',
        key: 'logout',
        show: login
      }
    ].filter(item => item.show)
  },
  { immediate: true }
)

function handleSelect(key: string): void {
  router.push({
    name: key
  })
}
</script>

<style scoped>
</style>
