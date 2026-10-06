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
import Cropper from 'cropperjs'

const emit = defineEmits<{ cropAvatar: [blob: Blob] }>()

const showModal = ref(false)
const loading = ref(false)
const containerRef = ref<HTMLElement>()
const imgSrc = ref('')
let cropper: Cropper | null = null

// 图片在 canvas 中的实际区域，选区不能超出它
let imgBounds = { x: 0, y: 0, w: 0, h: 0 }

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
  if (cropperImage) {
    cropperImage.$ready().then(() => {
      if (!localCropper || localCropper !== cropper) return

      const canvasEl = localCropper.getCropperCanvas()
      if (!canvasEl) return

      const imgRect = cropperImage.getBoundingClientRect()
      const canvasRect = canvasEl.getBoundingClientRect()

      const imgX = imgRect.left - canvasRect.left
      const imgY = imgRect.top - canvasRect.top
      const imgW = imgRect.width
      const imgH = imgRect.height

      imgBounds = { x: imgX, y: imgY, w: imgW, h: imgH }

      // 初始选区取图片短边的正方形并居中
      const size = Math.min(imgW, imgH)
      const sel = localCropper.getCropperSelection()
      if (sel) {
        sel.x = imgX + (imgW - size) / 2
        sel.y = imgY + (imgH - size) / 2
        sel.width = size
        sel.height = size
        sel.$render()
      }
    })
  }

  const selection = localCropper.getCropperSelection()
  if (selection) {
    selection.addEventListener('change', (event: Event) => {
      const e = event as CustomEvent
      const { x, y, width, height } = e.detail
      const { x: bx, y: by, w: bw, h: bh } = imgBounds

      if (bw === 0 || bh === 0) return

      const maxSize = Math.min(bw, bh)
      const clampedW = Math.min(width, maxSize)
      const clampedH = Math.min(height, maxSize)
      const clampedX = Math.max(bx, Math.min(x, bx + bw - clampedW))
      const clampedY = Math.max(by, Math.min(y, by + bh - clampedH))

      if (x !== clampedX || y !== clampedY || width !== clampedW || height !== clampedH) {
        e.preventDefault()
        selection.x = clampedX
        selection.y = clampedY
        selection.width = clampedW
        selection.height = clampedH
        selection.$render()
      }
    })
  }
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
</script>

<style scoped>
.cropper-wrapper :deep(cropper-canvas) {
  height: 100%;
}
</style>
