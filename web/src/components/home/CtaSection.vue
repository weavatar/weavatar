<template>
  <section class="relative border-t border-line py-16 sm:py-24">
    <div class="wrap">
      <div
        v-reveal
        class="relative overflow-hidden rounded-3xl bg-brand text-white px-8 py-14 sm:px-14 sm:py-20 text-center"
      >
        <div
          class="absolute inset-0 wa-grid opacity-30 [--wa-grid:rgba(255,255,255,0.35)]"
          aria-hidden="true"
        />
        <div
          class="absolute -top-32 -right-32 w-96 h-96 rounded-full bg-white/20 blur-3xl"
          aria-hidden="true"
        />
        <div class="relative">
          <h2 class="display-text text-3xl sm:text-5xl">现在就设置你的头像</h2>
          <p class="mt-5 text-base sm:text-lg text-white/80 max-w-xl mx-auto">
            一分钟完成设置，永久免费。
          </p>
          <div class="mt-9 flex flex-wrap justify-center gap-3">
            <router-link
              :to="{ name: 'login' }"
              class="btn btn-lg bg-white text-brand hover:bg-white/90 hover:-translate-y-0.5 shadow-card"
            >
              开始使用
              <span class="i-lucide-arrow-right text-base" />
            </router-link>
            <router-link
              :to="{ name: 'doc' }"
              class="btn btn-lg bg-white/10 text-white border border-white/30 hover:bg-white/20 hover:-translate-y-0.5"
            >
              阅读文档
            </router-link>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
