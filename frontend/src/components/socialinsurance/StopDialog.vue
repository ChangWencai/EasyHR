<template>
  <el-dialog
    v-model="visible"
    title="减员"
    width="480px"
    destroy-on-close
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="formRules"
      label-width="110px"
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

      <el-form-item label="终止月份" prop="stopYearMonth">
        <el-date-picker
          v-model="form.stopYearMonth"
          type="month"
          placeholder="选择终止月份"
          format="YYYY年MM月"
          value-format="YYYYMM"
          :disabled-date="disableStopDate"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item label="减员原因" prop="reason">
        <el-select v-model="form.reason" placeholder="请选择减员原因" style="width: 100%">
          <el-option label="跳槽" value="job_change" />
          <el-option label="退休" value="retirement" />
          <el-option label="其他" value="other" />
        </el-select>
      </el-form-item>

      <el-form-item label="转出日期">
        <div class="date-field-with-tooltip">
          <el-date-picker
            v-model="form.transferDate"
            type="date"
            placeholder="选择转出日期"
            value-format="YYYY-MM-DD"
            style="flex: 1"
          />
          <el-tooltip
            placement="top"
          >
            <template #content>
              <div class="rule-tooltip">
                <p>转出生效规则：</p>
                <p>- 每月5日前转出 → 当月生效</p>
                <p>- 5-25日转出 → 次月生效</p>
                <p>- 25日后转出 → 下下月生效</p>
              </div>
            </template>
            <el-icon class="tooltip-icon"><WarningFilled /></el-icon>
          </el-tooltip>
        </div>
      </el-form-item>

      <el-form-item label="公积金封存日期">
        <el-date-picker
          v-model="form.hfFreezeDate"
          type="date"
          placeholder="默认同转出日期"
          value-format="YYYY-MM-DD"
          style="width: 100%"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="danger" :loading="submitting" @click="handleSubmit">
        确认减员
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { WarningFilled } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import axios from '@/api/request'
import dayjs from 'dayjs'

interface EmployeeOption {
  id: number
  name: string
  id_number?: string
}

interface StopFormData {
  employeeID: number | undefined
  idNumber: string
  stopYearMonth: string
  reason: string
  transferDate: string
  hfFreezeDate: string
}

const props = defineProps<{
  modelValue: boolean
  employeeId?: number
  employeeName?: string
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

const form = reactive<StopFormData>({
  employeeID: undefined,
  idNumber: '',
  stopYearMonth: dayjs().format('YYYYMM'),
  reason: '',
  transferDate: '',
  hfFreezeDate: '',
})

const formRules: FormRules = {
  employeeID: [{ required: true, message: '请选择员工', trigger: 'change' }],
  stopYearMonth: [{ required: true, message: '请选择终止月份', trigger: 'change' }],
  reason: [{ required: true, message: '请选择减员原因', trigger: 'change' }],
}

watch(
  () => props.modelValue,
  (val) => {
    visible.value = val
    if (val && props.employeeId) {
      form.employeeID = props.employeeId
      if (props.employeeName) {
        employeeOptions.value = [{ id: props.employeeId, name: props.employeeName }]
      }
    }
  },
)

watch(visible, (val) => {
  emit('update:modelValue', val)
})

watch(
  () => form.transferDate,
  (newDate) => {
    if (newDate && !form.hfFreezeDate) {
      form.hfFreezeDate = newDate
    }
  },
)

function disableStopDate(date: Date): boolean {
  const currentMonthStart = dayjs().startOf('month')
  return dayjs(date).isBefore(currentMonthStart)
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

  const currentMonth = dayjs().format('YYYYMM')
  if (form.stopYearMonth < currentMonth) {
    ElMessage.error('终止月份不可早于当月')
    return
  }

  try {
    await ElMessageBox.confirm(
      '减员后不可逆，是否确认停缴？',
      '确认减员',
      {
        confirmButtonText: '确认减员',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )
  } catch {
    return
  }

  submitting.value = true
  try {
    await axios.post('/api/v1/socialinsurance/stop', {
      employeeID: form.employeeID,
      stopYearMonth: form.stopYearMonth,
      reason: form.reason,
      transferDate: form.transferDate,
      hfFreezeDate: form.hfFreezeDate,
    })
    ElMessage.success('减员成功')
    emit('success')
    handleClose()
  } catch {
    ElMessage.error('减员失败，请重试')
  } finally {
    submitting.value = false
  }
}

function handleClose(): void {
  visible.value = false
  form.employeeID = undefined
  form.idNumber = ''
  form.stopYearMonth = dayjs().format('YYYYMM')
  form.reason = ''
  form.transferDate = ''
  form.hfFreezeDate = ''
  employeeOptions.value = []
  formRef.value?.resetFields()
}
</script>

<style scoped lang="scss">
.date-field-with-tooltip {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.tooltip-icon {
  font-size: 16px;
  color: #e6a23c;
  cursor: pointer;
  flex-shrink: 0;
}

.rule-tooltip {
  line-height: 1.8;

  p {
    margin: 0;
  }
}
</style>
