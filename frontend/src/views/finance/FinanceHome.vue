<template>
  <div class="finance-page">
    <!-- 左侧子菜单 -->
    <aside class="finance-submenu glass-card">
      <div class="submenu-header">
        <div class="submenu-icon submenu-icon--finance">
          <el-icon><Wallet /></el-icon>
        </div>
        <div class="submenu-title-group">
          <span class="submenu-title">财务记账</span>
          <span class="submenu-desc">智能财税</span>
        </div>
      </div>
      <el-menu
        :default-active="activeSubMenu"
        router
        class="finance-menu"
      >
        <el-menu-item index="/finance/vouchers">
          <template #prefix>
            <el-icon><Document /></el-icon>
          </template>
          <span>凭证管理</span>
        </el-menu-item>
        <el-menu-item index="/finance/accounts">
          <template #prefix>
            <el-icon><Grid /></el-icon>
          </template>
          <span>科目管理</span>
        </el-menu-item>
        <el-menu-item index="/finance/invoices">
          <template #prefix>
            <el-icon><Tickets /></el-icon>
          </template>
          <span>发票管理</span>
        </el-menu-item>
        <el-menu-item index="/finance/expenses">
          <template #prefix>
            <el-icon><Coin /></el-icon>
          </template>
          <span>报销审批</span>
        </el-menu-item>
        <el-menu-item index="/finance/reports">
          <template #prefix>
            <el-icon><DataAnalysis /></el-icon>
          </template>
          <span>账簿报表</span>
        </el-menu-item>
      </el-menu>
    </aside>

    <!-- 右侧内容 -->
    <main class="finance-content">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Document, Grid, Tickets, Coin, DataAnalysis, Wallet } from '@element-plus/icons-vue'

const route = useRoute()

const activeSubMenu = computed(() => {
  if (route.path.startsWith('/finance/vouchers/create')) {
    return '/finance/vouchers/create'
  }
  return route.path
})
</script>

<style scoped lang="scss">
// ============================================================
// 变量定义
// ============================================================
$success: #10B981;
$success: #10B981;
$warning: #F59E0B;
$bg-page: #FAFBFC;
$bg-surface: #FFFFFF;
$text-primary: #1F2937;
$text-secondary: #6B7280;
$text-muted: #9CA3AF;
$border-color: #E5E7EB;
$shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
$shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
$radius-sm: 8px;
$radius-md: 12px;
$radius-lg: 16px;
$radius-xl: 24px;

// ============================================================
// 布局
// ============================================================
.finance-page {
  display: flex;
  min-height: calc(100vh - 56px);
  background: $bg-page;
  gap: 24px;
  padding: 24px 32px;
}

// ============================================================
// 左侧子菜单
// ============================================================
.finance-submenu {
  width: 260px;
  min-width: 260px;
  padding: 20px 16px;
  border-radius: $radius-xl;
  display: flex;
  flex-direction: column;
  gap: 16px;
  position: sticky;
  top: 24px;
  height: fit-content;
}

.submenu-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
}

.submenu-icon {
  width: 44px;
  height: 44px;
  background: linear-gradient(135deg, #06B6D4 0%, #0891B2 100%);
  border-radius: $radius-md;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 20px;
  box-shadow: 0 4px 12px rgba(#06B6D4, 0.3);

  &--finance {
    background: linear-gradient(135deg, #06B6D4 0%, #0891B2 100%);
    box-shadow: 0 4px 12px rgba(#06B6D4, 0.3);
  }
}

.submenu-title-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.submenu-title {
  font-size: 16px;
  font-weight: 600;
  color: $text-primary;
}

.submenu-desc {
  font-size: 12px;
  color: $text-muted;
}

.finance-menu {
  border-right: none !important;
  background: transparent;

  :deep(.el-menu-item) {
    height: 48px;
    line-height: 48px;
    font-size: 14px;
    font-weight: 500;
    color: $text-secondary;
    border-radius: $radius-md;
    margin-bottom: 4px;
    padding-left: 12px !important;
    transition: all 0.2s ease;

    .el-icon {
      font-size: 18px;
      margin-right: 12px;
      color: $text-muted;
      transition: color 0.2s ease;
    }

    &:hover {
      background: rgba(#06B6D4, 0.06);
      color: #0891B2;

      .el-icon {
        color: #0891B2;
      }
    }

    &.is-active {
      background: linear-gradient(135deg, rgba(#06B6D4, 0.1) 0%, rgba(#06B6D4, 0.05) 100%);
      color: #0891B2;
      font-weight: 600;

      .el-icon {
        color: #0891B2;
      }

      &::before {
        content: '';
        position: absolute;
        left: 0;
        top: 50%;
        transform: translateY(-50%);
        width: 4px;
        height: 24px;
        background: #06B6D4;
        border-radius: 0 4px 4px 0;
      }
    }
  }
}

// ============================================================
// 右侧内容
// ============================================================
.finance-content {
  flex: 1;
  min-width: 0;
}

// ============================================================
// 响应式
// ============================================================
@media (max-width: 1024px) {
  .finance-page {
    padding: 16px;
  }

  .finance-submenu {
    width: 220px;
    min-width: 220px;
  }
}

@media (max-width: 768px) {
  .finance-page {
    flex-direction: column;
    padding: 16px;
  }

  .finance-submenu {
    width: 100%;
    position: static;
  }
}
</style>
