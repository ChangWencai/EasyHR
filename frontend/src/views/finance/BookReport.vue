<template>
  <div class="book-report">
    <el-tabs v-model="activeTab" type="border-card">
      <!-- 科目余额表 -->
      <el-tab-pane label="科目余额表" name="trial">
        <div class="tab-toolbar">
          <el-select v-model="trialPeriod" placeholder="选择会计期间" style="width: 180px" @change="loadTrialBalance">
            <el-option v-for="p in periods" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
          <el-button :disabled="!trialPeriod" @click="handleExportTrial">导出Excel</el-button>
        </div>
        <el-table :data="trialData" stripe v-loading="trialLoading" border class="mt-2" show-summary>
          <el-table-column prop="account_code" label="编码" width="120" />
          <el-table-column prop="account_name" label="名称" />
          <el-table-column prop="opening_balance" label="期初余额" align="right" width="140">
            <template #default="{ row }">{{ formatAmount(row.opening_balance) }}</template>
          </el-table-column>
          <el-table-column prop="debit_amount" label="借方发生额" align="right" width="140">
            <template #default="{ row }">{{ formatAmount(row.debit_amount) }}</template>
          </el-table-column>
          <el-table-column prop="credit_amount" label="贷方发生额" align="right" width="140">
            <template #default="{ row }">{{ formatAmount(row.credit_amount) }}</template>
          </el-table-column>
          <el-table-column prop="closing_balance" label="期末余额" align="right" width="140">
            <template #default="{ row }">{{ formatAmount(row.closing_balance) }}</template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 报表 -->
      <el-tab-pane label="报表" name="report">
        <div class="tab-toolbar">
          <el-select v-model="reportPeriod" placeholder="选择会计期间" style="width: 180px">
            <el-option v-for="p in periods" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
          <el-button :disabled="!reportPeriod" @click="loadBalanceSheet">资产负债表</el-button>
          <el-button :disabled="!reportPeriod" @click="loadIncomeStatement">利润表</el-button>
        </div>

        <!-- 资产负债表 -->
        <div v-if="balanceSheetData" class="report-block mt-2">
          <h3 class="report-title">资产负债表</h3>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="资产总计">{{ formatAmount(balanceSheetData.total_assets) }}</el-descriptions-item>
            <el-descriptions-item label="负债合计">{{ formatAmount(balanceSheetData.total_liabilities) }}</el-descriptions-item>
            <el-descriptions-item label="所有者权益">{{ formatAmount(balanceSheetData.total_equity) }}</el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- 利润表 -->
        <div v-if="incomeStatementData" class="report-block mt-2">
          <h3 class="report-title">利润表</h3>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="营业收入">{{ formatAmount(incomeStatementData.revenue) }}</el-descriptions-item>
            <el-descriptions-item label="营业成本">{{ formatAmount(incomeStatementData.cost) }}</el-descriptions-item>
            <el-descriptions-item label="净利润">{{ formatAmount(incomeStatementData.net_profit) }}</el-descriptions-item>
          </el-descriptions>
        </div>
      </el-tab-pane>

      <!-- 期间管理 -->
      <el-tab-pane label="期间管理" name="period">
        <div class="toolbar">
          <el-button @click="loadPeriods">刷新</el-button>
        </div>
        <el-table :data="periodList" stripe v-loading="periodLoading" class="mt-2">
          <el-table-column prop="name" label="期间名称" width="160" />
          <el-table-column prop="year" label="年度" width="100" />
          <el-table-column prop="month" label="月份" width="80" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="periodStatusType(row.status)" size="small">
                {{ periodStatusLabel(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button
                v-if="row.status === 'open'"
                link
                type="warning"
                size="small"
                @click="handleClosePeriod(row)"
              >结账</el-button>
              <el-button
                v-if="row.status === 'closed'"
                link
                type="primary"
                size="small"
                @click="handleRevertPeriod(row)"
              >反结账</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { financeApi } from '@/api/finance'

interface Period {
  id: number
  name: string
  year: number
  month: number
  status: string
}

interface TrialRow {
  account_code: string
  account_name: string
  opening_balance: string
  debit_amount: string
  credit_amount: string
  closing_balance: string
}

const activeTab = ref('trial')

// Trial balance
const periods = ref<Period[]>([])
const trialPeriod = ref<number | null>(null)
const trialData = ref<TrialRow[]>([])
const trialLoading = ref(false)

// Reports
const reportPeriod = ref<number | null>(null)
const balanceSheetData = ref<any | null>(null)
const incomeStatementData = ref<any | null>(null)

// Periods
const periodList = ref<Period[]>([])
const periodLoading = ref(false)

function formatAmount(val?: string | number) {
  if (val === undefined || val === null) return '0.00'
  const num = typeof val === 'string' ? parseFloat(val) : val
  if (isNaN(num)) return '0.00'
  return num.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function periodStatusType(status: string) {
  const map: Record<string, string> = { open: 'success', closed: 'info' }
  return map[status] || 'info'
}

function periodStatusLabel(status: string) {
  const map: Record<string, string> = { open: '开放', closed: '已结账' }
  return map[status] || status
}

// Trial balance
async function loadPeriods() {
  try {
    const res = await financeApi.periods() as any
    periods.value = res.data?.data || res.data || []
  } catch {
    periods.value = []
  }
}

async function loadTrialBalance() {
  if (!trialPeriod.value) return
  trialLoading.value = true
  try {
    const res = await financeApi.trialBalance(trialPeriod.value) as any
    trialData.value = res.data?.data || res.data?.rows || []
  } catch {
    trialData.value = []
  } finally {
    trialLoading.value = false
  }
}

async function handleExportTrial() {
  if (!trialPeriod.value) return
  try {
    const res = await financeApi.bookExport(trialPeriod.value) as any
    const blob = new Blob([res], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `科目余额表_${trialPeriod.value}.xlsx`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (e: unknown) {
    ElMessage.error('导出失败')
  }
}

// Reports
async function loadBalanceSheet() {
  if (!reportPeriod.value) return
  try {
    const res = await financeApi.balanceSheet(reportPeriod.value) as any
    balanceSheetData.value = res.data?.data || res.data || null
    ElMessage.success('资产负债表加载成功')
  } catch {
    ElMessage.error('加载失败')
  }
}

async function loadIncomeStatement() {
  if (!reportPeriod.value) return
  try {
    const res = await financeApi.incomeStatement(reportPeriod.value) as any
    incomeStatementData.value = res.data?.data || res.data || null
    ElMessage.success('利润表加载成功')
  } catch {
    ElMessage.error('加载失败')
  }
}

// Period management
async function loadPeriodList() {
  periodLoading.value = true
  try {
    const res = await financeApi.periods() as any
    periodList.value = res.data?.data || res.data || []
  } catch {
    periodList.value = []
  } finally {
    periodLoading.value = false
  }
}

async function handleClosePeriod(row: Period) {
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
  }
}

async function handleRevertPeriod(row: Period) {
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
  }
}

onMounted(() => {
  loadPeriods()
  loadPeriodList()
})
</script>

<style scoped lang="scss">
.book-report {
  padding: 8px;
  .tab-toolbar {
    display: flex;
    gap: 8px;
    align-items: center;
    flex-wrap: wrap;
  }
  .toolbar {
    display: flex;
    gap: 8px;
  }
  .mt-2 {
    margin-top: 12px;
  }
  .report-block {
    background: #fafafa;
    padding: 16px;
    border-radius: 4px;
    .report-title {
      margin: 0 0 12px;
      font-size: 16px;
      font-weight: 600;
    }
  }
}
</style>
