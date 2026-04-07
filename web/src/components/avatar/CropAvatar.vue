<template>
  <NModal :show="showModal" @after-enter="initCropper">
    <NCard
      closable
      :mask-closable="false"
      title="裁剪头像"
      class="w-full max-w-125"
      @close="showModal = false"
    >
      <div ref="containerRef" class="cropper-wrapper h-90" />
      <div class="mt-4 flex justify-center">
        <NButton type="primary" size="large" :loading="loading" @click="handleConfirm">确定</NButton>
      </div>
    </NCard>
  </NModal>
</template>

<script setup lang="ts">
import { ref, onBeforeUnmount } from 'vue'
import { NButton, NCard, NModal } from 'naive-ui'
import Cropper from 'cropperjs'

const showModal = ref(false)
const loading = ref(false)
const containerRef = ref<HTMLElement>()
const imgSrc = ref('')
let cropper: Cropper | null = null

// 记录图片在 canvas 中的实际区域，用于限制选区
let imgBounds = { x: 0, y: 0, w: 0, h: 0 }

const emit = defineEmits(['cropAvatar'])

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

  // 图片加载后：计算图片区域 + 初始选区
  const cropperImage = localCropper.getCropperImage()
  if (cropperImage) {
    (cropperImage as any).$ready().then(() => {
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

      // 选区设为图片短边的正方形并居中在图片上
      const size = Math.min(imgW, imgH)
      const sel = localCropper.getCropperSelection()
      if (sel) {
        sel.x = imgX + (imgW - size) / 2
        sel.y = imgY + (imgH - size) / 2
        sel.width = size
        sel.height = size
        ;(sel as any).$render()
      }
    })
  }

  // 限制选区不超出图片实际区域
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
        ;(selection as any).$render()
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
    const canvas = await (selection as any).$toCanvas({ width: 800, height: 800 })
    canvas.toBlob((blob: Blob | null) => {
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
