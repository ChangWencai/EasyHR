<template>
  <div class="mine-view">
    <!-- 用户信息卡片 -->
    <div class="user-profile-card">
      <div class="profile-bg"></div>
      <div class="profile-content">
        <div class="avatar-wrapper" @click="triggerAvatarUpload">
          <el-avatar :size="80" :src="user?.avatar || undefined" class="avatar">
            <span class="avatar-fallback">{{ user?.name?.[0] || '?' }}</span>
          </el-avatar>
          <div class="avatar-edit">
            <el-icon><Camera /></el-icon>
          </div>
          <input
            ref="avatarInputRef"
            type="file"
            accept="image/jpeg,image/png"
            style="display: none"
            @change="handleAvatarChange"
          />
        </div>
        <div class="user-info">
          <h2 class="user-name">{{ user?.name || '点击设置姓名' }}</h2>
          <p class="user-phone">{{ formatPhone(user?.phone || '') }}</p>
          <div class="user-role">
            <el-icon><UserFilled /></el-icon>
            {{ user?.role === 'owner' ? '企业管理员' : '员工' }}
          </div>
        </div>
        <div class="profile-actions">
          <el-button @click="openNameDialog">
            <el-icon><EditPen /></el-icon>
            修改姓名
          </el-button>
          <el-button type="primary" @click="openEditOrgDialog">
            <el-icon><Setting /></el-icon>
            企业设置
          </el-button>
        </div>
      </div>
    </div>

    <!-- 企业信息 -->
    <div class="org-info-card glass-card">
      <div class="card-header">
        <h3 class="card-title">
          <el-icon><OfficeBuilding /></el-icon>
          企业信息
        </h3>
        <el-tag v-if="org" type="success" size="small">已认证</el-tag>
      </div>
      <div v-if="org" class="org-details">
        <div class="org-item">
          <span class="org-label">企业名称</span>
          <span class="org-value">{{ org.name }}</span>
        </div>
        <div class="org-item">
          <span class="org-label">统一社会信用代码</span>
          <span class="org-value">{{ org.credit_code || '—' }}</span>
        </div>
        <div class="org-item">
          <span class="org-label">所在城市</span>
          <span class="org-value">{{ org.city || '—' }}</span>
        </div>
        <div class="org-item">
          <span class="org-label">联系人</span>
          <span class="org-value">{{ org.contact_name || '—' }}</span>
        </div>
        <div class="org-item">
          <span class="org-label">联系电话</span>
          <span class="org-value">{{ org.contact_phone || '—' }}</span>
        </div>
      </div>
      <div v-else class="org-empty">
        <p>暂无企业信息</p>
        <el-button type="primary" @click="openEditOrgDialog">创建企业</el-button>
      </div>
    </div>

    <!-- 操作菜单 -->
    <div class="actions-card glass-card">
      <div class="card-header">
        <h3 class="card-title">
          <el-icon><Tools /></el-icon>
          账号设置
        </h3>
      </div>
      <div class="action-list">
        <div class="action-item" @click="openChangePasswordDialog">
          <div class="action-icon action-icon--security">
            <el-icon><Lock /></el-icon>
          </div>
          <div class="action-info">
            <span class="action-title">修改密码</span>
            <span class="action-desc">定期更换密码，保护账号安全</span>
          </div>
          <el-icon class="action-arrow"><ArrowRight /></el-icon>
        </div>
        <div class="action-item" @click="openEditOrgDialog">
          <div class="action-icon action-icon--org">
            <el-icon><OfficeBuilding /></el-icon>
          </div>
          <div class="action-info">
            <span class="action-title">编辑企业信息</span>
            <span class="action-desc">更新企业基本信息和联系方式</span>
          </div>
          <el-icon class="action-arrow"><ArrowRight /></el-icon>
        </div>
        <div class="action-item" @click="handleLogout">
          <div class="action-icon action-icon--danger">
            <el-icon><SwitchButton /></el-icon>
          </div>
          <div class="action-info">
            <span class="action-title logout-text">退出登录</span>
            <span class="action-desc">确定要退出当前账号吗？</span>
          </div>
          <el-icon class="action-arrow"><ArrowRight /></el-icon>
        </div>
      </div>
    </div>

    <!-- 版本信息 -->
    <div class="version-info">
      <span>易人事 v1.3.0</span>
      <span class="version-divider">·</span>
      <span>Made with for SMB</span>
    </div>

    <!-- 修改密码弹窗 -->
    <el-dialog
      v-model="passwordDialogVisible"
      title="修改密码"
      width="440px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        label-position="top"
      >
        <el-form-item label="旧密码" prop="old_password">
          <el-input
            v-model="passwordForm.old_password"
            type="password"
            placeholder="请输入旧密码"
            show-password
            size="large"
          />
        </el-form-item>
        <el-form-item label="新密码" prop="new_password">
          <el-input
            v-model="passwordForm.new_password"
            type="password"
            placeholder="6-20位字母或数字"
            show-password
            size="large"
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirm_password">
          <el-input
            v-model="passwordForm.confirm_password"
            type="password"
            placeholder="再次输入新密码"
            show-password
            size="large"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="passwordSaving" @click="handleChangePassword">
          确定修改
        </el-button>
      </template>
    </el-dialog>

    <!-- 修改姓名弹窗 -->
    <el-dialog
      v-model="nameDialogVisible"
      title="修改姓名"
      width="400px"
      :close-on-click-modal="false"
    >
      <el-form ref="nameFormRef" :model="nameForm" :rules="nameRules" label-position="top">
        <el-form-item label="姓名" prop="name">
          <el-input
            v-model="nameForm.name"
            placeholder="请输入姓名"
            maxlength="50"
            size="large"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="nameDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="nameSaving" @click="handleUpdateName">
          确定修改
        </el-button>
      </template>
    </el-dialog>

    <!-- 编辑企业信息弹窗 -->
    <el-dialog
      v-model="editOrgDialogVisible"
      title="编辑企业信息"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="editOrgFormRef"
        :model="editOrgForm"
        :rules="editOrgRules"
        label-position="top"
      >
        <el-form-item label="企业名称" prop="name">
          <el-input v-model="editOrgForm.name" placeholder="请输入企业全称" maxlength="100" size="large" />
        </el-form-item>
        <el-form-item label="统一社会信用代码" prop="credit_code">
          <el-input
            v-model="editOrgForm.credit_code"
            placeholder="18位统一社会信用代码"
            maxlength="18"
            size="large"
          />
        </el-form-item>
        <el-form-item label="城市" prop="city">
          <el-select
            v-model="editOrgForm.city"
            filterable
            reserve-keyword
            placeholder="点击选择或输入搜索城市"
            :loading="cityLoading"
            size="large"
            style="width: 100%"
            @focus="fetchCityOptions('')"
          >
            <el-option
              v-for="city in cityOptions"
              :key="city.code"
              :label="city.name"
              :value="city.name"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="联系人" prop="contact_name">
          <el-input v-model="editOrgForm.contact_name" placeholder="请输入联系人姓名" maxlength="50" size="large" />
        </el-form-item>
        <el-form-item label="联系电话" prop="contact_phone">
          <el-input v-model="editOrgForm.contact_phone" placeholder="请输入手机号" maxlength="11" size="large" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editOrgDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editOrgSaving" @click="handleEditOrg">
          保存修改
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAuthStore } from '@/stores/auth'
import {
  ElMessage,
  ElMessageBox,
  type FormInstance,
  type FormRules,
} from 'element-plus'
import {
  Setting,
  Lock,
  SwitchButton,
  EditPen,
  Camera,
  UserFilled,
  OfficeBuilding,
  Tools,
  ArrowRight,
} from '@element-plus/icons-vue'
import request from '@/api/request'

