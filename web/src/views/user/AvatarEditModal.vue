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
        :disabled="submitLoading"
        @click="handleSubmit"
      >
        <span v-if="submitLoading" class="i-lucide-loader-circle animate-spin text-base" />
        {{ submitLoading ? '正在保存…' : '保存修改' }}
      </button>
    </div>
  </n-modal>

  <geetest-captcha :config="{ product: 'bind' }" @initialized="onCaptchaInit" />
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { NModal } from 'naive-ui'
import { GeetestCaptcha } from 'vue3-geetest'
import { useRequest } from 'alova/client'
import StepLabel from '@/components/ui/StepLabel.vue'
import AvatarSourcePicker from '@/components/avatar/AvatarSourcePicker.vue'
import avatarApi from '@/api/avatar'

const props = defineProps<{ show: boolean; hash: string }>()
const emit = defineEmits<{
  'update:show': [value: boolean]
  success: []
}>()

const avatarBlob = ref<Blob | null>(null)
const submitLoading = ref(false)

watch(
  () => props.show,
  (v) => {
    if (!v) avatarBlob.value = null
  }
)

// 极验
let captchaInstance: any = null
const onCaptchaInit = (instance: any) => {
  captchaInstance = instance
  captchaInstance.onError((e: any) => window.$message.error(e.msg))
  captchaInstance.onSuccess(() => doSubmit(captchaInstance.getValidate()))
}

const handleSubmit = () => {
  if (!avatarBlob.value?.size) return window.$message.error('请先选择头像')
  if (!props.hash) return window.$message.error('头像哈希为空')
  captchaInstance?.showCaptcha()
}

const doSubmit = (captchaValidation: any) => {
  submitLoading.value = true
  const formData = new FormData()
  formData.append('avatar', avatarBlob.value!, 'avatar.png')
  formData.append('captcha', JSON.stringify(captchaValidation))

  useRequest(avatarApi.update(props.hash, formData))
    .onSuccess(() => {
      window.$message.success('修改成功，3 小时内全网生效')
      emit('update:show', false)
      emit('success')
      avatarBlob.value = null
    })
    .onComplete(() => {
      submitLoading.value = false
    })
}
</script>
