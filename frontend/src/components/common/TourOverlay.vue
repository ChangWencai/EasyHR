<script setup lang="ts">
/**
 * TourOverlay.vue — 首次使用引导遮罩气泡
 * 支持3个引导点 + 跳过 + localStorage 持久化
 * 设计规范: .planning/phases/10-UX基础-流程简化与引导体系/10-UI-SPEC.md
 */

export interface TourStep {
  title: string
  body: string
  /** CSS selector for the target element to highlight */
  target?: string
}

const props = defineProps<{
  steps: TourStep[]
  /** v-model compatible: controls visibility */
  visible: boolean
}>()

const emit = defineEmits<{
  'update:visible': [visible: boolean]
  close: []
  complete: []
}>()

const currentStep = ref(0)
const TOUR_DONE_KEY = 'hasSeenTour'

// Reset step when hidden
watch(() => props.visible, (val) => {
  if (!val) currentStep.value = 0
})

function next() {
  if (currentStep.value < props.steps.length - 1) {
    currentStep.value++
  } else {
    complete()
  }
}

function prev() {
  if (currentStep.value > 0) currentStep.value--
}

function skip() {
  complete()
}

function complete() {
  localStorage.setItem(TOUR_DONE_KEY, 'true')
  emit('update:visible', false)
  emit('complete')
}

/** Position tooltip relative to target element, falling back to center */
function getTooltipStyle(target?: string): Record<string, string> {
  if (!target) {
    return { top: '40%', left: '50%', transform: 'translate(-50%, -50%)' }
  }
  const el = document.querySelector(target)
  if (!el) {
    return { top: '40%', left: '50%', transform: 'translate(-50%, -50%)' }
  }
  const rect = el.getBoundingClientRect()
  return {
    position: 'fixed',
    top: `${rect.top + rect.height / 2}px`,
    left: `${rect.right + 16}px`,
    transform: 'translateY(-50%)',
  }
}

// Highlight target element and scroll it into view
watch(currentStep, () => {
  // Remove previous highlights
  document.querySelectorAll('.tour-highlight').forEach(el => el.classList.remove('tour-highlight'))
  const step = props.steps[currentStep.value]
  if (step?.target) {
    const el = document.querySelector(step.target)
    if (el) {
      el.classList.add('tour-highlight')
      el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
  }
})

// Clean up highlights when overlay unmounts
onUnmounted(() => {
  document.querySelectorAll('.tour-highlight').forEach(el => el.classList.remove('tour-highlight'))
})
</script>

<template>
  <Teleport to="body">
    <Transition name="tour-fade">
      <div v-if="visible" class="tour-overlay" @click.self="skip">
        <!-- Backdrop -->
        <div class="tour-backdrop" @click.self="skip" />

        <!-- Dot indicator + skip button -->
        <div class="tour-header">
          <div class="tour-dots">
            <span
              v-for="(_, i) in steps"
              :key="i"
              class="dot"
              :class="{ active: i === currentStep, done: i < currentStep }"
            />
          </div>
          <button class="skip-btn" @click="skip">跳过引导</button>
        </div>

        <!-- Tooltip bubble -->
        <div
          v-if="steps[currentStep]"
          class="tour-tooltip"
          :style="getTooltipStyle(steps[currentStep].target)"
        >
          <h4 class="tooltip-title">{{ steps[currentStep].title }}</h4>
          <p class="tooltip-body">{{ steps[currentStep].body }}</p>
          <div class="tooltip-actions">
            <el-button
              v-if="currentStep > 0"
              size="small"
              @click="prev"
            >上一步</el-button>
            <el-button
              type="primary"
              size="small"
              @click="next"
            >{{ currentStep === steps.length - 1 ? '完成' : '下一步' }}</el-button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.tour-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  pointer-events: auto;
}

.tour-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  pointer-events: auto;
}

.tour-header {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  padding: 16px 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  z-index: 10001;
}

.tour-dots {
  display: flex;
  gap: 8px;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.4);
  transition: background 0.2s;
}

.dot.active { background: var(--primary, #7C3AED) }
.dot.done { background: rgba(124, 58, 237, 0.6) }

.skip-btn {
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.8);
  font-size: 14px;
  cursor: pointer;
  padding: 4px 8px;
}

.skip-btn:hover { color: white }

.tour-tooltip {
  position: fixed;
  z-index: 10000;
  background: #FFFFFF;
  border-radius: 12px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
  padding: 20px;
  min-width: 240px;
  max-width: 320px;
}

.tooltip-title {
  font-size: 16px;
  font-weight: 700;
  margin: 0 0 8px 0;
  color: #374151;
}

.tooltip-body {
  font-size: 14px;
  color: #6B7280;
  margin: 0 0 16px 0;
  line-height: 1.5;
}

.tooltip-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.tour-fade-enter-active,
.tour-fade-leave-active {
  transition: opacity 0.3s ease;
}

.tour-fade-enter-from,
.tour-fade-leave-to {
  opacity: 0;
}
</style>

<!--
  Tour highlight — applied to target elements during tour.
  This is a global class (not scoped) so it can target elements outside this component.
-->
<style>
.tour-highlight {
  position: relative;
  z-index: 10000 !important;
  box-shadow: 0 0 0 4px var(--primary), 0 0 0 8px rgba(124, 58, 237, 0.2);
  border-radius: 8px;
}
</style>