// ========== 城市下拉 ==========
interface CityOption {
  code: string
  name: string
}

const cityOptions = ref<CityOption[]>([])
const cityLoading = ref(false)
let citySearchTimer: ReturnType<typeof setTimeout> | null = null

async function fetchCityOptions(query: string) {
  if (citySearchTimer) clearTimeout(citySearchTimer)
  if (!query && cityOptions.value.length > 0) return
  citySearchTimer = setTimeout(async () => {
    cityLoading.value = true
    try {
      const res = await request.get('/cities', { params: query ? { search: query } : {} })
      cityOptions.value = (res.data ?? []).map((item: any) => ({
        code: item.code ?? item.city_code ?? '',
        name: item.name ?? item.city_name ?? item,
      }))
    } catch {
      cityOptions.value = []
    } finally {
      cityLoading.value = false
    }
  }, 200)
}

// ========== 修改姓名 ==========
const router = useRouter()
const userStore = useUserStore()
const authStore = useAuthStore()
const user = ref(userStore.user)
const org = ref(userStore.org)

function formatPhone(phone: string): string {
  if (!phone || phone.length !== 11) return phone
  return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
}

// ========== 修改密码 ==========
const passwordDialogVisible = ref(false)
const passwordSaving = ref(false)
const passwordFormRef = ref<FormInstance>()

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

