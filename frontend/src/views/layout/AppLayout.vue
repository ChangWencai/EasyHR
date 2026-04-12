<template>
  <div class="app-layout">
    <!-- ============================================================
         移动端：顶部栏（< 768px 显示）
         ============================================================ -->
    <div class="mobile-header">
      <el-button text @click="drawerVisible = true">
        <el-icon size="22"><Menu /></el-icon>
      </el-button>
      <span class="page-title">{{ currentPageTitle }}</span>
      <div style="width: 40px" />
    </div>

    <!-- ============================================================
         桌面端侧边栏（>= 768px 显示）
         ============================================================ -->
    <aside class="sidebar" :class="{ collapsed: isCollapsed }">
      <!-- Logo -->
      <div class="sidebar-logo">
        <div class="logo-icon">
          <el-icon size="22"><Management /></el-icon>
        </div>
        <span class="logo-text">易人事</span>
      </div>

      <!-- 菜单 -->
      <el-scrollbar wrap-class="sidebar-scroll" :height="'calc(100vh - 120px)'">
        <el-menu
          :default-active="activeMenu"
          class="sidebar-el-menu"
          :collapse="isCollapsed"
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

          <el-sub-menu index="/tool">
            <template #title>
              <el-icon><Tools /></el-icon>
              <span>人事工具</span>
            </template>
            <el-menu-item index="/tool">概览</el-menu-item>
            <el-menu-item index="/tool/salary">薪资管理</el-menu-item>
            <el-menu-item index="/tool/socialinsurance">社保管理</el-menu-item>
            <el-menu-item index="/tool/tax">个税申报</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="/finance">
            <template #title>
              <el-icon><Money /></el-icon>
              <span>财务记账</span>
            </template>
            <el-menu-item index="/finance">概览</el-menu-item>
            <el-menu-item index="/finance/accounts">科目管理</el-menu-item>
            <el-menu-item index="/finance/vouchers">凭证管理</el-menu-item>
            <el-menu-item index="/finance/vouchers/create">填制凭证</el-menu-item>
            <el-menu-item index="/finance/invoices">发票管理</el-menu-item>
            <el-menu-item index="/finance/expenses">报销审批</el-menu-item>
            <el-menu-item index="/finance/reports">账簿报表</el-menu-item>
            <el-menu-item index="/finance/tax-declaration">纳税申报</el-menu-item>
          </el-sub-menu>

          <el-menu-item index="/mine">
            <el-icon><Avatar /></el-icon>
            <template #title>我的</template>
          </el-menu-item>
        </el-menu>
      </el-scrollbar>

      <!-- 折叠按钮 -->
      <div class="sidebar-footer">
        <el-button text class="collapse-btn" @click="isCollapsed = !isCollapsed">
          <el-icon size="18">
            <DArrowLeft v-if="!isCollapsed" />
            <DArrowRight v-else />
          </el-icon>
        </el-button>
      </div>
    </aside>

    <!-- ============================================================
         主内容区
         ============================================================ -->
    <div class="main-wrapper" :class="{ 'sidebar-collapsed': isCollapsed }">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
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
      :size="240"
      :show-close="false"
    >
      <template #header>
        <div class="drawer-header">
          <el-icon size="20"><Management /></el-icon>
          <span class="logo-text">易人事</span>
        </div>
      </template>
      <el-menu
        :default-active="activeMenu"
        router
        @select="drawerVisible = false"
      >
        <el-menu-item index="/home">
          <el-icon><HomeFilled /></el-icon><template #title>首页</template>
        </el-menu-item>

        <el-sub-menu index="/employee">
          <template #title><el-icon><UserFilled /></el-icon><span>员工管理</span></template>
          <el-menu-item index="/employee">员工列表</el-menu-item>
          <el-menu-item index="/employee/create">新增员工</el-menu-item>
          <el-menu-item index="/employee/invitations">入职邀请</el-menu-item>
          <el-menu-item index="/employee/offboardings">离职管理</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="/tool">
          <template #title><el-icon><Tools /></el-icon><span>人事工具</span></template>
          <el-menu-item index="/tool">概览</el-menu-item>
          <el-menu-item index="/tool/salary">薪资管理</el-menu-item>
          <el-menu-item index="/tool/socialinsurance">社保管理</el-menu-item>
          <el-menu-item index="/tool/tax">个税申报</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="/finance">
          <template #title><el-icon><Money /></el-icon><span>财务记账</span></template>
          <el-menu-item index="/finance">概览</el-menu-item>
          <el-menu-item index="/finance/accounts">科目管理</el-menu-item>
          <el-menu-item index="/finance/vouchers">凭证管理</el-menu-item>
          <el-menu-item index="/finance/vouchers/create">填制凭证</el-menu-item>
          <el-menu-item index="/finance/invoices">发票管理</el-menu-item>
          <el-menu-item index="/finance/expenses">报销审批</el-menu-item>
          <el-menu-item index="/finance/reports">账簿报表</el-menu-item>
          <el-menu-item index="/finance/tax-declaration">纳税申报</el-menu-item>
        </el-sub-menu>

        <el-menu-item index="/mine">
          <el-icon><Avatar /></el-icon><template #title>我的</template>
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
  '/finance/tax-declaration': '纳税申报',
  '/mine': '我的',
}

