<template>
  <div class="employee-list">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-stats">
        <div class="stat-chip">
          <span class="chip-num">{{ total }}</span>
          <span class="chip-label">员工总数</span>
        </div>
        <div class="stat-chip success">
          <span class="chip-num">{{ activeCount }}</span>
          <span class="chip-label">在职</span>
        </div>
        <div class="stat-chip warning">
          <span class="chip-num">{{ probationCount }}</span>
          <span class="chip-label">试用期</span>
        </div>
      </div>
      <div class="header-actions">
        <router-link to="/employee/invitations" class="action-btn secondary">
          <el-icon :size="16"><Plus /></el-icon>
          入职邀请
        </router-link>
        <router-link to="/employee/offboardings" class="action-btn secondary">
          <el-icon :size="16"><DArrowRight /></el-icon>
          离职
        </router-link>
      </div>
    </div>

    <!-- 搜索栏 -->
    <div class="search-bar">
      <div class="search-input-wrap">
        <el-icon class="search-icon" :size="16"><Search /></el-icon>
        <input
          v-model="search"
          type="search"
          class="search-input"
          placeholder="搜索姓名、手机号、岗位"
          @keydown.enter="load(1)"
        />
        <button v-if="search" class="search-clear" @click="clearSearch" aria-label="清除搜索">
          <el-icon :size="14"><Close /></el-icon>
        </button>
      </div>
      <button class="export-btn" @click="handleExport" :disabled="exporting">
        <el-icon :size="16"><Download /></el-icon>
        <span>导出</span>
      </button>
    </div>

    <!-- 筛选标签 -->
    <div class="filter-tabs">
      <button
        v-for="f in filters"
        :key="f.value"
        class="filter-tab"
        :class="{ active: activeFilter === f.value }"
        @click="setFilter(f.value)"
      >
        {{ f.label }}
        <span v-if="f.count > 0" class="filter-count">{{ f.count }}</span>
      </button>
    </div>

    <!-- 员工列表 -->
    <div class="employee-content">
      <!-- 加载态 -->
      <div v-if="loading" class="skeleton-list">
        <div v-for="i in 5" :key="i" class="skeleton-card" />
      </div>

      <!-- 空状态 -->
      <div v-else-if="list.length === 0" class="empty-state">
        <svg width="64" height="64" viewBox="0 0 64 64" fill="none">
          <circle cx="32" cy="32" r="32" fill="#F1F5F9"/>
          <circle cx="32" cy="26" r="10" fill="#CBD5E1"/>
          <path d="M16 50c0-8.837 7.163-16 16-16s16 7.163 16 16" fill="#CBD5E1"/>
        </svg>
        <p class="empty-title">{{ search ? '未找到符合条件的员工' : '暂无员工' }}</p>
        <p class="empty-sub">{{ search ? '尝试调整搜索关键词' : '点击右上角新增员工' }}</p>
        <router-link v-if="!search" to="/employee/create" class="add-first-btn">
          <el-icon :size="16"><Plus /></el-icon>
          新增第一个员工
        </router-link>
      </div>

      <!-- 员工卡片列表 -->
      <div v-else class="employee-list-items">
        <div
          v-for="emp in list"
          :key="emp.id"
          class="emp-card"
          role="button"
          tabindex="0"
          @click="goDetail(emp.id)"
          @keydown.enter="goDetail(emp.id)"
        >
          <!-- 头像 + 基本信息 -->
          <div class="emp-main">
            <div class="emp-avatar" :style="{ background: avatarColor(emp.name) }">
              {{ emp.name?.[0] || '?' }}
            </div>
            <div class="emp-info">
              <div class="emp-name-row">
                <span class="emp-name">{{ emp.name }}</span>
                <span class="emp-status" :class="emp.status">
                  {{ statusMap[emp.status] }}
                </span>
              </div>
              <div class="emp-meta">
                <span class="emp-phone">{{ emp.phone }}</span>
                <span class="emp-sep">·</span>
                <span class="emp-position">{{ emp.position || '未填岗位' }}</span>
              </div>
              <div v-if="emp.entry_date" class="emp-entry">
                <el-icon :size="12"><Calendar /></el-icon>
                入职 {{ emp.entry_date }}
              </div>
            </div>
          </div>
          <!-- 操作箭头 -->
          <div class="emp-action">
            <el-icon :size="18" color="#CBD5E1"><ArrowRight /></el-icon>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div v-if="total > pageSize && !loading" class="pagination-wrap">
        <button
          class="page-btn"
          :disabled="page <= 1"
          @click="load(page - 1)"
        >
          <el-icon :size="16"><ArrowLeft /></el-icon>
        </button>
        <span class="page-info">{{ page }} / {{ totalPages }}</span>
        <button
          class="page-btn"
          :disabled="page >= totalPages"
          @click="load(page + 1)"
        >
          <el-icon :size="16"><ArrowRight /></el-icon>
        </button>
      </div>
    </div>

    <!-- 底部新增按钮 -->
    <router-link to="/employee/create" class="fab" aria-label="新增员工">
      <el-icon :size="24" color="#fff"><Plus /></el-icon>
    </router-link>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { employeeApi } from '@/api/employee'
