<template>
  <div class="email-template-list">
    <el-card>
      <template #header>
        <div class="header">
          <span>邮箱模板管理</span>
          <el-button type="primary" @click="openDialog()">新增模板</el-button>
        </div>
      </template>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="name" label="模板名称" min-width="120" />
        <el-table-column prop="subject" label="邮件主题" min-width="180" show-overflow-tooltip />
        <el-table-column prop="content" label="正文预览" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.content.substring(0, 50) }}{{ row.content.length > 50 ? '…' : '' }}
          </template>
        </el-table-column>
        <el-table-column prop="is_default" label="默认模板" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_default" type="success" size="small">默认</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openDialog(row)">编辑</el-button>
            <el-popconfirm
              :title="`确认删除模板「${row.name}」？`"
              confirm-button-text="确认"
              cancel-button-text="取消"
              @confirm="handleDelete(row.id)"
            >
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="mt-4"
        layout="total,prev,pager,next"
        :total="total"
        :page="page"
        :page-size="pageSize"
        @current-change="load"
      />
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="editingId ? '编辑模板' : '新增模板'"
      width="560px"
      destroy-on-close
    >
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
.email-template-list {
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.mt-4 {
  margin-top: 16px;
}
</style>
