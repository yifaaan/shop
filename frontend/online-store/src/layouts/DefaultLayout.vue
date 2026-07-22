<template>
  <div class="layout">
    <AppHeader />
    <main class="layout-main">
      <router-view />
    </main>
    <AppFooter />
  </div>
</template>

<script setup lang="ts">
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
</script>

<style scoped lang="scss">
.layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}
.layout-main {
  flex: 1;
}
</style>
