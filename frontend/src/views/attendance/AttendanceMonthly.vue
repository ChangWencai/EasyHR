<template>
  <div class="page-view">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">出勤月报</h1>
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
          <el-option v-for="d in deptOptions" :key="d.id" :label="d.name" :value="d.id" />
        </el-select>
        <el-button type="primary" size="large" @click="handleExport">
          <el-icon><Download /></el-icon>
          导出月报
        </el-button>
      </div>
    </header>

    <!-- 视图切换 -->
    <div class="view-toggle-bar glass-card">
      <div class="toggle-group">
        <button
          class="toggle-btn"
          :class="{ active: viewMode === 'stat' }"
          @click="viewMode = 'stat'"
        >
          <el-icon><DataAnalysis /></el-icon>
          统计视图
        </button>
        <button
          class="toggle-btn"
          :class="{ active: viewMode === 'grid' }"
          @click="viewMode = 'grid'"
        >
          <el-icon><Grid /></el-icon>
          格子视图
        </button>
      </div>
    </div>

    <!-- 统计视图 -->
    <div v-if="viewMode === 'stat'" v-loading="loading">
      <!-- 统计概览 -->
      <div class="stats-grid" v-if="list.length > 0">
        <div class="stat-card glass-card">
          <div class="stat-icon stat-icon--actual">
            <el-icon><Calendar /></el-icon>
          </div>
          <div class="stat-body">
            <div class="stat-value">{{ stats.total_actual_days }}</div>
            <div class="stat-label">实际出勤（天）</div>
          </div>
        </div>
        <div class="stat-card glass-card">
          <div class="stat-icon stat-icon--required">
            <el-icon><Collection /></el-icon>
          </div>
          <div class="stat-body">
            <div class="stat-value">{{ stats.total_required_days }}</div>
            <div class="stat-label">应出勤（天）</div>
          </div>
        </div>
        <div class="stat-card glass-card">
          <div class="stat-icon stat-icon--overtime">
            <el-icon><Clock /></el-icon>
          </div>
          <div class="stat-body">
            <div class="stat-value">{{ stats.total_overtime_hours }}<span class="stat-unit">h</span></div>
            <div class="stat-label">加班时长</div>
          </div>
        </div>
        <div class="stat-card glass-card">
          <div class="stat-icon stat-icon--absent">
            <el-icon><WarningFilled /></el-icon>
          </div>
          <div class="stat-body">
            <div class="stat-value stat-value--danger">{{ stats.total_absent_days }}</div>
            <div class="stat-label">缺勤（天）</div>
          </div>
        </div>
      </div>

      <!-- 数据表格 -->
      <div class="table-card glass-card" v-if="list.length > 0">
        <el-table
          :data="list"
          stripe
          class="modern-table"
          :row-class-name="getRowClassName"
          :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
        >
          <el-table-column prop="employee_name" label="姓名" min-width="100">
            <template #default="{ row }">
              <div class="name-cell">
                <el-avatar :size="32" class="name-avatar">{{ row.employee_name?.[0] || '?' }}</el-avatar>
                <span class="name-text">{{ row.employee_name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="department_name" label="部门" min-width="110">
            <template #default="{ row }">
              <span class="dept-text">{{ row.department_name || '未分配' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="实际出勤" min-width="100">
            <template #default="{ row }">
              <span class="metric-value">{{ row.actual_days }}<span class="metric-unit">天</span></span>
            </template>
          </el-table-column>
          <el-table-column label="应出勤" min-width="100">
            <template #default="{ row }">
              <span class="metric-value metric-value--muted">{{ row.required_days }}<span class="metric-unit">天</span></span>
            </template>
          </el-table-column>
          <el-table-column label="加班" min-width="90">
            <template #default="{ row }">
              <span class="metric-value metric-value--warning">{{ row.overtime_hours }}<span class="metric-unit">h</span></span>
            </template>
          </el-table-column>
          <el-table-column label="缺勤" min-width="90">
            <template #default="{ row }">
              <span class="metric-value" :class="row.absent_days > 0 ? 'metric-value--danger' : 'metric-value--muted'">
                {{ row.absent_days }}<span class="metric-unit">天</span>
              </span>
            </template>
          </el-table-column>
          <el-table-column label="迟到(次)" min-width="90">
            <template #default="{ row }">
              <span class="metric-value" :class="row.late_count > 0 ? 'metric-value--warning' : ''">{{ row.late_count }}</span>
            </template>
          </el-table-column>
          <el-table-column label="早退(次)" min-width="90">
            <template #default="{ row }">
              <span class="metric-value">{{ row.early_leave_count }}</span>
            </template>
          </el-table-column>
          <el-table-column label="年假(天)" min-width="90">
            <template #default="{ row }">
              <span class="metric-value">{{ row.annual_leave_days }}</span>
            </template>
          </el-table-column>
          <el-table-column label="病假(天)" min-width="90">
            <template #default="{ row }">
              <span class="metric-value" :class="row.sick_leave_days > 0 ? 'metric-value--danger' : ''">{{ row.sick_leave_days }}</span>
            </template>
          </el-table-column>
          <el-table-column label="事假(天)" min-width="90">
            <template #default="{ row }">
              <span class="metric-value" :class="row.personal_leave_days > 0 ? 'metric-value--warning' : ''">{{ row.personal_leave_days }}</span>
            </template>
          </el-table-column>
          <el-table-column label="异常" width="90" align="center">
            <template #default="{ row }">
              <el-tag v-if="row.is_anomaly" type="danger" size="small">异常</el-tag>
              <el-tag v-else type="success" size="small">正常</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="出勤率" min-width="100">
            <template #default="{ row }">
              <div class="rate-cell">
                <span class="rate-value" :class="getRateClass(row.attendance_rate)">
                  {{ row.attendance_rate.toFixed(1) }}%
                </span>
                <el-progress
                  :percentage="row.attendance_rate"
                  :stroke-width="4"
                  :color="getRateColor(row.attendance_rate)"
                  :show-text="false"
                  style="width: 60px"
                />
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="130" fixed="right">
            <template #default="{ row }">
              <el-button size="small" text type="primary" @click="viewDaily(row)">
                <el-icon><View /></el-icon>
                打卡详情
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-wrapper">
          <el-pagination
            v-model:current-page="page"
            :page-size="pageSize"
            :total="total"
            layout="total, prev, pager, next"
            @current-change="load"
          />
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="!loading && list.length === 0" class="empty-state glass-card">
        <div class="empty-icon">
          <el-icon><Calendar /></el-icon>
        </div>
        <h3>暂无出勤数据</h3>
        <p>当前月份暂无员工打卡记录</p>
      </div>
    </div>

    <!-- 格子视图 -->
    <div v-if="viewMode === 'grid'" v-loading="loading">
      <div class="grid-card glass-card">
        <el-table
          :data="list"
          border
          class="grid-table"
          :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
        >
          <el-table-column prop="employee_name" label="姓名" width="110" fixed="left" />
          <el-table-column
            v-for="day in gridDays"
            :key="day"
            :label="day + '日'"
            width="50"
            align="center"
          >
            <template #default="{ row }">
              <span class="grid-cell" :class="getGridCellClass(row, day)">
                {{ getGridCellSymbol(row, day) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="actual_days" label="实际出勤" width="90" align="center" fixed="right" />
          <el-table-column prop="required_days" label="应出勤" width="90" align="center" fixed="right" />
        </el-table>
      </div>
    </div>

    <!-- 打卡详情抽屉 -->
    <el-drawer
      v-model="drawerVisible"
      :title="`${drawerName} - ${selectedMonth} 打卡详情`"
      size="560px"
      class="daily-drawer"
    >
      <div v-if="dailyRecords.length === 0" class="drawer-empty">
        <el-icon><Tickets /></el-icon>
        <span>暂无打卡数据</span>
      </div>
      <div v-else class="daily-timeline">
        <div
          v-for="r in dailyRecords"
          :key="r.date"
          class="timeline-item"
          :class="{ 'is-holiday': r.is_holiday, 'is-weekend': r.is_weekend }"
        >
          <div class="timeline-date">
            <span class="date-day">{{ formatDay(r.date) }}</span>
            <span class="date-symbol" :class="getSymbolClass(r.symbol)">{{ r.symbol }}</span>
          </div>
          <div class="timeline-clock" v-if="r.clock_in">
            <div class="clock-item">
              <el-icon><Upload /></el-icon>
              <span>{{ r.clock_in }}</span>
            </div>
            <div class="clock-item">
              <el-icon><Download /></el-icon>
              <span>{{ r.clock_out || '--:--' }}</span>
            </div>
          </div>
          <div class="timeline-clock timeline-clock--absent" v-else>
            <span class="absent-text">未打卡</span>
          </div>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { attendanceApi, type MonthlyComplianceItem, type ComplianceMonthlyStats, type DailyRecord } from '@/api/attendance'
import { departmentApi } from '@/api/department'
import {
  Download, DataAnalysis, Grid, Calendar, Collection,
  Clock, WarningFilled, View, Tickets, Upload,
} from '@element-plus/icons-vue'

const loading = ref(false)
const viewMode = ref<'stat' | 'grid'>('stat')
const selectedMonth = ref(new Date().toISOString().slice(0, 7))
const selectedDepts = ref<number[]>([])
const deptOptions = ref<{ id: number; name: string }[]>([])
const list = ref<MonthlyComplianceItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const stats = ref<ComplianceMonthlyStats>({
  total_actual_days: 0,
  total_required_days: 0,
  total_overtime_hours: 0,
  total_absent_days: 0,
  total_anomaly_count: 0,
})
const drawerVisible = ref(false)
const drawerName = ref('')
const dailyRecords = ref<DailyRecord[]>([])

const weekdayNames = ['日', '一', '二', '三', '四', '五', '六']

const gridDays = computed(() => {
  if (!selectedMonth.value) return []
  const [, month] = selectedMonth.value.split('-')
  const days = new Date(parseInt(selectedMonth.value.split('-')[0]), parseInt(month), 0).getDate()
  return Array.from({ length: days }, (_, i) => i + 1)
})

function formatDay(dateStr: string) {
  const [, m, d] = dateStr.split('-')
  const date = new Date(dateStr)
  return `${parseInt(m)}月${parseInt(d)}日 周${weekdayNames[date.getDay()]}`
}

function getRateClass(rate: number) {
  if (rate >= 95) return 'rate--good'
  if (rate >= 80) return 'rate--warning'
  return 'rate--danger'
}

function getRateColor(rate: number) {
  if (rate >= 95) return '#10B981'
  if (rate >= 80) return '#F59E0B'
  return '#EF4444'
}

function getSymbolClass(symbol: string) {
  if (symbol === '√') return 'symbol--ok'
  if (symbol === '迟到') return 'symbol--warning'
  if (symbol === '缺') return 'symbol--danger'
  return 'symbol--none'
}

function getGridCellClass(_row: any, _day: number) { return '' }
function getGridCellSymbol(_row: any, _day: number) { return '' }

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
    if (data?.stats) stats.value = data.stats
  } catch { ElMessage.error('加载月报数据失败') }
  finally { loading.value = false }
}

async function viewDaily(row: MonthlyComplianceItem) {
  drawerName.value = row.employee_name || '员工'
  try {
    const { data } = await attendanceApi.getDailyRecords({ employee_id: row.employee_id, year_month: selectedMonth.value })
    dailyRecords.value = data?.records ?? []
  } catch { dailyRecords.value = [] }
  drawerVisible.value = true
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
    ElMessage.success('月报导出成功')
  } catch { ElMessage.error('导出失败') }
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
.header-actions { display: flex; align-items: center; gap: 12px; }
.month-picker { width: 150px; }
.dept-select { width: 200px; }

.view-toggle-bar {
  padding: 14px 20px;
  margin-bottom: 20px;
}

.toggle-group { display: inline-flex; background: #F3F4F6; border-radius: var(--radius-md); padding: 4px; }

.toggle-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
  background: transparent;

  &.active {
    background: #fff;
    color: var(--primary);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  &:hover:not(.active) { color: var(--text-primary); }
  .el-icon { font-size: 15px; }
}

.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 20px; }

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  transition: all 0.2s ease;
  &:hover { transform: translateY(-2px); box-shadow: var(--shadow-lg); }
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;

  &--actual { background: linear-gradient(135deg, #D1FAE5, #A7F3D0); color: var(--success); }
  &--required { background: linear-gradient(135deg, #DBEAFE, #BFDBFE); color: #3B82F6; }
  &--overtime { background: linear-gradient(135deg, #FEF3C7, #FDE68A); color: var(--warning); }
  &--absent { background: linear-gradient(135deg, #FEE2E2, #FECACA); color: var(--danger); }
}

.stat-body { display: flex; flex-direction: column; gap: 2px; }
.stat-value { font-size: 26px; font-weight: 700; color: var(--text-primary); line-height: 1.2;
  &--danger { color: var(--danger); }
  .stat-unit { font-size: 14px; font-weight: 500; margin-left: 2px; color: var(--text-secondary); }
}
.stat-label { font-size: 12px; color: var(--text-tertiary); }

.table-card { padding: 0; overflow: hidden; }

:deep(.modern-table) {
  .el-table__header th { padding: 14px 12px; font-size: 13px; }
  .el-table__row { transition: background 0.2s ease; &:hover > td { background: rgba(var(--primary), 0.02) !important; } }
  .el-table__cell { padding: 14px 12px; border-bottom: 1px solid #F3F4F6; }
}

.name-cell { display: flex; align-items: center; gap: 10px; }
.name-avatar { background: linear-gradient(135deg, var(--primary-light), var(--primary)); color: #fff; font-size: 13px; font-weight: 600; }
.name-text { font-weight: 500; color: var(--text-primary); }
.dept-text { font-size: 13px; color: var(--text-tertiary); }

.metric-value { font-weight: 600; font-size: 14px; font-family: 'SF Mono', Monaco, monospace;
  .metric-unit { font-size: 11px; font-weight: 500; margin-left: 1px; color: var(--text-tertiary); }
  &--muted { color: var(--text-tertiary); }
  &--warning { color: var(--warning); }
  &--danger { color: var(--danger); }
}

.rate-cell { display: flex; align-items: center; gap: 8px; }
.rate-value { font-weight: 700; font-size: 14px; min-width: 48px;
  &.rate--good { color: var(--success); }
  &.rate--warning { color: var(--warning); }
  &.rate--danger { color: var(--danger); }
}

.pagination-wrapper { display: flex; justify-content: flex-end; padding: 16px 20px; border-top: 1px solid var(--border); }

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
  }

  h3 { font-size: 18px; font-weight: 600; color: var(--text-primary); margin: 0 0 8px; }
  p { font-size: 14px; color: var(--text-tertiary); margin: 0; }
}

.grid-card { padding: 0; overflow-x: auto; }
:deep(.grid-table) { .el-table__cell { padding: 8px 4px; font-size: 12px; } }
.grid-cell { font-size: 12px; }

.daily-timeline { padding: 4px 0; }

.timeline-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid var(--border);
  transition: background 0.2s ease;

  &.is-holiday { background: #FEF2F2; margin: 0 -20px; padding: 12px 20px; border-radius: var(--radius-sm); }
  &.is-weekend { background: #F9FAFB; }
  &:last-child { border-bottom: none; }
}

.timeline-date { display: flex; flex-direction: column; gap: 2px; min-width: 90px; }
.date-day { font-size: 13px; font-weight: 500; color: var(--text-primary); }
.date-symbol { font-size: 12px; font-weight: 600;
  &.symbol--ok { color: var(--success); }
  &.symbol--warning { color: var(--warning); }
  &.symbol--danger { color: var(--danger); }
  &.symbol--none { color: var(--text-tertiary); }
}

.timeline-clock { display: flex; gap: 12px; flex: 1; }
.clock-item { display: flex; align-items: center; gap: 4px; font-size: 14px; font-weight: 600; color: var(--text-primary); font-family: 'SF Mono', Monaco, monospace; .el-icon { color: var(--text-tertiary); } }
.clock-item--absent .absent-text { color: var(--text-tertiary); font-size: 13px; }

.drawer-empty { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 80px; color: var(--text-tertiary); .el-icon { font-size: 40px; } }

@media (max-width: 1024px) { .stats-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 768px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
  .header-actions { flex-wrap: wrap; }
}
</style>

<style lang="scss">
// Non-scoped to apply to el-table shadow DOM rows
.anomaly-row td {
  background-color: rgba(239, 68, 68, 0.04) !important;
}
</style>
