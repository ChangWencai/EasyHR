<template>
  <el-dialog
    v-model="visible"
    title="新增凭证"
    width="960px"
    :close-on-click-modal="false"
    @closed="resetForm"
    class="voucher-create-dialog"
  >
    <template #header>
      <div class="dialog-header">
        <div class="header-icon">
          <el-icon><Document /></el-icon>
        </div>
        <div class="header-text">
          <span class="header-title">新增凭证</span>
          <span class="header-subtitle">填写凭证分录信息</span>
        </div>
      </div>
    </template>

    <el-form :model="form" label-position="top" class="voucher-form">
      <div class="form-section">
        <el-form-item label="会计期间" required>
          <el-select v-model="form.period_id" placeholder="请选择会计期间" size="large" style="width: 100%">
            <template #prefix>
              <el-icon><Calendar /></el-icon>
            </template>
            <el-option v-for="p in periods" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
      </div>

      <div class="form-section">
        <div class="section-header">
          <h4 class="section-title">凭证分录</h4>
          <el-button size="small" @click="addEntry">
            <el-icon><Plus /></el-icon>
            新增一行
          </el-button>
        </div>

        <div class="entry-table">
          <div class="entry-header">
            <span class="col-account">科目</span>
            <span class="col-dc">借贷</span>
            <span class="col-amount">金额</span>
            <span class="col-summary">摘要</span>
            <span class="col-action">操作</span>
          </div>
          <TransitionGroup name="entry">
            <div v-for="(entry, idx) in form.entries" :key="idx" class="entry-row">
              <div class="col-account">
                <el-tree-select
                  v-model="entry.account_id"
                  :data="accountTreeData"
                  placeholder="选择科目"
                  clearable
                  check-strictly
                  size="large"
                  style="width: 100%"
                />
              </div>
              <div class="col-dc">
                <el-segmented v-model="entry.dc" :options="dcOptions" size="small" />
              </div>
              <div class="col-amount">
                <el-input-number
                  v-model="entry.amount"
                  :precision="2"
                  :min="0"
                  :controls="false"
                  placeholder="金额"
                  size="large"
                  style="width: 100%"
                />
              </div>
              <div class="col-summary">
                <el-input v-model="entry.summary" placeholder="摘要" size="large" />
              </div>
              <div class="col-action">
                <el-button
                  :disabled="form.entries.length <= 2"
                  type="danger"
                  text
                  @click="removeEntry(idx)"
                >
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>
          </TransitionGroup>
        </div>
      </div>

      <!-- 平衡提示 -->
      <div class="balance-card" :class="{ 'balance--error': !isBalanced }">
        <div class="balance-item">
          <span class="balance-label">借方合计</span>
          <span class="balance-value" :class="{ 'value--error': !isBalanced && totalDebit > 0 }">
            ¥{{ totalDebit.toLocaleString('zh-CN', { minimumFractionDigits: 2 }) }}
          </span>
        </div>
        <div class="balance-divider">
          <el-icon><Minus /></el-icon>
        </div>
        <div class="balance-item">
          <span class="balance-label">贷方合计</span>
          <span class="balance-value" :class="{ 'value--error': !isBalanced && totalCredit > 0 }">
            ¥{{ totalCredit.toLocaleString('zh-CN', { minimumFractionDigits: 2 }) }}
          </span>
        </div>
        <div class="balance-status" :class="isBalanced ? 'status--ok' : 'status--error'">
          <el-icon v-if="isBalanced"><CircleCheck /></el-icon>
          <el-icon v-else><Warning /></el-icon>
          <span>{{ isBalanced ? '借贷平衡' : '借贷不平衡' }}</span>
        </div>
      </div>
    </el-form>

    <template #footer>
      <el-button @click="visible = false" size="large">取消</el-button>
      <el-button
        type="primary"
        size="large"
        :disabled="!canSubmit"
        :loading="submitting"
        @click="handleSubmit"
      >
        <el-icon><Check /></el-icon>
        保存并提交
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Document, Calendar, Plus, Delete, Minus, CircleCheck, Warning, Check } from '@element-plus/icons-vue'
import { financeApi } from '@/api/finance'

interface Period { id: number; name: string }
interface TreeNode { id: number; label: string; code?: string; children?: TreeNode[] }

const props = defineProps<{ modelValue: boolean; periods: Period[] }>()
const emit = defineEmits<{ (e: 'update:modelValue', val: boolean): void; (e: 'success'): void }>()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

interface Entry { account_id: number | null; dc: 'debit' | 'credit'; amount: number | null; summary: string }

const form = ref<{ period_id: number | null; entries: Entry[] }>({ period_id: null, entries: [] })
const accountTreeData = ref<TreeNode[]>([])
const submitting = ref(false)

const dcOptions = [
  { label: '借', value: 'debit' },
  { label: '贷', value: 'credit' },
]

const totalDebit = computed(() =>
  form.value.entries.filter(e => e.dc === 'debit' && e.amount).reduce((acc, e) => acc + (e.amount || 0), 0)
)

const totalCredit = computed(() =>
  form.value.entries.filter(e => e.dc === 'credit' && e.amount).reduce((acc, e) => acc + (e.amount || 0), 0)
)

const isBalanced = computed(() => Math.abs(totalDebit.value - totalCredit.value) < 0.01)

