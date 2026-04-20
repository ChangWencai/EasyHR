<template>
  <div class="app-layout">
    <!-- ============================================================
         移动端顶部栏（< 768px 显示）
         ============================================================ -->
    <header class="mobile-header">
      <el-button text @click="drawerVisible = true">
        <el-icon size="22"><Menu /></el-icon>
      </el-button>
      <span class="page-title">{{ currentPageTitle }}</span>
      <div class="mobile-actions">
        <el-button text @click="$router.push('/home')">
          <el-icon size="20"><HomeFilled /></el-icon>
        </el-button>
      </div>
    </header>

    <!-- ============================================================
         桌面端侧边栏（>= 768px 显示）
         ============================================================ -->
    <aside class="sidebar" :class="{ collapsed: isCollapsed }">
      <!-- Logo -->
      <div class="sidebar-logo">
        <div class="logo-icon">
          <el-icon size="22"><Management /></el-icon>
        </div>
        <transition name="fade">
          <span v-if="!isCollapsed" class="logo-text">易人事</span>
        </transition>
      </div>

      <!-- 菜单 -->
      <el-scrollbar wrap-class="sidebar-scroll" :height="'calc(100vh - 140px)'">
        <el-menu
          :default-active="activeMenu"
          class="sidebar-el-menu"
          :collapse="isCollapsed"
          :collapse-transition="false"
          router
        >
          <el-menu-item index="/home">
            <el-icon><HomeFilled /></el-icon>
            <template #title>首页</template>
          </el-menu-item>

          <el-sub-menu index="/employee">
            <template #title>
              <el-icon><UserFilled /></el-icon>
              <span>员工管理</span>
            </template>
            <el-menu-item index="/employee">员工列表</el-menu-item>
            <el-menu-item index="/employee/create">新增员工</el-menu-item>
            <el-menu-item index="/employee/invitations">入职邀请</el-menu-item>
            <el-menu-item index="/employee/offboardings">离职管理</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="/attendance">
            <template #title>
              <el-icon><Clock /></el-icon>
              <span>考勤管理</span>
            </template>
            <el-menu-item index="/attendance/rule">打卡规则</el-menu-item>
            <el-menu-item index="/attendance/clock-live">今日实况</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="/compliance">
            <template #title>
              <el-icon><Document /></el-icon>
              <span>合规报表</span>
            </template>
            <el-menu-item index="/compliance/overtime">加班统计</el-menu-item>
            <el-menu-item index="/compliance/leave">请假合规</el-menu-item>
            <el-menu-item index="/compliance/anomaly">出勤异常</el-menu-item>
            <el-menu-item index="/compliance/monthly">月度汇总</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="/tool">
            <template #title>
              <el-icon><Tools /></el-icon>
              <span>人事工具</span>
            </template>
            <el-menu-item index="/tool">工具概览</el-menu-item>
            <el-menu-item index="/tool/salary">薪资管理</el-menu-item>
            <el-menu-item index="/tool/socialinsurance">社保管理</el-menu-item>
            <el-menu-item index="/tool/tax">个税申报</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="/finance">
            <template #title>
              <el-icon><Money /></el-icon>
              <span>财务记账</span>
            </template>
            <el-menu-item index="/finance">财务概览</el-menu-item>
            <el-menu-item index="/finance/vouchers">凭证管理</el-menu-item>
            <el-menu-item index="/finance/accounts">科目管理</el-menu-item>
            <el-menu-item index="/finance/invoices">发票管理</el-menu-item>
            <el-menu-item index="/finance/expenses">报销审批</el-menu-item>
            <el-menu-item index="/finance/reports">账簿报表</el-menu-item>
          </el-sub-menu>

          <el-menu-item index="/mine">
            <el-icon><Avatar /></el-icon>
            <template #title>我的</template>
          </el-menu-item>
        </el-menu>
      </el-scrollbar>

      <!-- 折叠按钮 -->
      <div class="sidebar-footer">
        <div class="collapse-btn-wrapper">
          <el-button
            text
            class="collapse-btn"
            @click="isCollapsed = !isCollapsed"
          >
            <el-icon size="18">
              <DArrowLeft v-if="!isCollapsed" />
              <DArrowRight v-else />
            </el-icon>
            <span v-if="!isCollapsed" class="collapse-text">收起</span>
          </el-button>
        </div>
      </div>
    </aside>

    <!-- ============================================================
         主内容区
         ============================================================ -->
    <div class="main-wrapper" :class="{ 'sidebar-collapsed': isCollapsed }">
      <router-view v-slot="{ Component }">
        <transition name="page-fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </div>

    <!-- ============================================================
         移动端抽屉导航（< 768px）
         ============================================================ -->
    <el-drawer
      v-model="drawerVisible"
      direction="ltr"
      :size="280"
      :show-close="false"
      class="mobile-drawer"
    >
      <template #header>
        <div class="drawer-header">
          <div class="drawer-logo">
            <el-icon size="24"><Management /></el-icon>
          </div>
          <div class="drawer-title-group">
            <span class="drawer-title">易人事</span>
            <span class="drawer-subtitle">人事管理平台</span>
          </div>
        </div>
      </template>
      <el-menu
        :default-active="activeMenu"
        router
        @select="drawerVisible = false"
        class="drawer-menu"
      >
        <el-menu-item index="/home">
          <el-icon><HomeFilled /></el-icon>
          <template #title>首页</template>
        </el-menu-item>

        <el-sub-menu index="/employee">
          <template #title><el-icon><UserFilled /></el-icon><span>员工管理</span></template>
          <el-menu-item index="/employee">员工列表</el-menu-item>
          <el-menu-item index="/employee/create">新增员工</el-menu-item>
          <el-menu-item index="/employee/invitations">入职邀请</el-menu-item>
          <el-menu-item index="/employee/offboardings">离职管理</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="/attendance">
          <template #title><el-icon><Clock /></el-icon><span>考勤管理</span></template>
          <el-menu-item index="/attendance/rule">打卡规则</el-menu-item>
          <el-menu-item index="/attendance/clock-live">今日实况</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="/compliance">
          <template #title><el-icon><Document /></el-icon><span>合规报表</span></template>
          <el-menu-item index="/compliance/overtime">加班统计</el-menu-item>
          <el-menu-item index="/compliance/leave">请假合规</el-menu-item>
          <el-menu-item index="/compliance/anomaly">出勤异常</el-menu-item>
          <el-menu-item index="/compliance/monthly">月度汇总</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="/tool">
          <template #title><el-icon><Tools /></el-icon><span>人事工具</span></template>
          <el-menu-item index="/tool">工具概览</el-menu-item>
          <el-menu-item index="/tool/salary">薪资管理</el-menu-item>
          <el-menu-item index="/tool/socialinsurance">社保管理</el-menu-item>
          <el-menu-item index="/tool/tax">个税申报</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="/finance">
          <template #title><el-icon><Money /></el-icon><span>财务记账</span></template>
          <el-menu-item index="/finance">财务概览</el-menu-item>
          <el-menu-item index="/finance/vouchers">凭证管理</el-menu-item>
          <el-menu-item index="/finance/accounts">科目管理</el-menu-item>
          <el-menu-item index="/finance/invoices">发票管理</el-menu-item>
          <el-menu-item index="/finance/expenses">报销审批</el-menu-item>
        </el-sub-menu>

        <el-menu-item index="/mine">
          <el-icon><Avatar /></el-icon>
          <template #title>我的</template>
        </el-menu-item>
      </el-menu>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import {
  HomeFilled,
  UserFilled,
  Tools,
  Money,
  Avatar,
  Menu,
  DArrowLeft,
  DArrowRight,
  Management,
  Clock,
  Document,
} from '@element-plus/icons-vue'

