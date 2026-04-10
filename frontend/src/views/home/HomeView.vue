<template>
  <div class="home-view">
    <!-- 顶部栏: 企业名称 + 我的入口 -->
    <div class="header">
      <span class="company-name">{{ companyName }}</span>
      <router-link to="/mine" class="mine-link">
        <el-icon><Avatar /></el-icon>
      </router-link>
    </div>

    <!-- 待办卡片区 -->
    <div class="section">
      <div v-if="store.loading" class="loading">
        <el-icon class="is-loading"><Loading /></el-icon>
      </div>
      <div v-else-if="store.todos.length === 0" class="empty-state">
        <el-empty description="暂无待办事项，轻松搞定人事~" :image-size="80" />
      </div>
      <div v-else class="todo-cards">
        <div
          v-for="todo in store.todos"
          :key="todo.type"
          class="todo-card"
          @click="handleTodoClick(todo)"
        >
          <div class="todo-info">
            <div class="todo-title">{{ todo.title }}</div>
            <div v-if="todo.deadline" class="todo-deadline">截止: {{ todo.deadline }}</div>
          </div>
          <div class="todo-count">
            <el-badge :value="todo.count" :max="99" />
          </div>
        </div>
      </div>
    </div>

    <!-- 5宫格入口 -->
    <div class="section">
      <div class="section-title">核心功能</div>
      <el-row :gutter="16" class="grid-5">
        <el-col v-for="item in gridItems" :key="item.path" :span="8">
          <router-link :to="item.path" class="grid-item">
            <el-icon :size="32" :color="item.color">
              <component :is="item.icon" />
            </el-icon>
            <span class="grid-label">{{ item.label }}</span>
          </router-link>
        </el-col>
      </el-row>
    </div>

    <!-- 数据概览 -->
    <div class="section overview-section">
      <div class="section-header" @click="store.toggleOverview">
        <span class="section-title">数据概览</span>
        <el-icon>
          <ArrowDown v-if="store.overviewExpanded" />
          <ArrowRight v-else />
        </el-icon>
      </div>
      <div v-if="store.overviewExpanded && store.overview" class="overview-grid">
        <div class="overview-item">
          <div class="overview-value">{{ store.overview.employee_count }}</div>
          <div class="overview-label">在职员工</div>
        </div>
        <div class="overview-item">
          <div class="overview-value">
            +{{ store.overview.joined_this_month }}/-{{ store.overview.left_this_month }}
          </div>
          <div class="overview-label">本月入/离职</div>
        </div>
        <div class="overview-item">
          <div class="overview-value">¥{{ store.overview.social_insurance_total }}</div>
          <div class="overview-label">本月社保总额</div>
        </div>
        <div class="overview-item">
          <div class="overview-value">¥{{ store.overview.payroll_total }}</div>
          <div class="overview-label">本月工资总额</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  User,
  Umbrella,
  Money,
  Document,
  Wallet,
  Avatar,
  ArrowDown,
  ArrowRight,
  Loading,
} from '@element-plus/icons-vue'
import { useDashboardStore } from '@/stores/dashboard'
import type { TodoItem } from '@/api/dashboard'

const store = useDashboardStore()
const router = useRouter()
const companyName = ref('我的企业')

const gridItems = [
  { path: '/employee', label: '员工管理', icon: User, color: '#1677ff' },
  { path: '/tool', label: '社保管理', icon: Umbrella, color: '#1677ff' },
  { path: '/tool', label: '工资管理', icon: Money, color: '#1677ff' },
  { path: '/tool', label: '个税申报', icon: Document, color: '#1677ff' },
  { path: '/finance', label: '财务管理', icon: Wallet, color: '#1677ff' },
]

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
  const path = routeMap[todo.type] || '/home'
  router.push(path)
}

onMounted(() => {
  store.load()
})
</script>

<style scoped lang="scss">
.home-view {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 70px;
}

.header {
  background: #1677ff;
  color: #fff;
  padding: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 18px;
}

.mine-link {
  color: #fff;
  text-decoration: none;
  display: flex;
  align-items: center;
}

.section {
  background: #fff;
  margin: 8px;
  border-radius: 8px;
  padding: 16px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin-bottom: 12px;
}

.loading {
  display: flex;
  justify-content: center;
  padding: 32px;
}

.empty-state {
  padding: 16px;
}

.todo-cards {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.todo-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  background: #f5f7ff;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;

  &:active {
    background: #e6efff;
  }
}

.todo-title {
  font-size: 14px;
  color: #333;
}

.todo-deadline {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}

.grid-5 {
  .grid-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    padding: 16px 8px;
    text-decoration: none;
    cursor: pointer;

    &:active {
      opacity: 0.7;
    }
  }

  .grid-label {
    font-size: 12px;
    color: #333;
  }
}

.overview-section {
  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    cursor: pointer;
  }

  .overview-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
    margin-top: 8px;
  }

  .overview-item {
    text-align: center;
    padding: 12px;
    background: #fafafa;
    border-radius: 6px;
  }

  .overview-value {
    font-size: 20px;
    font-weight: 700;
    color: #1677ff;
  }

  .overview-label {
    font-size: 12px;
    color: #999;
    margin-top: 4px;
  }
}
</style>
