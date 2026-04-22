<template>
  <div class="position-manage-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <el-button text @click="$router.back()">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1 class="page-title">岗位管理</h1>
      </div>
      <el-button type="primary" @click="showCreateDialog()">
        <el-icon><Plus /></el-icon>
        新建岗位
      </el-button>
    </div>

    <!-- 岗位列表 -->
    <div v-loading="loading" class="content-card">
      <el-table
        v-if="positions.length > 0"
        :data="positions"
        row-key="id"
        stripe
      >
        <el-table-column prop="name" label="岗位名称" min-width="150" />
        <el-table-column label="所属部门" min-width="150">
          <template #default="{ row }">
            <span v-if="row.department_id && deptMap[row.department_id]">
              {{ deptMap[row.department_id] }}
            </span>
            <el-tag v-else type="info" size="small">通用岗位</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sort_order" label="排序" width="100" align="center" />
        <el-table-column label="操作" width="150" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="showEditDialog(row)">编辑</el-button>
            <el-popconfirm
              :title="`确定删除岗位「${row.name}」？`"
              :desc="row.employee_count > 0 ? `该岗位下有 ${row.employee_count} 名员工，无法删除` : ''"
              :confirm-button-text="row.employee_count > 0 ? '' : '删除'"
              :cancel-button-text="'取消'"
              :disabled="row.employee_count > 0"
              @confirm="handleDelete(row.id)"
            >
              <template #reference>
                <el-button type="danger" link size="small" :disabled="row.employee_count > 0">
                  删除
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-else description="暂无岗位，点击上方按钮新建" />
    </div>

    <!-- 新建/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="editingId ? '编辑岗位' : '新建岗位'"
      width="440px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="岗位名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入岗位名称" maxlength="100" />
        </el-form-item>
        <el-form-item label="所属部门">
          <el-select v-model="form.department_id" clearable placeholder="通用岗位（所有部门可用）" class="full-width">
            <el-option
              v-for="dept in flatDepartments"
              :key="dept.id"
              :label="dept.name"
              :value="dept.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort_order" :min="0" :max="9999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, ArrowLeft } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import { positionApi, type Position } from '@/api/position'
import { departmentApi, type Department } from '@/api/department'
import request from '@/api/request'

const loading = ref(false)
const submitting = ref(false)
const positions = ref<(Position & { employee_count?: number })[]>([])
const flatDepartments = ref<Department[]>([])
const deptMap = ref<Record<number, string>>({})

const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()
const form = reactive({
  name: '',
  department_id: null as number | null,
  sort_order: 0,
})
const rules: FormRules = {
  name: [{ required: true, message: '请输入岗位名称', trigger: 'blur' }],
}

onMounted(async () => {
  await Promise.all([loadPositions(), loadDepartments()])
  await loadEmployeeCounts()
})

async function loadPositions() {
  loading.value = true
  try {
    positions.value = await positionApi.list()
  } catch {
    ElMessage.error('加载岗位列表失败')
  } finally {
    loading.value = false
  }
}

async function loadDepartments() {
  try {
    const data = await departmentApi.list()
    flatDepartments.value = Array.isArray(data) ? data : []
    // 构建部门 ID -> 名称映射
    deptMap.value = {}
    for (const d of flatDepartments.value) {
      deptMap.value[d.id] = d.name
    }
  } catch {
    // 静默失败
  }
}

async function loadEmployeeCounts() {
  // 获取每个岗位的员工数量
  try {
    const res = await request.get<Array<{ position_id: number; count: number }>>('/employees/position-counts')
    const data = (res as unknown as { data?: Array<{ position_id: number; count: number }> }).data ?? []
    const countMap: Record<number, number> = {}
    for (const item of data) {
      countMap[item.position_id] = item.count
    }
    for (const pos of positions.value) {
      pos.employee_count = countMap[pos.id] ?? 0
    }
  } catch {
    // 静默失败
  }
}

function showCreateDialog() {
  editingId.value = null
  form.name = ''
  form.department_id = null
  form.sort_order = 0
  dialogVisible.value = true
}

function showEditDialog(row: Position) {
  editingId.value = row.id
  form.name = row.name
  form.department_id = row.department_id
  form.sort_order = row.sort_order
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (editingId.value) {
      await positionApi.update(editingId.value, {
        name: form.name,
        department_id: form.department_id,
        sort_order: form.sort_order,
      })
      ElMessage.success('岗位已更新')
    } else {
      await positionApi.create({
        name: form.name,
        department_id: form.department_id,
        sort_order: form.sort_order,
      })
      ElMessage.success('岗位已创建')
    }
    dialogVisible.value = false
    await loadPositions()
    await loadEmployeeCounts()
  } catch (err: unknown) {
    const msg = (err as { message?: string })?.message ?? '操作失败'
    ElMessage.error(msg)
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await positionApi.delete(id)
    ElMessage.success('岗位已删除')
    await loadPositions()
  } catch (err: unknown) {
    const msg = (err as { message?: string })?.message ?? '删除失败'
    ElMessage.error(msg)
  }
}
</script>

<style scoped lang="scss">
.position-manage-page {
  padding: 20px 24px;
  width: 100%;
  box-sizing: border-box;
  max-width: 900px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-title {
  font-size: 16px;
  font-weight: 700;
  color: #1a1a1a;
  margin: 0;
}

.content-card {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  border: 1px solid #f0f0f0;
}

.full-width {
  width: 100%;
}
</style>
