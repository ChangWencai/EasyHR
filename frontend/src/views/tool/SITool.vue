<template>
  <div class="si-tool">
    <el-tabs v-model="activeTab">
      <!-- Tab 1: 政策库 -->
      <el-tab-pane label="政策库" name="policy">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>社保政策库</span>
              <el-form inline>
                <el-form-item label="城市">
                  <el-input v-model="cityFilter" placeholder="城市名称" clearable style="width: 120px" />
                </el-form-item>
                <el-form-item label="年份">
                  <el-input-number v-model="yearFilter" :min="2020" :max="2030" style="width: 100px" />
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" size="small" @click="loadPolicies">查询</el-button>
                </el-form-item>
              </el-form>
            </div>
          </template>

          <el-table :data="policies" stripe v-loading="loadingPolicy">
            <el-table-column prop="city" label="城市" min-width="100" />
            <el-table-column prop="year" label="年度" width="70" />
            <el-table-column prop="effective_date" label="生效日期" width="110" />
            <el-table-column label="公积金基数范围" min-width="150">
              <template #default="{ row }">
                {{ row.housing_fund_base_min }} ~ {{ row.housing_fund_base_max }}
              </template>
            </el-table-column>
            <el-table-column label="公积金比例" min-width="100">
              <template #default="{ row }">
                个人{{ (row.housing_fund_person_rate * 100).toFixed(1) }}% /
                单位{{ (row.housing_fund_company_rate * 100).toFixed(1) }}%
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-button size="small" @click="selectPolicy(row)">使用</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- Tab 2: 参保操作 -->
      <el-tab-pane label="参保操作" name="enroll">
        <!-- 红色欠缴横幅 -->
        <div v-if="overdueItems.length > 0" class="overdue-banner">
          <el-icon class="banner-icon"><WarningFilled /></el-icon>
          <div class="banner-content">
            <div class="banner-headline">存在欠缴记录，请及时处理</div>
            <div class="banner-detail">
              最大欠缴：{{ overdueItems[0].employeeName }} {{ overdueItems[0].city }} {{ overdueItems[0].yearMonth }} 欠缴 ¥{{ overdueItems[0].amount }}
            </div>
            <div v-if="visibleOverdueItems.length > 1" class="banner-scroll-list">
              <div
                v-for="item in visibleOverdueItems.slice(1)"
                :key="item.id"
                class="overdue-item"
              >
                {{ item.employeeName }} {{ item.city }} {{ item.yearMonth }} ¥{{ item.amount }}
              </div>
            </div>
            <div v-if="overdueItems.length > 5" class="banner-more">
              还有 {{ overdueItems.length - 5 }} 项
            </div>
          </div>
          <el-icon class="close-btn" @click="dismissBanner"><Close /></el-icon>
        </div>

        <el-card>
          <template #header>
            <div class="card-header">
              <span>参保操作</span>
            </div>
          </template>

          <el-form :model="enrollForm" label-width="100px">
            <el-form-item label="选择政策">
              <el-select v-model="enrollForm.policy_id" placeholder="请先在政策库选择" style="width: 300px">
                <el-option
                  v-for="p in policies"
                  :key="p.id"
                  :label="`${p.city} ${p.year}年`"
                  :value="p.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="选择员工">
              <el-select
                v-model="enrollForm.employee_ids"
                multiple
                placeholder="选择参保员工"
                style="width: 300px"
              >
                <el-option
                  v-for="e in employeeOptions"
                  :key="e.id"
                  :label="e.name"
                  :value="e.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="社保基数">
              <el-input-number v-model="enrollForm.salary_base" :min="0" :precision="2" style="width: 200px" />
            </el-form-item>
            <el-form-item label="参保月份">
              <el-date-picker
                v-model="enrollForm.start_month"
                type="month"
                placeholder="参保起始月份"
                value-format="YYYY-MM"
                style="width: 200px"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="previewing" @click="previewEnroll">参保预览</el-button>
              <el-button type="success" :loading="enrolling" @click="handleEnroll">确认参保</el-button>
              <el-button type="primary" @click="showEnrollDialog = true">增员</el-button>
              <el-button type="danger" :loading="stopping" @click="showStopDialog = true">批量停缴</el-button>
            </el-form-item>
          </el-form>

          <!-- 预览结果 -->
          <div v-if="previewResults.length > 0" class="preview-section">
            <div class="preview-title">参保预览</div>
            <el-table :data="previewResults" stripe size="small">
              <el-table-column prop="employee_name" label="员工" />
              <el-table-column prop="salary_base" label="社保基数" />
              <el-table-column label="个人合计">
                <template #default="{ row }">¥{{ row.calculation.total_personal }}</template>
              </el-table-column>
              <el-table-column label="单位合计">
                <template #default="{ row }">¥{{ row.calculation.total_company }}</template>
              </el-table-column>
            </el-table>
          </div>
        </el-card>
      </el-tab-pane>

      <!-- Tab 3: 参保记录 -->
      <el-tab-pane label="参保记录" name="records">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>参保记录</span>
            </div>
          </template>

          <el-table :data="siRecords" stripe v-loading="loadingRecords">
            <el-table-column prop="employee_name" label="员工" min-width="80" />
            <el-table-column prop="city" label="城市" min-width="80" />
            <el-table-column prop="salary_base" label="社保基数" min-width="100">
              <template #default="{ row }">¥{{ row.salary_base }}</template>
            </el-table-column>
            <el-table-column prop="start_month" label="参保月" min-width="90" />
            <el-table-column prop="stop_month" label="停缴月" min-width="90">
              <template #default="{ row }">{{ row.stop_month || '—' }}</template>
            </el-table-column>
            <el-table-column prop="status" label="状态" min-width="70">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
                  {{ row.status === 'active' ? '参保中' : '已停缴' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="个人月缴" min-width="100">
              <template #default="{ row }">¥{{ row.monthly_personal }}</template>
            </el-table-column>
            <el-table-column label="单位月缴" min-width="100">
              <template #default="{ row }">¥{{ row.monthly_company }}</template>
            </el-table-column>
          </el-table>

          <el-pagination
            class="mt-4"
            layout="total,prev,pager,next"
            :total="recordTotal"
            :page="recordPage"
            :page-size="recordPageSize"
            @current-change="loadRecords"
          />
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 增员弹窗 -->
    <EnrollDialog
      v-model="showEnrollDialog"
      @success="onEnrollSuccess"
    />

    <!-- 批量停缴对话框 -->
    <el-dialog v-model="showStopDialog" title="批量停缴" width="400px">
      <el-form :model="stopForm" label-width="100px">
        <el-form-item label="选择记录">
          <el-select v-model="stopForm.record_ids" multiple placeholder="选择要停缴的记录" style="width: 100%">
            <el-option
              v-for="r in siRecords.filter(r => r.status === 'active')"
              :key="r.id"
              :label="`${r.employee_name} - ${r.city}`"
              :value="r.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="停缴月份">
          <el-date-picker
            v-model="stopForm.stop_month"
            type="month"
            placeholder="停缴月份"
            value-format="YYYY-MM"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showStopDialog = false">取消</el-button>
        <el-button type="danger" :loading="stopping" @click="handleStop">确认停缴</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { siApi } from '@/api/socialinsurance'
import { employeeApi } from '@/api/employee'
import { ElMessage } from 'element-plus'
import { WarningFilled, Close } from '@element-plus/icons-vue'
import EnrollDialog from '@/components/socialinsurance/EnrollDialog.vue'
import axios from '@/api/request'

interface OverdueItem {
  id: number
  employeeName: string
  city: string
  yearMonth: string
  amount: string
}

const activeTab = ref('policy')

// Overdue items
const overdueItems = ref<OverdueItem[]>([])

const visibleOverdueItems = computed(() =>
  overdueItems.value.slice(0, 5)
)

function dismissBanner(): void {
  overdueItems.value = []
}

async function loadOverdueItems(): Promise<void> {
  try {
    const res = await axios.get('/api/v1/social-insurance/dashboard')
    const responseData = (res as { data?: Record<string, unknown> })?.data ?? res
    const dashboard = responseData as Record<string, unknown>
    if (Array.isArray(dashboard.overdue_items)) {
      overdueItems.value = dashboard.overdue_items as OverdueItem[]
    }
  } catch {
    // Dashboard loading failure should not block the page
  }
}

// Policy
const loadingPolicy = ref(false)
const policies = ref<any[]>([])
const cityFilter = ref('')
const yearFilter = ref(new Date().getFullYear())

async function loadPolicies() {
  loadingPolicy.value = true
  try {
    policies.value = await siApi.policies({ city: cityFilter.value || undefined, year: yearFilter.value })
  } catch {
    ElMessage.error('加载政策库失败')
  } finally {
    loadingPolicy.value = false
  }
}

function selectPolicy(policy: any) {
  enrollForm.policy_id = policy.id
  enrollForm.salary_base = policy.pension_base_min
  activeTab.value = 'enroll'
}

// Enroll (legacy form)
const previewing = ref(false)
const enrolling = ref(false)
const previewResults = ref<any[]>([])
const employeeOptions = ref<any[]>([])

const enrollForm = reactive({
  policy_id: undefined as number | undefined,
  employee_ids: [] as number[],
  salary_base: 0,
  start_month: '',
})

async function loadEmployees() {
  try {
    const res = await employeeApi.list({ page: 1, page_size: 100 })
    employeeOptions.value = res.list
  } catch {
    // ignore
  }
}

async function previewEnroll() {
  if (!enrollForm.policy_id || enrollForm.employee_ids.length === 0 || !enrollForm.salary_base) {
    ElMessage.warning('请填写完整信息')
    return
  }
  previewing.value = true
  try {
    previewResults.value = await siApi.enrollPreview({
      employee_ids: enrollForm.employee_ids,
      policy_id: enrollForm.policy_id,
      salary_base: enrollForm.salary_base,
    })
  } catch {
    ElMessage.error('预览失败')
  } finally {
    previewing.value = false
  }
}

async function handleEnroll() {
  if (enrolling.value) return
  if (!enrollForm.policy_id || enrollForm.employee_ids.length === 0 || !enrollForm.start_month) {
    ElMessage.warning('请填写完整信息')
    return
  }
  enrolling.value = true
  try {
    await siApi.enroll({
      employee_ids: enrollForm.employee_ids,
      policy_id: enrollForm.policy_id,
      salary_base: enrollForm.salary_base,
      start_month: enrollForm.start_month,
    })
    ElMessage.success('参保成功')
    previewResults.value = []
    loadRecords()
  } catch {
    ElMessage.error('参保失败')
  } finally {
    enrolling.value = false
  }
}

// EnrollDialog
const showEnrollDialog = ref(false)

function onEnrollSuccess(): void {
  loadRecords()
  loadOverdueItems()
}

// Records
const loadingRecords = ref(false)
const siRecords = ref<any[]>([])
const recordTotal = ref(0)
const recordPage = ref(1)
const recordPageSize = ref(20)

async function loadRecords(p = 1) {
  recordPage.value = p
  loadingRecords.value = true
  try {
    const res = await siApi.records({ page: p, page_size: recordPageSize.value })
    siRecords.value = res.list
    recordTotal.value = res.total
  } catch {
    ElMessage.error('加载记录失败')
  } finally {
    loadingRecords.value = false
  }
}

// Stop
const showStopDialog = ref(false)
const stopping = ref(false)
const stopForm = reactive({ record_ids: [] as number[], stop_month: '' })

async function handleStop() {
  if (stopping.value) return
  if (stopForm.record_ids.length === 0 || !stopForm.stop_month) {
    ElMessage.warning('请选择记录和停缴月份')
    return
  }
  stopping.value = true
  try {
    await siApi.stop({ record_ids: stopForm.record_ids, stop_month: stopForm.stop_month })
    ElMessage.success('停缴成功')
    showStopDialog.value = false
    stopForm.record_ids = []
    stopForm.stop_month = ''
    loadRecords()
  } catch {
    ElMessage.error('停缴失败')
  } finally {
    stopping.value = false
  }
}

onMounted(() => {
  loadPolicies()
  loadEmployees()
  loadOverdueItems()
})
</script>

<style scoped lang="scss">
.si-tool {
  padding: 8px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
}

.preview-section {
  margin-top: 16px;
}

.preview-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 8px;
  color: #333;
}

.mt-4 {
  margin-top: 16px;
}

// 红色欠缴横幅
.overdue-banner {
  background: #ff563015;
  border-left: 3px solid #ff5630;
  border-radius: 8px;
  padding: 12px 16px;
  color: #ff5630;
  font-size: 14px;
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 16px;
}

.banner-icon {
  font-size: 18px;
  margin-top: 2px;
  flex-shrink: 0;
}

.banner-content {
  flex: 1;
  min-width: 0;
}

.banner-headline {
  font-weight: 600;
  margin-bottom: 4px;
}

.banner-detail {
  font-size: 13px;
  line-height: 1.5;
}

.banner-scroll-list {
  max-height: 80px;
  overflow-y: auto;
  margin-top: 4px;
}

.overdue-item {
  font-size: 12px;
  line-height: 1.6;
}

.banner-more {
  font-size: 12px;
  margin-top: 4px;
  font-weight: 500;
}

.close-btn {
  font-size: 16px;
  cursor: pointer;
  flex-shrink: 0;
  margin-top: 2px;

  &:hover {
    opacity: 0.7;
  }
}
</style>
