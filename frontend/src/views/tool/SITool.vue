<template>
  <div class="si-tool">
    <!-- Page Header -->
    <div class="page-head">
      <div class="page-head-left">
        <div class="page-head-indicator"></div>
        <div>
          <h2 class="page-head-title">社保管理</h2>
          <p class="page-head-desc">政策库、参保操作与记录查询</p>
        </div>
      </div>
    </div>

    <!-- Tab Navigation -->
    <div class="tab-shell">
      <div class="tab-rail">
        <div
          v-for="tab in tabs"
          :key="tab.name"
          class="tab-btn"
          :class="{ 'tab-btn--active': activeTab === tab.name }"
          @click="activeTab = tab.name"
        >
          <el-icon :size="16"><component :is="tab.icon" /></el-icon>
          <span>{{ tab.label }}</span>
        </div>
      </div>

      <div class="tab-body">
        <!-- 政策库 -->
        <div v-show="activeTab === 'policy'" class="tab-panel">
          <div class="panel-bar">
            <span class="panel-title">社保政策库</span>
            <el-form inline>
              <el-form-item label="年份">
                <el-input-number v-model="yearFilter" :min="2020" :max="2030" style="width: 100px" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="loadPolicies">查询</el-button>
              </el-form-item>
            </el-form>
          </div>
          <el-table :data="policies" stripe v-loading="loadingPolicy">
            <el-table-column label="城市" min-width="100">
              <template #default="{ row }">{{ row.city_name || row.city || '未知城市' }}</template>
            </el-table-column>
            <el-table-column label="年度" width="70">
              <template #default="{ row }">{{ row.effective_year }}</template>
            </el-table-column>
            <el-table-column label="生效日期" width="110">
              <template #default="{ row }">{{ row.created_at?.substring(0, 10) || '-' }}</template>
            </el-table-column>
            <el-table-column label="公积金基数范围" min-width="150">
              <template #default="{ row }">{{ row.config?.housing_fund?.base_lower || 0 }} ~ {{ row.config?.housing_fund?.base_upper || 0 }}</template>
            </el-table-column>
            <el-table-column label="公积金比例" min-width="100">
              <template #default="{ row }">个人{{ ((row.config?.housing_fund?.personal_rate || 0) * 100).toFixed(1) }}% / 单位{{ ((row.config?.housing_fund?.company_rate || 0) * 100).toFixed(1) }}%</template>
            </el-table-column>
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-button size="small" type="primary" plain @click="selectPolicy(row)">使用</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 参保操作 -->
        <div v-show="activeTab === 'enroll'" class="tab-panel">
          <div v-if="overdueItems.length > 0" class="overdue-banner">
            <div class="overdue-banner-glow"></div>
            <el-icon class="banner-icon"><WarningFilled /></el-icon>
            <div class="banner-content">
              <div class="banner-headline">存在欠缴记录，请及时处理</div>
              <div class="banner-detail">{{ overdueItems[0].employee_name }} {{ overdueItems[0].city }} {{ overdueItems[0].year_month }} 欠缴 ¥{{ overdueItems[0].amount }}</div>
              <div v-if="visibleOverdueItems.length > 1" class="banner-scroll-list">
                <div v-for="item in visibleOverdueItems.slice(1)" :key="item.id" class="overdue-item">{{ item.employee_name }} {{ item.city }} {{ item.year_month }} ¥{{ item.amount }}</div>
              </div>
              <div v-if="overdueItems.length > 5" class="banner-more">还有 {{ overdueItems.length - 5 }} 项</div>
            </div>
            <el-icon class="close-btn" @click="dismissBanner"><Close /></el-icon>
          </div>

          <div class="panel-bar">
            <span class="panel-title">参保操作</span>
          </div>
          <el-form :model="enrollForm" label-width="100px">
            <el-form-item label="选择政策">
              <el-select v-model="enrollForm.policy_id" placeholder="请先在政策库选择" style="width: 300px">
                <el-option v-for="p in policies" :key="p.id" :label="`${p.city} ${p.year}年`" :value="p.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="选择员工">
              <el-select v-model="enrollForm.employee_ids" multiple placeholder="选择参保员工" style="width: 300px">
                <el-option v-for="e in employeeOptions" :key="e.id" :label="e.name" :value="e.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="社保基数">
              <el-input-number v-model="enrollForm.salary_base" :min="0" :precision="2" style="width: 200px" />
            </el-form-item>
            <el-form-item label="参保月份">
              <el-date-picker v-model="enrollForm.start_month" type="month" placeholder="参保起始月份" value-format="YYYY-MM" style="width: 200px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="previewing" @click="previewEnroll">参保预览</el-button>
              <el-button type="success" :loading="enrolling" @click="handleEnroll">确认参保</el-button>
              <el-button type="primary" @click="showEnrollDialog = true">增员</el-button>
              <el-button type="danger" plain :loading="stopping" @click="showStopDialog = true">批量停缴</el-button>
            </el-form-item>
          </el-form>

          <div v-if="previewResults.length > 0" class="preview-section">
            <div class="preview-title">参保预览</div>
            <el-table :data="previewResults" stripe size="small">
              <el-table-column prop="employee_name" label="员工" />
              <el-table-column prop="salary_base" label="社保基数" />
              <el-table-column label="个人合计"><template #default="{ row }">¥{{ row.calculation.total_personal }}</template></el-table-column>
              <el-table-column label="单位合计"><template #default="{ row }">¥{{ row.calculation.total_company }}</template></el-table-column>
            </el-table>
          </div>
        </div>

        <!-- 参保记录 -->
        <div v-show="activeTab === 'records'" class="tab-panel">
          <div class="panel-bar">
            <span class="panel-title">参保记录</span>
          </div>
          <el-table :data="siRecords" stripe v-loading="loadingRecords">
            <el-table-column prop="employee_name" label="员工" min-width="80" />
            <el-table-column prop="city" label="城市" min-width="80" />
            <el-table-column prop="salary_base" label="社保基数" min-width="100"><template #default="{ row }">¥{{ row.salary_base }}</template></el-table-column>
            <el-table-column prop="start_month" label="参保月" min-width="90" />
            <el-table-column prop="stop_month" label="停缴月" min-width="90"><template #default="{ row }">{{ row.stop_month || '—' }}</template></el-table-column>
            <el-table-column prop="status" label="状态" min-width="70">
              <template #default="{ row }"><el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">{{ row.status === 'active' ? '参保中' : '已停缴' }}</el-tag></template>
            </el-table-column>
            <el-table-column label="个人月缴" min-width="100"><template #default="{ row }">¥{{ row.monthly_personal }}</template></el-table-column>
            <el-table-column label="单位月缴" min-width="100"><template #default="{ row }">¥{{ row.monthly_company }}</template></el-table-column>
          </el-table>
          <el-pagination class="mt-4" layout="total,prev,pager,next" :total="recordTotal" :page="recordPage" :page-size="recordPageSize" @current-change="loadRecords" />
        </div>
      </div>
    </div>

    <EnrollDialog v-model="showEnrollDialog" @success="onEnrollSuccess" />

    <el-dialog v-model="showStopDialog" title="批量停缴" width="400px">
      <el-form :model="stopForm" label-width="100px">
        <el-form-item label="选择记录">
          <el-select v-model="stopForm.record_ids" multiple placeholder="选择要停缴的记录" style="width: 100%">
            <el-option v-for="r in siRecords.filter(r => r.status === 'active')" :key="r.id" :label="`${r.employee_name} - ${r.city}`" :value="r.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="停缴月份">
          <el-date-picker v-model="stopForm.stop_month" type="month" placeholder="停缴月份" value-format="YYYY-MM" style="width: 100%" />
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
import { WarningFilled, Close, Collection, Plus, List } from '@element-plus/icons-vue'
import EnrollDialog from '@/components/socialinsurance/EnrollDialog.vue'

