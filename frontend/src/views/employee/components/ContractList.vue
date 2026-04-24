<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { contractApi, type Contract } from '@/api/contract'
import ContractStatusBadge from '@/components/contract/ContractStatusBadge.vue'
import ExpiryCountdown from '@/components/contract/ExpiryCountdown.vue'
import ContractWizard from '@/components/contract/ContractWizard.vue'
import { useMessage } from '@/composables/useMessage'

const $msg = useMessage()

const props = defineProps<{
  employeeId?: number
  employeeName?: string
  employeeSalary?: number
  probationSalary?: number
}>()

const emit = defineEmits<{
  'open-wizard': []
}>()

const contracts = ref<Contract[]>([])
const loading = ref(false)
const showWizard = ref(false)
const editingContract = ref<Contract | null>(null)

// 加载合同列表
async function loadContracts() {
  if (!props.employeeId) return
  loading.value = true
  try {
    const res = await contractApi.list(props.employeeId)
    contracts.value = res.list
  } catch {
    $msg.error('加载合同列表失败')
  } finally {
    loading.value = false
  }
}

// 合同类型标签
const typeLabelMap: Record<string, string> = {
  fixed_term: '劳动合同（固定期限）',
  indefinite: '兼职合同',
  intern: '实习协议',
}

function getTypeLabel(type: string) {
  return typeLabelMap[type] || type
}

onMounted(() => {
  loadContracts()
})

// 终止合同
async function handleTerminate(contract: Contract) {
  try {
    await ElMessageBox.confirm(
      '确定要终止该合同吗？终止后将无法恢复',
      '终止合同',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await contractApi.terminate(contract.id, '老板主动终止', new Date().toISOString().split('T')[0])
    $msg.success('合同已终止')
    loadContracts()
  } catch {
    // 用户取消
  }
}

// 查看PDF
async function handleViewPdf(contract: Contract) {
  if (contract.signed_pdf_url) {
    window.open(contract.signed_pdf_url, '_blank')
  } else if (contract.pdf_url) {
    window.open(contract.pdf_url, '_blank')
  } else {
    // 未生成PDF时直接请求
    try {
      const blob = await contractApi.generatePdfBlob(contract.id)
      const url = URL.createObjectURL(blob)
      window.open(url, '_blank')
    } catch {
      $msg.error('生成合同PDF失败')
    }
  }
}

// 发起合同
function handleOpenWizard() {
  showWizard.value = true
}

// 签署成功后刷新列表
function handleWizardSuccess() {
  showWizard.value = false
  editingContract.value = null
  loadContracts()
}

// 编辑合同
function handleEdit(contract: Contract) {
  editingContract.value = contract
  showWizard.value = true
}

// 删除合同
async function handleDelete(contract: Contract) {
  try {
    await ElMessageBox.confirm('确定删除该合同吗？删除后无法恢复。', '删除合同', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    })
    await contractApi.delete(contract.id)
    $msg.success('合同已删除')
    loadContracts()
  } catch {
    // 用户取消
  }
}
</script>

<template>
  <div class="contract-list">
    <!-- Header -->
    <div class="contract-list-header">
      <span class="section-title">劳动合同</span>
      <el-button
        v-if="contracts.length > 0 && props.employeeId"
        type="primary"
        size="small"
        @click="handleOpenWizard"
      >
        发起合同
      </el-button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="contract-list-loading">
      <el-skeleton :rows="2" animated />
    </div>

    <!-- Empty State -->
    <div v-else-if="contracts.length === 0" class="contract-list-empty">
      <div class="empty-icon">
        <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
          <rect x="8" y="4" width="32" height="40" rx="2" stroke="#DCDFE6" stroke-width="2"/>
          <line x1="14" y1="14" x2="34" y2="14" stroke="#DCDFE6" stroke-width="2"/>
          <line x1="14" y1="22" x2="34" y2="22" stroke="#DCDFE6" stroke-width="2"/>
          <line x1="14" y1="30" x2="24" y2="30" stroke="#DCDFE6" stroke-width="2"/>
        </svg>
      </div>
      <p class="empty-title">暂无劳动合同</p>
      <p class="empty-desc">为员工创建劳动合同，保障双方权益</p>
      <el-button type="primary" @click="handleOpenWizard">发起合同</el-button>
    </div>

    <!-- Contract Items -->
    <div v-else class="contract-items">
      <div
        v-for="contract in contracts"
        :key="contract.id"
        class="contract-item"
      >
        <div class="contract-item-header">
          <span class="contract-type">{{ getTypeLabel(contract.contract_type) }}</span>
          <ContractStatusBadge :status="contract.status" />
        </div>
        <div class="contract-item-period">
          {{ contract.start_date }} ~ {{ contract.end_date || '无固定' }}
        </div>
        <div class="contract-item-footer">
          <ExpiryCountdown
            v-if="contract.expiry_days != null"
            :days="contract.expiry_days"
          />
          <div class="contract-item-actions">
            <el-button link size="small" @click="handleViewPdf(contract)">
              查看
            </el-button>
            <template v-if="contract.status === 'draft' || contract.status === 'pending_sign'">
              <el-button link size="small" type="primary" @click="handleEdit(contract)">
                编辑
              </el-button>
              <el-button link size="small" type="danger" @click="handleDelete(contract)">
                删除
              </el-button>
            </template>
            <el-button
              v-if="contract.status === 'active'"
              link
              size="small"
              type="danger"
              @click="handleTerminate(contract)"
            >
              终止
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- Contract Wizard Dialog -->
    <el-dialog
      v-model="showWizard"
      :title="editingContract ? '编辑合同' : '发起合同'"
      width="600px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <ContractWizard
        v-if="showWizard"
        :employee-id="props.employeeId!"
        :employee-name="props.employeeName || ''"
        :employee-salary="props.employeeSalary"
        :probation-salary="props.probationSalary"
        :editing-contract="editingContract"
        @success="handleWizardSuccess"
        @close="showWizard = false; editingContract = null"
      />
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.contract-list {
  padding: 0;
}

.contract-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;

  .section-title {
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
  }
}

.contract-list-loading {
  padding: 16px 0;
}

.contract-list-empty {
  text-align: center;
  padding: 32px 0;

  .empty-icon {
    margin-bottom: 12px;
  }
  .empty-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 4px;
  }
  .empty-desc {
    font-size: 12px;
    color: var(--text-secondary);
    margin: 0 0 16px;
  }
}

.contract-items {
  max-height: 400px;
  overflow-y: auto;
}

.contract-item {
  padding: 12px 0;
  border-bottom: 1px solid var(--border);

  &:last-child {
    border-bottom: none;
  }
}

.contract-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.contract-type {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.contract-item-period {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.contract-item-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.contract-item-actions {
  display: flex;
  gap: 8px;
}
</style>
