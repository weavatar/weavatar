<template>
  <div
    class="not-prose my-4 flex items-start gap-3 rounded-xl border px-4 py-3.5 text-sm leading-relaxed"
    :class="styles[type]"
  >
    <span :class="icons[type]" class="text-lg mt-0.5" />
    <div class="min-w-0 flex-1">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ type?: 'success' | 'info' | 'warning' }>(), { type: 'success' })

const styles = {
  success: 'bg-green-500/8 border-green-500/25 text-green-700 dark:text-green-300',
  info: 'bg-brand/8 border-brand/25 text-fg',
  warning: 'bg-amber-500/10 border-amber-500/30 text-amber-800 dark:text-amber-300'
}

const icons = {
  success: 'i-lucide-circle-check',
  info: 'i-lucide-info',
  warning: 'i-lucide-triangle-alert'
}
</script>
