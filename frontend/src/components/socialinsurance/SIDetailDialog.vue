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
            {{ formatCurrency(row.companyAmount) }}
          </template>
        </el-table-column>
        <el-table-column label="个人缴纳（元）" align="right">
          <template #default="{ row }">
            {{ formatCurrency(row.personalAmount) }}
          </template>
        </el-table-column>
      </el-table>

      <el-descriptions
        v-if="hasOtherFees"
        title="其他缴费"
        :column="2"
        border
        size="small"
        class="other-fees"
      >
        <el-descriptions-item label="滞纳金">
          {{ formatCurrency(detail.otherFees.lateFee) }}
        </el-descriptions-item>
        <el-descriptions-item label="残保金">
          {{ formatCurrency(detail.otherFees.disabilityFee) }}
        </el-descriptions-item>
        <el-descriptions-item label="漏缴">
          {{ formatCurrency(detail.otherFees.missedFee) }}
        </el-descriptions-item>
        <el-descriptions-item label="补缴">
          {{ formatCurrency(detail.otherFees.backPayFee) }}
        </el-descriptions-item>
      </el-descriptions>
    </div>

    <template #footer>
      <el-button @click="handleClose">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import axios from '@/api/request'

interface InsuranceItem {
  name: string
  companyAmount: number
  personalAmount: number
}

interface OtherFees {
  lateFee: number
  disabilityFee: number
  missedFee: number
  backPayFee: number
}

interface SIDetailData {
  pensionCompany: number
  pensionPersonal: number
  medicalCompany: number
  medicalPersonal: number
  unemploymentCompany: number
  unemploymentPersonal: number
  injuryCompany: number
  injuryPersonal: number
  maternityCompany: number
  maternityPersonal: number
  housingFundCompany: number
  housingFundPersonal: number
  otherFees: OtherFees
}

const defaultOtherFees: OtherFees = {
  lateFee: 0,
  disabilityFee: 0,
  missedFee: 0,
  backPayFee: 0,
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
  pensionCompany: 0,
  pensionPersonal: 0,
  medicalCompany: 0,
  medicalPersonal: 0,
  unemploymentCompany: 0,
  unemploymentPersonal: 0,
  injuryCompany: 0,
  injuryPersonal: 0,
  maternityCompany: 0,
  maternityPersonal: 0,
  housingFundCompany: 0,
  housingFundPersonal: 0,
  otherFees: { ...defaultOtherFees },
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

const insuranceItems = computed<InsuranceItem[]>(() => [
  { name: '养老保险', companyAmount: detail.value.pensionCompany, personalAmount: detail.value.pensionPersonal },
  { name: '医疗保险', companyAmount: detail.value.medicalCompany, personalAmount: detail.value.medicalPersonal },
  { name: '失业保险', companyAmount: detail.value.unemploymentCompany, personalAmount: detail.value.unemploymentPersonal },
  { name: '工伤保险', companyAmount: detail.value.injuryCompany, personalAmount: detail.value.injuryPersonal },
  { name: '生育保险', companyAmount: detail.value.maternityCompany, personalAmount: detail.value.maternityPersonal },
  { name: '住房公积金', companyAmount: detail.value.housingFundCompany, personalAmount: detail.value.housingFundPersonal },
])

const hasOtherFees = computed(() => {
  const fees = detail.value.otherFees
  return fees.lateFee > 0 || fees.disabilityFee > 0 || fees.missedFee > 0 || fees.backPayFee > 0
})

function formatCurrency(value: number | string | undefined): string {
  if (value === undefined || value === null) return '0.00'
  return Number(value).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function getSummary(param: { columns: { property: string }[]; data: InsuranceItem[] }): string[] {
  const { columns, data } = param
  const sums: string[] = []
  columns.forEach((column, index) => {
    if (index === 0) {
      sums[index] = '合计'
      return
    }
    const values = data.map((item) => {
      const field = column.property === 'companyAmount' ? 'companyAmount' : 'personalAmount'
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
    const res = await axios.get(`/api/v1/social-insurance/monthly-records/${recordId}`)
    const responseData = (res as { data?: SIDetailData })?.data ?? res
    detail.value = {
      ...(responseData as SIDetailData),
      otherFees: {
        ...defaultOtherFees,
        ...((responseData as SIDetailData).otherFees || {}),
      },
    }
  } catch {
    ElMessage.error('加载明细失败')
  } finally {
    loading.value = false
  }
}

function handleClose(): void {
  visible.value = false
  detail.value = {
    pensionCompany: 0,
    pensionPersonal: 0,
    medicalCompany: 0,
    medicalPersonal: 0,
    unemploymentCompany: 0,
    unemploymentPersonal: 0,
    injuryCompany: 0,
    injuryPersonal: 0,
    maternityCompany: 0,
    maternityPersonal: 0,
    housingFundCompany: 0,
    housingFundPersonal: 0,
    otherFees: { ...defaultOtherFees },
  }
}
</script>

<style scoped lang="scss">
.other-fees {
  margin-top: 16px;
}
</style>
