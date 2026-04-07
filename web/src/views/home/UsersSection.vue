<script setup lang="ts">
import logo_zmingcx from '@/assets/logo-zmingcx.png'
import logo_nicetheme from '@/assets/logo-nicetheme.png'
import logo_artalk from '@/assets/logo-artalk.png'
import logo_iro from '@/assets/logo-iro.png'
import logo_lxtx from '@/assets/logo-lxtx.png'
import logo_liyang from '@/assets/logo-liyang.png'
import logo_twikoo from '@/assets/logo-twikoo.png'
import logo_dami from '@/assets/logo-dami.png'

const users = [
  { url: 'https://zmingcx.com/', logo: logo_zmingcx },
  { url: 'https://www.nicetheme.cn/', logo: logo_nicetheme },
  { url: 'https://www.ilxtx.com/', logo: logo_lxtx },
  { url: 'https://github.com/mirai-mamori/Sakurairo', logo: logo_iro },
  { url: 'https://blog.qqsuu.cn/', logo: logo_dami },
  { url: 'https://artalk.js.org/', logo: logo_artalk },
  { url: 'https://twikoo.js.org/', logo: logo_twikoo },
  { url: 'https://www.liblog.cn/', logo: logo_liyang }
]
</script>

<template>
  <section class="py-20 px-6">
    <div class="max-w-360 mx-auto text-center">
      <h2 class="text-2xl sm:text-3xl font-700 mb-2">他们都在用</h2>
      <p class="opacity-50 mb-12">一些你可能认识的 TA 也在使用 WeAvatar，不妨来试试？</p>
      <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
        <a
          v-for="(u, i) in users"
          :key="i"
          :href="u.url"
          target="_blank"
          class="logo-card bg-white/5 dark:bg-white/5 border border-gray-200 dark:border-gray-700 backdrop-blur"
        >
          <img :src="u.logo" class="logo-img" />
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
