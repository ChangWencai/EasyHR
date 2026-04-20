<template>
  <div class="page-view">
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">月度汇总</h1>
        <p class="page-subtitle">{{ selectedMonth || '请选择月份' }}</p>
      </div>
      <div class="header-actions">
        <el-date-picker
          v-model="selectedMonth"
          type="month"
          format="YYYY-MM"
          value-format="YYYY-MM"
          placeholder="选择月份"
          size="large"
          class="month-picker"
          @change="load(1)"
        />
        <el-select v-model="selectedDepts" multiple collapse-tags collapse-tags-tooltip placeholder="选择部门" clearable class="dept-select" @change="load(1)">
          <el-option v-for="d in deptOptions" :key="d.id" :label="d.name" :value="d.id" />
        </el-select>
        <el-button type="primary" size="large" @click="handleExport">
          <el-icon><Download /></el-icon>
          导出 Excel
        </el-button>
      </div>
    </header>

    <div class="stats-grid" v-loading="loading">
      <ComplianceStatCard :value="stats.total_required_days" label="应出勤总计(天)" icon="Collection" icon-class="icon--required" />
      <ComplianceStatCard :value="stats.total_actual_days" label="实际出勤总计(天)" icon="Calendar" icon-class="icon--actual" />
      <ComplianceStatCard :value="stats.total_overtime_hours" label="加班总计(h)" icon="Clock" icon-class="icon--overtime" />
      <ComplianceStatCard :value="stats.total_absent_days" label="缺勤总计(天)" icon="WarningFilled" icon-class="icon--absent" />
      <ComplianceStatCard :value="stats.total_anomaly_count" label="异常人次" icon="WarnTriangleFilled" icon-class="icon--anomaly" />
    </div>

    <ComplianceTable
      v-loading="loading"
      :data="list"
      :total="total"
      :page="page"
      :page-size="pageSize"
      :row-class-name="getRowClass"
      @page-change="load"
    >
      <el-table-column prop="employee_name" label="姓名" min-width="100">
        <template #default="{ row }">
          <div class="name-cell">
            <el-avatar :size="32" class="name-avatar">{{ row.employee_name?.[0] || '?' }}</el-avatar>
            <span class="name-text">{{ row.employee_name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="department_name" label="部门" min-width="120" />
      <el-table-column label="应出勤(天)" min-width="90" align="right">
        <template #default="{ row }"><span class="num-val num-val--muted">{{ row.required_days }}</span></template>
      </el-table-column>
      <el-table-column label="实际(天)" min-width="90" align="right">
        <template #default="{ row }"><span class="num-val" :class="row.actual_days < row.required_days ? 'num-val--warning' : 'num-val--success'">{{ row.actual_days }}</span></template>
      </el-table-column>
      <el-table-column label="加班(h)" min-width="90" align="right">
        <template #default="{ row }"><span class="num-val num-val--warning">{{ row.overtime_hours }}</span></template>
      </el-table-column>
      <el-table-column label="缺勤(天)" min-width="90" align="right">
        <template #default="{ row }"><span class="num-val" :class="row.absent_days > 0 ? 'num-val--danger' : ''">{{ row.absent_days }}</span></template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag v-if="row.is_anomaly" type="danger">异常</el-tag>
          <el-tag v-else type="success">正常</el-tag>
        </template>
      </el-table-column>
    </ComplianceTable>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { attendanceApi } from '@/api/attendance'
import ComplianceStatCard from '@/components/compliance/ComplianceStatCard.vue'
import ComplianceTable from '@/components/compliance/ComplianceTable.vue'
import { Download } from '@element-plus/icons-vue'

const loading = ref(false)
const selectedMonth = ref(new Date().toISOString().slice(0, 7))
const selectedDepts = ref<number[]>([])
const deptOptions = ref<{ id: number; name: string }[]>([])
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const stats = ref({ total_required_days: 0, total_actual_days: 0, total_overtime_hours: 0, total_absent_days: 0, total_anomaly_count: 0 })

function getRowClass({ row }: { row: any }) {
  return row.is_anomaly ? 'anomaly-row' : ''
}

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const deptIds = selectedDepts.value.length ? selectedDepts.value.join(',') : undefined
    const { data } = await attendanceApi.getComplianceMonthly({ year_month: selectedMonth.value, dept_ids: deptIds, page: p, page_size: pageSize })
    list.value = data?.list ?? []
    total.value = data?.total ?? 0
    if (data?.stats) stats.value = data.stats
  } finally { loading.value = false }
}

async function handleExport() {
  try {
    const deptIds = selectedDepts.value.length ? selectedDepts.value.join(',') : undefined
    const blob = await attendanceApi.exportComplianceMonthly({ year_month: selectedMonth.value, dept_ids: deptIds }) as unknown as Blob
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `考勤月报汇总_${selectedMonth.value}.xlsx`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch { ElMessage.error('导出失败') }
}

onMounted(() => load())
</script>

<style scoped lang="scss">
.header-actions { display: flex; align-items: center; gap: 12px; }
.month-picker { width: 150px; }
.dept-select { width: 200px; }
.stats-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 16px; margin-bottom: 20px; }
.icon--required { background: linear-gradient(135deg, #DBEAFE, #BFDBFE); color: #3B82F6; }
.icon--actual { background: linear-gradient(135deg, #D1FAE5, #A7F3D0); color: #10B981; }
.icon--overtime { background: linear-gradient(135deg, #FEF3C7, #FDE68A); color: #F59E0B; }
.icon--absent { background: linear-gradient(135deg, #FEE2E2, #FECACA); color: #EF4444; }
.icon--anomaly { background: linear-gradient(135deg, #FEE2E2, #FECACA); color: #EF4444; }
.name-cell { display: flex; align-items: center; gap: 10px; }
.name-avatar { background: linear-gradient(135deg, var(--primary-light), var(--primary)); color: #fff; font-size: 13px; font-weight: 600; }
.name-text { font-weight: 500; }
.num-val { font-weight: 600; font-family: 'SF Mono', Monaco, monospace; &--muted { color: var(--text-tertiary); } &--warning { color: var(--warning); } &--danger { color: var(--danger); } &--success { color: var(--success); } }
</style>

<style lang="scss">
.anomaly-row td { background-color: rgba(239, 68, 68, 0.04) !important; }
</style>
