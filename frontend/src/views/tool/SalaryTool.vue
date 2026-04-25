<template>
  <div class="salary-tool">
    <!-- Page Header -->
    <div class="page-head">
      <div class="page-head-left">
        <div class="page-head-indicator"></div>
        <div>
          <h2 class="page-head-title">薪资管理</h2>
          <p class="page-head-desc">工资核算、模板配置与工资条发放</p>
        </div>
      </div>
    </div>

    <!-- Tab Navigation -->
    <div class="tab-shell">
      <div class="tab-rail">
        <div
          v-for="tab in tabs"
          :key="tab.name"
          class="tab-btn"
          :class="{ 'tab-btn--active': activeTab === tab.name }"
          @click="activeTab = tab.name"
        >
          <el-icon :size="16"><component :is="tab.icon" /></el-icon>
          <span>{{ tab.label }}</span>
        </div>
      </div>

      <div class="tab-body">
        <!-- 数据看板 -->
        <div v-show="activeTab === 'dashboard'" class="tab-panel">
          <SalaryDashboard />
        </div>

        <!-- 薪资模板 -->
        <div v-show="activeTab === 'template'" class="tab-panel">
          <div class="panel-bar">
            <span class="panel-title">薪资模板配置</span>
            <el-button type="primary" :loading="savingTemplate" @click="saveTemplate">
              保存配置
            </el-button>
          </div>
          <div v-loading="loadingTemplate">
            <el-table :data="templateItems" stripe>
              <el-table-column prop="name" label="项目名称" min-width="120" />
              <el-table-column prop="category" label="类别" width="100">
                <template #default="{ row }">{{ row.type === 'earning' ? '应发' : '扣款' }}</template>
              </el-table-column>
              <el-table-column prop="description" label="说明" min-width="160" show-overflow-tooltip />
              <el-table-column label="启用" width="80">
                <template #default="{ row }">
                  <el-switch v-model="row.is_enabled" :disabled="row.is_default" />
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>

        <!-- 工资核算 -->
        <div v-show="activeTab === 'payroll'" class="tab-panel">
          <div class="panel-bar">
            <span class="panel-title">工资核算</span>
            <el-form inline @submit.prevent>
              <el-form-item label="月份">
                <el-date-picker v-model="payrollYM" type="month" placeholder="选择月份" value-format="YYYY-MM" style="width: 140px" />
              </el-form-item>
              <el-form-item>
                <el-checkbox v-model="copyFromPrev">复制上月数据</el-checkbox>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="creating" @click="createPayroll">创建工资表</el-button>
                <el-button :loading="calculating" @click="calculatePayroll">一键核算</el-button>
                <el-button type="success" :loading="confirming" @click="confirmPayroll">确认锁定</el-button>
              </el-form-item>
            </el-form>
          </div>
          <el-table :data="payrollList" stripe v-loading="loadingPayroll">
            <el-table-column prop="year" label="年份" width="70" />
            <el-table-column prop="month" label="月份" width="70" />
            <el-table-column prop="employee_count" label="人数" width="70" />
            <el-table-column prop="total_gross" label="应发总额" width="110">
              <template #default="{ row }">¥{{ row.total_gross }}</template>
            </el-table-column>
            <el-table-column prop="total_deduction" label="扣款总额" width="110">
              <template #default="{ row }">¥{{ row.total_deduction }}</template>
            </el-table-column>
            <el-table-column prop="total_net" label="实发总额" width="110">
              <template #default="{ row }">¥{{ row.total_net }}</template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="90">
              <template #default="{ row }">
                <el-tag :type="payrollStatusTagType[row.status]" size="small">{{ payrollStatusMap[row.status] }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="viewPayrollDetail(row.id)">查看明细</el-button>
                <el-button v-if="row.status === 'confirmed'" size="small" type="success" @click="recordPayment(row)">登记发放</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination class="mt-4" layout="total,prev,pager,next" :total="payrollTotal" :page="payrollPage" :page-size="payrollPageSize" @current-change="loadPayroll" />
        </div>

        <!-- 调薪管理 -->
        <div v-show="activeTab === 'adjustment'" class="tab-panel">
          <SalaryAdjustment />
        </div>

        <!-- 导出 -->
        <div v-show="activeTab === 'export'" class="tab-panel">
          <div class="panel-bar">
            <span class="panel-title">工资表导出</span>
            <el-form inline @submit.prevent>
              <el-form-item label="选择月份">
                <el-date-picker v-model="exportYM" type="month" placeholder="选择月份" value-format="YYYY-MM" style="width: 140px" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :disabled="!exportYM" :loading="exporting" @click="handleExport">下载Excel</el-button>
              </el-form-item>
            </el-form>
          </div>
        </div>

        <!-- 个税上传 -->
        <div v-show="activeTab === 'tax-upload'" class="tab-panel">
          <TaxUpload />
        </div>

        <!-- 工资条发送 -->
        <div v-show="activeTab === 'slip-send'" class="tab-panel">
          <SalarySlipSend />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { salaryApi } from '@/api/salary'
import { ElMessage } from 'element-plus'
import { Odometer, Coin, List, Edit, Download, Upload, Message } from '@element-plus/icons-vue'
import SalaryDashboard from './SalaryDashboard.vue'
import SalaryAdjustment from './SalaryAdjustment.vue'
import TaxUpload from './TaxUpload.vue'
import SalarySlipSend from './SalarySlipSend.vue'

const activeTab = ref('dashboard')

const tabs = [
  { name: 'dashboard', label: '数据看板', icon: Odometer },
  { name: 'template', label: '薪资模板', icon: List },
  { name: 'payroll', label: '工资核算', icon: Coin },
  { name: 'adjustment', label: '调薪管理', icon: Edit },
  { name: 'export', label: '导出', icon: Download },
  { name: 'tax-upload', label: '个税上传', icon: Upload },
  { name: 'slip-send', label: '工资条发送', icon: Message },
]

// Template
const loadingTemplate = ref(false)
const savingTemplate = ref(false)
const templateItems = ref<any[]>([])

async function loadTemplate() {
  loadingTemplate.value = true
  try {
    const tpl = await salaryApi.template()
    templateItems.value = tpl.items
  } catch {
    ElMessage.error('加载模板失败')
  } finally {
    loadingTemplate.value = false
  }
}

async function saveTemplate() {
  savingTemplate.value = true
  try {
    const items = templateItems.value.map((i) => ({ id: i.id, is_enabled: i.is_enabled }))
    await salaryApi.updateTemplate({ items })
    ElMessage.success('保存成功')
  } catch {
    ElMessage.error('保存失败')
  } finally {
    savingTemplate.value = false
  }
}

// Payroll
const payrollYM = ref('')
const copyFromPrev = ref(false)
const creating = ref(false)
const calculating = ref(false)
const confirming = ref(false)
const loadingPayroll = ref(false)
const payrollList = ref<any[]>([])
const payrollTotal = ref(0)
const payrollPage = ref(1)
const payrollPageSize = ref(20)

const payrollStatusMap: Record<string, string> = {
  draft: '草稿',
  calculated: '已核算',
  confirmed: '已确认',
  paid: '已发放',
}
const payrollStatusTagType: Record<string, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
  draft: 'info',
  calculated: 'warning',
  confirmed: 'primary',
  paid: 'success',
}

