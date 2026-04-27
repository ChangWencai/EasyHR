<template>
  <div class="tax-tool">
    <!-- Page Header -->
    <div class="page-head">
      <div class="page-head-left">
        <div class="page-head-indicator"></div>
        <div>
          <h2 class="page-head-title">个税申报</h2>
          <p class="page-head-desc">专项附加扣除、个税计算与申报记录</p>
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
        <!-- 专项附加扣除 -->
        <div v-show="activeTab === 'deduction'" class="tab-panel">
          <div class="panel-bar">
            <span class="panel-title">专项附加扣除</span>
            <el-button type="primary" @click="showDeductionDialog = true">新增扣除项</el-button>
          </div>
          <el-table :data="deductions" stripe v-loading="loadingDeductions">
            <el-table-column prop="name" label="扣除名称" min-width="120" />
            <el-table-column prop="type" label="类型" min-width="120">
              <template #default="{ row }">{{ deductionTypeMap[row.type] }}</template>
            </el-table-column>
            <el-table-column label="扣除金额" min-width="100"><template #default="{ row }">¥{{ row.amount }}</template></el-table-column>
            <el-table-column label="年度上限" min-width="100"><template #default="{ row }">¥{{ row.max_amount }}</template></el-table-column>
            <el-table-column prop="year" label="年度" width="80" />
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{ row }">
                <el-button size="small" type="danger" plain @click="handleDeleteDeduction(row.id)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 个税计算 -->
        <div v-show="activeTab === 'calculate'" class="tab-panel">
          <div class="panel-bar">
            <span class="panel-title">个税计算</span>
          </div>
          <el-form :model="calcForm" label-width="100px" @submit.prevent="handleCalculate">
            <el-form-item label="员工">
              <el-select v-model="calcForm.employee_id" placeholder="选择员工" style="width: 200px" clearable>
                <el-option v-for="e in employeeOptions" :key="e.id" :label="e.name" :value="e.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="税前工资">
              <el-input-number v-model="calcForm.gross_income" :min="0" :precision="2" style="width: 200px" />
            </el-form-item>
            <el-form-item label="年份">
              <el-input-number v-model="calcForm.year" :min="2020" :max="2030" style="width: 120px" />
            </el-form-item>
            <el-form-item label="月份">
              <el-input-number v-model="calcForm.month" :min="1" :max="12" style="width: 120px" />
            </el-form-item>
            <el-form-item label="扣除项">
              <el-select v-model="calcForm.deduction_ids" multiple placeholder="选择专项附加扣除（可选）" style="width: 300px" clearable>
                <el-option v-for="d in deductions" :key="d.id" :label="`${d.name} ¥${d.amount}`" :value="d.id" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="calculating" @click="handleCalculate">计算</el-button>
            </el-form-item>
          </el-form>

          <div v-if="calcResult" class="calc-result">
            <div class="calc-result-header">
              <span class="calc-result-dot"></span>
              计算结果
            </div>
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="税前工资">¥{{ calcResult.gross_income }}</el-descriptions-item>
              <el-descriptions-item label="免税收入">¥{{ calcResult.tax_free_income }}</el-descriptions-item>
              <el-descriptions-item label="税前扣除合计">¥{{ calcResult.deduction_total }}</el-descriptions-item>
              <el-descriptions-item label="应纳税所得额">¥{{ calcResult.taxable_income }}</el-descriptions-item>
              <el-descriptions-item label="适用税率">{{ (calcResult.applicable_bracket.rate * 100).toFixed(1) }}%</el-descriptions-item>
              <el-descriptions-item label="速算扣除数">¥{{ calcResult.quick_deduction }}</el-descriptions-item>
              <el-descriptions-item label="应缴税额" label-class-name="tax-label">
                <span class="tax-amount">¥{{ calcResult.tax_amount }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="税后工资">¥{{ calcResult.net_income }}</el-descriptions-item>
            </el-descriptions>
          </div>
        </div>

        <!-- 申报记录 -->
        <div v-show="activeTab === 'records'" class="tab-panel">
          <div class="panel-bar">
            <span class="panel-title">个税申报记录</span>
            <el-form inline>
              <el-form-item label="年份">
                <el-input-number v-model="recordYear" :min="2020" :max="2030" style="width: 100px" @change="loadRecords" />
              </el-form-item>
            </el-form>
          </div>
          <el-table :data="taxRecords" stripe v-loading="loadingRecords">
            <el-table-column prop="employee_name" label="员工" min-width="80" />
            <el-table-column prop="year" label="年份" width="70" />
            <el-table-column prop="month" label="月份" width="70" />
            <el-table-column prop="gross_income" label="税前工资" min-width="100"><template #default="{ row }">¥{{ row.gross_income }}</template></el-table-column>
            <el-table-column prop="tax_amount" label="个税" min-width="90"><template #default="{ row }">¥{{ row.tax_amount }}</template></el-table-column>
            <el-table-column prop="status" label="状态" min-width="80">
              <template #default="{ row }"><el-tag :type="taxStatusTagType[row.status]" size="small">{{ taxStatusMap[row.status] }}</el-tag></template>
            </el-table-column>
            <el-table-column prop="declared_at" label="申报时间" min-width="160"><template #default="{ row }">{{ row.declared_at || '—' }}</template></el-table-column>
          </el-table>
          <el-pagination class="mt-4" layout="total,prev,pager,next" :total="recordTotal" :page="recordPage" :page-size="recordPageSize" @current-change="loadRecords" />
        </div>

        <!-- 申报管理 -->
        <div v-show="activeTab === 'declarations'" class="tab-panel">
          <div class="panel-bar">
            <span class="panel-title">税务申报管理</span>
            <el-form inline>
              <el-form-item label="年份">
                <el-input-number v-model="declYear" :min="2020" :max="2030" style="width: 100px" @change="loadDeclarations" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="loadDeclarations">查询</el-button>
                <el-button :disabled="!currentDecl" @click="handleExportDecl">
                  <el-icon><Download /></el-icon> 导出Excel
                </el-button>
              </el-form-item>
            </el-form>
          </div>

          <!-- 当月申报概览 -->
          <div v-if="currentDecl" class="decl-overview">
            <el-alert
              :title="declStatusLabel"
              :type="declStatusAlertType"
              :closable="false"
              show-icon
              style="margin-bottom: 12px"
            >
              <template #default>
                {{ declYear }}年{{ currentDecl.month }}月 &nbsp;|&nbsp;
                员工 {{ currentDecl.total_employees }} 人 &nbsp;|&nbsp;
                收入合计 ¥{{ currentDecl.total_income?.toLocaleString() }} &nbsp;|&nbsp;
                个税合计 ¥{{ currentDecl.total_tax?.toLocaleString() }}
              </template>
            </el-alert>
            <el-button
              v-if="currentDecl.status === 'pending'"
              type="danger"
              :loading="markingDeclared"
              @click="handleMarkDeclared"
            >
              标记为已申报
            </el-button>
            <el-tag v-else type="success">已申报 · {{ currentDecl.declared_at }}</el-tag>
          </div>

          <el-table :data="declarations" stripe v-loading="loadingDeclarations" class="mt-4">
            <el-table-column prop="year" label="年份" width="70" />
            <el-table-column prop="month" label="月份" width="70" />
            <el-table-column prop="total_employees" label="员工数" width="80" />
            <el-table-column prop="total_income" label="收入合计" min-width="120">
              <template #default="{ row }">¥{{ row.total_income?.toLocaleString() }}</template>
            </el-table-column>
            <el-table-column prop="total_tax" label="个税合计" min-width="120">
              <template #default="{ row }">¥{{ row.total_tax?.toLocaleString() }}</template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="90">
              <template #default="{ row }">
                <el-tag :type="declStatusTagType[row.status]" size="small">{{ declStatusMap[row.status] }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="declared_at" label="申报时间" min-width="160">
              <template #default="{ row }">{{ row.declared_at || '—' }}</template>
            </el-table-column>
          </el-table>
          <el-pagination class="mt-4" layout="total,prev,pager,next" :total="declTotal" :page="declPage" :page-size="declPageSize" @current-change="loadDeclarations" />
        </div>
      </div>
    </div>

    <el-dialog v-model="showDeductionDialog" title="新增扣除项" width="420px">
      <el-form ref="deductionFormRef" :model="deductionForm" :rules="deductionRules" label-width="100px">
        <el-form-item label="扣除类型" prop="type">
          <el-select v-model="deductionForm.type" placeholder="选择类型" style="width: 100%">
            <el-option v-for="(label, key) in deductionTypeMap" :key="key" :label="label" :value="key" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input v-model="deductionForm.name" placeholder="如：住房贷款利息" />
        </el-form-item>
        <el-form-item label="扣除金额" prop="amount">
          <el-input-number v-model="deductionForm.amount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="年度上限" prop="max_amount">
          <el-input-number v-model="deductionForm.max_amount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="年度" prop="year">
          <el-input-number v-model="deductionForm.year" :min="2020" :max="2030" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDeductionDialog = false">取消</el-button>
        <el-button type="primary" :loading="savingDeduction" @click="handleSaveDeduction">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { taxApi, type TaxDeclaration } from '@/api/tax'
import { employeeApi } from '@/api/employee'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Discount, DataAnalysis, List, Download } from '@element-plus/icons-vue'

const activeTab = ref('deduction')

const tabs = [
  { name: 'deduction', label: '专项附加扣除', icon: Discount },
  { name: 'calculate', label: '个税计算', icon: DataAnalysis },
  { name: 'records', label: '申报记录', icon: List },
  { name: 'declarations', label: '申报管理', icon: Download },
]

const deductionTypeMap: Record<string, string> = {
  housing_loan: '住房贷款利息',
  housing_rent: '住房租金',
  elderly_care: '赡养老人',
  children_education: '子女教育',
  continuing_education: '继续教育',
  serious_illness: '大病医疗',
  other: '其他',
}

const taxStatusMap: Record<string, string> = {
  pending: '待申报',
  declared: '已申报',
  paid: '已缴纳',
}
const taxStatusTagType: Record<string, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
  pending: 'warning',
  declared: 'primary',
  paid: 'success',
}

