<template>
  <div class="salary-adjustment">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">调薪管理</h1>
        <p class="page-subtitle">管理员工薪资调整与部门普调</p>
      </div>
    </header>

    <!-- 标签页 -->
    <div class="nav-tabs glass-card">
      <div class="tab-group">
        <button
          class="tab-btn"
          :class="{ active: activeSubTab === 'employee' }"
          @click="activeSubTab = 'employee'"
        >
          <el-icon><User /></el-icon>
          员工调薪
        </button>
        <button
          class="tab-btn"
          :class="{ active: activeSubTab === 'department' }"
          @click="activeSubTab = 'department'"
        >
          <el-icon><OfficeBuilding /></el-icon>
          部门普调
        </button>
      </div>
    </div>

    <!-- 员工调薪 -->
    <div v-show="activeSubTab === 'employee'" class="tab-content">
      <div class="form-card glass-card">
        <div class="form-grid">
          <div class="form-group">
            <label class="form-label">选择员工</label>
            <el-select
              v-model="empForm.employee_id"
              filterable
              remote
              placeholder="搜索员工姓名"
              :remote-method="searchEmployees"
              :loading="searchLoading"
              size="large"
              style="width: 100%"
              @focus="loadEmployees('')"
            >
              <el-option
                v-for="emp in employeeOptions"
                :key="emp.id"
                :label="emp.name"
                :value="emp.id"
              >
                <div class="emp-option">
                  <span>{{ emp.name }}</span>
                  <span class="emp-position">{{ emp.position }}</span>
                </div>
              </el-option>
            </el-select>
          </div>

          <div class="form-group">
            <label class="form-label">调整类型</label>
            <div class="type-chips">
              <label
                v-for="t in adjustmentTypes"
                :key="t.value"
                class="type-chip"
                :class="{ selected: empForm.adjustment_type === t.value }"
              >
                <input type="radio" :value="t.value" v-model="empForm.adjustment_type" class="hidden-check" />
                <span>{{ t.label }}</span>
              </label>
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">调整方式</label>
            <div class="toggle-group">
              <button
                class="toggle-btn"
                :class="{ active: empForm.adjust_by === 'amount' }"
                @click="empForm.adjust_by = 'amount'"
              >金额（元）</button>
              <button
                class="toggle-btn"
                :class="{ active: empForm.adjust_by === 'ratio' }"
                @click="empForm.adjust_by = 'ratio'"
              >比例（%）</button>
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">{{ empForm.adjust_by === 'amount' ? '调整金额（元）' : '调整比例（%）' }}</label>
            <el-input-number
              v-model="empForm.adjust_value"
              :min="0"
              :precision="empForm.adjust_by === 'ratio' ? 1 : 2"
              :step="empForm.adjust_by === 'ratio' ? 0.5 : 100"
              size="large"
              style="width: 100%"
            />
          </div>

          <div class="form-group">
            <label class="form-label">生效月份</label>
            <el-date-picker
              v-model="empForm.effective_month"
              type="month"
              placeholder="选择月份"
              value-format="YYYY-MM"
              size="large"
              style="width: 100%"
            />
          </div>
        </div>

        <div class="form-actions">
          <el-button :loading="previewLoading" size="large" @click="previewEmpAdjustment" class="preview-btn">
            <el-icon><View /></el-icon>
            预览影响
          </el-button>
          <el-button type="primary" :loading="submitLoading" size="large" @click="submitEmpAdjustment" class="submit-btn">
            <el-icon><Check /></el-icon>
            提交调薪
          </el-button>
        </div>
      </div>

      <!-- 预览面板 -->
      <Transition name="preview-slide">
        <div v-if="empPreview" class="preview-card glass-card">
          <div class="preview-header">
            <div class="preview-icon">
              <el-icon><DataLine /></el-icon>
            </div>
            <div>
              <h3>调薪预览</h3>
              <p>将于 {{ empPreview.effective_month }} 生效</p>
            </div>
          </div>
          <div class="preview-grid">
            <div class="preview-item">
              <span class="preview-label">影响员工</span>
              <span class="preview-value">{{ empPreview.employee_count }} 人</span>
            </div>
            <div class="preview-item">
              <span class="preview-label">生效月份</span>
              <span class="preview-value">{{ empPreview.effective_month }}</span>
            </div>
            <div class="preview-item">
              <span class="preview-label">月度影响</span>
              <span class="preview-value preview-value--warning">{{ formatAmount(empPreview.monthly_impact) }}</span>
            </div>
            <div class="preview-item">
              <span class="preview-label">年度影响</span>
              <span class="preview-value preview-value--warning">{{ formatAmount(empPreview.annual_impact) }}</span>
            </div>
          </div>
        </div>
      </Transition>
    </div>

    <!-- 部门普调 -->
    <div v-show="activeSubTab === 'department'" class="tab-content">
      <div class="form-card glass-card">
        <div class="form-grid">
          <div class="form-group form-group--full">
            <label class="form-label">选择部门</label>
            <el-select
              v-model="deptForm.department_ids"
              multiple
              filterable
              collapse-tags
              collapse-tags-tooltip
              placeholder="选择部门（可多选）"
              style="width: 100%"
              size="large"
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
          </div>

          <div class="form-group">
            <label class="form-label">调整类型</label>
            <div class="type-chips">
              <label
                v-for="t in adjustmentTypes"
                :key="t.value"
                class="type-chip"
                :class="{ selected: deptForm.adjustment_type === t.value }"
              >
                <input type="radio" :value="t.value" v-model="deptForm.adjustment_type" class="hidden-check" />
                <span>{{ t.label }}</span>
              </label>
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">调整方式</label>
            <div class="toggle-group">
              <button
                class="toggle-btn"
                :class="{ active: deptForm.adjust_by === 'amount' }"
                @click="deptForm.adjust_by = 'amount'"
              >金额（元）</button>
              <button
                class="toggle-btn"
                :class="{ active: deptForm.adjust_by === 'ratio' }"
                @click="deptForm.adjust_by = 'ratio'"
              >比例（%）</button>
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">{{ deptForm.adjust_by === 'amount' ? '调整金额（元）' : '调整比例（%）' }}</label>
            <el-input-number
              v-model="deptForm.adjust_value"
              :min="0"
              :precision="deptForm.adjust_by === 'ratio' ? 1 : 2"
              :step="deptForm.adjust_by === 'ratio' ? 0.5 : 100"
              size="large"
              style="width: 100%"
            />
          </div>

          <div class="form-group">
            <label class="form-label">生效月份</label>
            <el-date-picker
              v-model="deptForm.effective_month"
              type="month"
              placeholder="选择月份"
              value-format="YYYY-MM"
              size="large"
              style="width: 100%"
            />
          </div>
        </div>

        <div class="form-actions">
          <el-button :loading="previewLoading" size="large" @click="previewDeptAdjustment" class="preview-btn">
            <el-icon><View /></el-icon>
            预览影响
          </el-button>
          <el-button type="primary" :loading="submitLoading" size="large" @click="submitDeptAdjustment" class="submit-btn">
            <el-icon><Check /></el-icon>
            提交调薪
          </el-button>
        </div>
      </div>

      <!-- 预览面板 -->
      <Transition name="preview-slide">
        <div v-if="deptPreview" class="preview-card glass-card">
          <div class="preview-header">
            <div class="preview-icon">
              <el-icon><DataLine /></el-icon>
            </div>
            <div>
              <h3>调薪预览</h3>
              <p>将于 {{ deptPreview.effective_month }} 生效</p>
            </div>
          </div>
          <div class="preview-grid">
            <div class="preview-item">
              <span class="preview-label">影响部门</span>
              <span class="preview-value">{{ deptPreview.department_count }} 个</span>
            </div>
            <div class="preview-item">
              <span class="preview-label">影响员工</span>
              <span class="preview-value">{{ deptPreview.employee_count }} 人</span>
            </div>
            <div class="preview-item">
              <span class="preview-label">月度影响</span>
              <span class="preview-value preview-value--warning">{{ formatAmount(deptPreview.monthly_impact) }}</span>
            </div>
            <div class="preview-item">
              <span class="preview-label">年度影响</span>
              <span class="preview-value preview-value--warning">{{ formatAmount(deptPreview.annual_impact) }}</span>
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { salaryApi } from '@/api/salary'
import { employeeApi } from '@/api/employee'
import { departmentApi } from '@/api/department'
import { ElMessage } from 'element-plus'
import { User, OfficeBuilding, View, Check, DataLine } from '@element-plus/icons-vue'
import dayjs from 'dayjs'

