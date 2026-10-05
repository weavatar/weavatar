<template>
  <div
    class="relative overflow-hidden bg-muted shrink-0"
    :class="roundedMap[rounded]"
    :style="{ width: sizeStyle, height: sizeStyle }"
  >
    <img
      :src="src"
      :alt="alt"
      loading="lazy"
      decoding="async"
      draggable="false"
      class="w-full h-full object-cover transition-opacity duration-500"
      :class="loaded ? 'opacity-100' : 'opacity-0'"
      @load="loaded = true"
      @error="failed = true"
    />
    <div
      v-if="failed"
      class="absolute inset-0 flex items-center justify-center text-fg3"
      :style="{ fontSize: `calc(${sizeStyle} * 0.45)` }"
    >
      <span class="i-lucide-user-round" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    src: string
    alt?: string
    size?: number | string
    rounded?: 'full' | '2xl' | 'xl' | 'lg'
  }>(),
  { alt: '', size: 56, rounded: 'full' }
)

const roundedMap = {
  full: 'rounded-full',
  '2xl': 'rounded-2xl',
  xl: 'rounded-xl',
  lg: 'rounded-lg'
} as const

const loaded = ref(false)
const failed = ref(false)

watch(
  () => props.src,
  () => {
    loaded.value = false
    failed.value = false
  }
)

const sizeStyle = computed(() => (typeof props.size === 'number' ? `${props.size}px` : props.size))
</script>
