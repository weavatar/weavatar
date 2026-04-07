<template>
  <div class="w-100 max-w-md mx-auto mt-20">
    <NCard title="我的资料">
      <div class="flex items-end my-8">
        <div class="w-12.5 h-12.5 border border-gray-200">
          <NImage width="50" :src="userStore.info.avatar" preview-disabled lazy />
        </div>
        <div>
          <h1 class="m-0 pl-5 text-xl">
            {{ userStore.info.nickname }}
            <NTag type="success">{{ userStore.info.real_name ? '已实名' : '未实名' }}</NTag>
          </h1>
          <h4 class="m-0 ml-5">
            <small>ID: {{ userStore.info.id }}</small>
          </h4>
        </div>
      </div>
      <h4>下面的设置目前仅在 WeAvatar 平台显示使用</h4>
      <NSpin :show="pageLoading">
        <NForm :model="model">
          <NFormItem path="nickname" label="昵称">
            <NInput
              v-model:value="model.nickname"
              placeholder="输入一个昵称"
              @keydown.enter.prevent
            />
          </NFormItem>
          <NFormItem path="avatar" label="头像">
            <NInput
              v-model:value="model.avatar"
              placeholder="输入一个图片地址"
              @keydown.enter.prevent
            />
          </NFormItem>
          <NButton type="primary" block :loading="saveLoading" @click="handleSave">保存</NButton>
        </NForm>
      </NSpin>
    </NCard>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { NButton, NCard, NForm, NFormItem, NImage, NInput, NSpin, NTag } from 'naive-ui'
import { useUserStore } from '@/stores'
import userApi from '@/api/user'
import { useRequest } from 'alova/client'

const userStore = useUserStore()

const pageLoading = ref(true)
const saveLoading = ref(false)
const model = ref({ nickname: '', avatar: '' })

useRequest(userApi.info())
  .onSuccess(({ data }: any) => {
    model.value = { nickname: data.nickname, avatar: data.avatar }
  })
  .onComplete(() => {
    pageLoading.value = false
  })

const handleSave = () => {
  saveLoading.value = true
  useRequest(userApi.update(model.value))
    .onSuccess(() => {
      userStore.freshUserInfo()
      window.$message.success('保存成功')
    })
    .onComplete(() => {
      saveLoading.value = false
    })
}
</script>
