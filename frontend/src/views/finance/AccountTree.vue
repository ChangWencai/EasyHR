<template>
  <div class="account-tree">
    <!-- 页面标题 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">科目管理</h1>
        <p class="page-subtitle">管理会计科目体系与余额</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" size="large" @click="showCreate = true">
          <el-icon><Plus /></el-icon>
          新增科目
        </el-button>
      </div>
    </header>

    <!-- 数据表格 -->
    <div class="table-card glass-card">
      <el-table
        :data="flatAccounts"
        stripe
        v-loading="loading"
        row-key="id"
        class="modern-table"
        :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
      >
        <el-table-column prop="code" label="科目编码" width="150">
          <template #default="{ row }">
            <span class="account-code">{{ row.code }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="科目名称">
          <template #default="{ row }">
            <div class="account-name">
              <span class="account-icon" :style="{ background: getCategoryGradient(row.category) }">
                <el-icon><Coin /></el-icon>
              </span>
              <span class="account-text">{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="category" label="类别" width="120">
          <template #default="{ row }">
            <span class="category-badge" :class="`category--${getCategoryKey(row.category)}`">
              {{ row.category }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="balance" label="余额" align="right" width="160">
          <template #default="{ row }">
            <span class="balance-value" :class="{ 'balance--negative': Number(row.balance) < 0 }">
              {{ formatAmount(row.balance) }}
            </span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新增科目弹窗 -->
    <el-dialog
      v-model="showCreate"
      title="新增科目"
      width="500px"
      class="create-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <div class="header-icon">
            <el-icon><FolderAdd /></el-icon>
          </div>
          <div class="header-text">
            <span class="header-title">新增科目</span>
            <span class="header-subtitle">设置科目编码、名称与类别</span>
          </div>
        </div>
      </template>

      <el-form :model="form" label-position="top" class="create-form" ref="formRef">
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">科目编码</label>
            <el-input
              v-model="form.code"
              placeholder="如: 1001"
              size="large"
            >
              <template #prefix>
                <el-icon><Key /></el-icon>
              </template>
            </el-input>
          </div>
          <div class="form-group">
            <label class="form-label">科目名称</label>
            <el-input
              v-model="form.name"
              placeholder="如: 库存现金"
              size="large"
            >
              <template #prefix>
                <el-icon><Edit /></el-icon>
              </template>
            </el-input>
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">科目类别</label>
          <div class="category-selector">
            <label
              v-for="cat in categoryOptions"
              :key="cat.value"
              class="category-option"
              :class="{ selected: form.category === cat.value }"
            >
              <input type="radio" :value="cat.value" v-model="form.category" class="hidden-check" />
              <div class="cat-icon" :style="{ background: cat.gradient }">
                <el-icon><component :is="cat.icon" /></el-icon>
              </div>
              <span class="cat-name">{{ cat.label }}</span>
            </label>
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">上级科目（可选）</label>
          <el-tree-select
            v-model="form.parent_id"
            :data="treeData"
            placeholder="留空则创建一级科目"
            clearable
            check-strictly
            size="large"
            style="width: 100%"
          />
        </div>
      </el-form>

      <template #footer>
        <el-button @click="showCreate = false" size="large">取消</el-button>
        <el-button type="primary" :loading="saving" size="large" class="save-btn" @click="handleCreate">
          <el-icon><Check /></el-icon>
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { financeApi } from '@/api/finance'
import {
  Plus, FolderAdd, Key, Edit, Coin, Check,
} from '@element-plus/icons-vue'

interface AccountNode {
  id: number
  code: string
  name: string
  category: string
  balance?: number
  children?: AccountNode[]
}

interface TreeNode {
  id: number
  label: string
  children?: TreeNode[]
}

const accounts = ref<AccountNode[]>([])
const loading = ref(false)
const showCreate = ref(false)
const saving = ref(false)

const form = ref({ code: '', name: '', category: '', parent_id: null as number | null })

const categoryOptions = [
  { label: '资产', value: '资产', icon: 'Money', gradient: 'linear-gradient(135deg, #DBEAFE, #BFDBFE)' },
  { label: '负债', value: '负债', icon: 'TrendCharts', gradient: 'linear-gradient(135deg, #FEF3C7, #FDE68A)' },
  { label: '权益', value: '权益', icon: 'Trophy', gradient: 'linear-gradient(135deg, #D1FAE5, #A7F3D0)' },
  { label: '成本', value: '成本', icon: 'Coin', gradient: 'linear-gradient(135deg, #EDE9FE, #DDD6FE)' },
  { label: '损益', value: '损益', icon: 'Tickets', gradient: 'linear-gradient(135deg, #FEE2E2, #FECACA)' },
]

const categoryKeyMap: Record<string, string> = {
  资产: 'asset', 负债: 'liability', 权益: 'equity', 成本: 'cost', 损益: 'profit',
}

const categoryGradientMap: Record<string, string> = {
  资产: 'linear-gradient(135deg, #DBEAFE, #BFDBFE)',
  负债: 'linear-gradient(135deg, #FEF3C7, #FDE68A)',
  权益: 'linear-gradient(135deg, #D1FAE5, #A7F3D0)',
  成本: 'linear-gradient(135deg, #EDE9FE, #DDD6FE)',
  损益: 'linear-gradient(135deg, #FEE2E2, #FECACA)',
}

function getCategoryKey(cat: string) { return categoryKeyMap[cat] || 'default' }
function getCategoryGradient(cat: string) { return categoryGradientMap[cat] || '#F3F4F6' }

function formatAmount(val?: number) {
  if (val === undefined || val === null) return '—'
  const abs = Math.abs(val)
  return (val < 0 ? '-' : '') + '¥' + abs.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function flatten(nodes: AccountNode[], result: AccountNode[] = []): AccountNode[] {
  for (const node of nodes) {
    result.push({ ...node })
    if (node.children?.length) flatten(node.children, result)
  }
  return result
}

const flatAccounts = computed(() => flatten(accounts.value))

const treeData = computed<TreeNode[]>(() => {
  function toTreeNode(node: AccountNode): TreeNode {
    return {
      id: node.id,
      label: node.code ? `${node.code} ${node.name}` : node.name,
      children: node.children?.length ? node.children.map(toTreeNode) : undefined,
    }
  }
  return accounts.value.map(toTreeNode)
})

async function loadAccounts() {
  loading.value = true
  try {
    const res = await financeApi.accountTree() as any
    accounts.value = res.data?.data || res.data || []
  } catch { accounts.value = [] }
  finally { loading.value = false }
}

async function handleCreate() {
  if (saving.value) return
  if (!form.value.code || !form.value.name || !form.value.category) {
    ElMessage.warning('请填写完整信息')
    return
  }
  saving.value = true
  try {
    await financeApi.createAccount({
      code: form.value.code,
      name: form.value.name,
      category: form.value.category,
      parent_id: form.value.parent_id || undefined,
    })
    ElMessage.success('科目创建成功')
    showCreate.value = false
    form.value = { code: '', name: '', category: '', parent_id: null }
    loadAccounts()
  } catch (e: unknown) {
    const msg = (e as any)?.response?.data?.error || '创建失败'
    ElMessage.error(msg)
  } finally { saving.value = false }
}

onMounted(() => loadAccounts())
</script>

<style scoped lang="scss">
$success: #10B981;
$warning: #F59E0B;
$error: #EF4444;
$bg-page: #FAFBFC;
$text-primary: #1F2937;
$text-secondary: #6B7280;
$text-muted: #9CA3AF;
$border-color: #E5E7EB;
$radius-sm: 8px;
$radius-md: 12px;
$radius-lg: 16px;
$radius-xl: 24px;

.account-tree { padding: 24px 32px; width: 100%; box-sizing: border-box; background: $bg-page; min-height: 100vh; }

.glass-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: $radius-xl;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 24px;
  .page-title { font-size: 24px; font-weight: 700; color: $text-primary; margin: 0 0 4px; }
  .page-subtitle { font-size: 14px; color: $text-secondary; margin: 0; }
}

.table-card { padding: 0; overflow: hidden; }

:deep(.modern-table) {
  .el-table__header th { padding: 14px 16px; font-size: 13px; }
  .el-table__row { transition: background 0.2s ease; &:hover > td { background: rgba(var(--primary), 0.02) !important; } }
  .el-table__cell { padding: 14px 16px; border-bottom: 1px solid #F3F4F6; }
}

.account-code { font-family: 'SF Mono', Monaco, monospace; font-weight: 600; color: var(--primary); font-size: 13px; }

.account-name { display: flex; align-items: center; gap: 10px; }
.account-icon {
  width: 32px; height: 32px;
  border-radius: $radius-sm;
  display: flex; align-items: center; justify-content: center;
  font-size: 15px; color: #fff;
}
.account-text { font-weight: 500; color: $text-primary; }

.category-badge {
  display: inline-flex;
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 20px;

  &.category--asset     { background: #DBEAFE; color: #3B82F6; }
  &.category--liability  { background: #FEF3C7; color: #D97706; }
  &.category--equity     { background: #D1FAE5; color: #059669; }
  &.category--cost       { background: #EDE9FE; color: var(--primary); }
  &.category--profit     { background: #FEE2E2; color: #DC2626; }
  &.category--default    { background: #F3F4F6; color: #6B7280; }
}

.balance-value { font-family: 'SF Mono', Monaco, monospace; font-weight: 600; font-size: 14px; color: $text-primary; &.balance--negative { color: $error; } }

.dialog-header { display: flex; align-items: center; gap: 12px; }
.header-icon { width: 44px; height: 44px; background: linear-gradient(135deg, var(--primary-light), var(--primary)); border-radius: $radius-md; display: flex; align-items: center; justify-content: center; color: #fff; font-size: 20px; }
.header-text { display: flex; flex-direction: column; gap: 2px; }
.header-title { font-size: 18px; font-weight: 700; color: $text-primary; }
.header-subtitle { font-size: 13px; color: $text-muted; }

.create-form { padding-top: 8px; }
.form-row { display: flex; gap: 16px; margin-bottom: 20px; }
.form-group { flex: 1; margin-bottom: 20px; }

.form-label { display: block; font-size: 13px; font-weight: 500; color: $text-secondary; margin-bottom: 8px; }

.category-selector { display: flex; gap: 8px; flex-wrap: wrap; }

.category-option {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 14px 16px;
  border: 2px solid $border-color;
  border-radius: $radius-md;
  cursor: pointer;
  transition: all 0.2s ease;
  min-width: 70px;

  &.selected {
    border-color: var(--primary);
    background: rgba(var(--primary), 0.04);
  }

  &:hover:not(.selected) { border-color: rgba(var(--primary), 0.4); transform: translateY(-2px); }
  .hidden-check { display: none; }
}

.cat-icon {
  width: 36px; height: 36px;
  border-radius: $radius-sm;
  display: flex; align-items: center; justify-content: center;
  font-size: 18px; color: #fff;
}

.cat-name { font-size: 12px; font-weight: 500; color: $text-secondary; }

.save-btn {
  background: linear-gradient(135deg, var(--primary-light), var(--primary));
  border: none;
  box-shadow: 0 4px 14px rgba(var(--primary), 0.4);
  &:hover { box-shadow: 0 6px 20px rgba(var(--primary), 0.5); }
}

@media (max-width: 768px) {
  .account-tree { padding: 16px; }
  .form-row { flex-direction: column; }
  .category-selector { flex-direction: column; }
  .category-option { flex-direction: row; justify-content: flex-start; }
}
</style>
