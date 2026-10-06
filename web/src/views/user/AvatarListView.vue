<template>
  <div>
    <page-header title="头像管理" desc="为邮箱或手机号绑定头像，修改后 3 小时内全网生效。">
      <template #actions>
        <button type="button" class="btn-primary" @click="addModal = true">
          <span class="i-lucide-plus text-base" />
          添加头像
        </button>
      </template>
    </page-header>

    <div class="wrap py-10 sm:py-12">
      <div class="surface-muted px-5 py-4 flex items-start gap-3 text-sm text-fg2 leading-relaxed">
        <span class="i-lucide-info text-brand text-lg mt-0.5" />
        <div>
          你可以通过
          <code class="kbd">https://weavatar.com/avatar/地址的 SHA256 或 MD5</code>
          的方式访问自己的头像。
          <router-link :to="{ name: 'help' }" class="link">查看帮助</router-link>
          ·
          <router-link :to="{ name: 'doc' }" class="link">查看文档</router-link>
        </div>
      </div>

      <div v-if="loading && !data.length" class="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div v-for="i in 6" :key="i" class="surface p-5 flex gap-4 animate-pulse">
          <div class="w-[4.5rem] h-[4.5rem] rounded-2xl bg-muted shrink-0" />
          <div class="flex-1 space-y-3 pt-1">
            <div class="h-4 w-2/3 rounded bg-muted" />
            <div class="h-3 w-1/2 rounded bg-muted" />
            <div class="h-3 w-1/3 rounded bg-muted" />
          </div>
        </div>
      </div>

      <div
        v-else-if="!loading && !data.length"
        class="mt-8 rounded-3xl border border-dashed border-line-strong wa-dots py-20 px-6 text-center"
      >
        <div
          class="mx-auto w-16 h-16 rounded-2xl bg-elev border border-line shadow-card flex items-center justify-center text-brand"
        >
          <span class="i-lucide-image-plus text-3xl" />
        </div>
        <h2 class="mt-6 text-lg font-700 text-fg">还没有头像</h2>
        <p class="mt-2 text-sm text-fg2 max-w-sm mx-auto leading-relaxed">
          添加后，所有接入 WeAvatar 的网站都会自动显示。
        </p>
        <button type="button" class="btn-primary mt-7" @click="addModal = true">
          <span class="i-lucide-plus text-base" />
          添加第一个头像
        </button>
      </div>

      <div
        v-else
        class="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-3"
        :class="loading ? 'opacity-60 pointer-events-none' : ''"
      >
        <article
          v-for="row in data"
          :key="row.sha256"
          class="group surface p-5 flex flex-col gap-4 transition-all duration-300 hover:shadow-card hover:-translate-y-0.5"
        >
          <div class="flex gap-4">
            <avatar-image
              :src="avatarUrl(row.sha256)"
              :size="72"
              rounded="2xl"
              :alt="row.raw"
              class="ring-1 ring-line"
            />
            <div class="min-w-0 flex-1">
              <div class="font-600 text-fg truncate" :title="row.raw">{{ row.raw }}</div>
              <div class="mt-1.5 flex items-center gap-1 text-xs text-fg3 font-mono">
                <span class="truncate" :title="row.sha256">{{ shortHash(row.sha256, 10, 6) }}</span>
                <copy-button :text="row.sha256" class="!w-6 !h-6" />
              </div>
              <div v-if="row.created_at" class="mt-1 text-xs text-fg3">
                添加于 {{ formatDate(row.created_at) }}
              </div>
            </div>
          </div>
          <div class="flex items-center gap-2 pt-3 border-t border-line">
            <button type="button" class="btn-secondary btn-sm flex-1" @click="openEdit(row.sha256)">
              <span class="i-lucide-image text-sm" />
              修改图片
            </button>
            <n-popconfirm
              placement="top"
              positive-text="删除"
              negative-text="取消"
              @positive-click="handleDelete(row.sha256)"
            >
              <template #trigger>
                <button type="button" class="btn-danger btn-sm">
                  <span class="i-lucide-trash text-sm" />
                  删除
                </button>
              </template>
              删除后 3 小时内全网生效，确定删除这个头像吗？
            </n-popconfirm>
          </div>
        </article>
      </div>

      <div
        v-if="pagination.itemCount > 0"
        class="mt-10 flex flex-wrap items-center justify-between gap-4"
      >
        <div class="text-sm text-fg3">共 {{ pagination.itemCount }} 个头像</div>
        <n-pagination
          v-model:page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :item-count="pagination.itemCount"
          :page-sizes="[12, 24, 48]"
          show-size-picker
          @update:page="load"
          @update:page-size="onPageSize"
        />
      </div>
    </div>

    <avatar-add-modal v-model:show="addModal" @success="load(1)" />
    <avatar-edit-modal v-model:show="editModal" :hash="editHash" @success="load(pagination.page)" />
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { NPagination, NPopconfirm } from 'naive-ui'
import { useRequest } from 'alova/client'
import PageHeader from '@/components/ui/PageHeader.vue'
import AvatarImage from '@/components/ui/AvatarImage.vue'
import CopyButton from '@/components/ui/CopyButton.vue'
import AvatarAddModal from './AvatarAddModal.vue'
import AvatarEditModal from './AvatarEditModal.vue'
import avatarApi from '@/api/avatar'
import { API_AVATAR_BASE } from '@/constants/links'
import { formatDate, shortHash } from '@/utils/format'

interface AvatarRow {
  sha256: string
  md5: string
  raw: string
  created_at?: string
}

const loading = ref(true)
const data = ref<AvatarRow[]>([])
const addModal = ref(false)
const editModal = ref(false)
const editHash = ref('')
const cacheKey = ref(Date.now())

const pagination = reactive({
  page: 1,
  pageSize: 12,
  itemCount: 0
})

const avatarUrl = (hash: string) => `${API_AVATAR_BASE}/${hash}?s=144&t=${cacheKey.value}`

const load = (page = pagination.page) => {
  loading.value = true
  useRequest(avatarApi.list(page, pagination.pageSize))
    .onSuccess(({ data: res }: any) => {
      data.value = res.items ?? []
      pagination.page = page
      pagination.itemCount = res.total ?? 0
      cacheKey.value = Date.now()
    })
    .onComplete(() => {
      loading.value = false
    })
}

const onPageSize = (size: number) => {
  pagination.pageSize = size
  load(1)
}

const openEdit = (hash: string) => {
  editHash.value = hash
  editModal.value = true
}

const handleDelete = (hash: string) => {
  loading.value = true
  useRequest(avatarApi.delete(hash))
    .onSuccess(() => {
      window.$message.success('删除成功，3 小时内全网生效')
      const lastPage = Math.max(1, Math.ceil((pagination.itemCount - 1) / pagination.pageSize))
      load(Math.min(pagination.page, lastPage))
    })
    .onError(() => {
      loading.value = false
    })
}

load(1)
</script>
