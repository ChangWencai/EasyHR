<template>
  <div class="account-tree">
    <div class="toolbar">
      <el-button type="primary" @click="showCreate = true">新增科目</el-button>
    </div>

    <el-table :data="flatAccounts" stripe v-loading="loading" row-key="id" class="mt-2">
      <el-table-column prop="code" label="科目编码" width="130" />
      <el-table-column prop="name" label="科目名称" />
      <el-table-column prop="category" label="类别" width="100">
        <template #default="{ row }">
          <el-tag :type="categoryType(row.category)" size="small">{{ row.category }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="balance" label="余额" align="right" width="130">
        <template #default="{ row }">{{ formatAmount(row.balance) }}</template>
      </el-table-column>
    </el-table>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreate" title="新增科目" width="500px">
      <el-form :model="form" label-width="90px" ref="formRef">
        <el-form-item label="科目编码" prop="code" required>
          <el-input v-model="form.code" placeholder="如: 1001" />
        </el-form-item>
        <el-form-item label="科目名称" prop="name" required>
          <el-input v-model="form.name" placeholder="如: 库存现金" />
        </el-form-item>
        <el-form-item label="科目类别" prop="category" required>
          <el-select v-model="form.category" placeholder="请选择类别">
            <el-option label="资产" value="资产" />
            <el-option label="负债" value="负债" />
            <el-option label="权益" value="权益" />
            <el-option label="成本" value="成本" />
            <el-option label="损益" value="损益" />
          </el-select>
        </el-form-item>
        <el-form-item label="上级科目">
          <el-tree-select
            v-model="form.parent_id"
            :data="treeData"
            :props="{ label: 'label', value: 'id', children: 'children' }"
            placeholder="可留空"
            clearable
            check-strictly
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleCreate">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { financeApi } from '@/api/finance'

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
const formRef = ref()

const form = ref({
  code: '',
  name: '',
  category: '',
  parent_id: null as number | null,
})

function categoryType(cat: string) {
  const map: Record<string, string> = {
    资产: '',
    负债: 'warning',
    权益: 'success',
    成本: 'info',
    损益: 'danger',
  }
  return map[cat] || 'info'
}

function formatAmount(val?: number) {
  if (val === undefined || val === null) return '-'
  return val.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function flatten(nodes: AccountNode[], result: AccountNode[] = []): AccountNode[] {
  for (const node of nodes) {
    result.push({ ...node })
    if (node.children?.length) {
      flatten(node.children, result)
    }
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
  } catch {
    accounts.value = []
  } finally {
    loading.value = false
  }
}

async function handleCreate() {
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
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadAccounts()
})
</script>

<style scoped lang="scss">
.account-tree {
  padding: 8px;
  .toolbar {
    display: flex;
    gap: 8px;
  }
  .mt-2 {
    margin-top: 12px;
  }
}
</style>
