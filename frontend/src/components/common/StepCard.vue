<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <div class="step-card glass-card">
    <div v-if="title" class="section-header">
      <h3 class="section-title">{{ title }}</h3>
      <p v-if="description" class="section-desc">{{ description }}</p>
    </div>
    <div class="step-card-body">
      <slot />
    </div>
  </div>
</template>

<style scoped lang="scss">
.step-card {
  padding: 24px;
  padding-top: 20px;
  margin-bottom: 20px;
}

.section-header {
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border);
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.section-desc {
  font-size: 13px;
  color: var(--text-tertiary);
  margin: 4px 0 0;
}
</style>
