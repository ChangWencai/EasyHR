<template>
  <div class="org-chart-page">
    <!-- 页面头部 -->
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
        <el-button
          :type="viewMode === 'chart' ? 'primary' : 'default'"
          @click="viewMode = 'chart'"
        >
          <el-icon><DataAnalysis /></el-icon>
          架构图
        </el-button>
        <el-button type="primary" @click="showCreateDeptDialog()">新建部门</el-button>
        <el-button @click="showCreatePosDialog()">
          <el-icon><Plus /></el-icon>
          新建岗位
        </el-button>
      </div>
    </div>

    <!-- 列表视图：两列并排 -->
    <div v-if="viewMode === 'list'" class="list-layout">
      <div v-loading="listLoading" class="list-column">
        <!-- 部门列表 -->
        <div class="section-card">
          <div class="section-header">
            <h2 class="section-title">
              <el-icon color="#7C3AED"><OfficeBuilding /></el-icon>
              部门
              <el-tag type="info" size="small" style="margin-left: 8px">{{ flatDepartments.length }}</el-tag>
            </h2>
            <el-button text type="primary" @click="showCreateDeptDialog()">
              <el-icon><Plus /></el-icon> 新建
            </el-button>
          </div>
          <el-table :data="flatDepartments" row-key="id" stripe size="small">
            <el-table-column prop="name" label="部门名称" min-width="120" />
            <el-table-column label="上级部门" min-width="100">
              <template #default="{ row }">
                {{ getParentName(row.parent_id) }}
              </template>
            </el-table-column>
            <el-table-column prop="sort_order" label="排序" width="70" align="center" />
            <el-table-column label="操作" width="120" align="center">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="showEditDeptDialog(row)">编辑</el-button>
                <el-popconfirm
                  :title="`确定删除「${row.name}」？`"
                  :confirm-button-text="row.employee_count > 0 ? '' : '删除'"
                  :disabled="row.employee_count > 0"
                  @confirm="handleDeleteDept(row.id)"
                >
                  <template #reference>
                    <el-button type="danger" link size="small" :disabled="row.employee_count > 0">删除</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="flatDepartments.length === 0" description="暂无部门" :image-size="50" />
        </div>
      </div>

      <div v-loading="listLoading" class="list-column">
        <!-- 岗位列表 -->
        <div class="section-card">
          <div class="section-header">
            <h2 class="section-title">
              <el-icon color="#06B6D4"><Briefcase /></el-icon>
              岗位
              <el-tag type="info" size="small" style="margin-left: 8px">{{ positions.length }}</el-tag>
            </h2>
            <el-button text type="primary" @click="showCreatePosDialog()">
              <el-icon><Plus /></el-icon> 新建
            </el-button>
          </div>
          <el-table :data="positions" row-key="id" stripe size="small">
            <el-table-column prop="name" label="岗位名称" min-width="120" />
            <el-table-column label="所属部门" min-width="100">
              <template #default="{ row }">
                <span v-if="row.department_id && deptMap[row.department_id]">
                  {{ deptMap[row.department_id] }}
                </span>
                <el-tag v-else type="info" size="small">通用</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="sort_order" label="排序" width="70" align="center" />
            <el-table-column label="操作" width="120" align="center">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="showEditPosDialog(row)">编辑</el-button>
                <el-popconfirm
                  :title="`确定删除岗位「${row.name}」？`"
                  :confirm-button-text="row.employee_count > 0 ? '' : '删除'"
                  :disabled="row.employee_count > 0"
                  @confirm="handleDeletePos(row.id)"
                >
                  <template #reference>
                    <el-button type="danger" link size="small" :disabled="row.employee_count > 0">删除</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="positions.length === 0" description="暂无岗位" :image-size="50" />
        </div>
      </div>
    </div>

    <!-- 架构图视图 -->
    <div v-else v-loading="loading" class="chart-container">
      <div class="chart-toolbar">
        <el-button @click="viewMode = 'list'">
          <el-icon><Back /></el-icon>
          返回列表
        </el-button>
      </div>

      <div v-if="error" class="error-state">
        <el-empty description="加载组织架构失败，请刷新页面重试">
          <el-button type="primary" @click="loadTree">重新加载</el-button>
        </el-empty>
      </div>

      <div v-else-if="isEmpty" class="empty-state">
        <el-empty description="暂未设置部门架构，请新建部门" />
      </div>

      <div v-else class="chart-wrapper" @click="closeContextMenu" @contextmenu.prevent>
        <v-chart
          ref="chartRef"
          :option="chartOption"
          autoresize
          style="height: 600px"
        />
      </div>
    </div>

    <!-- 新建/编辑部门弹窗 -->
    <el-dialog
      v-model="createDialogVisible"
      :title="editingDeptId ? '编辑部门' : '新建部门'"
      width="440px"
      destroy-on-close
    >
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="80px">
        <el-form-item label="部门名称" prop="name">
          <el-input v-model="createForm.name" placeholder="请输入部门名称" maxlength="100" />
        </el-form-item>
        <el-form-item label="上级部门">
          <el-select v-model="createForm.parent_id" clearable placeholder="无（顶级部门）" class="full-width">
            <el-option
              v-for="dept in flatDepartments.filter(d => d.id !== editingDeptId)"
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
        <el-button type="primary" :loading="submitting" @click="handleCreateDept">确认</el-button>
      </template>
    </el-dialog>

    <!-- 右键菜单 -->
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

    <!-- 内联编辑浮层 -->
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

    <!-- 删除部门弹窗 -->
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

    <!-- 移动部门弹窗 -->
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

    <!-- 新建/编辑岗位弹窗 -->
    <el-dialog
      v-model="positionDialogVisible"
      :title="editingPosId ? '编辑岗位' : '新建岗位'"
      width="440px"
      destroy-on-close
    >
      <el-form ref="positionFormRef" :model="positionForm" :rules="positionRules" label-width="80px">
        <el-form-item label="岗位名称" prop="name">
          <el-input v-model="positionForm.name" placeholder="请输入岗位名称" maxlength="100" />
        </el-form-item>
        <el-form-item label="所属部门">
          <el-select v-model="positionForm.department_id" clearable placeholder="通用岗位（所有部门可用）" class="full-width">
            <el-option
              v-for="dept in flatDepartments"
              :key="dept.id"
              :label="dept.name"
              :value="dept.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="positionForm.sort_order" :min="0" :max="9999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="positionDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="positionSubmitting" @click="handleSubmitPosition">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Rank, Delete, Plus, Back,
  DataAnalysis, OfficeBuilding, Briefcase,
} from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { TreeChart } from 'echarts/charts'
import { TooltipComponent, TitleComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { departmentApi } from '@/api/department'
import type { Department, TreeNode } from '@/api/department'
import { positionApi, type Position } from '@/api/position'
import request from '@/api/request'

use([TreeChart, TooltipComponent, TitleComponent, CanvasRenderer])

// 视图模式
const viewMode = ref<'list' | 'chart'>('list')

const loading = ref(false)
const listLoading = ref(false)
const error = ref(false)
const treeData = ref<TreeNode[]>([])
const flatDepartments = ref<Department[]>([])
const positions = ref<Position[]>([])
const deptMap = ref<Record<number, string>>({})
const deptEmployeeCounts = ref<Record<number, number>>({})
const searchKeyword = ref('')

const createDialogVisible = ref(false)
const submitting = ref(false)
const editingDeptId = ref<number | null>(null)
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

// 岗位管理
const positionDialogVisible = ref(false)
const positionSubmitting = ref(false)
const editingPosId = ref<number | null>(null)
const positionFormRef = ref<FormInstance>()
const positionForm = ref({
  name: '',
  department_id: null as number | null,
  sort_order: 0,
})
const positionRules: FormRules = {
  name: [{ required: true, message: '请输入岗位名称', trigger: 'blur' }],
}
const posEmployeeCounts = ref<Record<number, number>>({})

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
        if (node.itemStyle?.color) {
          return { color: node.itemStyle.color as string, borderColor: node.itemStyle.color as string }
        }
        const colorMap: Record<string, string> = {
          department: '#7C3AED',
          position: '#06B6D4',
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

// ========== 列表视图数据加载 ==========

async function loadDepartments() {
  try {
    const data = await departmentApi.list()
    flatDepartments.value = Array.isArray(data) ? data : []
    // 构建部门 ID -> 名称映射
    deptMap.value = {}
    for (const d of flatDepartments.value) {
      deptMap.value[d.id] = d.name
    }
    // 加载员工数量
    await loadDeptEmployeeCounts()
  } catch {
    // 静默失败
  }
}

async function loadDeptEmployeeCounts() {
  try {
    const res = await request.get<Array<{ department_id: number; count: number }>>('/employees/dept-counts')
    const data = (res as unknown as { data?: Array<{ department_id: number; count: number }> }).data ?? []
    deptEmployeeCounts.value = {}
    for (const item of data) {
      deptEmployeeCounts.value[item.department_id] = item.count
    }
    for (const dept of flatDepartments.value) {
      dept.employee_count = deptEmployeeCounts.value[dept.id] ?? 0
    }
  } catch {
    // 静默失败
  }
}

async function loadPositions() {
  try {
    positions.value = (await positionApi.list()) ?? []
    await loadPosEmployeeCounts()
  } catch {
    // 静默失败
  }
}

async function loadPosEmployeeCounts() {
  try {
    const res = await request.get<Array<{ position_id: number; count: number }>>('/employees/position-counts')
    const data = (res as unknown as { data?: Array<{ position_id: number; count: number }> }).data ?? []
    posEmployeeCounts.value = {}
    for (const item of data) {
      posEmployeeCounts.value[item.position_id] = item.count
    }
    for (const pos of positions.value ?? []) {
      pos.employee_count = posEmployeeCounts.value[pos.id] ?? 0
    }
  } catch {
    // 静默失败
  }
}

function getParentName(parentId: number | null): string {
  if (!parentId) return '-'
  return deptMap.value[parentId] ?? '-'
}

// ========== 列表视图操作 ==========

function showCreateDeptDialog() {
  editingDeptId.value = null
  createForm.value = { name: '', parent_id: null, sort_order: 0 }
  createDialogVisible.value = true
}

function showEditDeptDialog(dept: Department) {
  editingDeptId.value = dept.id
  createForm.value = { name: dept.name, parent_id: dept.parent_id, sort_order: dept.sort_order }
  createDialogVisible.value = true
}

async function handleCreateDept() {
  if (!createFormRef.value) return
  const valid = await createFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (editingDeptId.value) {
      await departmentApi.update(editingDeptId.value, {
        name: createForm.value.name,
        parent_id: createForm.value.parent_id,
        sort_order: createForm.value.sort_order,
      })
      ElMessage.success('部门已更新')
    } else {
      await departmentApi.create({
        name: createForm.value.name,
        parent_id: createForm.value.parent_id,
        sort_order: createForm.value.sort_order,
      })
      ElMessage.success('部门已创建')
    }
    createDialogVisible.value = false
    await loadDepartments()
  } catch {
    ElMessage.error(editingDeptId.value ? '更新失败' : '创建失败')
  } finally {
    submitting.value = false
  }
}

async function handleDeleteDept(id: number) {
  try {
    await departmentApi.delete(id)
    ElMessage.success('部门已删除')
    await loadDepartments()
  } catch (err: unknown) {
    const msg = (err as { message?: string })?.message ?? '删除失败'
    ElMessage.error(msg)
  }
}

// ========== 岗位管理 ==========

function showCreatePosDialog() {
  editingPosId.value = null
  positionForm.value = { name: '', department_id: null, sort_order: 0 }
  positionDialogVisible.value = true
}

function showEditPosDialog(pos: Position) {
  editingPosId.value = pos.id
  positionForm.value = { name: pos.name, department_id: pos.department_id, sort_order: pos.sort_order }
  positionDialogVisible.value = true
}

async function handleSubmitPosition() {
  if (!positionFormRef.value) return
  const valid = await positionFormRef.value.validate().catch(() => false)
  if (!valid) return

  positionSubmitting.value = true
  try {
    if (editingPosId.value) {
      await positionApi.update(editingPosId.value, {
        name: positionForm.value.name,
        department_id: positionForm.value.department_id,
        sort_order: positionForm.value.sort_order,
      })
      ElMessage.success('岗位已更新')
    } else {
      await positionApi.create({
        name: positionForm.value.name,
        department_id: positionForm.value.department_id,
        sort_order: positionForm.value.sort_order,
      })
      ElMessage.success('岗位已创建')
    }
    positionDialogVisible.value = false
    await loadPositions()
  } catch (err: unknown) {
    const msg = (err as { message?: string })?.message ?? '操作失败'
    ElMessage.error(msg)
  } finally {
    positionSubmitting.value = false
  }
}

async function handleDeletePos(id: number) {
  try {
    await positionApi.delete(id)
    ElMessage.success('岗位已删除')
    await loadPositions()
  } catch (err: unknown) {
    const msg = (err as { message?: string })?.message ?? '删除失败'
    ElMessage.error(msg)
  }
}

// ========== 架构图 ==========

const chartRef = ref<InstanceType<typeof VChart> | null>(null)

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
  nextTick(() => bindChartEvents())
}

