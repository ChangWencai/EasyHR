<template>
  <div class="carousel-manage-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">轮播图管理</h1>
        <span class="carousel-hint">最多配置3张轮播图</span>
      </div>
      <div class="header-right">
        <el-button type="primary" :disabled="carouselList.length >= 3" @click="showAddDialog">
          <el-icon><Plus /></el-icon>
          添加轮播图
        </el-button>
      </div>
    </div>

    <!-- 轮播图列表 -->
    <div class="section">
      <div v-if="loading" class="loading">
        <el-icon class="is-loading" size="24"><Loading /></el-icon>
      </div>
      <div v-else-if="carouselList.length === 0" class="empty-state">
        <el-empty description="暂无轮播图，点击上方按钮添加" :image-size="80" />
      </div>
      <el-table v-else :data="carouselList" stripe>
        <el-table-column label="排序" width="80" align="center">
          <template #default="{ $index }">
            <span class="sort-num">{{ $index + 1 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="预览图" width="180">
          <template #default="{ row }">
            <el-image
              :src="row.image_url"
              fit="cover"
              style="width: 120px; height: 60px; border-radius: 4px"
              :preview-src-list="[row.image_url]"
            />
          </template>
        </el-table-column>
        <el-table-column prop="link_url" label="跳转链接" min-width="160">
          <template #default="{ row }">
            <span v-if="row.link_url" class="link-text">{{ row.link_url }}</span>
            <span v-else class="text-tertiary">--</span>
          </template>
        </el-table-column>
        <el-table-column label="生效时间" width="160">
          <template #default="{ row }">
            <span v-if="row.start_at">{{ row.start_at }}</span>
            <span v-else class="text-tertiary">--</span>
          </template>
        </el-table-column>
        <el-table-column label="失效时间" width="160">
          <template #default="{ row }">
            <span v-if="row.end_at">{{ row.end_at }}</span>
            <span v-else class="text-tertiary">--</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              :model-value="row.active"
              size="small"
              @change="handleToggleActive(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 添加/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'add' ? '添加轮播图' : '编辑轮播图'"
      width="560px"
      @close="handleDialogClose"
    >
      <el-form ref="formRef" :model="form" label-position="top">
        <el-form-item label="轮播图片" required>
          <div class="upload-area">
            <el-upload
              v-if="!form.image_url"
              class="image-uploader"
              :show-file-list="false"
              :before-upload="handleBeforeUpload"
              :http-request="handleUpload"
            >
              <el-icon class="upload-icon"><Plus /></el-icon>
              <div class="upload-text">点击上传</div>
            </el-upload>
            <div v-else class="upload-preview">
              <el-image :src="form.image_url" fit="cover" style="width: 200px; height: 100px; border-radius: 8px" />
              <el-button type="danger" size="small" class="replace-btn" @click="form.image_url = ''">替换</el-button>
            </div>
          </div>
          <div class="upload-tip">支持 jpg/jpeg/png/gif/webp，不超过5MB，建议尺寸 750x375</div>
        </el-form-item>
        <el-form-item label="跳转链接">
          <el-input v-model="form.link_url" placeholder="http:// 或 /path，可留空" />
        </el-form-item>
        <el-form-item label="生效时间">
          <el-date-picker
            v-model="form.start_at"
            type="datetime"
            placeholder="留空表示立即生效"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="失效时间">
          <el-date-picker
            v-model="form.end_at"
            type="datetime"
            placeholder="留空表示永久有效"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="是否启用">
          <el-switch v-model="form.active" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Delete, Loading } from '@element-plus/icons-vue'
import {
  listAllCarousels,
  createCarousel,
  updateCarousel,
  deleteCarousel,
  uploadImage,
  type CarouselItem,
} from '@/api/carousel'

const loading = ref(false)
const submitting = ref(false)
const carouselList = ref<CarouselItem[]>([])
const dialogVisible = ref(false)
const dialogMode = ref<'add' | 'edit'>('add')
const editingId = ref<number | null>(null)
const formRef = ref()
void formRef // used in template

const form = reactive({
  image_url: '',
  link_url: '',
  start_at: '',
  end_at: '',
  active: true,
})

async function loadData() {
  loading.value = true
  try {
    const res = await listAllCarousels()
    carouselList.value = res.data || []
  } catch {
    ElMessage.error('加载轮播图列表失败')
  } finally {
    loading.value = false
  }
}

function showAddDialog() {
  dialogMode.value = 'add'
  editingId.value = null
  resetForm()
  dialogVisible.value = true
}

function handleEdit(row: CarouselItem) {
  dialogMode.value = 'edit'
  editingId.value = row.id
  form.image_url = row.image_url
  form.link_url = row.link_url || ''
  form.start_at = row.start_at || ''
  form.end_at = row.end_at || ''
  form.active = row.active
  dialogVisible.value = true
}

function handleDialogClose() {
  resetForm()
}

function resetForm() {
  form.image_url = ''
  form.link_url = ''
  form.start_at = ''
  form.end_at = ''
  form.active = true
}

function handleBeforeUpload(file: File) {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size <= 5 * 1024 * 1024
  if (!isImage) {
    ElMessage.error('只能上传图片文件')
    return false
  }
  if (!isLt5M) {
    ElMessage.error('图片大小不能超过5MB')
    return false
  }
  return true
}

async function handleUpload(option: { file: File }) {
  try {
    const url = await uploadImage(option.file)
    form.image_url = url
    ElMessage.success('上传成功')
  } catch {
    ElMessage.error('上传失败')
  }
}

async function handleSubmit() {
  if (!form.image_url) {
    ElMessage.error('请上传轮播图片')
    return
  }
  submitting.value = true
  try {
    const data = {
      image_url: form.image_url,
      link_url: form.link_url || undefined,
      active: form.active,
      start_at: form.start_at || undefined,
      end_at: form.end_at || undefined,
    }
    if (dialogMode.value === 'add') {
      await createCarousel(data)
      ElMessage.success('添加成功')
    } else if (editingId.value !== null) {
      await updateCarousel(editingId.value, data)
      ElMessage.success('更新成功')
    }
    dialogVisible.value = false
    loadData()
  } catch (err: unknown) {
    const errorObj = err as { message?: string }
    ElMessage.error(errorObj.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function handleToggleActive(row: CarouselItem) {
  try {
    await updateCarousel(row.id, { active: !row.active })
    row.active = !row.active
    ElMessage.success(row.active ? '已启用' : '已停用')
  } catch {
    ElMessage.error('操作失败')
  }
}

async function handleDelete(row: CarouselItem) {
  try {
    await ElMessageBox.confirm('删除后不可恢复，确认删除？', '删除确认', {
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await deleteCarousel(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch {
    // user cancelled
  }
}

onMounted(loadData)
</script>

<style scoped lang="scss">
.carousel-manage-page {
  padding: 20px 24px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.header-left {
  display: flex;
  align-items: baseline;
  gap: 12px;
}

.page-title {
  font-size: 22px;
  font-weight: 700;
  color: #1a1a1a;
  margin: 0;
}

.carousel-hint {
  font-size: 13px;
  color: #97a0af;
}

.section {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
}

.loading {
  display: flex;
  justify-content: center;
  padding: 32px;
}

.sort-num {
  font-weight: 600;
  color: #4f6ef7;
}

.link-text {
  color: #4f6ef7;
  font-size: 13px;
  word-break: break-all;
}

.text-tertiary {
  color: #97a0af;
  font-size: 13px;
}

// 上传区域
.upload-area {
  display: flex;
  align-items: center;
  gap: 16px;
}

.image-uploader {
  width: 200px;
  height: 100px;
  border: 1px dashed #d9d9d9;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: border-color 0.2s;
  &:hover {
    border-color: #4f6ef7;
  }
}

.upload-icon {
  font-size: 28px;
  color: #8c8c8c;
}

.upload-text {
  font-size: 13px;
  color: #8c8c8c;
  margin-top: 4px;
}

.upload-preview {
  position: relative;
}

.replace-btn {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  opacity: 0;
  transition: opacity 0.2s;
}

.upload-preview:hover .replace-btn {
  opacity: 1;
}

.upload-tip {
  font-size: 12px;
  color: #97a0af;
  margin-top: 8px;
}
</style>
