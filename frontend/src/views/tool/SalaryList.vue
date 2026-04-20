<template>
  <div class="salary-list">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">薪资列表</h1>
        <p class="page-subtitle">{{ filterYM || '请选择月份' }}</p>
      </div>
      <div class="header-actions">
        <el-button @click="showExportDialog = true">
          <el-icon><Download /></el-icon>
          导出 Excel
        </el-button>
      </div>
    </header>

    <!-- 筛选栏 -->
    <div class="filter-bar glass-card">
      <div class="filter-group">
        <el-date-picker
          v-model="filterYM"
          type="month"
          placeholder="选择月份"
          value-format="YYYY-MM"
          size="large"
          class="filter-date"
          @change="loadList"
        />
        <el-select
          v-model="filterDeptId"
          placeholder="全部部门"
          clearable
          class="filter-select"
          @change="loadList"
        >
          <template #prefix>
            <el-icon><OfficeBuilding /></el-icon>
          </template>
          <el-option v-for="dept in departments" :key="dept.id" :label="dept.name" :value="dept.id" />
        </el-select>
        <div class="search-wrapper">
          <el-icon class="search-icon"><Search /></el-icon>
          <input
            v-model="filterKeyword"
            type="text"
            placeholder="搜索员工姓名..."
            class="search-input"
            @input="debouncedLoad"
          />
        </div>
      </div>
    </div>

    <!-- 统计概览 -->
    <div v-if="listData.length > 0" class="stats-overview">
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--total">
          <el-icon><Money /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ listData.length }}</span>
          <span class="stat-label">发放人数</span>
        </div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--gross">
          <el-icon><Coin /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value stat-value--money">
            ¥{{ totalGross.toLocaleString() }}
          </span>
          <span class="stat-label">应发合计</span>
        </div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--tax">
          <el-icon><Document /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value stat-value--tax">
            ¥{{ totalTax.toLocaleString() }}
          </span>
          <span class="stat-label">个税合计</span>
        </div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--net">
          <el-icon><Wallet /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value stat-value--net">
            ¥{{ totalNet.toLocaleString() }}
          </span>
          <span class="stat-label">实发合计</span>
        </div>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="table-container glass-card">
      <el-table
        :data="listData"
        stripe
        v-loading="loading"
        class="modern-table"
        :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
      >
        <el-table-column prop="employee_name" label="员工" min-width="120" fixed="left">
          <template #header>
            <el-tooltip content="点击查看员工详情" placement="top" :show-after="500">
              <span>员工</span>
            </el-tooltip>
          </template>
          <template #default="{ row }">
            <div class="employee-cell">
              <el-avatar :size="36" class="employee-avatar">
                {{ row.employee_name?.[0] || '?' }}
              </el-avatar>
              <div class="employee-info">
                <span class="employee-name">{{ row.employee_name }}</span>
                <span class="employee-dept">{{ row.department_name || '未分配' }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="应发合计" min-width="130">
          <template #default="{ row }">
            <span class="amount gross">
              ¥{{ (row.gross_income || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2 }) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="扣除" min-width="110">
          <template #default="{ row }">
            <span class="amount deduction">
              -¥{{ (row.total_deductions || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2 }) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="社保公积金" min-width="130">
          <template #default="{ row }">
            <span class="amount si">
              -¥{{ (row.si_deduction || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2 }) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="个税" min-width="110">
          <template #default="{ row }">
            <span class="amount tax">
              -¥{{ (row.tax || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2 }) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="实发" min-width="130" fixed="right">
          <template #header>
            <el-tooltip content="点击查看月度工资详情" placement="top" :show-after="500">
              <span>实发</span>
            </el-tooltip>
          </template>
          <template #default="{ row }">
            <span class="amount net">
              ¥{{ (row.net_income || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2 }) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template #header>
            <el-tooltip content="草稿→已核算→已确认→已发放" placement="top" :show-after="500">
              <span>状态</span>
            </el-tooltip>
          </template>
          <template #default="{ row }">
            <div class="status-cell">
              <span class="status-badge" :class="`status--${row.status}`">
                {{ statusMap[row.status] || row.status }}
              </span>
              <el-icon
                v-if="['confirmed', 'paid'].includes(row.status)"
                class="lock-icon"
                @click="openUnlockDialog(row)"
              >
                <Lock />
              </el-icon>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="loadList"
        />
      </div>
    </div>

    <!-- 解锁弹窗 -->
    <el-dialog
      v-model="showUnlockDialog"
      title="解锁工资记录"
      width="420px"
      class="unlock-dialog"
    >
      <div class="unlock-content">
        <el-alert type="warning" :closable="false" show-icon>
          <template #title>
            该月工资已确认锁定
          </template>
          <template #default>
            如需修改请输入解锁码
          </template>
        </el-alert>
        <el-form class="unlock-form">
          <el-form-item label="手机号">
            <el-input
              v-model="unlockPhone"
              placeholder="企业主手机号"
              size="large"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item v-if="codeSent" label="验证码">
            <el-input
              v-model="unlockCode"
              placeholder="请输入6位验证码"
              maxlength="6"
              size="large"
              style="width: 100%"
            />
          </el-form-item>
          <el-button
            v-if="!codeSent"
            type="primary"
            :loading="sendingCode"
            @click="sendUnlockCode"
            class="send-code-btn"
          >
            发送验证码
          </el-button>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="showUnlockDialog = false">取消</el-button>
        <el-button
          type="primary"
          :loading="unlocking"
          :disabled="!codeSent"
          @click="doUnlock"
        >
          确认解锁
        </el-button>
      </template>
    </el-dialog>

    <!-- 导出弹窗 -->
    <el-dialog
      v-model="showExportDialog"
      title="选择导出内容"
      width="420px"
      class="export-dialog"
    >
      <el-form>
        <el-form-item>
          <el-radio-group v-model="exportType" class="export-options">
            <el-radio value="current">
              <div class="radio-label">
                <span class="radio-title">当前页</span>
                <span class="radio-desc">{{ listData.length }} 条数据</span>
              </div>
            </el-radio>
            <el-radio value="full">
              <div class="radio-label">
                <span class="radio-title">全部数据</span>
                <span class="radio-desc">含税前明细</span>
              </div>
            </el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showExportDialog = false">取消</el-button>
        <el-button type="primary" :loading="exporting" @click="doExport">
          <el-icon><Download /></el-icon>
          导出
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { salaryApi } from '@/api/salary'
import { departmentApi } from '@/api/department'
import { ElMessage } from 'element-plus'
import { Lock, Search, Download, Money, Coin, Document, Wallet, OfficeBuilding } from '@element-plus/icons-vue'
import { useThrottleFn } from '@vueuse/core'
import dayjs from 'dayjs'

interface SalaryRecord {
  id: number
  employee_id: number
  employee_name: string
  department_name: string
  gross_income: number
  total_deductions: number
  tax: number
  si_deduction: number
  net_income: number
  status: string
}

const filterYM = ref(dayjs().format('YYYY-MM'))
const filterDeptId = ref<number | undefined>(undefined)
const filterKeyword = ref('')
const loading = ref(false)
const listData = ref<SalaryRecord[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const departments = ref<{ id: number; name: string }[]>([])

// 解锁
const showUnlockDialog = ref(false)
const unlockTarget = ref<SalaryRecord | null>(null)
const unlockPhone = ref('')
const unlockCode = ref('')
const sendingCode = ref(false)
const codeSent = ref(false)
const unlocking = ref(false)

// 导出
const showExportDialog = ref(false)
const exportType = ref('current')
const exporting = ref(false)

// 统计
const totalGross = computed(() =>
  listData.value.reduce((sum, r) => sum + (r.gross_income || 0), 0)
)

const totalTax = computed(() =>
  listData.value.reduce((sum, r) => sum + (r.tax || 0), 0)
)

const totalNet = computed(() =>
  listData.value.reduce((sum, r) => sum + (r.net_income || 0), 0)
)

const statusMap: Record<string, string> = {
  draft: '草稿',
  calculated: '已核算',
  confirmed: '已确认',
  paid: '已发放',
}

async function loadDepartments() {
  try {
    departments.value = await departmentApi.list()
  } catch { /* ignore */ }
}

async function loadList(p = 1) {
  if (!filterYM.value) return
  page.value = p
  loading.value = true
  try {
    const [yearStr, monthStr] = filterYM.value.split('-')
    const year = parseInt(yearStr, 10)
    const month = parseInt(monthStr, 10)
    const res = await salaryApi.getSalaryList({
      year,
      month,
      department_id: filterDeptId.value,
      keyword: filterKeyword.value || undefined,
      page: p,
      page_size: pageSize.value,
    })
    const data = (res as any)
    listData.value = data.list || []
    total.value = data.total || 0
  } catch {
    ElMessage.error('加载薪资列表失败')
  } finally {
    loading.value = false
  }
}

const debouncedLoad = useThrottleFn(() => loadList(), 400)

function openUnlockDialog(row: SalaryRecord) {
  unlockTarget.value = row
  unlockPhone.value = ''
  unlockCode.value = ''
  codeSent.value = false
  showUnlockDialog.value = true
}

async function sendUnlockCode() {
  if (!unlockPhone.value) {
    ElMessage.warning('请输入手机号')
    return
  }
  sendingCode.value = true
  try {
    await salaryApi.sendUnlockCode({ phone: unlockPhone.value })
    codeSent.value = true
    ElMessage.success('验证码已发送')
  } catch {
    ElMessage.error('发送失败')
  } finally {
    sendingCode.value = false
  }
}

async function doUnlock() {
  if (!unlockTarget.value || !unlockCode.value) return
  unlocking.value = true
  try {
    await salaryApi.unlockRecord({
      record_id: unlockTarget.value.id,
      sms_code: unlockCode.value,
    })
    ElMessage.success('已解锁，可重新编辑')
    showUnlockDialog.value = false
    loadList()
  } catch {
    ElMessage.error('解锁失败，验证码错误')
  } finally {
    unlocking.value = false
  }
}

async function doExport() {
  if (!filterYM.value) return
  const [yearStr, monthStr] = filterYM.value.split('-')
  const year = parseInt(yearStr, 10)
  const month = parseInt(monthStr, 10)
  exporting.value = true
  try {
    const blob = await salaryApi.exportWithDetails(year, month)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `工资表_${year}年${month}月.xlsx`
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

onMounted(async () => {
  await loadDepartments()
  await loadList()
})
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
$shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
$shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
$radius-sm: 8px;
$radius-md: 12px;
$radius-lg: 16px;
$radius-xl: 24px;

.salary-list {
  padding: 24px 32px;
  width: 100%;
  box-sizing: border-box;
  background: $bg-page;
  min-height: 100vh;
}

.glass-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: $radius-xl;
  box-shadow: $shadow-md;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;

  .page-title {
    font-size: 24px;
    font-weight: 700;
    color: $text-primary;
    margin: 0 0 4px;
  }

  .page-subtitle {
    font-size: 14px;
    color: $text-secondary;
    margin: 0;
  }
}

.header-actions {
  :deep(.el-button) {
    padding: 12px 24px;
    border-radius: $radius-md;
    font-weight: 600;
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
}

.filter-bar {
  padding: 16px 20px;
  margin-bottom: 20px;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.filter-date { width: 160px; }
.filter-select { width: 160px; }

.search-wrapper {
  position: relative;
  flex: 1;
  max-width: 280px;
}

.search-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  color: $text-muted;
  font-size: 18px;
}

.search-input {
  width: 100%;
  padding: 10px 14px 10px 42px;
  font-size: 14px;
  color: $text-primary;
  background: $bg-page;
  border: 1px solid $border-color;
  border-radius: $radius-md;
  outline: none;
  transition: all 0.2s ease;

  &::placeholder { color: $text-muted; }
  &:focus {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(var(--primary), 0.1);
  }
}

.stats-overview {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  transition: all 0.2s ease;

  &:hover { transform: translateY(-2px); box-shadow: 0 8px 16px rgba(0,0,0,0.08); }
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: $radius-md;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;

  &--total { background: #EDE9FE; color: var(--primary); }
  &--gross { background: #D1FAE5; color: $success; }
  &--tax { background: #FEF3C7; color: $warning; }
  &--net { background: #DBEAFE; color: #3B82F6; }
}

.stat-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: $text-primary;
  line-height: 1;

  &--money { color: $success; }
  &--tax { color: $warning; }
  &--net { color: #3B82F6; }
}

.stat-label {
  font-size: 12px;
  color: $text-secondary;
}

.table-container {
  padding: 0;
  overflow: hidden;
}

:deep(.modern-table) {
  .el-table__header th { padding: 16px 12px; font-size: 13px; }
  .el-table__row {
    transition: background 0.2s ease;
    &:hover > td { background: rgba(var(--primary), 0.02) !important; }
  }
  .el-table__cell {
    padding: 16px 12px;
    border-bottom: 1px solid #F3F4F6;
  }
}

.employee-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.employee-avatar {
  background: linear-gradient(135deg, var(--primary-light), var(--primary));
  color: #fff;
  font-size: 14px;
  font-weight: 600;
}

.employee-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.employee-name {
  font-weight: 500;
  color: $text-primary;
}

.employee-dept {
  font-size: 12px;
  color: $text-muted;
}

.amount {
  font-family: 'SF Mono', Monaco, monospace;
  font-weight: 600;
  font-size: 14px;
  &.gross { color: $success; }
  &.deduction { color: $error; }
  &.si { color: $warning; }
  &.tax { color: $warning; }
  &.net { color: #3B82F6; font-weight: 700; }
}

.status-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 20px;

  &.status--draft { background: #F3F4F6; color: #6B7280; }
  &.status--calculated { background: #FEF3C7; color: #D97706; }
  &.status--confirmed { background: #DBEAFE; color: #2563EB; }
  &.status--paid { background: #D1FAE5; color: #059669; }
}

.lock-icon {
  cursor: pointer;
  color: $text-muted;
  font-size: 14px;
  &:hover { color: var(--primary); }
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: 16px 20px;
  border-top: 1px solid $border-color;
}

.unlock-content {
  .unlock-form {
    margin-top: 20px;
  }

  .send-code-btn {
    width: 100%;
    margin-top: 8px;
  }
}

.export-options {
  display: flex;
  flex-direction: column;
  gap: 12px;

  :deep(.el-radio) {
    padding: 16px;
    border: 1px solid $border-color;
    border-radius: $radius-md;
    margin-right: 0;
    width: 100%;
    transition: all 0.2s ease;

    &:hover { border-color: var(--primary-light); }
    &.is-checked { border-color: var(--primary); background: rgba(var(--primary), 0.04); }
  }
}

.radio-label {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-left: 8px;
}

.radio-title {
  font-weight: 600;
  color: $text-primary;
}

.radio-desc {
  font-size: 12px;
  color: $text-muted;
}

@media (max-width: 1024px) {
  .stats-overview { grid-template-columns: repeat(2, 1fr); }
}

@media (max-width: 768px) {
  .salary-list { padding: 16px; }
  .filter-group { flex-direction: column; align-items: stretch; }
  .search-wrapper { max-width: none; }
  .stats-overview { grid-template-columns: 1fr; }
}
</style>
