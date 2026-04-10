<template>
  <div class="tax-tool">
    <el-tabs v-model="activeTab">
      <!-- Tab 1: 专项附加扣除 -->
      <el-tab-pane label="专项附加扣除" name="deduction">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>专项附加扣除</span>
              <el-button type="primary" size="small" @click="showDeductionDialog = true">新增扣除项</el-button>
            </div>
          </template>

          <el-table :data="deductions" stripe v-loading="loadingDeductions">
            <el-table-column prop="name" label="扣除名称" min-width="120" />
            <el-table-column prop="type" label="类型" min-width="120">
              <template #default="{ row }">{{ deductionTypeMap[row.type] }}</template>
            </el-table-column>
            <el-table-column label="扣除金额" min-width="100">
              <template #default="{ row }">¥{{ row.amount }}</template>
            </el-table-column>
            <el-table-column label="年度上限" min-width="100">
              <template #default="{ row }">¥{{ row.max_amount }}</template>
            </el-table-column>
            <el-table-column prop="year" label="年度" width="80" />
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{ row }">
                <el-button size="small" type="danger" @click="handleDeleteDeduction(row.id)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- Tab 2: 个税计算 -->
      <el-tab-pane label="个税计算" name="calculate">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>个税计算</span>
            </div>
          </template>

          <el-form :model="calcForm" label-width="100px" @submit.prevent="handleCalculate">
            <el-form-item label="税前工资">
              <el-input-number
                v-model="calcForm.gross_income"
                :min="0"
                :precision="2"
                style="width: 200px"
              />
            </el-form-item>
            <el-form-item label="年份">
              <el-input-number v-model="calcForm.year" :min="2020" :max="2030" style="width: 120px" />
            </el-form-item>
            <el-form-item label="月份">
              <el-input-number v-model="calcForm.month" :min="1" :max="12" style="width: 120px" />
            </el-form-item>
            <el-form-item label="扣除项">
              <el-select
                v-model="calcForm.deduction_ids"
                multiple
                placeholder="选择专项附加扣除（可选）"
                style="width: 300px"
                clearable
              >
                <el-option
                  v-for="d in deductions"
                  :key="d.id"
                  :label="`${d.name} ¥${d.amount}`"
                  :value="d.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="calculating" @click="handleCalculate">计算</el-button>
            </el-form-item>
          </el-form>

          <!-- 计算结果 -->
          <div v-if="calcResult" class="calc-result">
            <el-divider>计算结果</el-divider>
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="税前工资">¥{{ calcResult.gross_income }}</el-descriptions-item>
              <el-descriptions-item label="免税收入">¥{{ calcResult.tax_free_income }}</el-descriptions-item>
              <el-descriptions-item label="税前扣除合计">¥{{ calcResult.deduction_total }}</el-descriptions-item>
              <el-descriptions-item label="应纳税所得额">¥{{ calcResult.taxable_income }}</el-descriptions-item>
              <el-descriptions-item label="适用税率">
                {{ (calcResult.applicable_bracket.rate * 100).toFixed(1) }}%
              </el-descriptions-item>
              <el-descriptions-item label="速算扣除数">¥{{ calcResult.quick_deduction }}</el-descriptions-item>
              <el-descriptions-item label="应缴税额" label-class-name="tax-label">
                <span class="tax-amount">¥{{ calcResult.tax_amount }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="税后工资">¥{{ calcResult.net_income }}</el-descriptions-item>
            </el-descriptions>
          </div>
        </el-card>
      </el-tab-pane>

      <!-- Tab 3: 申报记录 -->
      <el-tab-pane label="申报记录" name="records">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>个税申报记录</span>
              <el-form inline>
                <el-form-item label="年份">
                  <el-input-number
                    v-model="recordYear"
                    :min="2020"
                    :max="2030"
                    style="width: 100px"
                    @change="loadRecords"
                  />
                </el-form-item>
              </el-form>
            </div>
          </template>

          <el-table :data="taxRecords" stripe v-loading="loadingRecords">
            <el-table-column prop="employee_name" label="员工" min-width="80" />
            <el-table-column prop="year" label="年份" width="70" />
            <el-table-column prop="month" label="月份" width="70" />
            <el-table-column prop="gross_income" label="税前工资" min-width="100">
              <template #default="{ row }">¥{{ row.gross_income }}</template>
            </el-table-column>
            <el-table-column prop="tax_amount" label="个税" min-width="90">
              <template #default="{ row }">¥{{ row.tax_amount }}</template>
            </el-table-column>
            <el-table-column prop="status" label="状态" min-width="80">
              <template #default="{ row }">
                <el-tag :type="taxStatusTagType[row.status]" size="small">
                  {{ taxStatusMap[row.status] }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="declared_at" label="申报时间" min-width="160">
              <template #default="{ row }">{{ row.declared_at || '—' }}</template>
            </el-table-column>
          </el-table>

          <el-pagination
            class="mt-4"
            layout="total,prev,pager,next"
            :total="recordTotal"
            :page="recordPage"
            :page-size="recordPageSize"
            @current-change="loadRecords"
          />
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 新增扣除项对话框 -->
    <el-dialog v-model="showDeductionDialog" title="新增扣除项" width="420px">
      <el-form ref="deductionFormRef" :model="deductionForm" :rules="deductionRules" label-width="100px">
        <el-form-item label="扣除类型" prop="type">
          <el-select v-model="deductionForm.type" placeholder="选择类型" style="width: 100%">
            <el-option
              v-for="(label, key) in deductionTypeMap"
              :key="key"
              :label="label"
              :value="key"
            />
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
import { ElMessage, FormInstance, FormRules } from 'element-plus'

const activeTab = ref('deduction')

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
const taxStatusTagType: Record<string, string> = {
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
    deductions.value = await taxApi.deductions({ year: deductionForm.year })
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loadingDeductions.value = false
  }
}

async function handleSaveDeduction() {
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
  try {
    await taxApi.deleteDeduction(id)
    ElMessage.success('已删除')
    loadDeductions()
  } catch {
    ElMessage.error('删除失败')
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
.tax-tool {
  padding: 8px;
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.calc-result {
  margin-top: 16px;
}
:deep(.tax-label) {
  color: #f56c6c;
  font-weight: 600;
}
.tax-amount {
  color: #f56c6c;
  font-size: 16px;
  font-weight: 700;
}
.mt-4 {
  margin-top: 16px;
}
</style>
