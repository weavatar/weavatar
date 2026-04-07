<script setup lang="ts">
import { ref, nextTick } from 'vue'
import type { NumberAnimationInst } from 'naive-ui'
import systemApi from '@/api/system'
import HeroSection from './HeroSection.vue'
import FeaturesSection from './FeaturesSection.vue'
import UsersSection from './UsersSection.vue'
import SponsorsSection from './SponsorsSection.vue'
import { useRequest } from 'alova/client'

const usage = ref(0)
const usageRef = ref<NumberAnimationInst | null>(null)
const avatars = ref<string[]>([])

useRequest(systemApi.count(), { meta: { noAlert: true } }).onSuccess(({ data }: any) => {
  usage.value = data.usage
  nextTick(() => usageRef.value?.play())
})

useRequest(systemApi.randomAvatars(), { meta: { noAlert: true } }).onSuccess(({ data }: any) => {
  avatars.value = data.avatars || []
})
</script>

<template>
  <main>
    <HeroSection v-model:usage-ref="usageRef" :usage="usage" :avatars="avatars" />
    <FeaturesSection />
    <UsersSection />
    <SponsorsSection />
  </main>
</template>
