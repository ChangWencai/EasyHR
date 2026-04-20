<template>
  <div class="page-view">
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">加班统计</h1>
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

    <!-- 统计概览 -->
    <div class="stats-grid" v-loading="loading">
      <ComplianceStatCard :value="stats.total_holiday_hours" label="法定节假日加班(h)" icon="Calendar" icon-class="icon--holiday" />
      <ComplianceStatCard :value="stats.total_weekday_hours" label="工作日延时加班(h)" icon="Clock" icon-class="icon--weekday" />
      <ComplianceStatCard :value="stats.total_weekend_hours" label="周末加班(h)" icon="Sunny" icon-class="icon--weekend" />
    </div>

    <!-- 数据表格 -->
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
      <el-table-column label="法定节假日(h)" min-width="120" align="right">
        <template #default="{ row }"><span class="num-val">{{ row.holiday_hours }}</span></template>
      </el-table-column>
      <el-table-column label="工作日延时(h)" min-width="120" align="right">
        <template #default="{ row }"><span class="num-val">{{ row.weekday_hours }}</span></template>
      </el-table-column>
      <el-table-column label="周末加班(h)" min-width="120" align="right">
        <template #default="{ row }"><span class="num-val">{{ row.weekend_hours }}</span></template>
      </el-table-column>
      <el-table-column label="合计(h)" min-width="100" align="right">
        <template #default="{ row }"><span class="num-val num-val--warning">{{ row.total_hours }}</span></template>
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
const deptOptions = ref<{ id: number; name: string }[]>([])
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const stats = ref({ total_holiday_hours: 0, total_weekday_hours: 0, total_weekend_hours: 0 })

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const deptIds = selectedDepts.value.length ? selectedDepts.value.join(',') : undefined
    const { data } = await attendanceApi.getComplianceOvertime({ year_month: selectedMonth.value, dept_ids: deptIds, page: p, page_size: pageSize })
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
.stats-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; margin-bottom: 20px; }
.icon--holiday { background: linear-gradient(135deg, #FEE2E2, #FECACA); color: #EF4444; }
.icon--weekday { background: linear-gradient(135deg, #FEF3C7, #FDE68A); color: #F59E0B; }
.icon--weekend { background: linear-gradient(135deg, #DBEAFE, #BFDBFE); color: #3B82F6; }
.name-cell { display: flex; align-items: center; gap: 10px; }
.name-avatar { background: linear-gradient(135deg, var(--primary-light), var(--primary)); color: #fff; font-size: 13px; font-weight: 600; }
.name-text { font-weight: 500; }
.num-val { font-weight: 600; font-family: 'SF Mono', Monaco, monospace; &--warning { color: var(--warning); } }
</style>
