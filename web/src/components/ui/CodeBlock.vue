<template>
  <div class="surface overflow-hidden text-[13px] flex flex-col">
    <div
      class="flex items-center justify-between h-10 pl-4 pr-2 border-b border-line bg-muted/60 shrink-0"
    >
      <div class="flex items-center gap-2 min-w-0">
        <span class="flex gap-1.5">
          <i class="w-2.5 h-2.5 rounded-full bg-fg/12" />
          <i class="w-2.5 h-2.5 rounded-full bg-fg/12" />
          <i class="w-2.5 h-2.5 rounded-full bg-fg/12" />
        </span>
        <span class="ml-2 text-xs font-500 text-fg3 font-mono truncate">{{ title || lang }}</span>
      </div>
      <copy-button :text="trimmed" />
    </div>
    <pre
      class="flex-1 p-4 leading-[1.7] m-0"
      :class="wrap ? 'whitespace-pre-wrap break-all' : 'overflow-x-auto'"
    ><code class="hljs font-mono" v-html="html" /></pre>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { highlight } from '@/utils/hljs'
import CopyButton from './CopyButton.vue'

const props = withDefaults(
  defineProps<{ code: string; lang?: string; title?: string; wrap?: boolean }>(),
  { lang: 'text', wrap: false }
)

const trimmed = computed(() => props.code.trim())
const html = computed(() => highlight(trimmed.value, props.lang))
</script>
