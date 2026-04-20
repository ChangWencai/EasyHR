<script setup lang="ts">
import { computed } from 'vue'
import type { ContractStatus } from '@/api/contract'

const props = defineProps<{
  status: ContractStatus
}>()

const config: Record<ContractStatus, { type: string; label: string }> = {
  draft:        { type: 'info',    label: '草稿' },
  pending_sign: { type: 'warning', label: '待签署' },
  signed:       { type: 'success', label: '已签' },
  active:       { type: 'success', label: '生效中' },
  terminated:   { type: 'info',    label: '已终止' },
  expired:      { type: 'danger',  label: '已过期' },
}

const item = computed(() => config[props.status] || { type: 'info', label: props.status })
</script>

<template>
  <el-tag :type="item.type" size="small">{{ item.label }}</el-tag>
</template>
