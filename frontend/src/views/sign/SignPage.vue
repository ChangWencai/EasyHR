<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import SmsVerifyInput from '@/components/contract/SmsVerifyInput.vue'
import { contractApi } from '@/api/contract'
import { useMessage } from '@/composables/useMessage'

const $msg = useMessage()
const route = useRoute()

// 从 URL 解析 contract_id
const contractId = computed(() => Number(route.params.contractId))

// Flow: input_phone -> input_code -> confirm_sign -> success
const flow = ref<'phone' | 'code' | 'confirm' | 'success'>('phone')
const phone = ref('')
const code = ref('')
const countdown = ref(0)
const signToken = ref('')
const confirmData = ref<{
  employee_name: string
  contract_type: string
  start_date: string
  end_date?: string
  org_name: string
} | null>(null)
const signedPdfUrl = ref('')

const loading = ref(false)

const typeLabelMap: Record<string, string> = {
  fixed_term: '劳动合同（固定期限）',
  indefinite: '兼职合同',
  intern: '实习协议',
}

// 发送验证码
async function handleSendCode() {
  if (!phone.value || phone.value.length !== 11) {
    $msg.warning('请输入正确的手机号')
    return
  }
  loading.value = true
  try {
    await contractApi.sendSignCode({ contract_id: contractId.value, phone: phone.value })
    $msg.success('验证码已发送')
    flow.value = 'code'
    startCountdown()
  } catch {
    $msg.error('发送失败，请重试')
  } finally {
    loading.value = false
  }
}

// 校验验证码
async function handleVerifyCode() {
  if (code.value.length !== 6) {
    $msg.warning('请输入6位验证码')
    return
  }
  loading.value = true
  try {
    const res = await contractApi.verifySignCode({
      contract_id: contractId.value,
      phone: phone.value,
      code: code.value,
    })
    signToken.value = res.sign_token
    confirmData.value = res
    flow.value = 'confirm'
  } catch (err: unknown) {
    const msg = (err as Error).message || ''
    if (msg.includes('过期')) {
      $msg.error('验证码已过期，请重新获取')
    } else {
      $msg.error('验证码错误，请重新输入')
    }
  } finally {
    loading.value = false
  }
}

// 确认签署
async function handleConfirmSign() {
  loading.value = true
  try {
    const res = await contractApi.confirmSign({
      contract_id: contractId.value,
      sign_token: signToken.value,
    })
    signedPdfUrl.value = res.signed_pdf_url
    flow.value = 'success'
    $msg.success('签署成功')
  } catch {
    $msg.error('签署失败，请重试')
  } finally {
    loading.value = false
  }
}

// 倒计时
let timer: ReturnType<typeof setInterval> | null = null
function startCountdown() {
  countdown.value = 60
  if (timer) clearInterval(timer)
  timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      if (timer) clearInterval(timer)
    }
  }, 1000)
}
</script>