const activeSubTab = ref('employee')
const searchLoading = ref(false)
const previewLoading = ref(false)
const submitLoading = ref(false)

interface Employee { id: number; name: string; position: string }
interface Department { id: number; name: string }
interface EmpPreview { employee_count: number; monthly_impact: number; annual_impact: number; effective_month: string }
interface DeptPreview { department_count: number; employee_count: number; monthly_impact: number; annual_impact: number; effective_month: string }

const employeeOptions = ref<Employee[]>([])
const departmentOptions = ref<Department[]>([])
const empPreview = ref<EmpPreview | null>(null)
const deptPreview = ref<DeptPreview | null>(null)

const adjustmentTypes = [
  { label: '岗位工资', value: 'base_salary' },
  { label: '补贴',     value: 'allowance'     },
  { label: '奖金',     value: 'bonus'         },
  { label: '年终奖',   value: 'year_end_bonus' },
  { label: '其他扣除', value: 'other'         },
]

const empForm = reactive({
  employee_id: null as number | null,
  adjustment_type: 'base_salary',
  adjust_by: 'amount' as 'amount' | 'ratio',
  adjust_value: 0,
  effective_month: dayjs().add(1, 'month').format('YYYY-MM'),
})

const deptForm = reactive({
  department_ids: [] as (number | string)[],
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
  } catch { /* ignore */ }
  finally { searchLoading.value = false }
}

