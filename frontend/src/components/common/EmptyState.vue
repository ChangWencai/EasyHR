<script setup lang="ts">
defineProps<{
  title: string
  description?: string
  actionText?: string
  actionRoute?: string
}>()
</script>

<template>
  <div class="empty-state">
    <!-- SVG illustration slot — parent passes contextual SVG or uses default -->
    <div class="empty-illustration">
      <slot name="illustration">
        <!-- Default: simple employee silhouette SVG, 120x120 -->
        <svg width="120" height="120" viewBox="0 0 120 120" fill="none" xmlns="http://www.w3.org/2000/svg">
          <circle cx="60" cy="40" r="20" fill="#E8EEFF"/>
          <path d="M30 100c0-16.569 13.431-30 30-30s30 13.431 30 30" fill="#E8EEFF"/>
        </svg>
      </slot>
    </div>
    <h3 class="empty-title">{{ title }}</h3>
    <p v-if="description" class="empty-desc">{{ description }}</p>
    <el-button
      v-if="actionText && actionRoute"
      type="primary"
      @click="$router.push(actionRoute)"
    >{{ actionText }}</el-button>
    <!-- Fallback: expose a named slot for custom action -->
    <slot name="action" />
  </div>
</template>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 48px 24px;
  text-align: center;
}
.empty-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #374151);
  margin: 0;
}
.empty-desc {
  font-size: 14px;
  color: var(--text-secondary, #6B7280);
  margin: 0;
  max-width: 280px;
}
</style>
