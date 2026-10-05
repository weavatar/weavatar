<template>
  <button
    type="button"
    class="btn-ghost btn-icon relative overflow-hidden"
    :aria-label="isDark ? '切换到浅色模式' : '切换到深色模式'"
    :title="isDark ? '浅色模式' : '深色模式'"
    @click="toggle($event)"
  >
    <span
      class="i-lucide-sun absolute text-lg transition-all duration-400"
      :class="isDark ? 'rotate-0 scale-100 opacity-100' : 'rotate-90 scale-50 opacity-0'"
    />
    <span
      class="i-lucide-moon absolute text-lg transition-all duration-400"
      :class="isDark ? '-rotate-90 scale-50 opacity-0' : 'rotate-0 scale-100 opacity-100'"
    />
  </button>
</template>

<script setup lang="ts">
import { useTheme } from '@/composables/useTheme'

const { isDark, toggle } = useTheme()
</script>
