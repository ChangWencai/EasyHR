<template>
  <div class="salary-slip-h5">
    <!-- 加载中 -->
    <div v-if="loading" class="loading-state">
      <el-icon class="is-loading"><Loading /></el-icon>
      <p>加载中...</p>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="errorMsg" class="error-state">
      <div class="error-icon">
        <el-icon><WarningFilled /></el-icon>
      </div>
      <h3>{{ errorMsg }}</h3>
      <p v-if="errorMsg === '工资单已过期'">请联系 HR 重新发送工资条</p>
      <p v-else-if="errorMsg === '工资单不存在'">工资条链接无效或已过期</p>
      <p v-else>{{ errorMsg }}</p>
    </div>

    <!-- 工资条详情 -->
    <div v-else-if="slip" class="slip-content">
      <!-- 头部 -->
      <div class="slip-header">
        <div class="logo-text">EasyHR</div>
        <div class="slip-month">{{ slip.year }}年{{ slip.month }}月 工资条</div>
      </div>

      <!-- 员工信息 -->
      <div class="employee-card">
        <div class="employee-name">{{ slip.employee_name }}</div>
        <div class="slip-status">
          <el-tag size="small" :type="slipStatusTagType[slip.status]">
            {{ slipStatusMap[slip.status] }}
          </el-tag>
        </div>
      </div>

      <!-- 应发明细 -->
      <div class="slip-section">
        <div class="section-title">应发明细</div>
        <div class="section-items">
          <div
            v-for="item in incomeItems"
            :key="item.item_name"
            class="slip-item"
          >
            <span class="item-name">{{ item.item_name }}</span>
            <span class="item-amount">+{{ item.amount.toFixed(2) }}</span>
          </div>
        </div>
        <div class="section-total">
          <span>应发合计</span>
          <span class="total-value">¥{{ slip.gross_income.toFixed(2) }}</span>
        </div>
      </div>

      <!-- 扣除明细 -->
      <div class="slip-section">
        <div class="section-title">扣除明细</div>
        <div class="section-items">
          <div
            v-for="item in deductionItems"
            :key="item.item_name"
            class="slip-item"
          >
            <span class="item-name">{{ item.item_name }}</span>
            <span class="item-amount deduction">-{{ item.amount.toFixed(2) }}</span>
          </div>
        </div>
        <div class="section-total">
          <span>扣除合计</span>
          <span class="total-value deduction">¥{{ slip.total_deductions.toFixed(2) }}</span>
        </div>
      </div>

      <!-- 实发工资 -->
      <div class="net-income-card">
        <div class="net-income-label">实发工资</div>
        <div class="net-income-value">¥{{ slip.net_income.toFixed(2) }}</div>
      </div>

      <!-- 签收状态 -->
      <div v-if="slip.signed_at" class="sign-status">
        <el-icon><CircleCheckFilled /></el-icon>
        已于 {{ slip.signed_at }} 签收确认
      </div>

      <!-- 签收按钮 -->
      <div v-else-if="slip.status === 'viewed'" class="sign-action">
        <el-button type="primary" size="large" :loading="signing" @click="handleSign">
          确认签收
        </el-button>
      </div>

      <!-- Footer -->
      <div class="slip-footer">
        <span>数据来源: 易人事（EasyHR）</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Loading, WarningFilled, CircleCheckFilled } from '@element-plus/icons-vue'
import request from '@/api/request'

const route = useRoute()
const token = route.params.token as string

const loading = ref(true)
const errorMsg = ref('')
const signing = ref(false)

interface SlipDetail {
  employee_name: string
  year: number
  month: number
  items: { item_name: string; item_type: string; amount: number }[]
  gross_income: number
  si_deduction: number
  tax: number
  total_deductions: number
  net_income: number
  status: string
  signed_at?: string
}

const slip = ref<SlipDetail | null>(null)

const slipStatusMap: Record<string, string> = {
  sent: '待查看',
  viewed: '已查看',
  signed: '已签收',
}
const slipStatusTagType: Record<string, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
  sent: 'info',
  viewed: 'warning',
  signed: 'success',
}

const incomeItems = computed(() =>
  slip.value?.items.filter((i) => i.item_type === 'income') ?? [],
)

const deductionItems = computed(() =>
  slip.value?.items.filter((i) => i.item_type === 'deduction') ?? [],
)

