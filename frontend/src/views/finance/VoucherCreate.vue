<template>
  <el-dialog
    v-model="visible"
    title="新增凭证"
    width="900px"
    :close-on-click-modal="false"
    @closed="resetForm"
  >
    <el-form :model="form" label-width="90px" class="voucher-form">
      <el-form-item label="会计期间" required>
        <el-select v-model="form.period_id" placeholder="请选择会计期间" style="width: 220px">
          <el-option v-for="p in periods" :key="p.id" :label="p.name" :value="p.id" />
        </el-select>
      </el-form-item>

      <el-form-item label="分录">
        <div class="entry-table">
          <div class="entry-header">
            <span class="col-account">科目</span>
            <span class="col-dc">借贷方向</span>
            <span class="col-amount">金额</span>
            <span class="col-summary">摘要</span>
            <span class="col-action">操作</span>
          </div>
          <div v-for="(entry, idx) in form.entries" :key="idx" class="entry-row">
            <div class="col-account">
              <el-tree-select
                v-model="entry.account_id"
                :data="accountTreeData"
                :props="{ label: 'label', value: 'id', children: 'children' }"
                placeholder="选择科目"
                clearable
                check-strictly
                style="width: 100%"
              />
            </div>
            <div class="col-dc">
              <el-select v-model="entry.dc" placeholder="方向" style="width: 80px">
                <el-option label="借" value="debit" />
                <el-option label="贷" value="credit" />
              </el-select>
            </div>
            <div class="col-amount">
              <el-input-number
                v-model="entry.amount"
                :precision="2"
                :min="0"
                :controls="false"
                placeholder="金额"
                style="width: 130px"
              />
            </div>
            <div class="col-summary">
              <el-input v-model="entry.summary" placeholder="摘要" />
            </div>
            <div class="col-action">
              <el-button link type="danger" size="small" @click="removeEntry(idx)">删除</el-button>
            </div>
          </div>
          <el-button size="small" plain @click="addEntry">新增一行</el-button>
        </div>
      </el-form-item>

      <el-form-item>
        <div class="balance-info">
          <span>借方合计: <strong :class="{ 'text-danger': !isBalanced }">{{ totalDebit }}</strong></span>
          <span>贷方合计: <strong :class="{ 'text-danger': !isBalanced }">{{ totalCredit }}</strong></span>
          <span v-if="!isBalanced" class="text-danger">借贷不平衡，请调整</span>
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :disabled="!canSubmit" :loading="submitting" @click="handleSubmit">
        保存并提交
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { financeApi } from '@/api/finance'

interface Period {
  id: number
  name: string
}

interface TreeNode {
  id: number
  label: string
  code?: string
  children?: TreeNode[]
}

const props = defineProps<{
  modelValue: boolean
  periods: Period[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', val: boolean): void
  (e: 'success'): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

interface Entry {
  account_id: number | null
  dc: 'debit' | 'credit'
  amount: number | null
  summary: string
}

const form = ref<{ period_id: number | null; entries: Entry[] }>({
  period_id: null,
  entries: [],
})

const accountTreeData = ref<TreeNode[]>([])
const submitting = ref(false)

function addEntry() {
  form.value.entries.push({ account_id: null, dc: 'debit', amount: null, summary: '' })
}

function removeEntry(idx: number) {
  form.value.entries.splice(idx, 1)
}

const totalDebit = computed(() => {
  const sum = form.value.entries
    .filter((e) => e.dc === 'debit' && e.amount)
    .reduce((acc, e) => acc + (e.amount || 0), 0)
  return sum.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
})

const totalCredit = computed(() => {
  const sum = form.value.entries
    .filter((e) => e.dc === 'credit' && e.amount)
    .reduce((acc, e) => acc + (e.amount || 0), 0)
  return sum.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
})

const isBalanced = computed(() => {
  const debit = form.value.entries
    .filter((e) => e.dc === 'debit' && e.amount)
    .reduce((acc, e) => acc + (e.amount || 0), 0)
  const credit = form.value.entries
    .filter((e) => e.dc === 'credit' && e.amount)
    .reduce((acc, e) => acc + (e.amount || 0), 0)
  return Math.abs(debit - credit) < 0.01
})

const canSubmit = computed(() => {
  return (
    form.value.period_id !== null &&
    form.value.entries.length >= 2 &&
    form.value.entries.every(
      (e) => e.account_id !== null && e.amount !== null && e.amount > 0 && e.summary.trim(),
    ) &&
    isBalanced.value
  )
})

async function loadAccounts() {
  try {
    const res = await financeApi.accountTree() as any
    const raw = res.data?.data || res.data || []
    accountTreeData.value = buildTree(raw)
  } catch {
    accountTreeData.value = []
  }
}

function buildTree(accounts: any[]): TreeNode[] {
  return accounts.map((acc) => ({
    id: acc.id,
    label: acc.code ? `${acc.code} ${acc.name}` : acc.name,
    code: acc.code,
    children: acc.children?.length ? buildTree(acc.children) : undefined,
  }))
}

async function handleSubmit() {
  if (!canSubmit.value) return
  submitting.value = true
  try {
    const payload = {
      period_id: form.value.period_id!,
      entries: form.value.entries.map((e) => ({
        account_id: e.account_id!,
        dc: e.dc,
        amount: String(e.amount),
        summary: e.summary,
      })),
    }
    const createRes = await financeApi.createVoucher(payload) as any
    const voucherId = createRes.data?.id
    if (voucherId) {
      await financeApi.submitVoucher(voucherId)
    }
    ElMessage.success('凭证创建并提交成功')
    visible.value = false
    emit('success')
  } catch (e: unknown) {
    const msg = (e as any)?.response?.data?.error || '创建失败'
    ElMessage.error(msg)
  } finally {
    submitting.value = false
  }
}

function resetForm() {
  form.value = { period_id: null, entries: [] }
}

watch(visible, (val) => {
  if (val) {
    loadAccounts()
    if (form.value.entries.length === 0) {
      addEntry()
      addEntry()
    }
  }
})
</script>

<style scoped lang="scss">
.voucher-form {
  .entry-table {
    width: 100%;
    .entry-header {
      display: flex;
      gap: 8px;
      padding: 6px 0;
      font-size: 13px;
      color: #666;
      font-weight: 600;
      .col-account { flex: 2; }
      .col-dc { width: 90px; }
      .col-amount { width: 140px; }
      .col-summary { flex: 1; }
      .col-action { width: 60px; text-align: center; }
    }
    .entry-row {
      display: flex;
      gap: 8px;
      align-items: center;
      margin-bottom: 8px;
      .col-account { flex: 2; }
      .col-dc { width: 90px; }
      .col-amount { width: 140px; }
      .col-summary { flex: 1; }
      .col-action { width: 60px; text-align: center; }
    }
  }
  .balance-info {
    display: flex;
    gap: 24px;
    font-size: 14px;
    .text-danger { color: #f56c6c; }
  }
}
</style>
