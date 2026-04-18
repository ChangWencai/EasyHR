<template>
  <div class="todo-list-view">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">待办中心</h1>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="handleExport">
          <el-icon><Download /></el-icon>
          导出Excel
        </el-button>
      </div>
    </div>

    <!-- 搜索筛选区域 -->
    <div class="filter-card section">
      <el-form :inline="true" :model="filters">
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px">
            <el-option label="全部" value="" />
            <el-option label="待办" value="pending" />
            <el-option label="已完成" value="completed" />
            <el-option label="已终止" value="terminated" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键字">
          <el-input
            v-model="filters.keyword"
            placeholder="事项名称/发起人/员工"
            clearable
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="时间段">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            :disabled-date="disabledDate"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 待办列表 -->
    <div class="section">
      <div v-if="loading" class="loading">
        <el-icon class="is-loading" size="24"><Loading /></el-icon>
      </div>
      <div v-else-if="items.length === 0" class="empty-state">
        <el-empty description="暂无待办事项" :image-size="80" />
      </div>
      <el-table v-else :data="items" stripe>
        <el-table-column label="序号" width="60" align="center">
          <template #default="{ $index }">
            {{ (pagination.page - 1) * pagination.pageSize + $index + 1 }}
          </template>
        </el-table-column>
        <el-table-column prop="title" label="事项名称" min-width="200" />
        <el-table-column prop="type" label="类型" width="120" />
        <el-table-column prop="employee_name" label="员工姓名" width="100" />
        <el-table-column prop="creator_name" label="发起人" width="100" />
        <el-table-column prop="created_at" label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="deadline" label="截止日期" width="120">
          <template #default="{ row }">
            <span v-if="row.deadline">{{ row.deadline }}</span>
            <span v-else class="text-tertiary">--</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.status === 'completed'" type="success" size="small">已完成</el-tag>
            <el-tag
              v-else-if="row.status === 'terminated'"
              size="small"
              style="background: #f5f5f5; border-color: #8c8c8c; color: #8c8c8c"
            >已终止</el-tag>
            <el-tag v-else type="warning" size="small">待办</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="紧迫状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              v-if="row.urgency_status === 'overdue'"
              type="danger"
              size="small"
            >超时</el-tag>
            <el-tag
              v-else-if="row.urgency_status === 'expired'"
              size="small"
              style="background: #f5f5f5; border-color: #8c8c8c; color: #8c8c8c"
            >失效</el-tag>
            <el-tag
              v-else-if="row.is_time_limited && row.urgency_status === 'normal'"
              type="success"
              size="small"
            >正常</el-tag>
            <span v-else class="text-tertiary">--</span>
          </template>
        </el-table-column>
        <el-table-column label="限时" width="70" align="center">
          <template #default="{ row }">
            <span v-if="row.is_time_limited" class="time-limited-badge">是</span>
            <span v-else class="text-tertiary">否</span>
          </template>
        </el-table-column>
        <el-table-column label="置顶" width="70" align="center">
          <template #default="{ row }">
            <el-button
              :icon="row.is_pinned ? StarFilled : Star"
              size="small"
              text
              :type="row.is_pinned ? 'warning' : 'info'"
              @click="togglePin(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-dropdown trigger="click" @command="(cmd: string) => handleAction(cmd, row)">
              <el-button type="primary" link size="small">
                更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="invite" :disabled="row.status === 'terminated'">
                    <el-icon><Message /></el-icon> 邀请协办
                  </el-dropdown-item>
                  <el-dropdown-item
                    v-if="row.status !== 'terminated' && row.status !== 'completed'"
                    command="terminate"
                  >
                    <el-icon style="color: #FF5630"><Close /></el-icon>
                    <span style="color: #FF5630">终止任务</span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div v-if="pagination.total > 0" class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadData"
          @current-change="loadData"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Download, Star, StarFilled, Loading, ArrowDown, Message, Close } from '@element-plus/icons-vue'
import { listTodos, pinTodo, exportTodos as triggerExport, inviteTodo, terminateTodo, type TodoItem } from '@/api/todo'

const loading = ref(false)
const items = ref<TodoItem[]>([])
const dateRange = ref<[string, string] | null>(null)

const filters = reactive({
  keyword: '',
  status: '' as '' | 'pending' | 'completed' | 'terminated',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

async function loadData() {
  loading.value = true
  try {
    const params = {
      keyword: filters.keyword || undefined,
      start_date: dateRange.value ? dateRange.value[0] : undefined,
      end_date: dateRange.value ? dateRange.value[1] : undefined,
      status: filters.status || undefined,
      page: pagination.page,
      page_size: pagination.pageSize,
    }
    const data = await listTodos(params)
    items.value = data.items || []
    pagination.total = data.total || 0
  } catch {
    ElMessage.error('加载待办列表失败')
  } finally {
    loading.value = false
  }
}

async function togglePin(row: TodoItem) {
  try {
    await pinTodo(row.id, !row.is_pinned)
    row.is_pinned = !row.is_pinned
    ElMessage.success(row.is_pinned ? '置顶成功' : '已取消置顶')
  } catch {
    ElMessage.error('操作失败')
  }
}

function handleSearch() {
  pagination.page = 1
  loadData()
}

function handleReset() {
  filters.keyword = ''
  filters.status = ''
  dateRange.value = null
  pagination.page = 1
  loadData()
}

function handleExport() {
  triggerExport()
}

async function handleAction(cmd: string, row: TodoItem) {
  if (cmd === 'invite') {
    try {
      const result = await inviteTodo(row.id)
      await navigator.clipboard.writeText(result.url)
      ElMessage.success('邀请链接已复制到剪贴板')
    } catch {
      ElMessage.error('邀请失败')
    }
  } else if (cmd === 'terminate') {
    try {
      await ElMessageBox.confirm(
        '终止后可在筛选中查看，数据保留',
        '确认终止此待办？',
        { confirmButtonText: '确认终止', cancelButtonText: '取消', type: 'warning' }
      )
      await terminateTodo(row.id)
      ElMessage.success('任务已终止')
      loadData()
    } catch {
      // user cancelled or error
    }
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '--'
  try {
    return new Date(dateStr).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' })
  } catch {
    return dateStr
  }
}

function disabledDate(date: Date): boolean {
  // 禁用未来日期
  const today = new Date()
  today.setHours(23, 59, 59, 999)
  return date.getTime() > today.getTime()
}

onMounted(loadData)
</script>

<style scoped lang="scss">
.todo-list-view {
  padding: 20px 24px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.page-title {
  font-size: 22px;
  font-weight: 700;
  color: #1a1a1a;
  margin: 0;
}

.section {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 16px;
}

.filter-card {
  padding: 16px 20px;
}

.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.loading {
  display: flex;
  justify-content: center;
  padding: 32px;
}

.text-tertiary {
  color: #97a0af;
  font-size: 13px;
}

.time-limited-badge {
  background: #eef1ff;
  color: #4f6ef7;
  border-radius: 4px;
  padding: 2px 6px;
  font-size: 12px;
}
</style>
