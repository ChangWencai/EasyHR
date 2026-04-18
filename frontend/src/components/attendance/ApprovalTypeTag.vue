<template>
  <el-tag :type="tagType" size="small">{{ typeName }}</el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ type: string }>()

const typeMap: Record<string, { name: string; color: '' | 'success' | 'warning' | 'danger' | 'info' }> = {
  makeup: { name: '补卡', color: 'info' },
  shift_swap: { name: '调班', color: 'info' },
  business_trip: { name: '出差', color: '' },
  outside: { name: '外出', color: '' },
  personal_leave: { name: '事假', color: 'warning' },
  sick_leave: { name: '病假', color: 'warning' },
  PTO: { name: '调休', color: 'success' },
  annual_leave: { name: '年假', color: 'success' },
  marriage_leave: { name: '婚假', color: '' },
  maternity_leave: { name: '产假', color: '' },
  paternity_leave: { name: '陪产假', color: '' },
  overtime: { name: '加班', color: 'danger' },
}

const typeName = computed(() => typeMap[props.type]?.name ?? props.type)
const tagType = computed(() => typeMap[props.type]?.color ?? 'info')
</script>
