<template>
  <div class="mine-view">
    <!-- 用户卡片 -->
    <div class="profile-card">
      <div class="profile-bg" />
      <div class="profile-main">
        <div class="avatar-wrap">
          <div class="avatar" :style="{ background: avatarColor(user?.name || '') }">
            {{ user?.name?.[0] || '?' }}
          </div>
          <span class="avatar-badge">
            <el-icon :size="10" color="#fff"><Check /></el-icon>
          </span>
        </div>
        <div class="profile-info">
          <h2 class="profile-name">{{ user?.name || '未登录' }}</h2>
          <p class="profile-phone">{{ formatPhone(user?.phone || '') }}</p>
          <div class="profile-org">
            <el-icon :size="12" color="#0F766E"><OfficeBuilding /></el-icon>
            {{ org?.name || '我的企业' }}
          </div>
        </div>
        <button class="edit-btn" @click="handleEditOrg" aria-label="编辑资料">
          <el-icon :size="16" color="#64748B"><EditPen /></el-icon>
        </button>
      </div>
    </div>

    <!-- 企业信息卡片 -->
    <div class="section">
      <div class="section-title">企业信息</div>
      <div class="info-list">
        <div class="info-row" v-if="org?.name">
          <span class="info-label">企业名称</span>
          <span class="info-value">{{ org.name }}</span>
        </div>
        <div class="info-row" v-if="org?.credit_code">
          <span class="info-label">统一社会信用代码</span>
          <span class="info-value">{{ org.credit_code }}</span>
        </div>
        <div class="info-row" v-if="org?.city">
          <span class="info-label">所在城市</span>
          <span class="info-value">{{ org.city }}</span>
        </div>
        <div class="info-row" v-if="org?.contact_name">
          <span class="info-label">联系人</span>
          <span class="info-value">{{ org.contact_name }}</span>
        </div>
        <div class="info-row" v-if="org?.contact_phone">
          <span class="info-label">联系电话</span>
          <a :href="`tel:${org.contact_phone}`" class="info-link">
            {{ org.contact_phone }}
            <el-icon :size="12"><Phone /></el-icon>
          </a>
        </div>
        <div v-if="!org?.name" class="empty-info">
          <p>尚未设置企业信息</p>
          <button class="setup-btn" @click="handleEditOrg">去设置</button>
        </div>
      </div>
    </div>

    <!-- 功能列表 -->
    <div class="section">
      <div class="section-title">设置</div>
      <div class="menu-list">
        <router-link to="/onboarding/org-setup" class="menu-item">
          <div class="menu-left">
            <div class="menu-icon" style="background: #DCFCE7; color: #16A34A">
              <el-icon :size="18"><Setting /></el-icon>
            </div>
            <span class="menu-label">编辑企业信息</span>
          </div>
          <el-icon :size="16" color="#CBD5E1"><ArrowRight /></el-icon>
        </router-link>

        <button class="menu-item" @click="handleChangePassword">
          <div class="menu-left">
            <div class="menu-icon" style="background: #E0E7FF; color: #4F46E5">
              <el-icon :size="18"><Lock /></el-icon>
            </div>
            <span class="menu-label">修改密码</span>
          </div>
          <el-icon :size="16" color="#CBD5E1"><ArrowRight /></el-icon>
        </button>

        <button class="menu-item" @click="handleAbout">
          <div class="menu-left">
            <div class="menu-icon" style="background: #FEF3C7; color: #D97706">
              <el-icon :size="18"><InfoFilled /></el-icon>
            </div>
            <span class="menu-label">关于我们</span>
          </div>
          <el-icon :size="16" color="#CBD5E1"><ArrowRight /></el-icon>
        </button>
      </div>
    </div>

    <!-- 退出登录 -->
    <div class="section logout-section">
      <button class="logout-btn" @click="handleLogout">
        <el-icon :size="18"><SwitchButton /></el-icon>
        退出登录
      </button>
    </div>

    <!-- 版本信息 -->
    <p class="version-text">易人事 v1.0.0</p>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAuthStore } from '@/stores/auth'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Check, OfficeBuilding, EditPen, Phone, Setting,
  Lock, InfoFilled, SwitchButton, ArrowRight,
} from '@element-plus/icons-vue'
import request from '@/api/request'

const router = useRouter()
const userStore = useUserStore()
const authStore = useAuthStore()
const user = ref(userStore.user)
const org = ref(userStore.org)

const AVATAR_COLORS = ['#0F766E', '#0EA5E9', '#8B5CF6', '#F59E0B', '#EF4444', '#10B981', '#EC4899', '#6366F1']
function avatarColor(name: string): string {
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length]
}

function formatPhone(phone: string): string {
  if (!phone || phone.length !== 11) return phone
  return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
}

