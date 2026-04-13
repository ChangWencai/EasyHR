<template>
  <div class="login-page">
    <!-- 品牌区 -->
    <div class="brand-section">
      <div class="brand-logo">
        <svg width="56" height="56" viewBox="0 0 56 56" fill="none" xmlns="http://www.w3.org/2000/svg">
          <rect width="56" height="56" rx="16" fill="#0F766E"/>
          <path d="M14 18h28M14 28h20M14 38h24" stroke="#fff" stroke-width="3.5" stroke-linecap="round"/>
        </svg>
      </div>
      <h1 class="brand-name">易人事</h1>
      <p class="brand-slogan">小微企业的人事管理神器</p>
    </div>

    <!-- 登录表单 -->
    <div class="form-section">
      <div class="phone-input-wrap">
        <span class="phone-prefix">+86</span>
        <input
          v-model="phone"
          type="tel"
          class="phone-input"
          placeholder="请输入手机号"
          maxlength="11"
          inputmode="numeric"
          @keydown.enter="handleSendCode"
        />
        <button v-if="phone" class="clear-btn" @click="phone = ''" aria-label="清除">
          <el-icon :size="14"><Close /></el-icon>
        </button>
      </div>

      <!-- 验证码输入（发送后显示） -->
      <transition name="fade-slide">
        <div v-if="codeSent" class="code-input-wrap">
          <input
            v-model="code"
            type="text"
            class="code-input"
            placeholder="请输入验证码"
            maxlength="6"
            inputmode="numeric"
            @keydown.enter="handleLogin"
          />
          <button class="resend-btn" :disabled="countdown > 0" @click="handleResend">
            {{ countdown > 0 ? `${countdown}s` : '重新获取' }}
          </button>
        </div>
      </transition>

      <!-- 发送验证码 / 登录按钮 -->
      <button
        class="primary-btn"
        :disabled="loading || (codeSent ? !code : !phoneValid)"
        @click="codeSent ? handleLogin() : handleSendCode()"
      >
        <span v-if="loading" class="btn-loading">
          <el-icon class="is-loading" :size="18"><Loading /></el-icon>
        </span>
        <span v-else>{{ codeSent ? '登录' : '获取验证码' }}</span>
      </button>

      <!-- 微信登录 -->
      <div class="divider">
        <span class="divider-line" />
        <span class="divider-text">其他登录方式</span>
        <span class="divider-line" />
      </div>

      <button class="wechat-btn" @click="handleWechatLogin">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
          <path d="M8.5 11a1.5 1.5 0 100-3 1.5 1.5 0 000 3zM15.5 11a1.5 1.5 0 100-3 1.5 1.5 0 000 3z" fill="#07C160"/>
          <path d="M12 2C6.477 2 2 6.145 2 11.243c0 2.937 1.519 5.547 3.893 7.267L5.54 21l3.823-1.65a10.32 10.32 0 004.637.563c5.523 0 10-4.145 10-9.243C22 6.145 17.523 2 12 2z" fill="#07C160"/>
        </svg>
        <span>微信一键登录</span>
      </button>

      <p class="agreement-text">
        登录即表示同意
        <a href="#" @click.prevent>《用户服务协议》</a>
        和
        <a href="#" @click.prevent>《隐私政策》</a>
      </p>

      <p class="register-hint">
        手机号即账号，无须单独注册
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Close, Loading } from '@element-plus/icons-vue'
import request from '@/api/request'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const phone = ref('')
const code = ref('')
const codeSent = ref(false)
const loading = ref(false)
const countdown = ref(0)

let countdownTimer: ReturnType<typeof setInterval> | null = null

const phoneValid = computed(() => /^1[3-9]\d{9}$/.test(phone.value))

async function handleSendCode() {
  if (!phoneValid.value || loading.value) return
  loading.value = true
  try {
    await request.post('/auth/send-code', { phone: phone.value })
    codeSent.value = true
    startCountdown()
    ElMessage.success('验证码已发送')
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || '发送失败，请稍后重试'
    ElMessage.error(msg)
  } finally {
    loading.value = false
  }
}

async function handleLogin() {
  if (!code.value || loading.value) return
  loading.value = true
  try {
    const res = await request.post('/auth/login', {
      phone: phone.value,
      code: code.value,
    })
    const token = res.data?.access_token || res.access_token
    const refreshToken = res.data?.refresh_token || res.refresh_token
    authStore.setToken(token)
    localStorage.setItem('refresh_token', refreshToken)
    ElMessage.success('登录成功')
    router.push('/home')
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || '登录失败，请检查验证码'
    ElMessage.error(msg)
  } finally {
    loading.value = false
  }
}

