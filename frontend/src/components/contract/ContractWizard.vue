<script setup lang="ts">
import { ref } from 'vue'
import ContractTypeSelect from './ContractTypeSelect.vue'
import ContractPeriodPicker from './ContractPeriodPicker.vue'
import PdfPreview from './PdfPreview.vue'
import { contractApi, type ContractType, type Contract } from '@/api/contract'
import { useMessage } from '@/composables/useMessage'

const $msg = useMessage()

const props = defineProps<{
  employeeId: number
  employeeName: string
  employeeSalary?: number
  probationSalary?: number
  editingContract?: Contract | null
}>()
const emit = defineEmits<{
  success: []
  close: []
}>()

const currentStep = ref(0)
const loading = ref(false)
const pdfUrl = ref('')

// Step 1 data
const contractType = ref<ContractType | null>(null)
const period = ref({ start: '', end: '' as string | null, probationMonths: 0 })

// Contract ID after creation
const contractId = ref<number | null>(null)

// Editing mode: pre-populate from existing contract
const isEditing = !!props.editingContract
if (isEditing && props.editingContract) {
  contractType.value = props.editingContract.contract_type as ContractType
  period.value = {
    start: props.editingContract.start_date || '',
    end: props.editingContract.end_date || null,
    probationMonths: props.editingContract.probation_months || 0,
  }
  contractId.value = props.editingContract.id
}

// Step 2: PDF preview URL
const pdfLoading = ref(false)

// Step 3: Confirmation
const sending = ref(false)

const steps = [
  { title: '选择合同类型' },
  { title: '预览合同' },
  { title: '发送签署链接' },
]

// Step 1 validation
function canProceedStep1() {
  return contractType.value && period.value.start
}

// Step 1 -> Step 2: Create contract + generate PDF
async function proceedToStep2() {
  if (!canProceedStep1()) return
  loading.value = true
  try {
    if (isEditing && contractId.value) {
      // 更新已有合同后直接关闭
      await contractApi.update(contractId.value, {
        contract_type: contractType.value!,
        start_date: period.value.start,
        end_date: period.value.end,
        probation_months: period.value.probationMonths,
        salary: props.employeeSalary,
        probation_salary: props.probationSalary,
      })
      $msg.success('合同已更新')
      emit('success')
      return
    } else {
      // 创建新合同（draft）
      const contract = await contractApi.create({
        employee_id: props.employeeId,
        contract_type: contractType.value!,
        start_date: period.value.start,
        end_date: period.value.end,
        probation_months: period.value.probationMonths,
      }, props.employeeSalary, props.probationSalary)
      contractId.value = contract.id
    }

    // Generate PDF
    pdfLoading.value = true
    const blob = await contractApi.generatePdfBlob(contractId.value!)
    pdfUrl.value = URL.createObjectURL(blob)
    currentStep.value = 1
  } catch {
    $msg.error(isEditing ? '更新合同失败，请重试' : '创建合同失败，请重试')
  } finally {
    loading.value = false
    pdfLoading.value = false
  }
}

// Step 3: Send sign link
async function handleComplete() {
  if (!contractId.value) return
  sending.value = true
  try {
    await contractApi.sendSignLink(contractId.value)
    $msg.success('签署链接发送成功')
    emit('success')
  } catch {
    $msg.error('发送签署链接失败，请重试')
  } finally {
    sending.value = false
  }
}

function handleClose() {
  emit('close')
}
</script>

