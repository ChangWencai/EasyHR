<template>
  <div class="employee-create">
    <el-card v-loading="saving">
      <template #header>
        <div class="header">
          <span>{{ isEdit ? '编辑员工' : '新增员工' }}</span>
          <el-button @click="$router.back()">取消</el-button>
        </div>
      </template>

      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="姓名" prop="name">
          <el-input v-model="form.name" placeholder="请输入员工姓名" maxlength="50" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入手机号" maxlength="11" />
        </el-form-item>
        <el-form-item label="身份证号" prop="id_number">
          <el-input v-model="form.id_number" placeholder="请输入身份证号" maxlength="18" />
        </el-form-item>
        <el-form-item label="岗位" prop="position">
          <el-input v-model="form.position" placeholder="请输入岗位名称" maxlength="100" />
        </el-form-item>
        <el-form-item label="入职日期" prop="entry_date">
          <el-date-picker
            v-model="form.entry_date"
            type="date"
            placeholder="选择入职日期"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="薪资" prop="salary">
          <el-input-number v-model="form.salary" :min="0" :precision="2" placeholder="正式薪资" style="width: 100%" />
        </el-form-item>
        <el-form-item label="试用期薪资" prop="probation_salary">
          <el-input-number v-model="form.probation_salary" :min="0" :precision="2" placeholder="试用期薪资（可选）" style="width: 100%" />
        </el-form-item>
        <el-form-item label="工资卡号" prop="bank_card">
          <el-input v-model="form.bank_card" placeholder="银行卡号（可选）" maxlength="30" />
        </el-form-item>
        <el-form-item label="紧急联系人" prop="emergency_contact">
          <el-input v-model="form.emergency_contact" placeholder="紧急联系人姓名（可选）" maxlength="50" />
        </el-form-item>
        <el-form-item label="紧急联系电话" prop="emergency_phone">
          <el-input v-model="form.emergency_phone" placeholder="紧急联系人电话（可选）" maxlength="11" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="saving">{{ isEdit ? '保存' : '创建' }}</el-button>
          <el-button @click="$router.back()">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { employeeApi } from '@/api/employee'
import { ElMessage, FormInstance, FormRules } from 'element-plus'
import type { Employee } from '@/api/employee'

const route = useRoute()
const router = useRouter()
const formRef = ref<FormInstance>()
const saving = ref(false)

const isEdit = computed(() => !!route.params.id)

const form = reactive({
  name: '',
  phone: '',
  id_number: '',
  position: '',
  entry_date: '',
  salary: undefined as number | undefined,
  probation_salary: undefined as number | undefined,
  bank_card: '',
  emergency_contact: '',
  emergency_phone: '',
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确', trigger: 'blur' },
  ],
  id_number: [
    { required: true, message: '请输入身份证号', trigger: 'blur' },
    { pattern: /^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$/, message: '身份证号格式不正确', trigger: 'blur' },
  ],
  position: [{ required: true, message: '请输入岗位', trigger: 'blur' }],
  entry_date: [{ required: true, message: '请选择入职日期', trigger: 'change' }],
}

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  saving.value = true
  try {
    const data = { ...form }
    if (isEdit.value) {
      await employeeApi.update(Number(route.params.id), data)
      ElMessage.success('保存成功')
    } else {
      await employeeApi.create(data)
      ElMessage.success('创建成功')
    }
    router.push('/employee')
  } catch {
    ElMessage.error(isEdit.value ? '保存失败' : '创建失败')
  } finally {
    saving.value = false
  }
}

async function loadEmployee() {
  if (!isEdit.value) return
  try {
    const emp = await employeeApi.get(Number(route.params.id))
    Object.assign(form, {
      name: emp.name,
      phone: emp.phone,
      id_number: emp.id_number,
      position: emp.position,
      entry_date: emp.entry_date,
      salary: emp.salary,
      probation_salary: emp.probation_salary,
      bank_card: emp.bank_card || '',
      emergency_contact: emp.emergency_contact || '',
      emergency_phone: emp.emergency_phone || '',
    })
  } catch {
    ElMessage.error('加载失败')
  }
}

onMounted(() => loadEmployee())
</script>

<style scoped lang="scss">
.employee-create {
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