const route = useRoute()
const isCollapsed = ref(false)
const drawerVisible = ref(false)

const activeMenu = computed(() => {
  const path = route.path
  if (route.matched.length > 1) {
    return route.matched[route.matched.length - 1].path
  }
  return path
})

const pageTitleMap: Record<string, string> = {
  '/home': '首页',
  '/attendance/rule': '打卡规则',
  '/attendance/clock-live': '今日实况',
  '/compliance/overtime': '加班统计',
  '/compliance/leave': '请假合规',
  '/compliance/anomaly': '出勤异常',
  '/compliance/monthly': '月度汇总',
  '/employee': '员工列表',
  '/employee/create': '新增员工',
  '/employee/invitations': '入职邀请',
  '/employee/offboardings': '离职管理',
  '/tool': '人事工具',
  '/tool/salary': '薪资管理',
  '/tool/socialinsurance': '社保管理',
  '/tool/tax': '个税申报',
  '/finance': '财务记账',
  '/finance/accounts': '科目管理',
  '/finance/vouchers': '凭证管理',
  '/finance/vouchers/create': '填制凭证',
  '/finance/invoices': '发票管理',
  '/finance/expenses': '报销审批',
  '/finance/reports': '账簿报表',
  '/mine': '我的',
}

const currentPageTitle = computed(() => pageTitleMap[route.path] || '易人事')
</script>

