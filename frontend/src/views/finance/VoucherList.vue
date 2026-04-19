<template>
  <div class="page-view">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">凭证管理</h1>
        <p class="page-subtitle">共 {{ total }} 条凭证</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="showCreate = true">
          <el-icon><Plus /></el-icon>
          新增凭证
        </el-button>
      </div>
    </header>

    <!-- 筛选栏 -->
    <div class="filter-bar glass-card">
      <div class="filter-group">
        <el-select
          v-model="filterPeriod"
          placeholder="选择会计期间"
          clearable
          class="filter-select"
          @change="loadVouchers"
        >
          <template #prefix>
            <el-icon><Calendar /></el-icon>
          </template>
          <el-option v-for="p in periods" :key="p.id" :label="p.name" :value="p.id" />
        </el-select>
        <div class="search-wrapper">
          <el-icon class="search-icon"><Search /></el-icon>
          <input
            v-model="keyword"
            type="text"
            placeholder="搜索凭证号/摘要..."
            class="search-input"
            @keyup.enter="loadVouchers"
          />
        </div>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="table-container glass-card">
      <el-table
        :data="vouchers"
        stripe
        v-loading="loading"
        class="modern-table"
        :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
      >
        <el-table-column prop="voucher_no" label="凭证号" min-width="130">
          <template #default="{ row }">
            <span class="voucher-no">{{ row.voucher_no }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="period" label="会计期间" min-width="120">
          <template #default="{ row }">
            <el-tag size="small" class="period-tag">{{ row.period }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="日期" min-width="160" />
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <span class="status-badge" :class="`status--${row.status}`">
              {{ statusLabel(row.status) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="total_debit" label="借方合计" align="right" min-width="130">
          <template #default="{ row }">
            <span class="amount debit">{{ formatAmount(row.total_debit) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="total_credit" label="贷方合计" align="right" min-width="130">
          <template #default="{ row }">
            <span class="amount credit">{{ formatAmount(row.total_credit) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="creator_name" label="制单人" min-width="100">
          <template #default="{ row }">
            <div class="creator-cell">
              <el-avatar :size="24" class="creator-avatar">
                {{ row.creator_name?.[0] || '?' }}
              </el-avatar>
              <span>{{ row.creator_name || '—' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-btns">
              <el-button size="small" text @click="viewDetail(row)">
                <el-icon><View /></el-icon>
                详情
              </el-button>
              <el-button
                v-if="row.status === 'draft'"
                size="small"
                text
                type="warning"
                @click="handleSubmit(row)"
              >
                <el-icon><Promotion /></el-icon>
                提交
              </el-button>
              <el-button
                v-if="row.status === 'submitted'"
                size="small"
                text
                type="success"
                @click="handleAudit(row)"
              >
                <el-icon><CircleCheck /></el-icon>
                审核
              </el-button>
              <el-button
                v-if="row.status === 'audited'"
                size="small"
                text
                type="danger"
                @click="handleReverse(row)"
              >
                <el-icon><RefreshRight /></el-icon>
                红冲
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          :page-size="20"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="loadVouchers"
        />
      </div>
    </div>

    <!-- 新增凭证 -->
    <VoucherCreate
      v-model="showCreate"
      :periods="periods"
      @success="loadVouchers"
    />

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      title="凭证详情"
      width="700px"
      class="detail-dialog"
    >
      <div v-if="currentVoucher" class="voucher-detail">
        <div class="detail-header">
          <div class="detail-row">
            <div class="detail-item">
              <span class="detail-label">凭证号</span>
              <span class="detail-value">{{ currentVoucher.voucher_no }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">状态</span>
              <span class="status-badge" :class="`status--${currentVoucher.status}`">
                {{ statusLabel(currentVoucher.status) }}
              </span>
            </div>
          </div>
          <div class="detail-row">
            <div class="detail-item">
              <span class="detail-label">借方合计</span>
              <span class="amount debit">{{ formatAmount(currentVoucher.total_debit) }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">贷方合计</span>
              <span class="amount credit">{{ formatAmount(currentVoucher.total_credit) }}</span>
            </div>
          </div>
          <div class="detail-row">
            <div class="detail-item">
              <span class="detail-label">制单人</span>
              <span class="detail-value">{{ currentVoucher.creator_name }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">日期</span>
              <span class="detail-value">{{ currentVoucher.created_at }}</span>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Calendar, Plus, View, Promotion, CircleCheck, RefreshRight } from '@element-plus/icons-vue'
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
const submitting = ref(false)
const auditing = ref(false)
const reversing = ref(false)

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
.filter-bar {
  padding: 16px 20px;
  margin-bottom: 20px;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 16px;
}

.filter-select {
  width: 180px;
}

.search-wrapper {
  position: relative;
  flex: 1;
  max-width: 320px;
}

.search-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-tertiary);
  font-size: 18px;
}

.search-input {
  width: 100%;
  padding: 10px 14px 10px 42px;
  font-size: 14px;
  color: var(--text-primary);
  background: var(--bg-page);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  outline: none;
  transition: all 0.2s ease;

  &::placeholder { color: var(--text-tertiary); }
  &:focus {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(var(--primary), 0.1);
  }
}

.table-container {
  padding: 0;
  overflow: hidden;
}

:deep(.modern-table) {
  .el-table__header th {
    padding: 16px 12px;
    font-size: 13px;
  }

  .el-table__row {
    transition: background 0.2s ease;
    &:hover > td { background: rgba(var(--primary), 0.02) !important; }
  }

  .el-table__cell {
    padding: 16px 12px;
    border-bottom: 1px solid #F3F4F6;
  }
}

.voucher-no {
  font-weight: 600;
  color: var(--primary);
  font-family: 'SF Mono', Monaco, monospace;
}

.period-tag {
  background: rgba(var(--primary), 0.08);
  color: var(--primary);
  border: none;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 20px;

  &.status--draft { background: #F3F4F6; color: #6B7280; }
  &.status--submitted { background: #FEF3C7; color: #D97706; }
  &.status--audited { background: #D1FAE5; color: #059669; }
  &.status--reversed { background: #FEE2E2; color: #DC2626; }
}

.amount {
  font-family: 'SF Mono', Monaco, monospace;
  font-weight: 600;
  &.debit { color: var(--danger); }
  &.credit { color: var(--success); }
}

.creator-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary);
}

.creator-avatar {
  background: linear-gradient(135deg, var(--primary-light), var(--primary));
  color: #fff;
  font-size: 11px;
  font-weight: 600;
}

.action-btns {
  display: flex;
  gap: 8px;

  :deep(.el-button) {
    padding: 4px 8px;
    border-radius: var(--radius-sm);
    display: inline-flex;
    align-items: center;
    gap: 4px;

    .el-icon { font-size: 14px; }
  }
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: 16px 20px;
  border-top: 1px solid var(--border);
}

.voucher-detail {
  .detail-header {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .detail-row {
    display: flex;
    gap: 24px;
  }

  .detail-item {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .detail-label {
    font-size: 12px;
    color: var(--text-tertiary);
  }

  .detail-value {
    font-size: 15px;
    font-weight: 500;
    color: var(--text-primary);
  }
}

@media (max-width: 768px) {
  .filter-group { flex-wrap: wrap; }
  .search-wrapper { max-width: none; width: 100%; }
  .detail-row { flex-direction: column; }
}
</style>
