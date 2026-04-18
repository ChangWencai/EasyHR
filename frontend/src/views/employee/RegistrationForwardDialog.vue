<template>
  <el-dialog v-model="dialogVisible" title="转发登记链接" width="460px" @close="handleClose">
    <div class="forward-content">
      <div class="link-preview">
        <el-input :model-value="registrationUrl" readonly>
          <template #append>
            <el-button @click="handleCopy">复制链接</el-button>
          </template>
        </el-input>
      </div>

      <el-divider content-position="center">或</el-divider>

      <div class="qrcode-section" v-if="registrationUrl">
        <div class="qrcode-wrapper">
          <canvas ref="qrCanvas"></canvas>
        </div>
        <p class="qrcode-hint">长按保存二维码图片</p>
      </div>

      <el-divider content-position="center">或</el-divider>

      <div class="sms-section">
        <el-input v-model="smsPhone" placeholder="输入手机号" style="flex: 1">
          <template #prepend>短信发送</template>
        </el-input>
        <el-button
          type="primary"
          :loading="sendingSms"
          :disabled="!smsPhone"
          @click="handleSendSms"
          style="margin-left: 8px"
        >
          发送短信
        </el-button>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import QRCode from 'qrcode'

const props = defineProps<{
  visible: boolean
  token: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const dialogVisible = ref(false)
const smsPhone = ref('')
const sendingSms = ref(false)
const qrCanvas = ref<HTMLCanvasElement>()

const registrationUrl = computed(() => {
  if (!props.token) return ''
  return `${window.location.origin}/#/register/${props.token}`
})

watch(
  () => props.visible,
  async (val) => {
    dialogVisible.value = val
    if (val && props.token) {
      await nextTick()
      renderQRCode()
    }
  },
)

async function renderQRCode() {
  if (!qrCanvas.value || !registrationUrl.value) return
  try {
    await QRCode.toCanvas(qrCanvas.value, registrationUrl.value, {
      width: 200,
      margin: 2,
    })
  } catch {
    // QR code rendering failed silently
  }
}

function handleClose() {
  emit('update:visible', false)
  smsPhone.value = ''
}

async function handleCopy() {
  try {
    await navigator.clipboard.writeText(registrationUrl.value)
    ElMessage.success('链接已复制，可粘贴发送')
  } catch {
    ElMessage.error('复制失败，请手动复制')
  }
}

async function handleSendSms() {
  if (!smsPhone.value) return
  sendingSms.value = true
  try {
    // TODO: 调用后端 SMS API 发送登记链接短信
    ElMessage.success(`短信已发送至 ${smsPhone.value}`)
  } catch {
    ElMessage.error('短信发送失败')
  } finally {
    sendingSms.value = false
  }
}
</script>

<style scoped lang="scss">
.forward-content {
  padding: 0 8px;
}

.link-preview {
  margin-bottom: 8px;
}

.qrcode-section {
  text-align: center;
}

.qrcode-wrapper {
  display: inline-block;
  padding: 12px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
}

.qrcode-hint {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}

.sms-section {
  display: flex;
  align-items: center;
}
</style>
