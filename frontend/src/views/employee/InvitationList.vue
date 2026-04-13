<template>
  <div class="invitation-list">
    <el-card>
      <template #header>
        <div class="header">
          <span>入职邀请</span>
          <el-button type="primary" @click="showDialog = true">发送邀请</el-button>
        </div>
      </template>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="name" label="姓名" min-width="80" />
        <el-table-column prop="phone" label="手机号" min-width="120" />
        <el-table-column prop="status" label="状态" min-width="90">
          <template #default="{ row }">
            <el-tag :type="invitationStatusTagType[row.status]" size="small">
              {{ invitationStatusMap[row.status] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="发送时间" min-width="160" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'pending'"
              size="small"
              @click="copyLink(row.invite_url)"
            >
              复制链接
            </el-button>
            <el-button
              v-if="row.status === 'pending'"
              size="small"
              type="danger"
              @click="handleCancel(row.id)"
            >
              取消邀请
            </el-button>
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

    <!-- 发送邀请对话框 -->
    <el-dialog v-model="showDialog" title="发送入职邀请" width="400px">
      <el-form ref="dialogFormRef" :model="dialogForm" :rules="dialogRules" label-width="80px">
        <el-form-item label="姓名" prop="name">
          <el-input v-model="dialogForm.name" placeholder="请输入员工姓名" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="dialogForm.phone" placeholder="请输入手机号" maxlength="11" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="sending" @click="handleSend">发送</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { employeeApi } from '@/api/employee'
import { ElMessage, FormInstance, FormRules } from 'element-plus'
import { invitationStatusMap, invitationStatusTagType } from './statusMap'

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const showDialog = ref(false)
const sending = ref(false)
const dialogFormRef = ref<FormInstance>()
const inviteUrl = ref('')

const dialogForm = reactive({ name: '', phone: '' })
const dialogRules: FormRules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确', trigger: 'blur' },
  ],
}

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const res = await employeeApi.invitations({ page: p, page_size: pageSize.value })
    list.value = res.list
    total.value = res.total
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

async function handleSend() {
  if (sending.value) return
  if (!dialogFormRef.value) return
  try {
    await dialogFormRef.value.validate()
  } catch {
    return
  }

  sending.value = true
  try {
    const res = await employeeApi.createInvitation(dialogForm)
    inviteUrl.value = res.invite_url
    ElMessage.success('邀请已发送')
    showDialog.value = false
    dialogForm.name = ''
    dialogForm.phone = ''
    load()
    // auto copy
    copyLink(res.invite_url)
  } catch {
    ElMessage.error('发送失败')
  } finally {
    sending.value = false
  }
}

function copyLink(url: string) {
  navigator.clipboard.writeText(url).then(() => {
    ElMessage.success('链接已复制到剪贴板')
  }).catch(() => {
    ElMessage.info(url)
  })
}

const cancelling = ref(false)

async function handleCancel(id: number) {
  if (cancelling.value) return
  cancelling.value = true
  try {
    await employeeApi.cancelInvitation(id)
    ElMessage.success('已取消')
    load()
  } catch {
    ElMessage.error('取消失败')
  } finally {
    cancelling.value = false
  }
}

onMounted(() => load())
</script>

<style scoped lang="scss">
.invitation-list {
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
