<template>
  <div class="attendance-rule">
    <el-card>
      <template #header>
        <div class="header">
          <span>打卡规则设置</span>
        </div>
      </template>

      <el-tabs v-model="activeTab" class="rule-tabs">
        <!-- 固定时间 Tab -->
        <el-tab-pane label="固定时间" name="fixed">
          <el-form label-width="120px" class="rule-form">
            <el-form-item label="上班日">
              <el-checkbox-group v-model="fixedWorkDays">
                <el-checkbox v-for="day in weekDays" :key="day.value" :value="day.value">
                  {{ day.label }}
                </el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item label="上班打卡时间">
              <el-time-picker v-model="fixedWorkStart" format="HH:mm" value-format="HH:mm" placeholder="选择时间" style="width: 150px" />
            </el-form-item>
            <el-form-item label="下班打卡时间">
              <el-time-picker v-model="fixedWorkEnd" format="HH:mm" value-format="HH:mm" placeholder="选择时间" style="width: 150px" />
            </el-form-item>
            <el-form-item label="打卡位置">
              <el-input v-model="fixedLocation" placeholder="请输入打卡位置（如：公司地址）" style="width: 300px" />
            </el-form-item>
            <el-form-item label="打卡方式">
              <el-radio-group v-model="fixedClockMethod">
                <el-radio value="click">手动打卡</el-radio>
                <el-radio value="photo">拍照打卡</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="节假日设置">
              <el-table :data="fixedHolidays" style="width: 500px" size="small">
                <el-table-column prop="date" label="日期" width="160">
                  <template #default="{ row }">
                    <el-date-picker
                      v-model="row.date"
                      type="date"
                      format="YYYY-MM-DD"
                      value-format="YYYY-MM-DD"
                      placeholder="选择日期"
                      style="width: 140px"
                    />
                  </template>
                </el-table-column>
                <el-table-column prop="name" label="名称" width="180">
                  <template #default="{ row }">
                    <el-input v-model="row.name" placeholder="节假日名称（如：元旦）" />
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="80">
                  <template #default="{ $index }">
                    <el-popconfirm title="确认删除？" @confirm="fixedHolidays.splice($index, 1)">
                      <template #reference>
                        <el-button size="small" type="danger" link>删除</el-button>
                      </template>
                    </el-popconfirm>
                  </template>
                </el-table-column>
              </el-table>
              <el-button size="small" @click="fixedHolidays.push({ date: '', name: '' })" style="margin-top: 8px">
                + 添加节假日
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 按排班 Tab -->
        <el-tab-pane label="按排班" name="scheduled">
          <el-form label-width="120px" class="rule-form">
            <el-form-item label="打卡方式">
              <el-radio-group v-model="scheduledClockMethod">
                <el-radio value="click">手动打卡</el-radio>
                <el-radio value="photo">拍照打卡</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-divider content-position="left">班次管理</el-divider>
            <div class="shift-list">
              <div v-for="shift in shifts" :key="shift.id" class="shift-item">
                <span class="shift-name">{{ shift.name }}</span>
                <span>{{ shift.work_start }} - {{ shift.work_end }}</span>
                <el-tag v-if="shift.work_date_offset !== 0" size="small" type="warning">跨天</el-tag>
                <el-button size="small" type="primary" link @click="openShiftDialog(shift)">编辑</el-button>
                <el-popconfirm title="确认删除？" @confirm="handleDeleteShift(shift.id)">
                  <template #reference>
                    <el-button size="small" type="danger" link>删除</el-button>
                  </template>
                </el-popconfirm>
              </div>
            </div>
            <el-button type="primary" @click="openShiftDialog(null)">+ 新建班次</el-button>
          </el-form>
        </el-tab-pane>

        <!-- 自由工时 Tab -->
        <el-tab-pane label="自由工时" name="free">
          <el-form label-width="120px" class="rule-form">
            <el-form-item label="上班日">
              <el-checkbox-group v-model="freeWorkDays">
                <el-checkbox v-for="day in weekDays" :key="day.value" :value="day.value">
                  {{ day.label }}
                </el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item label="打卡位置">
              <el-input v-model="freeLocation" placeholder="请输入打卡位置" style="width: 300px" />
            </el-form-item>
            <el-form-item label="打卡方式">
              <el-radio-group v-model="freeClockMethod">
                <el-radio value="click">手动打卡</el-radio>
                <el-radio value="photo">拍照打卡</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="节假日设置">
              <el-table :data="freeHolidays" style="width: 500px" size="small">
                <el-table-column prop="date" label="日期" width="160">
                  <template #default="{ row }">
                    <el-date-picker
                      v-model="row.date"
                      type="date"
                      format="YYYY-MM-DD"
                      value-format="YYYY-MM-DD"
                      placeholder="选择日期"
                      style="width: 140px"
                    />
                  </template>
                </el-table-column>
                <el-table-column prop="name" label="名称" width="180">
                  <template #default="{ row }">
                    <el-input v-model="row.name" placeholder="节假日名称（如：元旦）" />
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="80">
                  <template #default="{ $index }">
                    <el-popconfirm title="确认删除？" @confirm="freeHolidays.splice($index, 1)">
                      <template #reference>
                        <el-button size="small" type="danger" link>删除</el-button>
                      </template>
                    </el-popconfirm>
                  </template>
                </el-table-column>
              </el-table>
              <el-button size="small" @click="freeHolidays.push({ date: '', name: '' })" style="margin-top: 8px">
                + 添加节假日
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>

      <div class="footer-action">
        <el-button type="primary" :loading="saving" size="large" @click="handleSave">
          保存规则
        </el-button>
      </div>
    </el-card>

    <!-- 班次编辑弹窗 -->
    <el-dialog v-model="shiftDialogVisible" :title="editingShift ? '编辑班次' : '新建班次'" width="400px">
      <el-form label-width="100px">
        <el-form-item label="班次名称">
          <el-input v-model="shiftForm.name" placeholder="如：早班/晚班" />
        </el-form-item>
        <el-form-item label="上班时间">
          <el-time-picker v-model="shiftForm.work_start" format="HH:mm" value-format="HH:mm" style="width: 100%" />
        </el-form-item>
        <el-form-item label="下班时间">
          <el-time-picker v-model="shiftForm.work_end" format="HH:mm" value-format="HH:mm" style="width: 100%" />
        </el-form-item>
        <el-form-item label="跨天班次">
          <el-switch v-model="shiftForm.is_cross_day" />
          <span style="margin-left: 8px; color: #8c8c8c; font-size: 12px">开启后下班时间归属次日</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="shiftDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="shiftSaving" @click="handleSaveShift">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { attendanceApi } from '@/api/attendance'
