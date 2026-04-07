<template>
  <div>
    <NButton
      block
      :type="isActive ? 'tertiary' : 'primary'"
      :loading="loading"
      :disabled="isActive"
      @click="handleSend"
    >
      {{ isActive ? `${remaining} s` : '发送' }}
    </NButton>
    <GeetestCaptcha :config="{ product: 'bind' }" @initialized="onCaptchaInit" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import { NButton } from 'naive-ui'
import { GeetestCaptcha } from 'vue3-geetest'
import captchaApi from '@/api/captcha'
import { useRequest } from 'alova/client'

const props = defineProps<{
  to: string
  useFor: string
}>()

// 极验实例
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

// 倒计时
const remaining = ref(0)
const isActive = computed(() => remaining.value > 0)
let timer: ReturnType<typeof setInterval> | null = null

const startCountdown = () => {
  remaining.value = 60
  timer = setInterval(() => {
    remaining.value--
    if (remaining.value <= 0) {
      clearInterval(timer!)
      timer = null
    }
  }, 1000)
}

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

// 发送
const loading = ref(false)

const isPhone = (val: string) => /^1[3-9]\d{9}$/.test(val)
const isEmail = (val: string) => /^[\w.-]+@[\w.-]+\.\w+$/.test(val)

const handleSend = () => {
  if (!isPhone(props.to) && !isEmail(props.to)) {
    window.$message.error('请输入正确的手机号或邮箱')
    return
  }
  captchaInstance?.showCaptcha()
}

const doSend = async (validation: any) => {
  loading.value = true
  const api = isPhone(props.to)
    ? captchaApi.sms(props.to, props.useFor, validation)
    : captchaApi.email(props.to, props.useFor, validation)

  useRequest(api)
    .onSuccess(() => {
      window.$message.success('发送成功')
      startCountdown()
    })
    .onComplete(() => {
      loading.value = false
    })
}
</script>
