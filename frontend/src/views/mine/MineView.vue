<template>
  <div class="mine-view">
    <!-- User avatar card -->
    <el-card class="mb-4">
      <div class="user-header">
        <el-avatar
          :size="60"
          style="background: #409eff; font-size: 24px; cursor: pointer"
          :src="user?.avatar || undefined"
          @click="triggerAvatarUpload"
        >
          <img v-if="user?.avatar" :src="user.avatar" alt="avatar" />
          <span v-else>{{ user?.name?.[0] || '?' }}</span>
        </el-avatar>
        <input
          ref="avatarInputRef"
          type="file"
          accept="image/jpeg,image/png"
          style="display: none"
          @change="handleAvatarChange"
        />
        <div class="user-info">
          <div class="user-name" style="cursor: pointer" @click="openNameDialog">{{ user?.name || '点击设置姓名' }}</div>
          <div class="user-phone">{{ user?.phone || '' }}</div>
          <div class="user-org">{{ org?.name || '' }}</div>
        </div>
      </div>
    </el-card>

    <!-- Org info card -->
    <el-card class="mb-4">
      <template #header>
        <span>企业信息</span>
      </template>
      <el-descriptions :column="1" border size="small">
        <el-descriptions-item label="企业名称">{{ org?.name || '—' }}</el-descriptions-item>
        <el-descriptions-item label="统一社会信用代码">{{ org?.credit_code || '—' }}</el-descriptions-item>
        <el-descriptions-item label="城市">{{ org?.city || '—' }}</el-descriptions-item>
        <el-descriptions-item label="联系人">{{ org?.contact_name || '—' }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ org?.contact_phone || '—' }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- Account actions -->
    <el-card>
      <el-space direction="vertical" fill style="width: 100%">
        <el-button @click="openEditOrgDialog" style="width: 100%; text-align: left">
          <el-icon><Setting /></el-icon> 编辑企业信息
        </el-button>
        <el-button @click="openChangePasswordDialog" style="width: 100%; text-align: left">
          <el-icon><Lock /></el-icon> 修改密码
        </el-button>
        <el-button type="danger" plain @click="handleLogout" style="width: 100%; text-align: left">
          <el-icon><SwitchButton /></el-icon> 退出登录
        </el-button>
      </el-space>
    </el-card>

    <!-- 修改密码弹窗 -->
    <el-dialog v-model="passwordDialogVisible" title="修改密码" width="400px">
      <el-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" label-width="90px">
        <el-form-item label="旧密码" prop="old_password">
          <el-input v-model="passwordForm.old_password" type="password" placeholder="请输入旧密码" show-password />
        </el-form-item>
        <el-form-item label="新密码" prop="new_password">
          <el-input v-model="passwordForm.new_password" type="password" placeholder="6-20位字母或数字" show-password />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirm_password">
          <el-input v-model="passwordForm.confirm_password" type="password" placeholder="再次输入新密码" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="passwordSaving" @click="handleChangePassword">确定</el-button>
      </template>
    </el-dialog>

    <!-- 修改姓名弹窗 -->
    <el-dialog v-model="nameDialogVisible" title="修改姓名" width="360px">
      <el-form ref="nameFormRef" :model="nameForm" :rules="nameRules" label-width="80px">
        <el-form-item label="姓名" prop="name">
          <el-input v-model="nameForm.name" placeholder="请输入姓名" maxlength="50" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="nameDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="nameSaving" @click="handleUpdateName">确定</el-button>
      </template>
    </el-dialog>

    <!-- 编辑企业信息弹窗 -->
    <el-dialog v-model="editOrgDialogVisible" title="编辑企业信息" width="480px">
      <el-form ref="editOrgFormRef" :model="editOrgForm" :rules="editOrgRules" label-width="140px">
        <el-form-item label="企业名称" prop="name">
          <el-input v-model="editOrgForm.name" placeholder="请输入企业全称" maxlength="100" />
        </el-form-item>
        <el-form-item label="统一社会信用代码" prop="credit_code">
          <el-input v-model="editOrgForm.credit_code" placeholder="18位统一社会信用代码" maxlength="18" />
        </el-form-item>
        <el-form-item label="城市" prop="city">
          <el-input v-model="editOrgForm.city" placeholder="请输入所在城市" />
        </el-form-item>
        <el-form-item label="联系人" prop="contact_name">
          <el-input v-model="editOrgForm.contact_name" placeholder="请输入联系人姓名" maxlength="50" />
        </el-form-item>
        <el-form-item label="联系电话" prop="contact_phone">
          <el-input v-model="editOrgForm.contact_phone" placeholder="请输入手机号" maxlength="11" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editOrgDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editOrgSaving" @click="handleEditOrg">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAuthStore } from '@/stores/auth'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Setting, Lock, SwitchButton } from '@element-plus/icons-vue'
import request from '@/api/request'

const router = useRouter()
const userStore = useUserStore()
const authStore = useAuthStore()
const user = ref(userStore.user)
const org = ref(userStore.org)

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
  old_password: [
    { required: true, message: '请输入旧密码', trigger: 'blur' },
  ],
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
  if (passwordSaving.value) return
  if (!passwordFormRef.value) return
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
  if (nameSaving.value) return
  if (!nameFormRef.value) return
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
  // 预填充当前企业信息（如果已加载）
  if (org.value) {
    editOrgForm.name = org.value.name || ''
    editOrgForm.credit_code = org.value.credit_code || ''
    editOrgForm.city = org.value.city || ''
    editOrgForm.contact_name = org.value.contact_name || ''
    editOrgForm.contact_phone = org.value.contact_phone || ''
  } else {
    editOrgForm.name = ''
    editOrgForm.credit_code = ''
    editOrgForm.city = ''
    editOrgForm.contact_name = ''
    editOrgForm.contact_phone = ''
  }
  editOrgDialogVisible.value = true
}

async function handleEditOrg() {
  if (editOrgSaving.value) return
  if (!editOrgFormRef.value) return
  try {
    await editOrgFormRef.value.validate()
  } catch {
    return
  }
  editOrgSaving.value = true
  try {
    await request.put('/org', {
      name: editOrgForm.name,
      credit_code: editOrgForm.credit_code,
      city: editOrgForm.city,
      contact_name: editOrgForm.contact_name,
      contact_phone: editOrgForm.contact_phone,
    })
    ElMessage.success('企业信息更新成功')
    editOrgDialogVisible.value = false
    // 刷新本地 org 数据并重新加载页面信息
    await loadOrgInfo()
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '更新企业信息失败')
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
  } catch (err) {
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
  // 清空 input，允许重复选择同一文件
  if (avatarInputRef.value) avatarInputRef.value.value = ''
}

onMounted(() => {
  loadOrgInfo()
})
</script>

<style scoped lang="scss">
.mine-view {
  padding: 8px;
}
.mb-4 {
  margin-bottom: 12px;
}
.user-header {
  display: flex;
  align-items: center;
  gap: 16px;
}
.user-info {
  .user-name {
    font-size: 18px;
    font-weight: 600;
    color: #333;
  }
  .user-phone {
    font-size: 13px;
    color: #999;
    margin-top: 2px;
  }
  .user-org {
    font-size: 12px;
    color: #409eff;
    margin-top: 2px;
  }
}
</style>