import { ElMessage } from 'element-plus'
import {
  Plus, DArrowRight, Search, Close, Download,
  Calendar, ArrowRight, ArrowLeft,
} from '@element-plus/icons-vue'
import { statusMap } from '@/views/employee/statusMap'

const router = useRouter()

const search = ref('')
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const exporting = ref(false)
const activeFilter = ref('all')

const statusTagType: Record<string, string> = {
  active: 'success',
  probation: 'warning',
  resigned: 'info',
  archived: 'info',
}

// 按状态统计
const activeCount = computed(() => list.value.filter(e => e.status === 'active').length)
const probationCount = computed(() => list.value.filter(e => e.status === 'probation').length)
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

const filters = computed(() => [
  { label: '全部', value: 'all', count: 0 },
  { label: '在职', value: 'active', count: 0 },
  { label: '试用期', value: 'probation', count: 0 },
  { label: '已离职', value: 'resigned', count: 0 },
])

function setFilter(val: string) {
  activeFilter.value = val
  load(1)
}

function clearSearch() {
  search.value = ''
  load(1)
}

function goDetail(id: number) {
  router.push(`/employee/${id}`)
}

// 头像背景色：根据姓名生成稳定色
const AVATAR_COLORS = [
  '#0F766E', '#0EA5E9', '#8B5CF6', '#F59E0B',
  '#EF4444', '#10B981', '#EC4899', '#6366F1',
]
function avatarColor(name: string): string {
  let hash = 0
  for (let i = 0; i < (name || '').length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length]
}

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const params: Record<string, any> = { page: p, page_size: pageSize.value }
    if (search.value) params.search = search.value
    if (activeFilter.value !== 'all') params.status = activeFilter.value
    const res = await employeeApi.list(params)
    list.value = res.list
    total.value = res.total
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

function handleExport() {
  if (exporting.value) return
  exporting.value = true
  employeeApi.exportExcel()
    .then((blob) => {
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `员工列表_${Date.now()}.xlsx`
      a.click()
      URL.revokeObjectURL(url)
    })
    .catch(() => {
      ElMessage.error('导出失败')
    })
    .finally(() => {
      exporting.value = false
    })
}

onMounted(() => load())
</script>

<style scoped lang="scss">
.employee-list {
  background: #F8FAFC;
  min-height: 100%;
  padding-bottom: 80px; // 留空给 FAB
}

// ===== 页面头部统计 =====
.page-header {
  background: #fff;
  padding: 16px;
  border-bottom: 1px solid #F1F5F9;
}

