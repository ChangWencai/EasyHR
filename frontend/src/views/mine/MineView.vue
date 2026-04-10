<template>
  <div class="mine-view">
    <!-- User avatar card -->
    <el-card class="mb-4">
      <div class="user-header">
        <el-avatar :size="60" style="background: #409eff; font-size: 24px">
          {{ user?.name?.[0] || '?' }}
        </el-avatar>
        <div class="user-info">
          <div class="user-name">{{ user?.name || '未登录' }}</div>
          <div class="user-phone">{{ user?.phone || '' }}</div>
          <div class="user-org">{{ org?.name || '' }}</div>
        </div>
      </div>
    </el-card>

    <!-- Org info card -->
    <el-card class="mb-4">
      <template #header>
        <div class="card-header">
          <span>企业信息</span>
          <el-button size="small" type="primary" @click="handleEditOrg">编辑</el-button>
        </div>
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
        <el-button @click="handleEditOrg" style="width: 100%; text-align: left">
          <el-icon><Setting /></el-icon> 编辑企业信息
        </el-button>
        <el-button @click="handleChangePassword" style="width: 100%; text-align: left">
          <el-icon><Lock /></el-icon> 修改密码
        </el-button>
        <el-button type="danger" plain @click="handleLogout" style="width: 100%; text-align: left">
          <el-icon><SwitchButton /></el-icon> 退出登录
        </el-button>
      </el-space>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAuthStore } from '@/stores/auth'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Setting, Lock, SwitchButton } from '@element-plus/icons-vue'
import request from '@/api/request'

const router = useRouter()
const userStore = useUserStore()
const authStore = useAuthStore()
const user = ref(userStore.user)
const org = ref(userStore.org)

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

function handleEditOrg() {
  ElMessage.info('编辑企业信息功能')
}

function handleChangePassword() {
  ElMessage.info('修改密码功能')
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
  padding: 8px;
  padding-bottom: 70px;
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
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
