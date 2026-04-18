<template>
  <div class="salary-adjustment">
    <el-card>
      <template #header>
        <div class="page-header">
          <h1 class="page-title">调薪管理</h1>
        </div>
      </template>

      <el-tabs v-model="activeSubTab" class="adjustment-tabs">
        <!-- 员工调薪 Tab -->
        <el-tab-pane label="员工调薪" name="employee">
          <el-form :model="empForm" label-width="110px" class="adjustment-form">
            <el-form-item label="选择员工" required>
              <el-select
                v-model="empForm.employee_id"
                filterable
                remote
                placeholder="搜索员工姓名"
                :remote-method="searchEmployees"
                :loading="searchLoading"
                style="width: 280px"
                @focus="loadEmployees('')"
              >
                <el-option
                  v-for="emp in employeeOptions"
                  :key="emp.id"
                  :label="emp.name"
                  :value="emp.id"
                >
                  <span>{{ emp.name }}</span>
                  <span style="float:right;color:#8c8c8c;font-size:12px">{{ emp.position }}</span>
                </el-option>
              </el-select>
            </el-form-item>

            <el-form-item label="调整类型" required>
              <el-select v-model="empForm.adjustment_type" style="width: 200px">
                <el-option label="岗位工资" value="base_salary" />
                <el-option label="补贴" value="allowance" />
                <el-option label="奖金" value="bonus" />
                <el-option label="年终奖" value="year_end_bonus" />
                <el-option label="其他扣除" value="other" />
              </el-select>
            </el-form-item>

            <el-form-item label="调整方式" required>
              <el-radio-group v-model="empForm.adjust_by">
                <el-radio value="amount">金额（元）</el-radio>
                <el-radio value="ratio">比例（%）</el-radio>
              </el-radio-group>
            </el-form-item>

            <el-form-item :label="empForm.adjust_by === 'amount' ? '调整金额（元）' : '调整比例（%）'" required>
              <el-input-number
                v-model="empForm.adjust_value"
                :min="0"
                :precision="empForm.adjust_by === 'ratio' ? 1 : 2"
                :step="empForm.adjust_by === 'ratio' ? 0.5 : 100"
                style="width: 200px"
              />
            </el-form-item>

            <el-form-item label="生效月份" required>
              <el-date-picker
                v-model="empForm.effective_month"
                type="month"
                placeholder="选择月份"
                value-format="YYYY-MM"
                style="width: 200px"
              />
            </el-form-item>

            <el-form-item>
              <el-button :loading="previewLoading" @click="previewEmpAdjustment">预览影响</el-button>
              <el-button type="primary" :loading="submitLoading" @click="submitEmpAdjustment">提交调薪</el-button>
            </el-form-item>
          </el-form>

          <!-- 员工预览面板 -->
          <div v-if="empPreview" class="preview-panel">
            <el-card>
              <template #header>调薪预览</template>
              <el-descriptions :column="2" border>
                <el-descriptions-item label="影响员工">{{ empPreview.employee_count }} 人</el-descriptions-item>
                <el-descriptions-item label="生效月份">{{ empPreview.effective_month }}</el-descriptions-item>
                <el-descriptions-item label="月度影响">
                  <span class="impact-value">{{ formatAmount(empPreview.monthly_impact) }}</span>
                </el-descriptions-item>
                <el-descriptions-item label="年度影响">
                  <span class="impact-value">{{ formatAmount(empPreview.annual_impact) }}</span>
                </el-descriptions-item>
              </el-descriptions>
            </el-card>
          </div>
        </el-tab-pane>

        <!-- 部门普调 Tab -->
        <el-tab-pane label="部门普调" name="department">
          <el-form :model="deptForm" label-width="110px" class="adjustment-form">
            <el-form-item label="选择部门" required>
              <el-select
                v-model="deptForm.department_ids"
                multiple
                filterable
                collapse-tags
                collapse-tags-tooltip
                placeholder="选择部门"
                style="width: 400px"
                @focus="loadDepartments"
              >
                <el-option label="全选" value="__all__" @click="toggleSelectAllDepts" />
                <el-option
                  v-for="dept in departmentOptions"
                  :key="dept.id"
                  :label="dept.name"
                  :value="dept.id"
                />
              </el-select>
            </el-form-item>

            <el-form-item label="调整类型" required>
              <el-select v-model="deptForm.adjustment_type" style="width: 200px">
                <el-option label="岗位工资" value="base_salary" />
                <el-option label="补贴" value="allowance" />
                <el-option label="奖金" value="bonus" />
                <el-option label="年终奖" value="year_end_bonus" />
                <el-option label="其他扣除" value="other" />
              </el-select>
            </el-form-item>

            <el-form-item label="调整方式" required>
              <el-radio-group v-model="deptForm.adjust_by">
                <el-radio value="amount">金额（元）</el-radio>
                <el-radio value="ratio">比例（%）</el-radio>
              </el-radio-group>
            </el-form-item>

            <el-form-item :label="deptForm.adjust_by === 'amount' ? '调整金额（元）' : '调整比例（%）'" required>
              <el-input-number
                v-model="deptForm.adjust_value"
                :min="0"
                :precision="deptForm.adjust_by === 'ratio' ? 1 : 2"
                :step="deptForm.adjust_by === 'ratio' ? 0.5 : 100"
                style="width: 200px"
              />
            </el-form-item>

            <el-form-item label="生效月份" required>
              <el-date-picker
                v-model="deptForm.effective_month"
                type="month"
                placeholder="选择月份"
                value-format="YYYY-MM"
                style="width: 200px"
              />
            </el-form-item>

            <el-form-item>
              <el-button :loading="previewLoading" @click="previewDeptAdjustment">预览影响</el-button>
              <el-button type="primary" :loading="submitLoading" @click="submitDeptAdjustment">提交调薪</el-button>
            </el-form-item>
          </el-form>

          <!-- 部门预览面板 -->
          <div v-if="deptPreview" class="preview-panel">
            <el-card>
              <template #header>调薪预览</template>
              <el-descriptions :column="2" border>
                <el-descriptions-item label="影响部门">{{ deptPreview.department_count }} 个</el-descriptions-item>
                <el-descriptions-item label="影响员工">{{ deptPreview.employee_count }} 人</el-descriptions-item>
                <el-descriptions-item label="生效月份">{{ deptPreview.effective_month }}</el-descriptions-item>
                <el-descriptions-item label="月度影响">
                  <span class="impact-value">{{ formatAmount(deptPreview.monthly_impact) }}</span>
                </el-descriptions-item>
                <el-descriptions-item label="年度影响">
                  <span class="impact-value">{{ formatAmount(deptPreview.annual_impact) }}</span>
                </el-descriptions-item>
              </el-descriptions>
            </el-card>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { salaryApi } from '@/api/salary'
