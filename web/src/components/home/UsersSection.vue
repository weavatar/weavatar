<template>
  <section class="relative section border-t border-line overflow-hidden">
    <div class="wrap">
      <section-title title="他们都在用" desc="这些项目和站点都已接入 WeAvatar。" />
    </div>

    <div v-reveal class="mt-12 overflow-hidden wa-mask-x">
      <div class="flex w-max animate-marquee hover:[animation-play-state:paused]">
        <div v-for="copy in 2" :key="copy" class="flex gap-4 pr-4">
          <a
            v-for="u in users"
            :key="`${copy}-${u.url}`"
            :href="u.url"
            target="_blank"
            rel="noreferrer"
            :title="u.name"
            class="group flex items-center justify-center w-48 h-24 px-7 rounded-2xl border border-line bg-elev dark:bg-white/92 transition-all duration-300 hover:-translate-y-1 hover:shadow-card"
          >
            <img
              :src="u.logo"
              :alt="u.name"
              loading="lazy"
              class="h-10 w-auto max-w-full object-contain transition-transform duration-300 group-hover:scale-105"
            />
          </a>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import SectionTitle from '@/components/ui/SectionTitle.vue'
import logoZmingcx from '@/assets/logo-zmingcx.png'
import logoNicetheme from '@/assets/logo-nicetheme.png'
import logoArtalk from '@/assets/logo-artalk.png'
import logoIro from '@/assets/logo-iro.png'
import logoLxtx from '@/assets/logo-lxtx.png'
import logoLiyang from '@/assets/logo-liyang.png'
import logoTwikoo from '@/assets/logo-twikoo.png'
import logoForeverblog from '@/assets/logo-foreverblog-black.png'

const users = [
  { name: '知更鸟', url: 'https://zmingcx.com/', logo: logoZmingcx },
  { name: 'nicetheme', url: 'https://www.nicetheme.cn/', logo: logoNicetheme },
  { name: '龙笑天下', url: 'https://www.ilxtx.com/', logo: logoLxtx },
  { name: 'Sakurairo', url: 'https://github.com/mirai-mamori/Sakurairo', logo: logoIro },
  { name: '十年之约', url: 'https://www.foreverblog.cn/', logo: logoForeverblog },
  { name: 'Artalk', url: 'https://artalk.js.org/', logo: logoArtalk },
  { name: 'Twikoo', url: 'https://twikoo.js.org/', logo: logoTwikoo },
  { name: '李洋博客', url: 'https://www.liblog.cn/', logo: logoLiyang }
]
</script>
