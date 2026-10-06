<template>
  <div>
    <page-header title="关于 WeAvatar" desc="从 2022 年走到今天的公益头像服务。" />

    <div
      class="wrap py-12 sm:py-16 grid gap-14 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] lg:gap-20"
    >
      <div class="prose">
        <h2>WeAvatar 的前生</h2>
        <p>WeAvatar 的立项可以追溯至 2022 年 3 月，彼时我们依然在为 LitePress 项目工作。</p>
        <p>
          之后由于意见不合无法继续便开始着手筹备新项目，在随后的几个月里完成了 WeAvatar
          域名及商标注册，2022 年 6 月下旬启动开发。于 2022 年 7 月 25 日 Cravatar 成立 1
          周年之际灰度上线。
        </p>
        <p>由于琐事繁多加上仍对 LitePress 抱有期望，在之后的几个月内 WeAvatar 没有继续完善开发。</p>

        <h2>WeAvatar 的未来</h2>
        <p>WeAvatar 的目标市场不仅仅是网站，它将尝试进入桌面 App、移动 App、小程序等领域。</p>
        <p>路虽远行则将至，事虽难做则必成。期待 WeAvatar 的未来更加精彩！</p>

        <h2>赞助 WeAvatar</h2>
        <p>
          可以前往我们的
          <a target="_blank" rel="noreferrer" :href="AFDIAN_URL">爱发电主页</a>
          赞助我们。同时感谢以下支持者们：
        </p>
        <a
          :href="AFDIAN_URL"
          target="_blank"
          rel="noreferrer"
          class="not-prose block surface p-4 dark:bg-white/92"
        >
          <img :src="SPONSORS_SVG" alt="赞助者" loading="lazy" class="w-full h-auto" />
        </a>
      </div>

      <div>
        <h2 class="text-2xl font-700 tracking-tight text-fg pb-2.5 mb-6 border-b border-line">
          WeAvatar 的今世
        </h2>
        <ol class="relative ml-2 border-l border-line pl-8 space-y-7">
          <li v-for="(item, i) in timeline" :key="item.date" v-reveal="i * 40" class="relative">
            <span
              class="absolute -left-[2.3rem] top-1.5 w-[9px] h-[9px] rounded-full ring-4 ring-bg"
              :class="item.highlight ? 'bg-brand' : 'bg-fg3'"
            />
            <div class="text-xs font-600 text-fg3">{{ item.date }}</div>
            <div
              class="mt-1 text-[15px] leading-relaxed"
              :class="item.highlight ? 'font-600 text-fg' : 'text-fg2'"
            >
              {{ item.text }}
            </div>
          </li>
        </ol>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import PageHeader from '@/components/ui/PageHeader.vue'
import { AFDIAN_URL, SPONSORS_SVG } from '@/constants/links'

const timeline = [
  {
    date: '2022 年 11 月',
    text: '尝试接手 LitePress 项目未果后正式退出项目，随后不久完成 WeAvatar 初版控制台的开发。'
  },
  { date: '2023 年 1 月', text: '决定以 Golang 和 Vue 为主要开发语言进行 2.0 版本开发。' },
  {
    date: '2023 年 4 月',
    text: '缺乏维护的 LitePress 相关项目被打包出售，WeAvatar 开发开始提速。'
  },
  {
    date: '2023 年 5 月',
    text: 'WeAvatar 2.0 上线。这是一个全新的版本，宣告着最完善的头像服务就此诞生。',
    highlight: true
  },
  { date: '2023 年 7 月 25 日', text: 'WeAvatar 度过首个生日。' },
  { date: '2024 年 3 月', text: 'WeAvatar 日请求数达 500 万次。' },
  { date: '2024 年 12 月', text: 'WeAvatar 日请求数达 700 万次。' },
  { date: '2025 年 3 月', text: 'WeAvatar 3.0 上线。', highlight: true },
  { date: '2026 年 1 月', text: 'WeAvatar 日请求数达 1000 万次。', highlight: true }
]
</script>