function handleResend() {
  if (countdown.value > 0) return
  handleSendCode()
}

function startCountdown() {
  countdown.value = 60
  countdownTimer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(countdownTimer!)
      countdownTimer = null
    }
  }, 1000)
}

onUnmounted(() => {
  if (countdownTimer) clearInterval(countdownTimer)
})
</script>

<style scoped lang="scss">
.login-page {
  min-height: 100vh;
  background: #F8FAFC;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 24px;
  box-sizing: border-box;
}

// ===== 品牌区 =====
.brand-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 56px 0 40px;
  gap: 10px;
}

.brand-logo {
  width: 72px;
  height: 72px;
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 8px 24px rgba(15, 118, 110, 0.25);
}

.brand-name {
  font-size: 28px;
  font-weight: 800;
  color: #0F172A;
  margin: 0;
  letter-spacing: -0.5px;
}

.brand-slogan {
  font-size: 14px;
  color: #94A3B8;
  margin: 0;
}

// ===== 表单区 =====
.form-section {
  background: #fff;
  border-radius: 20px;
  padding: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
  width: 100%;
  max-width: 400px;
  box-sizing: border-box;
}

.phone-input-wrap {
  display: flex;
  align-items: center;
  border: 1.5px solid #E2E8F0;
  border-radius: 12px;
  padding: 0 14px;
  height: 52px;
  gap: 10px;
  transition: border-color 0.15s;
  background: #fff;

  &:focus-within {
    border-color: #0F766E;
  }
}

.phone-prefix {
  font-size: 16px;
  font-weight: 600;
  color: #0F172A;
  flex-shrink: 0;
  border-right: 1.5px solid #E2E8F0;
  padding-right: 10px;
}

.phone-input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 16px;
  color: #0F172A;
  background: transparent;
  height: 100%;
  caret-color: #0F766E;

  &::placeholder { color: #CBD5E1; }
}

.clear-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: #94A3B8;
  display: flex;
  align-items: center;
  padding: 4px;
  border-radius: 50%;
  transition: all 0.15s;

  &:hover { background: #F1F5F9; color: #64748B; }
}

.code-input-wrap {
  display: flex;
  align-items: center;
  border: 1.5px solid #E2E8F0;
  border-radius: 12px;
  padding: 0 14px;
  height: 52px;
  gap: 10px;
  margin-top: 12px;
  transition: border-color 0.15s;
  background: #fff;

  &:focus-within { border-color: #0F766E; }
}

.code-input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 16px;
  color: #0F172A;
  background: transparent;
  height: 100%;
  letter-spacing: 4px;
  font-feature-settings: "tnum";
  caret-color: #0F766E;

  &::placeholder { color: #CBD5E1; letter-spacing: 0; }
}

.resend-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 13px;
  color: #0F766E;
  font-weight: 500;
  white-space: nowrap;
  padding: 4px 0;
  transition: color 0.15s;

  &:hover { color: #0D6B62; }
  &:disabled { color: #CBD5E1; cursor: not-allowed; }
}

.primary-btn {
  width: 100%;
  height: 52px;
  background: #0F766E;
  color: #fff;
  border: none;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  margin-top: 16px;
  transition: background 0.15s, transform 0.1s, opacity 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;

  &:hover:not(:disabled) { background: #0D6B62; }
  &:active:not(:disabled) { transform: scale(0.99); }
  &:disabled { opacity: 0.5; cursor: not-allowed; }
}

.btn-loading {
  display: flex;
  align-items: center;
  justify-content: center;
}

// ===== 过渡动画 =====
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.fade-slide-enter-from,
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

// ===== 分隔线 =====
.divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 20px 0 16px;
}

.divider-line {
  flex: 1;
  height: 1px;
  background: #E2E8F0;
}

.divider-text {
  font-size: 12px;
  color: #94A3B8;
  white-space: nowrap;
}

// ===== 微信登录 =====
.wechat-btn {
  width: 100%;
  height: 52px;
  background: #07C160;
  color: #fff;
  border: none;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transition: background 0.15s, transform 0.1s;

  &:hover { background: #06a555; }
  &:active { transform: scale(0.99); }
}

// ===== 协议文字 =====
.agreement-text {
  font-size: 11px;
  color: #94A3B8;
  text-align: center;
  margin: 16px 0 0;
  line-height: 1.6;

  a {
    color: #0F766E;
    text-decoration: none;

    &:hover { text-decoration: underline; }
  }
}

.register-hint {
  font-size: 12px;
  color: #94A3B8;
  text-align: center;
  margin: 10px 0 0;
}
</style>
