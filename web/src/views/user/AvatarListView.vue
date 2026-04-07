<template>
  <div class="p-10 lt-md:px-0 lt-md:py-10">
    <NCard title="头像管理">
      <NSpace vertical>
        <NAlert type="info">
          你可以通过
          <b>https://weavatar.com/avatar/地址 SHA256 或 MD5 值</b> 的方式访问自己的头像。
          <RouterLink :to="{ name: 'help' }">查看帮助</RouterLink>
          /
          <RouterLink :to="{ name: 'doc' }">查看文档</RouterLink>
        </NAlert>
        <NDataTable
          striped
          remote
          :columns="columns"
          :data="data"
          :loading="loading"
          :pagination="pagination"
          :row-key="(row: any) => row.sha256"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
        <NCard :bordered="false">
          <NButton type="primary" size="large" class="block mx-auto" @click="addModal = true">
            添加头像
          </NButton>
        </NCard>
      </NSpace>
    </NCard>

    <AvatarAddModal v-model:show="addModal" @success="handlePageChange(1)" />
    <AvatarEditModal v-model:show="editModal" :hash="editHash" @success="handlePageChange(1)" />
  </div>
</template>

<script setup lang="ts">
import type { DataTableColumns } from 'naive-ui'
import { NAlert, NButton, NCard, NDataTable, NPopconfirm, NSpace } from 'naive-ui'
import { h, reactive, ref } from 'vue'
import avatarApi from '@/api/avatar'
import AvatarAddModal from './AvatarAddModal.vue'
import AvatarEditModal from './AvatarEditModal.vue'
import { useRequest } from 'alova/client'

const loading = ref(true)
const data = ref<any[]>([])
const addModal = ref(false)
const editModal = ref(false)
const editHash = ref('')

const pagination = reactive({
  page: 1,
  pageCount: 1,
  pageSize: 10,
  itemCount: 0,
  showQuickJumper: true,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100]
})

const handlePageSizeChange = (pageSize: number) => {
  pagination.pageSize = pageSize
  handlePageChange(1)
}

const handlePageChange = (page: number) => {
  loading.value = true
  useRequest(avatarApi.list(page, pagination.pageSize))
    .onSuccess(({ data: res }: any) => {
      data.value = res.items
      pagination.page = page
      pagination.itemCount = res.total
      pagination.pageCount = Math.ceil(res.total / pagination.pageSize)
    })
    .onComplete(() => {
      loading.value = false
    })
}

handlePageChange(1)

const handleDelete = (hash: string) => {
  loading.value = true
  useRequest(avatarApi.delete(hash))
    .onSuccess(() => {
      window.$message.success('删除成功，3 小时内全网生效')
      handlePageChange(1)
    })
    .onComplete(() => {
      loading.value = false
    })
}

const columns: DataTableColumns<any> = [
  {
    title: '头像',
    key: 'hash',
    render({ sha256 }) {
      return h('img', {
        src: `https://weavatar.com/api/avatar/${sha256}?s=50&t=${Date.now()}`,
        style: { borderRadius: '10%' }
      })
    }
  },
  {
    title: '地址',
    key: 'raw',
    ellipsis: { tooltip: true }
  },
  {
    title: '操作',
    key: 'actions',
    render(row) {
      return h(NSpace, {}, () => [
        h(
          NButton,
          {
            size: 'small',
            type: 'info',
            onClick: () => {
              editHash.value = row.sha256
              editModal.value = true
            }
          },
          () => '修改图片'
        ),
        h(
          NPopconfirm,
          { onPositiveClick: () => handleDelete(row.sha256) },
          {
            default: () => '确定删除头像吗？',
            trigger: () => h(NButton, { size: 'small', type: 'error' }, () => '删除头像')
          }
        )
      ])
    }
  }
]
</script>
