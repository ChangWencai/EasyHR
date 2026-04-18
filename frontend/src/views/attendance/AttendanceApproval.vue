<template>
  <div class="attendance-approval">
    <el-card>
      <template #header>
        <div class="header">
          <span>考勤审批</span>
          <div class="header-actions">
            <el-badge v-if="pendingCount > 0" :value="pendingCount" :max="99">
              <el-button type="primary" @click="applyDialogRef?.open()">+ 新建申请</el-button>
            </el-badge>
            <el-button v-else type="primary" @click="applyDialogRef?.open()">+ 新建申请</el-button>
          </div>
        </div>
      </template>

      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="全部" name="all" />
        <el-tab-pane name="pending">
          <template #label>
            待我审批
            <el-badge v-if="pendingCount > 0" :value="pendingCount" :max="99" class="tab-badge" />
          </template>
        </el-tab-pane>
        <el-tab-pane label="我发起的" name="mine" />
      </el-tabs>

      <el-table :data="list" stripe v-loading="loading" style="width: 100%">
        <el-table-column prop="employee_name" label="申请人" min-width="80" />
        <el-table-column label="类型" min-width="90">
          <template #default="{ row }">
            <ApprovalTypeTag :type="row.approval_type" />
          </template>
        </el-table-column>
        <el-table-column label="时段" min-width="160">
          <template #default="{ row }">
            {{ formatTime(row.start_time) }} ~ {{ formatTime(row.end_time) }}
          </template>
        </el-table-column>
        <el-table-column label="时长" width="100">
          <template #default="{ row }">
            {{ formatDuration(row.duration, row.approval_type) }}
          </template>
        </el-table-column>
        <el-table-column label="事由" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.reason || '--' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusMap[row.status]?.type ?? 'info'" size="small">
              {{ statusMap[row.status]?.label ?? row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 'pending' && activeTab === 'pending'">
              <el-popconfirm title="确认同意该申请？" @confirm="handleApprove(row.id)">
                <template #reference>
                  <el-button type="primary" size="small" link>同意</el-button>
                </template>
              </el-popconfirm>
              <el-button type="danger" size="small" link @click="handleReject(row)">驳回</el-button>
            </template>
            <el-popconfirm
              v-if="row.status === 'pending' && activeTab === 'mine'"
              title="确认撤回该申请？"
              @confirm="handleCancel(row.id)"
            >
              <template #reference>
                <el-button type="warning" size="small" link>撤回</el-button>
              </template>
            </el-popconfirm>
            <span v-if="!isActionable(row)" class="text-muted">--</span>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="loadData"
        />
      </div>
    </el-card>

    <ApprovalApplyDialog ref="applyDialogRef" @submitted="onSubmitted" />

    <el-dialog v-model="rejectVisible" title="驳回申请" width="400px">
      <el-input v-model="rejectNote" type="textarea" :rows="3" placeholder="请输入驳回原因（选填）" />
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" :loading="rejectLoading" @click="confirmReject">确认驳回</el-button>
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
import dayjs from 'dayjs'

const statusMap: Record<string, { label: string; type: '' | 'success' | 'warning' | 'danger' | 'info' }> = {
  draft: { label: '草稿', type: 'info' },
  pending: { label: '待审批', type: 'warning' },
  approved: { label: '已通过', type: 'success' },
  rejected: { label: '已驳回', type: 'danger' },
  cancelled: { label: '已撤回', type: 'info' },
  timeout: { label: '已过期', type: 'info' },
}

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

function formatTime(t: string) {
  return dayjs(t).format('MM-DD HH:mm')
}

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
    if (activeTab.value === 'pending') {
      params.status = 'pending'
    }
    const { data } = await approvalApi.list(params as Parameters<typeof approvalApi.list>[0])
    if (data) {
      list.value = data.list ?? []
      total.value = data.total
    }
  } catch {
    ElMessage.error('加载审批列表失败')
  } finally {
    loading.value = false
  }
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
  } catch {
    ElMessage.error('操作失败')
  }
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
  } catch {
    ElMessage.error('操作失败')
  } finally {
    rejectLoading.value = false
  }
}

async function handleCancel(id: number) {
  try {
    await approvalApi.cancel(id)
    ElMessage.success('已撤回')
    loadData()
  } catch {
    ElMessage.error('操作失败')
  }
}

function onSubmitted() {
  loadData()
  loadPendingCount()
}

onMounted(() => {
  loadData()
  loadPendingCount()
})
</script>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
.tab-badge {
  margin-left: 4px;
}
.text-muted {
  color: #c0c4cc;
}
</style>
