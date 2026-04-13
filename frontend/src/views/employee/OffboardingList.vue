<template>
  <div class="offboarding-list">
    <el-card>
      <template #header>
        <div class="header">
          <span>离职管理</span>
          <el-button @click="$router.push('/employee')">返回员工列表</el-button>
        </div>
      </template>

      <el-form inline @submit.prevent="load(1)">
        <el-form-item label="状态">
          <el-select v-model="filterStatus" placeholder="全部" clearable style="width: 120px" @change="load(1)">
            <el-option label="全部" value="" />
            <el-option label="待审核" value="pending_review" />
            <el-option label="已批准" value="approved" />
            <el-option label="已完成" value="completed" />
          </el-select>
        </el-form-item>
      </el-form>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="employee_name" label="员工姓名" min-width="90" />
        <el-table-column prop="resign_reason" label="离职原因" min-width="120" show-overflow-tooltip />
        <el-table-column prop="last_workday" label="最后工作日" min-width="120" />
        <el-table-column prop="status" label="状态" min-width="90">
          <template #default="{ row }">
            <el-tag :type="offboardingStatusTagType[row.status]" size="small">
              {{ offboardingStatusMap[row.status] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="离职清单" min-width="140">
          <template #default="{ row }">
            <span :class="{ done: row.checklist?.items_returned }">物品归还{{ row.checklist?.items_returned ? '✓' : '✗' }}</span>
            <span :class="{ done: row.checklist?.handover_done }">工作交接{{ row.checklist?.handover_done ? '✓' : '✗' }}</span>
            <span :class="{ done: row.checklist?.final_settlement }">结算{{ row.checklist?.final_settlement ? '✓' : '✗' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="申请时间" min-width="160" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'pending_review'"
              size="small"
              type="primary"
              @click="handleApprove(row.id)"
            >
              批准
            </el-button>
            <el-button
              v-if="row.status === 'approved'"
              size="small"
              type="success"
              @click="handleComplete(row.id)"
            >
              完成离职
            </el-button>
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
import { offboardingStatusMap, offboardingStatusTagType } from './statusMap'

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filterStatus = ref('')

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const res = await employeeApi.offboardings({
      page: p,
      page_size: pageSize.value,
      status: filterStatus.value || undefined,
    })
    list.value = res.list
    total.value = res.total
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

const approving = ref(false)
const completing = ref(false)

async function handleApprove(id: number) {
  if (approving.value) return
  approving.value = true
  try {
    await employeeApi.approveOffboarding(id)
    ElMessage.success('已批准')
    load()
  } catch {
    ElMessage.error('操作失败')
  } finally {
    approving.value = false
  }
}

async function handleComplete(id: number) {
  if (completing.value) return
  completing.value = true
  try {
    await employeeApi.completeOffboarding(id)
    ElMessage.success('离职已办理完成')
    load()
  } catch {
    ElMessage.error('操作失败')
  } finally {
    completing.value = false
  }
}

onMounted(() => load())
</script>

<style scoped lang="scss">
.offboarding-list {
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.mt-4 {
  margin-top: 16px;
}
.done {
  color: #67c23a;
  font-weight: 600;
}
</style>
