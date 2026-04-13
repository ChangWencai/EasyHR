<template>
  <div class="home-view">
    <!-- 欢迎区域 -->
    <div class="hero-section">
      <div class="hero-greeting">
        <p class="greeting-time">{{ greeting }}</p>
        <h1 class="greeting-name">{{ userName }}，{{ subGreeting }}</h1>
      </div>
      <div class="hero-stats">
        <div class="stat-card">
          <div class="stat-value primary">{{ overview?.employee_count ?? '--' }}</div>
          <div class="stat-label">在职员工</div>
        </div>
        <div class="stat-divider" />
        <div class="stat-card">
          <div class="stat-value success">+{{ overview?.joined_this_month ?? 0 }}</div>
          <div class="stat-label">本月入职</div>
        </div>
        <div class="stat-divider" />
        <div class="stat-card">
          <div class="stat-value danger">-{{ overview?.left_this_month ?? 0 }}</div>
          <div class="stat-label">本月离职</div>
        </div>
      </div>
    </div>

    <!-- 待办区域 -->
    <div class="section">
      <div class="section-header">
        <span class="section-title">
          <span class="title-dot" />
          待办事项
        </span>
        <span v-if="store.todos.length > 0" class="todo-total">{{ store.todos.reduce((s, t) => s + t.count, 0) }} 项</span>
      </div>

      <!-- 骨架屏加载态 -->
      <div v-if="store.loading" class="skeleton-todos">
        <div v-for="i in 3" :key="i" class="skeleton-card" />
      </div>

      <!-- 空状态 -->
      <div v-else-if="store.todos.length === 0" class="empty-todo">
        <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
          <circle cx="24" cy="24" r="24" fill="#F1F5F9"/>
          <path d="M16 24l6 6 10-12" stroke="#0F766E" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
        <p class="empty-title">暂无待办</p>
        <p class="empty-sub">人事事务处理得井井有条</p>
      </div>

      <!-- 待办卡片列表 -->
      <div v-else class="todo-list" role="list">
        <div
          v-for="todo in store.todos"
          :key="todo.type"
          class="todo-card"
          role="listitem"
          tabindex="0"
          @click="handleTodoClick(todo)"
          @keydown.enter="handleTodoClick(todo)"
        >
          <div class="todo-icon" :style="{ background: todoColor(todo.type).bg }">
            <el-icon :size="18" :color="todoColor(todo.type).fg">
              <component :is="todoIcon(todo.type)" />
            </el-icon>
          </div>
          <div class="todo-body">
            <div class="todo-title">{{ todo.title }}</div>
            <div v-if="todo.deadline" class="todo-deadline">
              <el-icon :size="12"><Clock /></el-icon>
              {{ todo.deadline }}
            </div>
          </div>
          <div class="todo-right">
            <span class="todo-count">{{ todo.count }}</span>
            <el-icon class="todo-arrow"><ArrowRight /></el-icon>
          </div>
        </div>
      </div>
    </div>

    <!-- 核心功能入口 -->
    <div class="section">
      <div class="section-header">
        <span class="section-title">
          <span class="title-dot" />
          核心功能
        </span>
      </div>
      <div class="grid-5">
        <router-link
          v-for="item in gridItems"
          :key="item.path"
          :to="item.path"
          class="grid-item"
          :style="{ '--item-color': item.color }"
        >
          <div class="grid-icon-wrap" :style="{ background: item.color + '18' }">
            <el-icon :size="24" :color="item.color">
              <component :is="item.icon" />
            </el-icon>
          </div>
          <span class="grid-label">{{ item.label }}</span>
          <span class="grid-sub">{{ item.sub }}</span>
        </router-link>
      </div>
    </div>

    <!-- 财务概览 -->
    <div class="section">
      <div class="section-header">
        <span class="section-title">
          <span class="title-dot" />
          本月财务
        </span>
        <router-link to="/finance" class="section-more">
          详情 <el-icon :size="12"><ArrowRight /></el-icon>
        </router-link>
      </div>
      <div class="finance-grid">
        <div class="finance-item">
          <div class="finance-value">¥{{ formatNum(overview?.payroll_total) }}</div>
          <div class="finance-label">工资总额</div>
        </div>
        <div class="finance-divider" />
        <div class="finance-item">
          <div class="finance-value">¥{{ formatNum(overview?.social_insurance_total) }}</div>
          <div class="finance-label">社保总额</div>
        </div>
        <div class="finance-divider" />
        <div class="finance-item">
          <div class="finance-value">¥{{ formatNum(overview?.tax_total) }}</div>
          <div class="finance-label">个税总额</div>
        </div>
      </div>
    </div>

    <!-- 快捷工具 -->
    <div class="section">
      <div class="section-header">
        <span class="section-title">
          <span class="title-dot" />
          快捷工具
        </span>
      </div>
      <div class="quick-tools">
        <router-link
          v-for="tool in quickTools"
          :key="tool.path"
          :to="tool.path"
          class="quick-tool"
        >
          <el-icon :size="20" :color="tool.color">
            <component :is="tool.icon" />
          </el-icon>
          <span>{{ tool.label }}</span>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  User, Umbrella, Money, Document, Wallet,
  Clock, ArrowRight, Plus, Ticket,
  Coin, Calendar, Download,
} from '@element-plus/icons-vue'
import { useDashboardStore } from '@/stores/dashboard'
import { useUserStore } from '@/stores/user'
import type { TodoItem } from '@/api/dashboard'

