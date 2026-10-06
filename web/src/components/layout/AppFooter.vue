<template>
  <footer class="border-t border-line bg-elev">
    <div class="wrap py-14 sm:py-16">
      <div class="grid grid-cols-2 gap-10 md:grid-cols-[1.5fr_1fr_1fr_1fr]">
        <div class="col-span-2 md:col-span-1 max-w-xs">
          <brand-logo :height="26" />
          <p class="mt-5 text-sm text-fg2 leading-relaxed">每个人的头像。一次设置，随处可见。</p>
          <div class="mt-6 flex items-center gap-2">
            <a :href="QQ_GROUP_URL" target="_blank" rel="noreferrer" class="btn-secondary btn-sm">
              <span class="i-lucide-message-circle text-base" />
              加入 QQ 群
            </a>
            <a
              :href="GITHUB_URL"
              target="_blank"
              rel="noreferrer"
              class="btn-ghost btn-sm btn-icon w-9 h-9"
              aria-label="GitHub"
            >
              <span class="i-lucide-github text-base" />
            </a>
          </div>
        </div>

        <div v-for="col in columns" :key="col.title">
          <div class="text-sm font-600 text-fg mb-4">{{ col.title }}</div>
          <ul class="space-y-2.5 text-sm">
            <li v-for="item in col.items" :key="item.label">
              <router-link
                v-if="item.to"
                :to="item.to"
                class="text-fg2 hover:text-fg transition-colors"
              >
                {{ item.label }}
              </router-link>
              <a
                v-else
                :href="item.href"
                target="_blank"
                rel="noreferrer"
                class="inline-flex items-center gap-1 text-fg2 hover:text-fg transition-colors"
              >
                {{ item.label }}
                <span class="i-lucide-arrow-up-right text-xs text-fg3" />
              </a>
            </li>
          </ul>
        </div>
      </div>

      <div
        class="mt-14 pt-6 border-t border-line flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between text-xs text-fg3"
      >
        <div>© 2022–{{ year }} WeAvatar 版权所有</div>
        <div class="flex flex-wrap items-center gap-x-4 gap-y-2">
          <a
            :href="ICP_URL"
            target="_blank"
            rel="noreferrer"
            class="hover:text-fg transition-colors"
          >
            {{ ICP_NUMBER }}
          </a>
          <a
            :href="GONGAN_URL"
            target="_blank"
            rel="noreferrer"
            class="inline-flex items-center gap-1.5 hover:text-fg transition-colors"
          >
            <img :src="beianGongan" alt="" width="14" height="14" class="inline-block" />
            {{ GONGAN_NUMBER }}
          </a>
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import BrandLogo from '@/components/ui/BrandLogo.vue'
import beianGongan from '@/assets/beian-gongan.png'
import {
  GITHUB_URL,
  GONGAN_NUMBER,
  GONGAN_URL,
  ICP_NUMBER,
  ICP_URL,
  QQ_GROUP_URL,
  STATUS_URL
} from '@/constants/links'

interface FooterItem {
  label: string
  to?: RouteLocationRaw
  href?: string
}

const year = new Date().getFullYear()

const columns: { title: string; items: FooterItem[] }[] = [
  {
    title: '产品',
    items: [
      { label: '文档', to: { name: 'doc' } },
      { label: '帮助', to: { name: 'help' } },
      { label: '服务状态', href: STATUS_URL },
      { label: '隐私政策', to: { name: 'privacy' } }
    ]
  },
  {
    title: '相关项目',
    items: [
      { label: '树新峰通行证', href: 'https://account.haozi.net/' },
      { label: 'Moe Tom', href: 'https://tom.moe/' },
      { label: 'Twikoo', href: 'https://twikoo.js.org/' },
      { label: 'Artalk', href: 'https://artalk.js.org/' }
    ]
  },
  {
    title: '开源项目',
    items: [
      { label: 'AcePanel', href: 'https://github.com/acepanel/panel' },
      { label: 'GORM SQLite', href: 'https://github.com/libtnb/sqlite' },
      { label: 'Validator', href: 'https://github.com/libtnb/validator' },
      { label: 'Cron', href: 'https://github.com/libtnb/cron' }
    ]
  }
]
</script>
