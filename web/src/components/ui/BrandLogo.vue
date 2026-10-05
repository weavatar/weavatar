<template>
  <img
    :src="isDark ? logoWhite : logo"
    alt="WeAvatar"
    :style="{ height: `${height}px`, width: 'auto' }"
    class="select-none"
    draggable="false"
  />
</template>

<script setup lang="ts">
import logo from '@/assets/logo.png'
import logoWhite from '@/assets/logo-white.png'
import { useTheme } from '@/composables/useTheme'

withDefaults(defineProps<{ height?: number }>(), { height: 26 })

const { isDark } = useTheme()
</script>
