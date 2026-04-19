<template>
  <div class="page-view">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">打卡规则设置</h1>
        <p class="page-subtitle">配置考勤打卡方式与时间规则</p>
      </div>
    </header>

    <!-- 规则卡片 -->
    <div class="rule-card glass-card">
      <div class="rule-card-header">
        <div class="mode-tabs">
          <button
            v-for="tab in modeTabs"
            :key="tab.value"
            class="mode-tab"
            :class="{ active: activeTab === tab.value }"
            @click="activeTab = tab.value as any"
          >
            <el-icon><component :is="tab.icon" /></el-icon>
            <span>{{ tab.label }}</span>
          </button>
        </div>
      </div>

      <!-- 固定时间 -->
      <div v-show="activeTab === 'fixed'" class="rule-content">
        <div class="form-section">
          <div class="section-label">
            <el-icon><Calendar /></el-icon>
            <span>上班日设置</span>
          </div>
          <div class="work-days-grid">
            <label
              v-for="day in weekDays"
              :key="day.value"
              class="day-chip"
              :class="{ selected: fixedWorkDays.includes(day.value) }"
            >
              <input type="checkbox" :value="day.value" v-model="fixedWorkDays" class="hidden-check" />
              <span>{{ day.label }}</span>
            </label>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label">上班打卡时间</label>
            <el-time-picker
              v-model="fixedWorkStart"
              format="HH:mm"
              value-format="HH:mm"
              placeholder="选择时间"
              size="large"
              style="width: 100%"
            />
          </div>
          <div class="form-group">
            <label class="form-label">下班打卡时间</label>
            <el-time-picker
              v-model="fixedWorkEnd"
              format="HH:mm"
              value-format="HH:mm"
              placeholder="选择时间"
              size="large"
              style="width: 100%"
            />
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">打卡位置</label>
          <el-input
            v-model="fixedLocation"
            placeholder="请输入打卡位置（如：公司地址）"
            size="large"
          >
            <template #prefix>
              <el-icon><Location /></el-icon>
            </template>
          </el-input>
        </div>

        <div class="form-group">
          <label class="form-label">打卡方式</label>
          <div class="method-cards">
            <label
              class="method-card"
              :class="{ selected: fixedClockMethod === 'click' }"
            >
              <input type="radio" value="click" v-model="fixedClockMethod" class="hidden-check" />
              <div class="method-icon">
                <el-icon><Pointer /></el-icon>
              </div>
              <span class="method-name">手动打卡</span>
              <span class="method-desc">员工主动签到</span>
            </label>
            <label
              class="method-card"
              :class="{ selected: fixedClockMethod === 'photo' }"
            >
              <input type="radio" value="photo" v-model="fixedClockMethod" class="hidden-check" />
              <div class="method-icon">
                <el-icon><Camera /></el-icon>
              </div>
              <span class="method-name">拍照打卡</span>
              <span class="method-desc">拍照留证更严格</span>
            </label>
          </div>
        </div>

        <div class="form-group">
          <div class="section-label">
            <el-icon><Present /></el-icon>
            <span>节假日设置</span>
          </div>
          <div class="holiday-table">
            <div class="holiday-header">
              <span class="col-date">日期</span>
              <span class="col-name">名称</span>
              <span class="col-action">操作</span>
            </div>
            <TransitionGroup name="holiday">
              <div v-for="(h, idx) in fixedHolidays" :key="idx" class="holiday-row">
                <div class="col-date">
                  <el-date-picker
                    v-model="h.date"
                    type="date"
                    format="YYYY-MM-DD"
                    value-format="YYYY-MM-DD"
                    placeholder="选择日期"
                    size="default"
                    style="width: 100%"
                  />
                </div>
                <div class="col-name">
                  <el-input v-model="h.name" placeholder="节假日名称（如：元旦）" size="default" />
                </div>
                <div class="col-action">
                  <el-popconfirm title="确认删除？" @confirm="fixedHolidays.splice(idx, 1)">
                    <template #reference>
                      <el-button type="danger" text size="small">
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </template>
                  </el-popconfirm>
                </div>
              </div>
            </TransitionGroup>
          </div>
          <el-button class="add-holiday-btn" @click="fixedHolidays.push({ date: '', name: '' })">
            <el-icon><Plus /></el-icon>
            添加节假日
          </el-button>
        </div>
      </div>

      <!-- 按排班 -->
      <div v-show="activeTab === 'scheduled'" class="rule-content">
        <div class="form-group">
          <label class="form-label">打卡方式</label>
          <div class="method-cards">
            <label class="method-card" :class="{ selected: scheduledClockMethod === 'click' }">
              <input type="radio" value="click" v-model="scheduledClockMethod" class="hidden-check" />
              <div class="method-icon"><el-icon><Pointer /></el-icon></div>
              <span class="method-name">手动打卡</span>
              <span class="method-desc">员工主动签到</span>
            </label>
            <label class="method-card" :class="{ selected: scheduledClockMethod === 'photo' }">
              <input type="radio" value="photo" v-model="scheduledClockMethod" class="hidden-check" />
              <div class="method-icon"><el-icon><Camera /></el-icon></div>
              <span class="method-name">拍照打卡</span>
              <span class="method-desc">拍照留证更严格</span>
            </label>
          </div>
        </div>

        <div class="form-group">
          <div class="section-label">
            <el-icon><Clock /></el-icon>
            <span>班次管理</span>
          </div>
          <div v-if="shifts.length === 0" class="empty-shifts">
            <el-icon><Folder /></el-icon>
            <span>暂无班次，点击下方按钮新建</span>
          </div>
          <div v-else class="shift-list">
            <div v-for="shift in shifts" :key="shift.id" class="shift-item">
              <div class="shift-left">
                <div class="shift-name">{{ shift.name }}</div>
                <div class="shift-time">
                  <el-icon><Timer /></el-icon>
                  {{ shift.work_start }} - {{ shift.work_end }}
                </div>
              </div>
              <div class="shift-right">
                <el-tag v-if="shift.work_date_offset !== 0" size="small" type="warning">跨天</el-tag>
                <el-button size="small" text type="primary" @click="openShiftDialog(shift)">
                  <el-icon><EditPen /></el-icon>
                  编辑
                </el-button>
                <el-popconfirm title="确认删除？" @confirm="handleDeleteShift(shift.id)">
                  <template #reference>
                    <el-button size="small" text type="danger">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </template>
                </el-popconfirm>
              </div>
            </div>
          </div>
          <el-button class="add-shift-btn" type="primary" @click="openShiftDialog(null)">
            <el-icon><Plus /></el-icon>
            新建班次
          </el-button>
        </div>
      </div>

      <!-- 自由工时 -->
      <div v-show="activeTab === 'free'" class="rule-content">
        <div class="form-section">
          <div class="section-label">
            <el-icon><Calendar /></el-icon>
            <span>上班日设置</span>
          </div>
          <div class="work-days-grid">
            <label
              v-for="day in weekDays"
              :key="day.value"
              class="day-chip"
              :class="{ selected: freeWorkDays.includes(day.value) }"
            >
              <input type="checkbox" :value="day.value" v-model="freeWorkDays" class="hidden-check" />
              <span>{{ day.label }}</span>
            </label>
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">打卡位置</label>
          <el-input v-model="freeLocation" placeholder="请输入打卡位置" size="large">
            <template #prefix>
              <el-icon><Location /></el-icon>
            </template>
          </el-input>
        </div>

        <div class="form-group">
          <label class="form-label">打卡方式</label>
          <div class="method-cards">
            <label class="method-card" :class="{ selected: freeClockMethod === 'click' }">
              <input type="radio" value="click" v-model="freeClockMethod" class="hidden-check" />
              <div class="method-icon"><el-icon><Pointer /></el-icon></div>
              <span class="method-name">手动打卡</span>
              <span class="method-desc">员工主动签到</span>
            </label>
            <label class="method-card" :class="{ selected: freeClockMethod === 'photo' }">
              <input type="radio" value="photo" v-model="freeClockMethod" class="hidden-check" />
              <div class="method-icon"><el-icon><Camera /></el-icon></div>
              <span class="method-name">拍照打卡</span>
              <span class="method-desc">拍照留证更严格</span>
            </label>
          </div>
        </div>

        <div class="form-group">
          <div class="section-label">
            <el-icon><Present /></el-icon>
            <span>节假日设置</span>
          </div>
          <div class="holiday-table">
            <div class="holiday-header">
              <span class="col-date">日期</span>
              <span class="col-name">名称</span>
              <span class="col-action">操作</span>
            </div>
            <TransitionGroup name="holiday">
              <div v-for="(h, idx) in freeHolidays" :key="idx" class="holiday-row">
                <div class="col-date">
                  <el-date-picker
                    v-model="h.date"
                    type="date"
                    format="YYYY-MM-DD"
                    value-format="YYYY-MM-DD"
                    placeholder="选择日期"
                    size="default"
                    style="width: 100%"
                  />
                </div>
                <div class="col-name">
                  <el-input v-model="h.name" placeholder="节假日名称" size="default" />
                </div>
                <div class="col-action">
                  <el-popconfirm title="确认删除？" @confirm="freeHolidays.splice(idx, 1)">
                    <template #reference>
                      <el-button type="danger" text size="small">
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </template>
                  </el-popconfirm>
                </div>
              </div>
            </TransitionGroup>
          </div>
          <el-button class="add-holiday-btn" @click="freeHolidays.push({ date: '', name: '' })">
            <el-icon><Plus /></el-icon>
            添加节假日
          </el-button>
        </div>
      </div>

      <div class="rule-footer">
        <el-button size="large" :loading="saving" type="primary" class="save-btn" @click="handleSave">
          <el-icon><Check /></el-icon>
          保存规则
        </el-button>
      </div>
    </div>

    <!-- 班次编辑弹窗 -->
    <el-dialog
      v-model="shiftDialogVisible"
      :title="editingShift ? '编辑班次' : '新建班次'"
      width="460px"
      class="shift-dialog"
    >
      <div class="shift-form">
        <div class="form-group">
          <label class="form-label">班次名称</label>
          <el-input v-model="shiftForm.name" placeholder="如：早班/晚班" size="large" />
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">上班时间</label>
            <el-time-picker
              v-model="shiftForm.work_start"
              format="HH:mm"
              value-format="HH:mm"
              size="large"
              style="width: 100%"
            />
          </div>
          <div class="form-group">
            <label class="form-label">下班时间</label>
            <el-time-picker
              v-model="shiftForm.work_end"
              format="HH:mm"
              value-format="HH:mm"
              size="large"
              style="width: 100%"
            />
          </div>
        </div>
        <div class="cross-day-toggle">
          <el-switch v-model="shiftForm.is_cross_day" />
          <div class="cross-day-info">
            <span class="cross-day-title">跨天班次</span>
            <span class="cross-day-desc">开启后下班时间归属次日</span>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="shiftDialogVisible = false" size="large">取消</el-button>
        <el-button type="primary" :loading="shiftSaving" size="large" @click="handleSaveShift">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { attendanceApi } from '@/api/attendance'
