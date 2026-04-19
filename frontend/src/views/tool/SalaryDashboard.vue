<template>
  <div class="salary-dashboard">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">薪资看板</h1>
        <p class="page-subtitle">实时掌握薪资发放情况</p>
      </div>
      <div class="header-actions">
        <el-date-picker
          v-model="selectedMonth"
          type="month"
          placeholder="选择月份"
          format="YYYY年MM月"
          value-format="YYYY-MM"
          :clearable="false"
          size="large"
          class="month-picker"
          @change="loadDashboard"
        />
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
      <h3>加载薪资数据失败</h3>
      <p>请检查网络连接后重试</p>
      <el-button type="primary" @click="loadDashboard">重新加载</el-button>
    </div>

    <!-- 空状态 -->
    <div v-else-if="isEmpty" class="empty-state glass-card">
      <div class="empty-icon">
        <el-icon><Money /></el-icon>
      </div>
      <h3>暂无薪资数据</h3>
      <p>请先创建工资表后再查看数据</p>
    </div>

    <!-- 统计卡片 -->
    <div v-else class="stats-grid">
      <div
        v-for="(stat, idx) in data?.stats"
        :key="stat.label"
        class="stat-card glass-card"
        :style="{ animationDelay: `${idx * 0.08}s` }"
      >
        <div class="stat-card-bg" :class="`bg--${idx}`"></div>
        <div class="stat-card-content">
          <div class="stat-header">
            <div class="stat-icon" :class="`icon--${idx}`">
              <el-icon><component :is="statIcons[idx] || 'Coin'" /></el-icon>
            </div>
            <div
              v-if="stat.trend_percent"
              class="stat-trend"
              :class="stat.trend_direction === 'up' ? 'trend--up' : 'trend--down'"
            >
              <el-icon v-if="stat.trend_direction === 'up'"><Top /></el-icon>
              <el-icon v-else><Bottom /></el-icon>
              {{ Math.abs(parseFloat(stat.trend_percent)) }}%
            </div>
            <div v-else class="stat-trend trend--neutral">—</div>
          </div>
          <div class="stat-value" :class="`value--${idx}`">{{ stat.value }}</div>
          <div class="stat-label">{{ stat.label }}</div>
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
import { WarningFilled, Money, Coin, Top, Bottom, TrendCharts, Timer, User } from '@element-plus/icons-vue'
import dayjs from 'dayjs'

// suppress unused import warnings - icons used via dynamic component
void Coin; void TrendCharts; void Timer; void User

const loading = ref(false)
const error = ref(false)
const data = ref<SalaryDashboardResponse | null>(null)
const selectedMonth = ref(dayjs().format('YYYY-MM'))

const statIcons = ['Coin', 'TrendCharts', 'Timer', 'User']

const isEmpty = computed(
  () => data.value !== null && data.value.stats.every((s) => s.value === '0.00' || s.value === '0'),
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
    data.value = res as SalaryDashboardResponse
  } catch {
    error.value = true
    ElMessage.error('加载薪资数据失败，请刷新页面重试')
  } finally {
    loading.value = false
  }
}

onMounted(() => loadDashboard())
</script>

<style scoped lang="scss">
$success: #10B981;
$warning: #F59E0B;
$error: #EF4444;
$bg-page: #FAFBFC;
$text-primary: #1F2937;
$text-secondary: #6B7280;
$text-muted: #9CA3AF;
$border-color: #E5E7EB;
$radius-sm: 8px;
$radius-md: 12px;
$radius-lg: 16px;
$radius-xl: 24px;

.salary-dashboard { padding: 24px 32px; width: 100%; box-sizing: border-box; background: $bg-page; min-height: 100vh; }

.glass-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: $radius-xl;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 24px;
  .page-title { font-size: 24px; font-weight: 700; color: $text-primary; margin: 0 0 4px; }
  .page-subtitle { font-size: 14px; color: $text-secondary; margin: 0; }
}

