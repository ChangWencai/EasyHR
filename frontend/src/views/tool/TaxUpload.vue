<template>
  <div class="tax-upload">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>个税上传</span>
          <el-button size="small" :icon="Download" link @click="downloadSample">
            下载模板
          </el-button>
        </div>
      </template>

      <!-- 上传区 -->
      <div v-if="!uploadResult" class="upload-section">
        <el-form inline @submit.prevent>
          <el-form-item label="对应月份">
            <el-date-picker
              v-model="uploadYM"
              type="month"
              placeholder="选择月份"
              value-format="YYYY-MM"
              style="width: 140px"
            />
          </el-form-item>
        </el-form>

        <el-upload
          ref="uploadRef"
          drag
          :auto-upload="false"
          :accept="'.xlsx,.xls'"
          :limit="1"
          :on-change="handleFileChange"
        >
          <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
          <div class="el-upload__text">
            拖拽 Excel 文件到此处，或 <em>点击上传</em>
          </div>
          <template #tip>
            <div class="el-upload__tip">支持 .xlsx 格式，文件大小不超过 5MB</div>
          </template>
        </el-upload>

        <div class="upload-actions">
          <el-button type="primary" :loading="uploading" :disabled="!selectedFile || !uploadYM" @click="handleUpload">
            上传并预览
          </el-button>
        </div>
      </div>

      <!-- 预览匹配结果 -->
      <div v-else class="preview-section">
        <el-alert
          v-if="uploadResult.matched_count > 0 && uploadResult.unmatched_rows.length > 0"
          type="warning"
          :title="`已成功处理 ${uploadResult.matched_count} 条，${uploadResult.unmatched_rows.length} 行无法匹配（见下方日志）`"
          :closable="false"
          show-icon
          style="margin-bottom: 12px"
        />
        <el-alert
          v-if="uploadResult.matched_count === 0"
          type="error"
          title="所有行都无法匹配，请检查 Excel 文件"
          :closable="false"
          show-icon
          style="margin-bottom: 12px"
        />

        <!-- 匹配成功表格 -->
        <div v-if="uploadResult.matched_count > 0" class="matched-section">
          <div class="section-title">已匹配（{{ uploadResult.matched_count }} 条）</div>
          <el-table :data="uploadResult.matched_rows" stripe size="small" max-height="300">
            <el-table-column prop="row_number" label="行号" width="60" />
            <el-table-column prop="name" label="Excel姓名" min-width="100" />
            <el-table-column prop="employee_name" label="匹配员工" min-width="100">
              <template #default="{ row }">
                <el-tag size="small" type="success">{{ row.employee_name }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="tax_amount" label="个税" width="100">
              <template #default="{ row }">¥{{ row.tax_amount.toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="adjustment" label="应补/应退" width="110">
              <template #default="{ row }">
                <span :class="row.adjustment >= 0 ? 'text-success' : 'text-danger'">
                  {{ row.adjustment >= 0 ? '+' : '' }}¥{{ row.adjustment.toFixed(2) }}
                </span>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 无法匹配表格 -->
        <div v-if="uploadResult.unmatched_rows.length > 0" class="unmatched-section">
          <div class="section-title">无法匹配（{{ uploadResult.unmatched_rows.length }} 条）</div>
          <el-table :data="uploadResult.unmatched_rows" stripe size="small" max-height="200">
            <el-table-column prop="row_number" label="行号" width="60" />
            <el-table-column prop="name" label="Excel姓名" min-width="100" />
            <el-table-column prop="reason" label="原因" min-width="140">
              <template #default="{ row }">
                <el-tag size="small" type="danger">{{ row.reason }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="preview-actions">
          <el-button @click="handleReset">重新上传</el-button>
          <el-button
            type="primary"
            :loading="confirming"
            :disabled="uploadResult.matched_count === 0"
            @click="handleConfirm"
          >
            确认并更新工资表
          </el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { UploadFilled, Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { salaryApi, type TaxUploadResult } from '@/api/salary'

const uploadYM = ref('')
const selectedFile = ref<File | null>(null)
const uploading = ref(false)
const confirming = ref(false)
const uploadResult = ref<TaxUploadResult | null>(null)

function handleFileChange(file: any) {
  selectedFile.value = file.raw as File
}

async function handleUpload() {
  if (!selectedFile.value || !uploadYM.value) return
  const [year, month] = uploadYM.value.split('-').map(Number)
  uploading.value = true
  try {
    uploadResult.value = await salaryApi.uploadTax(year, month, selectedFile.value)
  } catch (e: any) {
    ElMessage.error(e?.message || '上传失败')
  } finally {
    uploading.value = false
  }
}

async function handleConfirm() {
  if (!uploadResult.value || uploadResult.value.matched_count === 0) return
  const [year, month] = uploadYM.value.split('-').map(Number)
  confirming.value = true
  try {
    await salaryApi.confirmTaxUpload({
      year,
      month,
      matched_rows: uploadResult.value.matched_rows,
    })
    ElMessage.success('工资表个税数据已更新，请重新核算')
    handleReset()
  } catch (e: any) {
    ElMessage.error(e?.message || '确认失败')
  } finally {
    confirming.value = false
  }
}

function handleReset() {
  uploadResult.value = null
  selectedFile.value = null
  uploadYM.value = ''
}

function downloadSample() {
  ElMessage.info('请在后台管理系统下载个税导入模板')
}
</script>

<style scoped lang="scss">
.tax-upload {
  padding: 8px;
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.upload-section {
  max-width: 600px;
}
.upload-actions {
  margin-top: 16px;
}
.preview-section {
  margin-top: 8px;
}
.matched-section,
.unmatched-section {
  margin-bottom: 16px;
}
.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #606266;
  margin-bottom: 8px;
}
.text-success {
  color: #67c23a;
}
.text-danger {
  color: #f56c6c;
}
.preview-actions {
  margin-top: 16px;
  display: flex;
  gap: 8px;
}
</style>
