<template>
  <div class="page-view">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">财务概览</h1>
        <p class="page-subtitle">{{ currentPeriod }}</p>
      </div>
      <div class="header-actions">
        <el-select v-model="selectedPeriod" placeholder="选择会计期间" class="period-selector" @change="loadData">
          <el-option v-for="p in periods" :key="p.id" :label="p.name" :value="p.id" />
        </el-select>
      </div>
    </header>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--income">
          <el-icon><TrendCharts /></el-icon>
        </div>
        <div class="stat-content">
          <span class="stat-label">本月收入</span>
          <span class="stat-value income">{{ formatAmount(stats.income) }}</span>
        </div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--expense">
          <el-icon><Wallet /></el-icon>
        </div>
        <div class="stat-content">
          <span class="stat-label">本月支出</span>
          <span class="stat-value expense">{{ formatAmount(stats.expense) }}</span>
        </div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--balance">
          <el-icon><Money /></el-icon>
        </div>
        <div class="stat-content">
          <span class="stat-label">结余</span>
          <span class="stat-value" :class="stats.balance >= 0 ? 'income' : 'expense'">
            {{ formatAmount(stats.balance) }}
          </span>
        </div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon stat-icon--pending">
          <el-icon><Clock /></el-icon>
        </div>
        <div class="stat-content">
          <span class="stat-label">待审核凭证</span>
          <span class="stat-value">{{ stats.pendingVouchers }}</span>
        </div>
      </div>
    </div>

    <!-- 快捷操作 + 最近动态 -->
    <div class="content-grid">
      <!-- 快捷操作 -->
      <div class="quick-actions glass-card">
        <h3 class="section-title">快捷操作</h3>
        <div class="action-list">
          <router-link to="/finance/vouchers/create" class="action-item">
            <div class="action-icon">
              <el-icon><Document /></el-icon>
            </div>
            <span>新增凭证</span>
          </router-link>
          <router-link to="/finance/vouchers" class="action-item">
            <div class="action-icon">
              <el-icon><List /></el-icon>
            </div>
            <span>凭证管理</span>
          </router-link>
          <router-link to="/finance/invoices" class="action-item">
            <div class="action-icon">
              <el-icon><Tickets /></el-icon>
            </div>
            <span>发票管理</span>
          </router-link>
          <router-link to="/finance/expenses" class="action-item">
            <div class="action-icon">
              <el-icon><Coin /></el-icon>
            </div>
            <span>报销审批</span>
          </router-link>
          <router-link to="/finance/accounts" class="action-item">
            <div class="action-icon">
              <el-icon><Grid /></el-icon>
            </div>
            <span>科目管理</span>
          </router-link>
          <router-link to="/finance/reports" class="action-item">
            <div class="action-icon">
              <el-icon><DataAnalysis /></el-icon>
            </div>
            <span>账簿报表</span>
          </router-link>
        </div>
      </div>

      <!-- 最近凭证 -->
      <div class="recent-vouchers glass-card">
        <h3 class="section-title">最近凭证</h3>
        <div v-loading="loadingVouchers" class="voucher-list">
          <div v-for="v in recentVouchers" :key="v.id" class="voucher-item">
            <div class="voucher-info">
              <span class="voucher-no">{{ v.voucher_no }}</span>
              <span class="voucher-date">{{ v.created_at }}</span>
            </div>
            <div class="voucher-amount">
              <span class="amount">{{ formatAmount(v.total_debit) }}</span>
              <span class="status-badge" :class="`status--${v.status}`">{{ statusLabel(v.status) }}</span>
            </div>
          </div>
          <div v-if="!loadingVouchers && recentVouchers.length === 0" class="empty-state">
            暂无凭证记录
          </div>
        </div>
        <router-link to="/finance/vouchers" class="view-more">查看全部 →</router-link>
      </div>

      <!-- 本月发票统计 -->
      <div class="invoice-stats glass-card">
        <h3 class="section-title">本月发票</h3>
        <div class="invoice-summary">
          <div class="invoice-row">
            <span class="invoice-label">已开票</span>
            <span class="invoice-value">{{ invoiceStats.issued }} 笔</span>
            <span class="invoice-amount">{{ formatAmount(invoiceStats.issuedAmount) }}</span>
          </div>
          <div class="invoice-row">
            <span class="invoice-label">待收票</span>
            <span class="invoice-value">{{ invoiceStats.pending }} 笔</span>
            <span class="invoice-amount">{{ formatAmount(invoiceStats.pendingAmount) }}</span>
          </div>
        </div>
        <router-link to="/finance/invoices" class="view-more">查看全部 →</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  TrendCharts, Wallet, Money, Clock, Document, List, Tickets, Coin, Grid, DataAnalysis
} from '@element-plus/icons-vue'
import { financeApi } from '@/api/finance'
import type { Voucher } from '@/api/finance'

const selectedPeriod = ref<number | null>(null)
const periods = ref<{ id: number; name: string }[]>([])
const loadingVouchers = ref(false)
const recentVouchers = ref<Voucher[]>([])

const stats = ref({
  income: 0,
  expense: 0,
  balance: 0,
  pendingVouchers: 0,
})

const invoiceStats = ref({
  issued: 0,
  issuedAmount: 0,
  pending: 0,
  pendingAmount: 0,
})

const currentPeriod = computed(() => {
  const p = periods.value.find(p => p.id === selectedPeriod.value)
  return p ? p.name : '全部期间'
})

