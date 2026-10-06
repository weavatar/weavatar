<template>
  <div>
    <hero-section :usage="usage" :avatars="avatars" />
    <playground-section />
    <features-section />
    <flow-section />
    <users-section />
    <sponsors-section />
    <cta-section />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRequest } from 'alova/client'
import systemApi from '@/api/system'
import HeroSection from './HeroSection.vue'
import PlaygroundSection from './PlaygroundSection.vue'
import FeaturesSection from './FeaturesSection.vue'
import FlowSection from './FlowSection.vue'
import UsersSection from './UsersSection.vue'
import SponsorsSection from './SponsorsSection.vue'
import CtaSection from './CtaSection.vue'

const usage = ref(0)
const avatars = ref<string[]>([])

const countMethod = systemApi.count()
countMethod.meta = { noAlert: true }
useRequest(countMethod).onSuccess(({ data }: any) => {
  usage.value = Number(data?.usage) || 0
})

const avatarsMethod = systemApi.randomAvatars()
avatarsMethod.meta = { noAlert: true }
useRequest(avatarsMethod).onSuccess(({ data }: any) => {
  avatars.value = Array.isArray(data?.avatars) ? data.avatars : []
})
</script>
