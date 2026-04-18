<template>
  <div class="salary-tool">
    <el-tabs v-model="activeTab">
      <!-- Tab 1: 数据看板 -->
      <el-tab-pane label="数据看板" name="dashboard">
        <SalaryDashboard />
      </el-tab-pane>

      <!-- Tab 2: 薪资模板 -->
      <el-tab-pane label="薪资模板" name="template">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>薪资模板配置</span>
              <el-button type="primary" size="small" :loading="savingTemplate" @click="saveTemplate">
                保存配置
              </el-button>
            </div>
          </template>

          <div v-loading="loadingTemplate">
            <el-table :data="templateItems" stripe>
              <el-table-column prop="name" label="项目名称" min-width="120" />
              <el-table-column prop="category" label="类别" width="100">
                <template #default="{ row }">{{ row.type === 'earning' ? '应发' : '扣款' }}</template>
              </el-table-column>
              <el-table-column prop="description" label="说明" min-width="160" show-overflow-tooltip />
              <el-table-column label="启用" width="80">
                <template #default="{ row }">
                  <el-switch v-model="row.is_enabled" :disabled="row.is_default" />
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-card>
      </el-tab-pane>

      <!-- Tab 3: 工资核算 -->
      <el-tab-pane label="工资核算" name="payroll">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>工资核算</span>
            </div>
          </template>

          <el-form inline @submit.prevent>
            <el-form-item label="年份">
              <el-date-picker
                v-model="payrollYM"
                type="month"
                placeholder="选择月份"
                value-format="YYYY-MM"
                style="width: 140px"
              />
            </el-form-item>
            <el-form-item>
              <el-checkbox v-model="copyFromPrev">复制上月数据</el-checkbox>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="creating" @click="createPayroll">创建工资表</el-button>
              <el-button :loading="calculating" @click="calculatePayroll">一键核算</el-button>
              <el-button type="success" :loading="confirming" @click="confirmPayroll">确认锁定</el-button>
            </el-form-item>
          </el-form>

          <el-table :data="payrollList" stripe v-loading="loadingPayroll">
            <el-table-column prop="year" label="年份" width="70" />
            <el-table-column prop="month" label="月份" width="70" />
            <el-table-column prop="employee_count" label="人数" width="70" />
            <el-table-column prop="total_gross" label="应发总额" width="110">
              <template #default="{ row }">¥{{ row.total_gross }}</template>
            </el-table-column>
            <el-table-column prop="total_deduction" label="扣款总额" width="110">
              <template #default="{ row }">¥{{ row.total_deduction }}</template>
            </el-table-column>
            <el-table-column prop="total_net" label="实发总额" width="110">
              <template #default="{ row }">¥{{ row.total_net }}</template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="90">
              <template #default="{ row }">
                <el-tag :type="payrollStatusTagType[row.status]" size="small">
                  {{ payrollStatusMap[row.status] }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="viewPayrollDetail(row.id)">查看明细</el-button>
                <el-button
                  v-if="row.status === 'confirmed'"
                  size="small"
                  type="success"
                  @click="recordPayment(row)"
                >
                  登记发放
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            class="mt-4"
            layout="total,prev,pager,next"
            :total="payrollTotal"
            :page="payrollPage"
            :page-size="payrollPageSize"
            @current-change="loadPayroll"
          />
        </el-card>
      </el-tab-pane>

      <!-- Tab 4: 调薪管理 -->
      <el-tab-pane label="调薪管理" name="adjustment">
        <SalaryAdjustment />
      </el-tab-pane>

      <!-- Tab 5: 导出 -->
      <el-tab-pane label="导出" name="export">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>工资表导出</span>
            </div>
          </template>

          <el-form inline @submit.prevent>
            <el-form-item label="选择月份">
              <el-date-picker
                v-model="exportYM"
                type="month"
                placeholder="选择月份"
                value-format="YYYY-MM"
                style="width: 140px"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :disabled="!exportYM" :loading="exporting" @click="handleExport">
                下载Excel
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- Tab 6: 个税上传 -->
      <el-tab-pane label="个税上传" name="tax-upload">
        <TaxUpload />
      </el-tab-pane>

      <!-- Tab 7: 工资条发送 -->
      <el-tab-pane label="工资条发送" name="slip-send">
        <SalarySlipSend />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { salaryApi } from '@/api/salary'
import { ElMessage } from 'element-plus'
import SalaryDashboard from './SalaryDashboard.vue'
import SalaryAdjustment from './SalaryAdjustment.vue'
import SalaryList from './SalaryList.vue'
import TaxUpload from './TaxUpload.vue'
import SalarySlipSend from './SalarySlipSend.vue'

const activeTab = ref('dashboard')

// Template
const loadingTemplate = ref(false)
const savingTemplate = ref(false)
const templateItems = ref<any[]>([])

async function loadTemplate() {
  loadingTemplate.value = true
  try {
    const tpl = await salaryApi.template()
    templateItems.value = tpl.items
  } catch {
    ElMessage.error('加载模板失败')
  } finally {
    loadingTemplate.value = false
  }
}

async function saveTemplate() {
  savingTemplate.value = true
  try {
    const items = templateItems.value.map((i) => ({ id: i.id, is_enabled: i.is_enabled }))
    await salaryApi.updateTemplate({ items })
    ElMessage.success('保存成功')
  } catch {
    ElMessage.error('保存失败')
  } finally {
    savingTemplate.value = false
  }
}

// Payroll
const payrollYM = ref('')
const copyFromPrev = ref(false)
const creating = ref(false)
const calculating = ref(false)
const confirming = ref(false)
const loadingPayroll = ref(false)
const payrollList = ref<any[]>([])
const payrollTotal = ref(0)
const payrollPage = ref(1)
const payrollPageSize = ref(20)

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

async function loadPayroll(p = 1) {
  payrollPage.value = p
  if (!payrollYM.value) return
  const [year, month] = payrollYM.value.split('-').map(Number)
  loadingPayroll.value = true
  try {
    const res = await salaryApi.list({ year, month, page: p, page_size: payrollPageSize.value })
    payrollList.value = res.list
    payrollTotal.value = res.total
  } catch {
    ElMessage.error('加载工资表失败')
  } finally {
    loadingPayroll.value = false
  }
}

async function createPayroll() {
  if (!payrollYM.value) {
    ElMessage.warning('请先选择月份')
    return
  }
  const [year, month] = payrollYM.value.split('-').map(Number)
  creating.value = true
  try {
    await salaryApi.createPayroll({
      year,
      month,
      copy_from_month: copyFromPrev.value ? `${year}-${String(month).padStart(2, '0')}` : undefined,
    })
    ElMessage.success('工资表已创建')
    loadPayroll()
  } catch {
    ElMessage.error('创建失败')
  } finally {
    creating.value = false
  }
}

async function calculatePayroll() {
  if (!payrollYM.value) {
    ElMessage.warning('请先选择月份')
    return
  }
  const [year, month] = payrollYM.value.split('-').map(Number)
  calculating.value = true
  try {
    await salaryApi.calculatePayroll({ year, month })
    ElMessage.success('核算完成')
    loadPayroll()
  } catch {
    ElMessage.error('核算失败')
  } finally {
    calculating.value = false
  }
}

async function confirmPayroll() {
  if (!payrollYM.value) {
    ElMessage.warning('请先选择月份')
    return
  }
  const [year, month] = payrollYM.value.split('-').map(Number)
  confirming.value = true
  try {
    await salaryApi.confirmPayroll({ year, month })
    ElMessage.success('已确认锁定')
    loadPayroll()
  } catch {
    ElMessage.error('操作失败')
  } finally {
    confirming.value = false
  }
}

function viewPayrollDetail(id: number) {
  // TODO: navigate to payroll detail
  ElMessage.info(`查看工资表明细 ID=${id}`)
}

async function recordPayment(row: any) {
  try {
    await salaryApi.recordPayment(row.id, {
      method: 'bank_transfer',
      paid_at: new Date().toISOString().slice(0, 10),
    })
    ElMessage.success('已登记发放')
    loadPayroll()
  } catch {
    ElMessage.error('操作失败')
  }
}

// Export
const exportYM = ref('')
const exporting = ref(false)

async function handleExport() {
  if (!exportYM.value) return
  const [year, month] = exportYM.value.split('-').map(Number)
  exporting.value = true
  try {
    const blob = await salaryApi.export(year, month)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `工资表_${year}年${month}月.xlsx`
    a.click()
    URL.revokeObjectURL(url)
  } catch {
    ElMessage.error('导出失败')
  } finally {
    exporting.value = false
  }
}

onMounted(() => {
  loadTemplate()
  const now = new Date()
  payrollYM.value = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
})
</script>

<style scoped lang="scss">
.salary-tool {
  padding: 8px;
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.mt-4 {
  margin-top: 16px;
}
</style>
