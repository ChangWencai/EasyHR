<template>
  <el-dialog v-model="dialogVisible" title="创建登记表" width="500px" @close="handleClose">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
      <el-form-item label="姓名" prop="name">
        <el-input v-model="form.name" placeholder="请输入员工姓名" />
      </el-form-item>
      <el-form-item label="部门" prop="department_id">
        <el-select v-model="form.department_id" placeholder="请选择部门" clearable style="width: 100%">
          <el-option
            v-for="dept in departments"
            :key="dept.id"
            :label="dept.name"
            :value="dept.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="岗位" prop="position">
        <el-input v-model="form.position" placeholder="请输入岗位" />
      </el-form-item>
      <el-form-item label="入职日期" prop="hire_date">
        <el-date-picker
          v-model="form.hire_date"
          type="date"
          placeholder="请选择入职日期"
          value-format="YYYY-MM-DD"
          style="width: 100%"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">确认</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { registrationApi } from '@/api/employee'
import { departmentApi, type Department } from '@/api/department'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'created'): void
}>()

const dialogVisible = ref(false)
const formRef = ref<FormInstance>()
const submitting = ref(false)
const departments = ref<Department[]>([])

const form = ref({
  name: '',
  department_id: undefined as number | undefined,
  position: '',
  hire_date: '',
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入员工姓名', trigger: 'blur' }],
  position: [{ required: true, message: '请输入岗位', trigger: 'blur' }],
  hire_date: [{ required: true, message: '请选择入职日期', trigger: 'change' }],
}

watch(
  () => props.visible,
  (val) => {
    dialogVisible.value = val
    if (val) {
      loadDepartments()
    }
  },
)

async function loadDepartments() {
  try {
    departments.value = await departmentApi.list()
  } catch {
    // 部门加载失败不阻塞表单
  }
}

function handleClose() {
  emit('update:visible', false)
  form.value = { name: '', department_id: undefined, position: '', hire_date: '' }
}

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate()

  submitting.value = true
  try {
    await registrationApi.create({
      name: form.value.name,
      department_id: form.value.department_id,
      position: form.value.position,
      hire_date: form.value.hire_date,
    })
    ElMessage.success('登记表创建成功')
    emit('created')
  } catch {
    ElMessage.error('创建失败')
  } finally {
    submitting.value = false
  }
}
</script>
