<script setup lang="ts">
import logo_dunyun from '@/assets/logo-dunyun.png'
import logo_wafpro from '@/assets/logo-wafpro.png'
import logo_dk from '@/assets/logo-dk.png'
import logo_anycast from '@/assets/logo-anycast.png'

const sponsors = [
  { url: 'https://www.dkdun.cn/aff/MQZZNVHQ', logo: logo_dk },
  { url: 'https://waf.pro/', logo: logo_wafpro },
  { url: 'https://dun.ddunyun.com', logo: logo_dunyun },
  { url: 'https://www.anycast.ai', logo: logo_anycast }
]
</script>

<template>
  <section class="py-20 px-6">
    <div class="max-w-360 mx-auto text-center">
      <h2 class="text-2xl sm:text-3xl font-700 mb-2">我们的赞助商</h2>
      <p class="opacity-50 mb-12">作为公益性质的项目，WeAvatar 的稳定运行离不开它们的帮助</p>
      <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
        <a
          v-for="(s, i) in sponsors"
          :key="i"
          :href="s.url"
          target="_blank"
          class="logo-card bg-white/5 dark:bg-white/5 border border-gray-200 dark:border-gray-700 backdrop-blur"
        >
          <img :src="s.logo" class="logo-img" />
        </a>
      </div>
    </div>
  </section>
</template>

<style scoped>
.logo-card {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 120px;
  border-radius: 12px;
  padding: 0 1.5rem;
  overflow: hidden;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
  transition: all 0.3s;
}
.logo-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
}

.logo-img {
  display: block;
  height: 52px;
  width: auto;
  max-width: 100%;
  object-fit: contain;
}
</style>