function bindChartEvents() {
  const chartInstance = chartRef.value?.chart
  if (!chartInstance) return

  chartInstance.off('dblclick')
  chartInstance.off('contextmenu')

  const getPointerPos = (echartsEvent: { offsetX?: number; offsetY?: number } | undefined) => {
    const chartDom = chartInstance.getDom()
    const rect = chartDom.getBoundingClientRect()
    const offsetX = echartsEvent?.offsetX ?? 0
    const offsetY = echartsEvent?.offsetY ?? 0
    return { x: rect.left + offsetX, y: rect.top + offsetY }
  }

  chartInstance.on('contextmenu', (params: unknown) => {
    const p = params as { data?: TreeNode; event?: { preventDefault?: () => void; offsetX?: number; offsetY?: number } }
    if (p.data?.type === 'department' && p.data?.id) {
      p.event?.preventDefault?.()
      contextMenuDeptId.value = p.data.id
      contextMenuDeptName.value = p.data.name
      const pos = getPointerPos(p.event)
      contextMenuX.value = pos.x
      contextMenuY.value = pos.y
      contextMenuVisible.value = true
    }
  })

  chartInstance.on('dblclick', (params: unknown) => {
    const p = params as { data?: TreeNode; event?: { offsetX?: number; offsetY?: number } }
    if (p.data?.type === 'department' && p.data?.id) {
      contextMenuVisible.value = false
      inlineEditDeptId.value = p.data.id
      inlineEditValue.value = p.data.name
      const pos = getPointerPos(p.event)
      inlineEditX.value = pos.x
      inlineEditY.value = pos.y
      inlineEditVisible.value = true
    }
  })
}

