<template>
  <div class="clock-live">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">今日打卡实况</h1>
        <p class="page-subtitle">{{ currentDate }}</p>
      </div>
      <div class="header-actions">
        <el-select
          v-model="selectedDepartment"
          placeholder="全部部门"
          clearable
          class="filter-select"
          @change="load(1)"
        >
          <template #prefix>
            <el-icon><OfficeBuilding /></el-icon>
          </template>
          <el-option v-for="dept in departments" :key="dept.id" :label="dept.name" :value="dept.id" />
        </el-select>
        <div class="search-wrapper">
          <el-icon class="search-icon"><Search /></el-icon>
          <input
            v-model="search"
            type="text"
            placeholder="搜索员工姓名..."
            class="search-input"
            @input="debouncedLoad(1)"
          />
        </div>
      </div>
    </header>

    <!-- 统计概览 -->
    <div class="stats-overview">
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--total">
          <el-icon><User /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ total }}</span>
          <span class="stat-label">应打卡人数</span>
        </div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--checked">
          <el-icon><CircleCheck /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value stat-value--success">{{ checkedCount }}</span>
          <span class="stat-label">已打卡</span>
        </div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--late">
          <el-icon><Clock /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value stat-value--warning">{{ lateCount }}</span>
          <span class="stat-label">迟到</span>
        </div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--absent">
          <el-icon><RemoveFilled /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value stat-value--danger">{{ absentCount }}</span>
          <span class="stat-label">缺卡</span>
        </div>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="table-container glass-card">
      <el-table
        :data="filteredRecords"
        stripe
        v-loading="loading"
        @row-click="handleRowClick"
        class="modern-table"
        :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
      >
        <el-table-column prop="employee_name" label="员工" min-width="120" fixed="left">
          <template #default="{ row }">
            <div class="employee-cell">
              <el-avatar :size="36" class="employee-avatar">
                {{ row.employee_name?.[0] || '?' }}
              </el-avatar>
              <div class="employee-info">
                <span class="employee-name">{{ row.employee_name }}</span>
                <span class="employee-dept">{{ row.department_name || '未分配部门' }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="上班打卡" min-width="110">
          <template #default="{ row }">
            <div class="clock-time" :class="getClockClass(row.clock_in_time, row.status, 'in')">
              <el-icon v-if="row.clock_in_time">
                <Check v-if="row.status !== 'late'" />
                <Warning v-else />
              </el-icon>
              <span>{{ row.clock_in_time || '--:--' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="下班打卡" min-width="110">
          <template #default="{ row }">
            <div class="clock-time" :class="getClockClass(row.clock_out_time, row.status, 'out')">
              <el-icon v-if="row.clock_out_time">
                <Check />
              </el-icon>
              <span>{{ row.clock_out_time || '--:--' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="班次" min-width="100">
          <template #default="{ row }">
            <span class="shift-tag">{{ row.shift_name || '默认班次' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="打卡状态" width="120">
          <template #default="{ row }">
            <ClockStatusTag :status="row.status" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-popconfirm
              :title="`确认邀请 ${row.employee_name || '该员工'} 点签打卡？`"
              @confirm="handleInviteClock(row)"
            >
              <template #reference>
                <el-button size="small" type="primary" class="invite-btn">
                  <el-icon><Bell /></el-icon>
                  邀请
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
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

    <!-- 假勤统计弹窗 -->
    <el-dialog
      v-model="editDialogVisible"
      :title="`${selectedRecord?.employee_name || ''} - 假勤统计`"
      width="480px"
      class="stats-dialog"
    >
      <div class="leave-stats-grid">
        <div class="stat-item">
          <div class="stat-item-icon stat-item-icon--leave">
            <el-icon><Calendar /></el-icon>
          </div>
          <div class="stat-item-info">
            <span class="stat-item-value">{{ leaveStats.leave_days }}</span>
            <span class="stat-item-label">请假（天）</span>
            <span class="stat-item-detail">
              已通过 {{ leaveStats.approved_days }} · 待审批 {{ leaveStats.pending_days }}
            </span>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-item-icon stat-item-icon--business">
            <el-icon><Location /></el-icon>
          </div>
          <div class="stat-item-info">
            <span class="stat-item-value">{{ leaveStats.business_days }}</span>
            <span class="stat-item-label">出差（天）</span>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-item-icon stat-item-icon--outside">
            <el-icon><Position /></el-icon>
          </div>
          <div class="stat-item-info">
            <span class="stat-item-value">{{ leaveStats.outside_days }}</span>
            <span class="stat-item-label">外出（天）</span>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-item-icon stat-item-icon--overtime">
            <el-icon><Timer /></el-icon>
          </div>
          <div class="stat-item-info">
            <span class="stat-item-value">{{ leaveStats.overtime_hours }}</span>
            <span class="stat-item-label">加班（小时）</span>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-item-icon stat-item-icon--makeup">
            <el-icon><Edit /></el-icon>
          </div>
          <div class="stat-item-info">
            <span class="stat-item-value">{{ leaveStats.makeup_count }}</span>
            <span class="stat-item-label">补卡（次）</span>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-item-icon stat-item-icon--shift">
            <el-icon><Switch /></el-icon>
          </div>
          <div class="stat-item-info">
            <span class="stat-item-value">{{ leaveStats.shift_swap_count }}</span>
            <span class="stat-item-label">调班（次）</span>
          </div>
        </div>
      </div>

      <!-- 编辑表单 -->
      <div class="edit-form">
        <h4 class="edit-form-title">手动修正</h4>
        <div class="edit-form-grid">
          <el-form-item label="请假天数">
            <el-input-number v-model="editForm.leave_days" :min="0" :step="0.5" style="width: 100%" size="large" />
          </el-form-item>
          <el-form-item label="出差天数">
            <el-input-number v-model="editForm.business_days" :min="0" :step="0.5" style="width: 100%" size="large" />
          </el-form-item>
          <el-form-item label="外出天数">
            <el-input-number v-model="editForm.outside_days" :min="0" :step="0.5" style="width: 100%" size="large" />
          </el-form-item>
          <el-form-item label="加班（小时）">
            <el-input-number v-model="editForm.overtime_hours" :min="0" :step="0.5" style="width: 100%" size="large" />
          </el-form-item>
          <el-form-item label="补卡次数">
            <el-input-number v-model="editForm.makeup_count" :min="0" :step="1" style="width: 100%" size="large" />
          </el-form-item>
          <el-form-item label="调班次数">
            <el-input-number v-model="editForm.shift_swap_count" :min="0" :step="1" style="width: 100%" size="large" />
          </el-form-item>
        </div>
      </div>

      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSaveEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { attendanceApi, type ClockLiveRecord, type LeaveStats } from '@/api/attendance'
import { ElMessage } from 'element-plus'
import ClockStatusTag from '@/components/attendance/ClockStatusTag.vue'
import { useDebounceFn } from '@vueuse/core'
import {
  Search,
  OfficeBuilding,
  User,
  CircleCheck,
  Clock,
  RemoveFilled,
  Check,
  Warning,
  Bell,
  Calendar,
  Location,
  Position,
  Timer,
  Edit,
  Switch,
} from '@element-plus/icons-vue'

const loading = ref(false)
const records = ref<ClockLiveRecord[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const search = ref('')
const selectedDepartment = ref<number | null>(null)
const departments = ref<{ id: number; name: string }[]>([])

const editDialogVisible = ref(false)
const saving = ref(false)
const selectedRecord = ref<ClockLiveRecord | null>(null)
const leaveStats = ref<LeaveStats>({
  employee_id: 0,
  employee_name: '',
  year_month: '',
  leave_days: 0,
  business_days: 0,
  outside_days: 0,
  makeup_count: 0,
  shift_swap_count: 0,
  overtime_hours: 0,
  pending_days: 0,
  approved_days: 0,
})

const editForm = ref({
  leave_days: 0,
  business_days: 0,
  outside_days: 0,
  overtime_hours: 0,
  makeup_count: 0,
  shift_swap_count: 0,
})

const currentDate = computed(() => {
  const now = new Date()
  return `${now.getMonth() + 1}月${now.getDate()}日 ${['周日', '周一', '周二', '周三', '周四', '周五', '周六'][now.getDay()]}`
})

const checkedCount = computed(() =>
  records.value.filter(r => r.clock_in_time && r.clock_out_time).length
)

const lateCount = computed(() =>
  records.value.filter(r => r.status === 'late').length
)

const absentCount = computed(() =>
  records.value.filter(r => r.status === 'absent' || (!r.clock_in_time && !r.clock_out_time)).length
)

const filteredRecords = computed(() => {
  if (!search.value) return records.value
  const kw = search.value.toLowerCase()
  return records.value.filter(r =>
    (r.employee_name || '').toLowerCase().includes(kw)
  )
})

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const today = new Date().toISOString().split('T')[0]
    const res = await attendanceApi.getClockLive({
      date: today,
      page: p,
      page_size: pageSize.value,
    })
    records.value = res.data?.records || []
    total.value = res.data?.total || 0
  } catch {
    ElMessage.error('加载打卡数据失败，请刷新页面重试')
  } finally {
    loading.value = false
  }
}

const debouncedLoad = useDebounceFn((p: number) => load(p), 300)

async function handleRowClick(row: ClockLiveRecord) {
  selectedRecord.value = row
  const yearMonth = new Date().toISOString().slice(0, 7)
  try {
    const res = await attendanceApi.getLeaveStats({
      employee_id: row.employee_id,
      year_month: yearMonth,
    })
    leaveStats.value = res.data || leaveStats.value
  } catch {
    // ignore
  }
  editForm.value = {
    leave_days: leaveStats.value.leave_days || 0,
    business_days: leaveStats.value.business_days || 0,
    outside_days: leaveStats.value.outside_days || 0,
    overtime_hours: leaveStats.value.overtime_hours || 0,
    makeup_count: leaveStats.value.makeup_count || 0,
    shift_swap_count: leaveStats.value.shift_swap_count || 0,
  }
  editDialogVisible.value = true
}

function getClockClass(clockTime: string, status: string, type: 'in' | 'out'): string {
  if (!clockTime || clockTime === '--') return 'clock--empty'
  if (status === 'late' && type === 'in') return 'clock--late'
  if (status === 'absent') return 'clock--absent'
  return 'clock--normal'
}

async function handleInviteClock(row: ClockLiveRecord) {
  try {
    const today = new Date().toISOString().split('T')[0]
    const now = new Date()
    const hh = String(now.getHours()).padStart(2, '0')
    const mm = String(now.getMinutes()).padStart(2, '0')
    const ss = String(now.getSeconds()).padStart(2, '0')
    const clockTime = `${today}T${hh}:${mm}:${ss}Z`
    await attendanceApi.createClockRecord({
      employee_id: row.employee_id,
      clock_time: clockTime,
      clock_type: 'in',
    })
    ElMessage.success(`已向 ${row.employee_name || '员工'} 发送点签邀请`)
    await load()
  } catch {
    ElMessage.error('发送邀请失败，请稍后重试')
  }
}

async function handleSaveEdit() {
  if (!selectedRecord.value) return
  saving.value = true
  try {
    const yearMonth = new Date().toISOString().slice(0, 7)
    await attendanceApi.updateLeaveStats(selectedRecord.value.employee_id, yearMonth, editForm.value)
    ElMessage.success('假勤数据已更新')
    editDialogVisible.value = false
  } catch {
    ElMessage.error('更新失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => load())
</script>

<style scoped lang="scss">
// ============================================================
// 变量定义
// ============================================================
$success: #10B981;
$warning: #F59E0B;
$error: #EF4444;
$bg-page: #FAFBFC;
$bg-surface: #FFFFFF;
$text-primary: #1F2937;
$text-secondary: #6B7280;
$text-muted: #9CA3AF;
$border-color: #E5E7EB;
$shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
$shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
$shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
$radius-sm: 8px;
$radius-md: 12px;
$radius-lg: 16px;
$radius-xl: 24px;

// ============================================================
// 页面布局
// ============================================================
.clock-live {
  padding: 24px 32px;
  width: 100%;
  box-sizing: border-box;
  background: $bg-page;
  min-height: 100vh;
}

// ============================================================
// 玻璃态卡片
// ============================================================
.glass-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: $radius-xl;
  box-shadow: $shadow-md;
}

// ============================================================
// 页面头部
// ============================================================
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.header-left {
  .page-title {
    font-size: 24px;
    font-weight: 700;
    color: $text-primary;
    margin: 0 0 4px;
  }

  .page-subtitle {
    font-size: 14px;
    color: $text-secondary;
    margin: 0;
  }
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-select {
  width: 160px;
}

.search-wrapper {
  position: relative;
  width: 240px;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: $text-muted;
  font-size: 16px;
}

.search-input {
  width: 100%;
  padding: 10px 12px 10px 38px;
  font-size: 14px;
  color: $text-primary;
  background: $bg-surface;
  border: 1px solid $border-color;
  border-radius: $radius-md;
  outline: none;
  transition: all 0.2s ease;

  &::placeholder {
    color: $text-muted;
  }

  &:focus {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(var(--primary), 0.1);
  }
}

// ============================================================
// 统计概览
// ============================================================
.stats-overview {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  transition: all 0.2s ease;

  &:hover {
    transform: translateY(-2px);
    box-shadow: $shadow-lg;
  }
}

.stat-icon {
  width: 52px;
  height: 52px;
  border-radius: $radius-lg;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;

  &--total {
    background: linear-gradient(135deg, #EDE9FE 0%, #DDD6FE 100%);
    color: var(--primary);
  }

  &--checked {
    background: linear-gradient(135deg, #D1FAE5 0%, #A7F3D0 100%);
    color: $success;
  }

  &--late {
    background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%);
    color: $warning;
  }

  &--absent {
    background: linear-gradient(135deg, #FEE2E2 0%, #FECACA 100%);
    color: $error;
  }
}

.stat-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: $text-primary;
  line-height: 1;

  &--success { color: $success; }
  &--warning { color: $warning; }
  &--danger { color: $error; }
}

.stat-label {
  font-size: 13px;
  color: $text-secondary;
}

// ============================================================
// 数据表格
// ============================================================
.table-container {
  padding: 0;
  overflow: hidden;
}

:deep(.modern-table) {
  .el-table__header th {
    padding: 16px 12px;
    font-size: 13px;
  }

  .el-table__row {
    cursor: pointer;
    transition: background 0.2s ease;

    &:hover > td {
      background: rgba(var(--primary), 0.02) !important;
    }
  }

  .el-table__cell {
    padding: 16px 12px;
    border-bottom: 1px solid #F3F4F6;
  }
}

.employee-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.employee-avatar {
  background: linear-gradient(135deg, var(--primary-light) 0%, var(--primary) 100%);
  color: #fff;
  font-size: 14px;
  font-weight: 600;
}

.employee-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.employee-name {
  font-weight: 500;
  color: $text-primary;
}

.employee-dept {
  font-size: 12px;
  color: $text-muted;
}

.clock-time {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 500;
  padding: 4px 10px;
  border-radius: 6px;

  .el-icon {
    font-size: 14px;
  }

  &.clock--normal {
    background: #D1FAE5;
    color: #059669;
  }

  &.clock--late {
    background: #FEF3C7;
    color: #D97706;
  }

  &.clock--absent {
    background: #FEE2E2;
    color: #DC2626;
  }

  &.clock--empty {
    color: $text-muted;
  }
}

.shift-tag {
  font-size: 12px;
  color: $text-secondary;
  background: $bg-page;
  padding: 4px 10px;
  border-radius: 6px;
}

.invite-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border-radius: $radius-sm;

  .el-icon {
    font-size: 14px;
  }
}

// ============================================================
// 分页
// ============================================================
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: 16px 20px;
  border-top: 1px solid $border-color;
}

// ============================================================
// 假勤统计弹窗
// ============================================================
.leave-stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 24px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: $bg-page;
  border-radius: $radius-md;
}

.stat-item-icon {
  width: 40px;
  height: 40px;
  border-radius: $radius-sm;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;

  &--leave { background: #DBEAFE; color: #2563EB; }
  &--business { background: #FEF3C7; color: #D97706; }
  &--outside { background: #CFFAFE; color: #0891B2; }
  &--overtime { background: #EDE9FE; color: #7C3AED; }
  &--makeup { background: #FCE7F3; color: #DB2777; }
  &--shift { background: #D1FAE5; color: #059669; }
}

.stat-item-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stat-item-value {
  font-size: 20px;
  font-weight: 700;
  color: $text-primary;
  line-height: 1;
}

.stat-item-label {
  font-size: 12px;
  color: $text-secondary;
}

.stat-item-detail {
  font-size: 11px;
  color: $text-muted;
}

.edit-form {
  border-top: 1px solid $border-color;
  padding-top: 20px;
}

.edit-form-title {
  font-size: 14px;
  font-weight: 600;
  color: $text-primary;
  margin: 0 0 16px;
}

.edit-form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;

  :deep(.el-form-item) {
    margin-bottom: 0;
  }

  :deep(.el-form-item__label) {
    font-size: 12px;
    color: $text-secondary;
    padding-bottom: 6px;
  }
}

// ============================================================
// 响应式
// ============================================================
@media (max-width: 1200px) {
  .stats-overview {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .clock-live {
    padding: 16px;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .header-actions {
    width: 100%;
    flex-wrap: wrap;
  }

  .search-wrapper {
    width: 100%;
  }

  .stats-overview {
    grid-template-columns: 1fr;
  }

  .leave-stats-grid,
  .edit-form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
