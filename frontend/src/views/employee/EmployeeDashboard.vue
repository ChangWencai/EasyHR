<template>
  <div class="employee-dashboard">
    <div class="page-header">
      <h1 class="page-title">员工看板</h1>
    </div>

    <div v-loading="loading" class="dashboard-content">
      <div v-if="error" class="error-state">
        <el-empty description="加载看板数据失败，请刷新页面重试">
          <el-button type="primary" @click="loadDashboard">重新加载</el-button>
        </el-empty>
      </div>

      <div v-else-if="isEmpty" class="empty-state">
        <el-empty description="暂无员工数据 -- 请先添加员工" />
      </div>

      <div v-else class="stats-grid">
        <div class="stat-card">
          <div class="stat-value">{{ data?.active_count ?? 0 }}</div>
          <div class="stat-label">在职人数</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ data?.joined_this_month ?? 0 }}</div>
          <div class="stat-label">本月新入职</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ data?.left_this_month ?? 0 }}</div>
          <div class="stat-label">本月离职</div>
        </div>
        <div class="stat-card">
          <div class="stat-value turnover">{{ formatRate(data?.turnover_rate ?? 0) }}%</div>
          <div class="stat-label">当月离职率</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { employeeApi } from '@/api/employee'
import type { EmployeeDashboard } from '@/api/employee'

const loading = ref(false)
const error = ref(false)
const data = ref<EmployeeDashboard | null>(null)

const isEmpty = computed(
  () =>
    data.value !== null &&
    data.value.active_count === 0 &&
    data.value.joined_this_month === 0 &&
    data.value.left_this_month === 0,
)

function formatRate(rate: number): string {
  return rate.toFixed(2)
}

async function loadDashboard() {
  loading.value = true
  error.value = false
  try {
    const res = await employeeApi.getDashboard()
    data.value = res.data ?? res
  } catch {
    error.value = true
    ElMessage.error('加载看板数据失败，请刷新页面重试')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadDashboard()
})
</script>

<style scoped lang="scss">
.employee-dashboard {
  padding: 20px 24px;
  width: 100%;
  box-sizing: border-box;
}

.page-header {
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

  &.turnover {
    color: #f56c6c;
  }
}

.stat-label {
  font-size: 12px;
  color: #8c8c8c;
  margin-top: 4px;
}

// 响应式断点
@media (max-width: 900px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .employee-dashboard {
    padding: 12px;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