// Deductions
const loadingDeductions = ref(false)
const deductions = ref<any[]>([])
const showDeductionDialog = ref(false)
const savingDeduction = ref(false)
const deductionFormRef = ref<FormInstance>()
const deductionForm = reactive({
  type: '',
  name: '',
  amount: 0,
  max_amount: 0,
  year: new Date().getFullYear(),
})
const deductionRules: FormRules = {
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  amount: [{ required: true, message: '请输入扣除金额', trigger: 'blur' }],
  year: [{ required: true, message: '请输入年度', trigger: 'blur' }],
}

async function loadDeductions() {
  loadingDeductions.value = true
  try {
    deductions.value = (await taxApi.deductions({ year: deductionForm.year })) ?? []
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loadingDeductions.value = false
  }
}

async function handleSaveDeduction() {
  if (savingDeduction.value) return
  if (!deductionFormRef.value) return
  try {
    await deductionFormRef.value.validate()
  } catch {
    return
  }
  savingDeduction.value = true
  try {
    await taxApi.createDeduction({ ...deductionForm })
    ElMessage.success('保存成功')
    showDeductionDialog.value = false
    loadDeductions()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    savingDeduction.value = false
  }
}

async function handleDeleteDeduction(id: number) {
  const deleting = ref(false)
  deleting.value = true
  try {
    await taxApi.deleteDeduction(id)
    ElMessage.success('已删除')
    loadDeductions()
  } catch {
    ElMessage.error('删除失败')
  } finally {
    deleting.value = false
  }
}

