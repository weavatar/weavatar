<script setup lang="ts">
import { NIcon } from 'naive-ui'
import {
  AlbumsOutline as AlbumsIcon,
  CloudOutline as CloudIcon,
  IdCardOutline as IdCardIcon,
  ImageOutline as ImageIcon,
  SpeedometerOutline as SpeedIcon,
  ShieldCheckmarkOutline as ShieldIcon
} from '@vicons/ionicons5'

const features = [
  {
    icon: AlbumsIcon,
    title: '多级头像匹配',
    desc: '支持从 Gravatar、QQ 获取头像，可为 70% 以上的请求提供准确的头像'
  },
  {
    icon: IdCardIcon,
    title: '手机号 & 字母头像',
    desc: '首家支持手机号头像及字母默认头像，更符合国内使用习惯'
  },
  { icon: ImageIcon, title: '下一代图片格式', desc: '默认使用 WebP 格式，减少约 80% 的流量消耗' },
  {
    icon: ShieldIcon,
    title: 'AI 内容审核',
    desc: '全自动化 AI 审核机制，确保不会有违规内容被输出'
  },
  { icon: SpeedIcon, title: '极致性能', desc: 'Go 语言 + 多级缓存 + 全球 CDN，毫秒级头像响应' },
  { icon: CloudIcon, title: '开放平台', desc: '即将推出开放平台和配套 SDK，无缝对接您的应用' }
]
</script>

<template>
  <section class="py-20 px-6">
    <div class="max-w-360 mx-auto text-center">
      <h2 class="text-2xl sm:text-3xl font-700 mb-2">为什么选择 WeAvatar</h2>
      <p class="text-base opacity-50 mb-12">超越 Gravatar 的新一代头像服务</p>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
        <div
          v-for="(f, i) in features"
          :key="i"
          class="border border-gray-200 dark:border-gray-700 rounded-3 p-7 text-left transition-all duration-300 hover:shadow-lg hover:-translate-y-1"
        >
          <div
            class="w-11 h-11 flex items-center justify-center rounded-2.5 bg-blue-500/8 text-blue-500 mb-4"
          >
            <NIcon :size="26"><component :is="f.icon" /></NIcon>
          </div>
          <h3 class="text-base font-600 mb-1">{{ f.title }}</h3>
          <p class="text-sm leading-relaxed opacity-50 m-0">{{ f.desc }}</p>
        </div>
      </div>
    </div>
  </section>
</template>