import type { Shift } from '@/api/attendance'
import {
  Timer, Clock, Calendar, Location, Pointer, Camera,
  Present, Folder, EditPen, Delete, Plus, Check,
} from '@element-plus/icons-vue'

const activeTab = ref<'fixed' | 'scheduled' | 'free'>('fixed')

const modeTabs = [
  { label: '固定时间', value: 'fixed', icon: 'Timer' },
  { label: '按排班', value: 'scheduled', icon: 'Clock' },
  { label: '自由工时', value: 'free', icon: 'Sunrise' },
]

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

const shiftDialogVisible = ref(false)
const editingShift = ref<Shift | null>(null)
const shiftSaving = ref(false)
const shiftForm = ref({ name: '', work_start: '', work_end: '', is_cross_day: false })

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
  } catch { /* ignore */ }
}

async function loadShifts() {
  try {
    const res = await attendanceApi.listShifts() as any
    shifts.value = res.data || []
  } catch { /* ignore */ }
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
      data = { mode: 'scheduled', clock_method: scheduledClockMethod.value, holidays: [] }
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
  } catch { ElMessage.error('保存失败，请稍后重试') }
  finally { saving.value = false }
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
  } catch { ElMessage.error('保存失败') }
  finally { shiftSaving.value = false }
}

async function handleDeleteShift(id: number) {
  try {
    await attendanceApi.deleteShift(id)
    ElMessage.success('已删除')
    loadShifts()
  } catch { ElMessage.error('删除失败') }
}