async function searchEmployees(query: string) {
  if (query) await loadEmployees(query)
}

async function loadDepartments() {
  if (departmentOptions.value.length > 0) return
  try {
    departmentOptions.value = await departmentApi.list()
  } catch { /* ignore */ }
}

function toggleSelectAllDepts() {
  if (deptForm.department_ids.includes('__all__')) {
    deptForm.department_ids = departmentOptions.value.map((d) => d.id)
  } else {
    deptForm.department_ids = []
  }
}

async function previewEmpAdjustment() {
  if (!empForm.employee_id) { ElMessage.warning('请选择员工'); return }
  if (!empForm.effective_month) { ElMessage.warning('请选择生效月份'); return }
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
  } catch { ElMessage.error('预览失败') }
  finally { previewLoading.value = false }
}

async function submitEmpAdjustment() {
  if (!empForm.employee_id) { ElMessage.warning('请选择员工'); return }
  if (!empForm.effective_month) { ElMessage.warning('请选择生效月份'); return }
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
  } catch { ElMessage.error('提交失败') }
  finally { submitLoading.value = false }
}

async function previewDeptAdjustment() {
  if (deptForm.department_ids.length === 0) { ElMessage.warning('请选择部门'); return }
  if (!deptForm.effective_month) { ElMessage.warning('请选择生效月份'); return }
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
  } catch { ElMessage.error('预览失败') }
  finally { previewLoading.value = false }
}

async function submitDeptAdjustment() {
  if (deptForm.department_ids.length === 0) { ElMessage.warning('请选择部门'); return }
  if (!deptForm.effective_month) { ElMessage.warning('请选择生效月份'); return }
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
  } catch { ElMessage.error('提交失败') }
  finally { submitLoading.value = false }
}

function formatAmount(val: number): string {
  const prefix = val >= 0 ? '+' : ''
  return `${prefix}¥${Math.abs(val).toLocaleString('zh-CN', { minimumFractionDigits: 2 })}`
}
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

