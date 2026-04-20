<script setup lang="ts">
import type { ContractType } from '@/api/contract'

const props = defineProps<{
  modelValue: ContractType | null
}>()
const emit = defineEmits<{
  'update:modelValue': [val: ContractType]
}>()

const types = [
  {
    value: 'fixed_term' as ContractType,
    title: '劳动合同',
    subtitle: '（固定期限）',
    desc: '适用场景：',
    scenario: '正式员工',
  },
  {
    value: 'intern' as ContractType,
    title: '实习协议',
    subtitle: '',
    desc: '适用场景：',
    scenario: '在校学生实习',
  },
  {
    value: 'indefinite' as ContractType,
    title: '兼职合同',
    subtitle: '',
    desc: '适用场景：',
    scenario: '兼职人员',
  },
]

function select(val: ContractType) {
  emit('update:modelValue', val)
}
</script>

<template>
  <div class="contract-type-select">
    <div
      v-for="type in types"
      :key="type.value"
      class="type-card"
      :class="{ selected: modelValue === type.value }"
      @click="select(type.value)"
    >
      <div class="type-title">{{ type.title }}{{ type.subtitle }}</div>
      <div class="type-desc">{{ type.desc }}</div>
      <div class="type-scenario">{{ type.scenario }}</div>
      <div class="type-radio">
        <div class="radio-dot" :class="{ active: modelValue === type.value }" />
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.contract-type-select {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  @media (max-width: 768px) {
    grid-template-columns: 1fr;
  }
}

.type-card {
  position: relative;
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    border-color: var(--primary-light, #a78bfa);
    background: rgba(124, 58, 237, 0.02);
  }

  &.selected {
    border: 2px solid var(--primary, #7C3AED);
    background: rgba(124, 58, 237, 0.04);
  }
}

.type-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.type-desc {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 2px;
}

.type-scenario {
  font-size: 12px;
  color: var(--text-tertiary, #909399);
}

.type-radio {
  position: absolute;
  top: 12px;
  right: 12px;
  width: 16px;
  height: 16px;
  border: 2px solid var(--border);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;

  .radio-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: transparent;
    transition: background 0.2s;

    &.active {
      background: var(--primary, #7C3AED);
    }
  }
}
</style>
