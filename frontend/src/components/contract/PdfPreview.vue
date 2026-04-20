<script setup lang="ts">
const props = defineProps<{
  url: string
  loading?: boolean
}>()
</script>

<template>
  <div class="pdf-preview">
    <div v-if="loading" class="pdf-loading">
      <el-skeleton :rows="3" animated />
      <div class="loading-text">合同生成中...</div>
    </div>
    <iframe
      v-else-if="url"
      :src="url"
      class="pdf-iframe"
      sandbox="allow-scripts allow-same-origin"
    />
    <div v-else class="pdf-empty">
      <div class="empty-icon">
        <svg width="40" height="40" viewBox="0 0 40 40" fill="none">
          <rect x="4" y="2" width="32" height="36" rx="2" stroke="#DCDFE6" stroke-width="2"/>
          <line x1="10" y1="12" x2="30" y2="12" stroke="#DCDFE6" stroke-width="2"/>
          <line x1="10" y1="20" x2="30" y2="20" stroke="#DCDFE6" stroke-width="2"/>
        </svg>
      </div>
      <p>合同生成失败，请重试</p>
    </div>
  </div>
</template>

<style scoped lang="scss">
.pdf-preview {
  border: 1px solid var(--border);
  border-radius: 12px;
  overflow: hidden;
  min-height: 400px;
  display: flex;
  flex-direction: column;
}

.pdf-iframe {
  width: 100%;
  height: 480px;
  border: none;
}

.pdf-loading {
  padding: 24px;
  .loading-text {
    text-align: center;
    color: var(--text-secondary);
    font-size: 14px;
    margin-top: 12px;
  }
}

.pdf-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: var(--text-secondary);
  font-size: 14px;
  gap: 12px;
}
</style>
