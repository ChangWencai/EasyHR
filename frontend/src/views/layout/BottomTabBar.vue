<template>
  <div class="tab-bar-wrapper">
    <router-link
      v-for="tab in tabs"
      :key="tab.path"
      :to="tab.path"
      class="tab-item"
      :class="{ active: route.path.startsWith(tab.path) }"
    >
      <el-icon :size="22">
        <component :is="tab.icon" />
      </el-icon>
      <span class="tab-label">{{ tab.label }}</span>
    </router-link>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { HomeFilled, UserFilled, Tools, Money, Avatar } from '@element-plus/icons-vue'

const route = useRoute()

const tabs = [
  { path: '/home', label: '首页', icon: HomeFilled },
  { path: '/employee', label: '员工', icon: UserFilled },
  { path: '/tool', label: '工具', icon: Tools },
  { path: '/finance', label: '财务', icon: Money },
  { path: '/mine', label: '我的', icon: Avatar },
]
</script>

<style scoped lang="scss">
.tab-bar-wrapper {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  background: #fff;
  border-top: 1px solid #eee;
  padding-bottom: var(--safe-area-bottom, 0px);
  z-index: 1000;
  height: calc(60px + var(--safe-area-bottom, 0px));
}

.tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  text-decoration: none;
  color: #999;
  transition: color 0.2s;

  &.active {
    color: #1677ff;
  }

  .tab-label {
    font-size: 10px;
    line-height: 1;
  }
}
</style>
