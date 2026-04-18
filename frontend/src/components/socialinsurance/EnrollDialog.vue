<template>
  <el-dialog
    v-model="visible"
    title="增员"
    width="480px"
    destroy-on-close
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="formRules"
      label-width="100px"
      @submit.prevent
    >
      <el-form-item label="员工姓名" prop="employeeID">
        <el-select
          v-model="form.employeeID"
          filterable
          remote
          reserve-keyword
          placeholder="输入姓名搜索"
          :remote-method="searchEmployee"
          :loading="searching"
          style="width: 100%"
          @change="onEmployeeSelected"
        >
          <el-option
            v-for="emp in employeeOptions"
            :key="emp.id"
            :label="emp.name"
            :value="emp.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="身份证号">
        <el-input v-model="form.idNumber" readonly placeholder="选择员工后自动填充" />
      </el-form-item>

      <el-form-item label="起始月份" prop="startYearMonth">
        <el-date-picker
          v-model="form.startYearMonth"
          type="month"
          placeholder="选择起始月份"
          format="YYYY年MM月"
          value-format="YYYYMM"
          :disabled-date="disableStartDate"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item label="缴费城市" prop="cityCode">
        <el-input v-model="form.cityCode" placeholder="输入缴费城市" />
      </el-form-item>

      <el-form-item label="社保基数" prop="siBase">
        <el-input-number
          v-model="form.siBase"
          :min="0"
          :precision="2"
          :controls="false"
          placeholder="请输入社保基数"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item label="公积金基数" prop="hfBase">
        <el-input-number
          v-model="form.hfBase"
          :min="0"
          :precision="2"
          :controls="false"
          placeholder="选填"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item label="公积金比例" prop="hfRatio">
        <el-input-number
          v-model="form.hfRatio"
          :min="1"
          :max="24"
          :step="0.5"
          :precision="1"
          :controls="false"
          placeholder="默认12%"
          style="width: 100%"
        />
        <span class="form-unit">%</span>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">
        确认增员
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import axios from '@/api/request'
import dayjs from 'dayjs'

interface EmployeeOption {
  id: number
  name: string
  id_number?: string
}

interface EnrollForm {
  employeeID: number | undefined
  idNumber: string
  startYearMonth: string
  cityCode: string
  siBase: number | undefined
  hfBase: number | undefined
  hfRatio: number | undefined
}

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
}>()

const visible = ref(false)
const formRef = ref<FormInstance>()
const searching = ref(false)
const submitting = ref(false)
const employeeOptions = ref<EmployeeOption[]>([])

const form = reactive<EnrollForm>({
  employeeID: undefined,
  idNumber: '',
  startYearMonth: dayjs().format('YYYYMM'),
  cityCode: '',
  siBase: undefined,
  hfBase: undefined,
  hfRatio: 12,
})

const formRules: FormRules = {
  employeeID: [{ required: true, message: '请选择员工', trigger: 'change' }],
  startYearMonth: [{ required: true, message: '请选择起始月份', trigger: 'change' }],
  cityCode: [{ required: true, message: '请输入缴费城市', trigger: 'blur' }],
  siBase: [{ required: true, message: '请输入社保基数', trigger: 'blur' }],
}

watch(
  () => props.modelValue,
  (val) => {
    visible.value = val
  },
)

watch(visible, (val) => {
  emit('update:modelValue', val)
})

function disableStartDate(date: Date): boolean {
  const now = dayjs()
  const threeMonthsAgo = now.subtract(3, 'month').startOf('month')
  return dayjs(date).isBefore(threeMonthsAgo) || dayjs(date).isAfter(now.endOf('month'))
}

async function searchEmployee(query: string): Promise<void> {
  if (!query || query.length === 0) {
    employeeOptions.value = []
    return
  }

  searching.value = true
  try {
    const res = await axios.get('/api/v1/employees/search', {
      params: { name: query },
    })
    const responseData = (res as { data?: EmployeeOption[] })?.data ?? res
    employeeOptions.value = responseData as EmployeeOption[]
  } catch {
    employeeOptions.value = []
  } finally {
    searching.value = false
  }
}

function onEmployeeSelected(employeeID: number): void {
  const employee = employeeOptions.value.find((e) => e.id === employeeID)
  if (employee) {
    form.idNumber = employee.id_number ?? ''
  }
}

async function handleSubmit(): Promise<void> {
  if (!formRef.value) return

  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  if (form.hfRatio && form.hfRatio > 0 && (!form.hfBase || form.hfBase <= 0)) {
    ElMessage.warning('设置了公积金比例时，公积金基数不能为0')
    return
  }

  submitting.value = true
  try {
    await axios.post('/api/v1/social-insurance/enroll/single', {
      employeeID: form.employeeID,
      startYearMonth: form.startYearMonth,
      cityCode: form.cityCode,
      siBase: form.siBase,
      hfBase: form.hfBase,
      hfRatio: form.hfRatio,
    })
    ElMessage.success('增员成功')
    emit('success')
    handleClose()
  } catch {
    ElMessage.error('增员失败，请重试')
  } finally {
    submitting.value = false
  }
}

function handleClose(): void {
  visible.value = false
  form.employeeID = undefined
  form.idNumber = ''
  form.startYearMonth = dayjs().format('YYYYMM')
  form.cityCode = ''
  form.siBase = undefined
  form.hfBase = undefined
  form.hfRatio = 12
  employeeOptions.value = []
  formRef.value?.resetFields()
}
</script>

<style scoped lang="scss">
.form-unit {
  margin-left: 8px;
  color: #8c8c8c;
  font-size: 14px;
}
</style>