async function loadSlip() {
  loading.value = true
  errorMsg.value = ''
  try {
    const data = await request.get<SlipDetail>(`/salary/slip/${token}`)
    slip.value = data as unknown as SlipDetail
  } catch (e: any) {
    const status = e?.response?.status
    if (status === 404) {
      errorMsg.value = '工资单不存在'
    } else if (status === 403) {
      errorMsg.value = '工资单已过期'
    } else {
      errorMsg.value = '工资条链接无效或已过期'
    }
  } finally {
    loading.value = false
  }
}

async function handleSign() {
  signing.value = true
  try {
    await request.post(`/salary/slip/${token}/sign`)
    if (slip.value) {
      slip.value.status = 'signed'
      const now = new Date()
      slip.value.signed_at = now.toLocaleString('zh-CN')
    }
    ElMessage.success('签收成功')
  } catch {
    ElMessage.error('签收失败')
  } finally {
    signing.value = false
  }
}

onMounted(() => {
  loadSlip()
})
</script>

<style scoped lang="scss">
.salary-slip-h5 {
  min-height: 100vh;
  background: #f5f6f8;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'PingFang SC', 'Helvetica Neue', Helvetica, Arial, sans-serif;
}

/* Loading / Error */
.loading-state,
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  color: #606266;
  .error-icon {
    font-size: 48px;
    color: #f56c6c;
    margin-bottom: 16px;
  }
  h3 {
    margin: 0 0 8px;
    font-size: 16px;
    color: #303133;
  }
  p {
    margin: 0;
    font-size: 13px;
  }
}

/* Content */
.slip-content {
  max-width: 480px;
  margin: 0 auto;
  padding: 0 0 40px;
}

.slip-header {
  background: linear-gradient(135deg, #4f6ef7 0%, #6b8cff 100%);
  color: #fff;
  padding: 28px 20px 24px;
  text-align: center;
  .logo-text {
    font-size: 18px;
    font-weight: 700;
    letter-spacing: 1px;
    margin-bottom: 8px;
  }
  .slip-month {
    font-size: 15px;
    opacity: 0.9;
  }
}

.employee-card {
  background: #fff;
  margin: -12px 16px 12px;
  padding: 16px 20px;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  display: flex;
  align-items: center;
  justify-content: space-between;
  .employee-name {
    font-size: 17px;
    font-weight: 600;
    color: #303133;
  }
}

.slip-section {
  background: #fff;
  margin: 0 16px 12px;
  border-radius: 12px;
  padding: 16px 20px;
  box-shadow: 0 1px 6px rgba(0, 0, 0, 0.05);
  .section-title {
    font-size: 13px;
    color: #909399;
    margin-bottom: 12px;
    font-weight: 500;
  }
  .section-items {
    .slip-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 7px 0;
      border-bottom: 1px solid #f0f2f5;
      &:last-child {
        border-bottom: none;
      }
      .item-name {
        font-size: 14px;
        color: #606266;
      }
      .item-amount {
        font-size: 14px;
        color: #303133;
        font-weight: 500;
        &.deduction {
          color: #f56c6c;
        }
      }
    }
  }
  .section-total {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-top: 10px;
    margin-top: 4px;
    border-top: 1px solid #ebeef5;
    font-size: 14px;
    color: #303133;
    font-weight: 600;
    .total-value {
      font-size: 15px;
      &.deduction {
        color: #f56c6c;
      }
    }
  }
}

.net-income-card {
  background: #fafafa;
  margin: 0 16px 16px;
  border-radius: 12px;
  padding: 20px;
  text-align: center;
  border: 2px solid #4f6ef7;
  .net-income-label {
    font-size: 13px;
    color: #909399;
    margin-bottom: 6px;
  }
  .net-income-value {
    font-size: 28px;
    font-weight: 700;
    color: #4f6ef7;
    letter-spacing: -0.5px;
  }
}

.sign-status {
  margin: 0 16px 24px;
  text-align: center;
  color: #67c23a;
  font-size: 13px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.sign-action {
  margin: 0 16px 24px;
  .el-button {
    width: 100%;
    height: 44px;
    font-size: 16px;
    border-radius: 22px;
  }
}

.slip-footer {
  text-align: center;
  font-size: 11px;
  color: #c0c4cc;
  margin-top: 24px;
}
</style>
