<template>
  <div class="invoice-list">
    <div class="toolbar">
      <el-radio-group v-model="filterType" @change="loadInvoices">
        <el-radio-button label="">全部</el-radio-button>
        <el-radio-button label="input">进项</el-radio-button>
        <el-radio-button label="output">销项</el-radio-button>
      </el-radio-group>
      <el-button type="primary" @click="showCreate = true">登记发票</el-button>
    </div>

    <el-table :data="invoices" stripe v-loading="loading" class="mt-2">
      <el-table-column prop="invoice_no" label="发票号" width="180" />
      <el-table-column prop="type" label="类型" width="80">
        <template #default="{ row }">
          <el-tag :type="row.type === 'input' ? 'success' : 'warning'" size="small">
            {{ row.type === 'input' ? '进项' : '销项' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="amount" label="金额" align="right" width="130">
        <template #default="{ row }">{{ formatAmount(row.amount) }}</template>
      </el-table-column>
      <el-table-column prop="tax_rate" label="税率" width="80">
        <template #default="{ row }">{{ row.tax_rate || '-' }}</template>
      </el-table-column>
      <el-table-column prop="tax_amount" label="税额" align="right" width="120">
        <template #default="{ row }">{{ formatAmount(row.tax_amount) }}</template>
      </el-table-column>
      <el-table-column prop="voucher_no" label="关联凭证" width="130">
        <template #default="{ row }">{{ row.voucher_no || '-' }}</template>
      </el-table-column>
      <el-table-column prop="created_at" label="登记日期" width="160" />
    </el-table>

    <el-pagination
      v-model:current-page="page"
      :page-size="20"
      :total="total"
      layout="prev, pager, next"
      class="mt-2"
      @current-change="loadInvoices"
    />

    <!-- Create Dialog -->
    <el-dialog v-model="showCreate" title="登记发票" width="520px">
      <el-form :model="form" label-width="100px" ref="formRef">
        <el-form-item label="发票号码" prop="invoice_no" required>
          <el-input v-model="form.invoice_no" placeholder="请输入发票号码" />
        </el-form-item>
        <el-form-item label="发票类型" prop="type" required>
          <el-select v-model="form.type" placeholder="请选择类型">
            <el-option label="进项发票" value="input" />
            <el-option label="销项发票" value="output" />
          </el-select>
        </el-form-item>
        <el-form-item label="金额(元)" prop="amount" required>
          <el-input-number v-model="form.amount" :precision="2" :min="0" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="税率(%)">
          <el-input-number v-model="form.tax_rate" :precision="2" :min="0" :max="100" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="税额(元)">
          <el-input-number v-model="form.tax_amount" :precision="2" :min="0" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="开票日期">
          <el-date-picker v-model="form.invoice_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleCreate">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
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
const formRef = ref()

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
    form.value = { invoice_no: '', type: '', amount: null, tax_rate: null, tax_amount: null, invoice_date: '' }
    loadInvoices()
  } catch (e: unknown) {
    const msg = (e as any)?.response?.data?.error || '保存失败'
    ElMessage.error(msg)
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadInvoices()
})
</script>

<style scoped lang="scss">
.invoice-list {
  padding: 8px;
  .toolbar {
    display: flex;
    gap: 8px;
    align-items: center;
    flex-wrap: wrap;
  }
  .mt-2 {
    margin-top: 12px;
  }
}
</style>
