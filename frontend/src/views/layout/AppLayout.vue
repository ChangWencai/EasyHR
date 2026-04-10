<template>
  <div class="app-layout">
    <div class="main-content">
      <router-view />
    </div>
    <BottomTabBar />
  </div>
</template>

<script setup lang="ts">
import BottomTabBar from './BottomTabBar.vue'
</script>

<style scoped lang="scss">
.app-layout {
  min-height: 100vh;
  background: #f5f5f5;
}

.main-content {
  padding-bottom: calc(70px + var(--safe-area-bottom, 0px));
}
</style>