// ========== 搜索 ==========

function handleSearchInput() {
  if (searchTimer) clearTimeout(searchTimer)
  if (!searchKeyword.value.trim()) {
    if (viewMode.value === 'chart') loadTree()
    return
  }
  searchTimer = setTimeout(() => doSearch(searchKeyword.value.trim()), 300)
}

function handleSearchClear() {
  if (viewMode.value === 'chart') loadTree()
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

// ========== 部门操作 ==========

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

// ========== 生命周期 ==========

watch(searchKeyword, (val) => {
  if (!val && viewMode.value === 'chart') {
    loadTree()
  }
})

watch(viewMode, async (mode) => {
  if (mode === 'chart') {
    await loadTree()
    nextTick(() => bindChartEvents())
  }
})

onMounted(async () => {
  document.addEventListener('click', closeContextMenu)
  await Promise.all([loadDepartments(), loadPositions()])
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

// 列表视图：两列并排
.list-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  align-items: start;
}

.list-column {
  min-width: 0;
}

.section-card {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  border: 1px solid #f0f0f0;
  margin-bottom: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 6px;
}

// 架构图视图
.chart-container {
  min-height: 200px;
}

.chart-toolbar {
  margin-bottom: 12px;
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
  position: fixed;
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
  position: fixed;
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

  .list-layout {
    grid-template-columns: 1fr;
  }
}
</style>
