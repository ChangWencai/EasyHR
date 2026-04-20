<template>
  <div class="login-layout">
    <!-- 左侧品牌区 - 渐变背景 + 装饰 -->
    <aside class="login-brand">
      <div class="brand-bg">
        <div class="brand-blob brand-blob--1"></div>
        <div class="brand-blob brand-blob--2"></div>
        <div class="brand-blob brand-blob--3"></div>
      </div>

      <div class="brand-content">
        <!-- Logo -->
        <div class="brand-header">
          <div class="brand-logo">
            <div class="logo-icon">
              <Management />
            </div>
            <span class="brand-name">易人事</span>
          </div>
          <p class="brand-slogan">轻量一站式人事管理平台</p>
        </div>

        <div class="brand-divider"></div>

        <!-- 功能特性 -->
        <ul class="feature-list">
          <li v-for="(feature, index) in features" :key="index" class="feature-item" :style="{ animationDelay: `${index * 100}ms` }">
            <div class="feature-icon">
              <component :is="feature.icon" />
            </div>
            <div class="feature-text">
              <strong>{{ feature.title }}</strong>
              <span>{{ feature.desc }}</span>
            </div>
          </li>
        </ul>

        <!-- 数据展示 -->
        <div class="brand-stats">
          <div class="stat-item">
            <span class="stat-value">10,000+</span>
            <span class="stat-label">企业用户</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-item">
            <span class="stat-value">50,000+</span>
            <span class="stat-label">员工管理</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-item">
            <span class="stat-value">99.9%</span>
            <span class="stat-label">服务可用性</span>
          </div>
        </div>
      </div>

      <!-- 底部版权 -->
      <p class="brand-copyright">© 2025 易人事 · 专为小微企业打造</p>
    </aside>

    <!-- 右侧表单区 -->
    <main class="login-form-panel">
      <div class="login-card">
        <div class="form-brand-header">
          <h1 class="form-brand-title">欢迎回来</h1>
          <p class="form-brand-subtitle">登录到您的管理后台</p>
        </div>

        <!-- 登录方式切换 -->
        <div class="login-tabs">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="tab-btn"
            :class="{ active: activeTab === tab.key }"
            @click="activeTab = tab.key"
          >
            {{ tab.label }}
          </button>
        </div>

        <!-- 手机验证码登录 -->
        <form v-show="activeTab === 'sms'" class="login-form" @submit.prevent="handleSmsLogin">
          <div class="form-group">
            <label class="form-label">手机号</label>
            <div class="input-wrapper">
              <el-icon class="input-icon"><User /></el-icon>
              <input
                v-model="smsForm.phone"
                type="tel"
                placeholder="请输入手机号"
                maxlength="11"
                class="form-input"
              />
            </div>
          </div>
          <div class="form-group">
            <label class="form-label">验证码</label>
            <div class="input-wrapper code-input">
              <el-icon class="input-icon"><Lock /></el-icon>
              <input
                v-model="smsForm.code"
                type="text"
                placeholder="请输入验证码"
                maxlength="6"
                class="form-input"
              />
              <button
                type="button"
                class="code-btn"
                :disabled="countdown > 0"
                @click="handleSendCode"
              >
                {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
              </button>
            </div>
          </div>
          <button type="submit" class="submit-btn" :class="{ loading: submitting }">
            <span v-if="!submitting">登录</span>
            <span v-else class="loading-spinner"></span>
          </button>
        </form>

        <!-- 密码登录 -->
        <form v-show="activeTab === 'password'" class="login-form" @submit.prevent="handlePasswordLogin">
          <div class="form-group">
            <label class="form-label">手机号</label>
            <div class="input-wrapper">
              <el-icon class="input-icon"><User /></el-icon>
              <input
                v-model="passwordForm.phone"
                type="tel"
                placeholder="请输入手机号"
                maxlength="11"
                class="form-input"
              />
            </div>
          </div>
          <div class="form-group">
            <label class="form-label">密码</label>
            <div class="input-wrapper">
              <el-icon class="input-icon"><Lock /></el-icon>
              <input
                v-model="passwordForm.password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="请输入密码"
                class="form-input"
              />
              <button type="button" class="toggle-password" @click="showPassword = !showPassword">
                <el-icon>
                  <View v-if="!showPassword" />
                  <Hide v-else />
                </el-icon>
              </button>
            </div>
          </div>
          <button type="submit" class="submit-btn" :class="{ loading: submitting }">
            <span v-if="!submitting">登录</span>
            <span v-else class="loading-spinner"></span>
          </button>
        </form>

        <!-- 注册 -->
        <form v-show="activeTab === 'register'" class="login-form" @submit.prevent="handleRegister">
          <div class="form-group">
            <label class="form-label">手机号</label>
            <div class="input-wrapper">
              <el-icon class="input-icon"><User /></el-icon>
              <input
                v-model="registerForm.phone"
                type="tel"
                placeholder="请输入手机号"
                maxlength="11"
                class="form-input"
              />
            </div>
          </div>
          <div class="form-group">
            <label class="form-label">验证码</label>
            <div class="input-wrapper code-input">
              <el-icon class="input-icon"><Lock /></el-icon>
              <input
                v-model="registerForm.code"
                type="text"
                placeholder="请输入验证码"
                maxlength="6"
                class="form-input"
              />
              <button
                type="button"
                class="code-btn"
                :disabled="countdown > 0"
                @click="handleSendCode"
              >
                {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
              </button>
            </div>
          </div>
          <button type="submit" class="submit-btn register-btn" :class="{ loading: submitting }">
            <span v-if="!submitting">注册</span>
            <span v-else class="loading-spinner"></span>
          </button>
        </form>

        <p class="copyright">登录即表示同意 <a href="#">《用户协议》</a> 和 <a href="#">《隐私政策》</a></p>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  User,
  Lock,
  Management,
  Briefcase,
  Money,
  View,
  Hide,
  Calendar,
  DataLine,
  CircleCheck,
} from '@element-plus/icons-vue'
import request, { ERR_NEED_SMS_LOGIN } from '@/api/request'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const activeTab = ref('sms')
const showPassword = ref(false)
const submitting = ref(false)

const tabs = [
  { key: 'sms', label: '验证码登录' },
  { key: 'password', label: '密码登录' },
  { key: 'register', label: '注册' },
]

const features = [
  { icon: Briefcase, title: '入职管理', desc: '员工信息快速录入' },
  { icon: Money, title: '薪资核算', desc: '一键计算工资个税' },
  { icon: Calendar, title: '社保公积金', desc: '多城市政策自动更新' },
  { icon: DataLine, title: '财务记账', desc: '工资支出自动记账' },
  { icon: CircleCheck, title: '员工工资条', desc: '微信一键发送工资条' },
]

// 短信登录表单
const smsForm = ref({ phone: '', code: '' })
const countdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

// 密码登录表单
const passwordForm = ref({ phone: '', password: '' })

// 注册表单
const registerForm = ref({ phone: '', code: '' })

// 发送验证码
async function handleSendCode() {
  const phone = activeTab.value === 'register' ? registerForm.value.phone : smsForm.value.phone
  if (!phone || phone.length !== 11) {
    ElMessage.error('请输入正确的手机号')
    return
  }
  try {
    await request.post('/auth/send-code', { phone })
    ElMessage.success('验证码已发送')
    startCountdown()
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '发送失败')
  }
}

