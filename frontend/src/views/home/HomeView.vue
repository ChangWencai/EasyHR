<template>
  <div class="home-view">
    <!-- 页面标题区 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">工作台</h1>
        <span class="page-subtitle">{{ greeting }}，欢迎回来</span>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="$router.push('/employee/create')">
          <el-icon><Plus /></el-icon>
          新增员工
        </el-button>
      </div>
    </div>

    <!-- 待办事项（全宽） -->
    <div class="section todo-section">
      <div class="section-header">
        <span class="section-title">待办事项</span>
        <el-badge :value="store.todos.length" :hidden="store.todos.length === 0" />
      </div>

      <div v-if="store.loading" class="loading">
        <el-icon class="is-loading" size="24"><Loading /></el-icon>
      </div>
      <div v-else-if="store.todos.length === 0" class="empty-state">
        <el-empty description="暂无待办事项，轻松搞定人事~" :image-size="80" />
      </div>
      <div v-else class="todo-grid">
        <div
          v-for="todo in store.todos"
          :key="todo.type"
          class="todo-card"
          @click="handleTodoClick(todo)"
        >
          <div class="todo-info">
            <div class="todo-title">{{ todo.title }}</div>
            <div v-if="todo.deadline" class="todo-deadline">
              <el-icon><Clock /></el-icon>
              截止 {{ todo.deadline }}
            </div>
          </div>
          <div class="todo-action">
            <el-badge :value="todo.count" :max="99" class="todo-badge" />
            <el-icon class="todo-arrow"><ArrowRight /></el-icon>
          </div>
        </div>
      </div>
    </div>

    <!-- 快捷入口 + 数据概览 同行 -->
    <div class="bottom-row">
      <!-- 快捷入口（占2/3） -->
      <div class="section shortcuts-section">
        <div class="section-title">快捷入口</div>
        <div class="shortcuts-grid">
          <router-link
            v-for="item in gridItems"
            :key="item.path + item.label"
            :to="item.path"
            class="shortcut-item"
          >
            <div class="shortcut-icon" :style="{ background: item.bg }">
              <el-icon size="22" :color="item.color">
                <component :is="item.icon" />
              </el-icon>
            </div>
            <span class="shortcut-label">{{ item.label }}</span>
          </router-link>
        </div>
      </div>

      <!-- 数据概览（占1/3） -->
      <div class="section overview-section">
        <div class="section-header" @click="store.toggleOverview">
          <span class="section-title">数据概览</span>
          <el-icon class="toggle-icon">
            <ArrowUp v-if="store.overviewExpanded" />
            <ArrowDown v-else />
          </el-icon>
        </div>
        <div v-if="store.overviewExpanded && store.overview" class="overview-grid">
          <div class="overview-item">
            <div class="overview-value">{{ store.overview.employee_count }}</div>
            <div class="overview-label">在职员工</div>
          </div>
          <div class="overview-item">
            <div class="overview-value small">
              <span class="green">+{{ store.overview.joined_this_month }}</span>
              <span class="sep">/</span>
              <span class="red">-{{ store.overview.left_this_month }}</span>
            </div>
            <div class="overview-label">本月入/离职</div>
          </div>
          <div class="overview-item">
            <div class="overview-value">¥{{ store.overview.social_insurance_total }}</div>
            <div class="overview-label">本月社保</div>
          </div>
          <div class="overview-item">
            <div class="overview-value">¥{{ store.overview.payroll_total }}</div>
            <div class="overview-label">本月工资</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
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
  ArrowUp,
  Clock,
  Loading,
} from '@element-plus/icons-vue'
import { useDashboardStore } from '@/stores/dashboard'
import type { TodoItem } from '@/api/dashboard'

const store = useDashboardStore()
const router = useRouter()

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 12) return '上午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const gridItems = [
  { path: '/employee', label: '员工管理', icon: User, color: '#1677ff', bg: '#e6f4ff' },
  { path: '/tool/salary', label: '薪资管理', icon: Money, color: '#52c41a', bg: '#f6ffed' },
  { path: '/tool/socialinsurance', label: '社保管理', icon: Umbrella, color: '#722ed1', bg: '#f9f0ff' },
  { path: '/tool/tax', label: '个税申报', icon: Document, color: '#fa8c16', bg: '#fff7e6' },
  { path: '/finance/vouchers', label: '凭证管理', icon: Wallet, color: '#13c2c2', bg: '#e6fffa' },
  { path: '/finance/invoices', label: '发票管理', icon: Document, color: '#eb2f96', bg: '#fff0f6' },
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
.home-view {
  padding: 20px 24px;
  width: 100%;
  box-sizing: border-box;
}

