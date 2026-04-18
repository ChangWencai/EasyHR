<template>
  <div class="salary-list">
    <el-card>
      <template #header>
        <div class="page-header">
          <h1 class="page-title">薪资列表</h1>
          <el-button type="primary" @click="showExportDialog = true">导出 Excel</el-button>
        </div>
      </template>

      <!-- 筛选栏 -->
      <el-form inline @submit.prevent>
        <el-form-item label="选择月份">
          <el-date-picker
            v-model="filterYM"
            type="month"
            placeholder="选择月份"
            value-format="YYYY-MM"
            style="width: 160px"
            @change="loadList"
          />
        </el-form-item>
        <el-form-item label="部门">
          <el-select v-model="filterDeptId" placeholder="全部部门" clearable style="width: 160px" @change="loadList">
            <el-option label="全部部门" :value="undefined" />
            <el-option v-for="dept in departments" :key="dept.id" :label="dept.name" :value="dept.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="搜索员工">
          <el-input v-model="filterKeyword" placeholder="搜索员工姓名" clearable style="width: 160px" @input="debouncedLoad" />
        </el-form-item>
      </el-form>

      <!-- 薪资表格 -->
      <el-table :data="listData" stripe v-loading="loading">
        <el-table-column prop="employee_name" label="姓名" min-width="100" />
        <el-table-column prop="department_name" label="部门" min-width="120" />
        <el-table-column prop="gross_income" label="应发合计" width="120">
          <template #default="{ row }">¥{{ (row.gross_income || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2 }) }}</template>
        </el-table-column>
        <el-table-column prop="total_deductions" label="扣除" width="100">
          <template #default="{ row }">¥{{ (row.total_deductions || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2 }) }}</template>
        </el-table-column>
        <el-table-column prop="tax" label="个税" width="100">
          <template #default="{ row }">¥{{ (row.tax || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2 }) }}</template>
        </el-table-column>
        <el-table-column prop="si_deduction" label="社保公积金" width="120">
          <template #default="{ row }">¥{{ (row.si_deduction || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2 }) }}</template>
        </el-table-column>
        <el-table-column prop="net_income" label="实发" width="120">
          <template #default="{ row }">
            <span class="net-income">¥{{ (row.net_income || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2 }) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType[row.status] || 'info'" size="small">
              {{ statusMap[row.status] || row.status }}
            </el-tag>
            <el-icon v-if="['confirmed', 'paid'].includes(row.status)" class="lock-icon" @click="openUnlockDialog(row)">
              <Lock />
            </el-icon>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="mt-4"
        layout="total,prev,pager,next"
        :total="total"
        :page="page"
        :page-size="pageSize"
        @current-change="loadList"
      />
    </el-card>

    <!-- 解锁弹窗 -->
    <el-dialog v-model="showUnlockDialog" title="解锁工资记录" width="420px">
      <el-alert type="warning" :closable="false" style="margin-bottom: 16px">
        该月工资已确认锁定，如需修改请输入解锁码
      </el-alert>
      <el-form @submit.prevent="sendUnlockCode">
        <el-form-item label="手机号">
          <el-input v-model="unlockPhone" placeholder="企业主手机号" style="width: 220px" />
          <el-button style="margin-left: 8px" :loading="sendingCode" @click="sendUnlockCode">
            发送验证码
          </el-button>
        </el-form-item>
        <el-form-item label="验证码" v-if="codeSent">
          <el-input v-model="unlockCode" placeholder="请输入6位验证码" maxlength="6" style="width: 220px" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUnlockDialog = false">取消</el-button>
        <el-button type="primary" :loading="unlocking" :disabled="!codeSent" @click="doUnlock">
          确认解锁
        </el-button>
      </template>
    </el-dialog>

    <!-- 导出弹窗 -->
    <el-dialog v-model="showExportDialog" title="选择导出内容" width="420px">
      <el-form>
        <el-form-item>
          <el-radio-group v-model="exportType">
            <el-radio value="current">当前页（{{ listData.length }} 条）</el-radio>
            <el-radio value="full">全部数据（含税前明细）</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showExportDialog = false">取消</el-button>
        <el-button type="primary" :loading="exporting" @click="doExport">导出</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { salaryApi } from '@/api/salary'
import { departmentApi } from '@/api/department'
import { ElMessage } from 'element-plus'
import { Lock } from '@element-plus/icons-vue'
import { useThrottleFn } from '@vueuse/core'
import dayjs from 'dayjs'

interface SalaryRecord {
  id: number
  employee_id: number
  employee_name: string
  department_name: string
  gross_income: number
  total_deductions: number
  tax: number
  si_deduction: number
  net_income: number
  status: string
}

const filterYM = ref(dayjs().format('YYYY-MM'))
const filterDeptId = ref<number | undefined>(undefined)
const filterKeyword = ref('')
const loading = ref(false)
const listData = ref<SalaryRecord[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const departments = ref<{ id: number; name: string }[]>([])

const statusMap: Record<string, string> = {
  draft: '草稿',
  calculated: '已核算',
  confirmed: '已确认',
  paid: '已发放',
}
const statusTagType: Record<string, string> = {
  draft: 'info',
  calculated: 'warning',
  confirmed: 'primary',
  paid: 'success',
}

// Unlock
const showUnlockDialog = ref(false)
const unlockTarget = ref<SalaryRecord | null>(null)
const unlockPhone = ref('')
const unlockCode = ref('')
const sendingCode = ref(false)
const codeSent = ref(false)
const unlocking = ref(false)

// Export
const showExportDialog = ref(false)
const exportType = ref('current')
const exporting = ref(false)

async function loadDepartments() {
  try {
    const res = await departmentApi.list()
    departments.value = (res as any[]).map((d) => ({ id: d.id, name: d.name }))
  } catch {
    // ignore
  }
}

async function loadList(p = 1) {
  if (!filterYM.value) return
  page.value = p
  loading.value = true
  try {
    const [yearStr, monthStr] = filterYM.value.split('-')
    const year = parseInt(yearStr, 10)
    const month = parseInt(monthStr, 10)
    const res = await salaryApi.getSalaryList({
      year,
      month,
      department_id: filterDeptId.value,
      keyword: filterKeyword.value || undefined,
      page: p,
      page_size: pageSize.value,
    })
    const data = (res as any)
    listData.value = data.list || []
    total.value = data.total || 0
  } catch {
    ElMessage.error('加载薪资列表失败')
  } finally {
    loading.value = false
  }
}

const debouncedLoad = useThrottleFn(() => loadList(), 400)

function openUnlockDialog(row: SalaryRecord) {
  unlockTarget.value = row
  unlockPhone.value = ''
  unlockCode.value = ''
  codeSent.value = false
  showUnlockDialog.value = true
}

async function sendUnlockCode() {
  if (!unlockPhone.value) {
    ElMessage.warning('请输入手机号')
    return
  }
  sendingCode.value = true
  try {
    await salaryApi.sendUnlockCode({ phone: unlockPhone.value })
    codeSent.value = true
    ElMessage.success('验证码已发送')
  } catch {
    ElMessage.error('发送失败')
  } finally {
    sendingCode.value = false
  }
}

async function doUnlock() {
  if (!unlockTarget.value || !unlockCode.value) return
  unlocking.value = true
  try {
    await salaryApi.unlockRecord({
      record_id: unlockTarget.value.id,
      sms_code: unlockCode.value,
    })
    ElMessage.success('已解锁，可重新编辑')
    showUnlockDialog.value = false
    loadList()
  } catch {
    ElMessage.error('解锁失败，验证码错误')
  } finally {
    unlocking.value = false
  }
}

async function doExport() {
  if (!filterYM.value) return
  const [yearStr, monthStr] = filterYM.value.split('-')
  const year = parseInt(yearStr, 10)
  const month = parseInt(monthStr, 10)
  exporting.value = true
  try {
    const blob = await salaryApi.exportWithDetails(year, month, exportType.value === 'full')
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `工资表_${year}年${month}月.xlsx`
    a.click()
    URL.revokeObjectURL(url)
    showExportDialog.value = false
    ElMessage.success('导出成功')
  } catch {
    ElMessage.error('导出失败')
  } finally {
    exporting.value = false
  }
}

onMounted(async () => {
  await loadDepartments()
  await loadList()
})
</script>

<style scoped lang="scss">
.salary-list {
  padding: 8px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title {
  font-size: 16px;
  font-weight: 700;
  color: #1a1a1a;
  margin: 0;
}

.net-income {
  font-weight: 700;
  color: #52c41a;
}

.lock-icon {
  margin-left: 4px;
  cursor: pointer;
  color: #8c8c8c;
  vertical-align: middle;

  &:hover {
    color: #1677ff;
  }
}

.mt-4 {
  margin-top: 16px;
}
</style>
