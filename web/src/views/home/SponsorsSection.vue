<template>
  <section class="relative section border-t border-line">
    <div class="wrap">
      <section-title title="我们的赞助商" desc="WeAvatar 是公益项目，稳定运行离不开它们的支持。" />

      <div v-reveal class="mt-12 grid grid-cols-2 gap-4 lg:grid-cols-4">
        <a
          v-for="s in sponsors"
          :key="s.url"
          :href="s.url"
          target="_blank"
          rel="noreferrer"
          :title="s.name"
          class="flex items-center justify-center h-28 px-8 rounded-2xl border border-line bg-elev dark:bg-white/92 transition-all duration-300 hover:-translate-y-1 hover:shadow-card"
        >
          <img
            :src="s.logo"
            :alt="s.name"
            loading="lazy"
            class="h-11 w-auto max-w-full object-contain"
          />
        </a>
      </div>

      <div v-reveal class="mt-10 flex flex-wrap items-center justify-center gap-3 text-sm text-fg2">
        <span>想赞助或合作？</span>
        <a :href="AFDIAN_URL" target="_blank" rel="noreferrer" class="btn-secondary btn-sm">
          <span class="i-lucide-heart text-base text-red-500" />
          爱发电赞助
        </a>
        <a :href="QQ_GROUP_URL" target="_blank" rel="noreferrer" class="btn-ghost btn-sm">
          联系我们
          <span class="i-lucide-arrow-up-right text-sm" />
        </a>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import SectionTitle from '@/components/ui/SectionTitle.vue'
import { AFDIAN_URL, QQ_GROUP_URL } from '@/constants/links'
import logoDunyun from '@/assets/logo-dunyun.png'
import logoWafpro from '@/assets/logo-wafpro.png'
import logoDk from '@/assets/logo-dk.png'
import logoAnycast from '@/assets/logo-anycast.png'

const sponsors = [
  { name: 'DK 盾', url: 'https://www.dkdun.cn/aff/MQZZNVHQ', logo: logoDk },
  { name: 'WAF Pro', url: 'https://waf.pro/', logo: logoWafpro },
  { name: '盾云', url: 'https://dun.ddunyun.com', logo: logoDunyun },
  { name: 'Anycast', url: 'https://www.anycast.ai', logo: logoAnycast }
]
</script>
