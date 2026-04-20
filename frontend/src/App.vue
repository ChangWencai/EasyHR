<template>
  <router-view v-if="!showOnboarding" />
  <div v-if="showOnboarding" class="onboarding-overlay" @click.self="dismissOnboarding">
    <div class="onboarding-card">
      <h3>欢迎使用易人事！</h3>
      <div class="onboarding-steps">
        <div class="step">
          <span class="step-num">1</span>
          <span>这是您的首页工作台，待办事项一目了然</span>
        </div>
        <div class="step">
          <span class="step-num">2</span>
          <span>待办卡片告诉您最近需要处理的事项</span>
        </div>
        <div class="step">
          <span class="step-num">3</span>
          <span>点击左侧菜单切换不同功能模块</span>
        </div>
      </div>
      <el-button type="primary" class="onboarding-btn" @click="dismissOnboarding">
        知道了
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

// Initialize synchronously from localStorage to avoid a two-render gap.
// Previously showOnboarding was hard-coded to false, causing OrgSetup.vue
// to render on the first pass before onMounted could check localStorage,
// resulting in a visible flicker of the page before the overlay appeared.
const showOnboarding = ref(!localStorage.getItem('onboarding_done'))

function dismissOnboarding() {
  localStorage.setItem('onboarding_done', 'true')
  showOnboarding.value = false
}
</script>

<style scoped>
.onboarding-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
}
.onboarding-card {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  margin: 16px;
  max-width: 320px;
  width: 100%;
}
.onboarding-card h3 {
  font-size: 18px;
  color: #1677ff;
  margin: 0 0 16px 0;
  text-align: center;
}
.onboarding-steps {
  margin-bottom: 20px;
}
.step {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 12px;
  font-size: 14px;
  color: #333;
}
.step-num {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #1677ff;
  color: #fff;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.onboarding-btn {
  width: 100%;
}
</style>
