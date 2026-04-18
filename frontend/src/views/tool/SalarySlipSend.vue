<template>
  <div class="salary-slip-send">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>工资条发送</span>
        </div>
      </template>

      <el-tabs v-model="activeTab">
        <!-- Tab 1: 发送工资条 -->
        <el-tab-pane label="发送工资条" name="send">
          <el-form inline @submit.prevent>
            <el-form-item label="发放月份">
              <el-date-picker
                v-model="sendYM"
                type="month"
                placeholder="选择月份"
                value-format="YYYY-MM"
                style="width: 140px"
              />
            </el-form-item>
            <el-form-item label="发送渠道">
              <el-select v-model="sendChannel" placeholder="默认小程序" style="width: 120px">
                <el-option label="小程序" value="miniapp" />
                <el-option label="短信" value="sms" />
                <el-option label="H5链接" value="h5" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="sendingAll" :disabled="!sendYM" @click="handleSendAll">
                向全员发送
              </el-button>
              <el-button :disabled="!sendYM" @click="activeTab = 'select'">
                向选定员工发送
              </el-button>
            </el-form-item>
          </el-form>

          <!-- 已选员工表格 -->
          <div v-if="activeTab === 'select'" class="employee-select">
            <el-form-item>
              <el-button type="primary" :loading="sendingSelected" @click="handleSendSelected">
                确认发送（已选 {{ selectedEmployeeIds.length }} 人）
              </el-button>
              <el-button @click="activeTab = 'send'">取消</el-button>
            </el-form-item>
            <el-table
              ref="employeeTableRef"
              :data="employeeList"
              stripe
              @selection-change="handleSelectionChange"
            >
              <el-table-column type="selection" width="40" />
              <el-table-column prop="employee_name" label="员工姓名" min-width="100" />
              <el-table-column prop="status" label="工资状态" width="90">
                <template #default="{ row }">
                  <el-tag size="small" :type="payrollStatusTagType[row.status]">
                    {{ payrollStatusMap[row.status] }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>

        <!-- Tab 2: 发送记录 -->
        <el-tab-pane label="发送记录" name="logs">
          <el-form inline @submit.prevent>
            <el-form-item label="年份">
              <el-input-number
                v-model="logYear"
                :min="2020"
                :max="2030"
                style="width: 100px"
                @change="loadLogs"
              />
            </el-form-item>
            <el-form-item label="月份">
              <el-input-number
                v-model="logMonth"
                :min="1"
                :max="12"
                style="width: 80px"
                @change="loadLogs"
              />
            </el-form-item>
            <el-form-item>
              <el-button @click="loadLogs">刷新</el-button>
            </el-form-item>
          </el-form>

          <el-table :data="logs" stripe v-loading="loadingLogs">
            <el-table-column prop="employee_id" label="员工ID" width="80" />
            <el-table-column prop="channel" label="渠道" width="80">
              <template #default="{ row }">
                {{ slipChannelMap[row.channel] }}
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="90">
              <template #default="{ row }">
                <el-tag size="small" :type="slipStatusTagType[row.status]">
                  {{ slipStatusMap[row.status] }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="error_message" label="错误信息" min-width="160" show-overflow-tooltip />
            <el-table-column prop="sent_at" label="发送时间" width="160">
              <template #default="{ row }">{{ row.sent_at || '—' }}</template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="160" />
          </el-table>

          <el-pagination
            class="mt-4"
            layout="total,prev,pager,next"
            :total="logTotal"
            :page="logPage"
            :page-size="logPageSize"
            @current-change="loadLogs"
          />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { salaryApi, type SlipSendLog } from '@/api/salary'

const activeTab = ref('send')
const sendYM = ref('')
const sendChannel = ref('miniapp')

// 全员发送
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

// 选定发送
const employeeList = ref<any[]>([])
const selectedEmployeeIds = ref<number[]>([])
const sendingSelected = ref(false)

async function loadEmployeeList() {
  if (!sendYM.value) return
  const [year, month] = sendYM.value.split('-').map(Number)
  try {
    const res = await salaryApi.list({ year, month, page: 1, page_size: 200 })
    employeeList.value = res.list
  } catch {
    ElMessage.error('加载员工列表失败')
  }
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
      year,
      month,
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

// 发送记录
const loadingLogs = ref(false)
const logs = ref<SlipSendLog[]>([])
const logTotal = ref(0)
const logPage = ref(1)
const logPageSize = ref(20)
const logYear = ref(new Date().getFullYear())
const logMonth = ref(new Date().getMonth() + 1)

const payrollStatusMap: Record<string, string> = {
  draft: '草稿',
  calculated: '已核算',
  confirmed: '已确认',
  paid: '已发放',
}
const payrollStatusTagType: Record<string, string> = {
  draft: 'info',
  calculated: 'warning',
  confirmed: 'primary',
  paid: 'success',
}

const slipChannelMap: Record<string, string> = {
  miniapp: '小程序',
  sms: '短信',
  h5: 'H5链接',
}
const slipStatusMap: Record<string, string> = {
  pending: '等待中',
  sending: '发送中',
  sent: '已发送',
  failed: '失败',
}
const slipStatusTagType: Record<string, string> = {
  pending: 'info',
  sending: 'warning',
  sent: 'success',
  failed: 'danger',
}

async function loadLogs(p = 1) {
  logPage.value = p
  loadingLogs.value = true
  try {
    const res = await salaryApi.getSlipLogs({
      year: logYear.value,
      month: logMonth.value,
      page: p,
      page_size: logPageSize.value,
    })
    logs.value = res.logs
    logTotal.value = res.total
  } catch {
    ElMessage.error('加载发送记录失败')
  } finally {
    loadingLogs.value = false
  }
}

onMounted(() => {
  const now = new Date()
  sendYM.value = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  loadLogs()
  loadEmployeeList()
})
</script>

<style scoped lang="scss">
.salary-slip-send {
  padding: 8px;
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.employee-select {
  margin-top: 12px;
}
.mt-4 {
  margin-top: 16px;
}
</style>
