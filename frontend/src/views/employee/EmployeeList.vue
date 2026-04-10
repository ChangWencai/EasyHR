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
          <el-button type="primary" @click="load(1)">搜索</el-button>
        </el-form-item>
        <el-form-item>
          <el-button @click="handleExport">导出Excel</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="name" label="姓名" min-width="80" />
        <el-table-column prop="phone" label="手机号" min-width="120" />
        <el-table-column prop="position" label="岗位" min-width="100" />
        <el-table-column prop="entry_date" label="入职日期" min-width="110" />
        <el-table-column prop="status" label="状态" min-width="80">
          <template #default="{ row }">
            <el-tag :type="statusTagType[row.status]" size="small">{{ statusMap[row.status] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="$router.push(`/employee/${row.id}`)">查看</el-button>
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { employeeApi } from '@/api/employee'
import { ElMessage } from 'element-plus'
import { statusMap } from '@/views/employee/statusMap'

const search = ref('')
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const statusTagType: Record<string, string> = {
  active: 'success',
  probation: 'warning',
  resigned: 'info',
  archived: 'info',
}

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const res = await employeeApi.list({ page: p, page_size: pageSize.value, search: search.value })
    list.value = res.list
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
    a.download = `员工列表_${Date.now()}.xlsx`
    a.click()
    URL.revokeObjectURL(url)
  }).catch(() => {
    ElMessage.error('导出失败')
  })
}

onMounted(() => load())
</script>

<style scoped lang="scss">
.employee-list {
  padding-bottom: 70px;
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
