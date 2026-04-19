<template>
  <div class="invoice-list">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">发票管理</h1>
        <p class="page-subtitle">共 {{ total }} 张发票</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="showCreate = true">
          <el-icon><Plus /></el-icon>
          登记发票
        </el-button>
      </div>
    </header>

    <!-- 筛选栏 -->
    <div class="filter-bar glass-card">
      <el-radio-group v-model="filterType" @change="loadInvoices" class="filter-tabs">
        <el-radio-button label="">全部</el-radio-button>
        <el-radio-button label="input">
          <el-icon><Download /></el-icon>
          进项发票
        </el-radio-button>
        <el-radio-button label="output">
          <el-icon><Upload /></el-icon>
          销项发票
        </el-radio-button>
      </el-radio-group>
    </div>

    <!-- 数据表格 -->
    <div class="table-container glass-card">
      <el-table
        :data="invoices"
        stripe
        v-loading="loading"
        class="modern-table"
        :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
      >
        <el-table-column prop="invoice_no" label="发票号码" min-width="180">
          <template #default="{ row }">
            <span class="invoice-no">{{ row.invoice_no }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <span class="type-badge" :class="`type--${row.type}`">
              <el-icon>
                <Download v-if="row.type === 'input'" />
                <Upload v-else />
              </el-icon>
              {{ row.type === 'input' ? '进项' : '销项' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="amount" label="金额" align="right" min-width="140">
          <template #default="{ row }">
            <span class="amount">¥{{ formatAmount(row.amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="tax_rate" label="税率" width="90">
          <template #default="{ row }">
            <span class="tax-rate">{{ row.tax_rate ? row.tax_rate + '%' : '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="tax_amount" label="税额" align="right" min-width="120">
          <template #default="{ row }">
            <span class="tax-amount">¥{{ formatAmount(row.tax_amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="voucher_no" label="关联凭证" min-width="140">
          <template #default="{ row }">
            <span v-if="row.voucher_no" class="voucher-link">
              <el-icon><Link /></el-icon>
              {{ row.voucher_no }}
            </span>
            <span v-else class="no-voucher">未关联</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="登记日期" min-width="160" />
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          :page-size="20"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="loadInvoices"
        />
      </div>
    </div>

    <!-- 登记发票弹窗 -->
    <el-dialog
      v-model="showCreate"
      title="登记发票"
      width="500px"
      class="create-dialog"
    >
      <el-form :model="form" label-position="top" class="create-form">
        <el-form-item label="发票号码" prop="invoice_no" required>
          <el-input
            v-model="form.invoice_no"
            placeholder="请输入发票号码"
            size="large"
          >
            <template #prefix>
              <el-icon><Tickets /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="发票类型" prop="type" required>
          <el-radio-group v-model="form.type" class="type-selector">
            <el-radio value="input">
              <div class="radio-card">
                <el-icon><Download /></el-icon>
                <span>进项发票</span>
              </div>
            </el-radio>
            <el-radio value="output">
              <div class="radio-card">
                <el-icon><Upload /></el-icon>
                <span>销项发票</span>
              </div>
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="金额（元）" prop="amount" required>
          <el-input-number
            v-model="form.amount"
            :precision="2"
            :min="0"
            :controls="false"
            placeholder="请输入金额"
            size="large"
            style="width: 100%"
          >
            <template #prefix>
              <span class="currency-prefix">¥</span>
            </template>
          </el-input-number>
        </el-form-item>
        <div class="form-row">
          <el-form-item label="税率（%）" class="form-item-half">
            <el-input-number
              v-model="form.tax_rate"
              :precision="2"
              :min="0"
              :max="100"
              :controls="false"
              size="large"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="税额（元）" class="form-item-half">
            <el-input-number
              v-model="form.tax_amount"
              :precision="2"
              :min="0"
              :controls="false"
              size="large"
              style="width: 100%"
            >
              <template #prefix>
                <span class="currency-prefix">¥</span>
              </template>
            </el-input-number>
          </el-form-item>
        </div>
        <el-form-item label="开票日期">
          <el-date-picker
            v-model="form.invoice_date"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="选择开票日期"
            size="large"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleCreate">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Download, Upload, Link, Tickets } from '@element-plus/icons-vue'
import { financeApi } from '@/api/finance'

interface Invoice {
  id: number
  invoice_no: string
  type: string
  amount: string
  tax_rate?: string
  tax_amount: string
  voucher_no?: string
  status: string
  created_at: string
}

const invoices = ref<Invoice[]>([])
const loading = ref(false)
const filterType = ref('')
const page = ref(1)
const total = ref(0)
const showCreate = ref(false)
const saving = ref(false)

const form = ref({
  invoice_no: '',
  type: '',
  amount: null as number | null,
  tax_rate: null as number | null,
  tax_amount: null as number | null,
  invoice_date: '',
})

function formatAmount(val: string) {
  if (!val) return '0.00'
  return parseFloat(val).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

async function loadInvoices() {
  loading.value = true
  try {
    const res = await financeApi.invoices({
      page: page.value,
      type: filterType.value || undefined,
    }) as any
    invoices.value = res.data?.data || res.data?.list || []
    total.value = res.data?.total || 0
  } catch {
    invoices.value = []
  } finally {
    loading.value = false
  }
}

async function handleCreate() {
  if (saving.value) return
  if (!form.value.invoice_no || !form.value.type || form.value.amount === null) {
    ElMessage.warning('请填写必填项')
    return
  }
  saving.value = true
  try {
    await financeApi.createInvoice({
      invoice_no: form.value.invoice_no,
      type: form.value.type,
      amount: String(form.value.amount),
      tax_rate: form.value.tax_rate !== null ? String(form.value.tax_rate) : undefined,
      tax_amount: form.value.tax_amount !== null ? String(form.value.tax_amount) : undefined,
      invoice_date: form.value.invoice_date || undefined,
    })
    ElMessage.success('发票登记成功')
    showCreate.value = false
    loadInvoices()
    form.value = { invoice_no: '', type: '', amount: null, tax_rate: null, tax_amount: null, invoice_date: '' }
  } catch {
    ElMessage.error('登记失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => loadInvoices())
</script>

<style scoped lang="scss">
$success: #10B981;
$warning: #F59E0B;
$error: #EF4444;
$bg-page: #FAFBFC;
$text-primary: #1F2937;
$text-secondary: #6B7280;
$text-muted: #9CA3AF;
$border-color: #E5E7EB;
$shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
$radius-sm: 8px;
$radius-md: 12px;
$radius-lg: 16px;
$radius-xl: 24px;

.invoice-list { padding: 24px 32px; width: 100%; box-sizing: border-box; background: $bg-page; min-height: 100vh; }

.glass-card { background: rgba(255,255,255,0.95); backdrop-filter: blur(12px); border: 1px solid rgba(255,255,255,0.6); border-radius: $radius-xl; box-shadow: $shadow-md; }

.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 24px;
  .page-title { font-size: 24px; font-weight: 700; color: $text-primary; margin: 0 0 4px; }
  .page-subtitle { font-size: 14px; color: $text-secondary; margin: 0; }
}

.header-actions :deep(.el-button) { padding: 12px 24px; border-radius: $radius-md; font-weight: 600; display: inline-flex; align-items: center; gap: 6px; }

.filter-bar { padding: 16px 20px; margin-bottom: 20px; }

.filter-tabs { display: flex; gap: 8px;
  :deep(.el-radio-button__inner) { padding: 10px 20px; border-radius: $radius-md !important; display: inline-flex; align-items: center; gap: 6px; }
  :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) { background: var(--primary); border-color: var(--primary); }
}

.table-container { padding: 0; overflow: hidden; }

:deep(.modern-table) {
  .el-table__header th { padding: 16px 12px; font-size: 13px; }
  .el-table__row { transition: background 0.2s ease; &:hover > td { background: rgba(var(--primary), 0.02) !important; } }
  .el-table__cell { padding: 16px 12px; border-bottom: 1px solid #F3F4F6; }
}

.invoice-no { font-weight: 600; color: var(--primary); font-family: 'SF Mono', Monaco, monospace; }

.type-badge { display: inline-flex; align-items: center; gap: 4px; padding: 4px 12px; font-size: 12px; font-weight: 500; border-radius: 20px;
  &.type--input { background: #D1FAE5; color: #059669; }
  &.type--output { background: #FEF3C7; color: #D97706; }
  .el-icon { font-size: 14px; }
}

.amount { font-weight: 600; color: $text-primary; font-family: 'SF Mono', Monaco, monospace; }
.tax-rate { color: $text-secondary; }
.tax-amount { color: $warning; font-weight: 600; font-family: 'SF Mono', Monaco, monospace; }

.voucher-link { display: inline-flex; align-items: center; gap: 4px; color: var(--primary); font-weight: 500; cursor: pointer;
  .el-icon { font-size: 14px; }
  &:hover { text-decoration: underline; }
}
.no-voucher { color: $text-muted; }

.pagination-wrapper { display: flex; justify-content: flex-end; padding: 16px 20px; border-top: 1px solid $border-color; }

.create-form { :deep(.el-form-item) { margin-bottom: 20px; } :deep(.el-form-item__label) { font-weight: 500; color: $text-secondary; } }

.type-selector { display: flex; gap: 12px;
  :deep(.el-radio) { margin-right: 0; flex: 1;
    .el-radio__input { display: none; }
  }
  :deep(.el-radio__label) { width: 100%; }
}

.radio-card { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 20px; border: 2px solid $border-color; border-radius: $radius-md; cursor: pointer; transition: all 0.2s ease;
  .el-icon { font-size: 24px; }
  span { font-weight: 500; color: $text-secondary; }
  :deep(.is-checked) ~ .radio-card { border-color: var(--primary); background: rgba(var(--primary), 0.04);
    .el-icon, span { color: var(--primary); }
  }
}

.form-row { display: flex; gap: 16px; }
.form-item-half { flex: 1; }

.currency-prefix { font-weight: 600; color: $text-secondary; }

@media (max-width: 768px) {
  .invoice-list { padding: 16px; }
  .filter-tabs { flex-wrap: wrap; }
  .form-row { flex-direction: column; }
}
</style>