// 60秒倒计时
function startCountdown() {
  countdown.value = 60
  countdownTimer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      countdown.value = 0
      if (countdownTimer) clearInterval(countdownTimer)
    }
  }, 1000)
}

// 注册
async function handleRegister() {
  if (!registerForm.value.phone || !registerForm.value.code) {
    ElMessage.error('请填写手机号和验证码')
    return
  }
  try {
    const resp = await request.post('/auth/register', {
      phone: registerForm.value.phone,
      code: registerForm.value.code,
    })
    handleLoginSuccess(resp.data)
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '注册失败')
  }
}

// 短信登录
async function handleSmsLogin() {
  if (!smsForm.value.phone || !smsForm.value.code) {
    ElMessage.error('请填写手机号和验证码')
    return
  }
  submitting.value = true
  try {
    const resp = await request.post('/auth/login', {
      phone: smsForm.value.phone,
      code: smsForm.value.code,
    })
    handleLoginSuccess(resp.data)
  } catch (err: any) {
    if (err.response?.status === 403) {
      ElMessage.error('您的账号为员工账号，请使用员工端微信小程序登录')
      return
    }
    ElMessage.error(err.response?.data?.message || '登录失败')
  } finally {
    submitting.value = false
  }
}

// 密码登录
async function handlePasswordLogin() {
  if (!passwordForm.value.phone || !passwordForm.value.password) {
    ElMessage.error('请填写手机号和密码')
    return
  }
  submitting.value = true
  try {
    const resp = await request.post('/auth/login/password', {
      phone: passwordForm.value.phone,
      password: passwordForm.value.password,
    })
    handleLoginSuccess(resp.data)
  } catch (err: any) {
    const bizCode = err.response?.data?.code
    if (bizCode === ERR_NEED_SMS_LOGIN) {
      activeTab.value = 'sms'
      ElMessage.error(err.response?.data?.message || '该账号未设置密码，请使用手机验证码登录')
      return
    }
    if (err.response?.status === 403) {
      ElMessage.error('您的账号为员工账号，请使用员工端微信小程序登录')
      return
    }
    ElMessage.error(err.response?.data?.message || '登录失败')
  } finally {
    submitting.value = false
  }
}

