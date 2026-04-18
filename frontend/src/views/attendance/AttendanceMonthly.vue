<template>
  <div class="attendance-monthly">
    <el-card>
      <template #header>
        <div class="header">
          <span>出勤月报</span>
          <div class="header-actions">
            <el-date-picker v-model="selectedMonth" type="month" format="YYYY-MM" value-format="YYYY-MM" placeholder="选择月份" style="width: 140px" @change="load(1)" />
            <el-button type="primary" @click="handleExport">导出月报</el-button>
          </div>
        </div>
      </template>

      <div class="view-toggle">
        <el-button-group>
          <el-button :type="viewMode === 'stat' ? 'primary' : 'default'" @click="viewMode = 'stat'">统计视图</el-button>
          <el-button :type="viewMode === 'grid' ? 'primary' : 'default'" @click="viewMode = 'grid'">格子视图</el-button>
        </el-button-group>
      </div>

      <div v-if="viewMode === 'stat'" v-loading="loading">
        <div class="stats-grid">
          <div class="stat-card"><div class="stat-value">{{ stats.total_actual_days }}</div><div class="stat-label">实际出勤（天）</div></div>
          <div class="stat-card"><div class="stat-value">{{ stats.total_required_days }}</div><div class="stat-label">应出勤（天）</div></div>
          <div class="stat-card"><div class="stat-value">{{ stats.total_overtime_hours }}</div><div class="stat-label">加班时长（小时）</div></div>
          <div class="stat-card"><div class="stat-value">{{ stats.total_absent_days }}</div><div class="stat-label">缺勤（天）</div></div>
        </div>

        <el-table :data="list" stripe>
          <el-table-column prop="employee_name" label="姓名" min-width="90" />
          <el-table-column prop="department_name" label="部门" min-width="100" />
          <el-table-column label="实际出勤" min-width="90">
            <template #default="{ row }">{{ row.actual_days }} 天</template>
          </el-table-column>
          <el-table-column label="应出勤" min-width="90">
            <template #default="{ row }">{{ row.required_days }} 天</template>
          </el-table-column>
          <el-table-column label="加班" min-width="90">
            <template #default="{ row }">{{ row.overtime_hours }}h</template>
          </el-table-column>
          <el-table-column label="缺勤" min-width="80">
            <template #default="{ row }">{{ row.absent_days }} 天</template>
          </el-table-column>
          <el-table-column label="出勤率" min-width="90">
            <template #default="{ row }">
              <span :style="{ color: row.attendance_rate >= 95 ? '#52C41A' : row.attendance_rate >= 80 ? '#FA8C16' : '#FF4D4F' }">
                {{ row.attendance_rate.toFixed(1) }}%
              </span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="primary" link @click="viewDaily(row)">查看打卡详情</el-button>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-wrapper">
          <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />
        </div>
      </div>

      <div v-if="viewMode === 'grid'" v-loading="loading">
        <el-table :data="list" stripe border style="width: 100%">
          <el-table-column prop="employee_name" label="姓名" width="100" fixed="left" />
          <el-table-column prop="actual_days" label="实际出勤" width="80" fixed="right" />
          <el-table-column prop="required_days" label="应出勤" width="80" fixed="right" />
        </el-table>
      </div>
    </el-card>

    <el-drawer v-model="drawerVisible" :title="`${drawerName} - ${selectedMonth} 打卡详情`" size="600px">
      <div v-if="dailyRecords.length === 0" style="text-align:center;color:#8c8c8c;padding:40px">暂无打卡数据</div>
      <div v-else>
        <div v-for="r in dailyRecords" :key="r.date" class="daily-row" :class="{ 'is-holiday': r.is_holiday, 'is-weekend': r.is_weekend }">
          <span class="date-col">{{ formatDay(r.date) }}</span>
          <span class="symbol-col" :style="{ color: r.symbol === '√' ? '#52c41a' : r.symbol === '迟到' ? '#fa8c16' : r.symbol === '缺' ? '#ff4d4f' : '#bfbfbf' }">{{ r.symbol }}</span>
          <span v-if="r.clock_in" class="clock-col">上班 {{ r.clock_in }} 下班 {{ r.clock_out || '--' }}</span>
          <span v-else class="clock-col" style="color:#bfbfbf">--</span>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { attendanceApi, type MonthlyReportItem, type DailyRecord } from '@/api/attendance'

const loading = ref(false)
const viewMode = ref<'stat' | 'grid'>('stat')
const selectedMonth = ref(new Date().toISOString().slice(0, 7))
const list = ref<MonthlyReportItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const stats = ref({ total_actual_days: 0, total_required_days: 0, total_overtime_hours: 0, total_absent_days: 0 })
const drawerVisible = ref(false)
const drawerName = ref('')
const dailyRecords = ref<DailyRecord[]>([])

const weekdayNames = ['日', '一', '二', '三', '四', '五', '六']

function formatDay(dateStr: string) {
  const [, m, d] = dateStr.split('-')
  const date = new Date(dateStr)
  return `${parseInt(m)}月${parseInt(d)}日 周${weekdayNames[date.getDay()]}`
}

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const { data } = await attendanceApi.getMonthlyReport({ year_month: selectedMonth.value, page: p, page_size: pageSize })
    list.value = data?.list ?? []
    total.value = data?.total ?? 0
    if (data?.stats) stats.value = data.stats
  } catch { ElMessage.error('加载月报数据失败') }
  finally { loading.value = false }
}

async function viewDaily(row: MonthlyReportItem) {
  drawerName.value = row.employee_name || '员工'
  try {
    const { data } = await attendanceApi.getDailyRecords({ employee_id: row.employee_id, year_month: selectedMonth.value })
    dailyRecords.value = data?.records ?? []
  } catch { dailyRecords.value = [] }
  drawerVisible.value = true
}

async function handleExport() {
  try {
    const res = await attendanceApi.exportMonthlyExcel({ year_month: selectedMonth.value })
    const blob = res as unknown as Blob
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `出勤月报_${selectedMonth.value}.xlsx`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('月报导出成功')
  } catch { ElMessage.error('导出失败') }
}

onMounted(() => load())
</script>

<style scoped>
.header { display: flex; justify-content: space-between; align-items: center; }
.header-actions { display: flex; gap: 8px; }
.view-toggle { margin-bottom: 16px; }
.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 16px; }
.stat-card { text-align: center; padding: 16px; background: #fafafa; border-radius: 8px; }
.stat-value { font-size: 24px; font-weight: 700; color: #1a1a1a; }
.stat-label { font-size: 13px; color: #8c8c8c; margin-top: 4px; }
.pagination-wrapper { display: flex; justify-content: flex-end; margin-top: 16px; }
.daily-row { display: flex; align-items: center; gap: 12px; padding: 8px 0; border-bottom: 1px solid #f0f0f0; }
.daily-row.is-holiday { background: #fff1f0; padding: 8px 12px; margin: 0 -12px; }
.daily-row.is-weekend { background: #fafafa; }
.date-col { min-width: 90px; font-size: 13px; }
.symbol-col { min-width: 40px; font-weight: 700; }
.clock-col { font-size: 13px; color: #52c41a; }
</style>
