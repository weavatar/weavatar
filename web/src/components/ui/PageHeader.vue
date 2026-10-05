<template>
  <div :class="bordered ? 'border-b border-line' : ''">
    <div class="wrap pt-14 pb-10 sm:pt-20 sm:pb-12">
      <div class="flex flex-wrap items-end justify-between gap-6">
        <div class="max-w-2xl">
          <h1 class="display-text text-[2rem] sm:text-[2.75rem] text-fg">{{ title }}</h1>
          <p v-if="desc" class="mt-3 text-base sm:text-[17px] text-fg2 leading-relaxed">
            {{ desc }}
          </p>
        </div>
        <div v-if="$slots.actions" class="shrink-0">
          <slot name="actions" />
        </div>
      </div>
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string
    desc?: string
    bordered?: boolean
  }>(),
  { bordered: true }
)
</script>
