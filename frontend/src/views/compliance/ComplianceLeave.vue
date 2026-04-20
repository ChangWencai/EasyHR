<template>
  <div class="page-view">
    <!-- 页面标题 -->
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
        :value="stats.annual_quota_employee_count"
        label="年假员工人数"
        icon="User"
        icon-class="icon--quota"
      />
      <ComplianceStatCard
        :value="stats.total_annual_used"
        label="年假已用总计(天)"
        icon="Calendar"
        icon-class="icon--annual"
      />
      <ComplianceStatCard
        :value="stats.total_sick_days"
        label="病假总计(天)"
        icon="FirstAidKit"
        icon-class="icon--sick"
      />
      <ComplianceStatCard
        :value="stats.total_personal_days"
        label="事假总计(天)"
        icon="Memo"
        icon-class="icon--personal"
      />
    </div>

    <!-- 数据表格 -->
    <div class="table-card glass-card" v-loading="loading">
      <!-- 空状态 -->
      <div v-if="!loading && list.length === 0" class="empty-state">
        <div class="empty-icon">
          <el-icon><Calendar /></el-icon>
        </div>
        <h3>暂无请假数据</h3>
        <p>当前月份暂无员工请假记录</p>
      </div>

      <!-- 表格 -->
      <el-table
        v-else
        :data="list"
        stripe
        class="modern-table"
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
        <el-table-column label="年假额度(天)" min-width="130" align="right">
          <template #default="{ row }">
            <span class="num-val num-val--muted">{{ row.annual_quota }}</span>
          </template>
        </el-table-column>
        <el-table-column label="已用年假(天)" min-width="130" align="right">
          <template #default="{ row }">
            <span class="num-val num-val--warning">{{ row.annual_used }}</span>
          </template>
        </el-table-column>
        <el-table-column label="剩余年假(天)" min-width="130" align="right">
          <template #default="{ row }">
            <span class="num-val num-val--success">{{ row.annual_left }}</span>
          </template>
        </el-table-column>
        <el-table-column label="病假(天)" min-width="100" align="right">
          <template #default="{ row }">
            <span class="num-val" :class="row.sick_days > 0 ? 'num-val--danger' : ''">{{ row.sick_days }}</span>
          </template>
        </el-table-column>
        <el-table-column label="事假(天)" min-width="100" align="right">
          <template #default="{ row }">
            <span class="num-val" :class="row.personal_days > 0 ? 'num-val--warning' : 'num-val--muted'">{{ row.personal_days }}</span>
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
import { Calendar } from '@element-plus/icons-vue'
import { departmentApi } from '@/api/department'
import { attendanceApi, type ComplianceLeaveStats, type LeaveItem } from '@/api/attendance'
import ComplianceStatCard from '@/components/compliance/ComplianceStatCard.vue'

const loading = ref(false)
const selectedMonth = ref(new Date().toISOString().slice(0, 7))
const selectedDepts = ref<number[]>([])
const deptOptions = ref<{ id: number; name: string }[]>([])
const list = ref<LeaveItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const stats = ref<ComplianceLeaveStats>({
  annual_quota_employee_count: 0,
  total_annual_used: 0,
  total_sick_days: 0,
  total_personal_days: 0,
})

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const deptIds = selectedDepts.value.length ? selectedDepts.value.join(',') : undefined
    const { data } = await attendanceApi.getComplianceLeave({
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
    ElMessage.error('加载请假合规数据失败')
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

.icon--quota {
  background: linear-gradient(135deg, #EDE9FE, #DDD6FE);
  color: #7C3AED;
}

.icon--annual {
  background: linear-gradient(135deg, #D1FAE5, #A7F3D0);
  color: #10B981;
}

.icon--sick {
  background: linear-gradient(135deg, #FEE2E2, #FECACA);
  color: #EF4444;
}

.icon--personal {
  background: linear-gradient(135deg, #FEF3C7, #FDE68A);
  color: #F59E0B;
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
