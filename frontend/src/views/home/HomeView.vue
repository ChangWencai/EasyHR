<template>
  <div class="home-view">
    <!-- Tour Overlay -->
    <TourOverlay
      v-model:visible="showTour"
      :steps="tourSteps"
      @complete="completeTour"
    />

    <!-- 页面标题区 - 现代化头部 -->
    <header class="page-header">
      <div class="header-content">
        <div class="greeting-section">
          <h1 class="page-title">{{ greeting }}，欢迎回来</h1>
          <p class="page-subtitle">您有 {{ store.todos.length }} 项待办事项需要处理</p>
        </div>
      </div>
    </header>

    <!-- 完成率环形图 - 玻璃态卡片 -->
    <section class="ring-chart-section">
      <div class="glass-card ring-chart-card">
        <div class="ring-chart-row">
          <div class="ring-chart-item">
            <TodoRingChart type="all" />
            <div class="ring-chart-info">
              <span class="ring-chart-label">全部事项</span>
              <span class="ring-chart-desc">已完成 {{ completedRate.all }}%</span>
            </div>
          </div>
          <div class="ring-divider"></div>
          <div class="ring-chart-item">
            <TodoRingChart type="time-limited" />
            <div class="ring-chart-info">
              <span class="ring-chart-label">限时任务</span>
              <span class="ring-chart-desc">已完成 {{ completedRate.timeLimited }}%</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 轮播图公告 -->
    <HomeCarousel />

    <!-- 待办事项 - 现代化卡片网格 -->
    <section class="todo-section" data-tour="todo-section">
      <div class="section-header">
        <div class="section-title-group">
          <h2 class="section-title">待办事项</h2>
          <el-badge :value="store.todos.length" type="primary" :hidden="store.todos.length === 0" />
        </div>
      </div>

      <div v-if="store.loading" class="loading-state">
        <div class="skeleton-grid">
          <div v-for="i in 4" :key="i" class="skeleton-card"></div>
        </div>
      </div>
      <div v-else-if="store.todos.length === 0" class="empty-state">
        <div class="empty-icon">
          <el-icon><CircleCheck /></el-icon>
        </div>
        <h3 class="empty-title">太棒了！</h3>
        <p class="empty-desc">暂无待办事项，轻松搞定人事~</p>
      </div>
      <div v-else class="todo-grid">
        <div
          v-for="(todo, index) in store.todos"
          :key="todo.type"
          class="todo-card"
          :style="{ animationDelay: `${index * 50}ms` }"
          @click="handleTodoClick(todo)"
        >
          <div class="todo-card-content">
            <div class="todo-icon" :class="`todo-icon--${todo.type}`">
              <el-icon size="24"><component :is="getTodoIcon(todo.type)" /></el-icon>
            </div>
            <div class="todo-info">
              <h4 class="todo-title">{{ todo.title }}</h4>
              <p v-if="todo.deadline" class="todo-deadline">
                <el-icon><Clock /></el-icon>
                {{ todo.deadline }}
              </p>
            </div>
            <div class="todo-meta">
              <span class="todo-count">{{ todo.count }}</span>
              <el-icon class="todo-arrow"><ArrowRight /></el-icon>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 快捷入口 + 数据概览 -->
    <div class="bottom-row">
      <!-- 快捷入口 -->
      <section class="shortcuts-section glass-card">
        <div class="section-header">
          <h2 class="section-title">快捷入口</h2>
        </div>
        <div class="shortcuts-grid">
          <router-link
            v-for="(item, index) in gridItems"
            :key="item.path + item.label"
            :to="item.path"
            class="shortcut-item"
            :data-tour="item.dataTour"
            :style="{ animationDelay: `${index * 30}ms` }"
          >
            <div class="shortcut-icon" :style="{ background: item.bg }">
              <el-icon size="24" :color="item.color">
                <component :is="item.icon" />
              </el-icon>
            </div>
            <span class="shortcut-label">{{ item.label }}</span>
          </router-link>
        </div>
      </section>

      <!-- 数据概览 -->
      <section class="overview-section glass-card">
        <div class="section-header" @click="store.toggleOverview">
          <h2 class="section-title">数据概览</h2>
          <el-icon class="toggle-icon" :class="{ 'is-expanded': store.overviewExpanded }">
            <ArrowDown />
          </el-icon>
        </div>
        <div v-if="store.overviewExpanded && store.overview" class="overview-grid">
          <div class="overview-item">
            <div class="overview-value">{{ store.overview.employee_count }}</div>
            <div class="overview-label">在职员工</div>
          </div>
          <div class="overview-item">
            <div class="overview-value inflow">
              <span class="value-positive">+{{ store.overview.joined_this_month }}</span>
              <span class="value-divider">/</span>
              <span class="value-negative">-{{ store.overview.left_this_month }}</span>
            </div>
            <div class="overview-label">本月入/离职</div>
          </div>
          <div class="overview-item">
            <div class="overview-value">¥{{ formatNumber(Number(store.overview.social_insurance_total)) }}</div>
            <div class="overview-label">本月社保</div>
          </div>
          <div class="overview-item">
            <div class="overview-value">¥{{ formatNumber(Number(store.overview.payroll_total)) }}</div>
            <div class="overview-label">本月工资</div>
          </div>
        </div>
        <div v-else-if="store.overviewExpanded" class="overview-loading">
          <el-icon class="is-loading"><Loading /></el-icon>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  User,
  Umbrella,
  Money,
  Document,
  Wallet,
  Plus,
  ArrowRight,
  ArrowDown,
  Clock,
  Loading,
  TrendCharts,
  CircleCheck,
  FolderChecked,
  Tickets,
  Calendar,
  CreditCard,
  Ticket,
  OfficeBuilding,
} from '@element-plus/icons-vue'
import { useDashboardStore } from '@/stores/dashboard'
import type { TodoItem } from '@/api/dashboard'
import TodoRingChart from './components/TodoRingChart.vue'
import HomeCarousel from './components/HomeCarousel.vue'
import TourOverlay, { type TourStep } from '@/components/common/TourOverlay.vue'

