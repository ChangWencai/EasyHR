<template>
  <div class="finance-page">
    <!-- 左侧子菜单 -->
    <aside class="finance-submenu">
      <div class="submenu-title">财务记账</div>
      <el-menu
        :default-active="activeSubMenu"
        router
        class="finance-menu"
      >
        <el-menu-item index="/finance/vouchers">
          <el-icon><Document /></el-icon>
          <span>凭证管理</span>
        </el-menu-item>
        <el-menu-item index="/finance/accounts">
          <el-icon><Grid /></el-icon>
          <span>科目管理</span>
        </el-menu-item>
        <el-menu-item index="/finance/invoices">
          <el-icon><Tickets /></el-icon>
          <span>发票管理</span>
        </el-menu-item>
        <el-menu-item index="/finance/expenses">
          <el-icon><Coin /></el-icon>
          <span>报销审批</span>
        </el-menu-item>
        <el-menu-item index="/finance/reports">
          <el-icon><DataAnalysis /></el-icon>
          <span>账簿报表</span>
        </el-menu-item>
      </el-menu>
    </aside>

    <!-- 右侧内容 -->
    <div class="finance-content">
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import {
  Document,
  Grid,
  Tickets,
  Coin,
  DataAnalysis,
} from '@element-plus/icons-vue'

const route = useRoute()

const activeSubMenu = computed(() => {
  if (route.path.startsWith('/finance/vouchers/create')) {
    return '/finance/vouchers/create'
  }
  return route.path
})
</script>

<style scoped lang="scss">
.finance-page {
  display: flex;
  min-height: calc(100vh - 56px);
  background: #f0f2f5;
}

// ============================================================
// 左侧子菜单
// ============================================================
.finance-submenu {
  width: 180px;
  min-width: 180px;
  background: #fff;
  border-right: 1px solid #e8e8e8;
  padding-top: 8px;
  flex-shrink: 0;
  position: sticky;
  top: 0;
  height: calc(100vh - 56px);
  overflow-y: auto;
}

.submenu-title {
  padding: 12px 20px 8px;
  font-size: 12px;
  font-weight: 600;
  color: #999;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.finance-menu {
  border-right: none;
  background: transparent;

  .el-menu-item {
    height: 42px;
    line-height: 42px;
    font-size: 14px;
    color: #595959;
    display: flex;
    align-items: center;
    gap: 8px;

    .el-icon {
      font-size: 16px;
    }

    &:hover {
      background: #f5f5f5;
      color: #1677ff;
    }

    &.is-active {
      background: #e6f4ff;
      color: #1677ff;
      border-right: 2px solid #1677ff;
    }
  }
}

// ============================================================
// 右侧内容
// ============================================================
.finance-content {
  flex: 1;
  padding: 16px;
  min-width: 0;
}

// ============================================================
// 响应式：< 768px 子菜单折叠为标签栏
// ============================================================
@media (max-width: 768px) {
  .finance-submenu {
    display: none;
  }

  .finance-content {
    padding: 8px;
  }
}
</style>
