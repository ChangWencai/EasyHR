<template>
  <div class="sms-tool">
    <!-- Page Header -->
    <div class="page-head">
      <div class="page-head-left">
        <div class="page-head-indicator"></div>
        <div>
          <h2 class="page-head-title">短信模板</h2>
          <p class="page-head-desc">管理阿里云短信通知模板</p>
        </div>
      </div>
      <el-button type="primary" @click="openDialog()">新增模板</el-button>
    </div>

    <!-- Template Grid -->
    <div class="template-shell">
      <div v-if="list.length === 0 && !loading" class="empty-state">
        <div class="empty-icon">
          <el-icon :size="40"><ChatDotRound /></el-icon>
        </div>
        <p class="empty-title">暂无模板</p>
        <p class="empty-desc">点击右上角「新增模板」创建你的第一个短信模板</p>
      </div>

      <div v-else class="template-grid">
        <div
          v-for="tpl in list"
          :key="tpl.id"
          class="template-card"
        >
          <div class="template-card-head">
            <div class="template-card-icon">
              <el-icon :size="18"><ChatDotRound /></el-icon>
            </div>
            <div class="template-card-meta">
              <span class="template-card-name">{{ tpl.name }}</span>
              <el-tag size="small" effect="plain" type="warning">{{ sceneLabels[tpl.scene] || tpl.scene }}</el-tag>
              <el-tag v-if="tpl.is_default" type="success" size="small" effect="plain">默认</el-tag>
            </div>
          </div>
          <div class="template-card-code">模板代码：{{ tpl.template_code }}</div>
          <div class="template-card-preview">{{ tpl.content.substring(0, 80) }}{{ tpl.content.length > 80 ? '...' : '' }}</div>
          <div class="template-card-actions">
            <el-button size="small" @click="openDialog(tpl)">编辑</el-button>
            <el-popconfirm :title="`确认删除模板「${tpl.name}」？`" confirm-button-text="确认" cancel-button-text="取消" @confirm="handleDelete(tpl.id)">
              <template #reference>
                <el-button size="small" type="danger" plain>删除</el-button>
              </template>
            </el-popconfirm>
          </div>
        </div>
      </div>

      <el-pagination
        v-if="total > pageSize"
        class="mt-4"
        layout="total,prev,pager,next"
        :total="total"
        :page="page"
        :page-size="pageSize"
        @current-change="load"
      />
    </div>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑模板' : '新增模板'" width="560px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="模板名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入模板名称，如：验证码短信" maxlength="100" />
        </el-form-item>
        <el-form-item label="使用场景" prop="scene">
          <el-select v-model="form.scene" placeholder="请选择使用场景" style="width: 100%">
            <el-option v-for="(label, key) in sceneLabels" :key="key" :label="label" :value="key" />
          </el-select>
        </el-form-item>
        <el-form-item label="模板代码" prop="template_code">
          <el-input v-model="form.template_code" placeholder="阿里云短信模板代码，如：SMS_12345678" maxlength="50" />
        </el-form-item>
        <el-form-item label="模板内容" prop="content">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="6"
            placeholder="短信模板内容，支持变量：${name}、${code}、${company}、${url} 等"
          />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="form.is_default">设为默认模板</el-checkbox>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { ChatDotRound } from '@element-plus/icons-vue'
import { smsTemplateApi, sceneLabels } from '@/api/sms_template'
import type { SmsTemplate } from '@/api/sms_template'

