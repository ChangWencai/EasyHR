<template>
  <div class="invite-fill-page">
    <!-- 错误状态 -->
    <div v-if="errorState" class="error-state">
      <el-icon :size="48" color="#909399"><CircleCloseFilled /></el-icon>
      <p class="error-message">{{ errorMessage }}</p>
    </div>

    <!-- 成功状态 -->
    <div v-else-if="submitted" class="success-state">
      <el-icon :size="48" color="#67c23a"><CircleCheckFilled /></el-icon>
      <p class="success-message">信息已提交，感谢您的配合</p>
    </div>

    <!-- 表单状态 -->
    <div v-else-if="inviteDetail" class="form-container">
      <h1 class="page-title">协办任务</h1>
      <div class="todo-title-box">
        <el-tag type="info">{{ inviteDetail.title }}</el-tag>
      </div>

      <el-form ref="formRef" :model="form" label-position="top" class="fill-form">
        <!-- 协办填写表单（通用字段） -->
        <el-form-item label="姓名">
          <el-input v-model="form.name" placeholder="请输入您的姓名" />
        </el-form-item>
        <el-form-item label="手机号码">
          <el-input v-model="form.phone" placeholder="请输入手机号码" maxlength="11" />
        </el-form-item>
        <el-form-item label="备注说明">
          <el-input v-model="form.remark" type="textarea" placeholder="如有补充信息请在此填写" :rows="3" />
        </el-form-item>

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
import { ElMessage } from 'element-plus'
import { CircleCloseFilled, CircleCheckFilled, Loading } from '@element-plus/icons-vue'
import { verifyInviteToken, type VerifyResult } from '@/api/todo'
import request from '@/api/request'

const route = useRoute()
const token = route.query.token as string

const inviteDetail = ref<VerifyResult | null>(null)
const errorState = ref(false)
const errorMessage = ref('')
const submitted = ref(false)
const submitting = ref(false)

const form = ref({
  name: '',
  phone: '',
  remark: '',
})

async function loadDetail() {
  try {
    inviteDetail.value = await verifyInviteToken(token)
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
  if (!inviteDetail.value?.todo_id) return
  submitting.value = true
  try {
    await request.post(`/todos/invite/${token}/submit`, {
      name: form.value.name,
      phone: form.value.phone,
      remark: form.value.remark,
    })
    submitted.value = true
  } catch {
    ElMessage.error('提交失败，请稍后重试')
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
.invite-fill-page {
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

.todo-title-box {
  text-align: center;
  margin-bottom: 24px;
}

.fill-form {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
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