// ============================================================
// 页面标题区
// ============================================================
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;

  .header-left {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .page-title {
    font-size: 22px;
    font-weight: 700;
    color: #1a1a1a;
    margin: 0;
    line-height: 1.2;
  }

  .page-subtitle {
    font-size: 13px;
    color: #8c8c8c;
  }
}

// ============================================================
// 通用区块
// ============================================================
.section {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 16px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #1a1a1a;
}

.toggle-icon {
  color: #8c8c8c;
  cursor: pointer;
}

// ============================================================
// 待办
// ============================================================
.todo-section {
  // 全宽
}

.loading {
  display: flex;
  justify-content: center;
  padding: 32px;
}

.empty-state {
  padding: 8px 0;
}

.todo-grid {
  display: grid;
  // 响应式：4列(>1600px) → 3列(>1200px) → 2列(>768px) → 1列(<=768px)
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.todo-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  background: #fafafa;
  border-radius: 8px;
  border: 1px solid #f0f0f0;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    border-color: #1677ff;
    background: #f0f7ff;
    box-shadow: 0 2px 8px rgba(22, 119, 255, 0.1);
  }
}

.todo-info {
  flex: 1;
  min-width: 0;
}

.todo-title {
  font-size: 14px;
  font-weight: 500;
  color: #1a1a1a;
}

.todo-deadline {
  font-size: 12px;
  color: #8c8c8c;
  margin-top: 4px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.todo-action {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.todo-arrow {
  color: #bfbfbf;
  transition: color 0.2s;
}

.todo-card:hover .todo-arrow {
  color: #1677ff;
}

// ============================================================
// 快捷入口 + 数据概览同行
// ============================================================
.bottom-row {
  display: grid;
  // 默认: 2/3 + 1/3
  grid-template-columns: 2fr 1fr;
  gap: 16px;
  align-items: start;
}

.shortcuts-section {}

.shortcuts-grid {
  display: grid;
  // 6个入口: 6列(>1400px) → 3列(>900px) → 2列(<=768px)
  grid-template-columns: repeat(6, 1fr);
  gap: 8px;
  margin-top: 12px;
}

.shortcut-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 8px;
  text-decoration: none;
  border-radius: 8px;
  transition: background 0.2s;
  cursor: pointer;

  &:hover {
    background: #f5f5f5;
  }
}

.shortcut-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.shortcut-label {
  font-size: 13px;
  color: #595959;
  text-align: center;
  white-space: nowrap;
}

// ============================================================
// 数据概览
// ============================================================
.overview-section {}

.overview-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.overview-item {
  text-align: center;
  padding: 14px 12px;
  background: #fafafa;
  border-radius: 8px;
}

.overview-value {
  font-size: 24px;
  font-weight: 700;
  color: #1677ff;
  line-height: 1.2;

  &.small {
    font-size: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
  }

  .green { color: #52c41a; }
  .sep { color: #d9d9d9; font-weight: 400; }
  .red { color: #ff4d4f; }
}

.overview-label {
  font-size: 12px;
  color: #8c8c8c;
  margin-top: 4px;
}

// ============================================================
// 响应式断点
// ============================================================

// 大屏: 1440px - 1600px
@media (max-width: 1600px) {
  .todo-grid {
    grid-template-columns: repeat(3, 1fr);
  }
  .shortcuts-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

// 笔记本: 1200px - 1440px
@media (max-width: 1200px) {
  .bottom-row {
    grid-template-columns: 1fr;
  }
  .shortcuts-grid {
    grid-template-columns: repeat(6, 1fr);
  }
}

// iPad 横屏: 768px - 1200px
@media (max-width: 900px) {
  .todo-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .shortcuts-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

// 移动端
@media (max-width: 768px) {
  .home-view {
    padding: 12px;
  }

  .todo-grid {
    grid-template-columns: 1fr;
  }

  .shortcuts-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .bottom-row {
    grid-template-columns: 1fr;
  }
}
</style>
