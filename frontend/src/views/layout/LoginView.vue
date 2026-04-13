<template>
  <div class="login-layout">
    <!-- 左侧品牌区 -->
    <aside class="login-brand">
      <div class="brand-content">
        <!-- Logo + 标语 -->
        <div class="brand-header">
          <div class="brand-logo">
            <Management />
            <span class="brand-name">易人事</span>
          </div>
          <p class="brand-slogan">轻量一站式人事管理平台</p>
        </div>

        <div class="brand-divider"></div>

        <!-- 5个功能特性 -->
        <ul class="feature-list">
          <li class="feature-item">
            <div class="feature-icon"><BriefCase /></div>
            <div class="feature-text">
              <strong>入职管理</strong>
              <span>员工信息快速录入</span>
            </div>
          </li>
          <li class="feature-item">
            <div class="feature-icon"><Money /></div>
            <div class="feature-text">
              <strong>薪资核算</strong>
              <span>一键计算工资个税</span>
            </div>
          </li>
          <li class="feature-item">
            <div class="feature-icon"><Shield /></div>
            <div class="feature-text">
              <strong>社保公积金</strong>
              <span>多城市政策自动更新</span>
            </div>
          </li>
          <li class="feature-item">
            <div class="feature-icon"><Wallet /></div>
            <div class="feature-text">
              <strong>财务记账</strong>
              <span>工资支出自动记账</span>
            </div>
          </li>
          <li class="feature-item">
            <div class="feature-icon"><Document /></div>
            <div class="feature-text">
              <strong>员工工资条</strong>
              <span>微信一键发送工资条</span>
            </div>
          </li>
        </ul>
      </div>

      <!-- 底部版权 -->
      <p class="brand-copyright">© 2025 易人事 · 专为小微企业打造</p>
    </aside>

    <!-- 右侧表单区 -->
    <main class="login-form-panel">
      <div class="login-card">
        <div class="form-brand-header">
          <h1 class="form-brand-title">易人事</h1>
          <p class="form-brand-subtitle">老板管理后台</p>
        </div>

        <el-tabs v-model="activeTab" :animated="false" class="login-tabs">
          <!-- Tab 1: 手机验证码登录 -->
          <el-tab-pane label="手机验证码" name="sms">
            <el-form @submit.prevent="handleSmsLogin">
              <el-form-item>
                <el-input
                  v-model="smsForm.phone"
                  placeholder="请输入手机号"
                  maxlength="11"
                  type="number"
                  :prefix-icon="User"
                />
              </el-form-item>
              <el-form-item>
                <div class="code-row">
                  <el-input
                    v-model="smsForm.code"
                    placeholder="请输入验证码"
                    maxlength="6"
                    style="width: 60%"
                    :prefix-icon="Lock"
                  />
                  <el-button
                    :disabled="countdown > 0"
                    style="width: 38%"
                    @click="handleSendCode"
                  >
                    {{ countdown > 0 ? `已发送(${countdown}s)` : '获取验证码' }}
                  </el-button>
                </div>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" style="width: 100%; height: 44px; font-size: 16px" @click="handleSmsLogin">
                  登录
                </el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>

          <!-- Tab 2: 密码登录 -->
          <el-tab-pane label="密码登录" name="password">
            <el-form @submit.prevent="handlePasswordLogin">
              <el-form-item>
                <el-input
                  v-model="passwordForm.phone"
                  placeholder="请输入手机号"
                  maxlength="11"
                  type="number"
                  :prefix-icon="User"
                />
              </el-form-item>
              <el-form-item>
                <el-input
                  v-model="passwordForm.password"
                  placeholder="请输入密码"
                  show-password
                  type="password"
                  :prefix-icon="Lock"
                  @keyup.enter="handlePasswordLogin"
                />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" style="width: 100%; height: 44px; font-size: 16px" @click="handlePasswordLogin">
                  登录
                </el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>

          <!-- Tab 3: 注册 -->
          <el-tab-pane label="注册" name="register">
            <el-form @submit.prevent="handleRegister">
              <el-form-item>
                <el-input
                  v-model="registerForm.phone"
                  placeholder="请输入手机号"
                  maxlength="11"
                  type="number"
                  :prefix-icon="User"
                />
              </el-form-item>
              <el-form-item>
                <div class="code-row">
                  <el-input
                    v-model="registerForm.code"
                    placeholder="请输入验证码"
                    maxlength="6"
                    style="width: 60%"
                    :prefix-icon="Lock"
                  />
                  <el-button
                    :disabled="countdown > 0"
                    style="width: 38%"
                    @click="handleSendCode"
                  >
                    {{ countdown > 0 ? `已发送(${countdown}s)` : '获取验证码' }}
                  </el-button>
                </div>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" style="width: 100%; height: 44px; font-size: 16px" @click="handleRegister">
                  注册
                </el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>
        </el-tabs>

        <p class="copyright">© 2025 易人事 · 专为小微企业打造</p>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, Management, BriefCase, Money, Shield, Wallet, Document } from '@element-plus/icons-vue'
