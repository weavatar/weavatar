<template>
  <NLayoutHeader bordered class="fixed top-0 z-10 w-full bg-white">
    <div class="mx-auto px-16 h-16 flex items-center justify-between">
      <!-- Logo区域 -->
      <div class="flex items-center cursor-pointer" @click="router.push({ name: 'home' })">
        <NImage
          :src="logo"
          alt="WeAvatar"
          :height="32"
          preview-disabled
          class="mr-2"
        />
      </div>

      <!-- 导航区域 -->
      <div class="flex-1 flex justify-between items-center">
        <!-- PC端菜单 -->
        <div class="hidden md:block ml-8">
          <NMenu
            v-model:value="activeKey"
            mode="horizontal"
            :options="menuOptions"
            :watch-props="['defaultValue']"
          />
        </div>
        
        <!-- 右侧用户区域 -->
        <div class="flex items-center gap-4">
          <!-- 登录用户下拉菜单 -->
          <NDropdown v-if="userStore.auth.login" trigger="hover" :options="userOptions" @select="handleSelect">
            <div class="flex items-center gap-2 cursor-pointer">
              <NAvatar size="small" :src="userStore.info.avatar" />
              <NText class="hidden sm:block">{{ userStore.info.nickname }}</NText>
            </div>
          </NDropdown>
          
          <!-- 未登录状态的按钮 -->
          <div v-else class="flex gap-3">
            <NButton 
              quaternary 
              @click="router.push({ name: 'login' })"
              class="hidden sm:flex items-center border border-[#e5e7eb] hover:bg-gray-50"
            >
              登录
            </NButton>
            <NButton 
              type="primary" 
              @click="router.push({ name: 'login' })"
            >
              开始使用
            </NButton>
          </div>

          <!-- 移动端菜单按钮 -->
          <div class="md:hidden">
            <NDropdown trigger="click" :options="menuOptions" @select="handleSelect">
              <NIcon :component="MenuIcon" color="#2080f0" size="24" />
            </NDropdown>
          </div>
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
  PersonCircleOutline as PersonCircleIcon
} from '@vicons/ionicons5'

import logo from '@/assets/logo.png'
import { useUserStore } from '@/stores'

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
:deep(.n-space div:nth-child(1)) {
  display: flex !important;
  justify-content: center !important;
  align-items: center !important;
}

.right {
  flex: 1;
  width: 100%;
  display: flex;
  justify-content: space-between;
}

.nav {
  padding: 10px 32px;
  position: fixed;
  top: 0;
  z-index: 1;
  display: grid;
  align-items: center;
  --side-padding: 32px;
  grid-template-columns: calc(272px - var(--side-padding)) 1fr auto;
}

.nav-end {
  display: flex;
  align-items: center;
}

.pc-menu {
  display: flex;
  align-items: center;
}

.mobile-menu {
  display: none;
}

@media screen and (max-width: 768px) {
  .right {
    justify-content: flex-end;
  }

  .pc-menu {
    display: none;
  }

  .mobile-menu {
    display: flex;
    align-items: center;
    padding-left: 10px;
  }

  :deep(.n-menu-item-content-header) {
    display: none;
  }

  :deep(.n-space div:nth-child(2)) {
    display: none;
  }
}
</style>