// Calculate
const calculating = ref(false)
const calcResult = ref<any>(null)
const employeeOptions = ref<any[]>([])
const calcForm = reactive({
  employee_id: undefined as number | undefined,
  gross_income: 0,
  year: new Date().getFullYear(),
  month: new Date().getMonth() + 1,
  deduction_ids: [] as number[],
})

async function loadEmployees() {
  try {
    const res = await employeeApi.list({ page: 1, page_size: 100 }) as { list: any[] }
    employeeOptions.value = res.list || []
  } catch { /* ignore */ }
}

async function handleCalculate() {
  if (!calcForm.employee_id) {
    ElMessage.warning('请选择员工')
    return
  }
  calculating.value = true
  calcResult.value = null
  try {
    calcResult.value = await taxApi.calculate({
      employee_id: calcForm.employee_id,
      gross_income: calcForm.gross_income,
      year: calcForm.year,
      month: calcForm.month,
      deduction_ids: calcForm.deduction_ids.length > 0 ? calcForm.deduction_ids : undefined,
    })
  } catch {
    ElMessage.error('计算失败')
  } finally {
    calculating.value = false
  }
}

// Records
const loadingRecords = ref(false)
const taxRecords = ref<any[]>([])
const recordTotal = ref(0)
const recordPage = ref(1)
const recordPageSize = ref(20)
const recordYear = ref(new Date().getFullYear())

