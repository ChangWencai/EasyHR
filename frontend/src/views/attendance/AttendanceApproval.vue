<template>
  <div class="page-view">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">考勤审批</h1>
        <p class="page-subtitle">管理员工的请假、加班等审批流程</p>
      </div>
      <div class="header-actions">
        <el-badge :value="pendingCount" :max="99" :hidden="pendingCount === 0">
          <el-button type="primary" size="large" @click="applyDialogRef?.open()">
            <el-icon><Plus /></el-icon>
            新建申请
          </el-button>
        </el-badge>
      </div>
    </header>

    <!-- 标签页 -->
    <div class="filter-tabs glass-card">
      <div class="tab-group">
        <button
          v-for="tab in tabOptions"
          :key="tab.value"
          class="tab-btn"
          :class="{ active: activeTab === tab.value }"
          @click="activeTab = tab.value; handleTabChange()"
        >
          <el-icon><component :is="tab.icon" /></el-icon>
          {{ tab.label }}
          <span v-if="tab.value === 'pending' && pendingCount > 0" class="tab-badge">{{ pendingCount }}</span>
        </button>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="table-card glass-card" v-loading="loading">
      <el-table
        :data="list"
        stripe
        class="modern-table"
        :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
        v-if="list.length > 0"
      >
        <el-table-column prop="employee_name" label="申请人" min-width="110">
          <template #default="{ row }">
            <div class="applicant-cell">
              <el-avatar :size="32" class="applicant-avatar">{{ row.employee_name?.[0] || '?' }}</el-avatar>
              <span class="applicant-name">{{ row.employee_name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="类型" min-width="110">
          <template #default="{ row }">
            <ApprovalTypeTag :type="row.approval_type" />
          </template>
        </el-table-column>
        <el-table-column label="时段" min-width="180">
          <template #default="{ row }">
            <div class="time-range">
              <span class="time-text">{{ formatTime(row.start_time) }}</span>
              <span class="time-arrow">
                <el-icon><Right /></el-icon>
              </span>
              <span class="time-text">{{ formatTime(row.end_time) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="时长" width="100">
          <template #default="{ row }">
            <span class="duration-chip">{{ formatDuration(row.duration, row.approval_type) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="事由" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="reason-text">{{ row.reason || '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <span class="status-badge" :class="`status--${row.status}`">
              {{ statusMap[row.status]?.label ?? row.status }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <template v-if="row.status === 'pending' && activeTab === 'pending'">
                <el-popconfirm title="确认同意该申请？" @confirm="handleApprove(row.id)">
                  <template #reference>
                    <el-button type="success" size="small" class="action-btn action-btn--approve">
                      <el-icon><CircleCheck /></el-icon>
                      同意
                    </el-button>
                  </template>
                </el-popconfirm>
                <el-button size="small" class="action-btn action-btn--reject" @click="handleReject(row)">
                  <el-icon><CloseBold /></el-icon>
                  驳回
                </el-button>
              </template>
              <el-popconfirm
                v-if="row.status === 'pending' && activeTab === 'mine'"
                title="确认撤回该申请？"
                @confirm="handleCancel(row.id)"
              >
                <template #reference>
                  <el-button size="small" class="action-btn action-btn--cancel">
                    <el-icon><RefreshRight /></el-icon>
                    撤回
                  </el-button>
                </template>
              </el-popconfirm>
              <span v-if="!isActionable(row)" class="no-action">—</span>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 空状态 -->
      <div v-if="!loading && list.length === 0" class="empty-state">
        <div class="empty-icon">
          <el-icon><Tickets /></el-icon>
        </div>
        <h3>暂无审批记录</h3>
        <p>当前筛选条件下没有待处理的审批</p>
      </div>

      <div class="pagination-wrapper" v-if="list.length > 0">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="loadData"
        />
      </div>
    </div>

    <ApprovalApplyDialog ref="applyDialogRef" @submitted="onSubmitted" />

    <!-- 驳回弹窗 -->
    <el-dialog v-model="rejectVisible" title="驳回申请" width="460px" class="reject-dialog">
      <div class="reject-content">
        <div class="reject-icon">
          <el-icon><WarningFilled /></el-icon>
        </div>
        <div class="reject-info">
          <h4>请输入驳回原因</h4>
          <p>申请人会收到驳回通知，可修改后重新提交</p>
        </div>
        <el-input
          v-model="rejectNote"
          type="textarea"
          :rows="3"
          placeholder="请输入驳回原因（选填）"
          class="reject-textarea"
        />
      </div>
      <template #footer>
        <el-button @click="rejectVisible = false" size="large">取消</el-button>
        <el-button type="danger" :loading="rejectLoading" size="large" @click="confirmReject">
          确认驳回
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { approvalApi, type ApprovalRecord } from '@/api/attendance'
import ApprovalTypeTag from '@/components/attendance/ApprovalTypeTag.vue'
import ApprovalApplyDialog from '@/components/attendance/ApprovalApplyDialog.vue'
import {
  Plus, CircleCheck, CloseBold, RefreshRight, Right,
  Tickets, WarningFilled,
} from '@element-plus/icons-vue'
import dayjs from 'dayjs'

const statusMap: Record<string, { label: string; cls: string }> = {
  draft:    { label: '草稿',    cls: 'draft'    },
  pending:  { label: '待审批',  cls: 'pending'  },
  approved: { label: '已通过',  cls: 'approved' },
  rejected: { label: '已驳回',  cls: 'rejected' },
  cancelled:{ label: '已撤回',  cls: 'cancelled'},
  timeout:  { label: '已过期',  cls: 'timeout'  },
}

const tabOptions = [
  { label: '全部',      value: 'all',     icon: 'Document' },
  { label: '待我审批',  value: 'pending', icon: 'Clock'    },
  { label: '我发起的',  value: 'mine',    icon: 'Check'    },
]

const loading = ref(false)
const list = ref<ApprovalRecord[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const activeTab = ref('all')
const pendingCount = ref(0)
const applyDialogRef = ref<InstanceType<typeof ApprovalApplyDialog>>()

const rejectVisible = ref(false)
const rejectLoading = ref(false)
const rejectNote = ref('')
const rejectTargetId = ref<number>(0)

function formatTime(t: string) { return dayjs(t).format('MM-DD HH:mm') }

function formatDuration(hours: number, type: string) {
  if (['makeup'].includes(type)) return '1次'
  const days = Math.round((hours / 8) * 10) / 10
  return hours >= 8 ? `${days}天` : `${hours}h`
}

function isActionable(row: ApprovalRecord) {
  if (row.status === 'pending' && activeTab.value === 'pending') return true
  if (row.status === 'pending' && activeTab.value === 'mine') return true
  return false
}

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = { page: page.value, page_size: pageSize }
    if (activeTab.value === 'pending') params.status = 'pending'
    const { data } = await approvalApi.list(params as Parameters<typeof approvalApi.list>[0])
    if (data) {
      list.value = data.list ?? []
      total.value = data.total
    }
  } catch { ElMessage.error('加载审批列表失败') }
  finally { loading.value = false }
}

async function loadPendingCount() {
  try {
    const { data } = await approvalApi.pendingCount()
    if (data) pendingCount.value = data.pending_count
  } catch { /* ignore */ }
}

function handleTabChange() {
  page.value = 1
  loadData()
}

async function handleApprove(id: number) {
  try {
    await approvalApi.approve(id)
    ElMessage.success('审批通过')
    loadData()
    loadPendingCount()
  } catch { ElMessage.error('操作失败') }
}

function handleReject(row: ApprovalRecord) {
  rejectTargetId.value = row.id
  rejectNote.value = ''
  rejectVisible.value = true
}

async function confirmReject() {
  rejectLoading.value = true
  try {
    await approvalApi.reject(rejectTargetId.value, rejectNote.value)
    ElMessage.success('已驳回')
    rejectVisible.value = false
    loadData()
    loadPendingCount()
  } catch { ElMessage.error('操作失败') }
  finally { rejectLoading.value = false }
}

async function handleCancel(id: number) {
  try {
    await approvalApi.cancel(id)
    ElMessage.success('已撤回')
    loadData()
  } catch { ElMessage.error('操作失败') }
}

function onSubmitted() { loadData(); loadPendingCount() }

onMounted(() => { loadData(); loadPendingCount() })
</script>

<style scoped lang="scss">
.filter-tabs { padding: 14px 20px; margin-bottom: 20px; }

.tab-group { display: inline-flex; background: #F3F4F6; border-radius: var(--radius-md); padding: 4px; }

.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
  background: transparent;
  position: relative;

  &.active {
    background: #fff;
    color: var(--primary);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  &:hover:not(.active) { color: var(--text-primary); }
  .el-icon { font-size: 15px; }
}

.tab-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  background: var(--danger);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  border-radius: 9px;
}

.table-card { padding: 0; overflow: hidden; }

:deep(.modern-table) {
  .el-table__header th { padding: 14px 12px; font-size: 13px; }
  .el-table__row { transition: background 0.2s ease; &:hover > td { background: rgba(var(--primary), 0.02) !important; } }
  .el-table__cell { padding: 14px 12px; border-bottom: 1px solid #F3F4F6; }
}

.applicant-cell { display: flex; align-items: center; gap: 10px; }
.applicant-avatar { background: linear-gradient(135deg, var(--primary-light), var(--primary)); color: #fff; font-size: 13px; font-weight: 600; }
.applicant-name { font-weight: 500; color: var(--text-primary); }

.time-range { display: flex; align-items: center; gap: 6px; }
.time-text { font-size: 13px; font-weight: 500; color: var(--text-primary); font-family: 'SF Mono', Monaco, monospace; }
.time-arrow { color: var(--text-tertiary); .el-icon { font-size: 12px; } }

.duration-chip {
  display: inline-flex;
  padding: 3px 10px;
  background: #EDE9FE;
  color: var(--primary);
  font-size: 12px;
  font-weight: 600;
  border-radius: 12px;
  font-family: 'SF Mono', Monaco, monospace;
}

.reason-text { font-size: 13px; color: var(--text-secondary); }

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 20px;

  &.status--draft    { background: #F3F4F6; color: #6B7280; }
  &.status--pending  { background: #FEF3C7; color: #D97706; }
  &.status--approved { background: #D1FAE5; color: #059669; }
  &.status--rejected { background: #FEE2E2; color: #DC2626; }
  &.status--cancelled{ background: #F3F4F6; color: #6B7280; }
  &.status--timeout { background: #F3F4F6; color: #6B7280; }
}

.action-cell { display: flex; align-items: center; gap: 4px; }

.action-btn {
  padding: 4px 10px !important;
  border-radius: var(--radius-sm) !important;
  font-size: 12px !important;
  display: inline-flex !important;
  align-items: center !important;
  gap: 4px !important;

  &--approve {
    color: var(--success) !important;
    background: #D1FAE5 !important;
    border: none !important;
    &:hover { background: #A7F3D0 !important; }
  }

  &--reject {
    color: var(--danger) !important;
    background: #FEE2E2 !important;
    border: none !important;
    &:hover { background: #FECACA !important; }
  }

  &--cancel {
    color: var(--warning) !important;
    background: #FEF3C7 !important;
    border: none !important;
    &:hover { background: #FDE68A !important; }
  }
}

.no-action { color: var(--text-tertiary); font-size: 14px; }

.pagination-wrapper { display: flex; justify-content: flex-end; padding: 16px 20px; border-top: 1px solid var(--border); }

.empty-state {
  text-align: center;
  padding: 80px 32px;

  .empty-icon {
    width: 72px;
    height: 72px;
    margin: 0 auto 16px;
    background: linear-gradient(135deg, #EDE9FE, #DDD6FE);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 32px;
    color: var(--primary);
  }

  h3 { font-size: 18px; font-weight: 600; color: var(--text-primary); margin: 0 0 8px; }
  p { font-size: 14px; color: var(--text-tertiary); margin: 0; }
}

.reject-content { text-align: center; }

.reject-icon {
  width: 56px;
  height: 56px;
  margin: 0 auto 16px;
  background: #FEE2E2;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: var(--danger);
}

.reject-info {
  margin-bottom: 20px;
  h4 { font-size: 16px; font-weight: 600; color: var(--text-primary); margin: 0 0 4px; }
  p { font-size: 13px; color: var(--text-tertiary); margin: 0; }
}

.reject-textarea :deep(.el-textarea__inner) {
  border-radius: var(--radius-md);
  &:focus { border-color: var(--primary); box-shadow: 0 0 0 3px rgba(var(--primary), 0.1); }
}

@media (max-width: 768px) {
  .tab-group { flex-wrap: wrap; }
  .time-range { flex-wrap: wrap; }
}
</style>