const store = useDashboardStore()
const router = useRouter()
const TOUR_DONE_KEY = 'hasSeenTour'
const showTour = ref(!localStorage.getItem(TOUR_DONE_KEY))

const todoCount = computed(() => store.todos?.length ?? 0)

const tourSteps = computed<TourStep[]>(() => [
  {
    title: '新增员工',
    body: '点击这里快速添加新员工，3步完成入职',
    target: '[data-tour="new-employee"]',
  },
  {
    title: '待办事项',
    body: `您有 ${todoCount.value} 项待处理事项，记得及时处理`,
    target: '[data-tour="todo-section"]',
  },
  {
    title: '快速上手',
    body: '60秒内完成您的第一个人事任务，开始使用吧',
    target: undefined,
  },
])

function completeTour() {
  showTour.value = false
}

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 12) return '上午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const completedRate = computed(() => {
  const all = store.todos.length
  // 模拟数据，实际应从 store 获取
  return {
    all: all > 0 ? Math.round((all - store.todos.filter(t => t.deadline).length) / all * 100) : 100,
    timeLimited: all > 0 ? Math.round(store.todos.filter(t => !t.deadline).length / all * 100) : 100
  }
})

function formatNumber(num: number): string {
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return num.toLocaleString()
}

function getTodoIcon(type: string) {
  const iconMap: Record<string, any> = {
    social_insurance: Umbrella,
    tax: Document,
    employee: User,
    contract: Ticket,
    expense: Tickets,
    voucher: CreditCard,
  }
  return iconMap[type] || FolderChecked
}

const gridItems = [
  { path: '/employee', label: '员工管理', icon: User, color: '#7C3AED', bg: 'linear-gradient(135deg, #EDE9FE 0%, #DDD6FE 100%)' },
  { path: '/employee/org-chart', label: '组织架构', icon: OfficeBuilding, color: '#8B5CF6', bg: 'linear-gradient(135deg, #EDE9FE 0%, #DDD6FE 100%)' },
  { path: '/tool/salary', label: '薪资管理', icon: Money, color: '#10B981', bg: 'linear-gradient(135deg, #D1FAE5 0%, #A7F3D0 100%)' },
  { path: '/tool/socialinsurance', label: '社保管理', icon: Umbrella, color: '#F59E0B', bg: 'linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%)' },
  { path: '/tool/tax', label: '个税申报', icon: Document, color: '#3B82F6', bg: 'linear-gradient(135deg, #DBEAFE 0%, #BFDBFE 100%)' },
  { path: '/finance/vouchers', label: '凭证管理', icon: Wallet, color: '#06B6D4', bg: 'linear-gradient(135deg, #CFFAFE 0%, #A5F3FC 100%)' },
  { path: '/finance/invoices', label: '发票管理', icon: Tickets, color: '#EC4899', bg: 'linear-gradient(135deg, #FCE7F3 0%, #FBCFE8 100%)' },
  { path: '/employee/create', label: '新入职', icon: Plus, color: '#8B5CF6', bg: 'linear-gradient(135deg, #EDE9FE 0%, #DDD6FE 100%)', dataTour: 'new-employee' },
  { path: '/tool/salary', label: '调薪', icon: TrendCharts, color: '#059669', bg: 'linear-gradient(135deg, #D1FAE5 0%, #A7F3D0 100%)' },
  { path: '/attendance/clock-live', label: '考勤打卡', icon: Calendar, color: '#6366F1', bg: 'linear-gradient(135deg, #E0E7FF 0%, #C7D2FE 100%)' },
]

