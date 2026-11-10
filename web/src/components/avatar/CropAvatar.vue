<template>
  <n-modal
    :show="showModal"
    preset="card"
    title="裁剪头像"
    :style="{ width: '560px', maxWidth: '94vw' }"
    :bordered="false"
    :mask-closable="false"
    :auto-focus="false"
    @update:show="setShow"
    @after-enter="initCropper"
  >
    <div ref="containerRef" class="cropper-wrapper h-[22rem] rounded-xl overflow-hidden bg-muted" />
    <div class="mt-4 flex items-center justify-between gap-3">
      <span class="text-xs text-fg3">输出 800 × 800 正方形。</span>
      <div class="flex gap-2 shrink-0">
        <button type="button" class="btn-ghost btn-sm" @click="setShow(false)">取消</button>
        <button type="button" class="btn-primary btn-sm" :disabled="loading" @click="handleConfirm">
          <span v-if="loading" class="i-lucide-loader-circle animate-spin text-sm" />
          <span v-else class="i-lucide-check text-sm" />
          使用这张
        </button>
      </div>
    </div>
  </n-modal>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue'
import { NModal } from 'naive-ui'
import Cropper, { type CropperSelection } from 'cropperjs'

type Box = { x: number; y: number; width: number; height: number }

const emit = defineEmits<{ cropAvatar: [blob: Blob] }>()

const showModal = ref(false)
const loading = ref(false)
const containerRef = ref<HTMLElement>()
const imgSrc = ref('')
let cropper: Cropper | null = null

const destroyCropper = () => {
  if (cropper) {
    cropper.destroy()
    cropper = null
  }
  if (containerRef.value) {
    containerRef.value.innerHTML = ''
  }
}

const initCropper = () => {
  destroyCropper()
  if (!containerRef.value || !imgSrc.value) return

  const image = new Image()
  image.src = imgSrc.value
  image.alt = 'Avatar'

  cropper = new Cropper(image, {
    container: containerRef.value,
    template: `
      <cropper-canvas background>
        <cropper-image initial-center-size="contain" rotatable scalable translatable></cropper-image>
        <cropper-shade hidden></cropper-shade>
        <cropper-handle action="move" plain></cropper-handle>
        <cropper-selection aspect-ratio="1" movable resizable>
          <cropper-grid role="grid" bordered covered></cropper-grid>
          <cropper-crosshair centered></cropper-crosshair>
          <cropper-handle action="move" theme-color="rgba(255, 255, 255, 0.35)"></cropper-handle>
          <cropper-handle action="n-resize"></cropper-handle>
          <cropper-handle action="e-resize"></cropper-handle>
          <cropper-handle action="s-resize"></cropper-handle>
          <cropper-handle action="w-resize"></cropper-handle>
          <cropper-handle action="ne-resize"></cropper-handle>
          <cropper-handle action="nw-resize"></cropper-handle>
          <cropper-handle action="se-resize"></cropper-handle>
          <cropper-handle action="sw-resize"></cropper-handle>
        </cropper-selection>
      </cropper-canvas>
    `
  })

  const localCropper = cropper
  const cropperImage = localCropper.getCropperImage()
  const selection = localCropper.getCropperSelection()
  if (!cropperImage || !selection) return

  // 图片载入并居中后，初始选区取图片短边的正方形并居中
  cropperImage.$ready().then(() => {
    if (localCropper !== cropper) return
    const bounds = visibleBounds(localCropper)
    if (!bounds) return
    const size = Math.min(bounds.width, bounds.height)
    applySelection(selection, {
      x: bounds.x + (bounds.width - size) / 2,
      y: bounds.y + (bounds.height - size) / 2,
      width: size,
      height: size
    })
  })

  // 选区只能在图片可见范围内移动和缩放
  selection.addEventListener('change', (event) => {
    const e = event as CustomEvent<Box>
    const fitted = fit(e.detail, visibleBounds(localCropper))
    if (fitted && !same(fitted, e.detail)) {
      e.preventDefault()
      applySelection(selection, fitted)
    }
  })

  // 滚轮缩放或拖动图片后把选区压回图片内，否则输出会带透明区域
  cropperImage.addEventListener('change', (event) => {
    const e = event as CustomEvent<Box>
    const fitted = fit(selection, visibleBounds(localCropper, e.detail))
    if (fitted && !same(fitted, selection)) applySelection(selection, fitted)
  })
}

const handleConfirm = async () => {
  if (!cropper) return
  loading.value = true
  try {
    const selection = cropper.getCropperSelection()
    if (!selection) {
      loading.value = false
      return
    }
    const canvas = await selection.$toCanvas({ width: 800, height: 800 })
    canvas.toBlob((blob) => {
      if (blob) {
        emit('cropAvatar', blob)
        showModal.value = false
        destroyCropper()
      }
      loading.value = false
    }, 'image/png')
  } catch {
    loading.value = false
  }
}

const setShow = (value: boolean) => {
  showModal.value = value
  if (!value) {
    destroyCropper()
  }
}

const setImage = (value: Blob) => {
  if (imgSrc.value) URL.revokeObjectURL(imgSrc.value)
  imgSrc.value = URL.createObjectURL(value)
}

onBeforeUnmount(() => {
  destroyCropper()
  if (imgSrc.value) URL.revokeObjectURL(imgSrc.value)
})

defineExpose({ setShow, setImage })

// 图片与画布的重叠区域（画布坐标系），不传 image 时取图片当前位置
const visibleBounds = (c: Cropper, image?: Box): Box | null => {
  const canvasEl = c.getCropperCanvas()
  const imageEl = c.getCropperImage()
  if (!canvasEl || !imageEl) return null
  const canvasRect = canvasEl.getBoundingClientRect()
  if (!image) {
    const rect = imageEl.getBoundingClientRect()
    image = {
      x: rect.left - canvasRect.left,
      y: rect.top - canvasRect.top,
      width: rect.width,
      height: rect.height
    }
  }
  const x = Math.max(0, image.x)
  const y = Math.max(0, image.y)
  return {
    x,
    y,
    width: Math.min(canvasRect.width, image.x + image.width) - x,
    height: Math.min(canvasRect.height, image.y + image.height) - y
  }
}

// 把正方形选区压进图片区域，图片比选区小时选区缩成图片短边
const fit = (box: Box, bounds: Box | null): Box | null => {
  if (!bounds || bounds.width <= 0 || bounds.height <= 0) return null
  const size = Math.min(box.width, box.height, bounds.width, bounds.height)
  return {
    x: Math.max(bounds.x, Math.min(box.x, bounds.x + bounds.width - size)),
    y: Math.max(bounds.y, Math.min(box.y, bounds.y + bounds.height - size)),
    width: size,
    height: size
  }
}

const same = (a: Box, b: Box) =>
  a.x === b.x && a.y === b.y && a.width === b.width && a.height === b.height

const applySelection = (selection: CropperSelection, box: Box) => {
  selection.x = box.x
  selection.y = box.y
  selection.width = box.width
  selection.height = box.height
  selection.$render()
}
</script>

<style scoped>
.cropper-wrapper :deep(cropper-canvas) {
  height: 100%;
}
</style>