.salary-adjustment { padding: 24px 32px; width: 100%; box-sizing: border-box; background: $bg-page; min-height: 100vh; }

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
  display: inline-flex; align-items: center; gap: 6px;
  padding: 8px 20px;
  border-radius: $radius-sm;
  font-size: 14px; font-weight: 500; color: $text-secondary;
  cursor: pointer; transition: all 0.2s ease; border: none; background: transparent;

  &.active { background: #fff; color: var(--primary); box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1); }
  &:hover:not(.active) { color: $text-primary; }
  .el-icon { font-size: 15px; }
}

.tab-content { display: flex; flex-direction: column; gap: 20px; }

.form-card { padding: 28px; }

.form-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 20px; margin-bottom: 24px; }
.form-group { &--full { grid-column: 1 / -1; } }

.form-label { display: block; font-size: 13px; font-weight: 500; color: $text-secondary; margin-bottom: 8px; }

.type-chips { display: flex; gap: 8px; flex-wrap: wrap; }

.type-chip {
  display: inline-flex; align-items: center;
  padding: 8px 16px;
  border: 2px solid $border-color; border-radius: $radius-md;
  font-size: 13px; font-weight: 500; color: $text-secondary;
  cursor: pointer; transition: all 0.2s ease; user-select: none;

  &.selected { border-color: var(--primary); color: var(--primary); background: rgba(var(--primary), 0.06); }
  &:hover:not(.selected) { border-color: rgba(var(--primary), 0.4); color: var(--primary); }
  .hidden-check { display: none; }
}

.toggle-group { display: inline-flex; background: #F3F4F6; border-radius: $radius-md; padding: 4px; gap: 2px; }

.toggle-btn {
  padding: 8px 16px; border-radius: $radius-sm;
  font-size: 13px; font-weight: 500; color: $text-secondary;
  cursor: pointer; transition: all 0.2s ease; border: none; background: transparent;

  &.active { background: #fff; color: var(--primary); box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1); }
  &:hover:not(.active) { color: $text-primary; }
}

.emp-option { display: flex; justify-content: space-between; align-items: center; }
.emp-position { font-size: 12px; color: $text-muted; }

.form-actions { display: flex; gap: 12px; }

.preview-btn {
  border-style: dashed;
  color: var(--primary);
  border-color: rgba(var(--primary), 0.4);
  background: rgba(var(--primary), 0.04);
  &:hover { background: rgba(var(--primary), 0.08); border-color: var(--primary); }
}

.submit-btn {
  background: linear-gradient(135deg, var(--primary-light), var(--primary));
  border: none;
  box-shadow: 0 4px 14px rgba(var(--primary), 0.4);
  &:hover { box-shadow: 0 6px 20px rgba(var(--primary), 0.5); }
}

.preview-card {
  padding: 24px;
  animation: fadeInUp 0.3s ease;
}

.preview-header { display: flex; align-items: center; gap: 12px; margin-bottom: 20px; }

.preview-icon {
  width: 44px; height: 44px;
  border-radius: $radius-md;
  background: linear-gradient(135deg, #DBEAFE, #BFDBFE);
  color: #3B82F6;
  display: flex; align-items: center; justify-content: center;
  font-size: 20px;
}

.preview-header h3 { font-size: 16px; font-weight: 700; color: $text-primary; margin: 0 0 2px; }
.preview-header p { font-size: 13px; color: $text-muted; margin: 0; }

.preview-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; }

.preview-item {
  display: flex; flex-direction: column; gap: 4px;
  padding: 16px;
  background: $bg-page;
  border-radius: $radius-md;
  border: 1px solid $border-color;
}

.preview-label { font-size: 12px; color: $text-muted; }
.preview-value { font-size: 18px; font-weight: 700; color: $text-primary;
  &--warning { color: $warning; }
}

@keyframes fadeInUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }

.preview-slide-enter-active, .preview-slide-leave-active { transition: all 0.3s ease; }
.preview-slide-enter-from, .preview-slide-leave-to { opacity: 0; transform: translateY(-12px); }

@media (max-width: 768px) {
  .salary-adjustment { padding: 16px; }
  .form-grid { grid-template-columns: 1fr; }
  .form-group--full { grid-column: 1; }
  .preview-grid { grid-template-columns: repeat(2, 1fr); }
  .form-actions { flex-direction: column; }
}
</style>
