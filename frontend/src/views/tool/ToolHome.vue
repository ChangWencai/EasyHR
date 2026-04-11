<template>
  <div class="tool-page">
    <!-- 左侧子菜单 -->
    <aside class="tool-submenu">
      <div class="submenu-title">人事工具</div>
      <el-menu
        :default-active="activeSubMenu"
        router
        class="tool-menu"
      >
        <el-menu-item index="/tool">
          <el-icon><Odometer /></el-icon>
          <span>概览</span>
        </el-menu-item>
        <el-menu-item index="/tool/salary">
          <el-icon><Coin /></el-icon>
          <span>薪资管理</span>
        </el-menu-item>
        <el-menu-item index="/tool/socialinsurance">
          <el-icon><Umbrella /></el-icon>
          <span>社保管理</span>
        </el-menu-item>
        <el-menu-item index="/tool/tax">
          <el-icon><Document /></el-icon>
          <span>个税申报</span>
        </el-menu-item>
      </el-menu>
    </aside>

    <!-- 右侧内容 -->
    <div class="tool-content">
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import {
  Odometer,
  Coin,
  Umbrella,
  Document,
} from '@element-plus/icons-vue'

const route = useRoute()
const activeSubMenu = computed(() => route.path)
</script>

<style scoped lang="scss">
.tool-page {
  display: flex;
  min-height: calc(100vh - 56px);
  background: #f0f2f5;
}

.tool-submenu {
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

.tool-menu {
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

.tool-content {
  flex: 1;
  padding: 16px;
  min-width: 0;
}

@media (max-width: 768px) {
  .tool-submenu {
    display: none;
  }

  .tool-content {
    padding: 8px;
  }
}
</style>