function handleTodoClick(todo: TodoItem) {
  store.removeTodo(todo.type)
  const routeMap: Record<string, string> = {
    social_insurance: '/tool/socialinsurance',
    tax: '/tool/tax',
    employee: '/employee',
    contract: '/employee',
    expense: '/finance/expenses',
    voucher: '/finance/vouchers',
  }
  router.push(routeMap[todo.type] || '/home')
}

onMounted(() => {
  store.load()
})
</script>

<style scoped lang="scss">
// ============================================================
// 变量定义
// ============================================================
$success: #10B981;
$bg-page: #FAFBFC;
$bg-surface: #FFFFFF;
$text-primary: #1F2937;
$text-secondary: #6B7280;
$text-muted: #9CA3AF;
$border-color: #E5E7EB;
$success: #10B981;
$warning: #F59E0B;
$error: #EF4444;
$shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
$shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
$shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
$shadow-glow: 0 0 20px rgba(124, 58, 237, 0.3);
$radius-sm: 8px;
$radius-md: 12px;
$radius-lg: 16px;
$radius-xl: 24px;

// ============================================================
// 基础布局
// ============================================================
.home-view {
  padding: 24px 32px;
  width: 100%;
  box-sizing: border-box;
  background: $bg-page;
  min-height: 100vh;
}

// ============================================================
// 页面头部
// ============================================================
.page-header {
  margin-bottom: 24px;
}

.header-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.greeting-section {
  .page-title {
    font-size: 28px;
    font-weight: 700;
    color: $text-primary;
    margin: 0 0 4px;
    letter-spacing: -0.5px;
  }

  .page-subtitle {
    font-size: 14px;
    color: $text-secondary;
    margin: 0;
  }
}

// ============================================================
// 玻璃态卡片
// ============================================================
.glass-card {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: $radius-xl;
  box-shadow: $shadow-md;
  transition: all 0.3s ease;

  &:hover {
    box-shadow: $shadow-lg;
  }
}

// ============================================================
// 环形图区域
// ============================================================
.ring-chart-section {
  margin-bottom: 24px;
}

.ring-chart-card {
  padding: 24px 32px;
}

.ring-chart-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 48px;
}

.ring-chart-item {
  display: flex;
  align-items: center;
  gap: 16px;
}

.ring-chart-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.ring-chart-label {
  font-size: 15px;
  font-weight: 600;
  color: $text-primary;
}

.ring-chart-desc {
  font-size: 13px;
  color: $text-secondary;
}

.ring-divider {
  width: 1px;
  height: 60px;
  background: $border-color;
}

// ============================================================
// 区块通用
// ============================================================
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.section-title-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: $text-primary;
  margin: 0;
}

.toggle-icon {
  color: $text-muted;
  transition: transform 0.3s ease;

  &.is-expanded {
    transform: rotate(180deg);
  }
}

// ============================================================
// 待办事项
// ============================================================
.todo-section {
  margin-bottom: 24px;
}

.loading-state {
  padding: 16px 0;
}

.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.skeleton-card {
  height: 80px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  border-radius: $radius-md;
  animation: skeleton-loading 1.5s infinite;
}