function handleLogout() {
  ElMessageBox.confirm('确定要退出登录吗？退出后需要重新登录。', '退出登录', {
    confirmButtonText: '确定退出',
    cancelButtonText: '取消',
    type: 'warning',
    confirmButtonClass: 'logout-confirm-btn',
  }).then(() => {
    authStore.logout()
    userStore.clear()
    router.push('/login')
  }).catch(() => {})
}

function handleEditOrg() {
  router.push('/onboarding/org-setup')
}

function handleChangePassword() {
  ElMessage.info('修改密码功能开发中')
}

function handleAbout() {
  ElMessage.info('易人事 v1.0.0 — 专为小微企业打造的轻量化人事管理工具')
}

async function loadOrgInfo() {
  try {
    const res = await request.get('/orgs/current')
    userStore.setUser(res.user)
    userStore.setOrg(res.org)
    user.value = res.user
    org.value = res.org
  } catch {
    // ignore
  }
}

onMounted(() => {
  loadOrgInfo()
})
</script>

<style scoped lang="scss">
.mine-view {
  background: #F8FAFC;
  min-height: 100%;
  padding-bottom: 80px;
}

// ===== Profile Card =====
.profile-card {
  background: #fff;
  position: relative;
  overflow: hidden;
  margin-bottom: 12px;
}

.profile-bg {
  height: 80px;
  background: linear-gradient(135deg, #0F766E 0%, #14B8A6 100%);
}

.profile-main {
  display: flex;
  align-items: flex-end;
  gap: 14px;
  padding: 0 16px 20px;
  margin-top: -36px;
  position: relative;
}

.avatar-wrap {
  position: relative;
  flex-shrink: 0;
}

.avatar {
  width: 68px;
  height: 68px;
  border-radius: 50%;
  border: 3px solid #fff;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  font-weight: 700;
  color: #fff;
}

.avatar-badge {
  position: absolute;
  bottom: 2px;
  right: 2px;
  width: 18px;
  height: 18px;
  background: #0F766E;
  border-radius: 50%;
  border: 2px solid #fff;
  display: flex;
  align-items: center;
  justify-content: center;
}

.profile-info {
  flex: 1;
  min-width: 0;
  padding-bottom: 4px;
}

.profile-name {
  font-size: 20px;
  font-weight: 700;
  color: #0F172A;
  margin: 0 0 2px;
  letter-spacing: -0.3px;
}

.profile-phone {
  font-size: 13px;
  color: #94A3B8;
  margin: 0 0 4px;
  font-feature-settings: "tnum";
}

.profile-org {
  font-size: 12px;
  color: #64748B;
  display: flex;
  align-items: center;
  gap: 4px;
}

.edit-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s;
  flex-shrink: 0;

  &:hover { background: #F1F5F9; border-color: #CBD5E1; }
  &:active { background: #E2E8F0; }
}

// ===== 通用 Section =====
.section {
  background: #fff;
  margin: 0 12px 12px;
  border-radius: 16px;
  padding: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #94A3B8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 12px;
}

// ===== Info List =====
.info-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #F1F5F9;

  &:last-child { border-bottom: none; padding-bottom: 0; }
  &:first-child { padding-top: 0; }
}

.info-label {
  font-size: 14px;
  color: #64748B;
  flex-shrink: 0;
}

.info-value {
  font-size: 14px;
  color: #0F172A;
  font-weight: 500;
  text-align: right;
  word-break: break-all;
}

.info-link {
  font-size: 14px;
  color: #0F766E;
  font-weight: 500;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;

  &:hover { color: #0D6B62; }
}

.empty-info {
  text-align: center;
  padding: 16px 0 8px;

  p { font-size: 14px; color: #94A3B8; margin: 0 0 10px; }
}

.setup-btn {
  padding: 8px 20px;
  background: #0F766E;
  color: #fff;
  border: none;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s;

  &:hover { background: #0D6B62; }
  &:active { background: #115E59; }
}

// ===== Menu List =====
.menu-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border-radius: 10px;
  cursor: pointer;
  text-decoration: none;
  background: none;
  border: none;
  width: 100%;
  transition: background 0.15s;
  min-height: 44px;

  &:hover { background: #F8FAFC; }
  &:active { background: #F1F5F9; }
}

.menu-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.menu-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.menu-label {
  font-size: 15px;
  color: #0F172A;
  font-weight: 500;
}

// ===== 退出登录 =====
.logout-section {
  background: transparent;
  padding: 0;
  margin: 8px 12px;
  box-shadow: none;
}

.logout-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 14px;
  background: #fff;
  border: 1px solid #F1F5F9;
  border-radius: 16px;
  font-size: 15px;
  font-weight: 500;
  color: #EF4444;
  cursor: pointer;
  transition: background 0.15s;
  min-height: 50px;

  &:hover { background: #FEF2F2; }
  &:active { background: #FEE2E2; }
}

// ===== 版本信息 =====
.version-text {
  text-align: center;
  font-size: 12px;
  color: #CBD5E1;
  margin: 8px 0 0;
  padding-bottom: 8px;
}
</style>
