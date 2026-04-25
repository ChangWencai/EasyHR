<template>
  <div class="page-view">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">出勤异常</h1>
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
      </div>
    </header>

    <!-- 统计概览 -->
    <div class="stats-grid" v-loading="loading">
      <ComplianceStatCard
        :value="stats.anomaly_employee_count"
        label="异常人数"
        icon="WarningFilled"
        icon-class="icon--anomaly"
      />
      <ComplianceStatCard
        :value="stats.total_late_count"
        label="迟到总次数"
        icon="Clock"
        icon-class="icon--late"
      />
      <ComplianceStatCard
        :value="stats.total_absent_days"
        label="缺勤总天数"
        icon="CloseBold"
        icon-class="icon--absent"
      />
      <ComplianceStatCard
        :value="normalCount"
        label="正常员工数"
        icon="CircleCheckFilled"
        icon-class="icon--normal"
      />
    </div>

    <!-- 数据表格 -->
    <div class="table-card glass-card" v-loading="loading">
      <!-- 空状态 -->
      <div v-if="!loading && list.length === 0" class="empty-state">
        <div class="empty-icon">
          <el-icon><CircleCheckFilled /></el-icon>
        </div>
        <h3>暂无异常数据</h3>
        <p>当前月份暂无员工出勤异常记录</p>
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
        <el-table-column label="异常标记" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_anomaly" type="danger" size="small">
              {{ row.anomaly_count }} 项异常
            </el-tag>
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
import { CircleCheckFilled } from '@element-plus/icons-vue'
import { departmentApi } from '@/api/department'
import { attendanceApi, type ComplianceAnomalyStats, type AnomalyItem } from '@/api/attendance'
import ComplianceStatCard from '@/components/attendance/ComplianceStatCard.vue'

const loading = ref(false)
const selectedMonth = ref(new Date().toISOString().slice(0, 7))
const selectedDepts = ref<number[]>([])
const deptOptions = ref<{ id: number; name: string }[]>([])
const list = ref<AnomalyItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const stats = ref<ComplianceAnomalyStats>({
  anomaly_employee_count: 0,
  total_late_count: 0,
  total_absent_days: 0,
})
const normalCount = ref(0)

function getRowClassName({ row }: { row: AnomalyItem }) {
  return row.is_anomaly ? 'anomaly-row' : ''
}

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const deptIds = selectedDepts.value.length ? selectedDepts.value.join(',') : undefined
    const { data } = await attendanceApi.getComplianceAnomaly({
      year_month: selectedMonth.value,
      dept_ids: deptIds,
      page: p,
      page_size: pageSize,
    })
    list.value = data?.list ?? []
    total.value = data?.total ?? 0
    if (data?.stats) {
      stats.value = data.stats
      normalCount.value = (data.total ?? 0) - (data.stats.anomaly_employee_count ?? 0)
    }
  } catch {
    ElMessage.error('加载出勤异常数据失败')
  } finally {
    loading.value = false
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

.icon--anomaly {
  background: linear-gradient(135deg, #FEE2E2, #FECACA);
  color: #EF4444;
}

.icon--late {
  background: linear-gradient(135deg, #FEF3C7, #FDE68A);
  color: #F59E0B;
}

.icon--absent {
  background: linear-gradient(135deg, #FEE2E2, #FECACA);
  color: #EF4444;
}

.icon--normal {
  background: linear-gradient(135deg, #D1FAE5, #A7F3D0);
  color: #10B981;
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

  &--warning {
    color: var(--warning);
  }

  &--danger {
    color: var(--danger);
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
    background: linear-gradient(135deg, #D1FAE5, #A7F3D0);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 32px;
    color: var(--success);

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