.header-actions { display: flex; align-items: center; gap: 12px; }
.month-picker { width: 180px; }

.skeleton-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; }

.skeleton-card {
  height: 160px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  border-radius: $radius-xl;
  animation: skeleton-loading 1.5s infinite;
}

@keyframes skeleton-loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.error-state {
  text-align: center;
  padding: 64px 32px;
  .error-icon { width: 64px; height: 64px; margin: 0 auto 16px; background: #FEE2E2; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 28px; color: $error; }
  h3 { font-size: 18px; font-weight: 600; color: $text-primary; margin: 0 0 8px; }
  p { font-size: 14px; color: $text-secondary; margin: 0 0 24px; }
}

.empty-state {
  text-align: center;
  padding: 64px 32px;
  .empty-icon { width: 72px; height: 72px; margin: 0 auto 16px; background: linear-gradient(135deg, #EDE9FE, #DDD6FE); border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 32px; color: var(--primary); }
  h3 { font-size: 18px; font-weight: 600; color: $text-primary; margin: 0 0 8px; }
  p { font-size: 14px; color: $text-muted; margin: 0; }
}

.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; }

.stat-card {
  position: relative;
  overflow: hidden;
  padding: 20px;
  transition: all 0.3s ease;
  animation: fadeInUp 0.4s ease both;

  &:hover { transform: translateY(-4px); box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1); }
}

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(16px); }
  to   { opacity: 1; transform: translateY(0); }
}

.stat-card-bg {
  position: absolute;
  top: -20px; right: -20px;
  width: 120px; height: 120px;
  border-radius: 50%;

  &.bg--0 { background: linear-gradient(135deg, rgba(var(--primary-light), 0.3), rgba(var(--primary), 0.1)); }
  &.bg--1 { background: linear-gradient(135deg, rgba($success, 0.3), rgba($success, 0.1)); }
  &.bg--2 { background: linear-gradient(135deg, rgba($warning, 0.3), rgba($warning, 0.1)); }
  &.bg--3 { background: linear-gradient(135deg, rgba(#3B82F6, 0.3), rgba(#3B82F6, 0.1)); }
}

.stat-card-content { position: relative; z-index: 1; }

.stat-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }

.stat-icon {
  width: 44px; height: 44px;
  border-radius: $radius-md;
  display: flex; align-items: center; justify-content: center;
  font-size: 20px;

  &.icon--0 { background: linear-gradient(135deg, #EDE9FE, #DDD6FE); color: var(--primary); }
  &.icon--1 { background: linear-gradient(135deg, #D1FAE5, #A7F3D0); color: $success; }
  &.icon--2 { background: linear-gradient(135deg, #FEF3C7, #FDE68A); color: $warning; }
  &.icon--3 { background: linear-gradient(135deg, #DBEAFE, #BFDBFE); color: #3B82F6; }
}

.stat-trend {
  display: inline-flex; align-items: center; gap: 2px;
  font-size: 12px; font-weight: 600; padding: 4px 8px; border-radius: 20px;

  .el-icon { font-size: 12px; }

  &.trend--up    { background: #D1FAE5; color: $success; }
  &.trend--down  { background: #FEE2E2; color: $error; }
  &.trend--neutral { background: #F3F4F6; color: $text-muted; }
}

.stat-value {
  font-size: 28px; font-weight: 700; line-height: 1.2; margin-bottom: 4px;
  font-family: 'SF Mono', Monaco, monospace;

  &.value--0 { color: var(--primary); }
  &.value--1 { color: $success; }
  &.value--2 { color: $warning; }
  &.value--3 { color: #3B82F6; }
}

.stat-label { font-size: 13px; color: $text-secondary; }

@media (max-width: 1024px) { .stats-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 768px) {
  .salary-dashboard { padding: 16px; }
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
  .skeleton-grid { grid-template-columns: repeat(2, 1fr); }
}
</style>