const loading = ref(false)
const list = ref<SmsTemplate[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const formRef = ref<FormInstance>()

const form = reactive({
  name: '',
  scene: '',
  template_code: '',
  content: '',
  is_default: false,
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
  scene: [{ required: true, message: '请选择使用场景', trigger: 'change' }],
  template_code: [{ required: true, message: '请输入模板代码', trigger: 'blur' }],
  content: [{ required: true, message: '请输入模板内容', trigger: 'blur' }],
}

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const res = await smsTemplateApi.list({ page: p, page_size: pageSize.value })
    list.value = res?.list ?? []
    total.value = res?.total ?? 0
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

function openDialog(row?: SmsTemplate) {
  if (row) {
    editingId.value = row.id
    form.name = row.name
    form.scene = row.scene
    form.template_code = row.template_code
    form.content = row.content
    form.is_default = row.is_default
  } else {
    editingId.value = null
    form.name = ''
    form.scene = ''
    form.template_code = ''
    form.content = ''
    form.is_default = false
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  saving.value = true
  try {
    if (editingId.value) {
      await smsTemplateApi.update(editingId.value, {
        name: form.name,
        scene: form.scene,
        template_code: form.template_code,
        content: form.content,
        is_default: form.is_default,
      })
      ElMessage.success('保存成功')
    } else {
      await smsTemplateApi.create({
        name: form.name,
        scene: form.scene,
        template_code: form.template_code,
        content: form.content,
        is_default: form.is_default,
      })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    load(page.value)
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await smsTemplateApi.delete(id)
    ElMessage.success('已删除')
    load(page.value)
  } catch {
    ElMessage.error('删除失败')
  }
}

onMounted(() => load())
</script>

<style scoped lang="scss">
$primary: #7C3AED;
$text-primary: #1A1D2E;
$text-secondary: #5E6278;
$text-muted: #A0A3BD;
$border: #E8EBF0;
$surface: #FFFFFF;
$surface-alt: #F8F9FC;

.sms-tool {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.page-head-left {
  display: flex;
  align-items: center;
  gap: 14px;
}

.page-head-indicator {
  width: 4px;
  height: 36px;
  border-radius: 4px;
  background: linear-gradient(180deg, #F59E0B 0%, #FBBF24 100%);
}

.page-head-title {
  font-size: 20px;
  font-weight: 700;
  color: $text-primary;
  margin: 0;
  letter-spacing: -0.3px;
}

.page-head-desc {
  font-size: 13px;
  color: $text-muted;
  margin: 4px 0 0;
}

.template-shell {
  background: $surface;
  border: 1px solid $border;
  border-radius: 20px;
  padding: 24px;
}

.template-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.template-card {
  background: $surface-alt;
  border: 1px solid $border;
  border-radius: 16px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  transition: all 0.25s ease;

  &:hover {
    border-color: rgba($primary, 0.25);
    box-shadow: 0 4px 20px rgba(0,0,0,0.06);
    transform: translateY(-2px);
  }
}

.template-card-head {
  display: flex;
  align-items: center;
  gap: 12px;
}

.template-card-icon {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%);
  color: #D97706;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.template-card-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
  flex-wrap: wrap;
}

.template-card-name {
  font-size: 14px;
  font-weight: 600;
  color: $text-primary;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.template-card-code {
  font-size: 12px;
  color: $text-secondary;
  font-family: monospace;
  background: rgba($primary, 0.06);
  padding: 4px 8px;
  border-radius: 6px;
  display: inline-block;
  align-self: flex-start;
}

.template-card-preview {
  font-size: 12px;
  color: $text-muted;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.template-card-actions {
  display: flex;
  gap: 8px;
  margin-top: 4px;
  padding-top: 12px;
  border-top: 1px solid $border;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 60px 20px;
  text-align: center;
}

.empty-icon {
  width: 72px;
  height: 72px;
  border-radius: 20px;
  background: $surface-alt;
  color: $text-muted;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
}

.empty-title {
  font-size: 15px;
  font-weight: 600;
  color: $text-primary;
  margin: 0 0 4px;
}

.empty-desc {
  font-size: 13px;
  color: $text-muted;
  margin: 0;
}

.mt-4 {
  margin-top: 20px;
}

@media (max-width: 768px) {
  .template-grid {
    grid-template-columns: 1fr;
  }
}
</style>
