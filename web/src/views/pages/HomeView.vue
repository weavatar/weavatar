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
import HeroSection from '@/components/home/HeroSection.vue'
import PlaygroundSection from '@/components/home/PlaygroundSection.vue'
import FeaturesSection from '@/components/home/FeaturesSection.vue'
import FlowSection from '@/components/home/FlowSection.vue'
import UsersSection from '@/components/home/UsersSection.vue'
import SponsorsSection from '@/components/home/SponsorsSection.vue'
import CtaSection from '@/components/home/CtaSection.vue'
import { fetchRandomAvatars, fetchUsage } from '@/api/system'

const usage = ref(0)
const avatars = ref<string[]>([])

// 两个接口都不弹错误提示，失败时保留默认值
const load = async () => {
  const [count, random] = await Promise.allSettled([fetchUsage(), fetchRandomAvatars()])
  if (count.status === 'fulfilled') usage.value = count.value.usage
  if (random.status === 'fulfilled') avatars.value = random.value.avatars
}
load()
</script>
