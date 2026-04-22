<template>
  <div class="salary-slip-send">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">工资条发送</h1>
        <p class="page-subtitle">向员工发送电子工资条通知</p>
      </div>
    </header>

    <!-- 标签页 -->
    <div class="nav-tabs glass-card">
      <div class="tab-group">
        <button
          v-for="tab in navTabs"
          :key="tab.value"
          class="tab-btn"
          :class="{ active: activeTab === tab.value }"
          @click="activeTab = tab.value"
        >
          <el-icon><component :is="tab.icon" /></el-icon>
          {{ tab.label }}
        </button>
      </div>
    </div>

    <!-- 发送工资条 -->
    <div v-show="activeTab === 'send'" class="tab-content">
      <div class="send-card glass-card">
        <div class="send-form">
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">发放月份</label>
              <el-date-picker
                v-model="sendYM"
                type="month"
                placeholder="选择月份"
                value-format="YYYY-MM"
                size="large"
                style="width: 100%"
              >
                <template #prefix>
                  <el-icon><Calendar /></el-icon>
                </template>
              </el-date-picker>
            </div>
            <div class="form-group">
              <label class="form-label">发送渠道</label>
              <div class="channel-selector">
                <label
                  v-for="ch in channelOptions"
                  :key="ch.value"
                  class="channel-option"
                  :class="{ selected: sendChannel === ch.value }"
                >
                  <input type="radio" :value="ch.value" v-model="sendChannel" class="hidden-check" />
                  <el-icon><component :is="ch.icon" /></el-icon>
                  <span>{{ ch.label }}</span>
                </label>
              </div>
            </div>
          </div>

          <div class="send-actions">
            <el-button
              type="primary"
              size="large"
              :loading="sendingAll"
              :disabled="!sendYM"
              @click="handleSendAll"
              class="send-all-btn"
            >
              <el-icon><Promotion /></el-icon>
              向全员发送
            </el-button>
            <el-button
              size="large"
              :disabled="!sendYM"
              @click="activeTab = 'select'"
              class="send-select-btn"
            >
              <el-icon><User /></el-icon>
              向选定员工发送
            </el-button>
          </div>
        </div>
      </div>

      <!-- 已选员工表格 -->
      <div v-if="activeTab === 'select'" class="select-card glass-card">
        <div class="select-header">
          <div class="select-info">
            <el-icon><Tickets /></el-icon>
            <span>已选择 <strong>{{ selectedEmployeeIds.length }}</strong> 名员工</span>
          </div>
          <div class="select-actions">
            <el-button @click="activeTab = 'send'">取消</el-button>
            <el-button
              type="primary"
              :loading="sendingSelected"
              @click="handleSendSelected"
            >
              <el-icon><Promotion /></el-icon>
              确认发送
            </el-button>
          </div>
        </div>

        <el-table
          ref="employeeTableRef"
          :data="employeeList"
          stripe
          @selection-change="handleSelectionChange"
          class="employee-table"
          :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
        >
          <el-table-column type="selection" width="50" />
          <el-table-column prop="employee_name" label="员工姓名" min-width="120">
            <template #default="{ row }">
              <div class="name-cell">
                <el-avatar :size="28" class="name-avatar">{{ row.employee_name?.[0] }}</el-avatar>
                <span>{{ row.employee_name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="department_name" label="部门" min-width="120">
            <template #default="{ row }">
              <span class="dept-text">{{ row.department_name || '—' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="工资状态" width="110">
            <template #default="{ row }">
              <span class="status-badge" :class="`status--${row.status}`">
                {{ payrollStatusMap[row.status] }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <!-- 发送记录 -->
    <div v-show="activeTab === 'logs'" class="tab-content">
      <div class="toolbar-card glass-card">
        <div class="toolbar-form">
          <div class="toolbar-item">
            <label>年份</label>
            <el-input-number
              v-model="logYear"
              :min="2020"
              :max="2030"
              size="large"
              @change="loadLogs"
            />
          </div>
          <div class="toolbar-item">
            <label>月份</label>
            <el-input-number
              v-model="logMonth"
              :min="1"
              :max="12"
              size="large"
              @change="loadLogs"
            />
          </div>
        </div>
        <el-button size="large" @click="() => loadLogs()" class="refresh-btn">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>

      <div class="table-card glass-card" v-loading="loadingLogs">
        <el-table
          :data="logs"
          stripe
          class="modern-table"
          :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
        >
          <el-table-column prop="employee_id" label="员工ID" width="90">
            <template #default="{ row }">
              <span class="id-text">#{{ row.employee_id }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="channel" label="渠道" width="110">
            <template #default="{ row }">
              <span class="channel-tag">{{ slipChannelMap[row.channel] }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="110">
            <template #default="{ row }">
              <span class="status-badge" :class="`status--${row.status}`">
                {{ slipStatusMap[row.status] }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="error_message" label="错误信息" min-width="200" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="error-text" v-if="row.error_message">{{ row.error_message }}</span>
              <span class="no-error" v-else>—</span>
            </template>
          </el-table-column>
          <el-table-column prop="sent_at" label="发送时间" width="170">
            <template #default="{ row }">
              <span class="time-text">{{ row.sent_at || '—' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="confirmed_at" label="员工确认" width="120">
            <template #default="{ row }">
              <span v-if="row.confirmed_at" class="confirmed-tag">已确认</span>
              <span v-else class="unconfirmed-tag">未确认</span>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="170" />
        </el-table>

        <div class="pagination-wrapper">
          <el-pagination
            layout="total, prev, pager, next"
            :total="logTotal"
            :page="logPage"
            :page-size="logPageSize"
            @current-change="loadLogs"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { salaryApi, type SlipSendLog } from '@/api/salary'
import {
  Promotion, User, Tickets, Refresh, Calendar,
} from '@element-plus/icons-vue'

const navTabs = [
  { label: '发送工资条', value: 'send',   icon: 'Promotion' },
  { label: '发送记录',   value: 'logs',   icon: 'List'       },
]

const activeTab = ref('send')
const sendYM = ref('')
const sendChannel = ref('miniapp')

const channelOptions = [
  { label: '小程序', value: 'miniapp', icon: 'ChatDotRound' },
  { label: '短信',   value: 'sms',     icon: 'Message'      },
  { label: 'H5链接', value: 'h5',      icon: 'Link'         },
]

const sendingAll = ref(false)

async function handleSendAll() {
  if (!sendYM.value) return
  const [year, month] = sendYM.value.split('-').map(Number)
  sendingAll.value = true
  try {
    await salaryApi.sendSlipAll({ year, month, channel: sendChannel.value })
    ElMessage.success('工资条发送任务已入队，请稍后在"发送记录"中查看进度')
  } catch (e: any) {
    ElMessage.error(e?.message || '发送失败')
  } finally {
    sendingAll.value = false
  }
}

const employeeList = ref<any[]>([])
const selectedEmployeeIds = ref<number[]>([])
const sendingSelected = ref(false)

async function loadEmployeeList() {
  if (!sendYM.value) return
  const [year, month] = sendYM.value.split('-').map(Number)
  try {
    const res = await salaryApi.list({ year, month, page: 1, page_size: 200 })
    employeeList.value = res?.list ?? []
  } catch { ElMessage.error('加载员工列表失败') }
}

function handleSelectionChange(rows: any[]) {
  selectedEmployeeIds.value = rows.map((r) => r.id)
}

async function handleSendSelected() {
  if (selectedEmployeeIds.value.length === 0) {
    ElMessage.warning('请先选择员工')
    return
  }
  const [year, month] = sendYM.value.split('-').map(Number)
  sendingSelected.value = true
  try {
    await salaryApi.sendSlipAll({
      year, month,
      employee_ids: selectedEmployeeIds.value,
      channel: sendChannel.value,
    })
    ElMessage.success('工资条发送任务已入队')
    activeTab.value = 'logs'
  } catch (e: any) {
    ElMessage.error(e?.message || '发送失败')
  } finally {
    sendingSelected.value = false
  }
}

const loadingLogs = ref(false)
const logs = ref<SlipSendLog[]>([])
const logTotal = ref(0)
const logPage = ref(1)
const logPageSize = ref(20)
const logYear = ref(new Date().getFullYear())
const logMonth = ref(new Date().getMonth() + 1)

const payrollStatusMap: Record<string, string> = {
  draft: '草稿', calculated: '已核算', confirmed: '已确认', paid: '已发放',
}

const slipChannelMap: Record<string, string> = {
  miniapp: '小程序', sms: '短信', h5: 'H5链接',
}

const slipStatusMap: Record<string, string> = {
  pending: '等待中', sending: '发送中', sent: '已发送', failed: '失败',
}

async function loadLogs(p: number = 1) {
  logPage.value = p
  loadingLogs.value = true
  try {
    const res = await salaryApi.getSlipLogs({
      year: logYear.value, month: logMonth.value, page: p, page_size: logPageSize.value,
    })
    logs.value = res.logs
    logTotal.value = res.total
  } catch { ElMessage.error('加载发送记录失败') }
  finally { loadingLogs.value = false }
}

onMounted(() => {
  const now = new Date()
  sendYM.value = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  loadLogs()
  loadEmployeeList()
})
</script>

<style scoped lang="scss">
$success: #10B981;
$warning: #F59E0B;
$error: #EF4444;
$bg-page: #FAFBFC;
$text-primary: #1F2937;
$text-secondary: #6B7280;
$text-muted: #9CA3AF;
$border-color: #E5E7EB;
$radius-sm: 8px;
$radius-md: 12px;
$radius-lg: 16px;
$radius-xl: 24px;

.salary-slip-send { padding: 24px 32px; width: 100%; box-sizing: border-box; background: $bg-page; min-height: 100vh; }

.glass-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: $radius-xl;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 24px;
  .page-title { font-size: 24px; font-weight: 700; color: $text-primary; margin: 0 0 4px; }
  .page-subtitle { font-size: 14px; color: $text-secondary; margin: 0; }
}

.nav-tabs { padding: 14px 20px; margin-bottom: 20px; }

.tab-group { display: inline-flex; background: #F3F4F6; border-radius: $radius-md; padding: 4px; }

.tab-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 8px 20px;
  border-radius: $radius-sm;
  font-size: 14px; font-weight: 500; color: $text-secondary;
  cursor: pointer; transition: all 0.2s ease; border: none; background: transparent;

  &.active { background: #fff; color: var(--primary); box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1); }
  &:hover:not(.active) { color: $text-primary; }
  .el-icon { font-size: 15px; }
}

.tab-content { display: flex; flex-direction: column; gap: 16px; }

.send-card { padding: 28px; }
.send-form { max-width: 600px; }

.form-row { display: flex; gap: 20px; margin-bottom: 24px; }
.form-group { flex: 1; }
.form-label { display: block; font-size: 13px; font-weight: 500; color: $text-secondary; margin-bottom: 8px; }

.channel-selector { display: flex; gap: 8px; flex-wrap: wrap; }

.channel-option {
  display: flex; flex-direction: column; align-items: center; gap: 6px;
  padding: 16px 20px;
  border: 2px solid $border-color; border-radius: $radius-md;
  cursor: pointer; transition: all 0.2s ease; min-width: 80px;

  .el-icon { font-size: 22px; color: $text-muted; }
  span { font-size: 12px; font-weight: 500; color: $text-secondary; }

  &.selected {
    border-color: var(--primary); background: rgba(var(--primary), 0.04);
    .el-icon, span { color: var(--primary); }
  }
  &:hover:not(.selected) { border-color: rgba(var(--primary), 0.4); transform: translateY(-2px); }
  .hidden-check { display: none; }
}

.send-actions { display: flex; gap: 12px; }

.send-all-btn {
  background: linear-gradient(135deg, var(--primary-light), var(--primary));
  border: none;
  box-shadow: 0 4px 14px rgba(var(--primary), 0.4);
  &:hover { box-shadow: 0 6px 20px rgba(var(--primary), 0.5); }
}

.send-select-btn {
  border-style: dashed;
  color: var(--primary);
  border-color: rgba(var(--primary), 0.4);
  background: rgba(var(--primary), 0.04);
  &:hover { background: rgba(var(--primary), 0.08); border-color: var(--primary); }
}

.select-card { padding: 0; overflow: hidden; }

.select-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid $border-color;
}

.select-info { display: flex; align-items: center; gap: 8px; font-size: 14px; color: $text-secondary; .el-icon { font-size: 18px; } strong { color: var(--primary); } }
.select-actions { display: flex; gap: 8px; }

:deep(.employee-table) {
  .el-table__header th { padding: 14px 12px; font-size: 13px; }
  .el-table__row { transition: background 0.2s ease; &:hover > td { background: rgba(var(--primary), 0.02) !important; } }
  .el-table__cell { padding: 12px; border-bottom: 1px solid #F3F4F6; }
}

.name-cell { display: flex; align-items: center; gap: 8px; font-weight: 500; color: $text-primary; }
.name-avatar { background: linear-gradient(135deg, var(--primary-light), var(--primary)); color: #fff; font-size: 12px; font-weight: 600; }
.dept-text { font-size: 13px; color: $text-muted; }

.toolbar-card { display: flex; align-items: center; justify-content: space-between; padding: 16px 20px; }
.toolbar-form { display: flex; gap: 16px; align-items: center; }
.toolbar-item { display: flex; align-items: center; gap: 8px; label { font-size: 13px; font-weight: 500; color: $text-secondary; white-space: nowrap; } }

.table-card { padding: 0; overflow: hidden; }

:deep(.modern-table) {
  .el-table__header th { padding: 14px 16px; font-size: 13px; }
  .el-table__row { transition: background 0.2s ease; &:hover > td { background: rgba(var(--primary), 0.02) !important; } }
  .el-table__cell { padding: 14px 16px; border-bottom: 1px solid #F3F4F6; }
}

.status-badge {
  display: inline-flex; align-items: center; padding: 3px 10px;
  font-size: 12px; font-weight: 500; border-radius: 12px;

  &.status--draft, &.status--pending { background: #F3F4F6; color: #6B7280; }
  &.status--calculated { background: #FEF3C7; color: #D97706; }
  &.status--confirmed, &.status--paid, &.status--sent { background: #D1FAE5; color: #059669; }
  &.status--sending { background: #DBEAFE; color: #3B82F6; }
  &.status--failed { background: #FEE2E2; color: #DC2626; }
}

.confirmed-tag {
  display: inline-flex; align-items: center; padding: 3px 10px;
  background: #D1FAE5; color: #059669;
  font-size: 12px; font-weight: 500; border-radius: 12px;
}
.unconfirmed-tag {
  display: inline-flex; align-items: center; padding: 3px 10px;
  background: #FEF3C7; color: #D97706;
  font-size: 12px; font-weight: 500; border-radius: 12px;
}
.channel-tag {
  display: inline-flex; padding: 3px 10px;
  background: #EDE9FE; color: var(--primary);
  font-size: 12px; font-weight: 600; border-radius: 12px;
}

.id-text { font-family: 'SF Mono', Monaco, monospace; font-weight: 600; color: $text-muted; font-size: 13px; }
.error-text { color: $error; font-size: 13px; }
.no-error { color: $text-muted; }
.time-text { font-size: 13px; color: $text-muted; font-family: 'SF Mono', Monaco, monospace; }

.pagination-wrapper { display: flex; justify-content: flex-end; padding: 16px 20px; border-top: 1px solid $border-color; }

.refresh-btn { &:hover { color: var(--primary); border-color: rgba(var(--primary), 0.4); } }

@media (max-width: 768px) {
  .salary-slip-send { padding: 16px; }
  .form-row { flex-direction: column; }
  .channel-selector { flex-direction: column; }
  .channel-option { flex-direction: row; justify-content: flex-start; }
  .toolbar-card { flex-direction: column; align-items: stretch; gap: 12px; }
}
</style>
