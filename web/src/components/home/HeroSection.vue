<template>
  <section class="relative overflow-x-clip">
    <div
      class="absolute inset-0 -top-16 wa-grid wa-mask-radial pointer-events-none"
      aria-hidden="true"
    />
    <div
      class="absolute -top-48 left-1/2 -translate-x-1/2 w-[64rem] h-[32rem] rounded-full bg-brand/14 blur-[140px] pointer-events-none"
      aria-hidden="true"
    />

    <div
      class="relative wrap grid items-center gap-12 pt-16 pb-14 lg:grid-cols-[1.05fr_1fr] lg:gap-10 lg:py-0 lg:min-h-[calc(100vh-4rem)] lg:max-h-[56rem]"
    >
      <div class="max-w-2xl">
        <div v-reveal class="eyebrow mb-6">
          <span class="relative flex w-2 h-2">
            <span class="absolute inset-0 rounded-full bg-brand animate-pulse-ring" />
            <span class="relative w-2 h-2 rounded-full bg-brand" />
          </span>
          <span>新一代头像服务 · 完整兼容 Gravatar</span>
        </div>

        <h1
          v-reveal="60"
          class="display-text text-[2.75rem] sm:text-[3.75rem] lg:text-[4.5rem] text-fg"
        >
          你的头像，<br />
          <span class="text-brand">随处可见</span>。
        </h1>

        <p v-reveal="120" class="mt-6 text-base sm:text-lg text-fg2 leading-relaxed max-w-xl">
          将邮箱或手机号变成您的数字护照，<br
            class="hidden sm:block"
          />您在互联网上发帖、评论或在线互动时均可使用。
        </p>

        <div v-reveal="180" class="mt-8 flex flex-wrap items-center gap-3">
          <router-link :to="{ name: 'login' }" class="btn-primary btn-lg">
            开始使用
            <span class="i-lucide-arrow-right text-base" />
          </router-link>
          <router-link :to="{ name: 'doc' }" class="btn-secondary btn-lg">查看文档</router-link>
        </div>

        <div
          v-reveal="240"
          class="mt-10 flex flex-wrap items-center gap-x-8 gap-y-3 text-sm text-fg2"
        >
          <div v-if="usage > 0" class="flex items-baseline gap-1.5">
            <span>昨日响应</span>
            <span class="text-2xl font-800 text-fg tabular tracking-tight">{{ formatted }}</span>
            <span>次请求</span>
          </div>
        </div>
      </div>

      <div v-reveal="200" class="hidden lg:block">
        <avatar-wall :avatars="avatars" />
      </div>
    </div>

    <div class="relative lg:hidden pb-14 overflow-hidden wa-mask-x">
      <div class="flex w-max animate-marquee">
        <div v-for="copy in 2" :key="copy" class="flex gap-3 pr-3">
          <avatar-image
            v-for="(url, i) in mobileList"
            :key="`${copy}-${i}`"
            :src="url"
            :size="56"
            rounded="xl"
            class="ring-1 ring-line"
          />
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { TransitionPresets, useTransition } from '@vueuse/core'
import AvatarWall from './AvatarWall.vue'
import AvatarImage from '@/components/ui/AvatarImage.vue'
import { formatNumber } from '@/utils/format'
import { placeholderAvatars } from './placeholders'

const props = defineProps<{
  usage: number
  avatars: string[]
}>()

const source = ref(0)
watch(
  () => props.usage,
  (v) => {
    source.value = v
  }
)
const animated = useTransition(source, {
  duration: 2400,
  transition: TransitionPresets.easeOutExpo
})
const formatted = computed(() => formatNumber(animated.value))

const mobileList = computed(() =>
  props.avatars.length ? props.avatars.slice(0, 20) : placeholderAvatars(20)
)
</script>
