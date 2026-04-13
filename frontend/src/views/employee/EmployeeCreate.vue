<template>
  <div class="employee-create">
    <!-- 页面头部 -->
    <div class="header">
      <button class="back-btn" @click="$router.back()" aria-label="返回">
        <el-icon :size="20"><ArrowLeft /></el-icon>
      </button>
      <h1 class="header-title">{{ isEdit ? '编辑员工' : '新增员工' }}</h1>
      <div style="width: 36px" /> <!-- 占位平衡 -->
    </div>

    <el-card class="form-card" v-loading="saving" element-loading-text="保存中…">
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <!-- 基本信息 -->
        <div class="form-section">
          <div class="section-label">
            <el-icon :size="14" color="#0F766E"><UserFilled /></el-icon>
            基本信息
          </div>
          <el-form-item label="姓名" prop="name">
            <el-input v-model="form.name" placeholder="请输入员工姓名" maxlength="50" clearable />
          </el-form-item>
          <el-form-item label="手机号" prop="phone">
            <el-input v-model="form.phone" placeholder="请输入手机号" maxlength="11" clearable />
          </el-form-item>
          <el-form-item label="身份证号" prop="id_number">
            <el-input v-model="form.id_number" placeholder="请输入18位身份证号" maxlength="18" clearable />
          </el-form-item>
          <el-form-item label="岗位" prop="position">
            <el-input v-model="form.position" placeholder="请输入岗位名称" maxlength="100" clearable />
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
        </div>

        <!-- 薪资信息 -->
        <div class="form-section">
          <div class="section-label">
            <el-icon :size="14" color="#F59E0B"><Coin /></el-icon>
            薪资信息
          </div>
          <el-form-item label="正式薪资（元/月）" prop="salary">
            <el-input-number v-model="form.salary" :min="0" :precision="2" placeholder="税前薪资" style="width: 100%" controls-position="right" />
          </el-form-item>
          <el-form-item label="试用期薪资（元/月）" prop="probation_salary">
            <el-input-number v-model="form.probation_salary" :min="0" :precision="2" placeholder="试用期薪资（可选）" style="width: 100%" controls-position="right" />
            <div class="form-tip">
              <el-icon :size="12"><InfoFilled /></el-icon>
              试用期薪资通常为正式薪资的 80%
            </div>
          </el-form-item>
          <el-form-item label="工资卡号" prop="bank_card">
            <el-input v-model="form.bank_card" placeholder="银行卡号（可选）" maxlength="30" clearable />
          </el-form-item>
        </div>

        <!-- 紧急联系人 -->
        <div class="form-section">
          <div class="section-label">
            <el-icon :size="14" color="#EF4444"><Phone /></el-icon>
            紧急联系人
          </div>
          <el-form-item label="联系人姓名" prop="emergency_contact">
            <el-input v-model="form.emergency_contact" placeholder="紧急联系人姓名（可选）" maxlength="50" clearable />
          </el-form-item>
          <el-form-item label="联系人电话" prop="emergency_phone">
            <el-input v-model="form.emergency_phone" placeholder="紧急联系人电话（可选）" maxlength="11" clearable />
          </el-form-item>
        </div>

        <!-- 提交按钮 -->
        <div style="margin-top: 24px; display: flex; flex-direction: column; gap: 10px">
          <el-button
            type="primary"
            class="submit-btn"
            @click="handleSubmit"
            :loading="saving"
          >
            {{ isEdit ? '保存修改' : '创建员工' }}
          </el-button>
          <el-button class="cancel-btn" @click="$router.back()">取消</el-button>
        </div>
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
import {
  ArrowLeft,
  UserFilled,
  Coin,
  InfoFilled,
  Phone,
} from '@element-plus/icons-vue'

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
  if (saving.value) return
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
.employee-create {
  padding-bottom: 80px;
  background: #F8FAFC;
  min-height: 100%;
}
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  padding: 16px;
  border-bottom: 1px solid #F1F5F9;
  position: sticky;
  top: 0;
  z-index: 10;
}

.header-title {
  font-size: 17px;
  font-weight: 600;
  color: #0F172A;
}

.form-card {
  margin: 12px;
  border-radius: 16px;
  border: none;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
}

.form-section {
  margin-bottom: 8px;

  & + & {
    padding-top: 16px;
    border-top: 1px solid #F1F5F9;
  }
}

.section-label {
  font-size: 13px;
  font-weight: 600;
  color: #0F172A;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.form-tip {
  font-size: 12px;
  color: #94A3B8;
  margin-top: 6px;
  display: flex;
  align-items: flex-start;
  gap: 4px;
}

.submit-btn {
  width: 100%;
  height: 48px;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  background: #0F766E;
  border-color: #0F766E;
  cursor: pointer;
  transition: background 0.15s, transform 0.1s;

  &:hover { background: #0D6B62; }
  &:active { transform: scale(0.99); }
  &:disabled { background: #CBD5E1; border-color: #CBD5E1; cursor: not-allowed; }
}

.cancel-btn {
  width: 100%;
  height: 48px;
  border-radius: 12px;
  font-size: 16px;
  cursor: pointer;
  background: #fff;
  border: 1px solid #E2E8F0;
  color: #64748B;
  transition: all 0.15s;

  &:hover { border-color: #0F766E; color: #0F766E; }
  &:active { background: #F1F5F9; }
}

// Element Plus overrides
:deep(.el-form-item) {
  margin-bottom: 14px;
}

:deep(.el-form-item__label) {
  font-size: 14px;
  color: #0F172A;
  font-weight: 500;
  padding-bottom: 6px !important;
  line-height: 1.4;
}

:deep(.el-input__wrapper),
:deep(.el-textarea__inner) {
  border-radius: 10px;
  border-color: #E2E8F0;
  box-shadow: none !important;
  padding: 10px 12px;
  font-size: 15px;

  &:focus-within {
    border-color: #0F766E;
  }
}

:deep(.el-input-number .el-input__wrapper) {
  padding-left: 12px;
  padding-right: 12px;
}

:deep(.el-input-number__decrease),
:deep(.el-input-number__increase) {
  background: #F8FAFC;
  border-color: #E2E8F0;
  color: #64748B;

  &:hover { color: #0F766E; }
}

:deep(.el-date-editor) {
  width: 100% !important;
}

:deep(.el-form-item__error) {
  font-size: 12px;
  padding-top: 4px;
}

:deep(.el-button--primary) {
  background: #0F766E;
  border-color: #0F766E;

  &:hover { background: #0D6B62; border-color: #0D6B62; }
  &:active { background: #115E59; }
}
</style>
