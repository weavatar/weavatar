<template>
  <div class="flex-grow">
    <RouterView />
  </div>
</template>
<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<style scoped>
.content {
  flex-grow: 1;
}
</style>
