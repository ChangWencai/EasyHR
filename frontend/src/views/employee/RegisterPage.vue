<template>
  <div class="register-page">
    <!-- 错误状态 -->
    <div v-if="errorState" class="error-state">
      <el-icon :size="48" color="#909399"><CircleCloseFilled /></el-icon>
      <p class="error-message">{{ errorMessage }}</p>
    </div>

    <!-- 成功状态 -->
    <div v-else-if="submitted" class="success-state">
      <el-icon :size="48" color="#67c23a"><CircleCheckFilled /></el-icon>
      <p class="success-message">信息提交成功，感谢您的配合</p>
    </div>

    <!-- 表单状态 -->
    <div v-else-if="detail" class="form-container">
      <h1 class="page-title">员工信息登记</h1>

      <el-form ref="formRef" :model="form" label-position="top" class="register-form">
        <!-- 基本信息 -->
        <div class="section">
          <h3 class="section-title">基本信息</h3>
          <el-form-item label="姓名">
            <el-input :model-value="detail.name" disabled />
          </el-form-item>
          <el-form-item label="岗位">
            <el-input :model-value="detail.position" disabled />
          </el-form-item>
          <el-form-item label="入职日期">
            <el-input :model-value="detail.hire_date" disabled />
          </el-form-item>
          <el-form-item label="手机号码">
            <el-input v-model="form.phone" placeholder="请输入手机号码" maxlength="11" />
          </el-form-item>
          <el-form-item label="住址">
            <el-input v-model="form.address" placeholder="请输入居住地址" />
          </el-form-item>
        </div>

        <!-- 身份证信息 -->
        <div class="section">
          <h3 class="section-title">身份证信息</h3>
          <el-form-item label="身份证号">
            <el-input v-model="form.id_card" placeholder="请输入18位身份证号" maxlength="18" />
          </el-form-item>
          <el-form-item label="正面照">
            <el-input v-model="form.id_card_front_url" placeholder="身份证正面照URL" />
          </el-form-item>
          <el-form-item label="反面照">
            <el-input v-model="form.id_card_back_url" placeholder="身份证反面照URL" />
          </el-form-item>
        </div>

        <!-- 银行卡信息 -->
        <div class="section">
          <h3 class="section-title">银行卡信息</h3>
          <el-form-item label="银行卡号">
            <el-input v-model="form.bank_account" placeholder="请输入银行卡号" />
          </el-form-item>
          <el-form-item label="开户行">
            <el-input v-model="form.bank_name" placeholder="请输入开户银行名称" />
          </el-form-item>
          <el-form-item label="正面照">
            <el-input v-model="form.bank_card_front_url" placeholder="银行卡正面照URL" />
          </el-form-item>
          <el-form-item label="反面照">
            <el-input v-model="form.bank_card_back_url" placeholder="银行卡反面照URL" />
          </el-form-item>
        </div>

        <!-- 学历信息 -->
        <div class="section">
          <h3 class="section-title">学历信息</h3>
          <el-form-item label="学历证书">
            <el-input v-model="form.education_cert_url" placeholder="学历证书URL" />
          </el-form-item>
        </div>

        <!-- 紧急联系人 -->
        <div class="section">
          <h3 class="section-title">紧急联系人</h3>
          <el-form-item label="姓名">
            <el-input v-model="form.emergency_contact" placeholder="请输入紧急联系人姓名" />
          </el-form-item>
          <el-form-item label="电话">
            <el-input v-model="form.emergency_phone" placeholder="请输入紧急联系人电话" maxlength="11" />
          </el-form-item>
          <el-form-item label="与本人关系">
            <el-input v-model="form.emergency_relation" placeholder="如：父母、配偶、朋友" />
          </el-form-item>
        </div>

        <el-button
          type="primary"
          class="submit-btn"
          :loading="submitting"
          @click="handleSubmit"
        >
          提交信息
        </el-button>
      </el-form>
    </div>

    <!-- 加载状态 -->
    <div v-else class="loading-state">
      <el-icon :size="32" class="is-loading"><Loading /></el-icon>
      <p>正在加载...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { registrationApi, type RegistrationDetail, type SubmitRegistrationData } from '@/api/employee'
import { ElMessage } from 'element-plus'
import { CircleCloseFilled, CircleCheckFilled, Loading } from '@element-plus/icons-vue'

const route = useRoute()
const token = route.params.token as string

const detail = ref<RegistrationDetail | null>(null)
const errorState = ref(false)
const errorMessage = ref('')
const submitted = ref(false)
const submitting = ref(false)

const form = ref<SubmitRegistrationData>({
  phone: '',
  address: '',
  id_card: '',
  id_card_front_url: '',
  id_card_back_url: '',
  bank_account: '',
  bank_name: '',
  bank_card_front_url: '',
  bank_card_back_url: '',
  education_cert_url: '',
  emergency_contact: '',
  emergency_phone: '',
  emergency_relation: '',
})

async function loadDetail() {
  try {
    detail.value = await registrationApi.getDetail(token)
  } catch (err: unknown) {
    errorState.value = true
    const errorObj = err as { response?: { status?: number; data?: { message?: string } } }
    if (errorObj.response?.status === 410) {
      errorMessage.value = '该链接已过期，请联系管理员重新发送'
    } else if (errorObj.response?.status === 404) {
      errorMessage.value = '链接无效，请确认链接是否正确'
    } else {
      errorMessage.value = errorObj.response?.data?.message || '加载失败，请稍后重试'
    }
  }
}

async function handleSubmit() {
  submitting.value = true
  try {
    await registrationApi.submit(token, form.value)
    submitted.value = true
    ElMessage.success('信息提交成功，感谢您的配合')
  } catch (err: unknown) {
    const errorObj = err as { response?: { data?: { message?: string } } }
    ElMessage.error(errorObj.response?.data?.message || '提交失败，请稍后重试')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  if (token) {
    loadDetail()
  } else {
    errorState.value = true
    errorMessage.value = '链接无效，请确认链接是否正确'
  }
})
</script>

<style scoped lang="scss">
.register-page {
  min-height: 100vh;
  background: #f5f7fa;
  padding: 16px;
}

.form-container {
  max-width: 480px;
  margin: 0 auto;
}

.page-title {
  font-size: 22px;
  font-weight: 700;
  text-align: center;
  margin: 16px 0 24px;
  color: #303133;
}

.register-form {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
}

.section {
  margin-bottom: 24px;
}

.section-title {
  font-size: 16px;
  font-weight: 700;
  color: #303133;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #ebeef5;
}

.submit-btn {
  width: 100%;
  height: 44px;
  font-size: 16px;
  margin-top: 8px;
}

.error-state,
.success-state,
.loading-state {
  max-width: 480px;
  margin: 80px auto 0;
  text-align: center;
}

.error-message {
  margin-top: 16px;
  font-size: 16px;
  color: #909399;
}

.success-message {
  margin-top: 16px;
  font-size: 16px;
  color: #67c23a;
}

.loading-state {
  margin-top: 120px;
  color: #909399;
}
</style>
