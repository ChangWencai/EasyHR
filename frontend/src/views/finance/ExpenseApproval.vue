<template>
  <div class="expense-approval">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">报销审批</h1>
        <p class="page-subtitle">管理员工报销申请与审批流程</p>
      </div>
    </header>

    <!-- 标签页 -->
    <div class="filter-tabs glass-card">
      <div class="tab-group">
        <button
          v-for="tab in statusTabs"
          :key="tab.value"
          class="tab-btn"
          :class="{ active: activeStatus === tab.value }"
          @click="activeStatus = tab.value; loadExpenses()"
        >
          <el-icon><component :is="tab.icon" /></el-icon>
          {{ tab.label }}
        </button>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="table-card glass-card" v-loading="loading">
      <el-table
        :data="expenses"
        stripe
        class="modern-table"
        :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
        v-if="expenses.length > 0"
      >
        <el-table-column prop="employee_name" label="员工姓名" width="130">
          <template #default="{ row }">
            <div class="employee-cell">
              <el-avatar :size="32" class="employee-avatar">{{ row.employee_name?.[0] || '?' }}</el-avatar>
              <span class="employee-name">{{ row.employee_name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="报销类型" width="130">
          <template #default="{ row }">
            <span class="type-chip">{{ row.type || '一般报销' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="amount" label="金额" align="right" width="150">
          <template #default="{ row }">
            <span class="amount-value">¥{{ formatAmount(row.amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="说明" min-width="200">
          <template #default="{ row }">
            <span class="desc-text" :title="row.description">{{ row.description || '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <span class="status-badge" :class="`status--${row.status}`">
              <span class="status-dot"></span>
              {{ statusLabel(row.status) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="提交时间" width="170">
          <template #default="{ row }">
            <span class="time-text">{{ row.created_at }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <template v-if="row.status === 'pending'">
                <el-button
                  size="small"
                  class="action-btn action-btn--approve"
                  @click="handleApprove(row)"
                >
                  <el-icon><CircleCheck /></el-icon>
                  通过
                </el-button>
                <el-button
                  size="small"
                  class="action-btn action-btn--reject"
                  @click="showRejectDialog(row)"
                >
                  <el-icon><CloseBold /></el-icon>
                  驳回
                </el-button>
              </template>
              <span v-else class="no-action">—</span>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 空状态 -->
      <div v-if="!loading && expenses.length === 0" class="empty-state">
        <div class="empty-icon">
          <el-icon><Tickets /></el-icon>
        </div>
        <h3>暂无报销记录</h3>
        <p>当前筛选条件下没有待处理的报销申请</p>
      </div>

      <div class="pagination-wrapper" v-if="expenses.length > 0">
        <el-pagination
          v-model:current-page="page"
          :page-size="20"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="loadExpenses"
        />
      </div>
    </div>

    <!-- 驳回弹窗 -->
    <el-dialog v-model="rejectVisible" title="驳回报销" width="460px" class="reject-dialog">
      <div class="reject-content">
        <div class="reject-icon">
          <el-icon><WarningFilled /></el-icon>
        </div>
        <div class="reject-info">
          <h4>驳回报销申请</h4>
          <p>请输入驳回原因，申请人会收到通知</p>
        </div>
        <el-input
          v-model="rejectForm.reason"
          type="textarea"
          :rows="3"
          placeholder="请输入驳回原因"
          class="reject-textarea"
        />
      </div>
      <template #footer>
        <el-button @click="rejectVisible = false" size="large">取消</el-button>
        <el-button type="danger" :loading="rejecting" size="large" @click="confirmReject">
          确认驳回
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { financeApi } from '@/api/finance'
import {
  CircleCheck, CloseBold, Tickets, WarningFilled,
} from '@element-plus/icons-vue'

interface Expense {
  id: number; employee_name: string; type: string; amount: string
  description: string; status: string; created_at: string
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

const statusTabs = [
  { label: '全部',     value: '',        icon: 'List'   },
  { label: '待审批',   value: 'pending',  icon: 'Clock'  },
  { label: '已通过',   value: 'approved', icon: 'Check'  },
  { label: '已支付',   value: 'paid',     icon: 'Money'  },
  { label: '已驳回',   value: 'rejected', icon: 'Close'  },
]

function statusLabel(status: string) {
  const map: Record<string, string> = {
    pending: '待审批', approved: '已通过', paid: '已支付', rejected: '已驳回',
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
    const res = await financeApi.expenses({ page: page.value, status: activeStatus.value || undefined }) as any
    expenses.value = res.data?.data || res.data?.list || []
    total.value = res.data?.total || 0
  } catch { expenses.value = [] }
  finally { loading.value = false }
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
  } finally { actionLoading.value = false }
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
  } finally { rejecting.value = false }
}

onMounted(() => loadExpenses())
</script>

<style scoped lang="scss">
$success: #10B981;
$warning: #F59E0B;
$error: #EF4444;
$bg-page: #FAFBFC;
$text-primary: #1F2937;
$text-secondary: #6B7280;
$text-muted: #9CA3AF;
$border-color: #E5E7EB;
$radius-sm: 8px;
$radius-md: 12px;
$radius-lg: 16px;
$radius-xl: 24px;

.expense-approval { padding: 24px 32px; width: 100%; box-sizing: border-box; background: $bg-page; min-height: 100vh; }

.glass-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: $radius-xl;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 24px;
  .page-title { font-size: 24px; font-weight: 700; color: $text-primary; margin: 0 0 4px; }
  .page-subtitle { font-size: 14px; color: $text-secondary; margin: 0; }
}

.filter-tabs { padding: 14px 20px; margin-bottom: 20px; }

.tab-group { display: inline-flex; background: #F3F4F6; border-radius: $radius-md; padding: 4px; flex-wrap: wrap; gap: 2px; }

.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: $radius-sm;
  font-size: 14px;
  font-weight: 500;
  color: $text-secondary;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
  background: transparent;

  &.active {
    background: #fff;
    color: var(--primary);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  &:hover:not(.active) { color: $text-primary; }
  .el-icon { font-size: 15px; }
}

.table-card { padding: 0; overflow: hidden; }

:deep(.modern-table) {
  .el-table__header th { padding: 14px 12px; font-size: 13px; }
  .el-table__row { transition: background 0.2s ease; &:hover > td { background: rgba(var(--primary), 0.02) !important; } }
  .el-table__cell { padding: 14px 12px; border-bottom: 1px solid #F3F4F6; }
}

.employee-cell { display: flex; align-items: center; gap: 10px; }
.employee-avatar { background: linear-gradient(135deg, var(--primary-light), var(--primary)); color: #fff; font-size: 13px; font-weight: 600; }
.employee-name { font-weight: 500; color: $text-primary; }

.type-chip {
  display: inline-flex;
  padding: 3px 10px;
  background: #EDE9FE;
  color: var(--primary);
  font-size: 12px;
  font-weight: 600;
  border-radius: 12px;
}

.amount-value { font-family: 'SF Mono', Monaco, monospace; font-weight: 700; font-size: 14px; color: $text-primary; }

.desc-text { font-size: 13px; color: $text-secondary; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: block; max-width: 200px; }

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 20px;

  .status-dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; }

  &.status--pending  { background: #FEF3C7; color: #D97706; }
  &.status--approved { background: #D1FAE5; color: #059669; }
  &.status--paid     { background: #DBEAFE; color: #3B82F6; }
  &.status--rejected { background: #FEE2E2; color: #DC2626; }
}

.time-text { font-size: 13px; color: $text-muted; font-family: 'SF Mono', Monaco, monospace; }

.action-cell { display: flex; align-items: center; gap: 4px; }

.action-btn {
  padding: 4px 10px !important;
  border-radius: $radius-sm !important;
  font-size: 12px !important;
  display: inline-flex !important;
  align-items: center !important;
  gap: 4px !important;

  &--approve {
    color: $success !important;
    background: #D1FAE5 !important;
    border: none !important;
    &:hover { background: #A7F3D0 !important; }
  }

  &--reject {
    color: $error !important;
    background: #FEE2E2 !important;
    border: none !important;
    &:hover { background: #FECACA !important; }
  }
}

.no-action { color: $text-muted; font-size: 14px; }

.pagination-wrapper { display: flex; justify-content: flex-end; padding: 16px 20px; border-top: 1px solid $border-color; }

.empty-state {
  text-align: center;
  padding: 80px 32px;

  .empty-icon {
    width: 72px; height: 72px;
    margin: 0 auto 16px;
    background: linear-gradient(135deg, #EDE9FE, #DDD6FE);
    border-radius: 50%;
    display: flex; align-items: center; justify-content: center;
    font-size: 32px; color: var(--primary);
  }

  h3 { font-size: 18px; font-weight: 600; color: $text-primary; margin: 0 0 8px; }
  p { font-size: 14px; color: $text-muted; margin: 0; }
}

.reject-content { text-align: center; }

.reject-icon {
  width: 56px; height: 56px;
  margin: 0 auto 16px;
  background: #FEE2E2;
  border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 28px; color: $error;
}

.reject-info {
  margin-bottom: 20px;
  h4 { font-size: 16px; font-weight: 600; color: $text-primary; margin: 0 0 4px; }
  p { font-size: 13px; color: $text-muted; margin: 0; }
}

.reject-textarea :deep(.el-textarea__inner) {
  border-radius: $radius-md;
  &:focus { border-color: var(--primary); box-shadow: 0 0 0 3px rgba(var(--primary), 0.1); }
}

@media (max-width: 768px) {
  .expense-approval { padding: 16px; }
  .tab-group { flex-wrap: wrap; }
}
</style>
