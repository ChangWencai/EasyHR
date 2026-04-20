<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { UploadFilled, SuccessFilled } from '@element-plus/icons-vue'
import * as XLSX from 'xlsx'

// Props interface
const props = defineProps<{
  templateLabel: string
  templateFields: string[]
  importApi: (rows: Record<string, unknown>[]) => Promise<{ success: number; failed: number }>
}>()

// Emit events
const emit = defineEmits<{
  complete: [{ success: number; failed: number }]
  'update:dialogVisible': [visible: boolean]
}>()

// State
const currentStep = ref(0) // 0=upload, 1=preview, 2=confirm
const fileList = ref<File[]>([])
const loading = ref(false)
const importing = ref(false)
const excelFile = ref<File | null>(null)

interface ParsedRow {
  _raw: Record<string, unknown>
  _qualified: boolean
  _errors: Record<string, string>
}

const parsedRows = ref<ParsedRow[]>([])

const qualifiedCount = computed(() => parsedRows.value.filter(r => r._qualified).length)
const errorCount = computed(() => parsedRows.value.filter(r => !r._qualified).length)

/**
 * Download Excel template with column headers
 */
function downloadTemplate() {
  const ws = XLSX.utils.json_to_sheet([])
  XLSX.utils.sheet_add_aoa(ws, [props.templateFields])
  const wb = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(wb, ws, 'Sheet1')
  XLSX.writeFile(wb, `${props.templateLabel}导入模板.xlsx`)
}

/**
 * Handle file selection from el-upload
 */
function handleFileChange(file: { raw?: File; status?: string }) {
  const raw = file.raw
  if (!raw) return
  excelFile.value = raw
  fileList.value = [raw]
}

/**
 * Parse Excel file and validate each row
 */
async function parseFile() {
  if (!excelFile.value) return
  loading.value = true
  try {
    const buffer = await excelFile.value.arrayBuffer()
    const wb = XLSX.read(buffer, { type: 'array' })
    const ws = wb.Sheets[wb.SheetNames[0]]
    const rawData = XLSX.utils.sheet_to_json<Record<string, unknown>>(ws)

    parsedRows.value = rawData.map((row) => {
      const errors: Record<string, string> = {}

      // Required field: 姓名
      if (!row['姓名']) errors['姓名'] = '姓名不能为空'

      // Phone: 11 digits starting with 1
      const phone = String(row['手机号'] || '')
      if (phone && !/^1[3-9]\d{9}$/.test(phone)) {
        errors['手机号'] = '手机号格式错误，请输入11位数字'
      }

      // ID number: 18 digits
      const id = String(row['身份证号'] || '')
      if (id && !/^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$/.test(id)) {
        errors['身份证号'] = '身份证号格式不正确'
      }

      // entry_date: YYYY-MM-DD format
      const date = String(row['入职日期'] || '')
      if (date && !/^\d{4}-\d{2}-\d{2}$/.test(date)) {
        errors['入职日期'] = '日期格式需为YYYY-MM-DD'
      }

      return { _raw: row, _qualified: Object.keys(errors).length === 0, _errors: errors }
    })

    currentStep.value = 1
  } finally {
    loading.value = false
  }
}

/**
 * Get CSS class for table row based on qualification status
 */
function getRowClass({ row }: { row: ParsedRow }) {
  if (!row._qualified) return 'error-row'
  return 'qualified-row'
}

/**
 * Confirm and execute the import with qualified rows only
 */
async function confirmImport() {
  importing.value = true
  try {
    const qualified = parsedRows.value.filter(r => r._qualified).map(r => r._raw)
    const result = await props.importApi(qualified)
    ElMessage.success(`导入完成：成功${result.success}条，失败${result.failed}条`)
    emit('complete', result)
    emit('update:dialogVisible', false)
  } catch {
    ElMessage.error('导入失败，请稍后重试')
  } finally {
    importing.value = false
  }
}
</script>