async function loadPayroll(p = 1) {
  payrollPage.value = p
  if (!payrollYM.value) return
  const [year, month] = payrollYM.value.split('-').map(Number)
  loadingPayroll.value = true
  try {
    const res = await salaryApi.list({ year, month, page: p, page_size: payrollPageSize.value })
    payrollList.value = res?.list ?? []
    payrollTotal.value = res?.total ?? 0
  } catch {
    ElMessage.error('加载工资表失败')
  } finally {
    loadingPayroll.value = false
  }
}

async function createPayroll() {
  if (!payrollYM.value) {
    ElMessage.warning('请先选择月份')
    return
  }
  const [year, month] = payrollYM.value.split('-').map(Number)
  creating.value = true
  try {
    await salaryApi.createPayroll({
      year,
      month,
      copy_from_month: copyFromPrev.value ? `${year}-${String(month).padStart(2, '0')}` : undefined,
    })
    ElMessage.success('工资表已创建')
    loadPayroll()
  } catch {
    ElMessage.error('创建失败')
  } finally {
    creating.value = false
  }
}

async function calculatePayroll() {
  if (!payrollYM.value) {
    ElMessage.warning('请先选择月份')
    return
  }
  const [year, month] = payrollYM.value.split('-').map(Number)
  calculating.value = true
  try {
    await salaryApi.calculatePayroll({ year, month })
    ElMessage.success('核算完成')
    loadPayroll()
  } catch {
    ElMessage.error('核算失败')
  } finally {
    calculating.value = false
  }
}

