<template>
  <footer class="w-full bg-[#18181c] py-16 px-8">
    <div class="max-w-7xl mx-auto flex flex-col md:flex-row md:items-start md:justify-between gap-16">
      <!-- 品牌区 -->
      <div class="flex-1 flex flex-col items-start mb-8 md:mb-0 min-w-64">
        <div class="flex items-center gap-4">
          <img :src="logo" alt="WeAvatar" class="h-12 w-auto" />
        </div>
        <div class="text-gray-400 my-8">WeAvatar · 统一你的数字身份</div>
        <div class="flex gap-4">
          <NButton type="info" @click="handleJoinGroup">加 QQ 群</NButton>
          <NButton type="info" @click="handleJoinGroup">商务合作</NButton>
        </div>
      </div>
      <!-- 多列导航 -->
      <div class="flex-[2] grid grid-cols-2 sm:grid-cols-3 gap-16">
        <div>
          <div class="font-bold mb-6 text-white text-lg">WeAvatar</div>
          <ul class="space-y-4 list-none">
            <li><RouterLink :to="{ name: 'home' }" class="text-gray-400 hover:text-primary transition-colors">首页</RouterLink></li>
            <li><RouterLink :to="{ name: 'user-avatar' }" class="text-gray-400 hover:text-primary transition-colors">头像管理</RouterLink></li>
            <li><a href="https://status.haozi.net" target="_blank" class="text-gray-400 hover:text-primary transition-colors">服务状态</a></li>
            <li><RouterLink :to="{ name: 'about' }" class="text-gray-400 hover:text-primary transition-colors">关于我们</RouterLink></li>
          </ul>
        </div>
        <div>
          <div class="font-bold mb-6 text-white text-lg">相关项目</div>
          <ul class="space-y-4 list-none">
            <li><a href="https://account.haozi.net/" target="_blank" class="text-gray-400 hover:text-primary transition-colors">耗子通行证</a></li>
            <li><a href="https://tom.moe/" target="_blank" class="text-gray-400 hover:text-primary transition-colors">Moe Tom</a></li>
            <li><a href="https://twikoo.js.org/" target="_blank" class="text-gray-400 hover:text-primary transition-colors">Twikoo</a></li>
            <li><a href="https://artalk.js.org/" target="_blank" class="text-gray-400 hover:text-primary transition-colors">Artalk</a></li>
          </ul>
        </div>
        <div>
          <div class="font-bold mb-6 text-white text-lg">开源项目</div>
          <ul class="space-y-4 list-none">
            <li><a href="https://github.com/tnb-labs/panel" target="_blank" class="text-gray-400 hover:text-primary transition-colors">耗子面板</a></li>
            <li><a href="https://github.com/WenPai-org/wp-china-yes" target="_blank" class="text-gray-400 hover:text-primary transition-colors">WP-China-Yes</a></li>
            <li><a href="https://github.com/tnb-labs/LaravelInstaller" target="_blank" class="text-gray-400 hover:text-primary transition-colors">LaravelInstaller</a></li>
            <li><a href="https://github.com/tnb-labs/wp-setting-api" target="_blank" class="text-gray-400 hover:text-primary transition-colors">WP-Setting-API</a></li>
          </ul>
        </div>
      </div>
    </div>
    <!-- 版权和备案 -->
    <div class="mt-16 text-center text-gray-500 flex flex-col items-center gap-3">
      <div>
        Copyright © 2022-{{ new Date().getFullYear() }} WeAvatar. All Rights Reserved.
      </div>
      <div class="flex flex-col sm:flex-row items-center gap-3">
        <a href="https://beian.miit.gov.cn" target="_blank" rel="noreferrer" class="hover:text-primary text-gray-400">津ICP备2022009678号-2</a>
        <div class="hidden sm:flex items-center gap-2">
          <NImage :src="beian_gongan" width="16" alt="备案图标" lazy preview-disabled class="inline-block align-middle" />
          <a href="https://beian.mps.gov.cn/#/query/webSearch?code=12011502000847" target="_blank" rel="noreferrer" class="hover:text-primary text-gray-400">津公网安备12011502000847号</a>
        </div>
        <div class="flex sm:hidden items-center justify-center gap-2">
          <NImage :src="beian_gongan" width="16" alt="备案图标" lazy preview-disabled class="inline-block align-middle" />
          <a href="https://beian.mps.gov.cn/#/query/webSearch?code=12011502000847" target="_blank" rel="noreferrer" class="hover:text-primary text-gray-400">津公网安备12011502000847号</a>
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { NButton, NImage } from 'naive-ui'
import logo from '@/assets/logo-white.png'
import beian_gongan from '@/assets/beian-gongan.png'

const handleJoinGroup = () => {
  window.open('https://jq.qq.com/?_wv=1027&k=I1oJKSTH')
}
</script>
