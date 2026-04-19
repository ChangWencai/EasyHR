<template>
  <div class="page-view">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">员工看板</h1>
        <p class="page-subtitle">实时掌握团队人员变动</p>
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
      <h3>加载员工数据失败</h3>
      <p>请检查网络连接后重试</p>
      <el-button type="primary" @click="loadDashboard">重新加载</el-button>
    </div>

    <!-- 空状态 -->
    <div v-else-if="isEmpty" class="empty-state glass-card">
      <div class="empty-icon">
        <el-icon><User /></el-icon>
      </div>
      <h3>暂无员工数据</h3>
      <p>请先添加员工后再查看统计数据</p>
    </div>

    <!-- 统计卡片 -->
    <div v-else class="stats-grid">
      <div v-for="(stat, idx) in statCards" :key="stat.label" class="stat-card glass-card" :style="{ animationDelay: `${idx * 0.08}s` }">
        <div class="stat-bg" :class="`bg--${idx}`"></div>
        <div class="stat-content">
          <div class="stat-header">
            <div class="stat-icon" :class="`icon--${idx}`">
              <el-icon><component :is="stat.icon" /></el-icon>
            </div>
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
import { employeeApi } from '@/api/employee'
import type { EmployeeDashboard } from '@/api/employee'
import { Refresh, User, WarningFilled } from '@element-plus/icons-vue'

const loading = ref(false)
const error = ref(false)
const data = ref<EmployeeDashboard | null>(null)

const isEmpty = computed(
  () => data.value !== null && data.value.active_count === 0 && data.value.joined_this_month === 0 && data.value.left_this_month === 0,
)

const statCards = computed(() => data.value ? [
  { icon: 'User',        value: data.value.active_count,     label: '在职人数',    idx: 0 },
  { icon: 'UserPlus',    value: data.value.joined_this_month, label: '本月新入职',  idx: 1 },
  { icon: 'Delete', value: data.value.left_this_month,   label: '本月离职',    idx: 2 },
  { icon: 'TrendCharts', value: `${data.value.turnover_rate.toFixed(2)}%`, label: '当月离职率', idx: 3 },
] : [])

async function loadDashboard() {
  loading.value = true; error.value = false
  try {
    const res = await employeeApi.getDashboard()
    data.value = res
  } catch {
    error.value = true
    ElMessage.error('加载看板数据失败，请刷新页面重试')
  } finally {
    loading.value = false
  }
}

onMounted(() => loadDashboard())
</script>

<style scoped lang="scss">
.skeleton-grid { display: grid; grid-template-columns: repeat(4,1fr); gap: 16px; }
.skeleton-card { height: 160px; background: linear-gradient(90deg,#f0f0f0 25%,#e0e0e0 50%,#f0f0f0 75%); background-size: 200% 100%; border-radius: var(--radius-xl); animation: skeleton-loading 1.5s infinite; }
@keyframes skeleton-loading { 0%{background-position:200% 0} 100%{background-position:-200% 0} }

.error-state { text-align: center; padding: 64px 32px;
  .error-icon { width: 64px; height: 64px; margin: 0 auto 16px; background: #FEE2E2; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 28px; color: var(--danger); }
  h3 { font-size: 18px; font-weight: 600; color: var(--text-primary); margin: 0 0 8px; }
  p { font-size: 14px; color: var(--text-secondary); margin: 0 0 24px; }
}

.empty-state { text-align: center; padding: 64px 32px;
  .empty-icon { width: 72px; height: 72px; margin: 0 auto 16px; background: linear-gradient(135deg,#EDE9FE,#DDD6FE); border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 32px; color: var(--primary); }
  h3 { font-size: 18px; font-weight: 600; color: var(--text-primary); margin: 0 0 8px; }
  p { font-size: 14px; color: var(--text-tertiary); margin: 0; }
}

.stats-grid { display: grid; grid-template-columns: repeat(4,1fr); gap: 16px; }

.stat-card { position: relative; overflow: hidden; padding: 20px; transition: all 0.3s ease; animation: fadeInUp 0.4s ease both;
  &:hover { transform: translateY(-4px); box-shadow: var(--shadow-lg); }
}
@keyframes fadeInUp { from{opacity:0;transform:translateY(16px)} to{opacity:1;transform:translateY(0)} }

.stat-bg { position: absolute; top: -20px; right: -20px; width: 120px; height: 120px; border-radius: 50%;
  &.bg--0{background:linear-gradient(135deg,rgba(var(--primary-light),0.3),rgba(var(--primary),0.1))}
  &.bg--1{background:linear-gradient(135deg,rgba(var(--success),0.3),rgba(var(--success),0.1))}
  &.bg--2{background:linear-gradient(135deg,rgba(var(--danger),0.3),rgba(var(--danger),0.1))}
  &.bg--3{background:linear-gradient(135deg,rgba(var(--warning),0.3),rgba(var(--warning),0.1))}
}

.stat-content { position: relative; z-index: 1; }
.stat-header { margin-bottom: 16px; }
.stat-icon { width: 44px; height: 44px; border-radius: var(--radius-md); display: flex; align-items: center; justify-content: center; font-size: 20px;
  &.icon--0{background:linear-gradient(135deg,#EDE9FE,#DDD6FE);color:var(--primary)}
  &.icon--1{background:linear-gradient(135deg,#D1FAE5,#A7F3D0);color:var(--success)}
  &.icon--2{background:linear-gradient(135deg,#FEE2E2,#FECACA);color:var(--danger)}
  &.icon--3{background:linear-gradient(135deg,#FEF3C7,#FDE68A);color:var(--warning)}
}

.stat-value { font-size: 32px; font-weight: 700; line-height: 1.2; margin-bottom: 4px; font-family: 'SF Mono',Monaco,monospace;
  &.value--0{color:var(--primary)} &.value--1{color:var(--success)} &.value--2{color:var(--danger)} &.value--3{color:var(--warning)}
}
.stat-label { font-size: 13px; color: var(--text-secondary); }

@media (max-width:1024px){.stats-grid{grid-template-columns:repeat(2,1fr)}}
@media (max-width:768px){.stats-grid,.skeleton-grid{grid-template-columns:repeat(2,1fr)}}
</style>
