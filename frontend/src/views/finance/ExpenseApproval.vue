<template>
  <div class="expense-approval">
    <el-tabs v-model="activeStatus" @tab-change="loadExpenses">
      <el-tab-pane label="全部" name="" />
      <el-tab-pane label="待审批" name="pending" />
      <el-tab-pane label="已通过" name="approved" />
      <el-tab-pane label="已支付" name="paid" />
      <el-tab-pane label="已驳回" name="rejected" />
    </el-tabs>

    <el-table :data="expenses" stripe v-loading="loading" class="mt-2">
      <el-table-column prop="employee_name" label="员工姓名" width="100" />
      <el-table-column prop="type" label="报销类型" width="100" />
      <el-table-column prop="amount" label="金额" align="right" width="120">
        <template #default="{ row }">{{ formatAmount(row.amount) }}</template>
      </el-table-column>
      <el-table-column prop="description" label="说明" min-width="160" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="提交时间" width="160" />
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <template v-if="row.status === 'pending'">
            <el-button link type="success" size="small" @click="handleApprove(row)">通过</el-button>
            <el-button link type="danger" size="small" @click="showRejectDialog(row)">驳回</el-button>
          </template>
          <span v-else class="no-action">-</span>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="page"
      :page-size="20"
      :total="total"
      layout="prev, pager, next"
      class="mt-2"
      @current-change="loadExpenses"
    />

    <!-- Reject Dialog -->
    <el-dialog v-model="rejectVisible" title="驳回报销" width="420px">
      <el-form :model="rejectForm" label-width="80px">
        <el-form-item label="驳回原因" required>
          <el-input
            v-model="rejectForm.reason"
            type="textarea"
            :rows="3"
            placeholder="请输入驳回原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" :loading="actionLoading" @click="confirmReject">确认驳回</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { financeApi } from '@/api/finance'

interface Expense {
  id: number
  employee_name: string
  type: string
  amount: string
  description: string
  status: string
  created_at: string
}

const expenses = ref<Expense[]>([])
const loading = ref(false)
const activeStatus = ref('')
const page = ref(1)
const total = ref(0)
const rejectVisible = ref(false)
const actionLoading = ref(false)
const rejecting = ref(false)
const currentExpense = ref<Expense | null>(null)
const rejectForm = ref({ reason: '' })

function statusType(status: string) {
  const map: Record<string, string> = {
    pending: 'warning',
    approved: 'success',
    paid: 'primary',
    rejected: 'danger',
  }
  return map[status] || 'info'
}

function statusLabel(status: string) {
  const map: Record<string, string> = {
    pending: '待审批',
    approved: '已通过',
    paid: '已支付',
    rejected: '已驳回',
  }
  return map[status] || status
}

function formatAmount(val: string) {
  if (!val) return '0.00'
  return parseFloat(val).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

async function loadExpenses() {
  loading.value = true
  try {
    const res = await financeApi.expenses({
      page: page.value,
      status: activeStatus.value || undefined,
    }) as any
    expenses.value = res.data?.data || res.data?.list || []
    total.value = res.data?.total || 0
  } catch {
    expenses.value = []
  } finally {
    loading.value = false
  }
}

async function handleApprove(row: Expense) {
  if (actionLoading.value) return
  actionLoading.value = true
  try {
    await financeApi.approveExpense(row.id)
    ElMessage.success('审批通过')
    loadExpenses()
  } catch (e: unknown) {
    const msg = (e as any)?.response?.data?.error || '操作失败'
    ElMessage.error(msg)
  } finally {
    actionLoading.value = false
  }
}

function showRejectDialog(row: Expense) {
  currentExpense.value = row
  rejectForm.value.reason = ''
  rejectVisible.value = true
}

async function confirmReject() {
  if (rejecting.value) return
  if (!rejectForm.value.reason.trim()) {
    ElMessage.warning('请输入驳回原因')
    return
  }
  rejecting.value = true
  try {
    await financeApi.rejectExpense(currentExpense.value!.id, rejectForm.value.reason)
    ElMessage.success('已驳回')
    rejectVisible.value = false
    loadExpenses()
  } catch (e: unknown) {
    const msg = (e as any)?.response?.data?.error || '操作失败'
    ElMessage.error(msg)
  } finally {
    rejecting.value = false
  }
}

onMounted(() => {
  loadExpenses()
})
</script>

<style scoped lang="scss">
.expense-approval {
  padding: 8px;
  .toolbar {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }
  .mt-2 {
    margin-top: 12px;
  }
  .no-action {
    color: #999;
  }
}
</style>