.header-stats {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.stat-chip {
  flex: 1;
  background: #F8FAFC;
  border-radius: 10px;
  padding: 10px 8px;
  text-align: center;
  border: 1px solid #E2E8F0;

  &.success { border-color: #DCFCE7; background: #F0FDF4; }
  &.warning { border-color: #FEF3C7; background: #FFFBEB; }

  .chip-num {
    display: block;
    font-size: 20px;
    font-weight: 700;
    color: #0F172A;
    font-feature-settings: "tnum";
    line-height: 1;
  }

  .chip-label {
    display: block;
    font-size: 11px;
    color: #94A3B8;
    margin-top: 3px;
  }
}

.header-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  text-decoration: none;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
  border: none;
  background: #F8FAFC;
  color: #64748B;
  border: 1px solid #E2E8F0;
  min-height: 44px;

  &:active {
    background: #E2E8F0;
  }
}

// ===== 搜索栏 =====
.search-bar {
  display: flex;
  gap: 8px;
  padding: 12px;
  background: #fff;
  border-bottom: 1px solid #F1F5F9;
}

.search-input-wrap {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  color: #94A3B8;
  pointer-events: none;
}

.search-input {
  width: 100%;
  height: 40px;
  padding: 0 36px 0 36px;
  border: 1px solid #E2E8F0;
  border-radius: 10px;
  background: #F8FAFC;
  font-size: 14px;
  color: #0F172A;
  outline: none;
  box-sizing: border-box;
  transition: border-color 0.15s;

  &::placeholder { color: #94A3B8; }
  &:focus { border-color: #0F766E; background: #fff; }
}

.search-clear {
  position: absolute;
  right: 10px;
  background: none;
  border: none;
  cursor: pointer;
  color: #94A3B8;
  display: flex;
  align-items: center;
  padding: 4px;
  border-radius: 50%;

  &:hover { background: #F1F5F9; }
}

.export-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 0 14px;
  height: 40px;
  background: #fff;
  border: 1px solid #E2E8F0;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  color: #64748B;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
  min-height: 40px;

  &:hover { border-color: #0F766E; color: #0F766E; }
  &:disabled { opacity: 0.5; cursor: not-allowed; }
  &:active:not(:disabled) { background: #F1F5F9; }
}

// ===== 筛选标签 =====
.filter-tabs {
  display: flex;
  gap: 6px;
  padding: 10px 12px;
  background: #fff;
  border-bottom: 1px solid #F1F5F9;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  &::-webkit-scrollbar { display: none; }
}

.filter-tab {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  border-radius: 20px;
  font-size: 13px;
  color: #64748B;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s;
  min-height: 32px;

  &.active {
    background: #0F766E;
    border-color: #0F766E;
    color: #fff;

    .filter-count { background: rgba(255,255,255,0.25); color: #fff; }
  }
}

.filter-count {
  font-size: 11px;
  background: #E2E8F0;
  color: #64748B;
  border-radius: 10px;
  padding: 0 5px;
  min-width: 18px;
  text-align: center;
  font-weight: 600;
}

// ===== 员工内容区 =====
.employee-content {
  padding: 12px;
}

// ===== 骨架屏 =====
.skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.skeleton-card {
  height: 80px;
  background: linear-gradient(90deg, #F1F5F9 25%, #E2E8F0 50%, #F1F5F9 75%);
  background-size: 200% 100%;
  border-radius: 12px;
  animation: shimmer 1.4s infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

// ===== 空状态 =====
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48px 16px;
  gap: 8px;
}

.empty-title {
  font-size: 15px;
  font-weight: 600;
  color: #64748B;
  margin: 0;
}

.empty-sub {
  font-size: 13px;
  color: #94A3B8;
  margin: 0;
}

.add-first-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 12px;
  padding: 10px 20px;
  background: #0F766E;
  color: #fff;
  border-radius: 24px;
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  cursor: pointer;
  transition: background 0.15s;

  &:hover { background: #0D6B62; }
  &:active { background: #115E59; }
}

// ===== 员工卡片 =====
.employee-list-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.emp-card {
  display: flex;
  align-items: center;
  background: #fff;
  border-radius: 14px;
  padding: 14px;
  cursor: pointer;
  transition: transform 0.15s ease-out, box-shadow 0.15s ease-out;
  border: 1px solid transparent;
  box-shadow: 0 1px 3px rgba(0,0,0,0.03);
  min-height: 44px;

  &:hover {
    border-color: #E2E8F0;
    box-shadow: 0 2px 8px rgba(0,0,0,0.06);
  }

  &:active {
    transform: scale(0.99);
  }
}

.emp-main {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.emp-avatar {
  width: 46px;
  height: 46px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 600;
  color: #fff;
  flex-shrink: 0;
  letter-spacing: -0.5px;
}

.emp-info {
  flex: 1;
  min-width: 0;
}

.emp-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.emp-name {
  font-size: 15px;
  font-weight: 600;
  color: #0F172A;
}

.emp-status {
  font-size: 11px;
  padding: 1px 7px;
  border-radius: 10px;
  font-weight: 500;
  flex-shrink: 0;

  &.active { background: #DCFCE7; color: #16A34A; }
  &.probation { background: #FEF3C7; color: #D97706; }
  &.resigned { background: #F1F5F9; color: #94A3B8; }
  &.archived { background: #F1F5F9; color: #94A3B8; }
}

.emp-meta {
  font-size: 12px;
  color: #64748B;
  margin-top: 3px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.emp-sep { color: #CBD5E1; }

.emp-position {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.emp-entry {
  font-size: 11px;
  color: #94A3B8;
  margin-top: 3px;
  display: flex;
  align-items: center;
  gap: 3px;
}

.emp-action {
  flex-shrink: 0;
  padding-left: 8px;
}

// ===== 分页 =====
.pagination-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 16px 0 4px;
}

.page-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  border: 1px solid #E2E8F0;
  border-radius: 10px;
  cursor: pointer;
  color: #64748B;
  transition: all 0.15s;

  &:hover:not(:disabled) { border-color: #0F766E; color: #0F766E; }
  &:disabled { opacity: 0.4; cursor: not-allowed; }
}

.page-info {
  font-size: 13px;
  color: #64748B;
  font-feature-settings: "tnum";
}

// ===== FAB =====
.fab {
  position: fixed;
  right: 20px;
  bottom: calc(70px + env(safe-area-inset-bottom, 0px));
  bottom: calc(70px + constant(safe-area-inset-bottom, 0px));
  width: 52px;
  height: 52px;
  background: #0F766E;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 14px rgba(15, 118, 110, 0.4);
  cursor: pointer;
  text-decoration: none;
  transition: transform 0.2s ease-out, box-shadow 0.2s;
  z-index: 100;

  &:hover {
    transform: scale(1.05);
    box-shadow: 0 6px 20px rgba(15, 118, 110, 0.5);
  }

  &:active {
    transform: scale(0.96);
  }
}
</style>