<template>
  <div class="contract-wizard">
    <!-- Steps indicator -->
    <el-steps :active="currentStep" finish-status="success" align-center class="wizard-steps">
      <el-step v-for="(step, i) in steps" :key="i" :title="step.title" />
    </el-steps>

    <!-- Step content -->
    <div class="step-content">
      <!-- Step 0: Select type + period -->
      <div v-if="currentStep === 0" class="wizard-step-1">
        <div class="step-hint">
          当前为「{{ props.employeeName }}」发起合同
        </div>

        <div class="form-section">
          <div class="form-label">合同类型</div>
          <ContractTypeSelect v-model="contractType" />
        </div>

        <div class="form-section">
          <div class="form-label">合同期限</div>
          <ContractPeriodPicker v-model="period" />
        </div>
      </div>

      <!-- Step 1: PDF Preview -->
      <div v-else-if="currentStep === 1" class="wizard-step-2">
        <div class="pdf-hint">
          请确认合同内容无误后再发送签署链接
        </div>
        <PdfPreview :url="pdfUrl" :loading="pdfLoading" />
      </div>

      <!-- Step 2: Confirmation -->
      <div v-else-if="currentStep === 2" class="wizard-step-3">
        <div class="confirm-card">
          <div class="confirm-icon">
            <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
              <circle cx="16" cy="16" r="16" fill="#10B981" fill-opacity="0.1"/>
              <path d="M10 16l4 4 8-8" stroke="#10B981" stroke-width="2" stroke-linecap="round"/>
            </svg>
          </div>
          <div class="confirm-title">合同预览确认</div>
          <div class="confirm-detail">
            <div>员工：{{ props.employeeName }}</div>
            <div>类型：{{ contractType === 'fixed_term' ? '劳动合同（固定期限）' : contractType === 'intern' ? '实习协议' : '兼职合同' }}</div>
            <div>期限：{{ period.start }} ~ {{ period.end || '无固定' }}</div>
          </div>
        </div>
        <div class="confirm-note">
          <p>签署链接将发送至员工手机号</p>
          <p>签署链接有效期：7天</p>
        </div>
      </div>
    </div>

    <!-- Action buttons -->
    <div class="wizard-footer">
      <el-button @click="handleClose">取消</el-button>
      <el-button
        v-if="currentStep === 0"
        type="primary"
        :disabled="!canProceedStep1()"
        :loading="loading"
        @click="proceedToStep2"
      >
        {{ isEditing ? '保存' : '下一步' }}
      </el-button>
      <el-button
        v-if="currentStep === 1 && !isEditing"
        type="primary"
        @click="currentStep = 2"
      >
        下一步
      </el-button>
      <el-button
        v-if="currentStep === 2"
        type="primary"
        :loading="sending"
        @click="handleComplete"
      >
        发送签署链接
      </el-button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.contract-wizard {
  padding: 8px 0;
}

.wizard-steps {
  margin-bottom: 20px;

  :deep(.el-step__title) {
    font-size: 14px;
    font-weight: 500;
  }
  :deep(.el-step__title.is-finish) {
    color: var(--el-color-primary);
  }
}

.step-content {
  margin: 20px 0;
}

.wizard-step-1 {
  .step-hint {
    margin-bottom: 16px;
    color: var(--text-secondary);
    font-size: 13px;
  }
  .form-label {
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary);
    margin-bottom: 12px;
  }
}

.wizard-step-2 {
  .pdf-hint {
    margin-bottom: 12px;
    color: var(--text-secondary);
    font-size: 13px;
  }
}

.wizard-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--border);

  .el-button {
    height: 52px;
    padding: 0 32px;
    border-radius: 12px;
    font-size: 16px;
    font-weight: 600;
  }
}

.confirm-card {
  background: rgba(16, 185, 129, 0.06);
  border: 1px solid rgba(16, 185, 129, 0.2);
  border-radius: 16px;
  padding: 24px;
  text-align: center;
  margin-bottom: 16px;

  .confirm-icon {
    margin-bottom: 12px;
  }
  .confirm-title {
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 12px;
  }
  .confirm-detail {
    font-size: 13px;
    color: var(--text-secondary);
    text-align: left;
    div { margin-bottom: 4px; }
  }
}

.confirm-note {
  font-size: 13px;
  color: var(--text-secondary);
  p { margin: 4px 0; }
}
</style>