<template>
  <div class="excel-import-wizard">
    <!-- Step 0: Upload -->
    <div v-show="currentStep === 0" class="step-content">
      <el-upload
        ref="uploadRef"
        class="excel-upload-area"
        drag
        :auto-upload="false"
        :limit="1"
        accept=".xlsx,.xls"
        :on-change="handleFileChange"
        :file-list="fileList"
      >
        <el-icon class="upload-icon"><UploadFilled /></el-icon>
        <div class="upload-text">拖拽Excel文件到此处，或 <em>点击上传</em></div>
        <div class="upload-hint">支持 .xlsx .xls 格式</div>
      </el-upload>
      <div class="step-actions">
        <el-button @click="downloadTemplate">
          <el-icon><Download /></el-icon>
          下载模板
        </el-button>
        <el-button type="primary" :disabled="fileList.length === 0" :loading="loading" @click="parseFile">
          下一步
        </el-button>
      </div>
    </div>

    <!-- Step 1: Preview -->
    <div v-show="currentStep === 1" class="step-content">
      <div class="preview-header">
        <span>预览：{{ parsedRows.length }}条记录</span>
        <span class="qualified-badge">✓{{ qualifiedCount }}条合格</span>
        <span class="error-badge">✗{{ errorCount }}条有误</span>
      </div>
      <el-table
        :data="parsedRows"
        max-height="400"
        :row-class-name="getRowClass"
        v-loading="loading"
      >
        <el-table-column
          v-for="field in templateFields"
          :key="field"
          :prop="field"
          :label="field"
          min-width="120"
        >
          <template #default="{ row }">
            <span :class="{ 'cell-error': row._errors[field] }">
              {{ row._raw[field] ?? '—' }}
              <span v-if="row._errors[field]" class="field-error">{{ row._errors[field] }}</span>
            </span>
          </template>
        </el-table-column>
      </el-table>
      <div class="step-actions">
        <el-button @click="currentStep = 0">重新上传</el-button>
        <el-button
          type="primary"
          :disabled="qualifiedCount === 0"
          @click="currentStep = 2"
        >
          仅导入合格项 ({{ qualifiedCount }}条)
        </el-button>
      </div>
    </div>

    <!-- Step 2: Confirm -->
    <div v-show="currentStep === 2" class="step-content">
      <div class="confirm-box">
        <el-icon color="var(--success)" :size="48"><SuccessFilled /></el-icon>
        <p class="confirm-text">确认导入 {{ qualifiedCount }} 条{{ templateLabel }}信息？</p>
      </div>
      <div class="step-actions">
        <el-button @click="currentStep = 1">上一步</el-button>
        <el-button type="primary" :loading="importing" @click="confirmImport">
          确认导入
        </el-button>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.excel-import-wizard {
  width: 100%;
}

.step-content {
  padding: 8px 0;
}

.excel-upload-area {
  width: 100%;
  border: 2px dashed var(--el-border-color);
  border-radius: 12px;
  padding: 48px;
  text-align: center;
  transition: border-color 0.2s;

  :deep(.el-upload-dragger) {
    background: transparent;
    border: none;
    padding: 0;
  }

  &:hover {
    border-color: var(--el-color-primary);
  }
}

.upload-icon {
  font-size: 48px;
  color: var(--el-color-primary);
  margin-bottom: 16px;
  display: block;
}

.upload-text {
  font-size: 16px;
  color: var(--text-primary);
  margin-bottom: 8px;

  em {
    color: var(--el-color-primary);
    font-style: normal;
  }
}

.upload-hint {
  font-size: 12px;
  color: var(--text-secondary);
}

.step-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.preview-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
  font-size: 14px;
}

.qualified-badge {
  background: #10B981;
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.error-badge {
  background: #EF4444;
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.cell-error {
  color: #EF4444;
}

.field-error {
  display: block;
  font-size: 12px;
  color: #EF4444;
}

:deep(.qualified-row) {
  border-left: 4px solid #10B981;
  background: rgba(16, 185, 129, 0.05);
}

:deep(.error-row) {
  border-left: 4px solid #EF4444;
  background: rgba(239, 68, 68, 0.05);
}

.confirm-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 32px;
  text-align: center;
}

.confirm-text {
  font-size: 16px;
  font-weight: 500;
  color: var(--text-primary);
  margin: 0;
}
</style>
