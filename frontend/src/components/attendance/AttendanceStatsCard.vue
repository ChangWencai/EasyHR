<template>
  <div class="stats-grid">
    <div v-for="card in cards" :key="card.label" class="stat-card">
      <div class="stat-value" :style="{ color: card.color }">{{ card.value }}</div>
      <div class="stat-label">{{ card.label }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  actual: number | string
  required: number | string
  overtime: number | string
  absent: number | string
}>()

const cards = computed(() => [
  { label: '实际出勤', value: `${props.actual} 天`, color: '#1677FF' },
  { label: '应出勤', value: `${props.required} 天`, color: '#8C8C8C' },
  { label: '加班时长', value: `${props.overtime} 小时`, color: '#E6A23C' },
  { label: '缺勤天数', value: `${props.absent} 天`, color: '#F56C6C' },
])
</script>

<style scoped lang="scss">
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin-bottom: 16px;
}

.stat-card {
  background: #fafafa;
  padding: 16px;
  border-radius: 8px;
  text-align: center;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  line-height: 1.2;
}

.stat-label {
  font-size: 12px;
  color: #8c8c8c;
  margin-top: 4px;
}

@media (max-width: 900px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
