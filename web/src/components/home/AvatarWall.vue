<template>
  <div class="relative h-[36rem] overflow-hidden wa-mask-y select-none" aria-hidden="true">
    <div
      class="absolute inset-0 flex justify-center gap-3.5"
      style="
        transform: perspective(1400px) rotateX(14deg) rotateY(-16deg) rotateZ(8deg) scale(1.08);
        transform-origin: 50% 50%;
      "
    >
      <div
        v-for="(col, ci) in columns"
        :key="ci"
        class="flex flex-col w-[4.75rem]"
        :class="ci % 2 === 0 ? 'animate-scroll-up' : 'animate-scroll-down'"
        :style="{ animationDuration: `${38 + ci * 6}s` }"
      >
        <div v-for="copy in 2" :key="copy" class="flex flex-col gap-3.5 pb-3.5">
          <div
            v-for="(url, i) in col"
            :key="`${copy}-${i}`"
            class="w-[4.75rem] h-[4.75rem] rounded-2xl overflow-hidden bg-muted ring-1 ring-line shadow-card transition-transform duration-300 hover:scale-105"
          >
            <img
              :src="url"
              alt=""
              loading="lazy"
              decoding="async"
              class="w-full h-full object-cover"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { fillAvatars } from './placeholders'

const props = defineProps<{ avatars: string[] }>()

const COLS = 5
const ROWS = 9

const columns = computed(() => {
  const list = fillAvatars(props.avatars, COLS * ROWS)
  return Array.from({ length: COLS }, (_, c) => list.slice(c * ROWS, (c + 1) * ROWS))
})
</script>
