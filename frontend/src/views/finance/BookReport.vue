<template>
  <div class="book-report">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">账簿报表</h1>
        <p class="page-subtitle">科目余额表、财务报表与期间管理</p>
      </div>
    </header>

    <!-- 标签页导航 -->
    <div class="nav-tabs glass-card">
      <div class="tab-group">
        <button
          v-for="tab in navTabs"
          :key="tab.value"
          class="tab-btn"
          :class="{ active: activeTab === tab.value }"
          @click="activeTab = tab.value; if(tab.value==='period') loadPeriodList()"
        >
          <el-icon><component :is="tab.icon" /></el-icon>
          {{ tab.label }}
        </button>
      </div>
    </div>

    <!-- 科目余额表 -->
    <div v-show="activeTab === 'trial'" class="tab-content">
      <div class="toolbar-card glass-card">
        <div class="toolbar-left">
          <el-select
            v-model="trialPeriod"
            placeholder="选择会计期间"
            size="large"
            class="period-select"
            @change="loadTrialBalance"
          >
            <template #prefix>
              <el-icon><Calendar /></el-icon>
            </template>
            <el-option v-for="p in periods" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </div>
        <el-button
          :disabled="!trialPeriod"
          size="large"
          @click="handleExportTrial"
          class="export-btn"
        >
          <el-icon><Download /></el-icon>
          导出Excel
        </el-button>
      </div>

      <div class="table-card glass-card" v-loading="trialLoading">
        <el-table
          :data="trialData"
          stripe
          border
          show-summary
          class="modern-table"
          :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
        >
          <el-table-column prop="account_code" label="编码" width="130">
            <template #default="{ row }">
              <span class="code-text">{{ row.account_code }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="account_name" label="名称">
            <template #default="{ row }">
              <span class="name-text">{{ row.account_name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="opening_balance" label="期初余额" align="right" width="150">
            <template #default="{ row }">
              <span class="amount-text">{{ formatAmount(row.opening_balance) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="debit_amount" label="借方发生额" align="right" width="150">
            <template #default="{ row }">
              <span class="amount-text amount-text--debit">{{ formatAmount(row.debit_amount) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="credit_amount" label="贷方发生额" align="right" width="150">
            <template #default="{ row }">
              <span class="amount-text amount-text--credit">{{ formatAmount(row.credit_amount) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="closing_balance" label="期末余额" align="right" width="150">
            <template #default="{ row }">
              <span class="amount-text amount-text--bold">{{ formatAmount(row.closing_balance) }}</span>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <!-- 报表 -->
    <div v-show="activeTab === 'report'" class="tab-content">
      <div class="toolbar-card glass-card">
        <div class="toolbar-left">
          <el-select
            v-model="reportPeriod"
            placeholder="选择会计期间"
            size="large"
            class="period-select"
          >
            <template #prefix>
              <el-icon><Calendar /></el-icon>
            </template>
            <el-option v-for="p in periods" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </div>
        <div class="report-btns">
          <el-button
            :disabled="!reportPeriod"
            size="large"
            type="primary"
            @click="loadBalanceSheet"
            class="report-btn"
          >
            <el-icon><DataLine /></el-icon>
            资产负债表
          </el-button>
          <el-button
            :disabled="!reportPeriod"
            size="large"
            type="primary"
            @click="loadIncomeStatement"
            class="report-btn"
          >
            <el-icon><TrendCharts /></el-icon>
            利润表
          </el-button>
        </div>
      </div>

      <!-- 资产负债表 -->
      <div v-if="balanceSheetData" class="report-card glass-card">
        <div class="report-card-header">
          <div class="report-title-group">
            <div class="report-icon">
              <el-icon><DataLine /></el-icon>
            </div>
            <div>
              <h3 class="report-title">资产负债表</h3>
              <p class="report-subtitle">截至 {{ reportPeriod }} 期末</p>
            </div>
          </div>
        </div>
        <div class="report-grid">
          <div class="report-item">
            <span class="report-label">资产总计</span>
            <span class="report-value report-value--primary">¥{{ formatAmount(balanceSheetData.total_assets) }}</span>
          </div>
          <div class="report-item">
            <span class="report-label">负债合计</span>
            <span class="report-value">¥{{ formatAmount(balanceSheetData.total_liabilities) }}</span>
          </div>
          <div class="report-item">
            <span class="report-label">所有者权益</span>
            <span class="report-value report-value--success">¥{{ formatAmount(balanceSheetData.total_equity) }}</span>
          </div>
        </div>
      </div>

      <!-- 利润表 -->
      <div v-if="incomeStatementData" class="report-card glass-card">
        <div class="report-card-header">
          <div class="report-title-group">
            <div class="report-icon report-icon--profit">
              <el-icon><TrendCharts /></el-icon>
            </div>
            <div>
              <h3 class="report-title">利润表</h3>
              <p class="report-subtitle">{{ reportPeriod }} 期间</p>
            </div>
          </div>
        </div>
        <div class="report-grid">
          <div class="report-item">
            <span class="report-label">营业收入</span>
            <span class="report-value report-value--success">¥{{ formatAmount(incomeStatementData.revenue) }}</span>
          </div>
          <div class="report-item">
            <span class="report-label">营业成本</span>
            <span class="report-value report-value--danger">¥{{ formatAmount(incomeStatementData.cost) }}</span>
          </div>
          <div class="report-item">
            <span class="report-label">净利润</span>
            <span class="report-value report-value--primary">¥{{ formatAmount(incomeStatementData.net_profit) }}</span>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="!balanceSheetData && !incomeStatementData" class="empty-report glass-card">
        <div class="empty-icon">
          <el-icon><Document /></el-icon>
        </div>
        <h3>暂无报表数据</h3>
        <p>请先选择会计期间，然后点击上方按钮生成报表</p>
      </div>
    </div>

    <!-- 期间管理 -->
    <div v-show="activeTab === 'period'" class="tab-content">
      <div class="toolbar-card glass-card">
        <div class="toolbar-left">
          <span class="toolbar-label">
            <el-icon><Clock /></el-icon>
            共 {{ periodList.length }} 个会计期间
          </span>
        </div>
        <el-button size="large" @click="loadPeriodList" class="refresh-btn">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>

      <div class="table-card glass-card" v-loading="periodLoading">
        <el-table
          :data="periodList"
          stripe
          class="modern-table"
          :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
        >
          <el-table-column prop="name" label="期间名称" width="160">
            <template #default="{ row }">
              <div class="period-name">
                <el-icon><Calendar /></el-icon>
                {{ row.name }}
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="year" label="年度" width="100">
            <template #default="{ row }">
              <span class="year-badge">{{ row.year }}年</span>
            </template>
          </el-table-column>
          <el-table-column prop="month" label="月份" width="90">
            <template #default="{ row }">
              <span class="month-badge">{{ row.month }}月</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="110">
            <template #default="{ row }">
              <span class="status-badge" :class="`status--${row.status}`">
                <span class="status-dot"></span>
                {{ periodStatusLabel(row.status) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <div class="action-cell">
                <el-button
                  v-if="row.status === 'open'"
                  size="small"
                  type="warning"
                  class="action-btn"
                  @click="handleClosePeriod(row)"
                >
                  <el-icon><Lock /></el-icon>
                  结账
                </el-button>
                <el-button
                  v-if="row.status === 'closed'"
                  size="small"
                  type="primary"
                  class="action-btn"
                  @click="handleRevertPeriod(row)"
                >
                  <el-icon><Unlock /></el-icon>
                  反结账
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { financeApi } from '@/api/finance'
import {
  Download, Calendar, DataLine, TrendCharts, Document,
  Lock, Unlock, Refresh, Clock,
} from '@element-plus/icons-vue'

interface Period { id: number; name: string; year: number; month: number; status: string }

interface TrialRow {
  account_code: string; account_name: string
  opening_balance: string; debit_amount: string; credit_amount: string; closing_balance: string
}

const navTabs = [
  { label: '科目余额表', value: 'trial',  icon: 'FolderOpened' },
  { label: '财务报表',   value: 'report',  icon: 'DataLine'    },
  { label: '期间管理',   value: 'period',  icon: 'Clock'       },
]

const activeTab = ref('trial')

const periods = ref<Period[]>([])
const trialPeriod = ref<number | null>(null)
const trialData = ref<TrialRow[]>([])
const trialLoading = ref(false)

const reportPeriod = ref<number | null>(null)
const balanceSheetData = ref<any | null>(null)
const incomeStatementData = ref<any | null>(null)

const periodList = ref<Period[]>([])
const periodLoading = ref(false)

const closing = ref(false)
const reverting = ref(false)

function formatAmount(val?: string | number) {
  if (val === undefined || val === null) return '0.00'
  const num = typeof val === 'string' ? parseFloat(val) : val
  if (isNaN(num)) return '0.00'
  return num.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function periodStatusLabel(status: string) {
  const map: Record<string, string> = { open: '开放', closed: '已结账' }
  return map[status] || status
}

async function loadPeriods() {
  try {
    const res = await financeApi.periods() as any
    periods.value = res.data?.data || res.data || []
  } catch { periods.value = [] }
}

async function loadTrialBalance() {
  if (!trialPeriod.value) return
  trialLoading.value = true
  try {
    const res = await financeApi.trialBalance(trialPeriod.value) as any
    trialData.value = res.data?.data || res.data?.rows || []
  } catch { trialData.value = [] }
  finally { trialLoading.value = false }
}

async function handleExportTrial() {
  if (!trialPeriod.value) return
  try {
    const res = await financeApi.bookExport(trialPeriod.value) as any
    const blob = new Blob([res], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url; a.download = `科目余额表_${trialPeriod.value}.xlsx`; a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch { ElMessage.error('导出失败') }
}

async function loadBalanceSheet() {
  if (!reportPeriod.value) return
  try {
    const res = await financeApi.balanceSheet(reportPeriod.value) as any
    balanceSheetData.value = res.data?.data || res.data || null
    ElMessage.success('资产负债表加载成功')
  } catch { ElMessage.error('加载失败') }
}

async function loadIncomeStatement() {
  if (!reportPeriod.value) return
  try {
    const res = await financeApi.incomeStatement(reportPeriod.value) as any
    incomeStatementData.value = res.data?.data || res.data || null
    ElMessage.success('利润表加载成功')
  } catch { ElMessage.error('加载失败') }
}

async function loadPeriodList() {
  periodLoading.value = true
  try {
    const res = await financeApi.periods() as any
    periodList.value = res.data?.data || res.data || []
  } catch { periodList.value = [] }
  finally { periodLoading.value = false }
}

async function handleClosePeriod(row: Period) {
  if (closing.value) return
  closing.value = true
  try {
    await ElMessageBox.confirm(`确定要对「${row.name}」进行结账吗？`, '结账确认', { type: 'warning' })
    await financeApi.closePeriod(row.id)
    ElMessage.success('结账成功')
    loadPeriodList()
  } catch (e: unknown) {
    if ((e as any)?.message !== 'cancel') {
      const msg = (e as any)?.response?.data?.error || '结账失败'
      ElMessage.error(msg)
    }
  } finally { closing.value = false }
}

async function handleRevertPeriod(row: Period) {
  if (reverting.value) return
  reverting.value = true
  try {
    await ElMessageBox.confirm(`确定要对「${row.name}」进行反结账吗？此操作不可逆。`, '反结账确认', { type: 'warning' })
    await financeApi.revertPeriod(row.id)
    ElMessage.success('反结账成功')
    loadPeriodList()
  } catch (e: unknown) {
    if ((e as any)?.message !== 'cancel') {
      const msg = (e as any)?.response?.data?.error || '反结账失败'
      ElMessage.error(msg)
    }
  } finally { reverting.value = false }
}

onMounted(() => { loadPeriods(); loadPeriodList() })
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

.book-report { padding: 24px 32px; width: 100%; box-sizing: border-box; background: $bg-page; min-height: 100vh; }

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

.nav-tabs { padding: 14px 20px; margin-bottom: 20px; }

.tab-group { display: inline-flex; background: #F3F4F6; border-radius: $radius-md; padding: 4px; }

.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 20px;
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

.tab-content { display: flex; flex-direction: column; gap: 16px; }

.toolbar-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
}

.toolbar-left { display: flex; align-items: center; gap: 12px; }
.toolbar-label { display: flex; align-items: center; gap: 6px; font-size: 14px; color: $text-secondary; }

.period-select { width: 200px; }

.export-btn {
  border-style: dashed;
  color: var(--primary);
  border-color: rgba(var(--primary), 0.4);
  background: rgba(var(--primary), 0.04);
  &:hover { background: rgba(var(--primary), 0.08); border-color: var(--primary); }
}

.report-btns { display: flex; gap: 8px; }
.report-btn {
  background: linear-gradient(135deg, var(--primary-light), var(--primary));
  border: none;
  box-shadow: 0 4px 14px rgba(var(--primary), 0.3);
  &:hover { box-shadow: 0 6px 20px rgba(var(--primary), 0.4); }
}

.table-card { padding: 0; overflow: hidden; }

:deep(.modern-table) {
  .el-table__header th { padding: 14px 16px; font-size: 13px; }
  .el-table__row { transition: background 0.2s ease; &:hover > td { background: rgba(var(--primary), 0.02) !important; } }
  .el-table__cell { padding: 14px 16px; border-bottom: 1px solid #F3F4F6; }
}

.code-text { font-family: 'SF Mono', Monaco, monospace; font-weight: 600; color: var(--primary); font-size: 13px; }
.name-text { font-weight: 500; color: $text-primary; }

.amount-text { font-family: 'SF Mono', Monaco, monospace; font-size: 14px; font-weight: 500; color: $text-primary; &--debit { color: $error; } &--credit { color: $success; } &--bold { font-weight: 700; } }

.report-card {
  padding: 24px;
  animation: fadeInUp 0.3s ease;
}

.report-card-header { margin-bottom: 24px; }
.report-title-group { display: flex; align-items: center; gap: 12px; }

.report-icon {
  width: 48px; height: 48px;
  border-radius: $radius-md;
  background: linear-gradient(135deg, #DBEAFE, #BFDBFE);
  color: #3B82F6;
  display: flex; align-items: center; justify-content: center;
  font-size: 22px;

  &--profit { background: linear-gradient(135deg, #D1FAE5, #A7F3D0); color: $success; }
}

.report-title { font-size: 18px; font-weight: 700; color: $text-primary; margin: 0 0 4px; }
.report-subtitle { font-size: 13px; color: $text-muted; margin: 0; }

.report-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }

.report-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 20px;
  background: $bg-page;
  border-radius: $radius-md;
  border: 1px solid $border-color;
}

.report-label { font-size: 13px; color: $text-muted; }
.report-value { font-size: 20px; font-weight: 700; color: $text-primary; font-family: 'SF Mono', Monaco, monospace;
  &--primary { color: var(--primary); }
  &--success { color: $success; }
  &--danger  { color: $error; }
}

.empty-report {
  text-align: center;
  padding: 80px 32px;
  .empty-icon { width: 72px; height: 72px; margin: 0 auto 16px; background: linear-gradient(135deg, #EDE9FE, #DDD6FE); border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 32px; color: var(--primary); }
  h3 { font-size: 18px; font-weight: 600; color: $text-primary; margin: 0 0 8px; }
  p { font-size: 14px; color: $text-muted; margin: 0; }
}

.period-name { display: flex; align-items: center; gap: 6px; font-weight: 600; color: $text-primary; .el-icon { color: var(--primary); } }
.year-badge { font-weight: 700; color: $text-primary; font-size: 14px; }
.month-badge { color: $text-secondary; font-size: 14px; }

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 20px;

  .status-dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; }

  &.status--open    { background: #D1FAE5; color: #059669; }
  &.status--closed  { background: #F3F4F6; color: #6B7280; }
}

.action-cell { display: flex; gap: 4px; }
.action-btn {
  display: inline-flex !important;
  align-items: center !important;
  gap: 4px !important;
  padding: 4px 10px !important;
  border-radius: $radius-sm !important;
  font-size: 12px !important;
}

.refresh-btn { &:hover { color: var(--primary); border-color: rgba(var(--primary), 0.4); } }

@keyframes fadeInUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }

@media (max-width: 1024px) { .report-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 768px) {
  .book-report { padding: 16px; }
  .report-grid { grid-template-columns: 1fr; }
  .toolbar-card { flex-direction: column; align-items: stretch; gap: 12px; }
  .period-select { width: 100%; }
}
</style>