const store = useDashboardStore()
const userStore = useUserStore()
const router = useRouter()

const userName = computed(() => userStore.user?.name || '老板')
const overview = computed(() => store.overview)

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 12) return '早上好'
  if (h < 18) return '下午好'
  return '晚上好'
})
const subGreeting = computed(() => {
  const h = new Date().getHours()
  if (h < 12) return '新的一天，加油！'
  if (h < 18) return '继续努力，有条不紊'
  return '注意休息，早点休息'
})

const gridItems = [
  { path: '/employee', label: '员工管理', sub: '入职·在职·离职', icon: User, color: '#0F766E' },
  { path: '/employee/create', label: '新增员工', sub: '3步完成入职', icon: Plus, color: '#10B981' },
  { path: '/tool', label: '社保管理', sub: '公积金·社保', icon: Umbrella, color: '#0EA5E9' },
  { path: '/tool', label: '工资核算', sub: '一键工资条', icon: Coin, color: '#F59E0B' },
  { path: '/finance', label: '财务管理', sub: '凭证·发票·报销', icon: Wallet, color: '#8B5CF6' },
]

const quickTools = [
  { path: '/tool', label: '个税申报', icon: Document, color: '#EF4444' },
  { path: '/tool', label: '工资条导出', icon: Download, color: '#10B981' },
  { path: '/finance', label: '查看凭证', icon: Ticket, color: '#F59E0B' },
  { path: '/employee', label: '考勤记录', icon: Calendar, color: '#0EA5E9' },
]

// 待办类型颜色映射
const todoColorMap: Record<string, { bg: string; fg: string }> = {
  social_insurance: { bg: '#FEF3C7', fg: '#D97706' },
  tax: { bg: '#FEE2E2', fg: '#DC2626' },
  employee: { bg: '#DCFCE7', fg: '#16A34A' },
  contract: { bg: '#E0E7FF', fg: '#4F46E5' },
  expense: { bg: '#FEF3C7', fg: '#D97706' },
  voucher: { bg: '#DBEAFE', fg: '#2563EB' },
}

const todoIconMap: Record<string, any> = {
  social_insurance: Umbrella,
  tax: Document,
  employee: User,
  contract: Document,
  expense: Money,
  voucher: Ticket,
}

function todoColor(type: string) {
  return todoColorMap[type] || { bg: '#F1F5F9', fg: '#64748B' }
}

function todoIcon(type: string) {
  return todoIconMap[type] || Clock
}

function formatNum(val: number | undefined): string {
  if (val === undefined || val === null) return '--'
  if (val >= 10000) return (val / 10000).toFixed(1) + '万'
  return val.toLocaleString()
}

function handleTodoClick(todo: TodoItem) {
  store.removeTodo(todo.type)
  const routeMap: Record<string, string> = {
    social_insurance: '/tool',
    tax: '/tool',
    employee: '/employee',
    contract: '/employee',
    expense: '/finance',
    voucher: '/finance',
  }
  router.push(routeMap[todo.type] || '/home')
}

onMounted(() => {
  store.load()
})
</script>

<style scoped lang="scss">
.home-view {
  background: #F8FAFC;
  min-height: 100%;
  padding-bottom: 8px;
}

