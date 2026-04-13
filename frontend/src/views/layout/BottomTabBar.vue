<template>
  <nav class="tab-bar-wrapper" role="tablist" aria-label="主导航">
    <router-link
      v-for="tab in tabs"
      :key="tab.path"
      :to="tab.path"
      class="tab-item"
      :class="{ active: isActive(tab.path) }"
      role="tab"
      :aria-selected="isActive(tab.path)"
    >
      <div class="tab-icon-wrap">
        <el-icon :size="22" :color="isActive(tab.path) ? '#0F766E' : '#94A3B8'">
          <component :is="tab.icon" />
        </el-icon>
        <!-- 徽章 -->
        <span v-if="tab.badge" class="tab-badge" aria-label="待处理">{{ tab.badge }}</span>
      </div>
      <span class="tab-label" :class="{ active: isActive(tab.path) }">{{ tab.label }}</span>
      <!-- 激活指示条 -->
      <div v-if="isActive(tab.path)" class="active-indicator" />
    </router-link>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { HomeFilled, UserFilled, Tools, Money, Avatar } from '@element-plus/icons-vue'
import { useDashboardStore } from '@/stores/dashboard'

const route = useRoute()
const dashboard = useDashboardStore()

// 合并所有待办数量
const totalBadge = computed(() => {
  const total = dashboard.todos.reduce((sum, t) => sum + t.count, 0)
  return total > 0 ? (total > 99 ? '99+' : String(total)) : null
})

const tabs = [
  { path: '/home', label: '首页', icon: HomeFilled, badge: null },
  { path: '/employee', label: '员工', icon: UserFilled, badge: null },
  { path: '/tool', label: '工具', icon: Tools, badge: null },
  { path: '/finance', label: '财务', icon: Money, badge: null },
  { path: '/mine', label: '我的', icon: Avatar, badge: null },
]

function isActive(path: string): boolean {
  return route.path.startsWith(path)
}
</script>

<style scoped lang="scss">
.tab-bar-wrapper {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  background: #ffffff;
  border-top: 1px solid #E2E8F0;
  z-index: 1000;
  // 适配底部安全区（iPhone 等）
  padding-bottom: env(safe-area-inset-bottom, 0px);
  padding-bottom: constant(safe-area-inset-bottom, 0px);
  height: calc(60px + env(safe-area-inset-bottom, 0px));
  height: calc(60px + constant(safe-area-inset-bottom, 0px));
}

.tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  text-decoration: none;
  position: relative;
  // 触控区域扩大
  padding: 8px 4px;
  min-height: 60px;
  transition: background 0.15s ease-out;

  &:active {
    background: #F1F5F9;
  }
}

.tab-icon-wrap {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 22px;
}

.tab-badge {
  position: absolute;
  top: -4px;
  right: -8px;
  min-width: 16px;
  height: 16px;
  padding: 0 3px;
  background: #EF4444;
  color: #fff;
  font-size: 10px;
  font-weight: 600;
  line-height: 16px;
  text-align: center;
  border-radius: 8px;
  box-sizing: border-box;
}

.tab-label {
  font-size: 10px;
  font-weight: 500;
  color: #94A3B8;
  line-height: 1;
  transition: color 0.15s ease-out;

  &.active {
    color: #0F766E;
  }
}

.active-indicator {
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 20px;
  height: 2px;
  background: #0F766E;
  border-radius: 1px 1px 0 0;
}
</style>
