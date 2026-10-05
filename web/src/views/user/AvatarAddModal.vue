<template>
  <n-modal
    :show="show"
    preset="card"
    title="添加头像"
    :style="{ width: '520px', maxWidth: '94vw' }"
    :bordered="false"
    :mask-closable="false"
    :auto-focus="false"
    @update:show="$emit('update:show', $event)"
  >
    <div class="space-y-7">
      <!-- 第一步：地址 -->
      <section>
        <step-label :index="1" title="要绑定的地址" desc="需要验证码确认归属。" />
        <div class="mt-4 space-y-3">
          <n-input
            v-model:value="model.raw"
            size="large"
            placeholder="邮箱 / 手机号"
            :input-props="{ autocomplete: 'off', spellcheck: false }"
            @keydown.enter.prevent
          >
            <template #prefix>
              <span class="i-lucide-at-sign text-fg3" />
            </template>
          </n-input>
          <div class="flex gap-2">
            <n-input
              v-model:value="model.verify_code"
              size="large"
              placeholder="验证码"
              :input-props="{ autocomplete: 'one-time-code', inputmode: 'numeric' }"
              @keydown.enter.prevent
            >
              <template #prefix>
                <span class="i-lucide-shield-check text-fg3" />
              </template>
            </n-input>
            <verify-code-button :to="model.raw" use-for="avatar" />
          </div>
        </div>
      </section>

      <!-- 第二步：头像 -->
      <section>
        <step-label :index="2" title="选择头像" desc="上传图片，或直接获取社交头像。" />
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
        {{ submitLoading ? '正在提交…' : '添加头像' }}
      </button>
    </div>
  </n-modal>

  <geetest-captcha :config="{ product: 'bind' }" @initialized="onCaptchaInit" />
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { NInput, NModal } from 'naive-ui'
import { GeetestCaptcha } from 'vue3-geetest'
import { useRequest } from 'alova/client'
import StepLabel from '@/components/ui/StepLabel.vue'
import AvatarSourcePicker from '@/components/avatar/AvatarSourcePicker.vue'
import VerifyCodeButton from '@/components/captcha/VerifyCodeButton.vue'
import avatarApi from '@/api/avatar'
import { isPhoneOrEmail } from '@/utils/validate'

defineProps<{ show: boolean }>()
const emit = defineEmits<{
  'update:show': [value: boolean]
  success: []
}>()

const model = ref({ raw: '', verify_code: '' })
const avatarBlob = ref<Blob | null>(null)
const submitLoading = ref(false)

// 极验
let captchaInstance: any = null
const onCaptchaInit = (instance: any) => {
  captchaInstance = instance
  captchaInstance.onError((e: any) => window.$message.error(e.msg))
  captchaInstance.onSuccess(() => doSubmit(captchaInstance.getValidate()))
}

const handleSubmit = () => {
  if (!model.value.raw) return window.$message.error('请先输入地址')
  if (!isPhoneOrEmail(model.value.raw)) return window.$message.error('请输入正确的手机号或邮箱')
  if (!model.value.verify_code) return window.$message.error('请先输入验证码')
  if (!avatarBlob.value?.size) return window.$message.error('请先选择头像')
  captchaInstance?.showCaptcha()
}

const reset = () => {
  model.value = { raw: '', verify_code: '' }
  avatarBlob.value = null
}

const doSubmit = (captchaValidation: any) => {
  submitLoading.value = true

  const submitAvatar = () => {
    const formData = new FormData()
    formData.append('raw', model.value.raw)
    formData.append('avatar', avatarBlob.value!, 'avatar.png')
    formData.append('verify_code', model.value.verify_code)
    formData.append('captcha', JSON.stringify(captchaValidation))
    useRequest(avatarApi.create(formData))
      .onSuccess(() => {
        window.$message.success('添加成功，3 小时内全网生效')
        emit('update:show', false)
        emit('success')
        reset()
      })
      .onComplete(() => {
        submitLoading.value = false
      })
  }

  // 先检查该地址是否已被其他用户绑定
  useRequest(avatarApi.check(model.value.raw))
    .onSuccess(({ data: res }: any) => {
      if (res.bind) {
        window.$dialog.warning({
          title: '地址已被绑定',
          content: '该地址已被其他用户添加，继续添加将覆盖对方的头像。是否继续？',
          positiveText: '继续添加',
          negativeText: '取消',
          onPositiveClick: submitAvatar,
          onNegativeClick: () => {
            submitLoading.value = false
          },
          onClose: () => {
            submitLoading.value = false
          }
        })
      } else {
        submitAvatar()
      }
    })
    .onError(() => {
      submitLoading.value = false
    })
}
</script>
