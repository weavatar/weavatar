<template>
  <div>
    <!-- 已选择：预览 -->
    <div v-if="blob && previewUrl" class="surface p-4 flex items-center gap-4">
      <avatar-image
        :src="previewUrl"
        :size="72"
        rounded="2xl"
        alt="头像预览"
        class="ring-1 ring-line"
      />
      <div class="min-w-0 flex-1">
        <div class="text-sm font-600 text-fg">已裁剪完成</div>
        <div class="mt-0.5 text-xs text-fg3">800 × 800 · PNG · {{ sizeText }}</div>
      </div>
      <button type="button" class="btn-secondary btn-sm" @click="emit('update:blob', null)">
        重新选择
      </button>
    </div>

    <!-- 未选择：上传 / 社交头像 -->
    <div v-else>
      <div class="flex gap-1 p-1 rounded-xl bg-muted border border-line w-fit">
        <button
          v-for="t in tabs"
          :key="t.key"
          type="button"
          class="h-8 px-3.5 rounded-lg text-[13px] font-600 transition-colors"
          :class="tab === t.key ? 'bg-elev text-fg shadow-sm' : 'text-fg2 hover:text-fg'"
          @click="tab = t.key"
        >
          {{ t.label }}
        </button>
      </div>

      <!-- 上传 -->
      <label
        v-if="tab === 'upload'"
        class="mt-3 block rounded-2xl border border-dashed p-8 text-center cursor-pointer transition-colors"
        :class="
          dragging
            ? 'border-brand bg-brand/6'
            : 'border-line-strong hover:border-brand hover:bg-brand/4'
        "
        @dragover.prevent="dragging = true"
        @dragleave.prevent="dragging = false"
        @drop.prevent="onDrop"
      >
        <input
          type="file"
          accept="image/jpeg,image/png,image/gif,image/webp"
          class="sr-only"
          @change="onFileChange"
        />
        <div
          class="mx-auto w-12 h-12 rounded-xl bg-brand/10 text-brand flex items-center justify-center"
        >
          <span class="i-lucide-upload text-2xl" />
        </div>
        <div class="mt-4 text-sm font-600 text-fg">点击或拖拽图片到这里</div>
        <div class="mt-1 text-xs text-fg3">支持 JPG / PNG / GIF / WebP，不超过 5 MB</div>
        <div class="mt-3 text-xs text-fg3">上传的图片需符合中华人民共和国相关法律法规要求</div>
      </label>

      <!-- 社交头像 -->
      <div v-else class="mt-3 surface p-4">
        <div class="flex gap-2">
          <n-input
            v-model:value="qq"
            size="large"
            placeholder="社交账号"
            :input-props="{ inputmode: 'numeric', autocomplete: 'off' }"
            @keydown.enter.prevent="fetchQq"
          >
            <template #prefix>
              <span class="i-lucide-message-circle text-fg3" />
            </template>
          </n-input>
          <button
            type="button"
            class="btn-primary h-11 shrink-0"
            :disabled="qqLoading"
            @click="fetchQq"
          >
            <span v-if="qqLoading" class="i-lucide-loader-circle animate-spin text-base" />
            一键获取
          </button>
        </div>
      </div>
    </div>

    <crop-avatar ref="cropRef" @crop-avatar="onCropped" />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { NInput } from 'naive-ui'
import { useRequest } from 'alova/client'
import AvatarImage from '@/components/ui/AvatarImage.vue'
import CropAvatar from './CropAvatar.vue'
import avatarApi from '@/api/avatar'

const props = defineProps<{ blob: Blob | null }>()
const emit = defineEmits<{ 'update:blob': [value: Blob | null] }>()

const tabs = [
  { key: 'upload', label: '上传图片' },
  { key: 'qq', label: '社交头像' }
] as const
const tab = ref<'upload' | 'qq'>('upload')
const dragging = ref(false)
const qq = ref('')
const qqLoading = ref(false)
const cropRef = ref<InstanceType<typeof CropAvatar>>()

/* 预览 URL */
const previewUrl = ref('')
watch(
  () => props.blob,
  (b) => {
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = b ? URL.createObjectURL(b) : ''
  },
  { immediate: true }
)
onBeforeUnmount(() => {
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
})

const sizeText = computed(() => {
  const size = props.blob?.size ?? 0
  return size > 1024 * 1024
    ? `${(size / 1024 / 1024).toFixed(2)} MB`
    : `${Math.round(size / 1024)} KB`
})

/* 校验 + 打开裁剪 */
const accept = (file: File) => {
  const validType = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'].includes(file.type)
  const validSize = file.size / 1024 / 1024 < 5
  if (!validType) window.$message.error('只能上传 JPG / PNG / GIF / WebP 格式')
  else if (!validSize) window.$message.error('图片大小不能超过 5 MB')
  if (!validType || !validSize) return
  cropRef.value?.setImage(file)
  cropRef.value?.setShow(true)
}

const onFileChange = (e: Event) => {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) accept(file)
  input.value = ''
}

const onDrop = (e: DragEvent) => {
  dragging.value = false
  const file = e.dataTransfer?.files?.[0]
  if (file) accept(file)
}

const onCropped = (b: Blob) => emit('update:blob', b)

/* 社交头像 */
const fetchQq = () => {
  const value = qq.value.trim()
  if (!/^\d{5,12}$/.test(value)) {
    window.$message.error('请输入正确的社交账号')
    return
  }
  qqLoading.value = true
  useRequest(avatarApi.qq(value))
    .onSuccess(({ data }: any) => {
      const binary = atob(data)
      const bytes = new Uint8Array(binary.length)
      for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
      cropRef.value?.setImage(new Blob([bytes], { type: 'image/png' }))
      cropRef.value?.setShow(true)
    })
    .onComplete(() => {
      qqLoading.value = false
    })
}
</script>