// 登录成功后处理
function handleLoginSuccess(resp: any) {
  authStore.setToken(resp.access_token)
  if (resp.onboarding_required === true) {
    router.push('/onboarding/org-setup')
  } else {
    router.push('/home')
  }
}

// 微信回调：检查 URL 中是否有 code 参数
onMounted(() => {
  const params = new URLSearchParams(window.location.search)
  const code = params.get('code')
  if (code) {
    // 微信回调登录（Phase 1.5 功能，占位）
  }
})
</script>

<style scoped lang="scss">
// ============================================================
// 变量定义
// ============================================================
$success: #10B981;
$bg-page: #FAFBFC;
$bg-surface: #FFFFFF;
$text-primary: #1F2937;
$text-secondary: #6B7280;
$text-muted: #9CA3AF;
$border-color: #E5E7EB;
$success: #10B981;
$shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
$shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
$shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
$radius-sm: 8px;
$radius-md: 12px;
$radius-lg: 16px;
$radius-xl: 24px;

// ============================================================
// 布局
// ============================================================
.login-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  min-height: 100vh;
}

// ============================================================
// 左侧品牌区
// ============================================================
.login-brand {
  background: linear-gradient(135deg, #667EEA 0%, #764BA2 50%, #F093FB 100%);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 48px 64px;
  position: relative;
  overflow: hidden;
}

.brand-bg {
  position: absolute;
  inset: 0;
  overflow: hidden;
}

.brand-blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.4;
  animation: float 20s ease-in-out infinite;

  &--1 {
    width: 400px;
    height: 400px;
    background: rgba(255, 255, 255, 0.3);
    top: -100px;
    right: -100px;
    animation-delay: 0s;
  }

  &--2 {
    width: 300px;
    height: 300px;
    background: rgba(255, 255, 255, 0.2);
    bottom: -50px;
    left: -50px;
    animation-delay: -5s;
  }

  &--3 {
    width: 200px;
    height: 200px;
    background: rgba(255, 255, 255, 0.15);
    top: 50%;
    left: 30%;
    animation-delay: -10s;
  }
}

@keyframes float {
  0%, 100% { transform: translate(0, 0) rotate(0deg); }
  25% { transform: translate(20px, -20px) rotate(5deg); }
  50% { transform: translate(0, 20px) rotate(0deg); }
  75% { transform: translate(-20px, -10px) rotate(-5deg); }
}

.brand-content {
  position: relative;
  z-index: 1;
  animation: slideIn 0.6s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(-20px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.brand-header {
  margin-bottom: 40px;
}

.brand-logo {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}

.logo-icon {
  width: 48px;
  height: 48px;
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px);
  border-radius: $radius-md;
  display: flex;
  align-items: center;
  justify-content: center;

  :deep(.el-icon) {
    font-size: 28px;
    color: #fff;
  }
}

.brand-name {
  font-size: 32px;
  font-weight: 700;
  color: #fff;
  letter-spacing: 4px;
}

.brand-slogan {
  font-size: 16px;
  color: rgba(255, 255, 255, 0.9);
  margin: 0;
}

.brand-divider {
  width: 60px;
  height: 4px;
  background: rgba(255, 255, 255, 0.5);
  border-radius: 2px;
  margin-bottom: 40px;
}

.feature-list {
  list-style: none;
  padding: 0;
  margin: 0 0 48px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 16px;
  animation: fadeInUp 0.5s ease-out backwards;

  &:hover .feature-icon {
    transform: scale(1.1);
    background: rgba(255, 255, 255, 0.3);
  }
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.feature-icon {
  width: 48px;
  height: 48px;
  border-radius: $radius-md;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.2s ease;

  :deep(.el-icon) {
    font-size: 22px;
    color: rgba(255, 255, 255, 0.95);
  }
}

.feature-text {
  display: flex;
  flex-direction: column;
  gap: 2px;

  strong {
    font-size: 16px;
    font-weight: 600;
    color: #fff;
  }

  span {
    font-size: 14px;
    color: rgba(255, 255, 255, 0.7);
  }
}

.brand-stats {
  display: flex;
  align-items: center;
  gap: 32px;
  padding: 24px;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-radius: $radius-lg;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #fff;
}

.stat-label {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.7);
}

.stat-divider {
  width: 1px;
  height: 40px;
  background: rgba(255, 255, 255, 0.3);
}

.brand-copyright {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.6);
  text-align: center;
  position: relative;
  z-index: 1;
  margin: 0;
}

