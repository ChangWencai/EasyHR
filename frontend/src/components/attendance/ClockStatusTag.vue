<template>
  <el-tag :type="tagType" size="small" :style="{ color: tagColor }">
    {{ label }}
  </el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: 'normal' | 'late' | 'absent' | 'no_schedule' | 'not_clocked_in'
}>()

const labelMap: Record<string, string> = {
  normal: '正常',
  late: '迟到',
  absent: '缺勤',
  no_schedule: '未排班',
  not_clocked_in: '未打卡',
}

const tagColorMap: Record<string, string> = {
  normal: '#52C41A',
  late: '#FA8C16',
  absent: '#FF4D4F',
  no_schedule: '#BFBFBF',
  not_clocked_in: '#BFBFBF',
}

const tagType = computed(() => {
  if (props.status === 'normal') return 'success'
  if (props.status === 'late') return 'warning'
  if (props.status === 'absent') return 'danger'
  return 'info'
})

const label = computed(() => labelMap[props.status] || props.status)
const tagColor = computed(() => tagColorMap[props.status] || '#8C8C8C')
</script>
