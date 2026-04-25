<template>
  <div class="email-tool">
    <!-- Page Header -->
    <div class="page-head">
      <div class="page-head-left">
        <div class="page-head-indicator"></div>
        <div>
          <h2 class="page-head-title">邮箱模板</h2>
          <p class="page-head-desc">管理员工邮件通知模板</p>
        </div>
      </div>
      <el-button type="primary" @click="openDialog()">新增模板</el-button>
    </div>

    <!-- Template Grid -->
    <div class="template-shell">
      <div v-if="list.length === 0 && !loading" class="empty-state">
        <div class="empty-icon">
          <el-icon :size="40"><Message /></el-icon>
        </div>
        <p class="empty-title">暂无模板</p>
        <p class="empty-desc">点击右上角「新增模板」创建你的第一个邮件模板</p>
      </div>

      <div v-else class="template-grid">
        <div
          v-for="tpl in list"
          :key="tpl.id"
          class="template-card"
        >
          <div class="template-card-head">
            <div class="template-card-icon">
              <el-icon :size="18"><Message /></el-icon>
            </div>
            <div class="template-card-meta">
              <span class="template-card-name">{{ tpl.name }}</span>
              <el-tag v-if="tpl.is_default" type="success" size="small" effect="plain">默认</el-tag>
            </div>
          </div>
          <div class="template-card-subject">{{ tpl.subject }}</div>
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
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="模板名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入模板名称，如：入职邀请邮件" maxlength="100" />
        </el-form-item>
        <el-form-item label="邮件主题" prop="subject">
          <el-input v-model="form.subject" placeholder="请输入邮件主题，如：{{name}}，您有一份入职邀请" maxlength="200" />
        </el-form-item>
        <el-form-item label="邮件正文" prop="content">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="8"
            placeholder="请输入邮件正文内容，支持变量占位符：
{{name}} - 员工姓名
{{company}} - 公司名称
{{position}} - 岗位
{{invite_url}} - 邀请链接"
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
import { Message } from '@element-plus/icons-vue'
import { emailTemplateApi } from '@/api/email_template'
import type { EmailTemplate } from '@/api/email_template'

const loading = ref(false)
const list = ref<EmailTemplate[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const formRef = ref<FormInstance>()

const form = reactive({
  name: '',
  subject: '',
  content: '',
  is_default: false,
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
  subject: [{ required: true, message: '请输入邮件主题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入邮件正文', trigger: 'blur' }],
}

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const res = await emailTemplateApi.list({ page: p, page_size: pageSize.value })
    list.value = res?.list ?? []
    total.value = res?.total ?? 0
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

function openDialog(row?: EmailTemplate) {
  if (row) {
    editingId.value = row.id
    form.name = row.name
    form.subject = row.subject
    form.content = row.content
    form.is_default = row.is_default
  } else {
    editingId.value = null
    form.name = ''
    form.subject = ''
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
      await emailTemplateApi.update(editingId.value, {
        name: form.name,
        subject: form.subject,
        content: form.content,
        is_default: form.is_default,
      })
      ElMessage.success('保存成功')
    } else {
      await emailTemplateApi.create({
        name: form.name,
        subject: form.subject,
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
    await emailTemplateApi.delete(id)
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
$primary-light: #A78BFA;
$text-primary: #1A1D2E;
$text-secondary: #5E6278;
$text-muted: #A0A3BD;
$border: #E8EBF0;
$surface: #FFFFFF;
$surface-alt: #F8F9FC;

.email-tool {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* ─── Page Header ─── */
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
  background: linear-gradient(180deg, #DC2626 0%, #F87171 100%);
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

/* ─── Template Shell ─── */
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
  background: linear-gradient(135deg, #FEE2E2 0%, #FECACA 100%);
  color: #DC2626;
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
}

.template-card-name {
  font-size: 14px;
  font-weight: 600;
  color: $text-primary;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.template-card-subject {
  font-size: 13px;
  color: $text-secondary;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

/* ─── Empty State ─── */
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
