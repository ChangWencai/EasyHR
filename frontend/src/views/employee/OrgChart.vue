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

      <div v-else class="chart-wrapper" @click="closeContextMenu">
        <v-chart ref="chartRef" :option="chartOption" autoresize style="height: 600px" @contextmenu.prevent />
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

    <!-- Right-click context menu -->
    <div
      v-if="contextMenuVisible"
      class="context-menu"
      :style="{ left: contextMenuX + 'px', top: contextMenuY + 'px' }"
      @click.stop
    >
      <div class="context-menu-item" @click="showMoveDialog">
        <el-icon><Rank /></el-icon> 移动到...
      </div>
      <div class="context-menu-item" @click="showDeleteTransferDialog">
        <el-icon><Delete /></el-icon> 删除部门
      </div>
    </div>

    <!-- Inline edit overlay -->
    <div
      v-if="inlineEditVisible"
      class="inline-edit-overlay"
      :style="{ left: inlineEditX + 'px', top: inlineEditY + 'px' }"
    >
      <el-input
        v-model="inlineEditValue"
        size="small"
        class="inline-edit-input"
        @keyup.enter="handleInlineEditSave"
        @blur="handleInlineEditSave"
        @keyup.esc="inlineEditVisible = false"
      />
    </div>

    <!-- Delete transfer dialog -->
    <el-dialog
      v-model="deleteTransferVisible"
      :title="`删除部门「${deleteTransferDeptName}」`"
      width="480px"
      destroy-on-close
    >
      <el-alert type="warning" :closable="false">
        该部门下有 {{ deleteTransferEmployees.length }} 名员工，请选择接收部门后，再确认删除
      </el-alert>
      <el-table :data="deleteTransferEmployees" max-height="240" style="margin: 16px 0">
        <el-table-column prop="name" label="姓名" />
        <el-table-column prop="position" label="岗位" />
      </el-table>
      <el-form-item label="接收部门" required>
        <el-select v-model="deleteTransferTarget" placeholder="请选择接收部门" class="full-width">
          <el-option
            v-for="dept in availableTargetDepts"
            :key="dept.id"
            :label="dept.name"
            :value="dept.id"
          />
        </el-select>
      </el-form-item>
      <template #footer>
        <el-button @click="deleteTransferVisible = false">取消</el-button>
        <el-button
          type="danger"
          :loading="deleteTransferLoading"
          :disabled="!deleteTransferTarget"
          @click="handleDeleteTransfer"
        >
          确认转移并删除
        </el-button>
      </template>
    </el-dialog>

    <!-- Move to dialog -->
    <el-dialog
      v-model="moveDialogVisible"
      title="移动部门"
      width="440px"
      destroy-on-close
    >
      <p>将「{{ moveDeptName }}」移动到：</p>
      <el-select v-model="moveTargetDeptId" placeholder="请选择目标上级部门" class="full-width">
        <el-option
          v-for="dept in availableMoveTargets"
          :key="dept.id"
          :label="dept.name"
          :value="dept.id"
        />
      </el-select>
      <template #footer>
        <el-button @click="moveDialogVisible = false">取消</el-button>
        <el-button type="primary" :disabled="!moveTargetDeptId" @click="handleMoveDept">确认移动</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Rank, Delete } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { TreeChart } from 'echarts/charts'
import { TooltipComponent, TitleComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { departmentApi } from '@/api/department'
import type { Department, TreeNode } from '@/api/department'
import request from '@/api/request'

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
const chartRef = ref<InstanceType<typeof VChart> | null>(null)

const createForm = ref({
  name: '',
  parent_id: null as number | null,
  sort_order: 0,
})

const createRules: FormRules = {
  name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }],
}

let searchTimer: ReturnType<typeof setTimeout> | null = null

// Context menu state
const contextMenuVisible = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextMenuDeptId = ref<number | null>(null)
const contextMenuDeptName = ref('')

// Inline edit state
const inlineEditVisible = ref(false)
const inlineEditDeptId = ref<number | null>(null)
const inlineEditValue = ref('')
const inlineEditX = ref(0)
const inlineEditY = ref(0)

// Delete transfer dialog state
const deleteTransferVisible = ref(false)
const deleteTransferDeptId = ref<number | null>(null)
const deleteTransferDeptName = ref('')
const deleteTransferEmployees = ref<Array<{ id: number; name: string; position: string }>>([])
const deleteTransferTarget = ref<number | null>(null)
const deleteTransferLoading = ref(false)

// Move to dialog state
const moveDialogVisible = ref(false)
const moveTargetDeptId = ref<number | null>(null)
const moveDeptId = ref<number | null>(null)
const moveDeptName = ref('')

const isEmpty = computed(() => treeData.value.length === 0)

const availableTargetDepts = computed(() =>
  flatDepartments.value.filter(d => d.id !== deleteTransferDeptId.value),
)

