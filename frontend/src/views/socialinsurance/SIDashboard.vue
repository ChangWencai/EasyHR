<template>
  <div class="page-view">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">社保看板</h1>
        <p class="page-subtitle">实时掌握社保缴纳情况</p>
      </div>
      <div class="header-actions">
        <el-button @click="loadDashboard">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </header>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-state">
      <div class="skeleton-grid">
        <div v-for="i in 4" :key="i" class="skeleton-card"></div>
      </div>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="error-state glass-card">
      <div class="error-icon">
        <el-icon><WarningFilled /></el-icon>
      </div>
      <h3>加载社保数据失败</h3>
      <p>请检查网络连接后重试</p>
      <el-button type="primary" @click="loadDashboard">重新加载</el-button>
    </div>

    <!-- 统计卡片 -->
    <div v-else class="stats-grid">
      <!-- 应缴总额 -->
      <div class="stat-card glass-card stat-card--primary">
        <div class="stat-card-bg"></div>
        <div class="stat-card-content">
          <div class="stat-header">
            <div class="stat-icon stat-icon--primary">
              <el-icon><Coin /></el-icon>
            </div>
            <div class="stat-trend" :class="getTrendClass(data?.total_trend_percent, 'revenue')">
              <el-icon v-if="getTrendDirection(data?.total_trend_percent) === 'up'"><Top /></el-icon>
              <el-icon v-else-if="getTrendDirection(data?.total_trend_percent) === 'down'"><Bottom /></el-icon>
              <span v-if="data?.total_trend_percent">{{ Math.abs(Number(data.total_trend_percent)).toFixed(1) }}%</span>
              <span v-else>--</span>
            </div>
          </div>
          <div class="stat-value stat-value--primary">{{ data?.total || '0.00' }}</div>
          <div class="stat-label">应缴总额（元）</div>
        </div>
      </div>

      <!-- 单位部分 -->
      <div class="stat-card glass-card">
        <div class="stat-card-content">
          <div class="stat-header">
            <div class="stat-icon stat-icon--company">
              <el-icon><OfficeBuilding /></el-icon>
            </div>
            <div class="stat-trend" :class="getTrendClass(data?.company_trend_percent, 'revenue')">
              <el-icon v-if="getTrendDirection(data?.company_trend_percent) === 'up'"><Top /></el-icon>
              <el-icon v-else-if="getTrendDirection(data?.company_trend_percent) === 'down'"><Bottom /></el-icon>
              <span v-if="data?.company_trend_percent">{{ Math.abs(Number(data.company_trend_percent)).toFixed(1) }}%</span>
              <span v-else>--</span>
            </div>
          </div>
          <div class="stat-value">{{ data?.company || '0.00' }}</div>
          <div class="stat-label">单位部分合计（元）</div>
        </div>
      </div>

      <!-- 个人部分 -->
      <div class="stat-card glass-card">
        <div class="stat-card-content">
          <div class="stat-header">
            <div class="stat-icon stat-icon--personal">
              <el-icon><User /></el-icon>
            </div>
            <div class="stat-trend" :class="getTrendClass(data?.personal_trend_percent, 'revenue')">
              <el-icon v-if="getTrendDirection(data?.personal_trend_percent) === 'up'"><Top /></el-icon>
              <el-icon v-else-if="getTrendDirection(data?.personal_trend_percent) === 'down'"><Bottom /></el-icon>
              <span v-if="data?.personal_trend_percent">{{ Math.abs(Number(data.personal_trend_percent)).toFixed(1) }}%</span>
              <span v-else>--</span>
            </div>
          </div>
          <div class="stat-value">{{ data?.personal || '0.00' }}</div>
          <div class="stat-label">个人部分合计（元）</div>
        </div>
      </div>

      <!-- 欠缴金额 -->
      <div class="stat-card glass-card stat-card--danger">
        <div class="stat-card-bg stat-card-bg--danger"></div>
        <div class="stat-card-content">
          <div class="stat-header">
            <div class="stat-icon stat-icon--danger">
              <el-icon><WarningFilled /></el-icon>
            </div>
            <div v-if="data?.overdue_trend_percent" class="stat-trend stat-trend--danger">
              <el-icon><Bottom /></el-icon>
              <span>{{ Math.abs(Number(data.overdue_trend_percent)).toFixed(1) }}%</span>
            </div>
          </div>
          <div class="stat-value stat-value--danger">{{ data?.overdue || '0.00' }}</div>
          <div class="stat-label">欠缴金额（元）</div>
        </div>
      </div>
    </div>

    <!-- 底部图表占位 -->
    <div class="charts-row">
      <div class="chart-card glass-card">
        <div class="chart-header">
          <h3 class="chart-title">社保构成</h3>
          <span class="chart-subtitle">单位 vs 个人缴费比例</span>
        </div>
        <div class="chart-placeholder">
          <div class="placeholder-content">
            <el-icon class="placeholder-icon"><PieChart /></el-icon>
            <span>图表加载中...</span>
          </div>
        </div>
      </div>
      <div class="chart-card glass-card">
        <div class="chart-header">
          <h3 class="chart-title">月度趋势</h3>
          <span class="chart-subtitle">近6个月社保缴纳趋势</span>
        </div>
        <div class="chart-placeholder">
          <div class="placeholder-content">
            <el-icon class="placeholder-icon"><TrendCharts /></el-icon>
            <span>图表加载中...</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { siApi } from '@/api/socialinsurance'
