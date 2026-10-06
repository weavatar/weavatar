<template>
  <div>
    <page-header title="我的资料" desc="以下资料仅在 WeAvatar 站内显示。" />

    <div
      class="mx-auto w-full max-w-5xl px-6 sm:px-10 py-10 sm:py-12 grid gap-6 lg:grid-cols-[20rem_minmax(0,1fr)]"
    >
      <div class="surface p-6 flex flex-col items-center text-center">
        <div class="relative">
          <avatar-image
            :src="preview || userStore.info.avatar"
            :size="96"
            rounded="2xl"
            alt="头像"
            class="ring-1 ring-line shadow-card"
          />
          <span
            v-if="userStore.info.real_name"
            class="absolute -bottom-1.5 -right-1.5 w-7 h-7 rounded-full bg-elev border border-line flex items-center justify-center text-green-500"
            title="已实名"
          >
            <span class="i-lucide-badge-check text-base" />
          </span>
        </div>
        <div class="mt-4 text-lg font-700 text-fg">{{ userStore.info.nickname }}</div>
        <div class="mt-1 flex items-center gap-2">
          <span
            class="chip"
            :class="userStore.info.real_name ? 'text-green-600 dark:text-green-400' : ''"
          >
            {{ userStore.info.real_name ? '已实名' : '未实名' }}
          </span>
        </div>
        <dl class="mt-6 w-full text-sm divide-y divide-line">
          <div class="flex items-center justify-between py-2.5">
            <dt class="text-fg3">用户 ID</dt>
            <dd class="font-mono text-fg flex items-center gap-1">
              {{ userStore.info.id || '—' }}
              <copy-button v-if="userStore.info.id" :text="userStore.info.id" class="!w-6 !h-6" />
            </dd>
          </div>
          <div class="flex items-center justify-between py-2.5">
            <dt class="text-fg3">注册时间</dt>
            <dd class="text-fg tabular">{{ formatDate(userStore.info.created_at) || '—' }}</dd>
          </div>
        </dl>
      </div>

      <div class="surface p-6 sm:p-8">
        <h2 class="text-base font-700 text-fg">基本资料</h2>
        <n-spin :show="pageLoading">
          <div class="mt-6 space-y-5">
            <label class="block">
              <span class="block text-sm font-600 text-fg mb-2">昵称</span>
              <n-input
                v-model:value="model.nickname"
                size="large"
                placeholder="输入一个昵称"
                maxlength="30"
                show-count
                @keydown.enter.prevent
              />
            </label>
            <label class="block">
              <span class="block text-sm font-600 text-fg mb-2">头像地址</span>
              <n-input
                v-model:value="model.avatar"
                size="large"
                placeholder="输入一个图片地址（https://…）"
                :input-props="{ spellcheck: false }"
                @keydown.enter.prevent
              >
                <template #prefix>
                  <span class="i-lucide-link text-fg3" />
                </template>
              </n-input>
              <span class="mt-1.5 block text-xs text-fg3">仅用于控制台显示。</span>
            </label>
            <div class="pt-2 flex items-center gap-3">
              <button type="button" class="btn-primary" :disabled="saveLoading" @click="handleSave">
                <span v-if="saveLoading" class="i-lucide-loader-circle animate-spin text-base" />
                保存修改
              </button>
              <router-link :to="{ name: 'user-avatar' }" class="btn-ghost"
                >返回头像管理</router-link
              >
            </div>
          </div>
        </n-spin>
      </div>

      <div class="surface p-6 sm:p-8 lg:col-span-2">
        <div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
          <div>
            <h2 class="text-base font-700 text-fg">注销账号</h2>
            <p class="mt-1.5 text-sm text-fg2 leading-relaxed">
              注销后将删除你上传的全部头像，且不可恢复。
            </p>
          </div>
          <button type="button" class="btn-danger shrink-0" @click="showDeletion = true">
            <span class="i-lucide-user-x text-base" />
            注销账号
          </button>
        </div>
      </div>
    </div>

    <n-modal
      v-model:show="showDeletion"
      preset="card"
      title="注销账号"
      :style="{ width: '460px', maxWidth: '94vw' }"
      :bordered="false"
      :mask-closable="!deletionLoading"
      :auto-focus="false"
    >
      <p class="text-sm text-fg2 leading-relaxed">
        注销后将删除你上传的全部头像，且不可恢复。为确认身份，需要重新通过树新峰通行证授权，授权完成后账号即被注销。
      </p>
      <div class="mt-6 flex items-center justify-end gap-3">
        <button
          type="button"
          class="btn-ghost"
          :disabled="deletionLoading"
          @click="showDeletion = false"
        >
          取消
        </button>
        <button
          type="button"
          class="btn-danger"
          :disabled="deletionLoading"
          @click="handleDeletion"
        >
          <span v-if="deletionLoading" class="i-lucide-loader-circle animate-spin text-base" />
          {{ deletionLoading ? '正在跳转…' : '前往确认身份' }}
        </button>
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { NInput, NModal, NSpin } from 'naive-ui'
import { useRequest } from 'alova/client'
import PageHeader from '@/components/ui/PageHeader.vue'
import AvatarImage from '@/components/ui/AvatarImage.vue'
import CopyButton from '@/components/ui/CopyButton.vue'
import { useUserStore } from '@/stores'
import userApi from '@/api/user'
import { formatDate } from '@/utils/format'

const userStore = useUserStore()

const pageLoading = ref(true)
const saveLoading = ref(false)
const model = ref({ nickname: '', avatar: '' })
const showDeletion = ref(false)
const deletionLoading = ref(false)

const preview = computed(() => (/^https?:\/\//.test(model.value.avatar) ? model.value.avatar : ''))

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

// 回调页凭此意图完成注销而不是登录
const handleDeletion = () => {
  deletionLoading.value = true
  useRequest(userApi.deletionLogin())
    .onSuccess(({ data }: any) => {
      sessionStorage.setItem('oauth_intent', 'delete')
      window.location.href = data.url
    })
    .onError(() => {
      deletionLoading.value = false
    })
}
</script>