import { employeeApi } from '@/api/employee'
import { departmentApi } from '@/api/department'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'

const activeSubTab = ref('employee')
const searchLoading = ref(false)
const previewLoading = ref(false)
const submitLoading = ref(false)

interface Employee {
  id: number
  name: string
  position: string
}

interface Department {
  id: number
  name: string
}

interface EmpPreview {
  employee_count: number
  monthly_impact: number
  annual_impact: number
  effective_month: string
}

interface DeptPreview {
  department_count: number
  employee_count: number
  monthly_impact: number
  annual_impact: number
  effective_month: string
}

const employeeOptions = ref<Employee[]>([])
const departmentOptions = ref<Department[]>([])
const empPreview = ref<EmpPreview | null>(null)
const deptPreview = ref<DeptPreview | null>(null)

// 员工调薪表单
const empForm = reactive({
  employee_id: null as number | null,
  adjustment_type: 'base_salary',
  adjust_by: 'amount' as 'amount' | 'ratio',
  adjust_value: 0,
  effective_month: dayjs().add(1, 'month').format('YYYY-MM'),
})

// 部门普调表单
const deptForm = reactive({
  department_ids: [] as number[],
  adjustment_type: 'base_salary',
  adjust_by: 'amount' as 'amount' | 'ratio',
  adjust_value: 0,
  effective_month: dayjs().add(1, 'month').format('YYYY-MM'),
})

async function loadEmployees(query: string) {
  searchLoading.value = true
  try {
    const res = await employeeApi.list({ page: 1, page_size: 20, search: query })
    employeeOptions.value = res.list.map((e) => ({ id: e.id, name: e.name, position: e.position }))
  } catch {
    // ignore
  } finally {
    searchLoading.value = false
  }
}

async function searchEmployees(query: string) {
  if (query) {
    await loadEmployees(query)
  }
}

async function loadDepartments() {
  if (departmentOptions.value.length > 0) return
  try {
    const res = await departmentApi.list()
    departmentOptions.value = (res as any[]).map((d) => ({ id: d.id, name: d.name }))
  } catch {
    // ignore
  }
}

