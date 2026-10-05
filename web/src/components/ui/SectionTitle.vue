<template>
  <div :class="align === 'center' ? 'text-center mx-auto max-w-2xl' : 'max-w-2xl'">
    <h2 class="display-text text-[1.9rem] sm:text-[2.4rem] text-fg">
      <slot name="title">{{ title }}</slot>
    </h2>
    <p v-if="desc || $slots.desc" class="mt-4 text-base sm:text-[17px] text-fg2 leading-relaxed">
      <slot name="desc">{{ desc }}</slot>
    </p>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ title?: string; desc?: string; align?: 'center' | 'left' }>(), {
  align: 'center'
})
</script>