import request from '@/api/request'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const activeTab = ref('sms')

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
  }
}

// 密码登录
async function handlePasswordLogin() {
  if (!passwordForm.value.phone || !passwordForm.value.password) {
    ElMessage.error('请填写手机号和密码')
    return
  }
  try {
    const resp = await request.post('/auth/login/password', {
      phone: passwordForm.value.phone,
      password: passwordForm.value.password,
    })
    handleLoginSuccess(resp.data)
  } catch (err: any) {
    if (err.response?.status === 403) {
      ElMessage.error('您的账号为员工账号，请使用员工端微信小程序登录')
      return
    }
    ElMessage.error(err.response?.data?.message || '登录失败')
  }
}

// 微信登录（占位）
function handleWechatLogin() {
  ElMessage.info('微信登录功能开发中')
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
    // TODO: 实现微信登录逻辑
  }
})
</script>

<style scoped>
/* === 左右分栏布局 === */
.login-layout {
  display: grid;
  grid-template-columns: 720px 1fr;
  min-height: 100vh;
}

/* === 左侧品牌区 === */
.login-brand {
  background: linear-gradient(135deg, #1A2D6B 0%, #4F6EF7 60%, #7B9FFF 100%);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 48px 56px;
  position: relative;
  overflow: hidden;
}

/* 装饰圆 */
.login-brand::before {
  content: '';
  position: absolute;
  top: -120px;
  right: -120px;
  width: 360px;
  height: 360px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.06);
}

.login-brand::after {
  content: '';
  position: absolute;
  bottom: -80px;
  left: -80px;
  width: 240px;
  height: 240px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.04);
}

.brand-content {
  position: relative;
  z-index: 1;
}

.brand-header {
  margin-bottom: 32px;
}

.brand-logo {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.brand-logo .el-icon {
  font-size: 32px;
  color: #fff;
}

.brand-name {
  font-size: 24px;
  font-weight: 700;
  color: #fff;
  letter-spacing: 2px;
}

.brand-slogan {
  font-size: 15px;
  color: rgba(255, 255, 255, 0.8);
  margin: 0;
}

.brand-divider {
  width: 40px;
  height: 3px;
  background: rgba(255, 255, 255, 0.4);
  border-radius: 2px;
  margin-bottom: 32px;
}

.feature-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 14px;
}

.feature-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.feature-icon .el-icon {
  font-size: 18px;
  color: rgba(255, 255, 255, 0.9);
}

.feature-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.feature-text strong {
  font-size: 15px;
  font-weight: 600;
  color: #fff;
}

.feature-text span {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.65);
}

.brand-copyright {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.6);
  text-align: center;
  position: relative;
  z-index: 1;
  margin: 0;
}

/* === 右侧表单区 === */
.login-form-panel {
  background: var(--bg-surface, #fff);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  min-width: 0;
}

.login-card {
  width: 100%;
  max-width: 440px;
  padding: 48px;
  border-radius: var(--radius-lg, 12px);
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
  animation: slideUp 0.4s ease;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(16px);
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
  font-size: 26px;
  font-weight: 700;
  color: var(--primary, #4F6EF7);
  margin: 0 0 8px;
  letter-spacing: 2px;
}

.form-brand-subtitle {
  font-size: 14px;
  color: var(--text-secondary, #5E6C84);
  margin: 0;
}

.login-tabs {
  margin-bottom: 16px;
}

.code-row {
  display: flex;
  gap: 8px;
  width: 100%;
}

.code-row .el-input {
  flex: 1;
}

.copyright {
  text-align: center;
  font-size: 12px;
  color: #bbb;
  margin-top: 16px;
  margin-bottom: 0;
}

/* === 移动端响应式 === */
@media (max-width: 768px) {
  .login-layout {
    grid-template-columns: 1fr;
  }

  .login-brand {
    display: none;
  }

  .login-form-panel {
    align-items: flex-start;
    padding-top: 48px;
    min-height: 100vh;
  }

  .login-card {
    padding: 32px 24px;
  }
}
</style>
