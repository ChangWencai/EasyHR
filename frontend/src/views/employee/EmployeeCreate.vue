<template>
  <div class="employee-create">
    <!-- 页面头部 -->
    <header class="page-header">
      <button class="back-btn" @click="$router.back()" aria-label="返回">
        <el-icon :size="20"><ArrowLeft /></el-icon>
      </button>
      <div class="header-title-group">
        <h1 class="header-title">{{ isEdit ? '编辑员工' : '新增员工' }}</h1>
        <p class="header-subtitle">{{ isEdit ? '修改员工信息' : '完善员工基本信息' }}</p>
      </div>
      <div style="width: 36px" />
    </header>

    <!-- 创建模式：步骤向导 -->
    <div v-if="!isEdit" class="form-container">
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="left"
        label-width="110px"
        size="large"
      >
        <StepWizard
          :steps="steps"
          v-model:current-step="currentStep"
          @complete="handleCreate"
        >
          <template #default="{ step }">
            <!-- Step 0: 完整信息录入 -->
            <div v-show="step === 0">
              <StepCard title="基本信息" description="填写员工基本资料">
                <div class="form-grid">
                  <el-form-item label="姓名" prop="name" class="form-item">
                    <el-input v-model="form.name" placeholder="请输入员工姓名" maxlength="50" size="large">
                      <template #prefix><el-icon><User /></el-icon></template>
                    </el-input>
                  </el-form-item>
                  <el-form-item label="手机号" prop="phone" class="form-item">
                    <el-input v-model="form.phone" placeholder="请输入手机号" maxlength="11" size="large">
                      <template #prefix><el-icon><Phone /></el-icon></template>
                    </el-input>
                  </el-form-item>
                  <el-form-item label="性别" prop="gender" class="form-item">
                    <el-select v-model="form.gender" placeholder="请选择性别" clearable size="large" style="width: 100%">
                      <el-option label="男" value="男" />
                      <el-option label="女" value="女" />
                    </el-select>
                  </el-form-item>
                  <el-form-item label="邮箱" prop="email" class="form-item">
                    <el-input v-model="form.email" placeholder="邮箱地址（可选）" maxlength="100" size="large">
                      <template #prefix><el-icon><Message /></el-icon></template>
                    </el-input>
                  </el-form-item>
                  <el-form-item label="身份证号" prop="id_number" class="form-item form-item--full">
                    <el-input v-model="form.id_number" placeholder="请输入18位身份证号" maxlength="18" size="large">
                      <template #prefix><el-icon><Postcard /></el-icon></template>
                    </el-input>
                  </el-form-item>
                </div>
              </StepCard>
              <StepCard title="入职信息" description="填写入职相关信息">
                <div class="form-grid">
                  <el-form-item label="入职日期" prop="entry_date" class="form-item">
                    <el-date-picker
                      v-model="form.entry_date"
                      type="date"
                      placeholder="选择入职日期"
                      value-format="YYYY-MM-DD"
                      size="large"
                      style="width: 100%"
                    />
                  </el-form-item>
                  <el-form-item label="部门" prop="department_id" class="form-item">
                    <el-select
                      v-model="form.department_id"
                      placeholder="请选择部门"
                      clearable
                      size="large"
                      style="width: 100%"
                    >
                      <el-option
                        v-for="dept in flatDepartments"
                        :key="dept.id"
                        :value="dept.id"
                        :label="dept.name"
                      />
                    </el-select>
                  </el-form-item>
                  <el-form-item label="岗位" prop="position_id" class="form-item">
                    <el-select
                      v-model="form.position_id"
                      placeholder="请选择岗位"
                      clearable
                      size="large"
                      style="width: 100%"
                    >
                      <el-option-group v-if="deptPositions.length" label="部门专属岗位">
                        <el-option
                          v-for="p in deptPositions"
                          :key="p.id"
                          :value="p.id"
                          :label="p.name"
                        />
                      </el-option-group>
                      <el-option-group v-if="commonPositions.length" label="通用岗位">
                        <el-option
                          v-for="p in commonPositions"
                          :key="p.id"
                          :value="p.id"
                          :label="p.name"
                        />
                      </el-option-group>
                    </el-select>
                  </el-form-item>
                  <el-form-item label="正式薪资" prop="salary" class="form-item">
                    <el-input-number
                      v-model="form.salary"
                      :min="0"
                      :precision="2"
                      :controls="false"
                      placeholder="税前薪资"
                      size="large"
                      style="width: 100%"
                    >
                      <template #prefix><span class="currency-prefix">¥</span></template>
                    </el-input-number>
                  </el-form-item>
                  <el-form-item label="试用期薪资" prop="probation_salary" class="form-item">
                    <el-input-number
                      v-model="form.probation_salary"
                      :min="0"
                      :precision="2"
                      :controls="false"
                      placeholder="试用期薪资（可选）"
                      size="large"
                      style="width: 100%"
                    >
                      <template #prefix><span class="currency-prefix">¥</span></template>
                    </el-input-number>
                  </el-form-item>
                  <el-form-item label="工资卡号" prop="bank_card" class="form-item">
                    <el-input
                      v-model="form.bank_card"
                      placeholder="银行卡号（可选）"
                      maxlength="30"
                      size="large"
                    >
                      <template #prefix><el-icon><Postcard /></el-icon></template>
                    </el-input>
                  </el-form-item>
                  <el-form-item label="开户行" prop="bank_name" class="form-item">
                    <el-input
                      v-model="form.bank_name"
                      placeholder="开户银行名称（可选）"
                      maxlength="100"
                      size="large"
                    >
                      <template #prefix><el-icon><OfficeBuilding /></el-icon></template>
                    </el-input>
                  </el-form-item>
                  <el-form-item label="地址" prop="address" class="form-item form-item--full">
                    <el-input
                      v-model="form.address"
                      placeholder="居住地址（可选）"
                      maxlength="500"
                      size="large"
                    >
                      <template #prefix><el-icon><Location /></el-icon></template>
                    </el-input>
                  </el-form-item>
                  <el-form-item label="备注" prop="remark" class="form-item form-item--full">
                    <el-input
                      v-model="form.remark"
                      type="textarea"
                      :rows="2"
                      placeholder="备注信息（可选）"
                      maxlength="500"
                      size="large"
                    />
                  </el-form-item>
                </div>
              </StepCard>
              <StepCard title="紧急联系人" description="填写紧急联系方式">
                <div class="form-grid">
                  <el-form-item label="联系人姓名" prop="emergency_contact" class="form-item">
                    <el-input
                      v-model="form.emergency_contact"
                      placeholder="紧急联系人姓名（可选）"
                      maxlength="50"
                      size="large"
                    >
                      <template #prefix><el-icon><User /></el-icon></template>
                    </el-input>
                  </el-form-item>
                  <el-form-item label="联系人电话" prop="emergency_phone" class="form-item">
                    <el-input
                      v-model="form.emergency_phone"
                      placeholder="紧急联系人电话（可选）"
                      maxlength="11"
                      size="large"
                    >
                      <template #prefix><el-icon><Phone /></el-icon></template>
                    </el-input>
                  </el-form-item>
                </div>
              </StepCard>
            </div>
          </template>
        </StepWizard>
      </el-form>
    </div>

    <!-- 编辑模式：原始表单 -->
    <div v-else class="form-container">
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        class="modern-form"
        v-loading="saving"
        element-loading-text="保存中..."
      >
        <!-- 基本信息 -->
        <div class="form-section glass-card">
          <div class="section-header">
            <div class="section-icon section-icon--user">
              <el-icon><UserFilled /></el-icon>
            </div>
            <div class="section-title-group">
              <h3 class="section-title">基本信息</h3>
              <p class="section-desc">员工的个人基本资料</p>
            </div>
          </div>
          <div class="form-grid">
            <el-form-item label="姓名" prop="name" class="form-item">
              <el-input
                v-model="form.name"
                placeholder="请输入员工姓名"
                maxlength="50"
                size="large"
              >
                <template #prefix>
                  <el-icon><User /></el-icon>
                </template>
              </el-input>
            </el-form-item>
            <el-form-item label="手机号" prop="phone" class="form-item">
              <el-input
                v-model="form.phone"
                placeholder="请输入手机号"
                maxlength="11"
                size="large"
              >
                <template #prefix>
                  <el-icon><Phone /></el-icon>
                </template>
              </el-input>
            </el-form-item>
            <el-form-item label="性别" prop="gender" class="form-item">
              <el-select v-model="form.gender" placeholder="请选择性别" clearable size="large" style="width: 100%">
                <el-option label="男" value="男" />
                <el-option label="女" value="女" />
              </el-select>
            </el-form-item>
            <el-form-item label="邮箱" prop="email" class="form-item">
              <el-input
                v-model="form.email"
                placeholder="邮箱地址（可选）"
                maxlength="100"
                size="large"
              >
                <template #prefix>
                  <el-icon><Message /></el-icon>
                </template>
              </el-input>
            </el-form-item>
            <el-form-item label="身份证号" prop="id_number" class="form-item form-item--full">
              <el-input
                v-model="form.id_number"
                placeholder="请输入18位身份证号"
                maxlength="18"
                size="large"
              >
                <template #prefix>
                  <el-icon><Postcard /></el-icon>
                </template>
              </el-input>
            </el-form-item>
            <el-form-item label="岗位" prop="position_id" class="form-item">
              <el-select
                v-model="form.position_id"
                placeholder="请选择岗位"
                clearable
                size="large"
                style="width: 100%"
              >
                <el-option value="" label="未分配岗位" />
                <el-option-group v-if="deptPositions.length" label="部门专属岗位">
                  <el-option
                    v-for="p in deptPositions"
                    :key="p.id"
                    :value="p.id"
                    :label="p.name"
                  />
                </el-option-group>
                <el-option-group v-if="commonPositions.length" label="通用岗位">
                  <el-option
                    v-for="p in commonPositions"
                    :key="p.id"
                    :value="p.id"
                    :label="p.name"
                  />
                </el-option-group>
              </el-select>
            </el-form-item>
            <el-form-item label="部门" prop="department_id" class="form-item">
              <el-select
                v-model="form.department_id"
                placeholder="请选择部门"
                clearable
                size="large"
                style="width: 100%"
              >
                <el-option
                  v-for="dept in flatDepartments"
                  :key="dept.id"
                  :value="dept.id"
                  :label="dept.name"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="入职日期" prop="entry_date" class="form-item">
              <el-date-picker
                v-model="form.entry_date"
                type="date"
                placeholder="选择入职日期"
                value-format="YYYY-MM-DD"
                size="large"
                style="width: 100%"
              />
            </el-form-item>
          </div>
        </div>

        <!-- 薪资与银行卡 -->
        <div class="form-section glass-card">
          <div class="section-header">
            <div class="section-icon section-icon--salary">
              <el-icon><Coin /></el-icon>
            </div>
            <div class="section-title-group">
              <h3 class="section-title">薪资与银行卡</h3>
              <p class="section-desc">设置员工的薪资结构和银行卡信息</p>
            </div>
          </div>
          <div class="form-grid">
            <el-form-item label="正式薪资（元/月）" prop="salary" class="form-item">
              <el-input-number
                v-model="form.salary"
                :min="0"
                :precision="2"
                :controls="false"
                placeholder="税前薪资"
                size="large"
                style="width: 100%"
              >
                <template #prefix>
                  <span class="currency-prefix">¥</span>
                </template>
              </el-input-number>
            </el-form-item>
            <el-form-item label="试用期薪资（元/月）" prop="probation_salary" class="form-item">
              <el-input-number
                v-model="form.probation_salary"
                :min="0"
                :precision="2"
                :controls="false"
                placeholder="试用期薪资（可选）"
                size="large"
                style="width: 100%"
              >
                <template #prefix>
                  <span class="currency-prefix">¥</span>
                </template>
              </el-input-number>
              <div class="form-tip">
                <el-icon><InfoFilled /></el-icon>
                试用期薪资通常为正式薪资的 80%
              </div>
            </el-form-item>
            <el-form-item label="工资卡号" prop="bank_card" class="form-item">
              <el-input
                v-model="form.bank_card"
                placeholder="银行卡号（可选）"
                maxlength="30"
                size="large"
              >
                <template #prefix>
                  <el-icon><Postcard /></el-icon>
                </template>
              </el-input>
            </el-form-item>
            <el-form-item label="开户行" prop="bank_name" class="form-item">
              <el-input
                v-model="form.bank_name"
                placeholder="开户银行名称（可选）"
                maxlength="100"
                size="large"
              >
                <template #prefix>
                  <el-icon><OfficeBuilding /></el-icon>
                </template>
              </el-input>
            </el-form-item>
          </div>
        </div>

        <!-- 联系信息 -->
        <div class="form-section glass-card">
          <div class="section-header">
            <div class="section-icon section-icon--contact">
              <el-icon><Location /></el-icon>
            </div>
            <div class="section-title-group">
              <h3 class="section-title">联系信息</h3>
              <p class="section-desc">员工的居住地址和紧急联系方式</p>
            </div>
          </div>
          <div class="form-grid">
            <el-form-item label="地址" prop="address" class="form-item form-item--full">
              <el-input
                v-model="form.address"
                placeholder="居住地址（可选）"
                maxlength="500"
                size="large"
              >
                <template #prefix>
                  <el-icon><Location /></el-icon>
                </template>
              </el-input>
            </el-form-item>
            <el-form-item label="备注" prop="remark" class="form-item form-item--full">
              <el-input
                v-model="form.remark"
                type="textarea"
                :rows="2"
                placeholder="备注信息（可选）"
                maxlength="500"
                size="large"
              />
            </el-form-item>
            <el-form-item label="紧急联系人姓名" prop="emergency_contact" class="form-item">
              <el-input
                v-model="form.emergency_contact"
                placeholder="紧急联系人姓名（可选）"
                maxlength="50"
                size="large"
              >
                <template #prefix>
                  <el-icon><User /></el-icon>
                </template>
              </el-input>
            </el-form-item>
            <el-form-item label="紧急联系人电话" prop="emergency_phone" class="form-item">
              <el-input
                v-model="form.emergency_phone"
                placeholder="紧急联系人电话（可选）"
                maxlength="11"
                size="large"
              >
                <template #prefix>
                  <el-icon><Phone /></el-icon>
                </template>
              </el-input>
            </el-form-item>
          </div>
        </div>

        <!-- 提交按钮 -->
        <div class="form-actions">
          <el-button size="large" class="cancel-btn" @click="$router.back()">
            取消
          </el-button>
          <el-button
            type="primary"
            size="large"
            class="submit-btn"
            :loading="saving"
            @click="handleSubmit"
          >
            <el-icon v-if="!saving"><Check /></el-icon>
            {{ isEdit ? '保存修改' : '创建员工' }}
          </el-button>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { employeeApi } from '@/api/employee'
