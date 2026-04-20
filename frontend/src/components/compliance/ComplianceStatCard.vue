<template>
  <div class="compliance-stat-card glass-card">
    <div class="stat-icon" :class="iconClass">
      <el-icon><component :is="icon" /></el-icon>
    </div>
    <div class="stat-body">
      <div class="stat-value">{{ value }}</div>
      <div class="stat-label">{{ label }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  value: string | number
  label: string
  icon: string
  iconClass?: string
}>()
</script>

<style scoped lang="scss">
.compliance-stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  transition: all 0.2s ease;
  &:hover { transform: translateY(-2px); box-shadow: var(--shadow-lg); }
}
.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
}
.stat-body { display: flex; flex-direction: column; gap: 2px; }
.stat-value { font-size: 26px; font-weight: 700; color: var(--text-primary); line-height: 1.2; }
.stat-label { font-size: 12px; color: var(--text-tertiary); }
</style>
