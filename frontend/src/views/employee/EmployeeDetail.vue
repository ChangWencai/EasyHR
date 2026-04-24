<template>
  <div class="employee-detail">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">员工详情</h1>
        <p class="page-subtitle">查看员工完整信息</p>
      </div>
      <div class="header-actions">
        <el-button @click="$router.back()" size="large" class="back-btn">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
      </div>
    </header>

    <div v-loading="loading" class="detail-content">
      <div v-if="employee" class="profile-card glass-card">
        <!-- 头像区 -->
        <div class="profile-header">
          <div class="avatar-wrap">
            <el-avatar :size="80" class="profile-avatar">{{ employee.name?.[0] }}</el-avatar>
            <div class="avatar-glow"></div>
          </div>
          <div class="profile-info">
            <h2 class="profile-name">{{ employee.name }}</h2>
            <div class="profile-tags">
              <span class="status-badge" :class="`status--${employee.status}`">
                {{ statusMap[employee.status] }}
              </span>
              <span class="position-chip">
                <el-icon><Briefcase /></el-icon>
                {{ employee.position || '未分配岗位' }}
              </span>
            </div>
          </div>
        </div>

        <!-- 信息区 -->
        <div class="info-sections">
          <div class="info-section">
            <div class="section-title">
              <el-icon><User /></el-icon>
              基本信息
            </div>
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">姓名</span>
                <span class="info-value">{{ employee.name }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">性别</span>
                <span class="info-value">{{ employee.gender || '-' }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">手机号</span>
                <span class="info-value info-value--mono">{{ employee.phone }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">邮箱</span>
                <span class="info-value">{{ employee.email || '-' }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">部门</span>
                <span class="info-value">{{ employee.department_name || '-' }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">身份证号</span>
                <span class="info-value info-value--mono">{{ employee.id_card }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">入职日期</span>
                <span class="info-value">
                  <el-icon><Calendar /></el-icon>
                  {{ formatDate(employee.hire_date) }}
                </span>
              </div>
            </div>
          </div>

          <div class="info-section">
            <div class="section-title">
              <el-icon><Money /></el-icon>
              薪资信息
            </div>
            <div class="info-grid">
              <div class="info-item" v-if="employee.salary">
                <span class="info-label">薪资</span>
                <span class="info-value info-value--money">¥{{ employee.salary }}</span>
              </div>
              <div class="info-item" v-if="employee.probation_salary">
                <span class="info-label">试用期薪资</span>
                <span class="info-value info-value--warning">¥{{ employee.probation_salary }}</span>
              </div>
            </div>
          </div>

          <div class="info-section" v-if="employee.bank_account || employee.bank_name">
            <div class="section-title">
              <el-icon><Wallet /></el-icon>
              工资卡信息
            </div>
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">卡号</span>
                <span class="info-value info-value--mono">{{ employee.bank_account || '-' }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">开户行</span>
                <span class="info-value">{{ employee.bank_name || '-' }}</span>
              </div>
            </div>
          </div>

          <div class="info-section" v-if="employee.address || employee.remark">
            <div class="section-title">
              <el-icon><Location /></el-icon>
              其他信息
            </div>
            <div class="info-grid">
              <div class="info-item" v-if="employee.address">
                <span class="info-label">居住地址</span>
                <span class="info-value">{{ employee.address }}</span>
              </div>
              <div class="info-item" v-if="employee.remark">
                <span class="info-label">备注</span>
                <span class="info-value">{{ employee.remark }}</span>
              </div>
            </div>
          </div>

          <div class="info-section" v-if="employee.emergency_contact">
            <div class="section-title">
              <el-icon><Phone /></el-icon>
              紧急联系人
            </div>
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">姓名</span>
                <span class="info-value">{{ employee.emergency_contact }}</span>
              </div>
              <div class="info-item" v-if="employee.emergency_phone">
                <span class="info-label">电话</span>
                <span class="info-value info-value--mono">{{ employee.emergency_phone }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 加载失败 -->
      <div v-else class="empty-state glass-card">
        <div class="empty-icon"><el-icon><UserFilled /></el-icon></div>
        <h3>加载失败</h3>
        <p>无法获取员工信息</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { employeeApi } from '@/api/employee'
import { ElMessage } from 'element-plus'
import {
  ArrowLeft, Briefcase, User, Money, Wallet, Phone, Calendar, UserFilled, Location,
} from '@element-plus/icons-vue'
import { statusMap } from './statusMap'

const route = useRoute()
const loading = ref(false)
const employee = ref<any>(null)

function formatDate(dateStr: string | undefined): string {
  if (!dateStr) return '-'
  const s = String(dateStr)
  if (s.includes('T')) {
    return s.split('T')[0]
  }
  return s
}

async function load() {
  loading.value = true
  try {
    employee.value = await employeeApi.get(Number(route.params.id))
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => load())
</script>

<style scoped lang="scss">
$bg-page: #FAFBFC; $success: #10B981; $warning: #F59E0B; $error: #EF4444;
$text-primary: #1F2937; $text-secondary: #6B7280; $text-muted: #9CA3AF; $border-color: #E5E7EB;
$radius-sm: 8px; $radius-md: 12px; $radius-lg: 16px; $radius-xl: 24px;

.employee-detail { padding: 24px 32px; width: 100%; box-sizing: border-box; background: $bg-page; min-height: 100vh; }

.glass-card { background: rgba(255,255,255,0.95); backdrop-filter: blur(12px); border: 1px solid rgba(255,255,255,0.6); border-radius: $radius-xl; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); }

.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 24px;
  .page-title { font-size: 24px; font-weight: 700; color: $text-primary; margin: 0 0 4px; }
  .page-subtitle { font-size: 14px; color: $text-secondary; margin: 0; }
}

.back-btn { border-radius: $radius-md; }

.profile-card { padding: 32px; }

.profile-header { display: flex; align-items: center; gap: 24px; margin-bottom: 32px; }

.avatar-wrap { position: relative; flex-shrink: 0; }
.profile-avatar { background: linear-gradient(135deg, var(--primary-light), var(--primary)); color: #fff; font-size: 28px; font-weight: 700; }
.avatar-glow { position: absolute; inset: -4px; background: linear-gradient(135deg, var(--primary-light), var(--primary)); border-radius: 50%; opacity: 0.2; filter: blur(8px); z-index: -1; }

.profile-info { display: flex; flex-direction: column; gap: 10px; }
.profile-name { font-size: 24px; font-weight: 700; color: $text-primary; margin: 0; }
.profile-tags { display: flex; align-items: center; gap: 8px; }

.status-badge { display: inline-flex; padding: 4px 14px; font-size: 12px; font-weight: 600; border-radius: 20px;
  &.status--正式{background:#D1FAE5;color:#059669} &.status--试用{background:#FEF3C7;color:#D97706} &.status--待入职{background:#EDE9FE;color:var(--primary)} &.status--离职{background:#F3F4F6;color:#6B7280}
}

.position-chip { display: inline-flex; align-items: center; gap: 4px; padding: 4px 12px; background: #F3F4F6; color: $text-secondary; font-size: 13px; font-weight: 500; border-radius: 20px; .el-icon{font-size:14px} }

.info-sections { display: flex; flex-direction: column; gap: 24px; }

.info-section {
  border: 1px solid $border-color; border-radius: $radius-lg; overflow: hidden;
}

.section-title { display: flex; align-items: center; gap: 8px; padding: 12px 20px; background: #F9FAFB; font-size: 14px; font-weight: 600; color: $text-primary; border-bottom: 1px solid $border-color; .el-icon{color:var(--primary);font-size:16px} }

.info-grid { display: grid; grid-template-columns: repeat(2,1fr); gap: 0; }

.info-item { display: flex; flex-direction: column; gap: 4px; padding: 16px 20px; border-bottom: 1px solid $border-color; border-right: 1px solid $border-color;
  &:nth-child(2n){border-right:none} &:nth-last-child(-n+2){border-bottom:none}
}

.info-label { font-size: 12px; color: $text-muted; }
.info-value { font-size: 15px; font-weight: 500; color: $text-primary; display: flex; align-items: center; gap: 6px; .el-icon{color:$text-muted;font-size:14px}
  &--mono{font-family:'SF Mono',Monaco,monospace}
  &--money{color:var(--primary);font-weight:700;font-size:18px}
  &--warning{color:$warning;font-weight:600}
}

.empty-state { text-align:center; padding:80px 32px;
  .empty-icon{width:72px;height:72px;margin:0 auto 16px;background:linear-gradient(135deg,#EDE9FE,#DDD6FE);border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:32px;color:var(--primary)}
  h3{font-size:18px;font-weight:600;color:$text-primary;margin:0 0 8px}
  p{font-size:14px;color:$text-muted;margin:0}
}

@media (max-width:768px){
  .employee-detail{padding:16px}
  .info-grid{grid-template-columns:1fr}
  .info-item{border-right:none;border-bottom:1px solid $border-color}
  .info-item:last-child{border-bottom:none}
}
</style>
