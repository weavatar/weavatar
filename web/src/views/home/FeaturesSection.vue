<template>
  <section class="relative section border-t border-line">
    <div class="wrap">
      <section-title
        title="为什么选择 WeAvatar"
        desc="不只是 Gravatar 镜像，从匹配策略到图片格式均重新设计。"
      />

      <div
        v-reveal
        class="mt-12 grid gap-px bg-line border border-line rounded-3xl overflow-hidden md:grid-cols-6"
      >
        <!-- 多级头像匹配 -->
        <article class="bg-elev p-7 sm:p-8 md:col-span-4 flex flex-col">
          <feature-head icon="i-lucide-layers" title="多级头像匹配">
            按 WeAvatar、Gravatar、社交头像的顺序逐级匹配，70% 以上的请求都能命中真实头像。
          </feature-head>
          <div class="mt-auto pt-8">
            <div
              class="flex flex-col items-start gap-2 sm:flex-row sm:flex-wrap sm:items-center sm:gap-3"
            >
              <template v-for="(node, i) in chain" :key="node.label">
                <div
                  class="inline-flex items-center gap-2 h-10 px-3.5 rounded-xl border text-sm font-600 whitespace-nowrap"
                  :class="
                    i === 0
                      ? 'bg-brand text-white border-brand shadow-brand'
                      : 'bg-bg border-line-strong text-fg'
                  "
                >
                  <span :class="node.icon" class="text-base" />
                  {{ node.label }}
                </div>
                <template v-if="i < chain.length - 1">
                  <div class="dash-line-y h-5 ml-5 sm:hidden" />
                  <div class="dash-line hidden sm:block flex-1 min-w-6" />
                </template>
              </template>
            </div>
          </div>
        </article>

        <!-- 手机号 & 字母头像 -->
        <article class="bg-elev p-7 sm:p-8 md:col-span-2 flex flex-col">
          <feature-head icon="i-lucide-smartphone" title="手机号 & 字母头像">
            支持手机号作为头像标识，内置字母默认头像，更符合国内使用习惯。
          </feature-head>
          <div class="mt-auto pt-8 flex items-center gap-3">
            <avatar-image
              v-for="a in initialsDemo"
              :key="a"
              :src="a"
              :size="48"
              rounded="xl"
              class="ring-1 ring-line"
            />
            <span class="ml-1 font-mono text-xs text-fg3">d=initials</span>
          </div>
        </article>

        <!-- 下一代图片格式 -->
        <article class="bg-elev p-7 sm:p-8 md:col-span-2 flex flex-col">
          <feature-head icon="i-lucide-image" title="下一代图片格式">
            默认输出 WebP，节省约 80% 流量，另支持 AVIF、HEIC、JXL 等 10 种格式。
          </feature-head>
          <div class="mt-auto pt-8 flex flex-wrap gap-1.5">
            <span
              v-for="f in formats"
              :key="f"
              class="chip font-mono"
              :class="f === 'webp' ? 'chip-brand' : ''"
            >
              .{{ f }}
            </span>
          </div>
        </article>

        <!-- AI 内容审核 -->
        <article class="bg-elev p-7 sm:p-8 md:col-span-2 flex flex-col">
          <feature-head icon="i-lucide-shield-check" title="AI 内容审核">
            每张头像都经过 AI 自动审核，违规内容不会输出。
          </feature-head>
          <div class="mt-auto pt-8 flex items-center gap-3 text-sm">
            <span class="relative flex w-2.5 h-2.5">
              <span class="absolute inset-0 rounded-full bg-green-500 animate-pulse-ring" />
              <span class="relative w-2.5 h-2.5 rounded-full bg-green-500" />
            </span>
            <span class="text-fg2">全自动审核，无需人工介入</span>
          </div>
        </article>

        <!-- 极致性能 -->
        <article class="bg-elev p-7 sm:p-8 md:col-span-2 flex flex-col">
          <feature-head icon="i-lucide-gauge" title="极致性能">
            Go 语言编写，多级缓存加全球 CDN，毫秒级响应。
          </feature-head>
          <div class="mt-auto pt-8 flex items-end gap-1.5">
            <span class="text-3xl font-800 tracking-tight text-fg leading-none">毫秒级</span>
            <span class="text-sm text-fg2 pb-0.5">响应，全球 CDN 加速</span>
          </div>
        </article>

        <!-- 开源 -->
        <article class="bg-elev p-7 sm:p-8 md:col-span-3 flex flex-col">
          <feature-head icon="i-lucide-github" title="开源透明">
            前后端代码全部开源，可自行部署。
          </feature-head>
          <div class="mt-auto pt-8">
            <a
              :href="GITHUB_URL"
              target="_blank"
              rel="noreferrer"
              class="link inline-flex items-center gap-1 text-sm font-600"
            >
              github.com/weavatar/weavatar
              <span class="i-lucide-arrow-up-right text-sm" />
            </a>
          </div>
        </article>

        <!-- 开放平台 -->
        <article class="bg-elev p-7 sm:p-8 md:col-span-3 flex flex-col">
          <feature-head icon="i-lucide-blocks" title="开放平台">
            开放平台与 SDK 即将推出，可为不同应用设置不同头像。
          </feature-head>
          <div class="mt-auto pt-8">
            <span class="chip">
              <span class="i-lucide-sparkles text-sm text-brand" />
              即将推出
            </span>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import SectionTitle from '@/components/ui/SectionTitle.vue'
import AvatarImage from '@/components/ui/AvatarImage.vue'
import FeatureHead from './FeatureHead.vue'
import { AVATAR_BASE, GITHUB_URL } from '@/constants/links'

const chain = [
  { label: 'WeAvatar', icon: 'i-lucide-badge-check' },
  { label: 'Gravatar', icon: 'i-lucide-globe' },
  { label: '社交头像', icon: 'i-lucide-message-circle' },
  { label: '默认头像', icon: 'i-lucide-user-round' }
]

const formats = ['webp', 'avif', 'heic', 'jxl', 'png', 'jpg', 'gif', 'tiff']

const initialsDemo = [
  `${AVATAR_BASE}/demo-1?d=initials&initials=WA&f=y&s=96`,
  `${AVATAR_BASE}/demo-2?d=initials&name=Haozi&f=y&s=96`,
  `${AVATAR_BASE}/demo-3?d=color&f=y&s=96`
]
</script>
