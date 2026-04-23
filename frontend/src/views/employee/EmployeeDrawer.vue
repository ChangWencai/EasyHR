<template>
  <el-drawer
    :model-value="modelValue"
    @update:model-value="emit('update:modelValue', $event)"
    direction="rtl"
    size="480px"
    :title="`员工详情 - ${detail?.name || ''}`"
  >
    <div v-loading="loading" style="padding: 0 24px 24px">
      <template v-if="detail">
        <el-tabs v-model="activeTab" class="employee-drawer-tabs">
          <el-tab-pane label="基本信息" name="basic">
            <!-- 基本信息 -->
            <div class="section-title">基本信息</div>
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="姓名">{{ detail.name }}</el-descriptions-item>
              <el-descriptions-item label="性别">{{ detail.gender }}</el-descriptions-item>
              <el-descriptions-item label="手机号">{{ detail.phone }}</el-descriptions-item>
              <el-descriptions-item label="邮箱">{{ detail.email || '-' }}</el-descriptions-item>
              <el-descriptions-item label="部门">{{ detail.department_name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="岗位">{{ detail.position }}</el-descriptions-item>
              <el-descriptions-item label="入职日期">{{ formatDate(detail.hire_date) }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="statusTagType[detail.status]" size="small">{{ statusMap[detail.status] }}</el-tag>
              </el-descriptions-item>
            </el-descriptions>

            <!-- 身份证 -->
            <div class="section-title">身份证</div>
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="证件号">{{ detail.id_card }}</el-descriptions-item>
            </el-descriptions>

            <!-- 银行卡 -->
            <div class="section-title">银行卡</div>
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="卡号">{{ detail.bank_account || '-' }}</el-descriptions-item>
              <el-descriptions-item label="开户行">{{ detail.bank_name || '-' }}</el-descriptions-item>
            </el-descriptions>

            <!-- 其他信息 -->
            <div class="section-title">其他信息</div>
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="地址">{{ detail.address || '-' }}</el-descriptions-item>
              <el-descriptions-item label="紧急联系人">{{ detail.emergency_contact || '-' }}</el-descriptions-item>
              <el-descriptions-item label="紧急联系电话">{{ detail.emergency_phone || '-' }}</el-descriptions-item>
              <el-descriptions-item label="备注">{{ detail.remark || '-' }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>
          <el-tab-pane label="合同" name="contract">
            <ContractList
              :employee-id="detail.id"
              :employee-name="detail.name"
              :employee-salary="0"
              @open-wizard="activeTab = 'basic'"
            />
          </el-tab-pane>
        </el-tabs>

        <!-- 操作按钮 -->
        <div v-if="detail?.status === 'pending'" class="action-section">
          <el-button type="info" size="large" :loading="abandoning" @click="handleAbandonOnboarding">
            <el-icon><Close /></el-icon>
            放弃入职
          </el-button>
          <el-button type="success" size="large" :loading="confirming" @click="handleConfirmOnboarding">
            <el-icon><Check /></el-icon>
            确认入职
          </el-button>
        </div>
      </template>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { employeeApi } from '@/api/employee'
import { statusMap, statusTagType } from '@/views/employee/statusMap'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check, Close } from '@element-plus/icons-vue'
import ContractList from './components/ContractList.vue'

interface EmployeeDetail {
  id: number
  name: string
  gender: string
  phone: string
  email: string
  department_name: string
  position: string
  hire_date: string
  status: string
  id_card: string
  contract_type: string
  contract_start_date: string
  contract_end_date: string
  contract_expiry_days: number | null
  bank_account: string
  bank_name: string
  address: string
  emergency_contact: string
  emergency_phone: string
  remark: string
}

const props = defineProps<{
  modelValue: boolean
  employeeId: number
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const loading = ref(false)
const confirming = ref(false)
const abandoning = ref(false)
const detail = ref<EmployeeDetail | null>(null)
const activeTab = ref('basic')

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  return dateStr
}

async function handleConfirmOnboarding() {
  if (!detail.value) return

  // 检查入职日期是否超过30天
  if (detail.value.hire_date) {
    const hireDateStr = detail.value.hire_date.split('T')[0]  // 提取 YYYY-MM-DD 部分
    const hireDate = new Date(hireDateStr)
    const now = new Date()
    const daysDiff = Math.floor((now.getTime() - hireDate.getTime()) / (1000 * 60 * 60 * 24))
    if (daysDiff > 30) {
      ElMessage.warning(`该员工入职日期已超过30天（${daysDiff}天），请确认是否仍需入职`)
    }
  }

  confirming.value = true
  try {
    await employeeApi.confirmOnboarding(detail.value.id)
    ElMessage.success('入职确认成功')
    detail.value.status = 'active'
    emit('update:modelValue', false)
  } catch {
    ElMessage.error('确认入职失败')
  } finally {
    confirming.value = false
  }
}

async function handleAbandonOnboarding() {
  if (!detail.value) return
  try {
    await ElMessageBox.confirm(
      `确定要放弃入职吗？员工"${detail.value.name}"将被删除。`,
      '放弃入职',
      { confirmButtonText: '确定放弃', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return
  }

  abandoning.value = true
  try {
    await employeeApi.delete(detail.value.id)
    ElMessage.success('已放弃入职')
    emit('update:modelValue', false)
  } catch {
    ElMessage.error('操作失败')
  } finally {
    abandoning.value = false
  }
}

async function loadDetail() {
  if (!props.employeeId) return
  loading.value = true
  try {
    const res = await employeeApi.get(props.employeeId) as unknown as Record<string, unknown>
    detail.value = {
      id: (res.id as number) || 0,
      name: (res.name as string) || '',
      gender: (res.gender as string) || '',
      phone: (res.phone as string) || '',
      email: (res.email as string) || '',
      department_name: (res.department_name as string) || '',
      position: (res.position as string) || '',
      hire_date: (res.entry_date as string) || '',
      status: (res.status as string) || '',
      id_card: (res.id_number as string) || '',
      contract_type: (res.contract_type as string) || '',
      contract_start_date: (res.contract_start_date as string) || '',
      contract_end_date: (res.contract_end_date as string) || '',
      contract_expiry_days: (res.contract_expiry_days as number) ?? null,
      bank_account: (res.bank_card as string) || '',
      bank_name: (res.bank_name as string) || '',
      address: (res.address as string) || '',
      emergency_contact: (res.emergency_contact as string) || '',
      emergency_phone: (res.emergency_phone as string) || '',
      remark: (res.remark as string) || '',
    }
  } catch {
    detail.value = null
  } finally {
    loading.value = false
  }
}

watch(() => props.modelValue, (val) => {
  if (val && props.employeeId) {
    loadDetail()
  }
})
</script>

<style scoped lang="scss">
.section-title {
  font-size: 16px;
  font-weight: 700;
  color: #303133;
  margin: 20px 0 12px;
}

.section-title:first-child {
  margin-top: 0;
}

.employee-drawer-tabs {
  :deep(.el-tabs__item) {
    font-size: 14px;
    font-weight: 500;
    &.is-active {
      color: var(--el-color-primary);
    }
  }
  :deep(.el-tabs__nav-wrap::after) {
    background-color: var(--el-border-color);
  }
  :deep(.el-tabs__item):not(.is-active) {
    color: var(--el-text-color-secondary);
  }
}

.action-section {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid var(--el-border-color);
  display: flex;
  gap: 12px;

  :deep(.el-button) {
    flex: 1;
    font-weight: 600;
  }
}
</style>