import type { Shift } from '@/api/attendance'
import { ElMessage } from 'element-plus'

const activeTab = ref<'fixed' | 'scheduled' | 'free'>('fixed')

const weekDays = [
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
  { label: '周六', value: 6 },
  { label: '周日', value: 0 },
]

// 固定时间
const fixedWorkDays = ref<number[]>([1, 2, 3, 4, 5])
const fixedWorkStart = ref('09:00')
const fixedWorkEnd = ref('18:00')
const fixedLocation = ref('')
const fixedClockMethod = ref<'click' | 'photo'>('click')
const fixedHolidays = ref<{ date: string; name: string }[]>([])

// 按排班
const scheduledClockMethod = ref<'click' | 'photo'>('click')
const shifts = ref<Shift[]>([])

// 自由工时
const freeWorkDays = ref<number[]>([1, 2, 3, 4, 5])
const freeLocation = ref('')
const freeClockMethod = ref<'click' | 'photo'>('click')
const freeHolidays = ref<{ date: string; name: string }[]>([])

const saving = ref(false)

// 班次弹窗
const shiftDialogVisible = ref(false)
const editingShift = ref<Shift | null>(null)
const shiftSaving = ref(false)
const shiftForm = ref({
  name: '',
  work_start: '',
  work_end: '',
  is_cross_day: false,
})

