<template>
  <div class="tool-home">
    <el-tabs v-model="activeTab" class="tool-tabs">
      <el-tab-pane label="薪资管理" name="salary">
        <SalaryTool />
      </el-tab-pane>
      <el-tab-pane label="社保管理" name="social-insurance">
        <SITool />
      </el-tab-pane>
      <el-tab-pane label="个税申报" name="tax">
        <TaxTool />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SalaryTool from './SalaryTool.vue'
import SITool from './SITool.vue'
import TaxTool from './TaxTool.vue'

const activeTab = ref('salary')
</script>

<style scoped lang="scss">
.tool-home {
  padding-bottom: 70px;
}
.tool-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: 0;
  }
}
</style>
