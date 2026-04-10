<template>
  <div class="finance-home">
    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="凭证管理" name="voucher"><VoucherList /></el-tab-pane>
      <el-tab-pane label="科目管理" name="account"><AccountTree /></el-tab-pane>
      <el-tab-pane label="发票管理" name="invoice"><InvoiceList /></el-tab-pane>
      <el-tab-pane label="报销审批" name="expense"><ExpenseApproval /></el-tab-pane>
      <el-tab-pane label="账簿报表" name="book"><BookReport /></el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import VoucherList from './VoucherList.vue'
import AccountTree from './AccountTree.vue'
import InvoiceList from './InvoiceList.vue'
import ExpenseApproval from './ExpenseApproval.vue'
import BookReport from './BookReport.vue'

const activeTab = ref('voucher')
</script>

<style scoped lang="scss">
.finance-home {
  padding: 8px;
  padding-bottom: 70px;
}
</style>
