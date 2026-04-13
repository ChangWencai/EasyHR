<template>
  <div class="login-page">
    <div class="login-card">
      <div class="brand">
        <h1>易人事</h1>
        <p>老板管理后台</p>
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

      <p class="copyright">© 2024 易人事</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
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
.login-page {
  min-height: 100vh;
  background-color: #1677ff;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.login-card {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.brand {
  text-align: center;
  margin-bottom: 24px;
}

.brand h1 {
  font-size: 24px;
  font-weight: bold;
  color: #1677ff;
  margin: 0 0 8px;
}

.brand p {
  font-size: 14px;
  color: #999;
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

.wechat-placeholder {
  text-align: center;
  padding: 16px 0;
  color: #666;
  font-size: 14px;
}

.wechat-placeholder p {
  margin-bottom: 16px;
}

.copyright {
  text-align: center;
  font-size: 12px;
  color: #bbb;
  margin-top: 16px;
  margin-bottom: 0;
}

/* 移动端适配 */
@media (max-width: 480px) {
  .login-card {
    padding: 16px;
    border-radius: 8px;
  }

  .brand h1 {
    font-size: 20px;
  }

  .brand p {
    font-size: 12px;
  }
}
</style>