import { positionApi } from '@/api/position'
import { departmentApi } from '@/api/department'
import { salaryApi } from '@/api/salary'
import type { Department } from '@/api/department'
import StepWizard from '@/components/common/StepWizard.vue'
import StepCard from '@/components/common/StepCard.vue'
import { useMessage } from '@/composables/useMessage'
import { type FormInstance, type FormRules } from 'element-plus'
import {
  ArrowLeft,
  UserFilled,
  User,
  Phone,
  Postcard,
  Check,
  OfficeBuilding,
  Location,
  Message,
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const formRef = ref<FormInstance>()
const saving = ref(false)
const $msg = useMessage()
const currentStep = ref(0)

const isEdit = computed(() => !!route.params.id)

const steps = [
  { title: '填写信息' },
]

const deptPositions = ref<Array<{ id: number; name: string }>>([])
const commonPositions = ref<Array<{ id: number; name: string }>>([])
const flatDepartments = ref<Department[]>([])

const form = reactive({
  name: '',
  phone: '',
  id_number: '',
  gender: '',
  email: '',
  position: '',
  position_id: null as number | null,
  department_id: null as number | null,
  entry_date: '',
  salary: undefined as number | undefined,
  probation_salary: undefined as number | undefined,
  bank_card: '',
  bank_name: '',
  address: '',
  remark: '',
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
  position_id: [{ required: false, message: '请选择岗位', trigger: 'change' }],
  entry_date: [{ required: true, message: '请选择入职日期', trigger: 'change' }],
}

async function handleCreate() {
  if (saving.value) return

  saving.value = true
  try {
    // 转换字段名以匹配后端 DTO
    const createData = {
      name: form.name,
      phone: form.phone,
      id_card: form.id_number,
      position: form.position,
      position_id: form.position_id,
      department_id: form.department_id,
      hire_date: form.entry_date,
      gender: form.gender,
      email: form.email,
      salary: form.salary,
      probation_salary: form.probation_salary,
      bank_account: form.bank_card,
      bank_name: form.bank_name,
      address: form.address,
      remark: form.remark,
      emergency_contact: form.emergency_contact,
      emergency_phone: form.emergency_phone,
    }
    await employeeApi.create(createData)
    $msg.success('员工创建成功')
    router.push('/employee')
  } catch {
    $msg.error('创建失败，请稍后重试')
  } finally {
    saving.value = false
  }
}

async function handleSubmit() {
  if (saving.value) return

  saving.value = true
  try {
    // 转换字段名以匹配后端 DTO
    const data = {
      name: form.name,
      phone: form.phone,
      id_card: form.id_number,
      position: form.position,
      position_id: form.position_id,
      department_id: form.department_id,
      hire_date: form.entry_date,
      gender: form.gender,
      email: form.email,
      salary: form.salary,
      probation_salary: form.probation_salary,
      bank_account: form.bank_card,
      bank_name: form.bank_name,
      address: form.address,
      remark: form.remark,
      emergency_contact: form.emergency_contact,
      emergency_phone: form.emergency_phone,
    }
    if (isEdit.value) {
      await employeeApi.update(Number(route.params.id), data)
      $msg.success('保存成功')
    } else {
      await employeeApi.create(data)
      $msg.success('创建成功')
    }
    router.push('/employee')
  } catch (e) {
    console.error('保存失败:', e)
    $msg.error(isEdit.value ? '保存失败' : '创建失败')
  } finally {
    saving.value = false
  }
}

async function loadEmployee() {
  if (!isEdit.value) return
  try {
    const emp = await employeeApi.get(Number(route.params.id)) as any
    Object.assign(form, {
      name: emp.name || '',
      phone: emp.phone || '',
      id_number: emp.id_card || emp.id_number || '',  // 后端返回 id_card，前端用 id_number
      position: emp.position || '',
      position_id: emp.position_id ?? null,
      department_id: emp.department_id ?? null,
      gender: emp.gender || '',
      email: emp.email || '',
      entry_date: emp.hire_date ? emp.hire_date.split('T')[0] : '',  // 后端返回 hire_date，前端用 entry_date
      salary: emp.salary,
      probation_salary: emp.probation_salary,
      bank_card: emp.bank_account || '',
      bank_name: emp.bank_name || '',
      address: emp.address || '',
      remark: emp.remark || '',
      emergency_contact: emp.emergency_contact || '',
      emergency_phone: emp.emergency_phone || '',
    })

    // 加载薪资数据
    const now = new Date()
    const currentMonth = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
    try {
      const salaryItems = await salaryApi.employeeItems(Number(route.params.id), currentMonth) as any[]
      if (salaryItems && salaryItems.length > 0) {
        // 查找基本工资项
        const baseSalary = salaryItems.find((item) => item.name === '基本工资' || item.name === '岗位工资')
        if (baseSalary) {
          form.salary = baseSalary.amount
        }
        // 查找试用期工资项
        const probationSalary = salaryItems.find((item) => item.name === '试用期工资')
        if (probationSalary) {
          form.probation_salary = probationSalary.amount
        }
      }
    } catch {
      // 薪资加载失败不影响员工信息显示
    }
  } catch {
    $msg.error('加载失败')
  }
}

async function loadPositionOptions(deptId?: number | null) {
  try {
    const data = await positionApi.getSelectOptions(deptId ?? undefined)
    deptPositions.value = data.dept_positions ?? []
    commonPositions.value = data.common_positions ?? []
  } catch {
    deptPositions.value = []
    commonPositions.value = []
  }
}

async function loadDepartments() {
  try {
    const data = await departmentApi.list()
    flatDepartments.value = Array.isArray(data) ? data : []
  } catch {
    flatDepartments.value = []
  }
}

onMounted(() => {
  loadEmployee()
  loadDepartments()
  loadPositionOptions(undefined)
})
</script>

<style scoped lang="scss">
// ============================================================
// 变量定义
// ============================================================
$success: #10B981;
$warning: #F59E0B;
$error: #EF4444;
$bg-page: #FAFBFC;
$bg-surface: #FFFFFF;
$text-primary: #1F2937;
$text-secondary: #6B7280;
$text-muted: #9CA3AF;
$border-color: #E5E7EB;
$shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
$shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
$shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
$radius-sm: 8px;
$radius-md: 12px;
$radius-lg: 16px;
$radius-xl: 24px;

// ============================================================
// 页面布局
// ============================================================
.employee-create {
  padding-bottom: 80px;
  background: $bg-page;
  min-height: 100vh;
}

// ============================================================
// 页面头部
// ============================================================
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  padding: 16px 24px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  position: sticky;
  top: 0;
  z-index: 10;
}

.back-btn {
  width: 40px;
  height: 40px;
  border-radius: $radius-md;
  background: $bg-page;
  border: 1px solid $border-color;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
  color: $text-secondary;

  &:hover {
    background: rgba(var(--primary), 0.08);
    border-color: var(--primary-light);
    color: var(--primary);
  }
}

.header-title-group {
  text-align: center;

  .header-title {
    font-size: 18px;
    font-weight: 700;
    color: $text-primary;
    margin: 0;
  }

  .header-subtitle {
    font-size: 13px;
    color: $text-muted;
    margin: 4px 0 0;
  }
}

// ============================================================
// 表单容器
// ============================================================
.form-container {
  padding: 24px;
  max-width: 800px;
  margin: 0 auto;
}

// ============================================================
// 玻璃态卡片
// ============================================================
.glass-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.8);
  border-radius: $radius-xl;
  box-shadow: $shadow-md;
  padding: 24px;
  margin-bottom: 20px;
  transition: box-shadow 0.2s ease;

  &:hover {
    box-shadow: $shadow-lg;
  }
}

