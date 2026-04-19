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

        <!-- 合同 -->
        <div class="section-title">合同</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="合同类型">{{ contractTypeLabel }}</el-descriptions-item>
          <el-descriptions-item label="起止日期">{{ contractDateRange }}</el-descriptions-item>
          <el-descriptions-item label="到期天数">
            <template v-if="detail.contract_expiry_days !== null && detail.contract_expiry_days !== undefined">
              <span v-if="detail.contract_expiry_days > 0">{{ detail.contract_expiry_days }}天</span>
              <span v-else-if="detail.contract_expiry_days === 0" style="color: #E6A23C">今天到期</span>
              <span v-else style="color: #F56C6C">已过期{{ Math.abs(detail.contract_expiry_days) }}天</span>
            </template>
            <span v-else style="color: #8C8C8C">无固定期限</span>
          </el-descriptions-item>
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
      </template>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { employeeApi } from '@/api/employee'
import { statusMap, statusTagType } from '@/views/employee/statusMap'

interface EmployeeDetail {
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
const detail = ref<EmployeeDetail | null>(null)

const contractTypeMap: Record<string, string> = {
  fixed_term: '固定期限',
  indefinite: '无固定期限',
  intern: '实习',
}

const contractTypeLabel = computed(() => {
  if (!detail.value?.contract_type) return '-'
  return contractTypeMap[detail.value.contract_type] || detail.value.contract_type
})

const contractDateRange = computed(() => {
  if (!detail.value?.contract_start_date) return '-'
  const start = detail.value.contract_start_date
  const end = detail.value.contract_end_date || '无固定期限'
  return `${start} ~ ${end}`
})

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  return dateStr
}

async function loadDetail() {
  if (!props.employeeId) return
  loading.value = true
  try {
    const res = await employeeApi.get(props.employeeId) as unknown as Record<string, unknown>
    detail.value = {
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
</style>
