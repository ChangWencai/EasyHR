<template>
  <div class="si-records-table">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">参保记录</h1>
        <p class="page-subtitle">管理员工社保缴纳情况</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" size="large" @click="showExportDialog = true">
          <el-icon><Download /></el-icon>
          导出 Excel
        </el-button>
      </div>
    </header>

    <!-- 数据表格 -->
    <div class="table-card glass-card" v-loading="loading">
      <el-table
        :data="records"
        stripe
        :row-class-name="rowClassName"
        class="modern-table"
        :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
      >
        <el-table-column prop="employee_name" label="员工" min-width="110">
          <template #default="{ row }">
            <div class="employee-cell">
              <el-avatar :size="32" class="emp-avatar">{{ row.employee_name?.[0] }}</el-avatar>
              <span class="emp-name">{{ row.employee_name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="city" label="城市" width="110">
          <template #default="{ row }">
            <div class="city-cell">
              <el-icon><OfficeBuilding /></el-icon>
              {{ row.city }}
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="salary_base" label="社保基数" align="right" width="140">
          <template #default="{ row }">
            <span class="amount-text">{{ formatCurrency(row.salary_base) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="start_month" label="参保月" width="110">
          <template #default="{ row }">
            <span class="month-chip">{{ row.start_month }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="stop_month" label="停缴月" width="110">
          <template #default="{ row }">
            <span class="month-chip month-chip--stopped" v-if="row.stop_month">{{ row.stop_month }}</span>
            <span class="no-stop" v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column label="缴费渠道" width="130">
          <template #default="{ row }">
            <span class="channel-badge" :class="`channel--${row.payment_channel}`">
              {{ paymentChannelLabelMap[row.payment_channel] || '--' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <span class="status-badge" :class="`status--${row.status}`">
              <span class="status-dot"></span>
              {{ statusLabelMap[row.status] || row.status }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="个人月缴" align="right" width="130">
          <template #default="{ row }">
            <span class="amount-personal">{{ formatCurrency(row.monthly_personal) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="单位月缴" align="right" width="130">
          <template #default="{ row }">
            <span class="amount-company">{{ formatCurrency(row.monthly_company) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <el-button size="small" text type="primary" @click="openDetail(row)">
                <el-icon><View /></el-icon>
                详情
              </el-button>
              <el-button
                v-if="row.status === 'normal' || row.status === 'pending'"
                size="small"
                text
                type="danger"
                @click="openStopDialog(row)"
              >
                <el-icon><CircleClose /></el-icon>
                减员
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          :total="total"
          :page="page"
          :page-size="pageSize"
          layout="total, prev, pager, next"
          @current-change="fetchRecords"
        />
      </div>
    </div>

    <!-- 减员弹窗 -->
    <StopDialog
      v-model="showStopDialog"
      :employee-id="stopTarget?.employee_id"
      :employee-name="stopTarget?.employee_name"
      @success="onStopSuccess"
    />

    <!-- 详情弹窗 -->
    <SIDetailDialog
      v-model="showDetailDialog"
      :record-id="detailRecordId"
    />

    <!-- 导出弹窗 -->
    <el-dialog v-model="showExportDialog" title="选择导出内容" width="460px" class="export-dialog">
      <div class="export-content">
        <div class="export-options">
          <label class="export-option" :class="{ selected: exportType === 'current' }">
            <input type="radio" value="current" v-model="exportType" class="hidden-check" />
            <div class="option-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="option-text">
              <span class="option-title">当前页</span>
              <span class="option-desc">{{ records.length }} 条数据</span>
            </div>
          </label>
          <label class="export-option" :class="{ selected: exportType === 'full' }">
            <input type="radio" value="full" v-model="exportType" class="hidden-check" />
            <div class="option-icon option-icon--full">
              <el-icon><Files /></el-icon>
            </div>
            <div class="option-text">
              <span class="option-title">全部数据</span>
              <span class="option-desc">含五险分项明细</span>
            </div>
          </label>
        </div>
      </div>
      <template #footer>
        <el-button @click="showExportDialog = false" size="large">取消</el-button>
        <el-button type="primary" :loading="exporting" size="large" @click="doExport" class="export-btn">
          <el-icon><Download /></el-icon>
          导出
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import axios from '@/api/request'
import dayjs from 'dayjs'
import StopDialog from '@/components/socialinsurance/StopDialog.vue'
import SIDetailDialog from '@/components/socialinsurance/SIDetailDialog.vue'
import {
  Download, OfficeBuilding, View, CircleClose,
  Document, Files,
} from '@element-plus/icons-vue'

interface SIRecordRow {
  id: number; employee_id: number; employee_name: string; city: string
  salary_base: number; start_month: string; stop_month: string
  payment_channel: string; status: string; monthly_personal: number; monthly_company: number
}

const loading = ref(false)
const records = ref<SIRecordRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const statusLabelMap: Record<string, string> = {
  normal: '正常', pending: '待缴', overdue: '欠缴', transferred: '已转出', not_transferred: '未转出',
}

const paymentChannelLabelMap: Record<string, string> = {
  self: '自主缴费', agent_new: '代理新客', agent_existing: '代理已合作',
}

const showStopDialog = ref(false)
const stopTarget = ref<SIRecordRow | null>(null)
const showDetailDialog = ref(false)
const detailRecordId = ref<number | undefined>(undefined)
const showExportDialog = ref(false)
const exportType = ref('current')
const exporting = ref(false)

function formatCurrency(value: number | string | undefined): string {
  if (value === undefined || value === null) return '--'
  return '¥' + Number(value).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function rowClassName({ row }: { row: SIRecordRow }): string {
  return row.status === 'overdue' ? 'overdue-row' : ''
}

async function fetchRecords(p = 1): Promise<void> {
  page.value = p
  loading.value = true
  try {
    const res = await axios.get('/api/v1/social-insurance/monthly-records', {
      params: { page: p, page_size: pageSize.value },
    })
    const responseData = (res as { data?: { list: SIRecordRow[]; total: number } })?.data ?? res
    const data = responseData as { list: SIRecordRow[]; total: number }
    records.value = data.list || []
    total.value = data.total || 0
  } catch { ElMessage.error('加载记录失败') }
  finally { loading.value = false }
}

function openStopDialog(row: SIRecordRow): void {
  stopTarget.value = row
  showStopDialog.value = true
}

function onStopSuccess(): void { fetchRecords(page.value) }

function openDetail(row: SIRecordRow): void {
  detailRecordId.value = row.id
  showDetailDialog.value = true
}

async function doExport(): Promise<void> {
  exporting.value = true
  try {
    const params = exportType.value === 'full'
      ? { export: 'full', page_size: 9999 }
      : { page: page.value, page_size: pageSize.value }
    const res = await axios.get('/api/v1/social-insurance/records/export', {
      params,
      responseType: 'blob',
    })
    const blob = res as unknown as Blob
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `参保记录_${dayjs().format('YYYYMM')}.xlsx`
    a.click()
    URL.revokeObjectURL(url)
    showExportDialog.value = false
    ElMessage.success('导出成功')
  } catch { ElMessage.error('导出失败') }
  finally { exporting.value = false }
}

onMounted(() => fetchRecords())
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

.si-records-table { padding: 24px 32px; width: 100%; box-sizing: border-box; background: $bg-page; min-height: 100vh; }

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

.table-card { padding: 0; overflow: hidden; }

:deep(.modern-table) {
  .el-table__header th { padding: 14px 12px; font-size: 13px; }
  .el-table__row { transition: background 0.2s ease; &:hover > td { background: rgba(var(--primary), 0.02) !important; } }
  .el-table__cell { padding: 14px 12px; border-bottom: 1px solid #F3F4F6; }
}

:deep(.overdue-row) { background: rgba($error, 0.04) !important; }

.employee-cell { display: flex; align-items: center; gap: 10px; }
.emp-avatar { background: linear-gradient(135deg, var(--primary-light), var(--primary)); color: #fff; font-size: 13px; font-weight: 600; }
.emp-name { font-weight: 500; color: $text-primary; }

.city-cell { display: flex; align-items: center; gap: 6px; color: $text-secondary; font-size: 13px; .el-icon { color: $text-muted; } }

.amount-text { font-family: 'SF Mono', Monaco, monospace; font-weight: 600; color: $text-primary; }

.month-chip {
  display: inline-flex; padding: 3px 10px;
  background: #EDE9FE; color: var(--primary);
  font-size: 12px; font-weight: 600; border-radius: 12px;
  font-family: 'SF Mono', Monaco, monospace;

  &--stopped { background: #F3F4F6; color: $text-muted; }
}

.no-stop { color: $text-muted; }

.channel-badge {
  display: inline-flex; padding: 3px 10px;
  font-size: 12px; font-weight: 500; border-radius: 12px;

  &.channel--self          { background: #F3F4F6; color: $text-secondary; }
  &.channel--agent_new    { background: #FEF3C7; color: #D97706; }
  &.channel--agent_existing { background: #D1FAE5; color: #059669; }
}

.status-badge {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 4px 12px; font-size: 12px; font-weight: 500; border-radius: 20px;

  .status-dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; }

  &.status--normal   { background: #D1FAE5; color: #059669; }
  &.status--pending  { background: #FEF3C7; color: #D97706; }
  &.status--overdue  { background: #FEE2E2; color: #DC2626; }
  &.status--transferred { background: #F3F4F6; color: #6B7280; }
  &.status--not_transferred { background: #DBEAFE; color: #3B82F6; }
}

.amount-personal { font-family: 'SF Mono', Monaco, monospace; font-weight: 600; color: $text-primary; }
.amount-company { font-family: 'SF Mono', Monaco, monospace; font-weight: 700; color: var(--primary); }

.action-cell { display: flex; gap: 4px; }

.pagination-wrapper { display: flex; justify-content: flex-end; padding: 16px 20px; border-top: 1px solid $border-color; }

.export-content { padding: 4px 0; }

.export-options { display: flex; flex-direction: column; gap: 12px; }

.export-option {
  display: flex; align-items: center; gap: 16px;
  padding: 20px; border: 2px solid $border-color; border-radius: $radius-lg;
  cursor: pointer; transition: all 0.2s ease;

  &.selected { border-color: var(--primary); background: rgba(var(--primary), 0.04); }
  &:hover:not(.selected) { border-color: rgba(var(--primary), 0.4); transform: translateY(-2px); }
  .hidden-check { display: none; }
}

.option-icon {
  width: 44px; height: 44px;
  border-radius: $radius-md;
  background: linear-gradient(135deg, #EDE9FE, #DDD6FE);
  color: var(--primary);
  display: flex; align-items: center; justify-content: center; font-size: 20px;

  &--full { background: linear-gradient(135deg, #DBEAFE, #BFDBFE); color: #3B82F6; }
}

.option-text { display: flex; flex-direction: column; gap: 2px; }
.option-title { font-size: 15px; font-weight: 600; color: $text-primary; }
.option-desc { font-size: 12px; color: $text-muted; }

.export-btn {
  background: linear-gradient(135deg, var(--primary-light), var(--primary));
  border: none;
  box-shadow: 0 4px 14px rgba(var(--primary), 0.4);
  &:hover { box-shadow: 0 6px 20px rgba(var(--primary), 0.5); }
}

@media (max-width: 768px) {
  .si-records-table { padding: 16px; }
}
</style>