onMounted(() => { loadRule(); loadShifts() })
</script>

<style scoped lang="scss">
.rule-card {
  padding: 0;
  overflow: hidden;
}

.rule-card-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border);
}

.mode-tabs {
  display: flex;
  gap: 8px;
}

.mode-tab {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  border: 1.5px solid transparent;
  transition: all 0.2s ease;
  background: transparent;

  &:hover {
    color: var(--primary);
    background: rgba(var(--primary), 0.06);
  }

  &.active {
    color: var(--primary);
    background: rgba(var(--primary), 0.08);
    border-color: rgba(var(--primary), 0.3);
  }

  .el-icon { font-size: 16px; }
}

.rule-content {
  padding: 28px 24px;
}

.form-section { margin-bottom: 28px; }

.section-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
  .el-icon { color: var(--primary); }
}

.form-row { display: flex; gap: 16px; margin-bottom: 20px; }
.form-group { flex: 1; margin-bottom: 20px; }

.form-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.work-days-grid {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.day-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 56px;
  padding: 8px 14px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  background: #F3F4F6;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1.5px solid transparent;
  user-select: none;

  &.selected {
    color: #fff;
    background: var(--primary);
    border-color: var(--primary);
  }

  &:hover:not(.selected) {
    border-color: rgba(var(--primary), 0.4);
    color: var(--primary);
    background: rgba(var(--primary), 0.06);
  }

  .hidden-check { display: none; }
}

