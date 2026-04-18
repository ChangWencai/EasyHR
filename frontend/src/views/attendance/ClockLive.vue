<template>
  <div class="clock-live">
    <el-card>
      <template #header>
        <div class="header">
          <span>今日打卡实况</span>
          <div class="header-actions">
            <el-select v-model="selectedDepartment" placeholder="全部部门" clearable style="width: 140px" @change="load(1)">
              <el-option label="全部部门" value="" />
              <el-option v-for="dept in departments" :key="dept.id" :label="dept.name" :value="dept.id" />
            </el-select>
            <el-input v-model="search" placeholder="搜索员工姓名" clearable style="width: 180px" @input="debouncedLoad(1)" />
          </div>
        </div>
      </template>

      <el-table
        :data="filteredRecords"
        stripe
        v-loading="loading"
        @row-click="handleRowClick"
        :row-class-name="() => 'clickable-row'"
      >
        <el-table-column prop="employee_name" label="姓名" min-width="90" />
        <el-table-column prop="department_name" label="部门" min-width="100" />
        <el-table-column label="上班打卡" min-width="100">
          <template #default="{ row }">
            <span :style="{ color: getClockColor(row.clock_in_time, row.status, 'in') }">
              {{ row.clock_in_time || '--' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="下班打卡" min-width="100">
          <template #default="{ row }">
            <span :style="{ color: getClockColor(row.clock_out_time, row.status, 'out') }">
              {{ row.clock_out_time || '--' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="班次" min-width="80">
          <template #default="{ row }">
            {{ row.shift_name || '默认班次' }}
          </template>
        </el-table-column>
        <el-table-column label="打卡状态" width="110">
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
                <el-button size="small" type="primary" link>邀请点签</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <!-- 假勤统计 Popover -->
      <el-popover
        v-model:visible="popoverVisible"
        :width="320"
        trigger="click"
        placement="top"
        :title="`${selectedRecord?.employee_name || ''} - 假勤统计`"
      >
        <div class="leave-stats">
          <div class="stat-row">
            <span class="stat-label">请假:</span>
            <span class="stat-value">{{ leaveStats.leave_days }} 天</span>
            <span class="stat-detail">（已通过 {{ leaveStats.approved_days }} 天，待审批 {{ leaveStats.pending_days }} 天）</span>
          </div>
          <div class="stat-row">
            <span class="stat-label">出差:</span>
            <span class="stat-value">{{ leaveStats.business_days }} 天</span>
          </div>
          <div class="stat-row">
            <span class="stat-label">外出:</span>
            <span class="stat-value">{{ leaveStats.outside_days }} 天</span>
          </div>
          <div class="stat-row">
            <span class="stat-label">加班:</span>
            <span class="stat-value">{{ leaveStats.overtime_hours }} 小时</span>
          </div>
          <div class="stat-row">
            <span class="stat-label">补卡:</span>
            <span class="stat-value">{{ leaveStats.makeup_count }} 次</span>
          </div>
          <div class="stat-row">
            <span class="stat-label">调班:</span>
            <span class="stat-value">{{ leaveStats.shift_swap_count }} 次</span>
          </div>
        </div>
        <div class="popover-actions">
          <el-button size="small" @click="openEditDialog">手动修正</el-button>
        </div>
        <template #reference>
          <span />
        </template>
      </el-popover>

      <el-pagination
        class="mt-4"
        layout="total,prev,pager,next"
        :total="total"
        :page="page"
        :page-size="pageSize"
        @current-change="load"
      />
    </el-card>

    <!-- 手动修正弹窗 -->
    <el-dialog v-model="editDialogVisible" title="手动修正假勤数据" width="480px">
      <el-form label-width="120px">
        <el-form-item label="请假天数">
          <el-input-number v-model="editForm.leave_days" :min="0" :step="0.5" style="width: 100%" />
        </el-form-item>
        <el-form-item label="出差天数">
          <el-input-number v-model="editForm.business_days" :min="0" :step="0.5" style="width: 100%" />
        </el-form-item>
        <el-form-item label="外出天数">
          <el-input-number v-model="editForm.outside_days" :min="0" :step="0.5" style="width: 100%" />
        </el-form-item>
        <el-form-item label="加班时长（小时）">
          <el-input-number v-model="editForm.overtime_hours" :min="0" :step="0.5" style="width: 100%" />
        </el-form-item>
        <el-form-item label="补卡次数">
          <el-input-number v-model="editForm.makeup_count" :min="0" :step="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="调班次数">
          <el-input-number v-model="editForm.shift_swap_count" :min="0" :step="1" style="width: 100%" />
        </el-form-item>
      </el-form>
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

const loading = ref(false)
const records = ref<ClockLiveRecord[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const search = ref('')
const selectedDepartment = ref<number | null>(null)
const departments = ref<{ id: number; name: string }[]>([])

const popoverVisible = ref(false)
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

const editDialogVisible = ref(false)
const saving = ref(false)
const editForm = ref({
  leave_days: 0,
  business_days: 0,
  outside_days: 0,
  overtime_hours: 0,
  makeup_count: 0,
  shift_swap_count: 0,
})

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
  popoverVisible.value = true
}

function getClockColor(clockTime: string, status: string, type: 'in' | 'out'): string {
  if (!clockTime || clockTime === '--') {
    return '#BFBFBF'
  }
  if (status === 'late' && type === 'in') return '#FA8C16'
  if (status === 'absent') return '#FF4D4F'
  return '#52C41A'
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

function openEditDialog() {
  popoverVisible.value = false
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
.clock-live {
  padding: 20px 24px;
  width: 100%;
  box-sizing: border-box;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.mt-4 {
  margin-top: 16px;
}

.clickable-row {
  cursor: pointer;
}

.leave-stats {
  .stat-row {
    display: flex;
    align-items: baseline;
    gap: 8px;
    padding: 4px 0;
  }
  .stat-label {
    min-width: 50px;
    font-weight: 600;
    color: #1a1a1a;
  }
  .stat-value {
    font-size: 16px;
    font-weight: 700;
    color: #1677ff;
  }
  .stat-detail {
    font-size: 12px;
    color: #8c8c8c;
  }
}

.popover-actions {
  margin-top: 12px;
  text-align: right;
  border-top: 1px solid #f0f0f0;
  padding-top: 8px;
}
</style>
