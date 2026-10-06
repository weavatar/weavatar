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
    <geetest-captcha :config="{ product: 'bind' }" @initialized="onInit" />
  </div>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { GeetestCaptcha } from 'vue3-geetest'
import { sendEmail, sendSms, type VerifyCodeUse } from '@/api/verifyCode'
import { useGeetest } from '@/composables/useGeetest'
import { isEmail, isPhone } from '@/utils/validate'

const props = defineProps<{
  to: string
  useFor: VerifyCodeUse
  /** 限定只能发送到手机或邮箱 */
  only?: 'phone' | 'email'
}>()

const { onInit, verify } = useGeetest()

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

const handleSend = async () => {
  const to = props.to.trim()
  const phoneOk = props.only !== 'email' && isPhone(to)
  const emailOk = props.only !== 'phone' && isEmail(to)
  if (!phoneOk && !emailOk) {
    window.$message.error(
      props.only === 'phone'
        ? '请输入正确的手机号'
        : props.only === 'email'
          ? '请输入正确的邮箱'
          : '请输入正确的手机号或邮箱'
    )
    return
  }

  try {
    const captcha = await verify()
    loading.value = true
    if (phoneOk) await sendSms(to, props.useFor, captcha)
    else await sendEmail(to, props.useFor, captcha)
    window.$message.success('验证码已发送')
    startCountdown()
  } catch {
    // 取消验证或发送失败，错误已提示
  } finally {
    loading.value = false
  }
}
</script>