.method-cards {
  display: flex;
  gap: 12px;
}

.method-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 20px;
  border: 2px solid var(--border);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: all 0.2s ease;

  &.selected {
    border-color: var(--primary);
    background: rgba(var(--primary), 0.04);
  }

  &:hover:not(.selected) {
    border-color: rgba(var(--primary), 0.4);
    transform: translateY(-2px);
  }

  .hidden-check { display: none; }
}

.method-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, #EDE9FE, #DDD6FE);
  color: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;

  .method-card.selected & {
    background: linear-gradient(135deg, var(--primary-light), var(--primary));
    color: #fff;
  }
}

.method-name { font-size: 14px; font-weight: 600; color: var(--text-primary); }
.method-desc { font-size: 12px; color: var(--text-tertiary); }

.holiday-table {
  background: #FAFAFA;
  border-radius: var(--radius-md);
  overflow: hidden;
  margin-bottom: 8px;
}

.holiday-header {
  display: flex;
  gap: 12px;
  padding: 10px 12px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-tertiary);
  background: #F3F4F6;
}

.holiday-row {
  display: flex;
  gap: 12px;
  align-items: center;
  padding: 8px 12px;
  background: #fff;
  border-top: 1px solid var(--border);
}

.col-date { width: 180px; }
.col-name { flex: 1; }
.col-action { width: 40px; text-align: center; }

.add-holiday-btn {
  border-style: dashed;
  color: var(--primary);
  border-color: rgba(var(--primary), 0.4);
  background: rgba(var(--primary), 0.04);
  width: 100%;
  &:hover { background: rgba(var(--primary), 0.08); border-color: var(--primary); }
}

.empty-shifts {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 40px;
  color: var(--text-tertiary);
  font-size: 14px;
  .el-icon { font-size: 32px; }
}

.shift-list { display: flex; flex-direction: column; gap: 8px; margin-bottom: 12px; }

.shift-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  background: #FAFAFA;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
  transition: all 0.2s ease;
  &:hover { border-color: rgba(var(--primary), 0.3); background: rgba(var(--primary), 0.02); }
}

.shift-left { display: flex; flex-direction: column; gap: 4px; }
.shift-name { font-size: 14px; font-weight: 600; color: var(--text-primary); }
.shift-time { display: flex; align-items: center; gap: 4px; font-size: 13px; color: var(--text-tertiary); }

.shift-right { display: flex; align-items: center; gap: 8px; }

.add-shift-btn {
  width: 100%;
  border-style: dashed;
  border-color: rgba(var(--primary), 0.4);
  color: var(--primary);
  background: rgba(var(--primary), 0.04);
  &:hover { background: rgba(var(--primary), 0.08); border-color: var(--primary); }
}

.rule-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--border);
  display: flex;
  justify-content: flex-end;
}

.save-btn {
  background: linear-gradient(135deg, var(--primary-light), var(--primary));
  border: none;
  box-shadow: 0 4px 14px rgba(var(--primary), 0.4);
  &:hover { box-shadow: 0 6px 20px rgba(var(--primary), 0.5); transform: translateY(-1px); }
  transition: all 0.2s ease;
}

.shift-form { padding: 4px 0; }
.cross-day-toggle {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: #FEF9C3;
  border-radius: var(--radius-md);
  border: 1px solid #FDE68A;
}
.cross-day-info { display: flex; flex-direction: column; gap: 2px; }
.cross-day-title { font-size: 14px; font-weight: 600; color: var(--text-primary); }
.cross-day-desc { font-size: 12px; color: var(--text-tertiary); }

.holiday-enter-active, .holiday-leave-active { transition: all 0.25s ease; }
.holiday-enter-from { opacity: 0; transform: translateY(-8px); }
.holiday-leave-to { opacity: 0; transform: translateX(8px); }

@media (max-width: 768px) {
  .attendance-rule { padding: 16px; }
  .form-row { flex-direction: column; }
  .method-cards { flex-direction: column; }
  .mode-tabs { flex-wrap: wrap; }
}
</style>
