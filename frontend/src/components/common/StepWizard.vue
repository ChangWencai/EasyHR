<script setup lang="ts">
import { computed } from 'vue'

interface Step {
  title: string
}

const props = defineProps<{
  steps: Step[]
  currentStep: number
  finishText?: string
}>()

const emit = defineEmits<{
  'update:currentStep': [step: number]
  complete: []
}>()

const next = () => {
  if (props.currentStep < props.steps.length - 1) {
    emit('update:currentStep', props.currentStep + 1)
  }
}
</script>

<template>
  <div class="step-wizard">
    <el-steps :active="currentStep" finish-status="success" align-center>
      <el-step v-for="(step, i) in steps" :key="i" :title="step.title" />
    </el-steps>
    <div class="step-content">
      <slot :step="currentStep" />
    </div>
    <div class="step-actions">
      <el-button
        v-if="currentStep > 0"
        @click="emit('update:currentStep', currentStep - 1)"
      >
        上一步
      </el-button>
      <el-button
        v-if="currentStep < steps.length - 1"
        type="primary"
        @click="next"
      >
        下一步
      </el-button>
      <el-button
        v-else
        type="primary"
        @click="emit('complete')"
      >
        {{ finishText || '确认' }}
      </el-button>
    </div>
  </div>
</template>

<style scoped lang="scss">
$primary: var(--primary, #6366f1);
$primary-dark: var(--primary-dark, #4f46e5);
$primary-light: var(--primary-light, #a5b4fc);
$bg-surface: #FFFFFF;
$bg-page: #FAFBFC;
$text-secondary: #6B7280;
$border-color: #E5E7EB;
$radius-lg: 16px;

.step-wizard {
  width: 100%;
}

.step-wizard .el-step__title {
  font-size: 14px;
  font-weight: 500;
}

.step-wizard .el-step__title.is-finish {
  color: $primary;
}

.step-content {
  margin: 20px 0;
}

.step-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;

  .el-button {
    height: 52px;
    padding: 0 32px;
    border-radius: $radius-lg;
    font-size: 16px;
    font-weight: 600;
  }

  .el-button--primary {
    background: linear-gradient(135deg, $primary 0%, $primary-dark 100%);
    border: none;
    box-shadow: 0 4px 14px rgba($primary, 0.4);
    transition: all 0.2s ease;

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 6px 20px rgba($primary, 0.5);
    }

    &:active {
      transform: translateY(0);
    }
  }

  .el-button:not(.el-button--primary) {
    background: $bg-surface;
    border: 1px solid $border-color;
    color: $text-secondary;

    &:hover {
      border-color: $primary-light;
      color: $primary;
      background: rgba($primary, 0.04);
    }
  }
}
</style>
