<template>
  <el-dialog
    v-model="visible"
    title="社保明细"
    width="640px"
    destroy-on-close
    @close="handleClose"
  >
    <div v-loading="loading">
      <el-table
        :data="insuranceItems"
        border
        size="small"
        :summary-method="getSummary"
        show-summary
      >
        <el-table-column prop="name" label="险种" width="120" />
        <el-table-column label="单位缴纳（元）" align="right">
          <template #default="{ row }">
            {{ formatCurrency(row.company_amount) }}
          </template>
        </el-table-column>
        <el-table-column label="个人缴纳（元）" align="right">
          <template #default="{ row }">
            {{ formatCurrency(row.personal_amount) }}
          </template>
        </el-table-column>
      </el-table>
    </div>

    <template #footer>
      <el-button @click="handleClose">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { siApi } from '@/api/socialinsurance'

interface InsuranceAmountDetail {
  name: string
  base: number
  company_rate: number
  company_amount: number
  personal_rate: number
  personal_amount: number
}

interface SIDetailData {
  details: InsuranceAmountDetail[]
  total_company: number
  total_personal: number
}

const props = defineProps<{
  modelValue: boolean
  recordId?: number
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const visible = ref(false)
const loading = ref(false)
const detail = ref<SIDetailData>({
  details: [],
  total_company: 0,
  total_personal: 0,
})

watch(
  () => props.modelValue,
  (val) => {
    visible.value = val
    if (val && props.recordId) {
      fetchDetail(props.recordId)
    }
  },
)

watch(visible, (val) => {
  emit('update:modelValue', val)
})

const insuranceItems = computed<InsuranceAmountDetail[]>(() => {
  return detail.value.details || []
})

function formatCurrency(value: number | string | undefined): string {
  if (value === undefined || value === null) return '0.00'
  return Number(value).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function getSummary(param: { columns: { property: string }[]; data: InsuranceAmountDetail[] }): string[] {
  const { columns, data } = param
  const sums: string[] = []
  columns.forEach((column, index) => {
    if (index === 0) {
      sums[index] = '合计'
      return
    }
    const values = data.map((item) => {
      const field = column.property === 'company_amount' ? 'company_amount' : 'personal_amount'
      return Number(item[field]) || 0
    })
    const sum = values.reduce((acc, val) => acc + val, 0)
    sums[index] = formatCurrency(sum)
  })
  return sums
}

async function fetchDetail(recordId: number): Promise<void> {
  loading.value = true
  try {
    const data = await siApi.recordDetail(recordId)
    // details is a JSON string from backend, parse if needed
    const raw = (data as any).details
    const parsed = typeof raw === 'string' ? JSON.parse(raw) : (raw || [])
    detail.value = {
      details: parsed,
      total_company: data.total_company || 0,
      total_personal: data.total_personal || 0,
    }
  } catch {
    ElMessage.error('加载明细失败')
  } finally {
    loading.value = false
  }
}

function handleClose(): void {
  visible.value = false
  detail.value = { details: [], total_company: 0, total_personal: 0 }
}
</script>

<style scoped lang="scss">
</style>
