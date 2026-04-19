<template>
  <div class="page-view">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">人事工具</h1>
        <p class="page-subtitle">一站式薪资、社保、个税管理</p>
      </div>
    </header>

    <!-- 快捷入口 -->
    <div class="quick-actions">
      <div
        v-for="action in quickActions"
        :key="action.path"
        class="action-card glass-card"
        @click="navigateTo(action.path)"
      >
        <div class="action-icon" :style="{ background: action.bg }">
          <el-icon :style="{ color: action.color }">
            <component :is="action.icon" />
          </el-icon>
        </div>
        <div class="action-info">
          <span class="action-title">{{ action.title }}</span>
          <span class="action-desc">{{ action.desc }}</span>
        </div>
        <div class="action-arrow">
          <el-icon><ArrowRight /></el-icon>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--salary">
          <el-icon><Coin /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">--</span>
          <span class="stat-label">本月薪资总额</span>
        </div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--social">
          <el-icon><Umbrella /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">--</span>
          <span class="stat-label">本月社保总额</span>
        </div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--tax">
          <el-icon><Document /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">--</span>
          <span class="stat-label">待申报个税</span>
        </div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--employee">
          <el-icon><User /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">--</span>
          <span class="stat-label">在职员工</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { Coin, Umbrella, Document, User, ArrowRight } from '@element-plus/icons-vue'

const router = useRouter()

const quickActions = [
  {
    title: '薪资管理',
    desc: '工资核算、工资条发送',
    path: '/tool/salary',
    icon: Coin,
    bg: 'linear-gradient(135deg, #EDE9FE 0%, #DDD6FE 100%)',
    color: '#7C3AED',
  },
  {
    title: '社保管理',
    desc: '社保缴纳、查询与记录',
    path: '/tool/socialinsurance',
    icon: Umbrella,
    bg: 'linear-gradient(135deg, #DBEAFE 0%, #BFDBFE 100%)',
    color: '#3B82F6',
  },
  {
    title: '个税申报',
    desc: '专项附加扣除、申报记录',
    path: '/tool/tax',
    icon: Document,
    bg: 'linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%)',
    color: '#D97706',
  },
]

function navigateTo(path: string) {
  router.push(path)
}
</script>

<style scoped lang="scss">
.quick-actions {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.action-card {
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    transform: translateY(-4px);
    box-shadow: var(--shadow-lg);
  }
}

.action-icon {
  width: 52px;
  height: 52px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;

  .el-icon { font-size: 24px; }
}

.action-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.action-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.action-desc {
  font-size: 12px;
  color: var(--text-tertiary);
}

.action-arrow {
  color: var(--text-tertiary);
  font-size: 16px;
  transition: color 0.2s ease;

  .action-card:hover & { color: var(--primary); }
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-card {
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.2s ease;

  &:hover {
    transform: translateY(-2px);
    box-shadow: var(--shadow-md);
  }
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  flex-shrink: 0;

  &--salary { background: linear-gradient(135deg, #EDE9FE 0%, #DDD6FE 100%); color: var(--primary); }
  &--social { background: linear-gradient(135deg, #DBEAFE 0%, #BFDBFE 100%); color: #3B82F6; }
  &--tax { background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%); color: #D97706; }
  &--employee { background: linear-gradient(135deg, #D1FAE5 0%, #A7F3D0 100%); color: #059669; }
}

.stat-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-value {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
}

.stat-label {
  font-size: 12px;
  color: var(--text-tertiary);
}

@media (max-width: 1200px) {
  .quick-actions { grid-template-columns: repeat(2, 1fr); }
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
}

@media (max-width: 768px) {
  .quick-actions { grid-template-columns: 1fr; }
  .stats-grid { grid-template-columns: 1fr; }
}
</style>
