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
import { ref, reactive, onMounted } from 'vue'
import { taxApi } from '@/api/tax'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Discount, DataAnalysis, List } from '@element-plus/icons-vue'

const activeTab = ref('deduction')

const tabs = [
  { name: 'deduction', label: '专项附加扣除', icon: Discount },
  { name: 'calculate', label: '个税计算', icon: DataAnalysis },
  { name: 'records', label: '申报记录', icon: List },
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
const calcForm = reactive({
  gross_income: 0,
  year: new Date().getFullYear(),
  month: new Date().getMonth() + 1,
  deduction_ids: [] as number[],
})

async function handleCalculate() {
  calculating.value = true
  calcResult.value = null
  try {
    calcResult.value = await taxApi.calculate({
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

onMounted(() => {
  loadDeductions()
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