function toggleSelectAllDepts() {
  const allSelected = deptForm.department_ids.includes('__all__')
  if (allSelected) {
    deptForm.department_ids = departmentOptions.value.map((d) => d.id)
  } else {
    deptForm.department_ids = []
  }
}

async function previewEmpAdjustment() {
  if (!empForm.employee_id) {
    ElMessage.warning('请选择员工')
    return
  }
  if (!empForm.effective_month) {
    ElMessage.warning('请选择生效月份')
    return
  }
  previewLoading.value = true
  try {
    const res = await salaryApi.previewAdjustment({
      employee_id: empForm.employee_id,
      effective_month: empForm.effective_month,
      adjustment_type: empForm.adjustment_type as any,
      adjust_by: empForm.adjust_by,
      adjust_value: empForm.adjust_value,
    })
    empPreview.value = (res as any).data ?? res
  } catch {
    ElMessage.error('预览失败')
  } finally {
    previewLoading.value = false
  }
}

async function submitEmpAdjustment() {
  if (!empForm.employee_id) {
    ElMessage.warning('请选择员工')
    return
  }
  if (!empForm.effective_month) {
    ElMessage.warning('请选择生效月份')
    return
  }
  submitLoading.value = true
  try {
    await salaryApi.createAdjustment({
      employee_id: empForm.employee_id,
      effective_month: empForm.effective_month,
      adjustment_type: empForm.adjustment_type as any,
      adjust_by: empForm.adjust_by,
      adjust_value: empForm.adjust_value,
    } as any)
    ElMessage.success(`调薪记录已保存，将于 ${empForm.effective_month} 月生效`)
    empForm.employee_id = null
    empForm.adjust_value = 0
    empPreview.value = null
  } catch {
    ElMessage.error('提交失败')
  } finally {
    submitLoading.value = false
  }
}

async function previewDeptAdjustment() {
  if (deptForm.department_ids.length === 0) {
    ElMessage.warning('请选择部门')
    return
  }
  if (!deptForm.effective_month) {
    ElMessage.warning('请选择生效月份')
    return
  }
  const ids = deptForm.department_ids.includes('__all__')
    ? departmentOptions.value.map((d) => d.id)
    : deptForm.department_ids
  previewLoading.value = true
  try {
    const res = await salaryApi.previewAdjustment({
      department_ids: ids,
      effective_month: deptForm.effective_month,
      adjustment_type: deptForm.adjustment_type as any,
      adjust_by: deptForm.adjust_by,
      adjust_value: deptForm.adjust_value,
    } as any)
    deptPreview.value = (res as any).data ?? res
  } catch {
    ElMessage.error('预览失败')
  } finally {
    previewLoading.value = false
  }
}

async function submitDeptAdjustment() {
  if (deptForm.department_ids.length === 0) {
    ElMessage.warning('请选择部门')
    return
  }
  if (!deptForm.effective_month) {
    ElMessage.warning('请选择生效月份')
    return
  }
  const ids = deptForm.department_ids.includes('__all__')
    ? departmentOptions.value.map((d) => d.id)
    : deptForm.department_ids
  submitLoading.value = true
  try {
    await salaryApi.massAdjustment({
      department_ids: ids,
      effective_month: deptForm.effective_month,
      adjustment_type: deptForm.adjustment_type as any,
      adjust_by: deptForm.adjust_by,
      adjust_value: deptForm.adjust_value,
    } as any)
    ElMessage.success(`调薪记录已保存，将于 ${deptForm.effective_month} 月生效`)
    deptForm.department_ids = []
    deptForm.adjust_value = 0
    deptPreview.value = null
  } catch {
    ElMessage.error('提交失败')
  } finally {
    submitLoading.value = false
  }
}

function formatAmount(val: number): string {
  const prefix = val >= 0 ? '+' : ''
  return `${prefix}¥${Math.abs(val).toLocaleString('zh-CN', { minimumFractionDigits: 2 })}`
}
</script>

<style scoped lang="scss">
.salary-adjustment {
  padding: 8px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title {
  font-size: 16px;
  font-weight: 700;
  color: #1a1a1a;
  margin: 0;
}

.adjustment-tabs {
  padding: 8px 0;
}

.adjustment-form {
  max-width: 600px;
}

.preview-panel {
  margin-top: 20px;
  max-width: 600px;
}

.impact-value {
  color: #1677ff;
  font-weight: 700;
  font-size: 15px;
}
</style>
