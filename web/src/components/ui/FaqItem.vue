<template>
  <div class="border-b border-line">
    <button
      type="button"
      class="w-full flex items-center justify-between gap-6 py-5 text-left"
      :aria-expanded="open"
      @click="open = !open"
    >
      <span class="text-base font-600 text-fg">{{ question }}</span>
      <span
        class="i-lucide-plus text-fg3 text-lg transition-transform duration-300 shrink-0"
        :class="open ? 'rotate-45 text-brand' : ''"
      />
    </button>
    <div
      class="grid transition-[grid-template-rows] duration-300 ease-out"
      :style="{ gridTemplateRows: open ? '1fr' : '0fr' }"
    >
      <div class="overflow-hidden">
        <div class="pb-6 pr-10 text-[15px] text-fg2 leading-relaxed">
          <slot />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{ question: string; defaultOpen?: boolean }>()
const open = ref(!!props.defaultOpen)
</script>