const currentPageTitle = computed(() => pageTitleMap[route.path] || '易人事')
</script>

<style scoped lang="scss">
.app-layout {
  display: flex;
  min-height: 100vh;
  background: #f0f2f5;
}

// ============================================================
// 移动端顶部栏（默认隐藏）
// ============================================================
.mobile-header {
  display: none;
  position: sticky;
  top: 0;
  z-index: 100;
  height: 56px;
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
  padding: 0 16px;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.page-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

// ============================================================
// 侧边栏（桌面端默认显示）
// ============================================================
.sidebar {
  width: 220px;
  min-width: 220px;
  height: 100vh;
  position: fixed;
  left: 0;
  top: 0;
  background: #fff;
  border-right: 1px solid #e8e8e8;
  display: flex;
  flex-direction: column;
  transition: width 0.2s ease, min-width 0.2s ease;
  z-index: 200;
  overflow: hidden;

  &.collapsed {
    width: 64px;
    min-width: 64px;
  }
}

// ============================================================
// Logo 区域
// ============================================================
.sidebar-logo {
  height: 56px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 20px;
  border-bottom: 1px solid #f0f0f0;
  flex-shrink: 0;
  overflow: hidden;

  .logo-icon {
    width: 32px;
    height: 32px;
    background: #1677ff;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    flex-shrink: 0;
  }

  .logo-text {
    font-size: 16px;
    font-weight: 700;
    color: #1677ff;
    white-space: nowrap;
    overflow: hidden;
  }
}

// ============================================================
// Element Plus menu
// ============================================================
.sidebar-el-menu {
  border-right: none;
  background: transparent;

  &:not(.el-menu--collapse) {
    width: 100%;
  }

  // 子菜单激活态
  .is-active > .el-sub-menu__title {
    color: #1677ff !important;
  }

  .el-menu-item,
  .el-sub-menu__title {
    height: 44px;
    line-height: 44px;
    font-size: 14px;
    color: #595959;

    &:hover {
      background: #f5f5f5;
      color: #1677ff;
    }
  }

  .el-menu-item.is-active {
    background: #e6f4ff;
    color: #1677ff;
    border-right: 2px solid #1677ff;
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
  border-top: 1px solid #f0f0f0;
  padding: 8px 8px 8px 4px;
  display: flex;
  justify-content: flex-end;

  .collapse-btn {
    color: #8c8c8c;
    padding: 8px;

    &:hover {
      color: #1677ff;
      background: #f5f5f5;
    }
  }
}

// ============================================================
// 主内容区
// ============================================================
.main-wrapper {
  flex: 1;
  margin-left: 220px;
  min-height: 100vh;
  transition: margin-left 0.2s ease;

  &.sidebar-collapsed {
    margin-left: 64px;
  }
}

// ============================================================
// 移动端抽屉
// ============================================================
.drawer-header {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #1677ff;
  font-size: 16px;
  font-weight: 700;
}

// ============================================================
// 页面切换动画
// ============================================================
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

// ============================================================
// 响应式：< 768px
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

    &.sidebar-collapsed {
      margin-left: 0;
    }
  }
}
</style>
