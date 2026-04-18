<template>
  <div class="employee-list">
    <el-card>
      <template #header>
        <div class="header">
          <span>员工管理</span>
          <div class="header-actions">
            <el-button @click="$router.push('/employee/invitations')">入职邀请</el-button>
            <el-button @click="$router.push('/employee/offboardings')">离职管理</el-button>
            <el-button type="primary" @click="$router.push('/employee/create')">新增员工</el-button>
          </div>
        </div>
      </template>

      <el-form inline @submit.prevent="load(1)">
        <el-form-item>
          <el-input v-model="search" placeholder="搜索姓名/手机号/岗位" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-select v-model="departmentId" placeholder="全部部门" clearable style="width: 160px" @change="load(1)">
            <el-option v-for="dept in departments" :key="dept.id" :label="dept.name" :value="dept.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="load(1)">搜索</el-button>
        </el-form-item>
        <el-form-item>
          <el-button @click="handleExport">导出Excel</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="name" label="姓名" width="100" />
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType[row.status]" size="small">{{ statusMap[row.status] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="department_name" label="部门" width="120" />
        <el-table-column label="岗位薪资" width="120">
          <template #default="{ row }">
            <span v-if="row.salary_amount > 0">&yen;{{ row.salary_amount.toFixed(2) }}/月</span>
            <span v-else style="color: #C0C4CC">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="years_of_service" label="在职年限" width="100" />
        <el-table-column label="合同到期" width="120">
          <template #default="{ row }">
            <template v-if="row.contract_expiry_days !== null && row.contract_expiry_days !== undefined">
              <span v-if="row.contract_expiry_days > 0">{{ row.contract_expiry_days }}天</span>
              <span v-else-if="row.contract_expiry_days === 0" style="color: #E6A23C">今天到期</span>
              <span v-else style="color: #F56C6C">已过期{{ Math.abs(row.contract_expiry_days) }}天</span>
            </template>
            <span v-else style="color: #8C8C8C">无合同</span>
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openDrawer(row.id)">更多</el-button>
            <el-button size="small" type="primary" @click="$router.push(`/employee/${row.id}/edit`)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="mt-4"
        layout="total,prev,pager,next"
        :total="total"
        :page="page"
        :page-size="pageSize"
        @current-change="load"
      />
    </el-card>

    <EmployeeDrawer
      v-model="drawerVisible"
      :employee-id="selectedEmployeeId"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { employeeApi, type EmployeeRosterItem } from '@/api/employee'
import { departmentApi, type Department } from '@/api/department'
import { ElMessage } from 'element-plus'
import { statusMap, statusTagType } from '@/views/employee/statusMap'
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

function openDrawer(id: number) {
  selectedEmployeeId.value = id
  drawerVisible.value = true
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
  employeeApi.exportExcel().then((blob) => {
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
.employee-list {
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.header-actions {
  display: flex;
  gap: 8px;
}
.mt-4 {
  margin-top: 16px;
}
</style>