async function confirmPayroll() {
  if (!payrollYM.value) {
    ElMessage.warning('请先选择月份')
    return
  }
  const [year, month] = payrollYM.value.split('-').map(Number)
  confirming.value = true
  try {
    await salaryApi.confirmPayroll({ year, month })
    ElMessage.success('已确认锁定')
    loadPayroll()
  } catch {
    ElMessage.error('操作失败')
  } finally {
    confirming.value = false
  }
}

function viewPayrollDetail(id: number) {
  ElMessage.info(`查看工资表明细 ID=${id}`)
}

async function recordPayment(row: any) {
  try {
    await salaryApi.recordPayment(row.id, {
      method: 'bank_transfer',
      paid_at: new Date().toISOString().slice(0, 10),
    })
    ElMessage.success('已登记发放')
    loadPayroll()
  } catch {
    ElMessage.error('操作失败')
  }
}

// Export
const exportYM = ref('')
const exporting = ref(false)

async function handleExport() {
  if (!exportYM.value) return
  const [year, month] = exportYM.value.split('-').map(Number)
  exporting.value = true
  try {
    const blob = await salaryApi.export(year, month) as unknown as Blob
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `工资表_${year}年${month}月.xlsx`
    a.click()
    URL.revokeObjectURL(url)
  } catch {
    ElMessage.error('导出失败')
  } finally {
    exporting.value = false
  }
}

onMounted(() => {
  loadTemplate()
  const now = new Date()
  payrollYM.value = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
})
</script>

<style scoped lang="scss">
$primary: #7C3AED;
$primary-light: #A78BFA;
$text-primary: #1A1D2E;
$text-secondary: #5E6278;
$text-muted: #A0A3BD;
$border: #E8EBF0;
$surface: #FFFFFF;
$surface-alt: #F8F9FC;

.salary-tool {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* ─── Page Header ─── */
.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.page-head-left {
  display: flex;
  align-items: center;
  gap: 14px;
}

.page-head-indicator {
  width: 4px;
  height: 36px;
  border-radius: 4px;
  background: linear-gradient(180deg, $primary 0%, $primary-light 100%);
}

.page-head-title {
  font-size: 20px;
  font-weight: 700;
  color: $text-primary;
  margin: 0;
  letter-spacing: -0.3px;
}

.page-head-desc {
  font-size: 13px;
  color: $text-muted;
  margin: 4px 0 0;
}

/* ─── Tab Shell ─── */
.tab-shell {
  background: $surface;
  border: 1px solid $border;
  border-radius: 20px;
  overflow: hidden;
}

.tab-rail {
  display: flex;
  gap: 2px;
  padding: 10px 12px;
  border-bottom: 1px solid $border;
  overflow-x: auto;
  background: $surface-alt;

  &::-webkit-scrollbar {
    height: 0;
  }
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  color: $text-secondary;
  cursor: pointer;
  transition: all 0.25s ease;
  white-space: nowrap;
  user-select: none;

  &:hover {
    background: rgba($primary, 0.06);
    color: $primary;
  }

  &--active {
    background: $surface;
    color: $primary;
    font-weight: 600;
    box-shadow: 0 1px 4px rgba(0,0,0,0.06);
  }
}

.tab-body {
  padding: 0;
}

.tab-panel {
  padding: 24px;
  animation: panelFade 0.3s ease;
}

@keyframes panelFade {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}

.panel-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
}

.panel-title {
  font-size: 15px;
  font-weight: 600;
  color: $text-primary;
}

.mt-4 {
  margin-top: 16px;
}
</style>
