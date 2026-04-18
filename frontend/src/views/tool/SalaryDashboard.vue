<template>
  <div class="salary-dashboard">
    <div class="page-header">
      <h1 class="page-title">薪资看板</h1>
      <el-date-picker
        v-model="selectedMonth"
        type="month"
        placeholder="选择月份"
        format="YYYY年MM月"
        value-format="YYYY-MM"
        :clearable="false"
        @change="loadDashboard"
      />
    </div>

    <div v-loading="loading" class="dashboard-content">
      <div v-if="error" class="error-state">
        <el-empty description="加载薪资数据失败，请刷新页面重试">
          <el-button type="primary" @click="loadDashboard">重新加载</el-button>
        </el-empty>
      </div>

      <div v-else-if="isEmpty" class="empty-state">
        <el-empty description="暂无薪资数据 -- 请先创建工资表" />
      </div>

      <div v-else class="stats-grid">
        <div v-for="stat in data?.stats" :key="stat.label" class="stat-card">
          <div class="stat-value">{{ stat.value }}</div>
          <div class="stat-label">{{ stat.label }}</div>
          <div v-if="stat.trend_percent" class="stat-trend" :class="stat.trend_direction">
            <span v-if="stat.trend_direction === 'up'">&#8593;</span>
            <span v-else-if="stat.trend_direction === 'down'">&#8595;</span>
            {{ Math.abs(parseFloat(stat.trend_percent)) }}%
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
import { salaryApi } from '@/api/salary'
import type { SalaryDashboardResponse } from '@/api/salary'
import dayjs from 'dayjs'

const loading = ref(false)
const error = ref(false)
const data = ref<SalaryDashboardResponse | null>(null)
const selectedMonth = ref(dayjs().format('YYYY-MM'))

const isEmpty = computed(
  () =>
    data.value !== null &&
    data.value.stats.every(
      (s) => s.value === '0.00' || s.value === '0',
    ),
)

async function loadDashboard() {
  if (!selectedMonth.value) return

  const [yearStr, monthStr] = selectedMonth.value.split('-')
  const year = parseInt(yearStr, 10)
  const month = parseInt(monthStr, 10)

  loading.value = true
  error.value = false
  try {
    const res = await salaryApi.getSalaryDashboard(year, month)
    data.value = (res as { data?: SalaryDashboardResponse })?.data ?? (res as SalaryDashboardResponse)
  } catch {
    error.value = true
    ElMessage.error('加载薪资数据失败，请刷新页面重试')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadDashboard()
})
</script>

<style scoped lang="scss">
.salary-dashboard {
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

.error-state,
.empty-state {
  padding: 8px 0;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
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
  color: #1677ff;
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

// 响应式断点
@media (max-width: 900px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .salary-dashboard {
    padding: 12px;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
