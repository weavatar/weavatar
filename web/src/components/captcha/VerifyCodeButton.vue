<template>
  <div class="shrink-0">
    <button
      type="button"
      class="btn-secondary h-11 px-4 text-sm tabular"
      :disabled="isActive || loading"
      @click="handleSend"
    >
      <span v-if="loading" class="i-lucide-loader-circle animate-spin text-sm" />
      {{ isActive ? `${remaining} s 后重发` : '发送验证码' }}
    </button>
    <geetest-captcha :config="{ product: 'bind' }" @initialized="onCaptchaInit" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import { GeetestCaptcha } from 'vue3-geetest'
import { useRequest } from 'alova/client'
import captchaApi from '@/api/captcha'
import { isEmail, isPhone } from '@/utils/validate'

const props = defineProps<{
  to: string
  useFor: string
}>()

let captchaInstance: any = null
const onCaptchaInit = (instance: any) => {
  captchaInstance = instance
  captchaInstance.onError((e: any) => {
    window.$message.error(e.msg)
  })
  captchaInstance.onSuccess(() => {
    doSend(captchaInstance.getValidate())
  })
}

const remaining = ref(0)
const isActive = computed(() => remaining.value > 0)
let timer: ReturnType<typeof setInterval> | null = null

const startCountdown = () => {
  remaining.value = 60
  timer = setInterval(() => {
    remaining.value--
    if (remaining.value <= 0 && timer) {
      clearInterval(timer)
      timer = null
    }
  }, 1000)
}

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

const loading = ref(false)

const handleSend = () => {
  if (!isPhone(props.to) && !isEmail(props.to)) {
    window.$message.error('请输入正确的手机号或邮箱')
    return
  }
  captchaInstance?.showCaptcha()
}

const doSend = (validation: any) => {
  loading.value = true
  const api = isPhone(props.to)
    ? captchaApi.sms(props.to, props.useFor, validation)
    : captchaApi.email(props.to, props.useFor, validation)

  useRequest(api)
    .onSuccess(() => {
      window.$message.success('验证码已发送')
      startCountdown()
    })
    .onComplete(() => {
      loading.value = false
    })
}
</script>