<style scoped lang="scss">
// ============================================================
// 变量定义
// ============================================================
$success: #10B981;
$bg-page: #FAFBFC;
$bg-sidebar: linear-gradient(180deg, #1F2937 0%, #111827 100%);
$text-sidebar: rgba(255, 255, 255, 0.7);
$text-sidebar-active: #fff;
$border-sidebar: rgba(255, 255, 255, 0.1);
$bg-sidebar-hover: rgba(255, 255, 255, 0.08);
$shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1);

// ============================================================
// 布局容器
// ============================================================
.app-layout {
  display: flex;
  min-height: 100vh;
  background: $bg-page;
}

// ============================================================
// 移动端顶部栏
// ============================================================
.mobile-header {
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  height: 56px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  padding: 0 16px;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.page-title {
  font-size: 16px;
  font-weight: 600;
  color: #1F2937;
}

.mobile-actions {
  display: flex;
  gap: 4px;
}

// ============================================================
// 侧边栏
// ============================================================
.sidebar {
  width: 260px;
  min-width: 260px;
  height: 100vh;
  position: fixed;
  left: 0;
  top: 0;
  background: $bg-sidebar;
  display: flex;
  flex-direction: column;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1),
              min-width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 200;
  overflow: hidden;

  &.collapsed {
    width: 72px;
    min-width: 72px;

    .sidebar-logo {
      padding: 0 20px;
      justify-content: center;
    }

    .collapse-btn-wrapper {
      justify-content: center;
    }
  }
}

// ============================================================
// Logo 区域
// ============================================================
.sidebar-logo {
  height: 64px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 24px;
  border-bottom: 1px solid $border-sidebar;
  flex-shrink: 0;
}

.logo-icon {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, var(--primary-light) 0%, var(--primary) 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(var(--primary), 0.4);
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
  color: #fff;
  white-space: nowrap;
  letter-spacing: 2px;
}

