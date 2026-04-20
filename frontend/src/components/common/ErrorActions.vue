<script setup lang="ts">
import { CircleCloseFilled } from '@element-plus/icons-vue'

defineProps<{
  message: string
}>()
const emit = defineEmits<{
  retry: []
  contactAdmin: []
}>()
</script>

<template>
  <div class="error-actions">
    <div class="error-icon">
      <el-icon color="var(--danger)" :size="32"><CircleCloseFilled /></el-icon>
    </div>
    <p class="error-message">{{ message }}</p>
    <div class="error-btn-group">
      <el-button @click="emit('retry')">重试</el-button>
      <el-button type="primary" plain @click="emit('contactAdmin')">联系管理员</el-button>
    </div>
  </div>
</template>

<style scoped>
.error-actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 24px;
  text-align: center;
}
.error-message {
  font-size: 14px;
  color: var(--text-secondary, #6B7280);
  margin: 0;
  max-width: 320px;
}
.error-btn-group {
  display: flex;
  gap: 12px;
}
</style>