const validateConfirmPassword = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  if (value !== passwordForm.new_password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const passwordRules: FormRules = {
  old_password: [{ required: true, message: '请输入旧密码', trigger: 'blur' }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度为6-20位', trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
}

function openChangePasswordDialog() {
  passwordForm.old_password = ''
  passwordForm.new_password = ''
  passwordForm.confirm_password = ''
  passwordDialogVisible.value = true
}

async function handleChangePassword() {
  if (passwordSaving.value || !passwordFormRef.value) return
  try {
    await passwordFormRef.value.validate()
  } catch {
    return
  }
  passwordSaving.value = true
  try {
    await request.put('/auth/password', {
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password,
    })
    ElMessage.success('密码修改成功')
    passwordDialogVisible.value = false
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '修改密码失败')
  } finally {
    passwordSaving.value = false
  }
}

// ========== 修改姓名 ==========
const nameDialogVisible = ref(false)
const nameSaving = ref(false)
const nameFormRef = ref<FormInstance>()
const nameForm = reactive({ name: '' })

const nameRules: FormRules = {
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' },
    { min: 2, max: 50, message: '姓名长度为2-50位', trigger: 'blur' },
  ],
}

function openNameDialog() {
  nameForm.name = user.value?.name || ''
  nameDialogVisible.value = true
}

async function handleUpdateName() {
  if (nameSaving.value || !nameFormRef.value) return
  try {
    await nameFormRef.value.validate()
  } catch {
    return
  }
  nameSaving.value = true
  try {
    await request.put('/auth/name', { name: nameForm.name })
    user.value = { ...user.value!, name: nameForm.name }
    userStore.setUser({ ...user.value! })
    ElMessage.success('姓名修改成功')
    nameDialogVisible.value = false
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '姓名修改失败')
  } finally {
    nameSaving.value = false
  }
}

// ========== 编辑企业信息 ==========
const editOrgDialogVisible = ref(false)
const editOrgSaving = ref(false)
const editOrgFormRef = ref<FormInstance>()

const editOrgForm = reactive({
  name: '',
  credit_code: '',
  city: '',
  contact_name: '',
  contact_phone: '',
})

