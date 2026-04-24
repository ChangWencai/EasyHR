<script setup lang="ts">
const props = defineProps<{
  modelValue: { start: string; end: string | null; probationMonths: number }
}>()
const emit = defineEmits<{
  'update:modelValue': [val: { start: string; end: string | null; probationMonths: number }]
}>()

function updateStart(val: string) {
  emit('update:modelValue', { ...props.modelValue, start: val })
}
function updateEnd(val: string | null) {
  emit('update:modelValue', { ...props.modelValue, end: val })
}
function clearEnd() {
  emit('update:modelValue', { ...props.modelValue, end: null })
}
function updateProbation(val: number) {
  emit('update:modelValue', { ...props.modelValue, probationMonths: val })
}
</script>

<template>
  <div class="contract-period-picker">
    <div class="period-row">
      <label>起始日期</label>
      <el-date-picker
        :model-value="modelValue.start"
        type="date"
        value-format="YYYY-MM-DD"
        placeholder="选择开始日期"
        :disabled-date="(d: Date) => d < new Date(new Date().setHours(0,0,0,0))"
        @update:model-value="updateStart"
        style="flex: 1"
      />
    </div>
    <div class="period-row">
      <label>结束日期</label>
      <el-date-picker
        v-if="modelValue.end !== undefined"
        :model-value="modelValue.end"
        type="date"
        value-format="YYYY-MM-DD"
        placeholder="无固定期限可不选"
        :disabled-date="(d: Date) => props.modelValue.start && d < new Date(props.modelValue.start)"
        @update:model-value="(v: string) => updateEnd(v || null)"
        style="flex: 1"
      />
    </div>
    <div class="period-row">
      <label>试用期</label>
      <el-select
        :model-value="modelValue.probationMonths"
        placeholder="请选择试用期"
        style="flex: 1"
        @update:model-value="updateProbation"
      >
        <el-option :value="0" label="无试用期" />
        <el-option v-for="m in 6" :key="m" :value="m" :label="`${m} 个月`" />
      </el-select>
    </div>
    <el-button v-if="modelValue.end" link size="small" @click="clearEnd">
      设为无固定期限
    </el-button>
  </div>
</template>

<style scoped lang="scss">
.contract-period-picker {
  .period-row {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;

    label {
      width: 80px;
      font-size: 13px;
      color: var(--text-secondary);
      flex-shrink: 0;
    }
  }
}
</style>
