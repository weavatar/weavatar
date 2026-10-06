<template>
  <header
    class="sticky top-0 z-50 border-b transition-[background-color,border-color,backdrop-filter] duration-300"
    :class="
      scrolled || mobileOpen
        ? 'bg-bg/80 backdrop-blur-xl border-line'
        : 'bg-transparent border-transparent'
    "
  >
    <div class="wrap h-16 flex items-center justify-between gap-6">
      <router-link
        :to="{ name: 'home' }"
        class="flex items-center shrink-0"
        aria-label="WeAvatar 首页"
      >
        <brand-logo :height="26" />
      </router-link>

      <nav class="hidden md:flex items-center gap-0.5" aria-label="主导航">
        <router-link
          v-for="item in navItems"
          :key="item.name"
          :to="{ name: item.name }"
          class="nav-link"
          exact-active-class="text-fg bg-fg/6"
        >
          {{ item.label }}
        </router-link>
      </nav>

      <!-- 操作区 -->
      <div class="flex items-center gap-1.5">
        <a
          :href="GITHUB_URL"
          target="_blank"
          rel="noreferrer"
          class="btn-ghost btn-icon hidden sm:inline-flex"
          aria-label="GitHub"
          title="GitHub"
        >
          <span class="i-lucide-github text-lg" />
        </a>
        <theme-toggle />

        <template v-if="userStore.auth.login">
          <n-dropdown
            trigger="click"
            :options="userOptions"
            placement="bottom-end"
            @select="onSelect"
          >
            <button
              type="button"
              class="ml-1 flex items-center gap-2 pl-1 pr-2.5 h-10 rounded-full hover:bg-fg/6 transition-colors"
            >
              <avatar-image :src="userStore.info.avatar" :size="28" alt="" />
              <span class="hidden sm:block text-sm font-500 max-w-28 truncate">
                {{ userStore.info.nickname }}
              </span>
              <span class="i-lucide-chevron-down text-fg3 text-sm" />
            </button>
          </n-dropdown>
        </template>
        <template v-else>
          <router-link :to="{ name: 'login' }" class="btn-ghost btn-sm hidden sm:inline-flex">
            登录
          </router-link>
          <router-link :to="{ name: 'login' }" class="btn-primary btn-sm ml-1"
            >开始使用</router-link
          >
        </template>

        <button
          type="button"
          class="btn-ghost btn-icon md:hidden"
          :aria-expanded="mobileOpen"
          aria-label="打开菜单"
          @click="mobileOpen = !mobileOpen"
        >
          <span class="text-xl" :class="mobileOpen ? 'i-lucide-x' : 'i-lucide-menu'" />
        </button>
      </div>
    </div>

    <transition name="drop">
      <div
        v-if="mobileOpen"
        class="md:hidden absolute inset-x-0 top-full bg-bg border-b border-line shadow-card"
      >
        <nav class="wrap py-3 flex flex-col" aria-label="移动端导航">
          <router-link
            v-for="item in navItems"
            :key="item.name"
            :to="{ name: item.name }"
            class="nav-link py-3 text-base"
            exact-active-class="text-fg bg-fg/6"
          >
            {{ item.label }}
          </router-link>
          <template v-if="userStore.auth.login">
            <div class="my-2 border-t border-line" />
            <router-link :to="{ name: 'user-info' }" class="nav-link py-3 text-base"
              >我的资料</router-link
            >
            <router-link :to="{ name: 'logout' }" class="nav-link py-3 text-base text-red-500">
              退出登录
            </router-link>
          </template>
          <template v-else>
            <div class="my-2 border-t border-line" />
            <router-link :to="{ name: 'login' }" class="nav-link py-3 text-base">登录</router-link>
          </template>
        </nav>
      </div>
    </transition>
  </header>
</template>

<script setup lang="ts">
import { computed, h, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NDropdown, type DropdownOption } from 'naive-ui'
import { useWindowScroll } from '@vueuse/core'
import BrandLogo from '@/components/ui/BrandLogo.vue'
import ThemeToggle from '@/components/ui/ThemeToggle.vue'
import AvatarImage from '@/components/ui/AvatarImage.vue'
import { useUserStore } from '@/stores'
import { GITHUB_URL } from '@/constants/links'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const { y } = useWindowScroll()
const scrolled = computed(() => y.value > 8)
const mobileOpen = ref(false)

watch(
  () => route.fullPath,
  () => {
    mobileOpen.value = false
  }
)

const navItems = computed(() => {
  const items = [
    { label: '首页', name: 'home' },
    { label: '文档', name: 'doc' },
    { label: '帮助', name: 'help' },
    { label: '关于', name: 'about' }
  ]
  if (userStore.auth.login) items.splice(1, 0, { label: '头像管理', name: 'user-avatar' })
  return items
})

const icon = (cls: string) => () => h('span', { class: `${cls} text-base` })

const userOptions: DropdownOption[] = [
  { label: '头像管理', key: 'user-avatar', icon: icon('i-lucide-images') },
  { label: '我的资料', key: 'user-info', icon: icon('i-lucide-user-round') },
  { type: 'divider', key: 'd1' },
  { label: '退出登录', key: 'logout', icon: icon('i-lucide-log-out') }
]

const onSelect = (key: string) => router.push({ name: key })
</script>

<style scoped>
.drop-enter-active,
.drop-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.drop-enter-from,
.drop-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
