<template>
  <div class="tool-page">
    <!-- 左侧子菜单 -->
    <aside class="tool-submenu glass-card">
      <div class="submenu-header">
        <div class="submenu-icon">
          <el-icon><Tools /></el-icon>
        </div>
        <div class="submenu-title-group">
          <span class="submenu-title">人事工具</span>
          <span class="submenu-desc">高效管理</span>
        </div>
      </div>
      <el-menu
        :default-active="activeSubMenu"
        router
        class="tool-menu"
      >
        <el-menu-item index="/tool">
          <template #prefix>
            <el-icon><Odometer /></el-icon>
          </template>
          <span>概览</span>
        </el-menu-item>
        <el-menu-item index="/tool/salary">
          <template #prefix>
            <el-icon><Coin /></el-icon>
          </template>
          <span>薪资管理</span>
        </el-menu-item>
        <el-menu-item index="/tool/socialinsurance">
          <template #prefix>
            <el-icon><Umbrella /></el-icon>
          </template>
          <span>社保管理</span>
        </el-menu-item>
        <el-menu-item index="/tool/tax">
          <template #prefix>
            <el-icon><Document /></el-icon>
          </template>
          <span>个税申报</span>
        </el-menu-item>
      </el-menu>
    </aside>

    <!-- 右侧内容 -->
    <main class="tool-content">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Odometer, Coin, Umbrella, Document, Tools } from '@element-plus/icons-vue'

const route = useRoute()
const activeSubMenu = computed(() => route.path)
</script>

<style scoped lang="scss">
// ============================================================
// 变量定义
// ============================================================
$success: #10B981;
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
.tool-page {
  display: flex;
  min-height: calc(100vh - 56px);
  background: $bg-page;
  gap: 24px;
  padding: 24px 32px;
}

// ============================================================
// 左侧子菜单
// ============================================================
.tool-submenu {
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
  background: linear-gradient(135deg, var(--primary-light) 0%, var(--primary) 100%);
  border-radius: $radius-md;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 20px;
  box-shadow: 0 4px 12px rgba(var(--primary), 0.3);
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

.tool-menu {
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
      background: rgba(var(--primary), 0.06);
      color: var(--primary);

      .el-icon {
        color: var(--primary);
      }
    }

    &.is-active {
      background: linear-gradient(135deg, rgba(var(--primary), 0.1) 0%, rgba(var(--primary), 0.05) 100%);
      color: var(--primary);
      font-weight: 600;

      .el-icon {
        color: var(--primary);
      }

      &::before {
        content: '';
        position: absolute;
        left: 0;
        top: 50%;
        transform: translateY(-50%);
        width: 4px;
        height: 24px;
        background: var(--primary);
        border-radius: 0 4px 4px 0;
      }
    }
  }
}

// ============================================================
// 右侧内容
// ============================================================
.tool-content {
  flex: 1;
  min-width: 0;
}

// ============================================================
// 响应式
// ============================================================
@media (max-width: 1024px) {
  .tool-page {
    padding: 16px;
  }

  .tool-submenu {
    width: 220px;
    min-width: 220px;
  }
}

@media (max-width: 768px) {
  .tool-page {
    flex-direction: column;
    padding: 16px;
  }

  .tool-submenu {
    width: 100%;
    position: static;
  }
}
</style>
