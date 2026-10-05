<template>
  <div
    class="relative flex-1 flex items-center justify-center px-6 sm:px-10 py-24 min-h-[60vh] overflow-x-clip"
  >
    <div
      class="absolute inset-0 -top-16 wa-grid wa-mask-radial pointer-events-none"
      aria-hidden="true"
    />
    <div class="relative text-center">
      <div
        class="font-mono text-[6rem] sm:text-[8rem] font-800 leading-none tracking-tighter text-fg/10 select-none"
      >
        404
      </div>
      <h1 class="-mt-6 display-text text-2xl sm:text-3xl text-fg">你可能迷失了方向</h1>
      <p class="mt-3 text-fg2">页面不存在。</p>
      <div class="mt-8 flex justify-center gap-3">
        <router-link :to="{ name: 'home' }" class="btn-primary">
          <span class="i-lucide-house text-base" />
          回到首页
        </router-link>
        <router-link :to="{ name: 'doc' }" class="btn-secondary">查看文档</router-link>
      </div>
    </div>
  </div>
</template>
