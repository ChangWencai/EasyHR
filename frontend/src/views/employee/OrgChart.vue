<template>
  <div class="org-chart-page">
    <div class="page-header">
      <h1 class="page-title">组织架构</h1>
      <div class="header-actions">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索部门、岗位或员工"
          clearable
          class="search-input"
          @input="handleSearchInput"
          @clear="handleSearchClear"
        />
        <el-button type="primary" @click="showCreateDialog()">新建部门</el-button>
      </div>
    </div>

    <div v-loading="loading" class="chart-container">
      <div v-if="error" class="error-state">
        <el-empty description="加载组织架构失败，请刷新页面重试">
          <el-button type="primary" @click="loadTree">重新加载</el-button>
        </el-empty>
      </div>

      <div v-else-if="isEmpty" class="empty-state">
        <el-empty description="暂未设置部门架构，请新建部门" />
      </div>

      <div v-else class="chart-wrapper">
        <v-chart :option="chartOption" autoresize style="height: 600px" />
      </div>
    </div>

    <!-- 新建部门弹窗 -->
    <el-dialog
      v-model="createDialogVisible"
      :title="editingParentId ? '新建子部门' : '新建部门'"
      width="440px"
      destroy-on-close
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="80px"
      >
        <el-form-item label="部门名称" prop="name">
          <el-input v-model="createForm.name" placeholder="请输入部门名称" maxlength="100" />
        </el-form-item>
        <el-form-item label="上级部门">
          <el-select
            v-model="createForm.parent_id"
            clearable
            placeholder="无（顶级部门）"
            class="full-width"
          >
            <el-option
              v-for="dept in flatDepartments"
              :key="dept.id"
              :label="dept.name"
              :value="dept.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="createForm.sort_order" :min="0" :max="9999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleCreate">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { TreeChart } from 'echarts/charts'
import { TooltipComponent, TitleComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { departmentApi } from '@/api/department'
import type { Department, TreeNode } from '@/api/department'

use([TreeChart, TooltipComponent, TitleComponent, CanvasRenderer])

const loading = ref(false)
const error = ref(false)
const treeData = ref<TreeNode[]>([])
const flatDepartments = ref<Department[]>([])
const searchKeyword = ref('')
const createDialogVisible = ref(false)
const submitting = ref(false)
const editingParentId = ref<number | null>(null)
const createFormRef = ref<FormInstance>()

const createForm = ref({
  name: '',
  parent_id: null as number | null,
  sort_order: 0,
})

const createRules: FormRules = {
  name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }],
}

let searchTimer: ReturnType<typeof setTimeout> | null = null

const isEmpty = computed(() => treeData.value.length === 0)

const chartOption = computed(() => ({
  tooltip: {
    trigger: 'item',
    triggerOn: 'mousemove',
    formatter: (params: { data?: TreeNode }) => {
      if (!params.data) return ''
      const typeMap: Record<string, string> = {
        department: '部门',
        position: '岗位',
        employee: '员工',
      }
      return `${params.data.name} (${typeMap[params.data.type] ?? ''})`
    },
  },
  series: [
    {
      type: 'tree',
      data: treeData.value,
      layout: 'orthogonal',
      orient: 'LR',
      roam: true,
      animation: false,
      initialTreeDepth: -1,
      label: {
        fontSize: 14,
        fontWeight: 400,
        color: '#1A1A1A',
        position: 'left',
        verticalAlign: 'middle',
        align: 'right',
      },
      leaves: {
        label: {
          position: 'right',
          align: 'left',
        },
      },
      lineStyle: {
        color: '#D9D9D9',
        width: 1,
      },
      itemStyle: {
        color: '#4F6EF7',
        borderColor: '#4F6EF7',
      },
      emphasis: {
        focus: 'ancestor',
      },
    },
  ],
}))

async function loadTree() {
  loading.value = true
  error.value = false
  try {
    const res = await departmentApi.getTree()
    treeData.value = (res as { data?: TreeNode[] }).data ?? (res as unknown as TreeNode[])
  } catch {
    error.value = true
    ElMessage.error('加载组织架构失败，请刷新页面重试')
  } finally {
    loading.value = false
  }
}

async function loadDepartments() {
  try {
    const res = await departmentApi.list()
    flatDepartments.value = (res as { data?: Department[] }).data ?? (res as unknown as Department[])
  } catch {
    // Silently fail - departments list is auxiliary
  }
}

function handleSearchInput() {
  if (searchTimer) {
    clearTimeout(searchTimer)
  }
  if (!searchKeyword.value.trim()) {
    loadTree()
    return
  }
  searchTimer = setTimeout(() => {
    doSearch(searchKeyword.value.trim())
  }, 300)
}

function handleSearchClear() {
  loadTree()
}

async function doSearch(keyword: string) {
  loading.value = true
  try {
    const res = await departmentApi.searchTree(keyword)
    treeData.value = (res as { data?: TreeNode[] }).data ?? (res as unknown as TreeNode[])
  } catch {
    ElMessage.error('搜索失败')
  } finally {
    loading.value = false
  }
}

function showCreateDialog(parentId: number | null = null) {
  editingParentId.value = parentId
  createForm.value = {
    name: '',
    parent_id: parentId,
    sort_order: 0,
  }
  createDialogVisible.value = true
}

async function handleCreate() {
  if (!createFormRef.value) return
  const valid = await createFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    await departmentApi.create({
      name: createForm.value.name,
      parent_id: createForm.value.parent_id,
      sort_order: createForm.value.sort_order,
    })
    ElMessage.success('部门创建成功')
    createDialogVisible.value = false
    loadTree()
    loadDepartments()
  } catch {
    ElMessage.error('创建部门失败')
  } finally {
    submitting.value = false
  }
}

watch(searchKeyword, (val) => {
  if (!val) {
    loadTree()
  }
})

onMounted(() => {
  loadTree()
  loadDepartments()
})
</script>

<style scoped lang="scss">
.org-chart-page {
  padding: 20px 24px;
  width: 100%;
  box-sizing: border-box;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-title {
  font-size: 16px;
  font-weight: 700;
  color: #1a1a1a;
  margin: 0;
  line-height: 1.2;
}

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.search-input {
  width: 260px;
}

.chart-container {
  min-height: 200px;
}

.chart-wrapper {
  width: 100%;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  overflow: hidden;
}

.error-state,
.empty-state {
  padding: 40px 0;
}

.full-width {
  width: 100%;
}

@media (max-width: 768px) {
  .org-chart-page {
    padding: 12px;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .header-actions {
    width: 100%;
    flex-direction: column;
  }

  .search-input {
    width: 100%;
  }
}
</style>
