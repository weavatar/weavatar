<template>
  <div class="flex items-start gap-3">
    <span
      class="mt-0.5 w-6 h-6 shrink-0 rounded-full bg-brand/10 text-brand text-xs font-700 font-mono flex items-center justify-center tabular"
    >
      {{ index }}
    </span>
    <div>
      <div class="text-sm font-700 text-fg">{{ title }}</div>
      <div v-if="desc" class="mt-0.5 text-xs text-fg3 leading-relaxed">{{ desc }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ index: number; title: string; desc?: string }>()
</script>