// ===== Hero 区域 =====
.hero-section {
  background: linear-gradient(135deg, #0F766E 0%, #0D9488 60%, #14B8A6 100%);
  color: #fff;
  padding: 20px 16px 24px;
  border-radius: 0 0 20px 20px;
  // 抵消顶部导航
  margin-top: -1px;
}

.hero-greeting {
  margin-bottom: 16px;
}

.greeting-time {
  font-size: 13px;
  opacity: 0.8;
  margin: 0 0 2px;
}

.greeting-name {
  font-size: 22px;
  font-weight: 700;
  margin: 0;
  letter-spacing: -0.3px;
}

.hero-stats {
  display: flex;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  padding: 12px 8px;
  gap: 0;
}

.stat-card {
  flex: 1;
  text-align: center;
}

.stat-divider {
  width: 1px;
  background: rgba(255, 255, 255, 0.3);
  margin: 0 4px;
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  line-height: 1;

  &.primary { color: #fff; }
  &.success { color: #A7F3D0; }
  &.danger { color: #FCA5A5; }
}

.stat-label {
  font-size: 11px;
  opacity: 0.8;
  margin-top: 4px;
}

// ===== 通用 Section =====
.section {
  background: #fff;
  margin: 12px 12px 0;
  border-radius: 16px;
  padding: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #0F172A;
  display: flex;
  align-items: center;
  gap: 6px;
}

.title-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #0F766E;
  flex-shrink: 0;
}

.todo-total {
  font-size: 12px;
  color: #94A3B8;
}

// ===== 空状态 =====
.empty-todo {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 0 16px;
  gap: 6px;
}

.empty-title {
  font-size: 15px;
  font-weight: 600;
  color: #64748B;
  margin: 0;
}

.empty-sub {
  font-size: 12px;
  color: #94A3B8;
  margin: 0;
}

// ===== 骨架屏 =====
.skeleton-todos {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.skeleton-card {
  height: 60px;
  background: linear-gradient(90deg, #F1F5F9 25%, #E2E8F0 50%, #F1F5F9 75%);
  background-size: 200% 100%;
  border-radius: 10px;
  animation: shimmer 1.4s infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

// ===== 待办列表 =====
.todo-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.todo-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #F8FAFC;
  border-radius: 12px;
  cursor: pointer;
  transition: transform 0.15s ease-out, box-shadow 0.15s ease-out;
  border: 1px solid transparent;
  min-height: 44px; // 触控目标

  &:hover {
    border-color: #E2E8F0;
  }

  &:active {
    transform: scale(0.98);
    box-shadow: none;
  }
}

.todo-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.todo-body {
  flex: 1;
  min-width: 0;
}

.todo-title {
  font-size: 14px;
  font-weight: 500;
  color: #0F172A;
  line-height: 1.4;
}

.todo-deadline {
  font-size: 11px;
  color: #94A3B8;
  margin-top: 2px;
  display: flex;
  align-items: center;
  gap: 3px;
}

.todo-right {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.todo-count {
  min-width: 22px;
  height: 22px;
  background: #0F766E;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  border-radius: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 6px;
}

.todo-arrow {
  color: #CBD5E1;
  font-size: 14px;
}

// ===== 核心功能 5 宫格 =====
.grid-5 {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 4px;
}

.grid-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 12px 4px 10px;
  text-decoration: none;
  border-radius: 10px;
  cursor: pointer;
  transition: background 0.15s ease-out;
  min-height: 44px;

  &:active {
    background: #F1F5F9;
  }
}

.grid-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.grid-label {
  font-size: 12px;
  font-weight: 600;
  color: #0F172A;
  text-align: center;
  line-height: 1.2;
}

.grid-sub {
  font-size: 10px;
  color: #94A3B8;
  text-align: center;
  line-height: 1.2;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 64px;
}

// ===== 财务概览 =====
.section-more {
  font-size: 12px;
  color: #94A3B8;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 2px;
  cursor: pointer;
  transition: color 0.15s;

  &:hover {
    color: #0F766E;
  }
}

.finance-grid {
  display: flex;
  align-items: center;
  background: #F8FAFC;
  border-radius: 12px;
  padding: 14px 12px;
  gap: 0;
}

.finance-item {
  flex: 1;
  text-align: center;
}

.finance-divider {
  width: 1px;
  height: 32px;
  background: #E2E8F0;
}

.finance-value {
  font-size: 16px;
  font-weight: 700;
  color: #0F172A;
  font-feature-settings: "tnum";
}

.finance-label {
  font-size: 11px;
  color: #94A3B8;
  margin-top: 3px;
}

// ===== 快捷工具 =====
.quick-tools {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}

.quick-tool {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 12px 4px;
  text-decoration: none;
  background: #F8FAFC;
  border-radius: 10px;
  font-size: 11px;
  color: #64748B;
  cursor: pointer;
  transition: background 0.15s ease-out;
  min-height: 44px;

  &:active {
    background: #E2E8F0;
  }
}
</style>
