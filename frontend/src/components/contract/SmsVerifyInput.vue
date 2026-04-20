<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'

const props = defineProps<{
  modelValue: string
  countdown?: number
}>()
const emit = defineEmits<{
  'update:modelValue': [val: string]
  'send-code': []
}>()

const digits = ref(['', '', '', '', '', ''])
const inputRefs = ref<HTMLInputElement[]>([])

// Sync from external modelValue
watch(() => props.modelValue, (val) => {
  for (let i = 0; i < 6; i++) {
    digits.value[i] = val[i] || ''
  }
})

// Emit on change
function onInput(index: number, e: Event) {
  const target = e.target as HTMLInputElement
  const value = target.value.replace(/\D/g, '')
  digits.value[index] = value.slice(-1)

  if (value && index < 5) {
    nextTick(() => inputRefs.value[index + 1]?.focus())
  }

  emit('update:modelValue', digits.value.join(''))
}

function onKeydown(index: number, e: KeyboardEvent) {
  if (e.key === 'Backspace' && !digits.value[index] && index > 0) {
    inputRefs.value[index - 1]?.focus()
  }
}

function onPaste(e: ClipboardEvent) {
  e.preventDefault()
  const pasted = e.clipboardData?.getData('text').replace(/\D/g, '').slice(0, 6) || ''
  for (let i = 0; i < 6; i++) {
    digits.value[i] = pasted[i] || ''
  }
  emit('update:modelValue', digits.value.join(''))
}

function focusInput(index: number) {
  inputRefs.value[index]?.focus()
}
</script>

<template>
  <div class="sms-verify-input">
    <div class="digit-boxes" @paste="onPaste">
      <input
        v-for="(_, i) in 6"
        :key="i"
        :ref="el => { if (el) inputRefs[i] = el as HTMLInputElement }"
        type="number"
        inputmode="numeric"
        maxlength="1"
        class="digit-box"
        :value="digits[i]"
        @input="onInput(i, $event)"
        @keydown="onKeydown(i, $event)"
        @focus="focusInput(i)"
      />
    </div>
    <div class="countdown-hint" v-if="countdown && countdown > 0">
      {{ countdown }}秒后重新获取
    </div>
    <el-button v-else link type="primary" size="small" @click="emit('send-code')">
      重新获取验证码
    </el-button>
  </div>
</template>

<style scoped lang="scss">
.sms-verify-input {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.digit-boxes {
  display: flex;
  gap: 8px;
}

.digit-box {
  width: 44px;
  height: 52px;
  text-align: center;
  font-size: 20px;
  font-weight: 600;
  border: 1px solid var(--border);
  border-radius: 8px;
  outline: none;
  transition: border-color 0.2s;

  &:focus {
    border-color: var(--primary, #7C3AED);
  }

  /* Hide spinner arrows */
  &::-webkit-outer-spin-button,
  &::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }
  -moz-appearance: textfield;
}

.countdown-hint {
  font-size: 12px;
  color: var(--text-secondary);
}
</style>
