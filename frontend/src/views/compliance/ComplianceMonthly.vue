<template>
  <div class="page-view">
    <!-- 页面标题 -->
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
        <el-select
          v-model="selectedDepts"
          multiple
          collapse-tags
          collapse-tags-tooltip
          placeholder="选择部门"
          clearable
          class="dept-select"
          @change="load(1)"
        >
          <el-option
            v-for="d in deptOptions"
            :key="d.id"
            :label="d.name"
            :value="d.id"
          />
        </el-select>
        <el-button type="primary" size="large" @click="handleExport">
          <el-icon><Download /></el-icon>
          导出 Excel
        </el-button>
      </div>
    </header>

    <!-- 统计概览 -->
    <div class="stats-grid" v-loading="loading">
      <ComplianceStatCard
        :value="stats.total_required_days"
        label="应出勤天数"
        icon="Collection"
        icon-class="icon--required"
      />
      <ComplianceStatCard
        :value="stats.total_actual_days"
        label="实际出勤天数"
        icon="Calendar"
        icon-class="icon--actual"
      />
      <ComplianceStatCard
        :value="stats.total_overtime_hours"
        label="加班时长(h)"
        icon="Clock"
        icon-class="icon--overtime"
      />
      <ComplianceStatCard
        :value="stats.total_absent_days"
        label="缺勤天数"
        icon="WarningFilled"
        icon-class="icon--absent"
      />
    </div>

    <!-- 数据表格 -->
    <div class="table-card glass-card" v-loading="loading">
      <!-- 空状态 -->
      <div v-if="!loading && list.length === 0" class="empty-state">
        <div class="empty-icon">
          <el-icon><DataLine /></el-icon>
        </div>
        <h3>暂无月度汇总数据</h3>
        <p>当前月份暂无员工考勤汇总记录</p>
      </div>

      <!-- 表格 -->
      <el-table
        v-else
        :data="list"
        stripe
        class="modern-table"
        :row-class-name="getRowClassName"
        :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
      >
        <el-table-column prop="employee_name" label="姓名" min-width="120">
          <template #default="{ row }">
            <div class="name-cell">
              <el-avatar :size="32" class="name-avatar">
                {{ row.employee_name?.[0] || '?' }}
              </el-avatar>
              <span class="name-text">{{ row.employee_name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="department_name" label="部门" min-width="140">
          <template #default="{ row }">
            <span class="dept-text">{{ row.department_name || '未分配' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="应出勤(天)" min-width="110" align="right">
          <template #default="{ row }">
            <span class="num-val num-val--muted">{{ row.required_days }}</span>
          </template>
        </el-table-column>
        <el-table-column label="实际出勤(天)" min-width="120" align="right">
          <template #default="{ row }">
            <span
              class="num-val"
              :class="row.actual_days < row.required_days ? 'num-val--warning' : 'num-val--success'"
            >
              {{ row.actual_days }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="迟到(次)" min-width="100" align="right">
          <template #default="{ row }">
            <span class="num-val" :class="row.late_count > 0 ? 'num-val--warning' : ''">
              {{ row.late_count }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="早退(次)" min-width="100" align="right">
          <template #default="{ row }">
            <span class="num-val">{{ row.early_leave_count }}</span>
          </template>
        </el-table-column>
        <el-table-column label="缺勤(天)" min-width="100" align="right">
          <template #default="{ row }">
            <span class="num-val" :class="row.absent_days > 0 ? 'num-val--danger' : ''">
              {{ row.absent_days }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="加班(h)" min-width="100" align="right">
          <template #default="{ row }">
            <span class="num-val num-val--warning">{{ row.overtime_hours }}</span>
          </template>
        </el-table-column>
        <el-table-column label="年假(天)" min-width="100" align="right">
          <template #default="{ row }">
            <span class="num-val">{{ row.annual_leave_days }}</span>
          </template>
        </el-table-column>
        <el-table-column label="病假(天)" min-width="100" align="right">
          <template #default="{ row }">
            <span class="num-val" :class="row.sick_leave_days > 0 ? 'num-val--danger' : ''">
              {{ row.sick_leave_days }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="事假(天)" min-width="100" align="right">
          <template #default="{ row }">
            <span class="num-val" :class="row.personal_leave_days > 0 ? 'num-val--warning' : ''">
              {{ row.personal_leave_days }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="异常" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_anomaly" type="danger" size="small">异常</el-tag>
            <el-tag v-else type="success" size="small">正常</el-tag>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div v-if="list.length > 0" class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="load"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Download, DataLine } from '@element-plus/icons-vue'
import { departmentApi } from '@/api/department'
import { attendanceApi, type ComplianceMonthlyStats, type MonthlyComplianceItem } from '@/api/attendance'
import ComplianceStatCard from '@/components/compliance/ComplianceStatCard.vue'

const loading = ref(false)
const selectedMonth = ref(new Date().toISOString().slice(0, 7))
const selectedDepts = ref<number[]>([])
const deptOptions = ref<{ id: number; name: string }[]>([])
const list = ref<MonthlyComplianceItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const stats = ref<ComplianceMonthlyStats>({
  total_required_days: 0,
  total_actual_days: 0,
  total_overtime_hours: 0,
  total_absent_days: 0,
  total_anomaly_count: 0,
})

function getRowClassName({ row }: { row: MonthlyComplianceItem }) {
  return row.is_anomaly ? 'anomaly-row' : ''
}

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const deptIds = selectedDepts.value.length ? selectedDepts.value.join(',') : undefined
    const { data } = await attendanceApi.getComplianceMonthly({
      year_month: selectedMonth.value,
      dept_ids: deptIds,
      page: p,
      page_size: pageSize,
    })
    list.value = data?.list ?? []
    total.value = data?.total ?? 0
    if (data?.stats) {
      stats.value = data.stats
    }
  } catch {
    ElMessage.error('加载月度汇总数据失败')
  } finally {
    loading.value = false
  }
}

async function handleExport() {
  try {
    const deptIds = selectedDepts.value.length ? selectedDepts.value.join(',') : undefined
    const blob = await attendanceApi.exportComplianceMonthly({
      year_month: selectedMonth.value,
      dept_ids: deptIds,
    }) as unknown as Blob
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `月度考勤汇总_${selectedMonth.value}.xlsx`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch {
    ElMessage.error('导出失败')
  }
}

onMounted(async () => {
  try {
    const depts = await departmentApi.list()
    deptOptions.value = depts ?? []
  } catch {
    // dept options optional, continue without them
  }
  await load()
})
</script>

<style scoped lang="scss">
.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.month-picker {
  width: 150px;
}

.dept-select {
  width: 200px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}

.icon--required {
  background: linear-gradient(135deg, #DBEAFE, #BFDBFE);
  color: #3B82F6;
}

.icon--actual {
  background: linear-gradient(135deg, #D1FAE5, #A7F3D0);
  color: #10B981;
}

.icon--overtime {
  background: linear-gradient(135deg, #FEF3C7, #FDE68A);
  color: #F59E0B;
}

.icon--absent {
  background: linear-gradient(135deg, #FEE2E2, #FECACA);
  color: #EF4444;
}

.table-card {
  padding: 0;
  overflow: hidden;
}

:deep(.modern-table) {
  .el-table__header th {
    padding: 14px 12px;
    font-size: 13px;
  }

  .el-table__row {
    transition: background 0.2s ease;

    &:hover > td {
      background: rgba(124, 58, 237, 0.02) !important;
    }
  }

  .el-table__cell {
    padding: 14px 12px;
    border-bottom: 1px solid #F3F4F6;
  }
}

.name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.name-avatar {
  background: linear-gradient(135deg, var(--primary-light), var(--primary));
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
}

.name-text {
  font-weight: 500;
  color: var(--text-primary);
}

.dept-text {
  font-size: 13px;
  color: var(--text-tertiary);
}

.num-val {
  font-weight: 600;
  font-family: 'SF Mono', Monaco, monospace;
  font-size: 14px;

  &--muted {
    color: var(--text-tertiary);
  }

  &--warning {
    color: var(--warning);
  }

  &--danger {
    color: var(--danger);
  }

  &--success {
    color: var(--success);
  }
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: 16px 20px;
  border-top: 1px solid var(--border);
}

.empty-state {
  text-align: center;
  padding: 80px 32px;

  .empty-icon {
    width: 72px;
    height: 72px;
    margin: 0 auto 16px;
    background: linear-gradient(135deg, #EDE9FE, #DDD6FE);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 32px;
    color: var(--primary);

    .el-icon {
      font-size: 32px;
    }
  }

  h3 {
    font-size: 18px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 8px;
  }

  p {
    font-size: 14px;
    color: var(--text-tertiary);
    margin: 0;
  }
}

@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>

<style lang="scss">
// Non-scoped to apply to el-table shadow DOM rows
.anomaly-row td {
  background-color: rgba(239, 68, 68, 0.04) !important;
}
</style>
