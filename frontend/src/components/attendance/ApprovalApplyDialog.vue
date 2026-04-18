<template>
  <el-dialog v-model="visible" title="新建申请" width="480px" @close="resetForm">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
      <el-form-item label="申请类型" prop="approval_type">
        <el-select v-model="form.approval_type" placeholder="选择类型" style="width: 100%">
          <el-option-group label="假勤申请">
            <el-option v-for="t in leaveTypes" :key="t.value" :label="t.label" :value="t.value" />
          </el-option-group>
          <el-option-group label="其他申请">
            <el-option v-for="t in otherTypes" :key="t.value" :label="t.label" :value="t.value" />
          </el-option-group>
        </el-select>
      </el-form-item>

      <el-form-item v-if="isLeaveType" label="请假类型" prop="leave_type">
        <el-select v-model="form.leave_type" placeholder="选择请假类型" style="width: 100%">
          <el-option v-for="t in leaveTypes" :key="t.value" :label="t.label" :value="t.value" />
        </el-select>
      </el-form-item>

      <el-form-item label="开始时间" prop="start_time">
        <el-date-picker
          v-model="form.start_time"
          type="datetime"
          placeholder="选择开始时间"
          format="YYYY-MM-DD HH:mm"
          value-format="YYYY-MM-DDTHH:mm:ssZ"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item label="结束时间" prop="end_time">
        <el-date-picker
          v-model="form.end_time"
          type="datetime"
          placeholder="选择结束时间"
          format="YYYY-MM-DD HH:mm"
          value-format="YYYY-MM-DDTHH:mm:ssZ"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item v-if="durationDisplay" label="时长">
        <span class="duration-text">{{ durationDisplay }}</span>
      </el-form-item>

      <el-form-item label="事由" prop="reason">
        <el-input v-model="form.reason" type="textarea" :rows="3" placeholder="请输入事由" maxlength="500" show-word-limit />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">提交申请</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { approvalApi, type CreateApprovalRequest } from '@/api/attendance'
import dayjs from 'dayjs'

const emit = defineEmits<{ (e: 'submitted'): void }>()

const visible = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const form = ref<CreateApprovalRequest>({
  approval_type: '',
  start_time: '',
  end_time: '',
  reason: '',
  leave_type: '',
})

const rules: FormRules = {
  approval_type: [{ required: true, message: '请选择申请类型', trigger: 'change' }],
  start_time: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  end_time: [{ required: true, message: '请选择结束时间', trigger: 'change' }],
}

const leaveTypes = [
  { value: 'personal_leave', label: '事假' },
  { value: 'sick_leave', label: '病假' },
  { value: 'PTO', label: '调休' },
  { value: 'annual_leave', label: '年假' },
  { value: 'marriage_leave', label: '婚假' },
  { value: 'maternity_leave', label: '产假' },
  { value: 'paternity_leave', label: '陪产假' },
]

const otherTypes = [
  { value: 'makeup', label: '补卡申请' },
  { value: 'shift_swap', label: '调班申请' },
  { value: 'business_trip', label: '出差申请' },
  { value: 'outside', label: '外出申请' },
  { value: 'overtime', label: '加班申请' },
]

const leaveTypeValues = new Set(leaveTypes.map((t) => t.value))

const isLeaveType = computed(() => leaveTypeValues.has(form.value.approval_type))

const durationDisplay = computed(() => {
  if (!form.value.start_time || !form.value.end_time) return ''
  const start = dayjs(form.value.start_time)
  const end = dayjs(form.value.end_time)
  if (!end.isAfter(start)) return ''
  const hours = end.diff(start, 'minute') / 60
  const rounded = Math.round(hours * 100) / 100
  const days = Math.round((rounded / 8) * 10) / 10
  return `${rounded}小时（约${days}天）`
})

function open() {
  visible.value = true
}

function resetForm() {
  form.value = { approval_type: '', start_time: '', end_time: '', reason: '', leave_type: '' }
  formRef.value?.resetFields()
}

async function handleSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    await approvalApi.create(form.value)
    ElMessage.success('申请提交成功')
    visible.value = false
    emit('submitted')
  } catch {
    ElMessage.error('提交失败，请重试')
  } finally {
    submitting.value = false
  }
}

defineExpose({ open })
</script>

<style scoped>
.duration-text {
  color: var(--el-color-primary);
  font-weight: 500;
}
</style>
