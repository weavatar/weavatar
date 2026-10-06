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
      <section>
        <step-label :index="1" title="要绑定的地址" desc="需要验证码确认归属。" />
        <div class="mt-4 space-y-3">
          <n-input
            v-model:value="form.raw"
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
              v-model:value="form.verify_code"
              size="large"
              placeholder="验证码"
              :input-props="{ autocomplete: 'one-time-code', inputmode: 'numeric' }"
              @keydown.enter.prevent
            >
              <template #prefix>
                <span class="i-lucide-shield-check text-fg3" />
              </template>
            </n-input>
            <verify-code-button :to="form.raw" use-for="avatar" />
          </div>
        </div>
      </section>

      <section>
        <step-label :index="2" title="选择头像" desc="上传图片，或直接获取社交头像。" />
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
        {{ submitting ? '正在提交…' : '添加头像' }}
      </button>
    </div>
  </n-modal>

  <geetest-captcha :config="{ product: 'bind' }" @initialized="onInit" />
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { NInput, NModal } from 'naive-ui'
import { GeetestCaptcha } from 'vue3-geetest'
import StepLabel from '@/components/ui/StepLabel.vue'
import VerifyCodeButton from '@/components/captcha/VerifyCodeButton.vue'
import AvatarSourcePicker from './AvatarSourcePicker.vue'
import { checkAvatar, createAvatar } from '@/api/avatar'
import { useGeetest } from '@/composables/useGeetest'
import { isPhoneOrEmail } from '@/utils/validate'

defineProps<{ show: boolean }>()
const emit = defineEmits<{
  'update:show': [value: boolean]
  success: []
}>()

const { onInit, verify } = useGeetest()

const form = reactive({ raw: '', verify_code: '' })
const avatarBlob = ref<Blob | null>(null)
const submitting = ref(false)

const reset = () => {
  form.raw = ''
  form.verify_code = ''
  avatarBlob.value = null
}

const confirmOverwrite = () =>
  new Promise<boolean>((resolve) => {
    window.$dialog.warning({
      title: '地址已被绑定',
      content: '该地址已被其他用户添加，继续添加将覆盖对方的头像。是否继续？',
      positiveText: '继续添加',
      negativeText: '取消',
      onPositiveClick: () => resolve(true),
      // 取消、关闭与点击遮罩都会触发，确认后再触发不影响结果
      onAfterLeave: () => resolve(false)
    })
  })

const handleSubmit = async () => {
  const raw = form.raw.trim()
  const code = form.verify_code.trim()
  const blob = avatarBlob.value
  if (!isPhoneOrEmail(raw)) return window.$message.error('请输入正确的手机号或邮箱')
  if (!code) return window.$message.error('请输入验证码')
  if (!blob?.size) return window.$message.error('请选择头像')

  try {
    const captcha = await verify()
    submitting.value = true
    const { bind } = await checkAvatar(raw)
    if (bind && !(await confirmOverwrite())) return
    const data = new FormData()
    data.append('raw', raw)
    data.append('avatar', blob, 'avatar.png')
    data.append('verify_code', code)
    data.append('captcha', JSON.stringify(captcha))
    await createAvatar(data)
    window.$message.success('添加成功，3 小时内全网生效')
    emit('update:show', false)
    emit('success')
    reset()
  } catch {
    // 取消验证或添加失败，错误已提示
  } finally {
    submitting.value = false
  }
}
</script>
