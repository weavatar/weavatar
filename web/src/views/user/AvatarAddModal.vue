<template>
  <NModal :show="show" @update:show="$emit('update:show', $event)">
    <NCard closable title="添加头像" style="width: 40vh" @close="$emit('update:show', false)">
      <NForm :model="model">
        <NFormItem path="raw" label="地址">
          <NInput v-model:value="model.raw" placeholder="手机号 / 邮箱" @keydown.enter.prevent />
        </NFormItem>
        <NFormItem path="verify_code" label="验证码">
          <NRow :gutter="[0, 24]">
            <NCol :span="14">
              <NInput v-model:value="model.verify_code" @keydown.enter.prevent />
            </NCol>
            <NCol :span="2" />
            <NCol :span="8">
              <VerifyCodeButton :to="model.raw" use-for="avatar" />
            </NCol>
          </NRow>
        </NFormItem>
        <NDivider>上传头像或获取QQ头像</NDivider>
        <NFormItem path="avatar" label="上传头像">
          <NUpload
            v-show="!hasAvatar"
            directory-dnd
            :default-upload="false"
            :show-file-list="false"
            @change="handleUpload"
            @before-upload="sanitizeAvatar"
          >
            <NUploadDragger>
              <div class="mb-3">
                <NIcon size="48" :depth="3"><ArchiveIcon /></NIcon>
              </div>
              <NText class="text-base">点击或者拖动图片到该区域来上传</NText>
              <NP depth="3" class="mt-2 mb-0">上传的图片需符合中华人民共和国相关法律法规要求</NP>
            </NUploadDragger>
          </NUpload>
          <NButton v-show="hasAvatar" type="primary" block @click="avatarBlob = null">
            重新上传
          </NButton>
        </NFormItem>
        <NFormItem path="qq" label="获取QQ头像">
          <NRow :gutter="[0, 24]">
            <NCol :span="14">
              <NInput v-model:value="qq" @keydown.enter.prevent />
            </NCol>
            <NCol :span="2" />
            <NCol :span="8">
              <NButton block type="primary" :loading="qqLoading" @click="handleGetQQ"
                >一键获取</NButton
              >
            </NCol>
          </NRow>
        </NFormItem>
      </NForm>
      <NDivider />
      <NButton type="info" block :loading="submitLoading" @click="handleSubmit">提交</NButton>
    </NCard>
  </NModal>

  <CropAvatar v-model:show="cropShow" :image="cropImage" @crop="handleCrop" />
  <GeetestCaptcha :config="{ product: 'bind' }" @initialized="onCaptchaInit" />
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  NButton,
  NCard,
  NCol,
  NDivider,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  NModal,
  NP,
  NRow,
  NText,
  NUpload,
  NUploadDragger
} from 'naive-ui'
import { ArchiveOutline as ArchiveIcon } from '@vicons/ionicons5'
import { GeetestCaptcha } from 'vue3-geetest'
import avatarApi from '@/api/avatar'
import userApi from '@/api/user'
import CropAvatar from '@/components/avatar/CropAvatar.vue'
import VerifyCodeButton from '@/components/captcha/VerifyCodeButton.vue'
import { useRequest } from 'alova/client'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{
  'update:show': [value: boolean]
  success: []
}>()

const model = ref({ raw: '', verify_code: '' })
const avatarBlob = ref<Blob | null>(null)
const hasAvatar = computed(() => avatarBlob.value && avatarBlob.value.size > 0)
const qq = ref('')
const qqLoading = ref(false)
const submitLoading = ref(false)

// 裁剪
const cropShow = ref(false)
const cropImage = ref<Blob | null>(null)

const sanitizeAvatar = (data: { file: any }) => {
  const type = data.file.file?.type
  const size = data.file.file?.size || 0
  const validType = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'].includes(type)
  const validSize = size / 1024 / 1024 < 5
  if (!validType) window.$message.error('只能上传 JPG / PNG / GIF / WEBP 格式')
  if (!validSize) window.$message.error('图片大小不能超过 5 MB')
  return validType && validSize
}

const handleUpload = (options: { file: any }) => {
  cropImage.value = options.file.file as Blob
  cropShow.value = true
}

const handleCrop = (blob: Blob) => {
  avatarBlob.value = blob
}

const handleGetQQ = () => {
  if (!qq.value) {
    window.$message.error('请输入QQ号')
    return
  }
  qqLoading.value = true
  useRequest(userApi.qq(qq.value))
    .onSuccess(({ data }: any) => {
      window.$message.success('获取成功')
      const binary = atob(data)
      const bytes = new Uint8Array(binary.length)
      for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
      cropImage.value = new Blob([bytes], { type: 'image/png' })
      cropShow.value = true
    })
    .onComplete(() => {
      qqLoading.value = false
    })
}

// 极验
let captchaInstance: any = null
const onCaptchaInit = (instance: any) => {
  captchaInstance = instance
  captchaInstance.onError((e: any) => window.$message.error(e.msg))
  captchaInstance.onSuccess(() => doSubmit(captchaInstance.getValidate()))
}

const isPhone = (val: string) => /^1[3-9]\d{9}$/.test(val)
const isEmail = (val: string) => /^[\w.-]+@[\w.-]+\.\w+$/.test(val)

const handleSubmit = () => {
  if (!avatarBlob.value?.size) return window.$message.error('请先上传头像')
  if (!model.value.raw) return window.$message.error('请先输入地址')
  if (!isPhone(model.value.raw) && !isEmail(model.value.raw))
    return window.$message.error('请输入正确的手机号或邮箱')
  if (!model.value.verify_code) return window.$message.error('请先输入验证码')
  captchaInstance?.showCaptcha()
}

const doSubmit = (captchaValidation: any) => {
  submitLoading.value = true

  // 先检查绑定
  useRequest(avatarApi.check(model.value.raw))
    .onSuccess(({ data: res }: any) => {
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
            model.value = { raw: '', verify_code: '' }
            avatarBlob.value = null
          })
          .onComplete(() => {
            submitLoading.value = false
          })
      }

      if (res.bind) {
        window.$dialog.warning({
          title: '警告',
          content: '该地址已被其他用户添加，是否继续添加？',
          positiveText: '是',
          negativeText: '否',
          onPositiveClick: submitAvatar,
          onNegativeClick: () => {
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
