<template>
  <div class="voucher-list">
    <div class="toolbar">
      <el-select v-model="filterPeriod" placeholder="选择会计期间" clearable style="width: 160px" @change="loadVouchers">
        <el-option v-for="p in periods" :key="p.id" :label="p.name" :value="p.id" />
      </el-select>
      <el-input
        v-model="keyword"
        placeholder="搜索凭证号/摘要"
        clearable
        style="width: 200px"
        @clear="loadVouchers"
        @keyup.enter="loadVouchers"
      >
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-button type="primary" @click="showCreate = true">新增凭证</el-button>
    </div>

    <el-table :data="vouchers" stripe v-loading="loading" class="mt-2">
      <el-table-column prop="voucher_no" label="凭证号" width="130" />
      <el-table-column prop="period" label="会计期间" width="120" />
      <el-table-column prop="created_at" label="日期" width="160" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="total_debit" label="借方合计" align="right" width="130">
        <template #default="{ row }">{{ formatAmount(row.total_debit) }}</template>
      </el-table-column>
      <el-table-column prop="total_credit" label="贷方合计" align="right" width="130">
        <template #default="{ row }">{{ formatAmount(row.total_credit) }}</template>
      </el-table-column>
      <el-table-column prop="creator_name" label="制单人" width="100" />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="viewDetail(row)">详情</el-button>
          <el-button
            v-if="row.status === 'draft'"
            link
            type="warning"
            size="small"
            @click="handleSubmit(row)"
          >提交</el-button>
          <el-button
            v-if="row.status === 'submitted'"
            link
            type="success"
            size="small"
            @click="handleAudit(row)"
          >审核</el-button>
          <el-button
            v-if="row.status === 'audited'"
            link
            type="danger"
            size="small"
            @click="handleReverse(row)"
          >红冲</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="page"
      :page-size="20"
      :total="total"
      layout="prev, pager, next"
      class="mt-2"
      @current-change="loadVouchers"
    />

    <VoucherCreate
      v-model="showCreate"
      :periods="periods"
      @success="loadVouchers"
    />

    <!-- Detail Dialog -->
    <el-dialog v-model="detailVisible" title="凭证详情" width="700px">
      <el-descriptions :column="2" border v-if="currentVoucher">
        <el-descriptions-item label="凭证号">{{ currentVoucher.voucher_no }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusType(currentVoucher.status)" size="small">
            {{ statusLabel(currentVoucher.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="借方合计">{{ formatAmount(currentVoucher.total_debit) }}</el-descriptions-item>
        <el-descriptions-item label="贷方合计">{{ formatAmount(currentVoucher.total_credit) }}</el-descriptions-item>
        <el-descriptions-item label="制单人">{{ currentVoucher.creator_name }}</el-descriptions-item>
        <el-descriptions-item label="日期">{{ currentVoucher.created_at }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { financeApi } from '@/api/finance'
import VoucherCreate from './VoucherCreate.vue'

interface Voucher {
  id: number
  period: string
  period_id?: number
  voucher_no: string
  status: string
  total_debit: string
  total_credit: string
  creator_name: string
  created_at: string
}

interface Period {
  id: number
  name: string
}

const vouchers = ref<Voucher[]>([])
const periods = ref<Period[]>([])
const filterPeriod = ref<number | null>(null)
const keyword = ref('')
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const showCreate = ref(false)
const detailVisible = ref(false)
const currentVoucher = ref<Voucher | null>(null)

function statusType(status: string) {
  const map: Record<string, string> = {
    draft: 'info',
    submitted: 'warning',
    audited: 'success',
    reversed: 'danger',
  }
  return map[status] || 'info'
}

function statusLabel(status: string) {
  const map: Record<string, string> = {
    draft: '草稿',
    submitted: '已提交',
    audited: '已审核',
    reversed: '已红冲',
  }
  return map[status] || status
}

function formatAmount(val: string) {
  if (!val) return '0.00'
  return parseFloat(val).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

async function loadPeriods() {
  try {
    const res = await financeApi.periods() as any
    periods.value = res.data?.data || res.data || []
  } catch {
    periods.value = []
  }
}

async function loadVouchers() {
  loading.value = true
  try {
    const res = await financeApi.vouchers({
      page: page.value,
      period_id: filterPeriod.value || undefined,
      keyword: keyword.value || undefined,
    }) as any
    vouchers.value = res.data?.data || res.data?.list || []
    total.value = res.data?.total || 0
  } catch {
    vouchers.value = []
  } finally {
    loading.value = false
  }
}

function viewDetail(row: Voucher) {
  currentVoucher.value = row
  detailVisible.value = true
}

async function handleSubmit(row: Voucher) {
  if (submitting.value) return
  submitting.value = true
  try {
    await financeApi.submitVoucher(row.id)
    ElMessage.success('提交成功')
    loadVouchers()
  } catch (e: unknown) {
    const msg = (e as any)?.response?.data?.error || '提交失败'
    ElMessage.error(msg)
  } finally {
    submitting.value = false
  }
}

async function handleAudit(row: Voucher) {
  if (auditing.value) return
  auditing.value = true
  try {
    await financeApi.auditVoucher(row.id)
    ElMessage.success('审核成功')
    loadVouchers()
  } catch (e: unknown) {
    const msg = (e as any)?.response?.data?.error || '审核失败'
    ElMessage.error(msg)
  } finally {
    auditing.value = false
  }
}

async function handleReverse(row: Voucher) {
  if (reversing.value) return
  reversing.value = true
  try {
    await ElMessageBox.confirm('确定要红冲此凭证吗？', '红冲确认', { type: 'warning' })
    await financeApi.reverseVoucher(row.id)
    ElMessage.success('红冲成功')
    loadVouchers()
  } catch (e: unknown) {
    if ((e as any)?.message !== 'cancel') {
      const msg = (e as any)?.response?.data?.error || '红冲失败'
      ElMessage.error(msg)
    }
  } finally {
    reversing.value = false
  }
}

onMounted(() => {
  loadPeriods()
  loadVouchers()
})
</script>

<style scoped lang="scss">
.voucher-list {
  padding: 8px;
  .toolbar {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }
  .mt-2 {
    margin-top: 12px;
  }
}
</style>