const editOrgRules: FormRules = {
  name: [
    { required: true, message: '请输入企业名称', trigger: 'blur' },
    { min: 2, max: 100, message: '企业名称长度为2-100位', trigger: 'blur' },
  ],
  credit_code: [
    { required: true, message: '请输入统一社会信用代码', trigger: 'blur' },
    { pattern: /^[1-9][0-9A-HJ-NPQRTUWXY]{17}$/, message: '统一社会信用代码格式不正确', trigger: 'blur' },
  ],
  city: [{ required: true, message: '请输入所在城市', trigger: 'blur' }],
  contact_name: [{ required: true, message: '请输入联系人', trigger: 'blur' }],
  contact_phone: [
    { required: true, message: '请输入联系电话', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确', trigger: 'blur' },
  ],
}

function openEditOrgDialog() {
  if (org.value) {
    editOrgForm.name = org.value.name || ''
    editOrgForm.credit_code = org.value.credit_code || ''
    editOrgForm.city = org.value.city || ''
    editOrgForm.contact_name = org.value.contact_name || ''
    editOrgForm.contact_phone = org.value.contact_phone || ''
  } else {
    Object.assign(editOrgForm, { name: '', credit_code: '', city: '', contact_name: '', contact_phone: '' })
  }
  editOrgDialogVisible.value = true
  fetchCityOptions('')
}

async function handleEditOrg() {
  if (editOrgSaving.value || !editOrgFormRef.value) return
  try {
    await editOrgFormRef.value.validate()
  } catch {
    return
  }
  editOrgSaving.value = true
  try {
    if (org.value) {
      // 已有企业 → 更新
      await request.put('/org', {
        name: editOrgForm.name,
        credit_code: editOrgForm.credit_code,
        city: editOrgForm.city,
        contact_name: editOrgForm.contact_name,
        contact_phone: editOrgForm.contact_phone,
      })
      ElMessage.success('企业信息更新成功')
    } else {
      // 无企业 → 创建并获取含正确 org_id 的新 token
      const resp = await request.put('/auth/org/onboarding', {
        name: editOrgForm.name,
        credit_code: editOrgForm.credit_code,
        city: editOrgForm.city,
        contact_name: editOrgForm.contact_name,
        contact_phone: editOrgForm.contact_phone,
      })
      // 更新本地 JWT token（使其含正确 org_id）
      if (resp.data?.access_token) {
        authStore.setToken(resp.data.access_token)
      }
      ElMessage.success('企业创建成功')
    }
    editOrgDialogVisible.value = false
    await loadOrgInfo()
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '操作失败')
  } finally {
    editOrgSaving.value = false
  }
}

// ========== 退出登录 ==========
function handleLogout() {
  ElMessageBox.confirm('确定要退出登录吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
    buttonSize: 'large',
  }).then(() => {
    authStore.logout()
    userStore.clear()
    router.push('/login')
  }).catch(() => {})
}

// ========== 加载企业信息 ==========
async function loadOrgInfo() {
  try {
    const res = await request.get('/auth/me')
    const data = res.data
    const userInfo = {
      id: data.id,
      name: data.name || '',
      phone: data.phone || '',
      role: data.role,
      avatar: data.avatar || '',
    }
    userStore.setUser(userInfo)
    userStore.setOrg(data.org || null)
    user.value = userInfo
    org.value = data.org || null
  } catch {
    ElMessage.error('加载企业信息失败')
  }
}

// ========== 头像上传 ==========
const avatarInputRef = ref<HTMLInputElement>()

function triggerAvatarUpload() {
  avatarInputRef.value?.click()
}

async function handleAvatarChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = async (ev) => {
    const base64 = ev.target?.result as string
    try {
      await request.put('/auth/avatar', { avatar: base64 })
      user.value = { ...user.value!, avatar: base64 }
      userStore.setUser({ ...user.value! })
      ElMessage.success('头像更新成功')
    } catch (err: any) {
      ElMessage.error(err.response?.data?.message || '头像上传失败')
    }
  }
  reader.readAsDataURL(file)
  if (avatarInputRef.value) avatarInputRef.value.value = ''
}

onMounted(() => {
  loadOrgInfo()
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
$warning: #F59E0B;
$error: #EF4444;
$shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
$shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
$shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
$radius-sm: 8px;
$radius-md: 12px;
$radius-lg: 16px;
$radius-xl: 24px;

// ============================================================
// 基础布局
// ============================================================
.mine-view {
  padding: 24px 32px;
  width: 100%;
  box-sizing: border-box;
  background: $bg-page;
  min-height: 100vh;
}

// ============================================================
// 玻璃态卡片
// ============================================================
.glass-card {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: $radius-xl;
  box-shadow: $shadow-md;
}

// ============================================================
// 用户信息卡片
// ============================================================
.user-profile-card {
  position: relative;
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
  border-radius: $radius-xl;
  overflow: hidden;
  margin-bottom: 24px;
}

.profile-bg {
  position: absolute;
  inset: 0;
  background: url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='0.05'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
  opacity: 0.5;
}

.profile-content {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 32px;
}

.avatar-wrapper {
  position: relative;
  cursor: pointer;

  &:hover .avatar-edit {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1);
  }
}

.avatar {
  width: 80px;
  height: 80px;
  border: 3px solid rgba(255, 255, 255, 0.3);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);

  .avatar-fallback {
    font-size: 32px;
    font-weight: 600;
    background: rgba(255, 255, 255, 0.2);
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}

.avatar-edit {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) scale(0.8);
  width: 32px;
  height: 32px;
  background: rgba(0, 0, 0, 0.6);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: all 0.2s ease;
  color: #fff;
  font-size: 14px;
}

