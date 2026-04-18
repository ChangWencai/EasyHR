<template>
  <div class="si-dashboard">
    <div class="page-header">
      <h1 class="page-title">社保看板</h1>
    </div>

    <div v-loading="loading" class="dashboard-content">
      <div v-if="error" class="error-state">
        <el-empty description="加载社保数据失败，请刷新页面重试">
          <el-button type="primary" @click="loadDashboard">重新加载</el-button>
        </el-empty>
      </div>

      <div v-else class="stats-grid">
        <div
          v-for="card in statCards"
          :key="card.label"
          class="stat-card"
        >
          <div class="stat-value" :style="{ color: card.accent }">
            {{ card.value }}
          </div>
          <div class="stat-label">{{ card.label }}</div>
          <div v-if="card.trendPercent !== null" class="stat-trend" :class="card.trendClass">
            <span v-if="card.trendDirection === 'up'">&#8593;</span>
            <span v-else-if="card.trendDirection === 'down'">&#8595;</span>
            {{ card.trendDisplay }}%
          </div>
          <div v-else class="stat-trend neutral">--</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import axios from '@/api/request'

interface DashboardData {
  total: string
  total_trend_percent: number | null
  company: string
  company_trend_percent: number | null
  personal: string
  personal_trend_percent: number | null
  overdue: string
  overdue_trend_percent: number | null
}

interface StatCard {
  label: string
  value: string
  accent: string
  trendPercent: number | null
  trendDirection: 'up' | 'down' | null
  trendClass: string
  trendDisplay: string
}

const loading = ref(false)
const error = ref(false)
const data = ref<DashboardData | null>(null)

const statCards = computed<StatCard[]>(() => {
  if (!data.value) {
    return getEmptyCards()
  }

  return [
    buildCard('应缴总额（元）', data.value.total, '#4F6EF7', data.value.total_trend_percent, 'revenue'),
    buildCard('单位部分合计（元）', data.value.company, '#4F6EF7', data.value.company_trend_percent, 'revenue'),
    buildCard('个人部分合计（元）', data.value.personal, '#4F6EF7', data.value.personal_trend_percent, 'revenue'),
    buildCard('欠缴金额（元）', data.value.overdue, '#FF5630', data.value.overdue_trend_percent, 'overdue'),
  ]
})

function getEmptyCards(): StatCard[] {
  return [
    buildCard('应缴总额（元）', '0.00', '#4F6EF7', null, 'revenue'),
    buildCard('单位部分合计（元）', '0.00', '#4F6EF7', null, 'revenue'),
    buildCard('个人部分合计（元）', '0.00', '#4F6EF7', null, 'revenue'),
    buildCard('欠缴金额（元）', '0.00', '#FF5630', null, 'overdue'),
  ]
}

function buildCard(
  label: string,
  value: string,
  accent: string,
  trendPercent: number | null,
  metricType: 'revenue' | 'overdue'
): StatCard {
  if (trendPercent === null) {
    return { label, value, accent, trendPercent: null, trendDirection: null, trendClass: 'neutral', trendDisplay: '--' }
  }

  const absPercent = Math.abs(trendPercent)
  const isPositive = trendPercent > 0
  const direction = isPositive ? 'up' : 'down'

  let trendClass: string
  if (metricType === 'revenue') {
    trendClass = isPositive ? 'up' : 'down'
  } else {
    trendClass = isPositive ? 'down' : 'up'
  }

  return {
    label,
    value,
    accent,
    trendPercent,
    trendDirection: direction,
    trendClass,
    trendDisplay: absPercent.toFixed(1),
  }
}

async function loadDashboard() {
  loading.value = true
  error.value = false
  try {
    const res = await axios.get('/api/v1/socialinsurance/dashboard')
    const responseData = (res as { data?: DashboardData })?.data ?? (res as DashboardData)
    data.value = responseData as DashboardData
  } catch {
    error.value = true
    ElMessage.error('加载社保数据失败，请刷新页面重试')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadDashboard()
})
</script>

<style scoped lang="scss">
.si-dashboard {
  padding: 20px 24px;
  width: 100%;
  box-sizing: border-box;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-title {
  font-size: 16px;
  font-weight: 700;
  color: #1a1a1a;
  margin: 0;
  line-height: 1.2;
}

.dashboard-content {
  min-height: 120px;
}

.error-state {
  padding: 8px 0;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}

.stat-card {
  background: #fafbff;
  padding: 16px;
  border-radius: 12px;
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

.stat-trend {
  font-size: 12px;
  margin-top: 4px;
  font-weight: 500;

  &.up {
    color: #52c41a;
  }

  &.down {
    color: #ff4d4f;
  }

  &.neutral {
    color: #8c8c8c;
  }
}

@media (max-width: 900px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .si-dashboard {
    padding: 12px;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