// ============================================================
// 右侧表单区
// ============================================================
.login-form-panel {
  background: $bg-page;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px;
}

.login-card {
  width: 100%;
  max-width: 420px;
  animation: slideUp 0.5s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(24px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.form-brand-header {
  text-align: center;
  margin-bottom: 32px;
}

.form-brand-title {
  font-size: 28px;
  font-weight: 700;
  color: $text-primary;
  margin: 0 0 8px;
  letter-spacing: -0.5px;
}

.form-brand-subtitle {
  font-size: 15px;
  color: $text-secondary;
  margin: 0;
}

// ============================================================
// 登录标签切换
// ============================================================
.login-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 32px;
  padding: 4px;
  background: $bg-surface;
  border-radius: $radius-lg;
  box-shadow: $shadow-sm;
}

.tab-btn {
  flex: 1;
  padding: 12px 16px;
  font-size: 14px;
  font-weight: 500;
  color: $text-secondary;
  background: transparent;
  border: none;
  border-radius: $radius-md;
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    color: $text-primary;
  }

  &.active {
    background: var(--primary);
    color: #fff;
    box-shadow: 0 2px 8px rgba(var(--primary), 0.3);
  }
}

// ============================================================
// 表单
// ============================================================
.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  font-size: 14px;
  font-weight: 500;
  color: $text-primary;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;

  &.code-input {
    .form-input {
      padding-right: 120px;
    }
  }
}

.input-icon {
  position: absolute;
  left: 16px;
  color: $text-muted;
  font-size: 18px;
  pointer-events: none;
}

.form-input {
  width: 100%;
  padding: 14px 16px 14px 48px;
  font-size: 15px;
  color: $text-primary;
  background: $bg-surface;
  border: 1px solid $border-color;
  border-radius: $radius-md;
  transition: all 0.2s ease;
  outline: none;

  &::placeholder {
    color: $text-muted;
  }

  &:focus {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(var(--primary), 0.1);
  }
}

.code-btn {
  position: absolute;
  right: 8px;
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 500;
  color: var(--primary);
  background: rgba(var(--primary), 0.08);
  border: none;
  border-radius: $radius-sm;
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover:not(:disabled) {
    background: rgba(var(--primary), 0.15);
  }

  &:disabled {
    color: $text-muted;
    cursor: not-allowed;
  }
}

.toggle-password {
  position: absolute;
  right: 12px;
  padding: 4px;
  color: $text-muted;
  background: none;
  border: none;
  cursor: pointer;

  &:hover {
    color: $text-secondary;
  }

  .el-icon {
    font-size: 18px;
  }
}

.submit-btn {
  width: 100%;
  padding: 16px;
  font-size: 16px;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
  border: none;
  border-radius: $radius-lg;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 4px 14px rgba(var(--primary), 0.4);

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 20px rgba(var(--primary), 0.5);
  }

  &:active {
    transform: translateY(0);
  }

  &.register-btn {
    background: $bg-surface;
    color: var(--primary);
    box-shadow: none;
    border: 2px solid var(--primary);

    &:hover {
      background: rgba(var(--primary), 0.05);
      box-shadow: 0 4px 14px rgba(var(--primary), 0.2);
    }
  }

  &.loading {
    pointer-events: none;
    opacity: 0.8;
  }
}

.loading-spinner {
  display: inline-block;
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: #fff;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.copyright {
  text-align: center;
  font-size: 12px;
  color: $text-muted;
  margin-top: 24px;

  a {
    color: var(--primary);
    text-decoration: none;

    &:hover {
      text-decoration: underline;
    }
  }
}

// ============================================================
// 移动端响应式
// ============================================================
@media (max-width: 768px) {
  .login-layout {
    grid-template-columns: 1fr;
  }

  .login-brand {
    display: none;
  }

  .login-form-panel {
    padding: 32px 24px;
    align-items: flex-start;
    padding-top: 64px;
  }
}
</style>
