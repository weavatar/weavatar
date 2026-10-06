<template>
  <n-modal
    :show="show"
    preset="card"
    title="修改头像图片"
    :style="{ width: '520px', maxWidth: '94vw' }"
    :bordered="false"
    :mask-closable="false"
    :auto-focus="false"
    @update:show="$emit('update:show', $event)"
  >
    <div class="space-y-7">
      <section>
        <step-label :index="1" title="选择新头像" desc="上传图片，或直接获取社交头像。" />
        <div class="mt-4">
          <avatar-source-picker v-model:blob="avatarBlob" />
        </div>
      </section>

      <button
        type="button"
        class="btn-primary btn-lg w-full"
        :disabled="submitting"
        @click="handleSubmit"
      >
        <span v-if="submitting" class="i-lucide-loader-circle animate-spin text-base" />
        {{ submitting ? '正在保存…' : '保存修改' }}
      </button>
    </div>
  </n-modal>

  <geetest-captcha :config="{ product: 'bind' }" @initialized="onInit" />
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { NModal } from 'naive-ui'
import { GeetestCaptcha } from 'vue3-geetest'
import StepLabel from '@/components/ui/StepLabel.vue'
import AvatarSourcePicker from './AvatarSourcePicker.vue'
import { updateAvatar } from '@/api/avatar'
import { useGeetest } from '@/composables/useGeetest'

const props = defineProps<{ show: boolean; hash: string }>()
const emit = defineEmits<{
  'update:show': [value: boolean]
  success: []
}>()

const { onInit, verify } = useGeetest()

const avatarBlob = ref<Blob | null>(null)
const submitting = ref(false)

watch(
  () => props.show,
  (v) => {
    if (!v) avatarBlob.value = null
  }
)

const handleSubmit = async () => {
  const blob = avatarBlob.value
  if (!blob?.size) return window.$message.error('请选择头像')

  try {
    const captcha = await verify()
    submitting.value = true
    const data = new FormData()
    data.append('avatar', blob, 'avatar.png')
    data.append('captcha', JSON.stringify(captcha))
    await updateAvatar(props.hash, data)
    window.$message.success('修改成功，3 小时内全网生效')
    emit('update:show', false)
    emit('success')
  } catch {
    // 取消验证或修改失败，错误已提示
  } finally {
    submitting.value = false
  }
}
</script>
