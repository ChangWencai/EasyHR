<template>
  <div class="page-view">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">员工管理</h1>
        <p class="page-subtitle">共 {{ total }} 名员工</p>
      </div>
      <div class="header-actions">
        <el-button @click="showBatchImport = true">
          <el-icon><Upload /></el-icon>
          批量入职
        </el-button>
        <el-button @click="$router.push('/employee/invitations')">
          <el-icon><Plus /></el-icon>
          入职邀请
        </el-button>
        <el-button @click="$router.push('/employee/offboardings')">
          <el-icon><Delete /></el-icon>
          离职管理
        </el-button>
        <el-button type="primary" @click="$router.push('/employee/create')">
          <el-icon><Plus /></el-icon>
          新增员工
        </el-button>
      </div>
    </header>

    <!-- 搜索筛选栏 -->
    <div class="filter-bar glass-card">
      <div class="search-wrapper">
        <el-icon class="search-icon"><Search /></el-icon>
        <input
          v-model="search"
          type="text"
          placeholder="搜索姓名、手机号、岗位..."
          class="search-input"
          @keyup.enter="load(1)"
        />
        <el-button v-if="search" type="text" class="clear-btn" @click="search = ''; load(1)">
          <el-icon><Close /></el-icon>
        </el-button>
      </div>
      <div class="filter-group">
        <el-select
          v-model="departmentId"
          placeholder="全部部门"
          clearable
          class="filter-select"
          @change="load(1)"
        >
          <template #prefix>
            <el-icon><OfficeBuilding /></el-icon>
          </template>
          <el-option v-for="dept in departments" :key="dept.id" :label="dept.name" :value="dept.id" />
        </el-select>
        <el-button type="primary" @click="load(1)">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
        <el-button @click="handleExport">
          <el-icon><Download /></el-icon>
          导出
        </el-button>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="table-container glass-card">
      <el-table
        :data="list"
        stripe
        v-loading="loading"
        class="modern-table"
        :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
      >
        <el-table-column prop="name" label="姓名" min-width="100" fixed="left">
          <template #header>
            <el-tooltip content="点击查看员工详情" placement="top" :show-after="500">
              <span>姓名</span>
            </el-tooltip>
          </template>
          <template #default="{ row }">
            <div class="employee-cell">
              <el-avatar :size="36" class="employee-avatar">
                {{ row.name?.[0] || '?' }}
              </el-avatar>
              <span class="employee-name">{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <span class="status-badge" :class="`status--${row.status}`">
              {{ statusMap[row.status] }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="department_name" label="部门" min-width="120">
          <template #default="{ row }">
            <span class="department-tag">{{ row.department_name || '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="岗位薪资" min-width="130">
          <template #default="{ row }">
            <span v-if="row.salary_amount > 0" class="salary-value">
              ¥{{ row.salary_amount.toLocaleString() }}
            </span>
            <span v-else class="salary-empty">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="years_of_service" label="在职年限" width="100">
          <template #default="{ row }">
            <span v-if="row.years_of_service" class="years-value">
              {{ row.years_of_service }}年
            </span>
            <span v-else class="salary-empty">—</span>
          </template>
        </el-table-column>
        <el-table-column label="合同到期" min-width="140">
          <template #default="{ row }">
            <template v-if="row.contract_expiry_days !== null && row.contract_expiry_days !== undefined">
              <span
                class="contract-status"
                :class="{
                  'contract--safe': row.contract_expiry_days > 30,
                  'contract--warning': row.contract_expiry_days > 0 && row.contract_expiry_days <= 30,
                  'contract--danger': row.contract_expiry_days === 0 || row.contract_expiry_days < 0
                }"
              >
                <el-icon>
                  <WarningFilled v-if="row.contract_expiry_days <= 30" />
                  <CircleCheckFilled v-else />
                </el-icon>
                <span v-if="row.contract_expiry_days > 0">{{ row.contract_expiry_days }}天后</span>
                <span v-else-if="row.contract_expiry_days === 0">今天到期</span>
                <span v-else>已过期{{ Math.abs(row.contract_expiry_days) }}天</span>
              </span>
            </template>
            <span v-else class="salary-empty">无合同</span>
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="手机号" min-width="130">
          <template #default="{ row }">
            <span class="phone-value">{{ formatPhone(row.phone) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #header>
            <el-tooltip content="编辑、查看员工信息" placement="top" :show-after="500">
              <span>操作</span>
            </el-tooltip>
          </template>
          <template #default="{ row }">
            <div class="action-btns">
              <el-button size="small" text @click="openDrawer(row.id)">
                <el-icon><View /></el-icon>
                详情
              </el-button>
              <el-button size="small" type="primary" @click="$router.push(`/employee/${row.id}/edit`)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          :page-sizes="[10, 20, 50]"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next, sizes"
          small
          @current-change="load"
        />
      </div>
    </div>

    <!-- 员工详情抽屉 -->
    <EmployeeDrawer
      v-model="drawerVisible"
      :employee-id="selectedEmployeeId"
    />

    <!-- 批量入职弹窗 -->
    <el-dialog
      v-model="showBatchImport"
      title="批量入职"
      width="680px"
      :close-on-click-modal="false"
    >
      <ExcelImportWizard
        template-label="员工"
        :template-fields="['姓名', '手机号', '身份证号', '入职日期', '岗位', '薪资']"
        :import-api="batchImportEmployees"
        @complete="handleBatchComplete"
        @update:dialog-visible="showBatchImport = $event"
      />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { employeeApi, type EmployeeRosterItem, batchImportEmployees } from '@/api/employee'
import { departmentApi, type Department } from '@/api/department'
import { ElMessage } from 'element-plus'
import ExcelImportWizard from '@/components/common/ExcelImportWizard.vue'
import {
  Search,
  Close,
  Plus,
  Download,
  Delete,
  OfficeBuilding,
  View,
  Edit,
  Upload,
  WarningFilled,
  CircleCheckFilled,
} from '@element-plus/icons-vue'
import { statusMap } from '@/views/employee/statusMap'
import EmployeeDrawer from '@/views/employee/EmployeeDrawer.vue'

const search = ref('')
const departmentId = ref<number | undefined>(undefined)
const departments = ref<Department[]>([])
const list = ref<EmployeeRosterItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

// Drawer 状态
const drawerVisible = ref(false)
const selectedEmployeeId = ref<number>(0)

// 批量入职弹窗状态
const showBatchImport = ref(false)

function openDrawer(id: number) {
  selectedEmployeeId.value = id
  drawerVisible.value = true
}

async function handleBatchComplete(result: { success: number; failed: number }) {
  showBatchImport.value = false
  await load()
}

function formatPhone(phone: string): string {
  if (!phone || phone.length !== 11) return phone || '—'
  return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
}

async function loadDepartments() {
  try {
    departments.value = await departmentApi.list()
  } catch {
    // 部门加载失败不阻塞主流程
  }
}

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const params: { page: number; page_size: number; search?: string; department_id?: number } = {
      page: p,
      page_size: pageSize.value,
    }
    if (search.value) {
      params.search = search.value
    }
    if (departmentId.value !== undefined) {
      params.department_id = departmentId.value
    }
    const res = await employeeApi.getRoster(params)
    list.value = res.list || []
    total.value = res.total
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

function handleExport() {
  employeeApi.exportExcel().then((blob: Blob) => {
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `员工花名册_${Date.now()}.xlsx`
    a.click()
    URL.revokeObjectURL(url)
  }).catch(() => {
    ElMessage.error('导出失败')
  })
}

onMounted(() => {
  loadDepartments()
  load()
})
</script>

<style scoped lang="scss">
// 筛选栏
.filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  margin-bottom: 20px;
  gap: 16px;
}

.search-wrapper {
  position: relative;
  flex: 1;
  max-width: 400px;
}

.search-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-tertiary);
  font-size: 18px;
}

.search-input {
  width: 100%;
  padding: 10px 40px 10px 42px;
  font-size: 14px;
  color: var(--text-primary);
  background: var(--bg-page);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  outline: none;
  transition: all 0.2s ease;

  &::placeholder {
    color: var(--text-tertiary);
  }

  &:focus {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(var(--primary), 0.1);
  }
}

.clear-btn {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  padding: 4px;
  color: var(--text-tertiary);

  &:hover {
    color: var(--text-secondary);
  }
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-select {
  width: 160px;

  :deep(.el-input__wrapper) {
    border-radius: var(--radius-md);
  }
}

// 数据表格
.table-container {
  padding: 0;
  overflow: hidden;
}

:deep(.modern-table) {
  .el-table__header th {
    padding: 16px 12px;
    font-size: 13px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .el-table__row {
    transition: background 0.2s ease;

    &:hover > td {
      background: rgba(var(--primary), 0.02) !important;
    }
  }

  .el-table__cell {
    padding: 16px 12px;
    border-bottom: 1px solid #F3F4F6;
  }
}

.employee-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.employee-avatar {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, var(--primary-light) 0%, var(--primary) 100%);
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  flex-shrink: 0;
}

.employee-name {
  font-weight: 500;
  color: var(--text-primary);
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 20px;

  &.status--active {
    background: #D1FAE5;
    color: #059669;
  }

  &.status--probation {
    background: #DBEAFE;
    color: #2563EB;
  }

  &.status--leave {
    background: #FEF3C7;
    color: #D97706;
  }

  &.status--resigned {
    background: #F3F4F6;
    color: #6B7280;
  }
}

.department-tag {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  font-size: 12px;
  background: var(--bg-page);
  color: var(--text-secondary);
  border-radius: 6px;
}

.salary-value {
  font-weight: 600;
  color: var(--text-primary);
  font-family: 'SF Mono', Monaco, monospace;
}

.salary-empty {
  color: var(--text-tertiary);
}

.years-value {
  color: var(--text-secondary);
}

.contract-status {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 500;

  &.contract--safe {
    color: var(--success);
  }

  &.contract--warning {
    color: var(--warning);
  }

  &.contract--danger {
    color: var(--danger);
  }
}

.phone-value {
  font-family: 'SF Mono', Monaco, monospace;
  color: var(--text-secondary);
  font-size: 13px;
}

.action-btns {
  display: flex;
  gap: 8px;

  :deep(.el-button) {
    padding: 6px 12px;
    border-radius: var(--radius-sm);
    font-size: 13px;
    display: inline-flex;
    align-items: center;
    gap: 4px;

    .el-icon {
      font-size: 14px;
    }
  }
}

// 分页
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: 16px 20px;
  border-top: 1px solid var(--border);

  :deep(.el-pagination) {
    .el-pager li {
      border-radius: var(--radius-sm);
      margin: 0 2px;
      min-width: 32px;
      height: 32px;
      line-height: 32px;

      &.is-active {
        background: var(--primary);
      }
    }

    .btn-prev,
    .btn-next {
      border-radius: var(--radius-sm);
      min-width: 32px;
      height: 32px;
    }
  }
}

// 响应式
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .header-actions {
    width: 100%;
    flex-wrap: wrap;
  }

  .filter-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .search-wrapper {
    max-width: none;
  }

  .filter-group {
    flex-wrap: wrap;
  }
}
</style>