// ============================================================
// 表单区块
// ============================================================
.form-section {
  .section-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 20px;
    padding-bottom: 16px;
    border-bottom: 1px solid $border-color;
  }
}

.section-icon {
  width: 44px;
  height: 44px;
  border-radius: $radius-md;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;

  &--user {
    background: linear-gradient(135deg, #EDE9FE 0%, #DDD6FE 100%);
    color: var(--primary);
  }

  &--salary {
    background: linear-gradient(135deg, #D1FAE5 0%, #A7F3D0 100%);
    color: $success;
  }

  &--emergency {
    background: linear-gradient(135deg, #FEE2E2 0%, #FECACA 100%);
    color: $error;
  }

  &--contact {
    background: linear-gradient(135deg, #DBEAFE 0%, #BFDBFE 100%);
    color: #2563EB;
  }
}

.section-title-group {
  flex: 1;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: $text-primary;
  margin: 0;
}

.section-desc {
  font-size: 13px;
  color: $text-muted;
  margin: 2px 0 0;
}

// ============================================================
// 表单网格
// ============================================================
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  align-items: flex-start;
}

.form-item {
  &--full {
    grid-column: 1 / -1;
  }
}

// ============================================================
// Element Plus 覆盖
// ============================================================
:deep(.el-form-item) {
  margin-bottom: 0;
}

:deep(.el-input__wrapper),
:deep(.el-textarea__inner) {
  border-radius: 10px !important;
  border-color: #E5E7EB;
  box-shadow: none !important;
  transition: all 0.2s ease;

  &:hover {
    border-color: var(--primary-light);
  }

  &:focus-within {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(var(--primary), 0.1) !important;
  }
}

:deep(.el-input-number) {
  width: 100%;

  .el-input__inner {
    text-align: left;
  }
}

:deep(.el-date-editor) {
  width: 100% !important;
}

// ============================================================
// 提交按钮
// ============================================================
.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
}

.submit-btn {
  flex: 1;
  height: 52px;
  border-radius: $radius-lg;
  font-size: 16px;
  font-weight: 600;
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
  border: none;
  box-shadow: 0 4px 14px rgba(var(--primary), 0.4);
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 20px rgba(var(--primary), 0.5);
  }

  &:active {
    transform: translateY(0);
  }

  .el-icon {
    font-size: 18px;
  }
}

.cancel-btn {
  flex: 1;
  height: 52px;
  border-radius: $radius-lg;
  font-size: 16px;
  background: $bg-surface;
  border: 1px solid $border-color;
  color: $text-secondary;
  transition: all 0.2s ease;

  &:hover {
    border-color: var(--primary-light);
    color: var(--primary);
    background: rgba(var(--primary), 0.04);
  }
}

// ============================================================
// 确认页样式
// ============================================================
.confirm-summary {
  p {
    font-size: 14px;
    color: $text-secondary;
    margin-bottom: 16px;
  }
}

.confirm-details {
  background: $bg-page;
  border-radius: $radius-md;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-row {
  font-size: 14px;
  color: $text-primary;

  span {
    color: $text-secondary;
  }
}

.post-create-actions {
  p {
    font-size: 14px;
    color: $text-secondary;
    margin-bottom: 16px;
  }
}

.channel-select {
  margin-top: 16px;
}

// ============================================================
// 响应式
// ============================================================
@media (max-width: 768px) {
  .form-container {
    padding: 16px;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .form-item--full {
    grid-column: 1;
  }

  .form-actions {
    flex-direction: column-reverse;
  }
}
</style>