async function loadRecords(p = 1) {
  recordPage.value = p
  loadingRecords.value = true
  try {
    const res = await taxApi.records({ page: p, year: recordYear.value })
    taxRecords.value = res.list
    recordTotal.value = res.total
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loadingRecords.value = false
  }
}

// Declarations
const declYear = ref(new Date().getFullYear())
const loadingDeclarations = ref(false)
const declarations = ref<TaxDeclaration[]>([])
const currentDecl = ref<TaxDeclaration | null>(null)
const declTotal = ref(0)
const declPage = ref(1)
const declPageSize = ref(20)
const markingDeclared = ref(false)

const declStatusMap: Record<string, string> = {
  pending: '待申报',
  declared: '已申报',
  paid: '已缴纳',
}
const declStatusTagType: Record<string, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
  pending: 'warning',
  declared: 'success',
  paid: 'info',
}

const declStatusLabel = computed(() => {
  if (!currentDecl.value) return ''
  return declStatusMap[currentDecl.value.status] || currentDecl.value.status
})
const declStatusAlertType = computed((): 'warning' | 'success' | 'info' => {
  if (!currentDecl.value) return 'info'
  return currentDecl.value.status === 'pending' ? 'warning' : 'success'
})

async function loadDeclarations(p = 1) {
  declPage.value = p
  loadingDeclarations.value = true
  try {
    const [res, cur] = await Promise.all([
      taxApi.declarations({ year: declYear.value, page: p, page_size: declPageSize.value }),
      taxApi.getCurrentDeclaration().catch(() => null),
    ])
    declarations.value = res.list || []
    declTotal.value = res.total || 0
    currentDecl.value = cur
  } catch {
    ElMessage.error('加载申报列表失败')
  } finally {
    loadingDeclarations.value = false
  }
}

async function handleMarkDeclared() {
  if (!currentDecl.value) return
  markingDeclared.value = true
  try {
    await taxApi.markDeclared(currentDecl.value.id)
    ElMessage.success('已标记为申报')
    await loadDeclarations()
  } catch {
    ElMessage.error('操作失败')
  } finally {
    markingDeclared.value = false
  }
}

async function handleExportDecl() {
  if (!currentDecl.value) return
  const y = currentDecl.value.year
  const m = currentDecl.value.month
  try {
    const blob = await taxApi.exportDeclarationExcel(y, m)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `个税申报表_${y}年${m}月.xlsx`
    a.click()
    URL.revokeObjectURL(url)
  } catch {
    ElMessage.error('导出失败')
  }
}

onMounted(() => {
  loadDeductions()
  loadEmployees()
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

.tax-tool {
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
  background: linear-gradient(180deg, #D97706 0%, #FBBF24 100%);
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

/* ─── Calc Result ─── */
.calc-result {
  margin-top: 24px;
  background: $surface-alt;
  border: 1px solid $border;
  border-radius: 16px;
  padding: 20px;
}

.calc-result-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: $text-primary;
  margin-bottom: 16px;
}

.calc-result-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: linear-gradient(135deg, #D97706, #FBBF24);
}

:deep(.tax-label) {
  color: #EF4444;
  font-weight: 600;
}

.tax-amount {
  color: #EF4444;
  font-size: 16px;
  font-weight: 700;
}

.mt-4 {
  margin-top: 16px;
}
</style>