@keyframes skeleton-loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.empty-state {
  text-align: center;
  padding: 48px 24px;
  background: $bg-surface;
  border-radius: $radius-xl;
  border: 1px dashed $border-color;

  .empty-icon {
    width: 64px;
    height: 64px;
    margin: 0 auto 16px;
    background: linear-gradient(135deg, #D1FAE5 0%, #A7F3D0 100%);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;

    .el-icon {
      font-size: 32px;
      color: $success;
    }
  }

  .empty-title {
    font-size: 18px;
    font-weight: 600;
    color: $text-primary;
    margin: 0 0 8px;
  }

  .empty-desc {
    font-size: 14px;
    color: $text-secondary;
    margin: 0;
  }
}

.todo-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.todo-card {
  background: $bg-surface;
  border-radius: $radius-lg;
  border: 1px solid $border-color;
  cursor: pointer;
  transition: all 0.2s ease-out;
  animation: fadeInUp 0.4s ease-out backwards;

  &:hover {
    transform: translateY(-4px);
    box-shadow: $shadow-lg;
    border-color: var(--primary-light);

    .todo-arrow {
      color: var(--primary);
      transform: translateX(4px);
    }

    .todo-icon {
      transform: scale(1.1);
    }
  }
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.todo-card-content {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
}

.todo-icon {
  width: 48px;
  height: 48px;
  border-radius: $radius-md;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: transform 0.2s ease;

  &--social_insurance { background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%); color: $warning; }
  &--tax { background: linear-gradient(135deg, #DBEAFE 0%, #BFDBFE 100%); color: #3B82F6; }
  &--employee { background: linear-gradient(135deg, #EDE9FE 0%, #DDD6FE 100%); color: var(--primary); }
  &--contract { background: linear-gradient(135deg, #FCE7F3 0%, #FBCFE8 100%); color: #EC4899; }
  &--expense { background: linear-gradient(135deg, #CFFAFE 0%, #A5F3FC 100%); color: #06B6D4; }
  &--voucher { background: linear-gradient(135deg, #D1FAE5 0%, #A7F3D0 100%); color: $success; }
}

.todo-info {
  flex: 1;
  min-width: 0;
}

.todo-title {
  font-size: 14px;
  font-weight: 600;
  color: $text-primary;
  margin: 0 0 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.todo-deadline {
  font-size: 12px;
  color: $text-muted;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 4px;

  .el-icon {
    font-size: 12px;
  }
}

.todo-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.todo-count {
  min-width: 24px;
  height: 24px;
  padding: 0 8px;
  background: var(--primary);
  color: white;
  font-size: 12px;
  font-weight: 600;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.todo-arrow {
  color: $text-muted;
  transition: all 0.2s ease;
}

// ============================================================
// 快捷入口 + 数据概览
// ============================================================
.bottom-row {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 24px;
}

.shortcuts-section {
  padding: 24px;
}

.shortcuts-grid {
  display: grid;
  grid-template-columns: repeat(9, 1fr);
  gap: 8px;
}

.shortcut-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 20px 8px;
  text-decoration: none;
  border-radius: $radius-md;
  transition: all 0.2s ease-out;
  cursor: pointer;
  animation: fadeInUp 0.4s ease-out backwards;

  &:hover {
    transform: translateY(-4px);
    box-shadow: $shadow-md;
    background: rgba(124, 58, 237, 0.04);

    .shortcut-icon {
      transform: scale(1.1);
    }
  }
}

.shortcut-icon {
  width: 56px;
  height: 56px;
  border-radius: $radius-lg;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.2s ease;
  box-shadow: $shadow-sm;
}

.shortcut-label {
  font-size: 13px;
  font-weight: 500;
  color: $text-secondary;
  text-align: center;
  white-space: nowrap;
}

.overview-section {
  padding: 24px;
}

.overview-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.overview-item {
  text-align: center;
  padding: 20px 12px;
  background: $bg-page;
  border-radius: $radius-md;
  transition: all 0.2s ease;

  &:hover {
    background: rgba(124, 58, 237, 0.04);
  }
}

.overview-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--primary);
  line-height: 1.2;
  margin-bottom: 4px;

  &.inflow {
    font-size: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;

    .value-positive { color: $success; }
    .value-divider { color: $text-muted; font-weight: 400; }
    .value-negative { color: $error; }
  }
}

.overview-label {
  font-size: 12px;
  color: $text-secondary;
}

.overview-loading {
  display: flex;
  justify-content: center;
  padding: 32px;
  color: var(--primary);
}

// ============================================================
// 响应式断点
// ============================================================
@media (max-width: 1600px) {
  .todo-grid {
    grid-template-columns: repeat(3, 1fr);
  }
  .shortcuts-grid {
    grid-template-columns: repeat(6, 1fr);
  }
}

@media (max-width: 1200px) {
  .bottom-row {
    grid-template-columns: 1fr;
  }
  .shortcuts-grid {
    grid-template-columns: repeat(6, 1fr);
  }
  .ring-chart-row {
    flex-direction: column;
    gap: 24px;
  }
  .ring-divider {
    width: 60px;
    height: 1px;
  }
}

@media (max-width: 900px) {
  .todo-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .shortcuts-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .home-view {
    padding: 16px;
  }

  .header-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .page-title {
    font-size: 22px !important;
  }

  .todo-grid {
    grid-template-columns: 1fr;
  }

  .shortcuts-grid {
    grid-template-columns: repeat(3, 1fr);
  }

  .bottom-row {
    grid-template-columns: 1fr;
  }
}
</style>
