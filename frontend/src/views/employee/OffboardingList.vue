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
            <el-option label="待审核" value="pending" />
            <el-option label="已批准" value="approved" />
            <el-option label="已驳回" value="rejected" />
            <el-option label="已完成" value="completed" />
          </el-select>
        </el-form-item>
      </el-form>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="employee_name" label="发起人" min-width="90" />
        <el-table-column prop="type" label="事项" min-width="80">
          <template #default="{ row }">
            {{ offboardingTypeMap[row.type] || row.type }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" min-width="160" />
        <el-table-column prop="status" label="状态" min-width="90">
          <template #default="{ row }">
            <el-tag :type="offboardingStatusTagType[row.status]" size="small">
              {{ offboardingStatusMap[row.status] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <!-- pending: 同意 + 驳回 -->
            <template v-if="row.status === 'pending'">
              <el-popconfirm
                :title="`确认同意 ${row.employee_name} 的离职申请？`"
                confirm-button-text="确认"
                cancel-button-text="取消"
                @confirm="handleApprove(row.id)"
              >
                <template #reference>
                  <el-button size="small" type="primary" :loading="approving">同意</el-button>
                </template>
              </el-popconfirm>
              <el-button size="small" type="danger" @click="openRejectDialog(row)">驳回</el-button>
            </template>
            <!-- approved: 去减员 + 完成离职 -->
            <template v-if="row.status === 'approved'">
              <el-button size="small" type="warning" @click="goToSIRegister(row.employee_id, row.employee_name)">去减员</el-button>
              <el-popconfirm
                title="确认完成离职？"
                confirm-button-text="确认"
                cancel-button-text="取消"
                @confirm="handleComplete(row.id)"
              >
                <template #reference>
                  <el-button size="small" type="success" :loading="completing">完成离职</el-button>
                </template>
              </el-popconfirm>
            </template>
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

    <!-- 驳回弹窗 -->
    <el-dialog v-model="rejectDialogVisible" title="驳回离职申请" width="440px" destroy-on-close>
      <el-form>
        <el-form-item label="驳回原因（选填）">
          <el-input
            v-model="rejectReason"
            type="textarea"
            :rows="3"
            placeholder="请输入驳回原因"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="rejecting" @click="handleReject">确认驳回</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { employeeApi } from '@/api/employee'
import { ElMessage } from 'element-plus'
import { offboardingStatusMap, offboardingStatusTagType } from './statusMap'

const router = useRouter()

const offboardingTypeMap: Record<string, string> = {
  voluntary: '主动离职',
  involuntary: '公司发起',
}

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

async function handleApprove(id: number) {
  if (approving.value) return
  approving.value = true
  try {
    await employeeApi.approveOffboarding(id)
    ElMessage.success('已同意')
    load(page.value)
  } catch {
    ElMessage.error('操作失败')
  } finally {
    approving.value = false
  }
}

const completing = ref(false)

async function handleComplete(id: number) {
  if (completing.value) return
  completing.value = true
  try {
    await employeeApi.completeOffboarding(id)
    ElMessage.success('离职已办理完成')
    load(page.value)
  } catch {
    ElMessage.error('操作失败')
  } finally {
    completing.value = false
  }
}

// 驳回相关
const rejectDialogVisible = ref(false)
const rejectReason = ref('')
const rejecting = ref(false)
const rejectingRow = ref<any>(null)

function openRejectDialog(row: any) {
  rejectingRow.value = row
  rejectReason.value = ''
  rejectDialogVisible.value = true
}

async function handleReject() {
  if (rejecting.value || !rejectingRow.value) return
  rejecting.value = true
  try {
    await employeeApi.rejectOffboarding(rejectingRow.value.id, rejectReason.value || undefined)
    ElMessage.success('已驳回')
    rejectDialogVisible.value = false
    load(page.value)
  } catch {
    ElMessage.error('操作失败')
  } finally {
    rejecting.value = false
  }
}

// 去减员：跳转社保减员页面
function goToSIRegister(employeeId: number, employeeName: string) {
  router.push({
    path: '/tool/socialinsurance',
    query: {
      action: 'reduce',
      employee_id: String(employeeId),
      employee_name: employeeName,
    },
  })
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
</style>
