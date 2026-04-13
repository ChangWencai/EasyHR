<template>
  <div class="app-layout" :class="{ 'is-desktop': isDesktop }">
    <!-- 顶部导航栏（移动端） -->
    <header class="top-nav mobile-only">
      <div class="nav-brand">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
          <rect width="24" height="24" rx="6" fill="#0F766E"/>
          <path d="M6 8h12M6 12h8M6 16h10" stroke="#fff" stroke-width="2" stroke-linecap="round"/>
        </svg>
        <span class="brand-name">易人事</span>
      </div>
      <router-link to="/mine" class="nav-avatar">
        <el-avatar :size="32" style="background: #0F766E; font-size: 14px">
          {{ avatarText }}
        </el-avatar>
      </router-link>
    </header>

    <!-- 左侧导航栏（桌面端） -->
    <aside class="sidebar desktop-only">
      <div class="sidebar-brand">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none">
          <rect width="24" height="24" rx="6" fill="#0F766E"/>
          <path d="M6 8h12M6 12h8M6 16h10" stroke="#fff" stroke-width="2" stroke-linecap="round"/>
        </svg>
        <span class="sidebar-brand-name">易人事</span>
      </div>

      <nav class="sidebar-nav">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="sidebar-item"
          :class="{ active: isActive(item.path) }"
        >
          <el-icon :size="20">
            <component :is="item.icon" />
          </el-icon>
          <span>{{ item.label }}</span>
          <span v-if="item.badge" class="sidebar-badge">{{ item.badge }}</span>
        </router-link>
      </nav>

      <div class="sidebar-user">
        <router-link to="/mine" class="sidebar-user-link">
          <el-avatar :size="36" style="background: #0F766E; font-size: 14px; flex-shrink: 0">
            {{ avatarText }}
          </el-avatar>
          <div class="sidebar-user-info">
            <div class="sidebar-user-name">{{ userName }}</div>
            <div class="sidebar-user-role">管理员</div>
          </div>
        </router-link>
      </div>
    </aside>

    <!-- 主内容区 -->
    <main class="main-content">
      <router-view />
    </main>

    <!-- 底部 Tab 栏（仅移动端） -->
    <BottomTabBar class="mobile-only" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import BottomTabBar from './BottomTabBar.vue'
import { useUserStore } from '@/stores/user'
import { useDashboardStore } from '@/stores/dashboard'
import { HomeFilled, UserFilled, Tools, Money, Avatar } from '@element-plus/icons-vue'

const route = useRoute()
const userStore = useUserStore()
const dashboard = useDashboardStore()

const isDesktop = ref(window.innerWidth >= 768)
const userName = computed(() => userStore.user?.name || '管理员')
const avatarText = computed(() => userStore.user?.name?.[0] || '管')

const totalBadge = computed(() => {
  const total = dashboard.todos.reduce((sum, t) => sum + t.count, 0)
  return total > 0 ? String(total) : null
})

const navItems = computed(() => [
  { path: '/home', label: '工作台', icon: HomeFilled, badge: totalBadge.value },
  { path: '/employee', label: '员工管理', icon: UserFilled, badge: null },
  { path: '/tool', label: '工具中心', icon: Tools, badge: null },
  { path: '/finance', label: '财务管理', icon: Money, badge: null },
  { path: '/mine', label: '我的', icon: Avatar, badge: null },
])

function isActive(path: string) {
  return route.path.startsWith(path)
}

function handleResize() {
  isDesktop.value = window.innerWidth >= 768
}

onMounted(() => window.addEventListener('resize', handleResize))
onUnmounted(() => window.removeEventListener('resize', handleResize))
</script>

<style scoped lang="scss">
// ===== 响应式基础布局 =====
.app-layout {
  min-height: 100vh;
  background: #F8FAFC;
  display: flex;
  flex-direction: column;
}

// ===== 移动端顶部导航 =====
.top-nav {
  position: sticky;
  top: 0;
  z-index: 100;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  height: 52px;
  box-shadow: 0 1px 0 0 #E2E8F0;
  padding-top: env(safe-area-inset-top, 0);
  padding-top: constant(safe-area-inset-top, 0);
}

.nav-brand {
  display: flex;
  align-items: center;
  gap: 8px;
}

.brand-name {
  font-size: 17px;
  font-weight: 700;
  color: #0F172A;
  letter-spacing: -0.3px;
}

.nav-avatar {
  display: flex;
  align-items: center;
  text-decoration: none;
  cursor: pointer;
  transition: opacity 0.15s ease-out;

  &:active { opacity: 0.7; }
}

// ===== 桌面端侧边栏 =====
.sidebar {
  width: 220px;
  min-width: 220px;
  background: #fff;
  border-right: 1px solid #E2E8F0;
  display: flex;
  flex-direction: column;
  height: 100vh;
  position: sticky;
  top: 0;
  overflow-y: auto;
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px 20px 16px;
  border-bottom: 1px solid #F1F5F9;
}

.sidebar-brand-name {
  font-size: 18px;
  font-weight: 700;
  color: #0F172A;
  letter-spacing: -0.3px;
}

.sidebar-nav {
  flex: 1;
  padding: 12px 12px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 14px;
  border-radius: 10px;
  text-decoration: none;
  color: #64748B;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.15s ease-out;
  cursor: pointer;
  position: relative;
  min-height: 44px;

  &:hover {
    background: #F8FAFC;
    color: #0F172A;
  }

  &.active {
    background: #F0FDFA;
    color: #0F766E;

    .sidebar-badge {
      background: #0F766E;
      color: #fff;
    }
  }
}

.sidebar-badge {
  margin-left: auto;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  background: #E2E8F0;
  color: #64748B;
  font-size: 11px;
  font-weight: 600;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.sidebar-user {
  padding: 16px;
  border-top: 1px solid #F1F5F9;
}

.sidebar-user-link {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  padding: 8px;
  border-radius: 10px;
  transition: background 0.15s;
  cursor: pointer;

  &:hover { background: #F8FAFC; }
}

.sidebar-user-info {
  min-width: 0;
}

.sidebar-user-name {
  font-size: 14px;
  font-weight: 600;
  color: #0F172A;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sidebar-user-role {
  font-size: 12px;
  color: #94A3B8;
  margin-top: 1px;
}

// ===== 主内容区 =====
.main-content {
  flex: 1;
  padding-bottom: calc(60px + env(safe-area-inset-bottom, 0));
  padding-bottom: calc(60px + constant(safe-area-inset-bottom, 0));
  max-width: 100%;
}

// ===== 桌面端布局 =====
.is-desktop {
  flex-direction: row;

  .main-content {
    padding-bottom: 0;
    max-width: calc(100vw - 220px);
    overflow-x: hidden;
  }
}

// ===== 响应式显示/隐藏 =====
.mobile-only {
  display: flex;

  .desktop-only & {
    display: none;
  }
}

.desktop-only {
  display: none;

  .desktop-only & {
    display: flex;
  }

  .is-desktop & {
    display: flex;
  }
}

// 在桌面端隐藏底部 Tab 栏
:deep(.mobile-only) {
  display: flex;

  .is-desktop & {
    display: none !important;
  }
}
</style>