async function loadRule() {
  try {
    const res = await attendanceApi.getRule() as any
    if (res.data) {
      const rule = res.data
      if (rule.mode === 'fixed') {
        activeTab.value = 'fixed'
        fixedWorkDays.value = rule.work_days || []
        fixedWorkStart.value = rule.work_start || ''
        fixedWorkEnd.value = rule.work_end || ''
        fixedLocation.value = rule.location || ''
        fixedClockMethod.value = rule.clock_method || 'click'
        fixedHolidays.value = rule.holidays || []
      } else if (rule.mode === 'free') {
        activeTab.value = 'free'
        freeWorkDays.value = rule.work_days || []
        freeLocation.value = rule.location || ''
        freeClockMethod.value = rule.clock_method || 'click'
        freeHolidays.value = rule.holidays || []
      } else if (rule.mode === 'scheduled') {
        activeTab.value = 'scheduled'
        scheduledClockMethod.value = rule.clock_method || 'click'
      }
    }
  } catch {
    // 无规则，正常
  }
}

async function loadShifts() {
  try {
    const res = await attendanceApi.listShifts() as any
    shifts.value = res.data || []
  } catch {
    // ignore
  }
}

async function handleSave() {
  saving.value = true
  try {
    let data: Record<string, unknown> = {}
    if (activeTab.value === 'fixed') {
      data = {
        mode: 'fixed',
        work_days: fixedWorkDays.value,
        work_start: fixedWorkStart.value,
        work_end: fixedWorkEnd.value,
        location: fixedLocation.value,
        clock_method: fixedClockMethod.value,
        holidays: fixedHolidays.value.filter(h => h.date),
      }
    } else if (activeTab.value === 'scheduled') {
      data = {
        mode: 'scheduled',
        clock_method: scheduledClockMethod.value,
        holidays: [],
      }
    } else {
      data = {
        mode: 'free',
        work_days: freeWorkDays.value,
        location: freeLocation.value,
        clock_method: freeClockMethod.value,
        holidays: freeHolidays.value.filter(h => h.date),
      }
    }
    await attendanceApi.saveRule(data as any)
    ElMessage.success('打卡规则已保存')
  } catch {
    ElMessage.error('保存失败，请稍后重试')
  } finally {
    saving.value = false
  }
}

function openShiftDialog(shift: Shift | null) {
  editingShift.value = shift
  if (shift) {
    shiftForm.value = {
      name: shift.name,
      work_start: shift.work_start,
      work_end: shift.work_end,
      is_cross_day: shift.work_date_offset !== 0,
    }
  } else {
    shiftForm.value = { name: '', work_start: '', work_end: '', is_cross_day: false }
  }
  shiftDialogVisible.value = true
}

async function handleSaveShift() {
  shiftSaving.value = true
  try {
    const payload = {
      name: shiftForm.value.name,
      work_start: shiftForm.value.work_start,
      work_end: shiftForm.value.work_end,
      work_date_offset: shiftForm.value.is_cross_day ? 1 : 0,
    }
    if (editingShift.value) {
      await attendanceApi.updateShift(editingShift.value.id, payload)
    } else {
      await attendanceApi.createShift(payload)
    }
    ElMessage.success('班次已保存')
    shiftDialogVisible.value = false
    loadShifts()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    shiftSaving.value = false
  }
}

async function handleDeleteShift(id: number) {
  try {
    await attendanceApi.deleteShift(id)
    ElMessage.success('已删除')
    loadShifts()
  } catch {
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadRule()
  loadShifts()
})
</script>

<style scoped lang="scss">
.attendance-rule {
  padding: 20px 24px;
  width: 100%;
  box-sizing: border-box;
}
.header {
  font-size: 16px;
  font-weight: 700;
  color: #1a1a1a;
}
.rule-tabs {
  margin-bottom: 80px;
}
.rule-form {
  padding: 16px 0;
  max-width: 700px;
}
.shift-list {
  margin-bottom: 16px;
}
.shift-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
}
.shift-name {
  font-weight: 600;
  min-width: 80px;
}
.footer-action {
  position: fixed;
  bottom: 0;
  left: 220px;
  right: 0;
  padding: 16px 24px;
  background: #fff;
  border-top: 1px solid #f0f0f0;
  display: flex;
  justify-content: flex-end;
  z-index: 100;
}
</style>