<template>
  <div class="sign-page">
    <!-- Header -->
    <div class="sign-header">
      <span class="logo-text">易人事</span>
    </div>

    <!-- Content -->
    <div class="sign-content">
      <!-- Step dots -->
      <div class="step-dots">
        <div class="dot" :class="{ active: flow !== 'phone', completed: flow !== 'phone' && flow !== 'success' }" />
        <div class="dot" :class="{ active: flow === 'code' || flow === 'confirm', completed: flow === 'confirm' || flow === 'success' }" />
        <div class="dot" :class="{ active: flow === 'confirm' || flow === 'success', completed: flow === 'success' }" />
      </div>

      <!-- Phone input -->
      <div v-if="flow === 'phone'" class="sign-card">
        <h2 class="card-title">验证手机号</h2>
        <p class="card-desc">请输入您的手机号以接收签署验证码</p>
        <el-input
          v-model="phone"
          placeholder="请输入手机号"
          size="large"
          maxlength="11"
          style="margin-bottom: 16px;"
        />
        <el-button
          type="primary"
          size="large"
          style="width: 100%; border-radius: 12px; height: 52px;"
          :loading="loading"
          @click="handleSendCode"
        >
          获取验证码
        </el-button>
      </div>

      <!-- Code input -->
      <div v-else-if="flow === 'code'" class="sign-card">
        <h2 class="card-title">输入验证码</h2>
        <p class="card-desc">已发送至 {{ phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2') }}</p>
        <div style="margin: 24px 0;">
          <SmsVerifyInput
            v-model="code"
            :countdown="countdown"
            @send-code="handleSendCode"
          />
        </div>
        <el-button
          type="primary"
          size="large"
          style="width: 100%; border-radius: 12px; height: 52px;"
          :disabled="code.length !== 6"
          :loading="loading"
          @click="handleVerifyCode"
        >
          下一步
        </el-button>
      </div>

      <!-- Confirm sign -->
      <div v-else-if="flow === 'confirm'" class="sign-card">
        <h2 class="card-title">确认签署合同</h2>
        <div class="confirm-info" v-if="confirmData">
          <div class="info-row">
            <span class="info-label">甲方</span>
            <span class="info-value">{{ confirmData.org_name }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">乙方</span>
            <span class="info-value">{{ confirmData.employee_name }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">类型</span>
            <span class="info-value">{{ typeLabelMap[confirmData.contract_type] || confirmData.contract_type }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">期限</span>
            <span class="info-value">{{ confirmData.start_date }} ~ {{ confirmData.end_date || '无固定' }}</span>
          </div>
        </div>
        <el-button
          type="danger"
          size="large"
          style="width: 100%; border-radius: 12px; height: 52px; margin-top: 24px;"
          :loading="loading"
          @click="handleConfirmSign"
        >
          确认签署
        </el-button>
      </div>

      <!-- Success -->
      <div v-else-if="flow === 'success'" class="sign-card success-card">
        <div class="success-icon">
          <svg width="64" height="64" viewBox="0 0 64 64" fill="none">
            <circle cx="32" cy="32" r="32" fill="#10B981" fill-opacity="0.1"/>
            <path d="M20 32l8 8 16-16" stroke="#10B981" stroke-width="3" stroke-linecap="round"/>
          </svg>
        </div>
        <h2 class="card-title" style="color: var(--success);">签署成功</h2>
        <p class="card-desc">合同已签署完成，您可以查看或下载合同PDF</p>
        <div class="success-actions">
          <el-button
            v-if="signedPdfUrl"
            type="primary"
            size="large"
            style="width: 100%; border-radius: 12px; height: 52px;"
            @click="window.open(signedPdfUrl, '_blank')"
          >
            查看合同
          </el-button>
          <!-- FLAG-1: 返回首页按钮目的待定，暂时隐藏 -->
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.sign-page {
  min-height: 100vh;
  background: var(--bg-page, #FAFBFC);
}

.sign-header {
  height: 48px;
  background: var(--primary, #7C3AED);
  display: flex;
  align-items: center;
  justify-content: center;

  .logo-text {
    color: white;
    font-size: 16px;
    font-weight: 600;
  }
}

.sign-content {
  max-width: 480px;
  margin: 0 auto;
  padding: 24px 16px;
}

.step-dots {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-bottom: 32px;

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--border, #DCDFE6);
    transition: all 0.3s;

    &.active {
      background: var(--primary, #7C3AED);
      width: 24px;
      border-radius: 4px;
    }

    &.completed {
      background: var(--success, #10B981);
    }
  }
}

.sign-card {
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.card-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 8px;
  text-align: center;
}

.card-desc {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 0 0 20px;
  text-align: center;
}

.confirm-info {
  background: var(--bg-page, #FAFBFC);
  border-radius: 12px;
  padding: 16px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  font-size: 14px;

  .info-label {
    color: var(--text-secondary);
  }
  .info-value {
    color: var(--text-primary);
    font-weight: 500;
  }
}

.success-card {
  text-align: center;
  .success-icon {
    margin-bottom: 16px;
  }
}

.success-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 16px;
}
</style>
