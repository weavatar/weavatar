<template>
  <button
    type="button"
    class="inline-flex items-center gap-1.5 rounded-lg text-fg3 hover:text-fg hover:bg-fg/6 transition-colors"
    :class="label ? 'h-8 px-2.5 text-xs font-500' : 'h-8 w-8 justify-center'"
    :title="copied ? '已复制' : '复制'"
    :aria-label="copied ? '已复制' : '复制'"
    @click.stop="handleCopy"
  >
    <span
      class="text-[15px] transition-transform duration-300"
      :class="copied ? 'i-lucide-check text-green-500 scale-110' : 'i-lucide-copy'"
    />
    <span v-if="label">{{ copied ? '已复制' : label }}</span>
  </button>
</template>

<script setup lang="ts">
import { useClipboard } from '@vueuse/core'

const props = defineProps<{ text: string; label?: string }>()

const { copy, copied, isSupported } = useClipboard({ copiedDuring: 1600 })

const handleCopy = async () => {
  if (!isSupported.value) {
    window.$message.warning('当前浏览器不支持自动复制，请手动选择复制')
    return
  }
  await copy(props.text)
}
</script>
