<script setup lang="ts">
import type { NumberAnimationInst } from 'naive-ui'
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NNumberAnimation, NAvatar } from 'naive-ui'

const props = defineProps<{
  usage: number
  avatars: string[]
}>()

const usageRef = defineModel<NumberAnimationInst | null>('usageRef')
const router = useRouter()

const column1 = computed(() => props.avatars.slice(0, 10))
const column2 = computed(() => props.avatars.slice(10, 20))
const column3 = computed(() => props.avatars.slice(20, 30))
</script>

<template>
  <section class="relative overflow-hidden">
    <!-- 背景光晕 -->
    <div class="hero-glow" />

    <div
      class="max-w-360 mx-auto grid grid-cols-1 lg:grid-cols-[1fr_auto] items-center gap-12 px-6 py-20 lg:py-0 lg:min-h-[calc(100vh-4rem)]"
    >
      <!-- 左侧文字 -->
      <div class="text-center lg:text-left">
        <div
          class="inline-block text-sm font-500 text-blue-500 bg-blue-500/8 border border-blue-500/15 rounded-full px-3.5 py-1 mb-6"
        >
          新一代头像服务
        </div>
        <h1 class="text-4xl sm:text-5xl lg:text-6xl font-800 leading-tight tracking-tight mb-5">
          您的免费网络资料卡
        </h1>
        <p class="text-sm lg:text-base leading-relaxed opacity-80 mb-2">
          将电子邮箱或手机号变成您的数字护照，<br class="hidden sm:block" />
          您在互联网上发帖、评论或在线互动时均可使用。
        </p>
        <p class="text-base opacity-60 mb-8">一次设置，随处可见。</p>
        <div class="flex gap-3 mb-8 justify-center lg:justify-start">
          <NButton type="info" size="large" @click="router.push({ name: 'login' })"
            >开始使用</NButton
          >
          <NButton size="large" @click="router.push({ name: 'doc' })">浏览文档</NButton>
        </div>
        <div class="flex items-baseline gap-1.5 text-sm justify-center lg:justify-start">
          <span>昨日共响应</span>
          <span class="text-2xl font-700 text-blue-500 tabular-nums">
            <NNumberAnimation
              ref="usageRef"
              :from="0"
              :to="usage"
              :active="false"
              :duration="3000"
              show-separator
            />
          </span>
          <span>次请求</span>
        </div>
      </div>

      <!-- 右侧头像滚动列 -->
      <div v-if="avatars.length >= 30" class="avatar-area hidden lg:block">
        <div class="flex gap-3.5 h-full">
          <div class="w-14 h-full overflow-hidden">
            <div class="scroll-track scroll-up flex flex-col gap-3.5">
              <template v-for="c in 2" :key="c">
                <NAvatar
                  v-for="(url, i) in column1"
                  :key="`${c}-${i}`"
                  :src="url"
                  round
                  :size="56"
                  class="shrink-0 shadow-sm transition-transform duration-300 hover:scale-112"
                />
              </template>
            </div>
          </div>
          <div class="w-14 h-full overflow-hidden">
            <div class="scroll-track scroll-down flex flex-col gap-3.5">
              <template v-for="c in 2" :key="c">
                <NAvatar
                  v-for="(url, i) in column2"
                  :key="`${c}-${i}`"
                  :src="url"
                  round
                  :size="56"
                  class="shrink-0 shadow-sm transition-transform duration-300 hover:scale-112"
                />
              </template>
            </div>
          </div>
          <div class="w-14 h-full overflow-hidden">
            <div class="scroll-track scroll-up-slow flex flex-col gap-3.5">
              <template v-for="c in 2" :key="c">
                <NAvatar
                  v-for="(url, i) in column3"
                  :key="`${c}-${i}`"
                  :src="url"
                  round
                  :size="56"
                  class="shrink-0 shadow-sm transition-transform duration-300 hover:scale-112"
                />
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.hero-glow {
  position: absolute;
  top: -12.5rem;
  left: 50%;
  transform: translateX(-50%);
  width: 56rem;
  height: 37.5rem;
  background: radial-gradient(ellipse at center, rgba(59, 130, 246, 0.08) 0%, transparent 70%);
  pointer-events: none;
}

.avatar-area {
  height: 32.5rem;
  overflow: hidden;
  mask-image: linear-gradient(to bottom, transparent 0%, black 12%, black 88%, transparent 100%);
}

.scroll-up {
  animation: scrollUp 28s linear infinite;
}
.scroll-down {
  animation: scrollDown 24s linear infinite;
}
.scroll-up-slow {
  animation: scrollUp 34s linear infinite;
}

@keyframes scrollUp {
  0% {
    transform: translateY(0);
  }
  100% {
    transform: translateY(calc(-50% - 0.44rem));
  }
}
@keyframes scrollDown {
  0% {
    transform: translateY(calc(-50% - 0.44rem));
  }
  100% {
    transform: translateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .scroll-up,
  .scroll-down,
  .scroll-up-slow {
    animation: none !important;
  }
}
</style>
