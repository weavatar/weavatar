<template>
  <div>
    <div class="w-10 h-10 rounded-xl bg-brand/10 text-brand flex items-center justify-center">
      <span :class="icon" class="text-xl" />
    </div>
    <h3 class="mt-5 text-[17px] font-700 text-fg tracking-tight">{{ title }}</h3>
    <p class="mt-2 text-sm text-fg2 leading-relaxed">
      <slot />
    </p>
  </div>
</template>

<script setup lang="ts">
defineProps<{ icon: string; title: string }>()
</script>