function formatAmount(value: string | number): string {
  const num = typeof value === 'string' ? parseFloat(value) : value
  return num.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function statusLabel(status: string): string {
  const map: Record<string, string> = {
    draft: '草稿',
    pending: '待审核',
    approved: '已审核',
    rejected: '已驳回',
  }
  return map[status] || status
}

async function loadPeriods() {
  try {
    const res = await financeApi.periods()
    periods.value = res.data || []
    if (periods.value.length > 0) {
      selectedPeriod.value = periods.value[0].id
    }
  } catch (e) {
    console.error('加载会计期间失败', e)
  }
}

async function loadRecentVouchers() {
  loadingVouchers.value = true
  try {
    const res = await financeApi.vouchers({ page: 1 })
    recentVouchers.value = (res.data?.list || []).slice(0, 5)
    stats.value.pendingVouchers = res.data?.pending_count || 0
  } catch (e) {
    console.error('加载凭证失败', e)
  } finally {
    loadingVouchers.value = false
  }
}

async function loadInvoiceStats() {
  try {
    const res = await financeApi.invoices({ page: 1 })
    const list = res.data?.list || []
    invoiceStats.value = {
      issued: list.filter((i: { status: string }) => i.status === 'issued').length,
      issuedAmount: list
        .filter((i: { status: string }) => i.status === 'issued')
        .reduce((sum: number, i: { amount: string }) => sum + parseFloat(i.amount || '0'), 0),
      pending: list.filter((i: { status: string }) => i.status === 'pending').length,
      pendingAmount: list
        .filter((i: { status: string }) => i.status === 'pending')
        .reduce((sum: number, i: { amount: string }) => sum + parseFloat(i.amount || '0'), 0),
    }
  } catch (e) {
    console.error('加载发票统计失败', e)
  }
}

async function loadData() {
  await Promise.all([loadRecentVouchers(), loadInvoiceStats()])
}

onMounted(async () => {
  await loadPeriods()
  await loadData()
})
</script>

<style scoped lang="scss">
$bg-page: #FAFBFC;
$bg-surface: #FFFFFF;
$text-primary: #1F2937;
$text-secondary: #6B7280;
$text-muted: #9CA3AF;
$border-color: #E5E7EB;
$radius-md: 12px;
$radius-lg: 16px;

.page-view {
  padding: 24px 32px;
  background: $bg-page;
  min-height: calc(100vh - 56px);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-left {
  .page-title {
    font-size: 24px;
    font-weight: 600;
    color: $text-primary;
    margin: 0;
  }
  .page-subtitle {
    font-size: 14px;
    color: $text-muted;
    margin: 4px 0 0;
  }
}

.period-selector {
  width: 180px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  border-radius: $radius-lg;
  background: $bg-surface;
  transition: box-shadow 0.2s;
  cursor: default;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  }
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: $radius-md;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;

  &--income {
    background: linear-gradient(135deg, #10B981 0%, #059669 100%);
    color: #fff;
  }

  &--expense {
    background: linear-gradient(135deg, #EF4444 0%, #DC2626 100%);
    color: #fff;
  }

  &--balance {
    background: linear-gradient(135deg, #3B82F6 0%, #2563EB 100%);
    color: #fff;
  }

  &--pending {
    background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%);
    color: #fff;
  }
}

.stat-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-label {
  font-size: 13px;
  color: $text-muted;
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  color: $text-primary;

  &.income {
    color: #10B981;
  }

  &.expense {
    color: #EF4444;
  }
}

.content-grid {
  display: grid;
  grid-template-columns: 1fr 2fr 1fr;
  gap: 20px;
}

.glass-card {
  background: $bg-surface;
  border-radius: $radius-lg;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: $text-primary;
  margin: 0 0 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid $border-color;
}

.action-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.action-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  border-radius: $radius-md;
  background: #F9FAFB;
  text-decoration: none;
  color: $text-secondary;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s;
  cursor: pointer;

  &:hover {
    background: #EEF2FF;
    color: #4F46E5;
    transform: translateY(-1px);
  }
}

.action-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: $bg-surface;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: $text-muted;
  transition: color 0.2s;
}

.action-item:hover .action-icon {
  color: #4F46E5;
}

.voucher-list {
  min-height: 120px;
}

.voucher-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #F3F4F6;

  &:last-of-type {
    border-bottom: none;
  }
}

.voucher-info {
  display: flex;
  flex-direction: column;
  gap: 2px;

  .voucher-no {
    font-size: 14px;
    font-weight: 500;
    color: $text-primary;
  }

  .voucher-date {
    font-size: 12px;
    color: $text-muted;
  }
}

.voucher-amount {
  display: flex;
  align-items: center;
  gap: 8px;

  .amount {
    font-size: 14px;
    font-weight: 600;
    color: $text-primary;
  }
}

.status-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;

  &.status--draft {
    background: #F3F4F6;
    color: $text-muted;
  }

  &.status--pending {
    background: #FEF3C7;
    color: #D97706;
  }

  &.status--approved {
    background: #D1FAE5;
    color: #059669;
  }

  &.status--rejected {
    background: #FEE2E2;
    color: #DC2626;
  }
}

.view-more {
  display: block;
  text-align: center;
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid $border-color;
  color: #4F46E5;
  font-size: 13px;
  text-decoration: none;
  font-weight: 500;
  cursor: pointer;

  &:hover {
    color: #3730A3;
  }
}

.empty-state {
  text-align: center;
  color: $text-muted;
  font-size: 14px;
  padding: 24px 0;
}

.invoice-summary {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.invoice-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background: #F9FAFB;
  border-radius: $radius-md;
}

.invoice-label {
  font-size: 13px;
  color: $text-secondary;
  flex: 1;
}

.invoice-value {
  font-size: 14px;
  font-weight: 600;
  color: $text-primary;
}

.invoice-amount {
  font-size: 13px;
  color: #10B981;
  font-weight: 500;
}
</style>