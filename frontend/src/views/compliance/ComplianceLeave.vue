<template>
  <div class="page-view">
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">请假合规</h1>
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
      </div>
    </header>

    <div class="stats-grid" v-loading="loading">
      <ComplianceStatCard :value="stats.annual_quota_employee_count" label="年假员工人数" icon="User" icon-class="icon--quota" />
      <ComplianceStatCard :value="stats.total_annual_used" label="年假已用总计(天)" icon="Calendar" icon-class="icon--annual" />
      <ComplianceStatCard :value="stats.total_sick_days" label="病假总计(天)" icon="FirstAidKit" icon-class="icon--sick" />
      <ComplianceStatCard :value="stats.total_personal_days" label="事假总计(天)" icon="Memo" icon-class="icon--personal" />
    </div>

    <ComplianceTable
      v-loading="loading"
      :data="list"
      :total="total"
      :page="page"
      :page-size="pageSize"
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
      <el-table-column label="年假额度(天)" min-width="100" align="right">
        <template #default="{ row }"><span class="num-val">{{ row.annual_quota }}</span></template>
      </el-table-column>
      <el-table-column label="已用(天)" min-width="100" align="right">
        <template #default="{ row }"><span class="num-val num-val--warning">{{ row.annual_used }}</span></template>
      </el-table-column>
      <el-table-column label="剩余(天)" min-width="100" align="right">
        <template #default="{ row }"><span class="num-val num-val--success">{{ row.annual_left }}</span></template>
      </el-table-column>
      <el-table-column label="病假(天)" min-width="90" align="right">
        <template #default="{ row }"><span class="num-val">{{ row.sick_days }}</span></template>
      </el-table-column>
      <el-table-column label="事假(天)" min-width="90" align="right">
        <template #default="{ row }"><span class="num-val">{{ row.personal_days }}</span></template>
      </el-table-column>
    </ComplianceTable>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { attendanceApi } from '@/api/attendance'
import ComplianceStatCard from '@/components/compliance/ComplianceStatCard.vue'
import ComplianceTable from '@/components/compliance/ComplianceTable.vue'

const loading = ref(false)
const selectedMonth = ref(new Date().toISOString().slice(0, 7))
const selectedDepts = ref<number[]>([])
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const stats = ref({ annual_quota_employee_count: 0, total_annual_used: 0, total_sick_days: 0, total_personal_days: 0 })
const deptOptions = ref<{ id: number; name: string }[]>([])

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const deptIds = selectedDepts.value.length ? selectedDepts.value.join(',') : undefined
    const { data } = await attendanceApi.getComplianceLeave({ year_month: selectedMonth.value, dept_ids: deptIds, page: p, page_size: pageSize })
    list.value = data?.list ?? []
    total.value = data?.total ?? 0
    if (data?.stats) stats.value = data.stats
  } finally { loading.value = false }
}

onMounted(() => load())
</script>

<style scoped lang="scss">
.header-actions { display: flex; align-items: center; gap: 12px; }
.month-picker { width: 150px; }
.dept-select { width: 200px; }
.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 20px; }
.icon--quota { background: linear-gradient(135deg, #EDE9FE, #DDD6FE); color: #7C3AED; }
.icon--annual { background: linear-gradient(135deg, #D1FAE5, #A7F3D0); color: #10B981; }
.icon--sick { background: linear-gradient(135deg, #FEE2E2, #FECACA); color: #EF4444; }
.icon--personal { background: linear-gradient(135deg, #FEF3C7, #FDE68A); color: #F59E0B; }
.name-cell { display: flex; align-items: center; gap: 10px; }
.name-avatar { background: linear-gradient(135deg, var(--primary-light), var(--primary)); color: #fff; font-size: 13px; font-weight: 600; }
.name-text { font-weight: 500; }
.num-val { font-weight: 600; font-family: 'SF Mono', Monaco, monospace; &--success { color: var(--success); } &--warning { color: var(--warning); } }
</style>