interface OverdueItem {
  id: number
  employee_name: string
  city: string
  year_month: string
  amount: string
}

const activeTab = ref('policy')

const tabs = [
  { name: 'policy', label: '政策库', icon: Collection },
  { name: 'enroll', label: '参保操作', icon: Plus },
  { name: 'records', label: '参保记录', icon: List },
]

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
    const dashboard = await siApi.dashboard()
    if (Array.isArray(dashboard?.overdue_items)) {
      overdueItems.value = dashboard.overdue_items
    }
  } catch {
    // Dashboard loading failure should not block the page
  }
}

// Policy
const loadingPolicy = ref(false)
const policies = ref<any[]>([])
const cityFilter = ref<number | undefined>()
const yearFilter = ref(new Date().getFullYear())

async function loadPolicies() {
  loadingPolicy.value = true
  try {
    policies.value = (await siApi.policies({ city_code: cityFilter.value || undefined, year: yearFilter.value })) ?? []
  } catch {
    ElMessage.error('加载政策库失败')
  } finally {
    loadingPolicy.value = false
  }
}

function selectPolicy(policy: any) {
  enrollForm.policy_id = policy.id
  enrollForm.salary_base = policy.config?.pension?.base_lower || 0
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
    const res = await employeeApi.list({ page: 1, page_size: 100 }) as { list: any[] }
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
    previewResults.value = (await siApi.enrollPreview({
      employee_ids: enrollForm.employee_ids,
      policy_id: enrollForm.policy_id,
      salary_base: enrollForm.salary_base,
    }))
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
$primary: #7C3AED;
$primary-light: #A78BFA;
$text-primary: #1A1D2E;
$text-secondary: #5E6278;
$text-muted: #A0A3BD;
$border: #E8EBF0;
$surface: #FFFFFF;
$surface-alt: #F8F9FC;

.si-tool {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* ─── Page Header ─── */
.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.page-head-left {
  display: flex;
  align-items: center;
  gap: 14px;
}

.page-head-indicator {
  width: 4px;
  height: 36px;
  border-radius: 4px;
  background: linear-gradient(180deg, #2563EB 0%, #60A5FA 100%);
}

.page-head-title {
  font-size: 20px;
  font-weight: 700;
  color: $text-primary;
  margin: 0;
  letter-spacing: -0.3px;
}

.page-head-desc {
  font-size: 13px;
  color: $text-muted;
  margin: 4px 0 0;
}

/* ─── Tab Shell ─── */
.tab-shell {
  background: $surface;
  border: 1px solid $border;
  border-radius: 20px;
  overflow: hidden;
}

.tab-rail {
  display: flex;
  gap: 2px;
  padding: 10px 12px;
  border-bottom: 1px solid $border;
  overflow-x: auto;
  background: $surface-alt;

  &::-webkit-scrollbar {
    height: 0;
  }
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  color: $text-secondary;
  cursor: pointer;
  transition: all 0.25s ease;
  white-space: nowrap;
  user-select: none;

  &:hover {
    background: rgba($primary, 0.06);
    color: $primary;
  }

  &--active {
    background: $surface;
    color: $primary;
    font-weight: 600;
    box-shadow: 0 1px 4px rgba(0,0,0,0.06);
  }
}

.tab-body {
  padding: 0;
}

.tab-panel {
  padding: 24px;
  animation: panelFade 0.3s ease;
}

@keyframes panelFade {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}

.panel-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
}

.panel-title {
  font-size: 15px;
  font-weight: 600;
  color: $text-primary;
}

/* ─── Overdue Banner ─── */
.overdue-banner {
  position: relative;
  background: linear-gradient(135deg, rgba(255,86,48,0.06) 0%, rgba(255,86,48,0.02) 100%);
  border: 1px solid rgba(255,86,48,0.15);
  border-left: 3px solid #FF5630;
  border-radius: 14px;
  padding: 16px 20px;
  color: #FF5630;
  font-size: 14px;
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 20px;
  overflow: hidden;
}

.overdue-banner-glow {
  position: absolute;
  top: -20px;
  right: -20px;
  width: 80px;
  height: 80px;
  background: radial-gradient(circle, rgba(255,86,48,0.08) 0%, transparent 70%);
  pointer-events: none;
}

.banner-icon {
  font-size: 20px;
  margin-top: 2px;
  flex-shrink: 0;
}

.banner-content {
  flex: 1;
  min-width: 0;
}

.banner-headline {
  font-weight: 700;
  margin-bottom: 4px;
  font-size: 14px;
}

.banner-detail {
  font-size: 13px;
  line-height: 1.5;
}

.banner-scroll-list {
  max-height: 80px;
  overflow-y: auto;
  margin-top: 6px;
}

.overdue-item {
  font-size: 12px;
  line-height: 1.8;
  opacity: 0.85;
}

.banner-more {
  font-size: 12px;
  margin-top: 6px;
  font-weight: 600;
}

.close-btn {
  font-size: 16px;
  cursor: pointer;
  flex-shrink: 0;
  margin-top: 2px;
  transition: opacity 0.2s;

  &:hover { opacity: 0.6; }
}

/* ─── Preview ─── */
.preview-section {
  margin-top: 20px;
}

.preview-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 10px;
  color: $text-primary;
}

.mt-4 {
  margin-top: 16px;
}
</style>
