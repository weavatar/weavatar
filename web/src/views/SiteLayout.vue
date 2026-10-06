<template>
  <div class="min-h-screen flex flex-col">
    <app-header />
    <main class="flex-1 flex flex-col">
      <router-view v-slot="{ Component }">
        <transition name="page" mode="out-in">
          <component :is="Component" class="flex-1" />
        </transition>
      </router-view>
    </main>
    <app-footer />
  </div>
</template>

<script setup lang="ts">
import AppHeader from '@/components/layout/AppHeader.vue'
import AppFooter from '@/components/layout/AppFooter.vue'
</script>