// ============================================================
// 菜单
// ============================================================
.sidebar-el-menu {
  border-right: none !important;
  background: transparent !important;
  padding: 12px 8px;

  // 菜单项
  .el-menu-item,
  .el-sub-menu__title {
    height: 48px;
    line-height: 48px;
    font-size: 14px;
    font-weight: 500;
    color: $text-sidebar !important;
    background: transparent !important;
    border-radius: 10px;
    margin-bottom: 2px;
    padding-left: 12px !important;
    transition: all 0.2s ease;

    .el-icon {
      font-size: 20px;
      margin-right: 12px;
      color: $text-sidebar;
      transition: color 0.2s ease;
    }

    &:hover {
      background: $bg-sidebar-hover !important;
      color: $text-sidebar-active !important;

      .el-icon {
        color: $text-sidebar-active;
      }
    }
  }

  // 激活状态
  .el-menu-item.is-active {
    background: linear-gradient(135deg, rgba(var(--primary-light), 0.3) 0%, rgba(var(--primary), 0.2) 100%) !important;
    color: $text-sidebar-active !important;
    border-right: none;

    .el-icon {
      color: var(--primary-light);
    }
  }

  // 子菜单
  :deep(.el-sub-menu .el-menu--inline) {
    background: rgba(255, 255, 255, 0.05) !important;
    border-radius: 8px;
    margin: 4px 0;

    .el-menu-item {
      height: 42px;
      line-height: 42px;
      font-size: 13px;
      font-weight: 400;
      color: rgba(255, 255, 255, 0.8) !important;
      background: transparent !important;
      border-radius: 8px;
      margin-bottom: 2px;
      padding-left: 72px !important;

      &:hover {
        background: rgba(255, 255, 255, 0.12) !important;
        color: #fff !important;
      }

      &.is-active {
        background: rgba(var(--primary), 0.35) !important;
        color: #fff !important;
        font-weight: 600;
      }
    }
  }

  // 子菜单箭头
  .el-sub-menu__icon-arrow {
    color: $text-sidebar;
    transition: transform 0.3s ease;
  }
}

// ============================================================
// 滚动区域
// ============================================================
.sidebar-scroll {
  overflow-x: hidden !important;
  flex: 1;
}

// ============================================================
// 折叠按钮
// ============================================================
.sidebar-footer {
  flex-shrink: 0;
  border-top: 1px solid $border-sidebar;
  padding: 12px 8px;
}

.collapse-btn-wrapper {
  display: flex;
  justify-content: flex-end;
}

.collapse-btn {
  color: rgba(255, 255, 255, 0.5);
  padding: 8px 12px;
  border-radius: 8px;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 6px;

  &:hover {
    color: #fff;
    background: $bg-sidebar-hover;
  }
}

.collapse-text {
  font-size: 13px;
  white-space: nowrap;
}

// ============================================================
// 主内容区
// ============================================================
.main-wrapper {
  flex: 1;
  margin-left: 260px;
  min-height: 100vh;
  transition: margin-left 0.3s cubic-bezier(0.4, 0, 0.2, 1);

  &.sidebar-collapsed {
    margin-left: 72px;
  }
}

// ============================================================
// 移动端抽屉
// ============================================================
.mobile-drawer {
  :deep(.el-drawer) {
    background: #1F2937 !important;
  }

  :deep(.el-drawer__header) {
    border-bottom: 1px solid $border-sidebar;
    margin-bottom: 0;
    padding: 20px;
    color: #fff;
  }

  .drawer-header {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .drawer-logo {
    width: 40px;
    height: 40px;
    background: linear-gradient(135deg, var(--primary-light) 0%, var(--primary) 100%);
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
  }

  .drawer-title-group {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .drawer-title {
    font-size: 18px;
    font-weight: 700;
    color: #fff;
  }

  .drawer-subtitle {
    font-size: 12px;
    color: $text-sidebar;
  }

  .drawer-menu {
    background: transparent !important;
    border-right: none !important;
    padding: 8px;

    .el-menu-item,
    .el-sub-menu__title {
      color: $text-sidebar !important;
      border-radius: 8px;
      margin-bottom: 2px;

      .el-icon {
        color: $text-sidebar;
      }

      &:hover {
        background: $bg-sidebar-hover !important;
      }

      &.is-active {
        background: linear-gradient(135deg, rgba(var(--primary-light), 0.3) 0%, rgba(var(--primary), 0.2) 100%) !important;
        color: #fff !important;

        .el-icon {
          color: var(--primary-light);
        }
      }
    }

    .el-menu--inline {
      padding-left: 44px;
      background: transparent !important;
    }
  }
}

// ============================================================
// 动画
// ============================================================
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.page-fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

// ============================================================
// 响应式
// ============================================================
@media (max-width: 768px) {
  .mobile-header {
    display: flex;
  }

  .sidebar {
    display: none;
  }

  .main-wrapper {
    margin-left: 0;
    padding-top: 56px;

    &.sidebar-collapsed {
      margin-left: 0;
    }
  }
}
</style>