import type { SIDashboardData } from '@/api/socialinsurance'
import { Refresh, Coin, OfficeBuilding, User, WarningFilled, Top, Bottom, PieChart, TrendCharts } from '@element-plus/icons-vue'

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

const loading = ref(false)
const error = ref(false)
const data = ref<SIDashboardData | null>(null)

function getTrendDirection(val: string | number | null | undefined): 'up' | 'down' | null {
  if (val === null || val === undefined) return null
  const n = typeof val === 'string' ? parseFloat(val) : Number(val)
  if (isNaN(n)) return null
  return n > 0 ? 'up' : n < 0 ? 'down' : null
}

function getTrendClass(val: string | number | null | undefined, type: 'revenue' | 'overdue'): string {
  const direction = getTrendDirection(val)
  if (!direction) return ''
  if (type === 'revenue') {
    return direction === 'up' ? 'stat-trend--up' : 'stat-trend--down'
  } else {
    return direction === 'down' ? 'stat-trend--up' : 'stat-trend--down'
  }
}

async function loadDashboard() {
  loading.value = true
  error.value = false
  try {
    data.value = await siApi.dashboard()
  } catch {
    error.value = true
    ElMessage.error('加载社保数据失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadDashboard()
})
</script>

<style scoped lang="scss">
.loading-state {
  margin-bottom: 24px;
}

.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.skeleton-card {
  height: 160px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  border-radius: var(--radius-xl);
  animation: skeleton-loading 1.5s infinite;
}

@keyframes skeleton-loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.error-state {
  text-align: center;
  padding: 64px 32px;

  .error-icon {
    width: 64px;
    height: 64px;
    margin: 0 auto 16px;
    background: #FEE2E2;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
    color: var(--danger);
  }

  h3 {
    font-size: 18px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 8px;
  }

  p {
    font-size: 14px;
    color: var(--text-secondary);
    margin: 0 0 24px;
  }
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  position: relative;
  overflow: hidden;
  padding: 20px;
  transition: all 0.3s ease;

  &:hover {
    transform: translateY(-4px);
    box-shadow: var(--shadow-lg);
  }

  &--primary {
    .stat-card-bg {
      position: absolute;
      top: -20px;
      right: -20px;
      width: 120px;
      height: 120px;
      background: linear-gradient(135deg, rgba(var(--primary-light), 0.3) 0%, rgba(var(--primary), 0.1) 100%);
      border-radius: 50%;
    }
  }

  &--danger {
    .stat-card-bg--danger {
      position: absolute;
      top: -20px;
      right: -20px;
      width: 120px;
      height: 120px;
      background: linear-gradient(135deg, rgba(var(--danger), 0.3) 0%, rgba(var(--danger), 0.1) 100%);
      border-radius: 50%;
    }
  }
}

.stat-card-content {
  position: relative;
  z-index: 1;
}

.stat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.stat-icon {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;

  &--primary { background: linear-gradient(135deg, #EDE9FE 0%, #DDD6FE 100%); color: var(--primary); }
  &--company { background: linear-gradient(135deg, #DBEAFE 0%, #BFDBFE 100%); color: #3B82F6; }
  &--personal { background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%); color: var(--warning); }
  &--danger { background: linear-gradient(135deg, #FEE2E2 0%, #FECACA 100%); color: var(--danger); }
}

.stat-trend {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  font-size: 12px;
  font-weight: 600;
  padding: 4px 8px;
  border-radius: 20px;

  .el-icon { font-size: 12px; }

  &--up {
    background: #D1FAE5;
    color: var(--success);
  }

  &--down {
    background: #FEE2E2;
    color: var(--danger);
  }

  &--danger {
    background: #FEE2E2;
    color: var(--danger);
  }
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.2;
  margin-bottom: 4px;
  font-family: 'SF Mono', Monaco, monospace;

  &--primary { color: var(--primary); }
  &--danger { color: var(--danger); }
}

.stat-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.charts-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.chart-card {
  padding: 24px;
}

.chart-header {
  margin-bottom: 20px;
}

.chart-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 4px;
}

.chart-subtitle {
  font-size: 13px;
  color: var(--text-tertiary);
}

.chart-placeholder {
  height: 200px;
  background: var(--bg-page);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
}

.placeholder-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: var(--text-tertiary);
  font-size: 14px;
}

.placeholder-icon {
  font-size: 32px;
  opacity: 0.5;
}

@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .si-dashboard { padding: 16px; }
  .stats-grid { grid-template-columns: 1fr; }
  .charts-row { grid-template-columns: 1fr; }
  .skeleton-grid { grid-template-columns: 1fr; }
}
</style>
