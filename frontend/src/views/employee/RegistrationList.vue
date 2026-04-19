<template>
  <div class="registration-list">
    <el-card>
      <template #header>
        <div class="header">
          <span class="title">信息登记管理</span>
          <el-button type="primary" @click="showCreateDialog = true">创建登记表</el-button>
        </div>
      </template>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="employee_name" label="员工姓名" min-width="100" />
        <el-table-column prop="department_name" label="部门" min-width="100" />
        <el-table-column prop="status" label="状态" min-width="80">
          <template #default="{ row }">
            <el-tag :type="statusTagType[row.status]" size="small">{{ statusMap[row.status] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="expires_at" label="过期时间" min-width="160">
          <template #default="{ row }">
            {{ formatDate(row.expires_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 'pending'">
              <el-button size="small" @click="openForwardDialog(row)">转发</el-button>
              <el-popconfirm
                title="确认删除此登记表？删除后链接失效，已提交的数据不受影响"
                @confirm="handleDelete(row.id)"
              >
                <template #reference>
                  <el-button size="small" type="danger">删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </template>
        </el-table-column>
      </el-table>

      <el-empty
        v-if="!loading && list.length === 0"
        description="暂无登记记录 —— 点击「创建登记表」邀请员工填写信息"
      />

      <el-pagination
        v-if="total > 0"
        class="mt-4"
        layout="total,prev,pager,next"
        :total="total"
        :page="page"
        :page-size="pageSize"
        @current-change="load"
      />
    </el-card>

    <RegistrationCreate
      v-model:visible="showCreateDialog"
      @created="handleCreated"
    />

    <RegistrationForwardDialog
      v-model:visible="showForwardDialog"
      :token="forwardToken"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { registrationApi, type Registration } from '@/api/employee'
import { ElMessage } from 'element-plus'
import RegistrationCreate from './RegistrationCreate.vue'
import RegistrationForwardDialog from './RegistrationForwardDialog.vue'

const list = ref<Registration[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const showCreateDialog = ref(false)
const showForwardDialog = ref(false)
const forwardToken = ref('')

const statusMap: Record<string, string> = {
  pending: '待填写',
  used: '已提交',
  expired: '已过期',
}

const statusTagType: Record<string, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
  pending: 'info',
  used: 'success',
  expired: 'danger',
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const res = await registrationApi.list({ page: p, page_size: pageSize.value })
    list.value = res.list
    total.value = res.total
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

function openForwardDialog(row: Registration) {
  forwardToken.value = row.token
  showForwardDialog.value = true
}

async function handleDelete(id: number) {
  try {
    await registrationApi.delete(id)
    ElMessage.success('删除成功')
    load(page.value)
  } catch {
    ElMessage.error('删除失败')
  }
}

function handleCreated() {
  showCreateDialog.value = false
  load(1)
}

onMounted(() => load())
</script>

<style scoped lang="scss">
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.title {
  font-size: 16px;
  font-weight: 700;
}

.mt-4 {
  margin-top: 16px;
}
</style>
