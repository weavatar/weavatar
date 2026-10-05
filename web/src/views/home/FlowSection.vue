<template>
  <section class="relative section border-t border-line">
    <div class="wrap">
      <section-title
        title="三步接入，无需 SDK"
        desc="接口即图片地址，任何能显示图片的地方都可使用。"
      />

      <div v-reveal class="mt-12 grid gap-6 lg:grid-cols-3 lg:gap-0">
        <div
          v-for="(step, i) in steps"
          :key="step.title"
          class="relative flex flex-col lg:px-8 first:lg:pl-0 last:lg:pr-0"
        >
          <!-- 步骤间连接线 -->
          <div
            v-if="i < steps.length - 1"
            class="hidden lg:block absolute top-5 left-[calc(100%-2rem)] w-16 dash-line z-10"
            aria-hidden="true"
          />
          <div class="flex items-center gap-3">
            <span
              class="w-10 h-10 rounded-full border border-line-strong bg-elev flex items-center justify-center font-mono text-sm font-600 text-fg tabular"
            >
              {{ String(i + 1).padStart(2, '0') }}
            </span>
            <h3 class="text-lg font-700 text-fg tracking-tight">{{ step.title }}</h3>
          </div>
          <p class="mt-3 text-sm text-fg2 leading-relaxed lg:min-h-16">{{ step.desc }}</p>
          <div class="mt-4 flex-1 flex flex-col">
            <code-block
              :code="step.code"
              :lang="step.lang"
              :title="step.file"
              wrap
              class="flex-1"
            />
          </div>
        </div>
      </div>

      <div v-reveal class="mt-12 flex flex-wrap items-center gap-x-6 gap-y-3">
        <span class="text-sm text-fg2">常用平台有现成方案：</span>
        <router-link
          v-for="p in platforms"
          :key="p.hash"
          :to="{ name: 'doc', hash: p.hash }"
          class="chip hover:border-fg/30 hover:text-fg transition-colors"
        >
          {{ p.label }}
        </router-link>
        <router-link
          :to="{ name: 'doc' }"
          class="link text-sm font-600 inline-flex items-center gap-1"
        >
          完整文档
          <span class="i-lucide-arrow-right text-sm" />
        </router-link>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import SectionTitle from '@/components/ui/SectionTitle.vue'
import CodeBlock from '@/components/ui/CodeBlock.vue'

const steps = [
  {
    title: '计算哈希',
    desc: '去除首尾空格并转为小写，计算 SHA256，也支持 MD5。',
    file: 'hash.js',
    lang: 'javascript',
    code: `const raw = ' Hi@WeAvatar.com '
const hash = sha256(raw.trim().toLowerCase())
// 2f9f6c…`
  },
  {
    title: '拼接地址',
    desc: '哈希拼接到地址末尾，按需追加尺寸等参数。',
    file: 'url',
    lang: 'text',
    code: `https://weavatar.com/avatar/{hash}?s=200&d=mp`
  },
  {
    title: '直接使用',
    desc: '像普通图片一样引用即可。',
    file: 'index.html',
    lang: 'html',
    code: `<img src="https://weavatar.com/avatar/{hash}?s=200" alt="avatar">`
  }
]

const platforms = [
  { label: 'WordPress', hash: '#wordpress' },
  { label: 'Typecho', hash: '#typecho' },
  { label: 'Emlog', hash: '#emlog' },
  { label: 'Z-Blog', hash: '#zblog' },
  { label: 'Twikoo', hash: '#twikoo' },
  { label: 'Artalk', hash: '#artalk' }
]
</script>