.user-info {
  flex: 1;

  .user-name {
    font-size: 24px;
    font-weight: 700;
    color: #fff;
    margin: 0 0 8px;
  }

  .user-phone {
    font-size: 14px;
    color: rgba(255, 255, 255, 0.8);
    margin: 0 0 8px;
  }

  .user-role {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 12px;
    font-size: 12px;
    color: #fff;
    background: rgba(255, 255, 255, 0.15);
    border-radius: 20px;
  }
}

.profile-actions {
  display: flex;
  gap: 12px;

  :deep(.el-button) {
    padding: 10px 20px;
    border-radius: $radius-md;
    font-weight: 500;

    &.el-button--default {
      background: rgba(255, 255, 255, 0.15);
      border-color: rgba(255, 255, 255, 0.3);
      color: #fff;

      &:hover {
        background: rgba(255, 255, 255, 0.25);
      }
    }

    .el-icon {
      margin-right: 4px;
    }
  }
}

// ============================================================
// 企业信息卡片
// ============================================================
.org-info-card {
  padding: 24px;
  margin-bottom: 24px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: $text-primary;
  margin: 0;

  .el-icon {
    color: var(--primary);
  }
}

.org-details {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.org-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px 16px;
  background: $bg-page;
  border-radius: $radius-md;
}

.org-label {
  font-size: 12px;
  color: $text-muted;
}

.org-value {
  font-size: 14px;
  font-weight: 500;
  color: $text-primary;
}

.org-empty {
  text-align: center;
  padding: 32px;

  p {
    color: $text-secondary;
    margin: 0 0 16px;
  }
}

// ============================================================
// 操作菜单卡片
// ============================================================
.actions-card {
  padding: 24px;
  margin-bottom: 24px;
}

.action-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.action-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  border-radius: $radius-md;
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    background: rgba(var(--primary), 0.04);

    .action-arrow {
      color: var(--primary);
      transform: translateX(4px);
    }
  }
}

.action-icon {
  width: 44px;
  height: 44px;
  border-radius: $radius-md;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;

  .el-icon {
    font-size: 20px;
  }

  &--security {
    background: linear-gradient(135deg, #DBEAFE 0%, #BFDBFE 100%);
    color: #3B82F6;
  }

  &--org {
    background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%);
    color: $warning;
  }

  &--danger {
    background: linear-gradient(135deg, #FEE2E2 0%, #FECACA 100%);
    color: $error;
  }
}

.action-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.action-title {
  font-size: 15px;
  font-weight: 500;
  color: $text-primary;

  &.logout-text {
    color: $error;
  }
}

.action-desc {
  font-size: 13px;
  color: $text-secondary;
}

.action-arrow {
  color: $text-muted;
  transition: all 0.2s ease;
}

// ============================================================
// 版本信息
// ============================================================
.version-info {
  text-align: center;
  font-size: 12px;
  color: $text-muted;
  padding: 16px;
}

.version-divider {
  margin: 0 8px;
}

// ============================================================
// 响应式
// ============================================================
@media (max-width: 768px) {
  .mine-view {
    padding: 16px;
  }

  .profile-content {
    flex-direction: column;
    text-align: center;
    padding: 24px 16px;
  }

  .profile-actions {
    width: 100%;
    justify-content: center;
  }

  .org-details {
    grid-template-columns: 1fr;
  }
}
</style>