const availableMoveTargets = computed(() =>
  flatDepartments.value.filter(d => d.id !== moveDeptId.value),
)

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
      itemStyle: (params: { data?: TreeNode }) => {
        const node = params.data
        if (!node) return { color: '#4F6EF7', borderColor: '#4F6EF7' }
        // Let backend-set colors flow through (search highlights), fall back to type-based color
        if (node.itemStyle?.color) {
          return { color: node.itemStyle.color as string, borderColor: node.itemStyle.color as string }
        }
        const colorMap: Record<string, string> = {
          department: '#7C3AED',
          position: '#A78BFA',
          employee: '#F59E0B',
        }
        const color = colorMap[node.type] ?? '#4F6EF7'
        return { color, borderColor: color }
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
  // Bind ECharts events after tree data loads
  bindChartEvents()
}

function bindChartEvents() {
  const chartInstance = chartRef.value?.chart
  if (!chartInstance) return

  // Context menu on right-click department nodes
  chartInstance.on('contextmenu', (params: unknown) => {
    const p = params as { data?: TreeNode; offsetX?: number; offsetY?: number; event?: MouseEvent }
    if (p.data?.type === 'department' && p.data?.id) {
      p.event?.preventDefault?.()
      contextMenuDeptId.value = p.data.id
      contextMenuDeptName.value = p.data.name
      contextMenuX.value = p.offsetX ?? 0
      contextMenuY.value = p.offsetY ?? 0
      contextMenuVisible.value = true
    }
  })

  // Click on department node for inline edit
  chartInstance.on('click', (params: unknown) => {
    const p = params as { data?: TreeNode; offsetX?: number; offsetY?: number }
    if (p.data?.type === 'department' && p.data?.id) {
      contextMenuVisible.value = false
      inlineEditDeptId.value = p.data.id
      inlineEditValue.value = p.data.name
      inlineEditX.value = p.offsetX ?? 0
      inlineEditY.value = p.offsetY ?? 0
      inlineEditVisible.value = true
    }
  })
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

function closeContextMenu() {
  contextMenuVisible.value = false
}

function showMoveDialog() {
  contextMenuVisible.value = false
  moveDeptId.value = contextMenuDeptId.value
  moveDeptName.value = contextMenuDeptName.value
  moveTargetDeptId.value = null
  moveDialogVisible.value = true
}

async function handleMoveDept() {
  if (!moveDeptId.value || !moveTargetDeptId.value) return
  try {
    await departmentApi.update(moveDeptId.value, { parent_id: moveTargetDeptId.value })
    ElMessage.success('部门已移动')
    moveDialogVisible.value = false
    loadTree()
    loadDepartments()
  } catch {
    ElMessage.error('移动失败')
  }
}

function showDeleteTransferDialog() {
  contextMenuVisible.value = false
  deleteTransferDeptId.value = contextMenuDeptId.value
  deleteTransferDeptName.value = contextMenuDeptName.value
  deleteTransferTarget.value = null
  if (deleteTransferDeptId.value) {
    loadDeptEmployees(deleteTransferDeptId.value)
  }
}

async function loadDeptEmployees(deptId: number) {
  try {
    const res = await request.get<Array<{ id: number; name: string; position: string }>>('/employees', {
      params: { department_id: deptId },
    })
    const data = (res as { data?: Array<{ id: number; name: string; position: string }> }).data
      ?? (res as unknown as Array<{ id: number; name: string; position: string }>)
    deleteTransferEmployees.value = data
    deleteTransferVisible.value = true
  } catch {
    deleteTransferEmployees.value = []
    deleteTransferVisible.value = true
  }
}

async function handleDeleteTransfer() {
  if (!deleteTransferDeptId.value || !deleteTransferTarget.value) return
  deleteTransferLoading.value = true
  try {
    await departmentApi.transferDelete(deleteTransferDeptId.value, {
      target_department_id: deleteTransferTarget.value,
      employee_ids: deleteTransferEmployees.value.map(e => e.id),
    })
    ElMessage.success('部门已删除，员工已转移')
    deleteTransferVisible.value = false
    loadTree()
    loadDepartments()
  } catch {
    ElMessage.error('删除失败')
  } finally {
    deleteTransferLoading.value = false
  }
}

function handleInlineEditSave() {
  if (!inlineEditDeptId.value) return
  const newName = inlineEditValue.value.trim()
  if (!newName) {
    inlineEditVisible.value = false
    return
  }
  departmentApi.update(inlineEditDeptId.value, { name: newName })
    .then(() => {
      ElMessage.success('部门名称已更新')
      inlineEditVisible.value = false
      loadTree()
      loadDepartments()
    })
    .catch(() => {
      ElMessage.error('保存失败，请重试')
      inlineEditVisible.value = false
    })
}

watch(searchKeyword, (val) => {
  if (!val) {
    loadTree()
  }
})

onMounted(() => {
  document.addEventListener('click', closeContextMenu)
  loadTree()
  loadDepartments()
})

onUnmounted(() => {
  document.removeEventListener('click', closeContextMenu)
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

.context-menu {
  position: absolute;
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 4px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  z-index: 9999;
  padding: 4px 0;
}

.context-menu-item {
  padding: 8px 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #333;
}

.context-menu-item:hover {
  background: #f5f5f5;
}

.inline-edit-overlay {
  position: absolute;
  z-index: 9998;
}

.inline-edit-input {
  min-width: 120px;

  :deep(.el-input__wrapper) {
    border-color: #7C3AED !important;
  }
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
