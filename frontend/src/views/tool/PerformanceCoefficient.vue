<template>
  <div class="performance-coefficient">
    <el-card>
      <template #header>
        <div class="page-header">
          <h1 class="page-title">绩效系数设置</h1>
          <div class="header-actions">
            <el-date-picker
              v-model="selectedYM"
              type="month"
              placeholder="选择月份"
              value-format="YYYY-MM"
              style="width: 160px"
              @change="loadData"
            />
            <el-button type="primary" :loading="saving" :disabled="!hasDirty" @click="saveCoefficients">
              保存系数
            </el-button>
          </div>
        </div>
      </template>

      <!-- 筛选栏 -->
      <el-form inline @submit.prevent>
        <el-form-item label="搜索员工">
          <el-input v-model="searchKeyword" placeholder="搜索员工姓名" clearable style="width: 160px" @input="loadEmployees" />
        </el-form-item>
        <el-form-item label="部门">
          <el-select v-model="selectedDeptId" placeholder="全部部门" clearable style="width: 160px" @change="loadEmployees">
            <el-option label="全部部门" :value="undefined" />
            <el-option v-for="dept in departments" :key="dept.id" :label="dept.name" :value="dept.id" />
          </el-select>
        </el-form-item>
      </el-form>

      <!-- 提示信息 -->
      <el-alert type="info" :closable="false" style="margin-bottom: 16px">
        实际绩效工资 = 绩效工资标准 &times; 绩效系数。系数范围: 0% - 200%，步进 5%。100% 为默认系数。
      </el-alert>

      <!-- 员工系数表格 -->
      <el-table :data="employeeList" stripe v-loading="loading">
        <el-table-column prop="employee_name" label="姓名" min-width="100" />
        <el-table-column prop="department_name" label="部门" min-width="120" />
        <el-table-column prop="standard_amount" label="绩效工资标准" width="140">
          <template #default="{ row }">¥{{ (row.standard_amount || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2 }) }}</template>
        </el-table-column>
        <el-table-column label="绩效系数" width="280">
          <template #default="{ row }">
            <div class="slider-cell">
              <el-slider
                v-model="row.coefficient"
                :min="0"
                :max="200"
                :step="5"
                :show-tooltip="false"
                style="flex: 1"
                @change="markDirty(row)"
              />
              <span class="coef-label" :class="{ 'is-default': row.coefficient === 100 }">
                {{ row.coefficient }}%
                <el-tag v-if="row.coefficient === 100" size="small" type="info" style="margin-left: 4px">默认</el-tag>
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="实际绩效工资" width="140">
          <template #default="{ row }">
            <span class="actual-amount">
              ¥{{ ((row.standard_amount || 0) * (row.coefficient / 100)).toLocaleString('zh-CN', { minimumFractionDigits: 2 }) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.coefficient !== 100"
              size="small"
              link
              @click="resetRow(row)"
            >
              重置
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="mt-4"
        layout="total,prev,pager,next"
        :total="total"
        :page="page"
        :page-size="pageSize"
        @current-change="loadEmployees"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { salaryApi } from '@/api/salary'
import { employeeApi } from '@/api/employee'
import { departmentApi } from '@/api/department'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'

interface EmployeeCoeffRow {
  employee_id: number
  employee_name: string
  department_name: string
  standard_amount: number
  coefficient: number
  _dirty: boolean
  _original: number
}

const selectedYM = ref(dayjs().format('YYYY-MM'))
const searchKeyword = ref('')
const selectedDeptId = ref<number | undefined>(undefined)
const loading = ref(false)
const saving = ref(false)
const departments = ref<{ id: number; name: string }[]>([])
const employeeList = ref<EmployeeCoeffRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const hasDirty = computed(() => employeeList.value.some((r) => r._dirty))

async function loadData() {
  await Promise.all([loadDepartments(), loadEmployees()])
}

async function loadDepartments() {
  try {
    const res = await departmentApi.list()
    departments.value = (res as any[]).map((d) => ({ id: d.id, name: d.name }))
  } catch {
    // ignore
  }
}

async function loadEmployees(p = 1) {
  if (!selectedYM.value) return
  page.value = p
  loading.value = true
  try {
    const [yearStr, monthStr] = selectedYM.value.split('-')
    const year = parseInt(yearStr, 10)
    const month = parseInt(monthStr, 10)

    // Load performance coefficients
    const coeffRes = await salaryApi.getPerformance(year, month)
    const coeffMap: Record<number, number> = {}
    ;(coeffRes as any[] || []).forEach((c) => {
      coeffMap[c.employee_id] = c.coefficient
    })

    // Load employees
    const empRes = await employeeApi.list({ page: p, page_size: pageSize.value, search: searchKeyword.value || undefined })
    const rawList = (empRes as any).list || empRes || []

    // Load performance standard amounts via employee items
    const rows: EmployeeCoeffRow[] = []
    for (const emp of rawList) {
      if (emp.status === 'inactive') continue
      try {
        const itemsRes = await salaryApi.employeeItems(emp.id, selectedYM.value)
        const items = (itemsRes as any[]) || []
        // Find performance salary item
        const perfItem = items.find((i: any) => i.name && i.name.includes('绩效'))
        const perfAmount = perfItem ? perfItem.amount : 0
        const coeff = coeffMap[emp.id] ?? 100
        rows.push({
          employee_id: emp.id,
          employee_name: emp.name,
          department_name: (emp as any).department_name || '-',
          standard_amount: perfAmount,
          coefficient: coeff,
          _dirty: false,
          _original: coeff,
        })
      } catch {
        rows.push({
          employee_id: emp.id,
          employee_name: emp.name,
          department_name: (emp as any).department_name || '-',
          standard_amount: 0,
          coefficient: coeffMap[emp.id] ?? 100,
          _dirty: false,
          _original: coeffMap[emp.id] ?? 100,
        })
      }
    }

    employeeList.value = rows
    total.value = typeof (empRes as any).total === 'number' ? (empRes as any).total : rawList.length
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

function markDirty(row: EmployeeCoeffRow) {
  if (row.coefficient !== row._original) {
    row._dirty = true
  } else {
    row._dirty = false
  }
}

function resetRow(row: EmployeeCoeffRow) {
  row.coefficient = 100
  row._dirty = false
}

async function saveCoefficients() {
  if (!selectedYM.value) return
  const [yearStr, monthStr] = selectedYM.value.split('-')
  const year = parseInt(yearStr, 10)
  const month = parseInt(monthStr, 10)
  const dirtyRows = employeeList.value.filter((r) => r._dirty)
  if (dirtyRows.length === 0) return

  saving.value = true
  try {
    await salaryApi.setPerformance({
      year,
      month,
      coefficients: dirtyRows.map((r) => ({
        employee_id: r.employee_id,
        coefficient: r.coefficient,
      })),
    })
    ElMessage.success('绩效系数已保存')
    // Mark saved rows as clean
    dirtyRows.forEach((r) => {
      r._original = r.coefficient
      r._dirty = false
    })
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.performance-coefficient {
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

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.slider-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.coef-label {
  min-width: 64px;
  text-align: right;
  font-weight: 700;
  color: #1677ff;
  font-size: 14px;

  &.is-default {
    color: #8c8c8c;
  }
}

.actual-amount {
  font-weight: 700;
  color: #52c41a;
  font-size: 14px;
}

.mt-4 {
  margin-top: 16px;
}
</style>
