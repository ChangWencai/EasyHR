<template>
  <div class="si-records-table">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>参保记录</span>
          <div class="header-actions">
            <el-button type="primary" size="small" @click="showExportDialog = true">导出 Excel</el-button>
          </div>
        </div>
      </template>

      <el-table
        :data="records"
        stripe
        v-loading="loading"
        :row-class-name="rowClassName"
      >
        <el-table-column prop="employee_name" label="员工" min-width="80" />
        <el-table-column prop="city" label="城市" min-width="80" />
        <el-table-column prop="salary_base" label="社保基数" min-width="100">
          <template #default="{ row }">
            {{ formatCurrency(row.salary_base) }}
          </template>
        </el-table-column>
        <el-table-column prop="start_month" label="参保月" min-width="90" />
        <el-table-column prop="stop_month" label="停缴月" min-width="90">
          <template #default="{ row }">{{ row.stop_month || '--' }}</template>
        </el-table-column>
        <el-table-column label="缴费渠道" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="paymentChannelTagType[row.payment_channel] || 'info'">
              {{ paymentChannelLabelMap[row.payment_channel] || '--' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag
              :type="statusTagType[row.status] || 'info'"
              size="small"
              :class="{ 'status-blue': row.status === 'not_transferred' }"
            >
              {{ statusLabelMap[row.status] || row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="个人月缴" min-width="100">
          <template #default="{ row }">
            {{ formatCurrency(row.monthly_personal) }}
          </template>
        </el-table-column>
        <el-table-column label="单位月缴" min-width="100">
          <template #default="{ row }">
            {{ formatCurrency(row.monthly_company) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button size="small" link @click="openDetail(row)">详情</el-button>
            <el-button
              v-if="row.status === 'normal' || row.status === 'pending'"
              size="small"
              link
              type="danger"
              @click="openStopDialog(row)"
            >
              减员
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="mt-4"
        layout="total, prev, pager, next"
        :total="total"
        :page="page"
        :page-size="pageSize"
        @current-change="fetchRecords"
      />
    </el-card>

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
    <el-dialog v-model="showExportDialog" title="选择导出内容" width="420px">
      <el-form>
        <el-form-item>
          <el-radio-group v-model="exportType">
            <el-radio value="current">当前页（{{ records.length }} 条）</el-radio>
            <el-radio value="full">全部数据（含五险分项明细）</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showExportDialog = false">取消</el-button>
        <el-button type="primary" :loading="exporting" @click="doExport">导出</el-button>
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

interface SIRecordRow {
  id: number
  employee_id: number
  employee_name: string
  city: string
  salary_base: number
  start_month: string
  stop_month: string
  payment_channel: string
  status: string
  monthly_personal: number
  monthly_company: number
}

const loading = ref(false)
const records = ref<SIRecordRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const statusTagType: Record<string, string> = {
  normal: 'success',
  pending: 'warning',
  overdue: 'danger',
  transferred: 'info',
  not_transferred: '',
}

const statusLabelMap: Record<string, string> = {
  normal: '正常',
  pending: '待缴',
  overdue: '欠缴',
  transferred: '已转出',
  not_transferred: '未转出',
}

const paymentChannelLabelMap: Record<string, string> = {
  self: '自主缴费',
  agent_new: '代理新客',
  agent_existing: '代理已合作',
}

const paymentChannelTagType: Record<string, string> = {
  self: '',
  agent_new: 'warning',
  agent_existing: 'success',
}

// Stop dialog
const showStopDialog = ref(false)
const stopTarget = ref<SIRecordRow | null>(null)

// Detail dialog
const showDetailDialog = ref(false)
const detailRecordId = ref<number | undefined>(undefined)

// Export dialog
const showExportDialog = ref(false)
const exportType = ref('current')
const exporting = ref(false)

function formatCurrency(value: number | string | undefined): string {
  if (value === undefined || value === null) return '--'
  return Number(value).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function rowClassName({ row }: { row: SIRecordRow }): string {
  return row.status === 'overdue' ? 'overdue-row' : ''
}

async function fetchRecords(p = 1): Promise<void> {
  page.value = p
  loading.value = true
  try {
    const res = await axios.get('/api/v1/socialinsurance/records', {
      params: { page: p, page_size: pageSize.value },
    })
    const responseData = (res as { data?: { list: SIRecordRow[]; total: number } })?.data ?? res
    const data = responseData as { list: SIRecordRow[]; total: number }
    records.value = data.list || []
    total.value = data.total || 0
  } catch {
    ElMessage.error('加载记录失败')
  } finally {
    loading.value = false
  }
}

function openStopDialog(row: SIRecordRow): void {
  stopTarget.value = row
  showStopDialog.value = true
}

function onStopSuccess(): void {
  fetchRecords(page.value)
}

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

    const res = await axios.get('/api/v1/socialinsurance/records/export', {
      params,
      responseType: 'blob',
    })
    const blob = res as unknown as Blob
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    const ym = dayjs().format('YYYYMM')
    a.download = `参保记录_${ym}.xlsx`
    a.click()
    URL.revokeObjectURL(url)
    showExportDialog.value = false
    ElMessage.success('导出成功')
  } catch {
    ElMessage.error('导出失败')
  } finally {
    exporting.value = false
  }
}

onMounted(() => {
  fetchRecords()
})
</script>

<style scoped lang="scss">
.si-records-table {
  padding: 0;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.mt-4 {
  margin-top: 16px;
}

.status-blue {
  --el-tag-bg-color: #4f6ef720;
  --el-tag-text-color: #4f6ef7;
  --el-tag-border-color: #4f6ef740;
}

:deep(.overdue-row) {
  background: #ff563010 !important;
}
</style>