const canSubmit = computed(() =>
  form.value.period_id !== null &&
  form.value.entries.length >= 2 &&
  form.value.entries.every(e => e.account_id !== null && e.amount !== null && e.amount > 0 && e.summary.trim()) &&
  isBalanced.value
)

function addEntry() { form.value.entries.push({ account_id: null, dc: 'debit', amount: null, summary: '' }) }
function removeEntry(idx: number) { form.value.entries.splice(idx, 1) }

async function loadAccounts() {
  try {
    const res = await financeApi.accountTree() as any
    const raw = res.data?.data || res.data || []
    accountTreeData.value = buildTree(raw)
  } catch { accountTreeData.value = [] }
}

function buildTree(accounts: any[]): TreeNode[] {
  return accounts.map(acc => ({
    id: acc.id, label: acc.code ? `${acc.code} ${acc.name}` : acc.name,
    code: acc.code, children: acc.children?.length ? buildTree(acc.children) : undefined,
  }))
}

async function handleSubmit() {
  if (submitting.value || !canSubmit.value) return
  submitting.value = true
  try {
    const payload = {
      period_id: form.value.period_id!,
      entries: form.value.entries.map(e => ({ account_id: e.account_id!, dc: e.dc, amount: String(e.amount), summary: e.summary })),
    }
    const createRes = await financeApi.createVoucher(payload) as any
    const voucherId = createRes.data?.id
    if (voucherId) await financeApi.submitVoucher(voucherId)
    ElMessage.success('凭证创建并提交成功')
    visible.value = false
    emit('success')
  } catch (e: unknown) {
    const msg = (e as any)?.response?.data?.error || '创建失败'
    ElMessage.error(msg)
  } finally { submitting.value = false }
}

function resetForm() { form.value = { period_id: null, entries: [] } }

watch(visible, (val) => {
  if (val) {
    loadAccounts()
    if (form.value.entries.length === 0) { addEntry(); addEntry() }
  }
})
</script>

<style scoped lang="scss">
$success: #10B981;
$warning: #F59E0B;
$error: #EF4444;
$text-primary: #1F2937;
$text-secondary: #6B7280;
$text-muted: #9CA3AF;
$border-color: #E5E7EB;
$radius-sm: 8px;
$radius-md: 12px;
$radius-lg: 16px;
$radius-xl: 24px;

:deep(.voucher-create-dialog) {
  .el-dialog__header { padding: 20px 24px; border-bottom: 1px solid $border-color; margin-right: 0; }
  .el-dialog__body { padding: 24px; }
  .el-dialog__footer { padding: 16px 24px; border-top: 1px solid $border-color; }
}

.dialog-header { display: flex; align-items: center; gap: 12px; }
.header-icon { width: 44px; height: 44px; background: linear-gradient(135deg, var(--primary-light), var(--primary)); border-radius: $radius-md; display: flex; align-items: center; justify-content: center; color: #fff; font-size: 20px; }
.header-text { display: flex; flex-direction: column; gap: 2px; }
.header-title { font-size: 18px; font-weight: 700; color: $text-primary; }
.header-subtitle { font-size: 13px; color: $text-muted; }

.voucher-form { }
.form-section { margin-bottom: 24px;
  :deep(.el-form-item) { margin-bottom: 0; }
  :deep(.el-form-item__label) { font-weight: 500; color: $text-secondary; padding-bottom: 8px; }
}

.section-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px;
  .section-title { font-size: 14px; font-weight: 600; color: $text-primary; margin: 0; }
}

.entry-table { background: #FAFAFA; border-radius: $radius-md; padding: 12px; }
.entry-header { display: flex; gap: 8px; padding: 8px 4px; font-size: 12px; font-weight: 600; color: $text-muted;
  .col-account { flex: 2; }
  .col-dc { width: 100px; }
  .col-amount { width: 140px; }
  .col-summary { flex: 1; }
  .col-action { width: 40px; text-align: center; }
}
.entry-row { display: flex; gap: 8px; align-items: center; padding: 8px 4px; background: #fff; border-radius: $radius-sm; margin-bottom: 8px;
  .col-account { flex: 2; }
  .col-dc { width: 100px; }
  .col-amount { width: 140px; }
  .col-summary { flex: 1; }
  .col-action { width: 40px; display: flex; justify-content: center; }
  &:last-child { margin-bottom: 0; }
}

.entry-enter-active, .entry-leave-active { transition: all 0.3s ease; }
.entry-enter-from { opacity: 0; transform: translateX(-20px); }
.entry-leave-to { opacity: 0; transform: translateX(20px); }

.balance-card { display: flex; align-items: center; gap: 24px; padding: 16px 20px; background: #F0FDF4; border: 1px solid #BBF7D0; border-radius: $radius-md; margin-top: 16px;
  &.balance--error { background: #FEF2F2; border-color: #FECACA; }
}
.balance-item { display: flex; flex-direction: column; gap: 2px; }
.balance-label { font-size: 12px; color: $text-muted; }
.balance-value { font-size: 20px; font-weight: 700; color: $success; font-family: 'SF Mono', Monaco, monospace;
  &.value--error { color: $error; }
}
.balance-divider { color: $text-muted; font-size: 20px; }
.balance-status { margin-left: auto; display: flex; align-items: center; gap: 6px; padding: 8px 16px; border-radius: 20px; font-size: 13px; font-weight: 600;
  &.status--ok { background: #D1FAE5; color: $success; }
  &.status--error { background: #FEE2E2; color: $error; }
}
</style>
